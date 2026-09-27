package container

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	standardRuntime "runtime"
	fileLog "server/app/util/log"
	"sort"
	"strings"
	"sync"
	"time"
)

// NewManager 创建镜像、实例和运行时状态目录，并恢复已有实例状态。
// options 可选；未提供或字段为零时使用内置默认配置。
func NewManager(root string, options ...ManagerOptions) (_ *Manager, err error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("容器数据目录不能为空")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("解析容器数据目录失败: %w", err)
	}
	rootExists := false
	var anchorPath string
	var anchorInfo os.FileInfo
	if info, statErr := os.Lstat(absRoot); statErr == nil {
		rootExists = true
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, errors.New("容器数据目录不能是符号链接")
		}
		if !info.IsDir() {
			return nil, errors.New("容器数据目录不是目录")
		}
		resolved, resolveErr := filepath.EvalSymlinks(absRoot)
		if resolveErr != nil {
			return nil, fmt.Errorf("解析容器数据目录失败: %w", resolveErr)
		}
		absRoot = filepath.Clean(resolved)
		anchorPath, anchorInfo = absRoot, info
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return nil, fmt.Errorf("读取容器数据目录失败: %w", statErr)
	} else {
		missing := make([]string, 0, 4)
		parent := filepath.Clean(absRoot)
		for {
			info, parentErr := os.Lstat(parent)
			if parentErr == nil {
				if !info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
					return nil, fmt.Errorf("容器数据目录父路径不是目录: %s", parent)
				}
				if info.Mode()&os.ModeSymlink != 0 {
					info, parentErr = os.Stat(parent)
					if parentErr != nil {
						return nil, fmt.Errorf("读取容器数据目录父路径失败: %w", parentErr)
					}
				}
				anchorInfo = info
				break
			}
			if !errors.Is(parentErr, os.ErrNotExist) {
				return nil, fmt.Errorf("读取容器数据目录父路径失败: %w", parentErr)
			}
			next := filepath.Dir(parent)
			if next == parent {
				return nil, errors.New("找不到容器数据目录的有效父路径")
			}
			missing = append(missing, filepath.Base(parent))
			parent = next
		}
		resolved, resolveErr := filepath.EvalSymlinks(parent)
		if resolveErr != nil {
			return nil, fmt.Errorf("解析容器数据目录父路径失败: %w", resolveErr)
		}
		anchorPath = filepath.Clean(resolved)
		for index := len(missing) - 1; index >= 0; index-- {
			resolved = filepath.Join(resolved, missing[index])
		}
		absRoot = filepath.Clean(resolved)
	}
	if absRoot == filepath.Clean(filepath.VolumeName(absRoot)+string(filepath.Separator)) {
		return nil, errors.New("容器数据目录不能是磁盘根目录")
	}
	prepareLinuxDirectories := func(start string) error {
		if standardRuntime.GOOS != "linux" {
			return nil
		}
		if os.Geteuid() != 0 {
			return errors.New("容器运行时必须由 root 启动")
		}
		// 路径解析可能经过尚未收紧的目录，先确认没有被重定向到其他目录。
		info, err := os.Lstat(anchorPath)
		if err != nil || !os.SameFile(anchorInfo, info) {
			return fmt.Errorf("容器数据目录在路径解析期间发生变化: %s", anchorPath)
		}
		// 从上到下收紧权限，先保护父目录，再打开子目录，避免修复期间路径被替换。
		ancestors := make([]string, 0, 8)
		for current := filepath.Clean(start); ; current = filepath.Dir(current) {
			ancestors = append(ancestors, current)
			if current == filepath.Dir(current) {
				break
			}
		}
		for index := len(ancestors) - 1; index >= 0; index-- {
			current := ancestors[index]
			info, err := os.Lstat(current)
			if err != nil {
				return fmt.Errorf("读取容器数据目录祖先失败: %w", err)
			}
			if current == anchorPath && !os.SameFile(anchorInfo, info) {
				return fmt.Errorf("容器数据目录在权限准备期间发生变化: %s", current)
			}
			uid, gid, ownerOK := archiveOwnership(info)
			if !ownerOK || uid != 0 || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("容器数据目录及其祖先必须由 root 持有且不能是符号链接: %s", current)
			}
			mode := info.Mode() &^ 0022
			if current == absRoot {
				if gid != 0 {
					return fmt.Errorf("容器数据目录必须由 root:root 持有: %s", current)
				}
				mode = 0700
			} else if info.Mode().Perm()&0022 == 0 || info.Mode()&os.ModeSticky != 0 {
				// 保留 /tmp 等 sticky 目录的共享用途，其他祖先只移除组和其他用户的写权限。
				continue
			}
			directory, err := os.Open(current)
			if err != nil {
				return fmt.Errorf("打开容器数据目录失败 %s: %w", current, err)
			}
			opened, err := directory.Stat()
			if err != nil || !os.SameFile(info, opened) {
				_ = directory.Close()
				return fmt.Errorf("容器数据目录在权限准备期间发生变化: %s", current)
			}
			// 使用已核对的目录句柄修改权限，不跟随重新解析后的路径。
			if err := errors.Join(directory.Chmod(mode), directory.Close()); err != nil {
				return fmt.Errorf("收紧容器数据目录权限失败 %s: %w", current, err)
			}
		}
		return nil
	}
	ancestor := absRoot
	for {
		if _, err := os.Lstat(ancestor); err == nil {
			break
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("读取容器数据目录祖先失败: %w", err)
		}
		next := filepath.Dir(ancestor)
		if next == ancestor {
			return nil, errors.New("找不到容器数据目录的有效祖先")
		}
		ancestor = next
	}
	if err := prepareLinuxDirectories(ancestor); err != nil {
		return nil, err
	}
	if rootExists {
		entries, readErr := os.ReadDir(absRoot)
		if readErr != nil {
			return nil, fmt.Errorf("读取容器数据目录失败: %w", readErr)
		}
		if len(entries) > 0 {
			for _, name := range []string{"images", "instances", "runtime"} {
				info, statErr := os.Lstat(filepath.Join(absRoot, name))
				if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
					return nil, errors.New("已有非空目录不是可识别的容器数据目录")
				}
			}
		}
	}
	installation := sha256.Sum256([]byte(absRoot))
	managerOptions := ManagerOptions{}
	if len(options) > 0 {
		managerOptions = options[0]
	}
	if managerOptions.UIDMapStart < 0 || managerOptions.GIDMapStart < 0 || managerOptions.IDMapSize < 0 {
		return nil, errors.New("User Namespace 映射参数不能小于 0")
	}
	if managerOptions.UIDMapStart == 0 != (managerOptions.GIDMapStart == 0) {
		return nil, errors.New("UIDMapStart 和 GIDMapStart 必须同时配置")
	}
	requestedSecurity := securityMetadata{
		UIDMapStart: managerOptions.UIDMapStart,
		GIDMapStart: managerOptions.GIDMapStart,
		IDMapSize:   managerOptions.IDMapSize,
	}
	managerOptions = normalizeManagerOptions(managerOptions)
	managerContext, cancel := context.WithCancel(context.Background())
	m := &Manager{
		dataRoot:          absRoot,
		imagesRoot:        filepath.Join(absRoot, "images"),
		instancesRoot:     filepath.Join(absRoot, "instances"),
		runtimeRoot:       filepath.Join(absRoot, "runtime"),
		volumesRoot:       filepath.Join(absRoot, "volumes"),
		mountsRoot:        filepath.Join(absRoot, ".mount-sources"),
		installation:      fmt.Sprintf("%x", installation[:16]),
		options:           managerOptions,
		requestedSecurity: requestedSecurity,
		images:            make(map[string]*Image),
		instances:         make(map[string]*Info),
		runtime:           make(map[string]interface{}),
		terminals:         make(map[TerminalSession]struct{}),
		autoRestart:       make(map[string]uint64),
		ctx:               managerContext,
		cancel:            cancel,
	}
	ready := false
	defer func() {
		if !ready {
			err = errors.Join(err, m.Close())
		}
	}()
	if err := m.ensureManagedDirectory(absRoot); err != nil {
		return nil, err
	}
	if err := prepareLinuxDirectories(absRoot); err != nil {
		return nil, err
	}
	_, mountsRootErr := os.Lstat(m.mountsRoot)
	mountsRootExists := mountsRootErr == nil
	if mountsRootErr != nil && !errors.Is(mountsRootErr, os.ErrNotExist) {
		return nil, fmt.Errorf("读取临时挂载源目录失败: %w", mountsRootErr)
	}
	for _, path := range []string{m.imagesRoot, m.instancesRoot, m.runtimeRoot, m.volumesRoot, m.mountsRoot} {
		if err := m.ensureManagedDirectory(path); err != nil {
			return nil, err
		}
	}
	markerPath := filepath.Join(m.mountsRoot, ".managed")
	marker := []byte(m.installation + "\n")
	validateMarker := func() error {
		info, err := os.Lstat(markerPath)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() != int64(len(marker)) {
			return errors.New("已有临时挂载源目录缺少有效管理标记，拒绝接管")
		}
		data, err := os.ReadFile(markerPath)
		if err != nil || string(data) != string(marker) {
			return errors.New("临时挂载源目录管理标记不匹配，拒绝接管")
		}
		return nil
	}
	if mountsRootExists {
		if err := validateMarker(); err != nil {
			return nil, err
		}
	} else {
		file, err := os.OpenFile(markerPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if errors.Is(err, os.ErrExist) {
			if err := validateMarker(); err != nil {
				return nil, err
			}
			file = nil
			err = nil
		}
		if err != nil {
			return nil, fmt.Errorf("创建临时挂载源管理标记失败: %w", err)
		}
		if file != nil {
			_, writeErr := file.Write(marker)
			syncErr := file.Sync()
			closeErr := file.Close()
			if err := errors.Join(writeErr, syncErr, closeErr); err != nil {
				return nil, fmt.Errorf("保存临时挂载源管理标记失败: %w", err)
			}
		}
	}
	mountRoots := append([]string{m.volumesRoot}, managerOptions.MountRoots...)
	seenMountRoots := make(map[string]struct{}, len(mountRoots))
	for _, mountRoot := range mountRoots {
		mountRoot = strings.TrimSpace(mountRoot)
		if mountRoot == "" {
			return nil, errors.New("挂载根目录不能为空")
		}
		mountRoot, err = filepath.Abs(mountRoot)
		if err != nil {
			return nil, fmt.Errorf("解析挂载根目录失败: %w", err)
		}
		info, statErr := os.Lstat(mountRoot)
		if statErr != nil {
			return nil, fmt.Errorf("读取挂载根目录失败: %w", statErr)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return nil, fmt.Errorf("挂载根目录必须是普通目录: %s", mountRoot)
		}
		mountRoot, err = filepath.EvalSymlinks(mountRoot)
		if err != nil {
			return nil, fmt.Errorf("解析挂载根目录失败: %w", err)
		}
		mountRoot = filepath.Clean(mountRoot)
		if mountRoot == filepath.Clean(filepath.VolumeName(mountRoot)+string(filepath.Separator)) {
			return nil, errors.New("挂载根目录不能是磁盘根目录")
		}
		if mountRoot != m.volumesRoot && (m.ensureInside(m.dataRoot, mountRoot) == nil || m.ensureInside(mountRoot, m.dataRoot) == nil) {
			return nil, errors.New("外部挂载根目录不能与容器数据目录重叠")
		}
		if _, exists := seenMountRoots[mountRoot]; exists {
			continue
		}
		seenMountRoots[mountRoot] = struct{}{}
		m.mountRoots = append(m.mountRoots, mountRoot)
	}
	m.options.MountRoots = append([]string(nil), m.mountRoots...)
	if err := m.configurePlatformSecurity(absRoot); err != nil {
		return nil, err
	}
	if err := m.loadMetadata(); err != nil {
		return nil, err
	}
	if err := m.reconcileRuntimeStates(); err != nil {
		return nil, err
	}
	if err := m.recoverInstances(); err != nil {
		return nil, err
	}
	ready = true
	return m, nil
}

