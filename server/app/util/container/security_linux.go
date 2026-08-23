//go:build linux && cgo

package container

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	pathrs "github.com/cyphar/filepath-securejoin/pathrs-lite"
	mobySeccomp "github.com/moby/profiles/seccomp"
	"github.com/moby/sys/mountinfo"
	userUtil "github.com/moby/sys/user"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runc/libcontainer/seccomp"
	"github.com/opencontainers/runc/libcontainer/specconv"
	"github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/sys/unix"
)

type linuxSecurity struct {
	capabilities *configs.Capabilities
	seccomp      *configs.Seccomp
}

func (m *Manager) processCapabilities(uid int) *configs.Capabilities {
	if uid == 0 {
		return nil
	}
	return &configs.Capabilities{}
}

func (m *Manager) configurePlatformSecurity(root string) error {
	if os.Geteuid() != 0 {
		return errors.New("容器运行时必须由 root 启动")
	}
	if !seccomp.Enabled {
		return errors.New("当前二进制未启用 seccomp，请使用静态 Linux 发布构建")
	}
	probe, err := unix.Openat2(unix.AT_FDCWD, root, &unix.OpenHow{
		Flags:   uint64(unix.O_PATH | unix.O_DIRECTORY | unix.O_CLOEXEC | unix.O_NOFOLLOW),
		Resolve: unix.RESOLVE_NO_MAGICLINKS | unix.RESOLVE_NO_SYMLINKS,
	})
	if err != nil {
		if errors.Is(err, unix.ENOSYS) {
			return errors.New("当前内核不支持 openat2，无法安全运行容器")
		}
		return fmt.Errorf("安全打开容器数据目录失败: %w", err)
	}
	_ = unix.Close(probe)
	managedRoots := []string{m.dataRoot, m.imagesRoot, m.instancesRoot, m.runtimeRoot, m.volumesRoot, m.mountsRoot}
	for _, managedRoot := range managedRoots {
		info, err := os.Lstat(managedRoot)
		if err != nil {
			return fmt.Errorf("读取容器受管目录失败: %w", err)
		}
		uid, gid, ok := archiveOwnership(info)
		if !ok || uid != 0 || gid != 0 || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("容器受管目录必须由 root:root 持有且不能是符号链接: %s", managedRoot)
		}
		if err := os.Chmod(managedRoot, 0700); err != nil {
			return fmt.Errorf("收紧容器受管目录权限失败: %w", err)
		}
	}
	for _, mountRoot := range m.mountRoots {
		info, err := os.Lstat(mountRoot)
		if err != nil {
			return fmt.Errorf("读取挂载根目录失败: %w", err)
		}
		uid, gid, ok := archiveOwnership(info)
		if !ok || uid != 0 || gid != 0 || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0077 != 0 {
			return fmt.Errorf("挂载根目录必须由 root:root 持有且不能允许组或其他用户访问: %s", mountRoot)
		}
		for current := filepath.Dir(mountRoot); ; current = filepath.Dir(current) {
			parentInfo, parentErr := os.Lstat(current)
			if parentErr != nil {
				return fmt.Errorf("读取挂载根目录父路径失败: %w", parentErr)
			}
			parentUID, _, ownerOK := archiveOwnership(parentInfo)
			writable := parentInfo.Mode().Perm()&0022 != 0
			if !ownerOK || parentUID != 0 || !parentInfo.IsDir() || parentInfo.Mode()&os.ModeSymlink != 0 ||
				writable && parentInfo.Mode()&os.ModeSticky == 0 {
				return fmt.Errorf("挂载根目录父路径可被非 root 替换: %s", current)
			}
			if current == filepath.Dir(current) {
				break
			}
		}
	}
	if err := m.cleanupPlatformMountStaging(); err != nil {
		return fmt.Errorf("清理遗留挂载源失败: %w", err)
	}
	options := m.managerOptions()
	if options.UIDMapStart == 0 != (options.GIDMapStart == 0) {
		return errors.New("UIDMapStart 和 GIDMapStart 必须同时配置")
	}
	validate := func(value securityMetadata) error {
		if value.Version != securityMetadataVersion {
			return fmt.Errorf("不支持的容器安全元数据版本: %d", value.Version)
		}
		if value.UIDMapStart < defaultIDMapSize || value.GIDMapStart < defaultIDMapSize {
			return errors.New("User Namespace 的宿主 UID/GID 起点不能小于 65536")
		}
		if value.IDMapSize < defaultIDMapSize {
			return errors.New("User Namespace 至少需要映射 65536 个 UID/GID")
		}
		if value.IDMapSize > value.UIDMapStart || value.IDMapSize > value.GIDMapStart {
			return errors.New("User Namespace 原始 ID 与宿主映射区间不能重叠")
		}
		if value.UIDMapStart > math.MaxInt32-value.IDMapSize+1 || value.GIDMapStart > math.MaxInt32-value.IDMapSize+1 {
			return errors.New("User Namespace UID/GID 映射超出系统范围")
		}
		return nil
	}
	metadataPath := filepath.Join(root, "security.json")
	var metadata securityMetadata
	if err := m.validateManagedNode(metadataPath, false, true); err != nil {
		return fmt.Errorf("容器安全配置路径不安全: %w", err)
	}
	if err := m.readJSON(metadataPath, &metadata); err == nil {
		if err := validate(metadata); err != nil {
			return fmt.Errorf("容器安全配置无效: %w", err)
		}
		if (m.requestedSecurity.UIDMapStart > 0 && m.requestedSecurity.UIDMapStart != metadata.UIDMapStart) ||
			(m.requestedSecurity.GIDMapStart > 0 && m.requestedSecurity.GIDMapStart != metadata.GIDMapStart) ||
			(m.requestedSecurity.IDMapSize > 0 && m.requestedSecurity.IDMapSize != metadata.IDMapSize) {
			return errors.New("传入的 User Namespace 映射与容器数据目录已有配置不一致")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("读取容器安全配置失败: %w", err)
	} else {
		metadata = securityMetadata{
			Version:     securityMetadataVersion,
			UIDMapStart: options.UIDMapStart,
			GIDMapStart: options.GIDMapStart,
			IDMapSize:   options.IDMapSize,
		}
		if metadata.UIDMapStart == 0 {
			uids, uidErr := userUtil.CurrentUserSubUIDs()
			gids, gidErr := userUtil.CurrentUserSubGIDs()
			if uidErr != nil || gidErr != nil {
				return fmt.Errorf("读取 /etc/subuid 或 /etc/subgid 失败: %w", errors.Join(uidErr, gidErr))
			}
			for _, value := range uids {
				if value.Count >= metadata.IDMapSize {
					metadata.UIDMapStart = value.SubID
					break
				}
			}
			for _, value := range gids {
				if value.Count >= metadata.IDMapSize {
					metadata.GIDMapStart = value.SubID
					break
				}
			}
			if metadata.UIDMapStart == 0 || metadata.GIDMapStart == 0 {
				return errors.New("当前用户缺少至少 65536 个连续的 subuid/subgid，请配置 /etc/subuid、/etc/subgid 或显式传入映射")
			}
		}
		if err := validate(metadata); err != nil {
			return err
		}
		if err := m.writeJSON(metadataPath, &metadata); err != nil {
			return fmt.Errorf("保存容器安全配置失败: %w", err)
		}
	}

	capabilities := []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_FSETID",
		"CAP_KILL",
		"CAP_SETGID",
		"CAP_SETUID",
		"CAP_SYS_CHROOT",
	}
	profile, err := mobySeccomp.GetDefaultProfile(&specs.Spec{
		Process: &specs.Process{Capabilities: &specs.LinuxCapabilities{
			Bounding:  append([]string(nil), capabilities...),
			Effective: append([]string(nil), capabilities...),
			Permitted: append([]string(nil), capabilities...),
		}},
	})
	if err != nil {
		return fmt.Errorf("生成 seccomp 配置失败: %w", err)
	}
	seccompConfig, err := specconv.SetupSeccomp(profile)
	if err != nil {
		return fmt.Errorf("转换 seccomp 配置失败: %w", err)
	}
	m.security = metadata
	m.platformSecurity = &linuxSecurity{
		capabilities: &configs.Capabilities{
			Bounding:  append([]string(nil), capabilities...),
			Effective: append([]string(nil), capabilities...),
			Permitted: append([]string(nil), capabilities...),
		},
		seccomp: seccompConfig,
	}
	return nil
}