// VolumePath 返回实例在当前管理器下的数据卷目录，不创建目录，也不要求实例已存在。
func (m *Manager) VolumePath(id string) (string, error) {
	if m == nil || m.volumesRoot == "" {
		return "", errors.New("容器管理器尚未初始化")
	}
	if err := m.validateID(id); err != nil {
		return "", err
	}
	return filepath.Join(m.volumesRoot, id), nil
}

// HostID 将容器内 UID/GID 转换为宿主 ID，用于准备可写 bind mount 的所有权。
func (m *Manager) HostID(uid, gid int) (int, int, error) {
	return m.mapOwnership(uid, gid)
}

// Import 将本机 Docker save 归档导入镜像存储，并返回镜像信息。
func (m *Manager) Import(source string) (*Image, error) {
	m.importMu.Lock()
	defer m.importMu.Unlock()
	m.imageMu.Lock()
	defer m.imageMu.Unlock()
	image, err := m.importSave(source)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	m.images[image.ID] = image
	m.mu.Unlock()
	return m.cloneImage(image), nil
}

// Images 返回全部镜像的只读副本。
func (m *Manager) Images() []*Image {
	m.mu.RLock()
	result := make([]*Image, 0, len(m.images))
	for _, image := range m.images {
		result = append(result, m.cloneImage(image))
	}
	m.mu.RUnlock()
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt > result[j].CreatedAt })
	return result
}