func (m *Manager) normalizePlatformRootfsOwnership(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		uid, gid, ok := archiveOwnership(info)
		if !ok {
			return nil
		}
		mapID := func(value int, start int64, name string) (int, error) {
			current := int64(value)
			if current >= start && current < start+m.security.IDMapSize {
				return value, nil
			}
			if current < 0 || current >= m.security.IDMapSize {
				return 0, fmt.Errorf("镜像路径 %s 的 %s %d 超出 User Namespace 映射范围", path, name, value)
			}
			return int(start + current), nil
		}
		mappedUID, err := mapID(uid, m.security.UIDMapStart, "UID")
		if err != nil {
			return err
		}
		mappedGID, err := mapID(gid, m.security.GIDMapStart, "GID")
		if err != nil {
			return err
		}
		if uid == mappedUID && gid == mappedGID {
			return nil
		}
		if err := os.Lchown(path, mappedUID, mappedGID); err != nil {
			return fmt.Errorf("映射镜像路径所有者失败 %s: %w", path, err)
		}
		return nil
	})
}

func (m *Manager) normalizePlatformPathOwnership(path, root string) error {
	root = filepath.Clean(root)
	for current := filepath.Clean(path); ; current = filepath.Dir(current) {
		if current != root {
			relative, err := filepath.Rel(root, current)
			if err != nil || relative == ".." || filepath.IsAbs(relative) {
				return fmt.Errorf("rootfs 路径超出容器根目录: %s", current)
			}
		}
		info, err := os.Lstat(current)
		if err != nil {
			return err
		}
		uid, gid, ok := archiveOwnership(info)
		if ok {
			mapID := func(value int, start int64) (int, error) {
				currentID := int64(value)
				if currentID >= start && currentID < start+m.security.IDMapSize {
					return value, nil
				}
				if currentID < 0 || currentID >= m.security.IDMapSize {
					return 0, fmt.Errorf("rootfs 路径所有者 %d 超出 User Namespace 映射范围", value)
				}
				return int(start + currentID), nil
			}
			mappedUID, uidErr := mapID(uid, m.security.UIDMapStart)
			mappedGID, gidErr := mapID(gid, m.security.GIDMapStart)
			if uidErr != nil || gidErr != nil {
				return errors.Join(uidErr, gidErr)
			}
			if uid != mappedUID || gid != mappedGID {
				if err := os.Lchown(current, mappedUID, mappedGID); err != nil {
					return fmt.Errorf("映射 rootfs 路径所有者失败 %s: %w", current, err)
				}
			}
		}
		if current == root {
			return nil
		}
	}
}