// Image 按 ID、名称或标签查询已导入镜像的完整元数据。
func (m *Manager) Image(id string) (*Image, error) {
	m.mu.RLock()
	image, err := m.resolveImage(id)
	if err == nil {
		image = m.cloneImage(image)
	}
	m.mu.RUnlock()
	return image, err
}

// ExportImage 将本地镜像导出为 Docker save 兼容的 tar 归档。
// destination 必须是新的文件路径，方法不会覆盖已有文件。
func (m *Manager) ExportImage(id, destination string) error {
	m.imageMu.RLock()
	defer m.imageMu.RUnlock()
	m.mu.RLock()
	image, err := m.resolveImage(id)
	if err == nil {
		image = m.cloneImage(image)
	}
	m.mu.RUnlock()
	if err != nil {
		return err
	}
	return m.exportImage(image, destination)
}

// Commit 将实例当前文件系统快照保存为一个新镜像。
// 运行中的实例可以直接提交；bind mount 对应的宿主机内容不会写入镜像。
// name 为空时使用实例名称，tags 为空时不附加仓库标签。
func (m *Manager) Commit(id, name string, tags []string) (*Image, error) {
	if err := m.validateID(id); err != nil {
		return nil, err
	}
	unlock := m.lockInstance(id)
	defer unlock()
	instance, err := m.refreshInstanceLocked(id)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, errors.New("实例不存在")
	}
	return m.commitInstance(instance, name, tags)
}