func (m *Manager) openPlatformMount(source string) (*os.File, error) {
	mountRoot := ""
	for _, candidate := range m.mountRoots {
		if m.ensureInside(candidate, source) == nil && len(candidate) > len(mountRoot) {
			mountRoot = candidate
		}
	}
	if mountRoot == "" {
		return nil, errors.New("挂载源不在允许的挂载根目录内")
	}
	mountRootFD, err := unix.Openat2(unix.AT_FDCWD, mountRoot, &unix.OpenHow{
		Flags:   uint64(unix.O_PATH | unix.O_DIRECTORY | unix.O_CLOEXEC | unix.O_NOFOLLOW),
		Resolve: unix.RESOLVE_NO_MAGICLINKS | unix.RESOLVE_NO_SYMLINKS,
	})
	if err != nil {
		return nil, fmt.Errorf("安全打开挂载根目录失败: %w", err)
	}
	defer unix.Close(mountRootFD)
	relativeRoot, err := filepath.Rel(mountRoot, source)
	if err != nil || relativeRoot == ".." || filepath.IsAbs(relativeRoot) {
		return nil, errors.New("挂载源超出允许的挂载根目录")
	}
	sourceFD, err := unix.Openat2(mountRootFD, relativeRoot, &unix.OpenHow{
		Flags:   uint64(unix.O_PATH | unix.O_CLOEXEC | unix.O_NOFOLLOW),
		Resolve: unix.RESOLVE_BENEATH | unix.RESOLVE_NO_MAGICLINKS | unix.RESOLVE_NO_SYMLINKS | unix.RESOLVE_NO_XDEV,
	})
	if err != nil {
		return nil, fmt.Errorf("安全打开挂载源失败: %w", err)
	}
	return os.NewFile(uintptr(sourceFD), source), nil
}

func (m *Manager) scanPlatformMount(root string, prepare, requireMapped bool) error {
	source, err := m.openPlatformMount(root)
	if err != nil {
		return err
	}
	defer source.Close()
	return m.scanPlatformMountHandle(root, source, prepare, requireMapped)
}

func (m *Manager) scanPlatformMountHandle(root string, source *os.File, prepare, requireMapped bool) error {
	rootDevice, err := m.validatePlatformMountTree(root)
	if err != nil {
		return err
	}
	sourceFD := int(source.Fd())
	var sourceStat unix.Stat_t
	if err := unix.Fstat(sourceFD, &sourceStat); err != nil {
		return fmt.Errorf("读取挂载源状态失败: %w", err)
	}
	if uint64(sourceStat.Dev) != rootDevice {
		return errors.New("挂载源在检查期间跨越了文件系统")
	}

	type nodeState struct {
		device uint64
		inode  uint64
		links  uint64
		mode   uint32
		uid    uint32
		gid    uint32
	}
	stateOf := func(stat *unix.Stat_t) nodeState {
		return nodeState{
			device: uint64(stat.Dev),
			inode:  uint64(stat.Ino),
			links:  uint64(stat.Nlink),
			mode:   stat.Mode,
			uid:    stat.Uid,
			gid:    stat.Gid,
		}
	}
	equalState := func(left, right nodeState) bool {
		return left == right
	}
	mapID := func(value uint32, start int64, currentPath, name string) (int, error) {
		current := int64(value)
		if current >= start && current < start+m.security.IDMapSize {
			return int(current), nil
		}
		if current < 0 || current >= m.security.IDMapSize {
			return 0, fmt.Errorf("挂载路径 %s 的 %s %d 超出 User Namespace 映射范围", currentPath, name, value)
		}
		return int(start + current), nil
	}
	snapshots := make(map[string]nodeState)
	walk := func(record, mutate bool) error {
		queue := []string{"."}
		seen := make(map[string]struct{})
		for len(queue) > 0 {
			relative := queue[0]
			queue = queue[1:]
			currentPath := root
			if relative != "." {
				currentPath = filepath.Join(root, relative)
			}
			pathFD := -1
			var openErr error
			if relative == "." {
				pathFD, openErr = unix.FcntlInt(uintptr(sourceFD), unix.F_DUPFD_CLOEXEC, 0)
			} else {
				pathFD, openErr = unix.Openat2(sourceFD, relative, &unix.OpenHow{
					Flags:   uint64(unix.O_PATH | unix.O_CLOEXEC | unix.O_NOFOLLOW),
					Resolve: unix.RESOLVE_BENEATH | unix.RESOLVE_NO_MAGICLINKS | unix.RESOLVE_NO_SYMLINKS | unix.RESOLVE_NO_XDEV,
				})
			}
			if openErr != nil {
				return fmt.Errorf("安全打开挂载路径失败 %s: %w", currentPath, openErr)
			}
			var pathStat unix.Stat_t
			if statErr := unix.Fstat(pathFD, &pathStat); statErr != nil {
				_ = unix.Close(pathFD)
				return fmt.Errorf("读取挂载路径状态失败 %s: %w", currentPath, statErr)
			}
			state := stateOf(&pathStat)
			if state.device != rootDevice {
				_ = unix.Close(pathFD)
				return fmt.Errorf("挂载目录不能跨文件系统: %s", currentPath)
			}
			kind := pathStat.Mode & unix.S_IFMT
			if kind != unix.S_IFREG && kind != unix.S_IFDIR && kind != unix.S_IFLNK {
				_ = unix.Close(pathFD)
				return fmt.Errorf("挂载目录只能包含普通文件、目录或符号链接: %s", currentPath)
			}
			if kind != unix.S_IFDIR && pathStat.Nlink > 1 {
				_ = unix.Close(pathFD)
				return fmt.Errorf("挂载目录不能包含硬链接: %s", currentPath)
			}
			if record {
				snapshots[relative] = state
			} else if prepare {
				expected, exists := snapshots[relative]
				if !exists || !equalState(expected, state) {
					_ = unix.Close(pathFD)
					return fmt.Errorf("挂载路径在准备期间被替换: %s", currentPath)
				}
				seen[relative] = struct{}{}
			}
			mappedUID, uidErr := mapID(pathStat.Uid, m.security.UIDMapStart, currentPath, "UID")
			mappedGID, gidErr := mapID(pathStat.Gid, m.security.GIDMapStart, currentPath, "GID")
			if prepare && (uidErr != nil || gidErr != nil) {
				_ = unix.Close(pathFD)
				return errors.Join(uidErr, gidErr)
			}
			if requireMapped && (int64(pathStat.Uid) < m.security.UIDMapStart || int64(pathStat.Uid) >= m.security.UIDMapStart+m.security.IDMapSize ||
				int64(pathStat.Gid) < m.security.GIDMapStart || int64(pathStat.Gid) >= m.security.GIDMapStart+m.security.IDMapSize) {
				_ = unix.Close(pathFD)
				return fmt.Errorf("可写挂载路径未准备 User Namespace 所有权: %s，请先调用 PrepareMount", currentPath)
			}
			if kind == unix.S_IFLNK {
				if mutate && (int(pathStat.Uid) != mappedUID || int(pathStat.Gid) != mappedGID) {
					if chownErr := unix.Fchownat(pathFD, "", mappedUID, mappedGID, unix.AT_EMPTY_PATH|unix.AT_SYMLINK_NOFOLLOW); chownErr != nil {
						_ = unix.Close(pathFD)
						return fmt.Errorf("映射挂载符号链接所有者失败 %s: %w", currentPath, chownErr)
					}
				}
				_ = unix.Close(pathFD)
				continue
			}
			flags := unix.O_RDONLY | unix.O_NONBLOCK
			if kind == unix.S_IFDIR {
				flags |= unix.O_DIRECTORY
			}
			handle := os.NewFile(uintptr(pathFD), currentPath)
			file, reopenErr := pathrs.Reopen(handle, flags)
			if reopenErr != nil {
				_ = handle.Close()
				return fmt.Errorf("安全重开挂载路径失败 %s: %w", currentPath, reopenErr)
			}
			fd := int(file.Fd())
			var reopenedStat unix.Stat_t
			if statErr := unix.Fstat(fd, &reopenedStat); statErr != nil || !equalState(state, stateOf(&reopenedStat)) {
				_ = file.Close()
				_ = handle.Close()
				if statErr != nil {
					return fmt.Errorf("复核挂载路径状态失败 %s: %w", currentPath, statErr)
				}
				return fmt.Errorf("挂载路径在打开期间被替换: %s", currentPath)
			}
			_ = handle.Close()
			if mutate && (int(reopenedStat.Uid) != mappedUID || int(reopenedStat.Gid) != mappedGID) {
				if chownErr := unix.Fchown(fd, mappedUID, mappedGID); chownErr != nil {
					_ = file.Close()
					return fmt.Errorf("映射挂载路径所有者失败 %s: %w", currentPath, chownErr)
				}
			}
			if mutate && kind == unix.S_IFREG {
				if reopenedStat.Mode&(unix.S_ISUID|unix.S_ISGID) != 0 {
					if chmodErr := unix.Fchmod(fd, reopenedStat.Mode&0777); chmodErr != nil {
						_ = file.Close()
						return fmt.Errorf("移除挂载文件提权位失败 %s: %w", currentPath, chmodErr)
					}
				}
				if xattrErr := unix.Fremovexattr(fd, "security.capability"); xattrErr != nil && !errors.Is(xattrErr, unix.ENODATA) && !errors.Is(xattrErr, unix.ENOTSUP) {
					_ = file.Close()
					return fmt.Errorf("移除挂载文件能力失败 %s: %w", currentPath, xattrErr)
				}
			}
			if kind == unix.S_IFDIR {
				entries, readErr := file.ReadDir(-1)
				closeErr := file.Close()
				if readErr != nil {
					return fmt.Errorf("读取挂载目录失败 %s: %w", currentPath, readErr)
				}
				if closeErr != nil {
					return fmt.Errorf("关闭挂载目录失败 %s: %w", currentPath, closeErr)
				}
				for _, entry := range entries {
					queue = append(queue, filepath.Join(relative, entry.Name()))
				}
			} else {
				_ = file.Close()
			}
		}
		if prepare && !record && len(seen) != len(snapshots) {
			return errors.New("挂载目录在准备期间发生变化")
		}
		return nil
	}
	if prepare {
		if err := walk(true, false); err != nil {
			return err
		}
		return walk(false, true)
	}
	return walk(false, false)
}