// RemoveImage 删除未被任何实例引用的镜像。
func (m *Manager) RemoveImage(id string) error {
	m.imageMu.Lock()
	defer m.imageMu.Unlock()
	m.mu.Lock()
	image, err := m.resolveImage(id)
	if err != nil {
		m.mu.Unlock()
		return err
	}
	for _, instance := range m.instances {
		if instance.ImageID == image.ID {
			m.mu.Unlock()
			return fmt.Errorf("镜像仍被实例使用: %s", instance.ID)
		}
	}
	digest := strings.TrimPrefix(image.ID, "sha256:")
	rootfsPath := filepath.Join(m.imagesRoot, digest)
	if filepath.Clean(image.Rootfs) != filepath.Clean(rootfsPath) || rootfsPath == filepath.Clean(m.imagesRoot) {
		m.mu.Unlock()
		return errors.New("镜像 rootfs 路径不安全")
	}
	delete(m.images, image.ID)
	m.mu.Unlock()

	if err := os.RemoveAll(rootfsPath); err != nil {
		m.mu.Lock()
		m.images[image.ID] = image
		m.mu.Unlock()
		return fmt.Errorf("删除镜像 rootfs 失败: %w", err)
	}
	metadataPath := filepath.Join(m.imagesRoot, digest+".json")
	if err := m.ensureInside(m.imagesRoot, metadataPath); err != nil {
		return fmt.Errorf("镜像元数据路径不安全: %w", err)
	}
	if err := os.Remove(metadataPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("删除镜像元数据失败: %w", err)
	}
	return nil
}