func (m *Manager) preparePlatformMountOwnership(root string) error {
	unlock, err := m.lockPlatformMountStaging()
	if err != nil {
		return err
	}
	return errors.Join(m.scanPlatformMount(root, true, false), unlock())
}

func (m *Manager) validatePlatformMountOwnership(root string) error {
	return m.scanPlatformMount(root, false, true)
}

func (m *Manager) validatePlatformMountContent(root string) error {
	return m.scanPlatformMount(root, false, false)
}

func (m *Manager) lockPlatformMountStaging() (func() error, error) {
	lockPath := filepath.Join(m.dataRoot, ".mount-staging.lock")
	fd, err := unix.Open(lockPath, unix.O_RDWR|unix.O_CREAT|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, fmt.Errorf("打开临时挂载锁失败: %w", err)
	}
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("读取临时挂载锁失败: %w", err)
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFREG || stat.Uid != 0 || stat.Gid != 0 || stat.Nlink != 1 || stat.Mode&0077 != 0 {
		_ = unix.Close(fd)
		return nil, errors.New("临时挂载锁必须是 root:root 持有的 0600 普通文件且不能是硬链接")
	}
	if err := unix.Flock(fd, unix.LOCK_EX); err != nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("锁定临时挂载目录失败: %w", err)
	}
	var once sync.Once
	var unlockErr error
	return func() error {
		once.Do(func() {
			unlockErr = errors.Join(unix.Flock(fd, unix.LOCK_UN), unix.Close(fd))
		})
		return unlockErr
	}, nil
}