// Run 自动解析镜像路径、ID 或标签，继承镜像的启动参数并准备可写挂载后启动容器。
// 实例保存固定镜像 ID；已映射的挂载文件保留原属主。
func (m *Manager) Run(options RunOptions) (*Info, error) {
	release, err := m.lockOpen()
	if err != nil {
		return nil, err
	}
	defer release()
	if options.ID == "" {
		options.ID = fmt.Sprintf("container-%d", time.Now().UnixNano())
	}
	if err = m.normalizeRunOptions(&options); err != nil {
		return nil, err
	}
	// 导入需要镜像写锁，在启动流程获取镜像读锁前解析，并固定实例使用的镜像 ID。
	var image *Image
	if filepath.IsAbs(options.ImageID) {
		image, err = m.Import(options.ImageID)
	} else {
		image, err = m.Image(options.ImageID)
	}
	if err != nil {
		return nil, err
	}
	options.ImageID = image.ID
	unlock := m.lockInstance(options.ID)
	defer unlock()
	if !options.CreateOnly {
		m.cancelAutoRestart(options.ID)
	}
	return m.runLocked(options)
}

func (m *Manager) Start(id string) (*Info, error) {
	release, err := m.lockOpen()
	if err != nil {
		return nil, err
	}
	defer release()
	if err := m.validateID(id); err != nil {
		return nil, err
	}
	unlock := m.lockInstance(id)
	defer unlock()
	m.cancelAutoRestart(id)
	instance, err := m.refreshInstanceLocked(id)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, errors.New("实例不存在")
	}
	if m.isActiveStatus(instance.Status) || instance.PID > 0 {
		return nil, errors.New("实例已经在运行")
	}
	options := m.optionsFromInstance(instance)
	if err := m.normalizeRunOptions(&options); err != nil {
		return nil, err
	}
	return m.runLocked(options)
}

// Stop 手动停止实例并暂停自动恢复，超时后由平台运行时强制结束。
func (m *Manager) Stop(id string) error {
	if err := m.validateID(id); err != nil {
		return err
	}
	unlock := m.lockInstance(id)
	defer unlock()
	m.cancelAutoRestart(id)
	m.mu.RLock()
	instance := m.cloneInstance(m.instances[id])
	m.mu.RUnlock()
	if instance == nil {
		return errors.New("实例不存在")
	}
	if !instance.ManualStop {
		instance.ManualStop = true
		if err := m.writeJSON(filepath.Join(m.instancesRoot, id, "instance.json"), instance); err != nil {
			return fmt.Errorf("保存实例手动停止状态失败: %w", err)
		}
		m.storeInstance(instance)
	}
	return m.stopLocked(id)
}

func (m *Manager) Restart(id string) (*Info, error) {
	release, err := m.lockOpen()
	if err != nil {
		return nil, err
	}
	defer release()
	if err := m.validateID(id); err != nil {
		return nil, err
	}
	unlock := m.lockInstance(id)
	defer unlock()
	m.cancelAutoRestart(id)
	instance, err := m.refreshInstanceLocked(id)
	if err != nil {
		return nil, err
	}
	if m.isActiveStatus(instance.Status) || instance.PID > 0 {
		if err := m.stopLocked(id); err != nil {
			return nil, err
		}
	}
	options := m.optionsFromInstance(instance)
	if err := m.normalizeRunOptions(&options); err != nil {
		return nil, err
	}
	return m.runLocked(options)
}

// Remove 停止实例并删除写层、日志和运行状态。
func (m *Manager) Remove(id string) error {
	if err := m.validateID(id); err != nil {
		return err
	}
	unlock := m.lockInstance(id)
	defer unlock()
	m.cancelAutoRestart(id)
	m.mu.RLock()
	instance := m.cloneInstance(m.instances[id])
	_, hasRuntime := m.runtime[id]
	m.mu.RUnlock()
	hasRuntimeState, stateErr := m.hasRuntimeState(id)
	if instance == nil {
		if stateErr != nil {
			return fmt.Errorf("实例元数据不存在，异常运行状态已保留等待人工处理: %w", stateErr)
		}
		if hasRuntimeState {
			if err := m.cleanupOrphanPlatformRuntime(id); err != nil {
				return fmt.Errorf("孤立容器运行状态无法安全清理，状态已保留等待人工处理: %w", err)
			}
			if _, err := os.Lstat(filepath.Join(m.runtimeRoot, id)); !errors.Is(err, os.ErrNotExist) {
				if err == nil {
					err = errors.New("运行状态目录仍然存在")
				}
				return fmt.Errorf("确认孤立容器运行状态删除失败: %w", err)
			}
			return m.cleanupOrphanInstanceStorage(id)
		}
		instanceDir := filepath.Join(m.instancesRoot, id)
		if _, err := os.Lstat(instanceDir); err == nil {
			return m.cleanupOrphanInstanceStorage(id)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("读取孤立实例目录失败: %w", err)
		}
		return errors.New("实例不存在")
	}
	if stateErr != nil {
		return stateErr
	}
	if m.isActiveStatus(instance.Status) || instance.PID > 0 || hasRuntime || hasRuntimeState {
		if err := m.stopLocked(id); err != nil {
			return err
		}
	}
	if err := m.cleanupRuntimeRootfs(instance); err != nil {
		return err
	}
	instanceDir := filepath.Join(m.instancesRoot, id)
	runtimeDir := filepath.Join(m.runtimeRoot, id)
	if err := m.ensureInside(m.instancesRoot, instanceDir); err != nil {
		return fmt.Errorf("实例目录不安全: %w", err)
	}
	if err := m.ensureInside(m.runtimeRoot, runtimeDir); err != nil {
		return fmt.Errorf("运行时目录不安全: %w", err)
	}
	if err := os.RemoveAll(runtimeDir); err != nil {
		return fmt.Errorf("删除实例运行状态失败: %w", err)
	}
	if _, err := os.Lstat(runtimeDir); !errors.Is(err, os.ErrNotExist) {
		if err == nil {
			err = errors.New("运行状态目录仍然存在")
		}
		return fmt.Errorf("确认实例运行状态删除失败: %w", err)
	}
	tombstone := filepath.Join(m.instancesRoot, fmt.Sprintf(".deleted-%s-%d", id, time.Now().UnixNano()))
	if err := m.ensureInside(m.instancesRoot, tombstone); err != nil {
		return fmt.Errorf("实例删除临时目录不安全: %w", err)
	}
	if err := os.Rename(instanceDir, tombstone); err != nil {
		return fmt.Errorf("标记实例删除失败: %w", err)
	}
	m.mu.Lock()
	delete(m.instances, id)
	delete(m.runtime, id)
	m.mu.Unlock()
	if err := os.RemoveAll(tombstone); err != nil {
		return fmt.Errorf("实例已删除，但清理残留目录失败: %w", err)
	}
	return nil
}