func (m *Manager) cleanupPlatformMountStaging() error {
	unlock, err := m.lockPlatformMountStaging()
	if err != nil {
		return err
	}
	return errors.Join(m.cleanupPlatformMountStagingLocked(), unlock())
}

func (m *Manager) cleanupPlatformMountStagingLocked() error {
	markerPath := filepath.Join(m.mountsRoot, ".managed")
	markerInfo, err := os.Lstat(markerPath)
	if err != nil {
		return fmt.Errorf("读取临时挂载源管理标记失败: %w", err)
	}
	uid, gid, ownerOK := archiveOwnership(markerInfo)
	expectedMarker := m.installation + "\n"
	if !ownerOK || uid != 0 || gid != 0 || !markerInfo.Mode().IsRegular() || markerInfo.Mode()&os.ModeSymlink != 0 ||
		markerInfo.Mode().Perm()&0077 != 0 || markerInfo.Size() != int64(len(expectedMarker)) {
		return errors.New("临时挂载源管理标记无效，拒绝清理")
	}
	marker, err := os.ReadFile(markerPath)
	if err != nil || string(marker) != expectedMarker {
		return errors.New("临时挂载源管理标记不匹配，拒绝清理")
	}
	entries, err := os.ReadDir(m.mountsRoot)
	if err != nil {
		return err
	}
	stageDirs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Name() == ".managed" {
			continue
		}
		if !strings.HasPrefix(entry.Name(), ".source-") {
			return fmt.Errorf("临时挂载源目录包含未知条目，拒绝清理: %s", entry.Name())
		}
		stageDir := filepath.Join(m.mountsRoot, entry.Name())
		info, err := os.Lstat(stageDir)
		if err != nil {
			return fmt.Errorf("读取临时挂载目录失败: %w", err)
		}
		uid, gid, ownerOK := archiveOwnership(info)
		if !ownerOK || uid != 0 || gid != 0 || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0077 != 0 {
			return fmt.Errorf("临时挂载目录不安全，拒绝清理: %s", stageDir)
		}
		children, err := os.ReadDir(stageDir)
		if err != nil {
			return fmt.Errorf("读取临时挂载目录失败: %w", err)
		}
		for _, child := range children {
			index, parseErr := strconv.Atoi(child.Name())
			if parseErr != nil || index < 0 {
				return fmt.Errorf("临时挂载目录包含未知条目，拒绝清理: %s", filepath.Join(stageDir, child.Name()))
			}
			childInfo, statErr := os.Lstat(filepath.Join(stageDir, child.Name()))
			if statErr != nil {
				return fmt.Errorf("读取临时挂载目标失败: %w", statErr)
			}
			if childInfo.Mode()&os.ModeSymlink != 0 || !childInfo.IsDir() && !childInfo.Mode().IsRegular() {
				return fmt.Errorf("临时挂载目标类型无效，拒绝清理: %s", filepath.Join(stageDir, child.Name()))
			}
		}
		stageDirs = append(stageDirs, stageDir)
	}
	mounts, err := mountinfo.GetMounts(mountinfo.PrefixFilter(m.mountsRoot))
	if err != nil {
		return err
	}
	for _, mount := range mounts {
		mountPoint := filepath.Clean(mount.Mountpoint)
		if mountPoint != m.mountsRoot && m.ensureInside(m.mountsRoot, mountPoint) != nil {
			continue
		}
		allowed := mountPoint == m.mountsRoot
		for _, stageDir := range stageDirs {
			allowed = allowed || m.ensureInside(stageDir, mountPoint) == nil
		}
		if !allowed {
			return fmt.Errorf("临时挂载源目录包含未知挂载点，拒绝清理: %s", mountPoint)
		}
	}
	sort.Slice(mounts, func(left, right int) bool {
		return len(mounts[left].Mountpoint) > len(mounts[right].Mountpoint)
	})
	var result error
	for _, mount := range mounts {
		mountPoint := filepath.Clean(mount.Mountpoint)
		if mountPoint != m.mountsRoot && m.ensureInside(m.mountsRoot, mountPoint) != nil {
			continue
		}
		if err := unix.Unmount(mountPoint, unix.MNT_DETACH); err != nil && !errors.Is(err, unix.EINVAL) && !errors.Is(err, unix.ENOENT) {
			result = errors.Join(result, fmt.Errorf("卸载遗留挂载源失败 %s: %w", mountPoint, err))
		}
	}
	if result != nil {
		return result
	}
	remaining, err := mountinfo.GetMounts(mountinfo.PrefixFilter(m.mountsRoot))
	if err != nil {
		return err
	}
	for _, mount := range remaining {
		mountPoint := filepath.Clean(mount.Mountpoint)
		if mountPoint == m.mountsRoot || m.ensureInside(m.mountsRoot, mountPoint) == nil {
			return fmt.Errorf("临时挂载源仍在使用，拒绝清理目录: %s", mountPoint)
		}
	}
	for _, stageDir := range stageDirs {
		if err := os.RemoveAll(stageDir); err != nil {
			result = errors.Join(result, err)
		}
	}
	return result
}