// List 同步容器状态并返回摘要，不包含环境变量、命令和挂载等详情。
func (m *Manager) List() []*Instance {
	m.mu.RLock()
	ids := make([]string, 0, len(m.instances))
	for id := range m.instances {
		ids = append(ids, id)
	}
	m.mu.RUnlock()
	for _, id := range ids {
		unlock := m.lockInstance(id)
		_, _ = m.refreshInstanceLocked(id)
		unlock()
	}
	m.mu.RLock()
	result := make([]*Instance, 0, len(m.instances))
	for _, instance := range m.instances {
		item := &Instance{
			ID: instance.ID, Name: instance.Name, ImageID: instance.ImageID,
			Status: instance.Status, CreatedAt: instance.CreatedAt,
			AutoRestart: instance.AutoRestart, ManualStop: instance.ManualStop,
		}
		if image := m.images[instance.ImageID]; image != nil {
			item.ImageName = image.Name
		}
		result = append(result, item)
	}
	m.mu.RUnlock()
	sort.Slice(result, func(i, j int) bool {
		if result[i].CreatedAt == result[j].CreatedAt {
			return result[i].ID < result[j].ID
		}
		return result[i].CreatedAt > result[j].CreatedAt
	})
	return result
}

// Info 返回一个实例的最新状态和完整配置。
func (m *Manager) Info(id string) (*Info, error) {
	if err := m.validateID(id); err != nil {
		return nil, err
	}
	unlock := m.lockInstance(id)
	defer unlock()
	return m.refreshInstanceLocked(id)
}

// Status 返回容器生命周期和资源动态；停止或失败的实例也可查询。
func (m *Manager) Status(id string) (*Status, error) {
	if err := m.validateID(id); err != nil {
		return nil, err
	}
	unlock := m.lockInstance(id)
	defer unlock()
	instance, refreshErr := m.refreshInstanceLocked(id)
	if instance == nil {
		return nil, refreshErr
	}
	result := &Status{ReadTime: time.Now()}
	if refreshErr == nil && instance.Status == StatusRunning {
		resources, err := m.statusPlatform(id)
		if err == nil {
			result = resources
			result.ResourcesAvailable = true
		} else {
			result.ResourceError = err.Error()
			// 进程可能恰好在采样期间退出，重新确认生命周期，不把采样失败当作运行失败。
			if current, err := m.refreshInstanceLocked(id); current != nil {
				instance, refreshErr = current, err
			} else if err != nil {
				instance.Status, refreshErr = StatusUnknown, err
			}
			result.ReadTime = time.Now()
		}
	}
	result.ID, result.Status, result.PID = instance.ID, instance.Status, instance.PID
	result.StartedAt, result.EndedAt = instance.StartedAt, instance.EndedAt
	result.ExitCode, result.Error = instance.ExitCode, instance.Error
	result.AutoRestart, result.ManualStop = instance.AutoRestart, instance.ManualStop
	result.Generation = instance.Generation
	if refreshErr != nil {
		result.Error = refreshErr.Error()
	}
	if instance.StartedAt > 0 {
		end := result.ReadTime.Unix()
		if instance.EndedAt > 0 {
			end = instance.EndedAt
		}
		result.Uptime = max(end-instance.StartedAt, 0)
	}
	m.mu.RLock()
	generation, pending := m.autoRestart[id]
	result.RestartPending = pending && generation == instance.Generation && instance.AutoRestart && !instance.ManualStop && !m.closing
	m.mu.RUnlock()
	return result, nil
}

// StopAll 停止全部容器并关闭管理器，保留各实例的自动启动配置。
func (m *Manager) StopAll() error {
	return m.Close(true)
}

// Close 关闭终端和后台管理，默认保留容器运行，供下一次启动接管。
// 传入 true 时同时停止容器；可重复调用，完成后不再接受运行操作。
func (m *Manager) Close(stopContainers ...bool) error {
	m.closeOnce.Do(func() {
		m.lifecycleMu.Lock()
		m.mu.Lock()
		m.closing = true
		m.autoRestart = make(map[string]uint64)
		terminals := make([]TerminalSession, 0, len(m.terminals))
		for session := range m.terminals {
			terminals = append(terminals, session)
		}
		m.terminals = make(map[TerminalSession]struct{})
		cancel := m.cancel
		m.mu.Unlock()
		if cancel != nil {
			cancel()
		}
		m.lifecycleMu.Unlock()

		m.mu.RLock()
		ids := make([]string, 0, len(m.instances))
		if len(stopContainers) > 0 && stopContainers[0] {
			for id, instance := range m.instances {
				_, hasRuntime := m.runtime[id]
				if m.isActiveStatus(instance.Status) || instance.PID > 0 || hasRuntime {
					ids = append(ids, id)
				}
			}
		}
		m.mu.RUnlock()
		errorsChannel := make(chan error, len(ids)+len(terminals))
		var group sync.WaitGroup
		for _, session := range terminals {
			group.Add(1)
			go func() {
				defer group.Done()
				if err := session.Close(); err != nil {
					errorsChannel <- fmt.Errorf("关闭容器终端失败: %w", err)
				}
			}()
		}
		for _, id := range ids {
			group.Add(1)
			go func(instanceID string) {
				defer group.Done()
				unlock := m.lockInstance(instanceID)
				defer unlock()
				if err := m.stopLocked(instanceID); err != nil {
					errorsChannel <- fmt.Errorf("停止实例 %s 失败: %w", instanceID, err)
				}
			}(id)
		}
		group.Wait()
		close(errorsChannel)
		for err := range errorsChannel {
			m.closeErr = errors.Join(m.closeErr, err)
		}
		m.workers.Wait()
	})
	return m.closeErr
}

// Exec 在运行实例中执行命令，并限制执行时间和输出大小。
func (m *Manager) Exec(id string, options ExecOptions) (*ExecResult, error) {
	if err := m.validateID(id); err != nil {
		return nil, err
	}
	if len(options.Command) == 0 || strings.TrimSpace(options.Command[0]) == "" {
		return nil, errors.New("容器命令不能为空")
	}
	if options.Timeout <= 0 {
		options.Timeout = 30 * time.Second
	}
	if options.MaxOutput <= 0 {
		options.MaxOutput = 512 * 1024
	}
	return m.execPlatform(id, options)
}

// Terminal 在运行实例内启动交互式 shell。关闭会话不会停止容器主进程。
func (m *Manager) Terminal(id string, height, width int) (TerminalSession, error) {
	release, err := m.lockOpen()
	if err != nil {
		return nil, err
	}
	defer release()
	if err := m.validateID(id); err != nil {
		return nil, err
	}
	if height < 1 || height > 1000 || width < 1 || width > 1000 {
		return nil, errors.New("终端尺寸必须在 1-1000 范围内")
	}
	unlock := m.lockInstance(id)
	defer unlock()
	instance, err := m.refreshInstanceLocked(id)
	if err != nil {
		return nil, err
	}
	if instance.Status != StatusRunning {
		return nil, errors.New("实例未运行")
	}
	return m.openPlatformTerminal(id, height, width)
}

// ReadLog 按字节游标读取实例日志。首次读取最新内容，after 读取新增内容，before 读取更早内容。
func (m *Manager) ReadLog(id string, after int64, before int64, pageSize int) (map[string]interface{}, error) {
	if err := m.validateID(id); err != nil {
		return nil, err
	}
	unlock := m.lockInstance(id)
	defer unlock()
	m.mu.RLock()
	instance := m.instances[id]
	if instance == nil {
		m.mu.RUnlock()
		return nil, errors.New("实例不存在")
	}
	path := instance.LogPath
	m.mu.RUnlock()
	unlockLog := m.lockPlatformLog(id)
	defer unlockLog()
	return fileLog.Read(path, after, before, pageSize, path+".1")
}

// ClearLog 清空一个已存在实例的日志。
func (m *Manager) ClearLog(id string) error {
	if err := m.validateID(id); err != nil {
		return err
	}
	unlock := m.lockInstance(id)
	defer unlock()
	m.mu.RLock()
	instance := m.instances[id]
	if instance == nil {
		m.mu.RUnlock()
		return errors.New("实例不存在")
	}
	path := instance.LogPath
	m.mu.RUnlock()
	return m.clearPlatformLog(id, path)
}