func (m *Manager) stagePlatformMounts(mounts []Mount) ([]string, func() error, error) {
	if len(mounts) == 0 {
		return nil, func() error { return nil }, nil
	}
	unlock, err := m.lockPlatformMountStaging()
	if err != nil {
		return nil, nil, err
	}
	if err := m.cleanupPlatformMountStagingLocked(); err != nil {
		return nil, nil, errors.Join(err, unlock())
	}
	parents, err := mountinfo.GetMounts(mountinfo.ParentsFilter(m.mountsRoot))
	if err != nil {
		return nil, nil, errors.Join(fmt.Errorf("读取临时挂载源传播状态失败: %w", err), unlock())
	}
	var parentMount *mountinfo.Info
	for _, candidate := range parents {
		if m.ensureInside(candidate.Mountpoint, m.mountsRoot) != nil {
			continue
		}
		if parentMount == nil || len(candidate.Mountpoint) > len(parentMount.Mountpoint) ||
			len(candidate.Mountpoint) == len(parentMount.Mountpoint) && candidate.ID > parentMount.ID {
			parentMount = candidate
		}
	}
	if parentMount == nil {
		return nil, nil, errors.Join(errors.New("找不到临时挂载源所在的宿主挂载点"), unlock())
	}
	for _, field := range strings.Fields(parentMount.Optional) {
		if field == "shared" || strings.HasPrefix(field, "shared:") {
			return nil, nil, errors.Join(errors.New("临时挂载源位于 shared mount，请先为服务启用独立或非共享 mount namespace"), unlock())
		}
	}
	if err := unix.Mount(m.mountsRoot, m.mountsRoot, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
		return nil, nil, errors.Join(fmt.Errorf("创建私有挂载源目录失败: %w", err), unlock())
	}
	if err := unix.Mount("", m.mountsRoot, "", unix.MS_PRIVATE|unix.MS_REC, ""); err != nil {
		cleanupErr := m.cleanupPlatformMountStagingLocked()
		return nil, nil, errors.Join(fmt.Errorf("隔离挂载源目录传播失败: %w", err), cleanupErr, unlock())
	}
	var cleanupMu sync.Mutex
	cleaned := false
	ownsLock := true
	cleanup := func() error {
		cleanupMu.Lock()
		defer cleanupMu.Unlock()
		if cleaned {
			return nil
		}
		release := unlock
		if !ownsLock {
			var err error
			release, err = m.lockPlatformMountStaging()
			if err != nil {
				return err
			}
		}
		ownsLock = false
		cleanupErr := m.cleanupPlatformMountStagingLocked()
		unlockErr := release()
		if cleanupErr == nil && unlockErr == nil {
			cleaned = true
		}
		return errors.Join(cleanupErr, unlockErr)
	}
	fail := func(stageErr error) ([]string, func() error, error) {
		cleanupErr := cleanup()
		if cleanupErr != nil {
			cleanupErr = errors.Join(cleanupErr, cleanup())
		}
		return nil, nil, errors.Join(stageErr, cleanupErr)
	}
	stageDir, err := os.MkdirTemp(m.mountsRoot, ".source-")
	if err != nil {
		return fail(err)
	}
	if err := os.Chmod(stageDir, 0700); err != nil {
		return fail(err)
	}
	stageDirFD, err := unix.Open(stageDir, unix.O_PATH|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return fail(err)
	}
	defer unix.Close(stageDirFD)
	staged := make([]string, 0, len(mounts))
	for index, mount := range mounts {
		source, err := m.openPlatformMount(mount.Source)
		if err != nil {
			return fail(err)
		}
		if err := m.scanPlatformMountHandle(mount.Source, source, false, !mount.ReadOnly); err != nil {
			_ = source.Close()
			return fail(err)
		}
		var sourceStat unix.Stat_t
		if err := unix.Fstat(int(source.Fd()), &sourceStat); err != nil {
			_ = source.Close()
			return fail(err)
		}
		kind := sourceStat.Mode & unix.S_IFMT
		if kind != unix.S_IFREG && kind != unix.S_IFDIR {
			_ = source.Close()
			return fail(fmt.Errorf("挂载源只能是普通文件或目录: %s", mount.Source))
		}
		treeFD, err := unix.OpenTree(int(source.Fd()), "", uint(unix.OPEN_TREE_CLONE|unix.OPEN_TREE_CLOEXEC|unix.AT_EMPTY_PATH))
		_ = source.Close()
		if err != nil {
			return fail(fmt.Errorf("固定挂载源失败 %s: %w", mount.Source, err))
		}
		name := fmt.Sprintf("%d", index)
		if kind == unix.S_IFDIR {
			err = unix.Mkdirat(stageDirFD, name, 0700)
		} else {
			var targetFile int
			targetFile, err = unix.Openat(stageDirFD, name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0600)
			if err == nil {
				_ = unix.Close(targetFile)
			}
		}
		if err != nil {
			_ = unix.Close(treeFD)
			return fail(fmt.Errorf("创建临时挂载目标失败: %w", err))
		}
		targetFlags := unix.O_PATH | unix.O_CLOEXEC | unix.O_NOFOLLOW
		if kind == unix.S_IFDIR {
			targetFlags |= unix.O_DIRECTORY
		}
		targetFD, err := unix.Openat(stageDirFD, name, targetFlags, 0)
		if err != nil {
			_ = unix.Close(treeFD)
			return fail(fmt.Errorf("打开临时挂载目标失败: %w", err))
		}
		err = unix.MoveMount(treeFD, "", targetFD, "", unix.MOVE_MOUNT_F_EMPTY_PATH|unix.MOVE_MOUNT_T_EMPTY_PATH)
		_ = unix.Close(targetFD)
		_ = unix.Close(treeFD)
		if err != nil {
			return fail(fmt.Errorf("挂接固定挂载源失败: %w", err))
		}
		target := filepath.Join(stageDir, name)
		if err := unix.Mount("", target, "", unix.MS_PRIVATE|unix.MS_REC, ""); err != nil {
			if !errors.Is(err, unix.EINVAL) {
				return fail(fmt.Errorf("隔离固定挂载源传播失败: %w", err))
			}
			if err := unix.Mount("", target, "", unix.MS_PRIVATE, ""); err != nil {
				return fail(fmt.Errorf("隔离固定挂载源传播失败: %w", err))
			}
		}
		staged = append(staged, target)
	}
	return staged, cleanup, nil
}

func (m *Manager) validatePlatformMountTree(root string) (uint64, error) {
	mounts, err := mountinfo.GetMounts(mountinfo.PrefixFilter(root))
	if err != nil {
		return 0, fmt.Errorf("读取可写挂载目录的挂载点失败: %w", err)
	}
	for _, mount := range mounts {
		mountPoint := filepath.Clean(mount.Mountpoint)
		if mountPoint != root && m.ensureInside(root, mountPoint) == nil {
			return 0, fmt.Errorf("可写挂载目录不能包含其他挂载点: %s", mountPoint)
		}
	}
	info, err := os.Lstat(root)
	if err != nil {
		return 0, err
	}
	device, ok := archiveStatUint(info, "Dev")
	if !ok {
		return 0, errors.New("无法读取可写挂载目录所在设备")
	}
	return device, nil
}
