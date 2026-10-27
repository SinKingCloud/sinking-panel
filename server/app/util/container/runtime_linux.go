//go:build linux && cgo

package container

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	standardRuntime "runtime"
	fileLog "server/app/util/log"
	"server/app/util/platform"
	"strings"
	"sync"
	"time"

	pathrs "github.com/cyphar/filepath-securejoin/pathrs-lite"
	"github.com/moby/sys/mountinfo"
	userUtil "github.com/moby/sys/user"
	"github.com/opencontainers/cgroups"
	_ "github.com/opencontainers/cgroups/devices"
	devices "github.com/opencontainers/cgroups/devices/config"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"
	_ "github.com/opencontainers/runc/libcontainer/nsenter"
	"golang.org/x/sys/unix"
)

type linuxRuntime struct {
	container *libcontainer.Container
	process   *libcontainer.Process
	logMu     sync.Mutex
	logFile   *os.File
	logPath   string
	cleanup   sync.Once
	exitCode  int
	exitErr   error
	failure   error
}

func (m *Manager) startPlatformRuntime(instance *Info, image *Image, options RunOptions) (returnErr error) {
	if image.OS != "" && image.OS != "linux" {
		return fmt.Errorf("镜像操作系统不支持: %s", image.OS)
	}
	if architecture := m.normalizeArchitecture(image.Architecture); architecture != "" && architecture != standardRuntime.GOARCH {
		return fmt.Errorf("镜像架构 %s 与宿主机 %s 不匹配", image.Architecture, standardRuntime.GOARCH)
	}
	if image.Rootfs == "" {
		return errors.New("镜像 rootfs 为空")
	}
	if _, err := os.Stat(image.Rootfs); err != nil {
		return fmt.Errorf("镜像 rootfs 不存在: %w", err)
	}
	// cgroups 会缓存根目录句柄，必须在宿主 namespace 中初始化，供后续启停复用。
	// cgroup v1 根目录没有 cgroup.procs，读取失败不影响句柄初始化。
	_, _ = cgroups.ReadFile("/sys/fs/cgroup", "cgroup.procs")
	standardRuntime.LockOSThread()
	originalNamespace, err := os.Open("/proc/thread-self/ns/mnt")
	if err != nil {
		standardRuntime.UnlockOSThread()
		return fmt.Errorf("打开当前 mount namespace 失败: %w", err)
	}
	originalDirectory, err := os.Open(".")
	if err != nil {
		_ = originalNamespace.Close()
		standardRuntime.UnlockOSThread()
		return fmt.Errorf("打开当前工作目录失败: %w", err)
	}
	if err := unix.Unshare(unix.CLONE_NEWNS); err != nil {
		_ = originalDirectory.Close()
		_ = originalNamespace.Close()
		standardRuntime.UnlockOSThread()
		return fmt.Errorf("创建容器 mount namespace 失败: %w", err)
	}
	namespaceActive := true
	restoreNamespace := func() error {
		if !namespaceActive {
			return nil
		}
		if err := unix.Setns(int(originalNamespace.Fd()), unix.CLONE_NEWNS); err != nil {
			return fmt.Errorf("恢复 mount namespace 失败: %w", err)
		}
		// setns 会把当前线程的工作目录重置为根目录，必须还原后才能归还 Go 线程池。
		// 否则协程切换线程时，相对上传、日志等路径会随机指向另一目录。
		if err := unix.Fchdir(int(originalDirectory.Fd())); err != nil {
			return fmt.Errorf("恢复工作目录失败: %w", err)
		}
		namespaceActive = false
		return nil
	}
	defer func() {
		err := restoreNamespace()
		_ = originalDirectory.Close()
		_ = originalNamespace.Close()
		if err != nil {
			// 恢复失败的线程不能归还 Go 线程池，协程退出时由 runtime 回收。
			returnErr = errors.Join(returnErr, err)
			return
		}
		standardRuntime.UnlockOSThread()
	}()
	if err := unix.Mount("", "/", "", unix.MS_PRIVATE|unix.MS_REC, ""); err != nil {
		return fmt.Errorf("隔离挂载传播失败: %w", err)
	}
	rootfs := instance.Rootfs
	if rootfs == "" {
		rootfs = image.Rootfs
	}
	for _, mount := range options.Mounts {
		if err := m.validateMount(mount); err != nil {
			return err
		}
	}
	user := options.User
	if user == "" {
		user = image.User
	}
	uid, gid, groups, err := m.resolveUser(rootfs, user)
	if err != nil {
		return fmt.Errorf("解析容器运行用户失败: %w", err)
	}
	hostUID, hostGID, err := m.mapOwnership(uid, gid)
	if err != nil {
		return fmt.Errorf("容器运行用户超出 User Namespace 映射范围: %w", err)
	}
	for _, group := range groups {
		if _, _, err := m.mapOwnership(0, group); err != nil {
			return fmt.Errorf("镜像附加用户组超出 User Namespace 映射范围: %w", err)
		}
	}
	stagedSources, cleanupStaging, err := m.stagePlatformMounts(options.Mounts, [2]int{hostUID, hostGID})
	if err != nil {
		return err
	}
	cleanupPending := true
	defer func() {
		if cleanupPending {
			returnErr = errors.Join(returnErr, cleanupStaging())
		}
	}()
	runtimeMounts := append([]Mount(nil), options.Mounts...)
	for index, source := range stagedSources {
		runtimeMounts[index].Source = source
	}
	if err := m.prepareMountTargets(rootfs, runtimeMounts, instance.WritableLayer); err != nil {
		return err
	}
	// 容器 root 映射到普通宿主 UID，无法遍历受管目录的 0700 父路径。
	// 只在私有 namespace 中提供可访问的 rootfs 入口，不放宽真实目录权限。
	// 固定使用 /tmp，避免 TMPDIR 本身位于 /root 等私有目录下。
	initRootfs, err := os.MkdirTemp("/tmp", "sinking-rootfs-")
	if err != nil {
		return fmt.Errorf("创建容器 rootfs 入口失败: %w", err)
	}
	rootfsMounted, rootfsRemoved := false, false
	cleanupRootfs := func() error {
		if rootfsMounted {
			if err := unix.Unmount(initRootfs, unix.MNT_DETACH); err != nil {
				return fmt.Errorf("卸载容器 rootfs 入口失败: %w", err)
			}
			rootfsMounted = false
		}
		if !rootfsRemoved {
			if err := os.Remove(initRootfs); err != nil {
				return fmt.Errorf("清理容器 rootfs 入口失败: %w", err)
			}
			rootfsRemoved = true
		}
		return nil
	}
	defer func() { returnErr = errors.Join(returnErr, cleanupRootfs()) }()
	containerRootfs := initRootfs
	if instance.WritableLayer {
		if err := unix.Mount(rootfs, initRootfs, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
			return fmt.Errorf("挂载容器 rootfs 入口失败: %w", err)
		}
		rootfsMounted = true
	} else {
		// 只读实例也需要准备 /etc、/proc 等挂载目标，不能写入共享镜像。
		// 临时层随容器命名空间释放，父线程仅卸载入口，不删除仍在使用的 upper/work。
		if err := unix.Mount("tmpfs", initRootfs, "tmpfs", unix.MS_NOSUID|unix.MS_NODEV, "mode=755,size=16m"); err != nil {
			return fmt.Errorf("创建只读容器准备层失败: %w", err)
		}
		rootfsMounted = true
		for _, name := range []string{"upper", "work", "rootfs"} {
			dir := filepath.Join(initRootfs, name)
			if err := os.Mkdir(dir, 0755); err != nil {
				return err
			}
			if err := os.Chown(dir, int(m.security.UIDMapStart), int(m.security.GIDMapStart)); err != nil {
				return err
			}
			if err := os.Chmod(dir, 0755); err != nil {
				return err
			}
		}
		containerRootfs = filepath.Join(initRootfs, "rootfs")
		data := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", rootfs, filepath.Join(initRootfs, "upper"), filepath.Join(initRootfs, "work"))
		if err := unix.Mount("overlay", containerRootfs, "overlay", 0, data); err != nil {
			return fmt.Errorf("挂载只读容器准备层失败: %w", err)
		}
	}
	config, err := m.buildConfig(instance, image, options, true)
	if err != nil {
		return err
	}
	for index, source := range stagedSources {
		config.Mounts[len(config.Mounts)-len(stagedSources)+index].Source = source
	}
	config.Rootfs = containerRootfs
	automaticMountCount := len(config.Mounts) - len(options.Mounts)
	var namedTimezoneMounts []*configs.Mount
	for _, mount := range config.Mounts[:automaticMountCount] {
		if mount.Destination != containerTimezonePath {
			continue
		}
		// 保存本次启动的 TZif 数据，避免 UID 映射受宿主文件权限影响。
		// 宿主未配置 localtime 时按 Linux 的默认行为使用 UTC。
		var timezone []byte
		info, err := os.Stat(mount.Source)
		if os.IsNotExist(err) {
			timezone = make([]byte, 54)
			copy(timezone, "TZif")
			timezone[39], timezone[43] = 1, 4 // 一个 UTC 类型及 "UTC\x00" 缩写，无切换记录
			copy(timezone[50:], "UTC")
		} else if err != nil {
			return fmt.Errorf("读取服务器时区失败: %w", err)
		} else {
			if !info.Mode().IsRegular() {
				return errors.New("服务器时区文件不是普通文件")
			}
			timezone, err = os.ReadFile(mount.Source)
			if err != nil {
				return fmt.Errorf("读取服务器时区失败: %w", err)
			}
		}
		if _, err := time.LoadLocationFromTZData("localtime", timezone); err != nil {
			return fmt.Errorf("服务器时区数据无效: %w", err)
		}
		file, err := os.CreateTemp("/tmp", "sinking-timezone-")
		if err != nil {
			return err
		}
		defer os.Remove(file.Name())
		_, writeErr := file.Write(timezone)
		if err := errors.Join(writeErr, file.Chmod(0444), file.Close()); err != nil {
			return fmt.Errorf("准备容器时区数据失败: %w", err)
		}
		mount.Source = file.Name()
		destination, err := m.prepareTimezoneMount(containerRootfs, containerTimezonePath)
		if err != nil {
			return err
		}
		mount.Destination = destination
		config.Labels = append(config.Labels, timezoneMountLabel+destination)
		// SQL Server PAL 和 JVM 需要时区名称；libc 同时需要对应的 TZif 文件。
		// 使用同一份宿主数据，兼容镜像没有 tzdata 或镜像内规则版本不同的情况。
		if name := m.hostTimezoneName(timezone); name != "" {
			destination, err := m.prepareTimezoneMount(containerRootfs, "/usr/share/zoneinfo/"+name)
			if err != nil {
				return err
			}
			namedTimezoneMounts = append(namedTimezoneMounts, &configs.Mount{
				Source:      file.Name(),
				Destination: destination,
				Device:      "bind",
				Flags:       mount.Flags,
			})
			config.Labels = append(config.Labels, timezoneNameLabel+name, timezoneMountLabel+destination)
		}
	}
	// 自动挂载放在用户挂载前，保留用户自定义 zoneinfo 挂载的覆盖顺序。
	if len(namedTimezoneMounts) > 0 {
		mounts := append(config.Mounts[:automaticMountCount:automaticMountCount], namedTimezoneMounts...)
		config.Mounts = append(mounts, config.Mounts[automaticMountCount:]...)
	}
	container, err := libcontainer.Create(m.runtimeRoot, instance.ID, config)
	if err != nil {
		return fmt.Errorf("创建容器失败: %w", err)
	}
	logFile, err := m.openRuntimeLog(instance.LogPath)
	if err != nil {
		failure := fmt.Errorf("打开容器日志失败: %w", err)
		return errors.Join(failure, m.abortPlatformStart(instance, container, nil, nil, failure))
	}
	args := append([]string(nil), image.Entrypoint...)
	if len(options.Entrypoint) > 0 {
		args = append([]string(nil), options.Entrypoint...)
	}
	command := options.Command
	if len(command) == 0 {
		command = image.Command
	}
	args = append(args, command...)
	if len(args) == 0 {
		failure := errors.New("镜像没有入口命令")
		return errors.Join(failure, m.abortPlatformStart(instance, container, nil, logFile, failure))
	}
	env := m.mergeEnv(
		[]string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
		image.Env,
		m.timezoneEnv(*config),
		options.Env,
	)
	workingDir := options.WorkingDir
	if workingDir == "" {
		workingDir = image.WorkingDir
	}
	if workingDir == "" {
		workingDir = "/"
	}
	process := &libcontainer.Process{
		Args:             args,
		Env:              env,
		UID:              uid,
		GID:              gid,
		AdditionalGroups: groups,
		Capabilities:     m.processCapabilities(uid),
		Cwd:              workingDir,
		Stdout:           logFile,
		Stderr:           logFile,
		Init:             true,
	}
	flushInitOutput, err := m.prepareInitOutput(process)
	if err != nil {
		return errors.Join(err, m.abortPlatformStart(instance, container, nil, logFile, err))
	}
	startErr := container.Start(process)
	if err := errors.Join(startErr, flushInitOutput()); err != nil {
		failure := fmt.Errorf("启动容器进程失败: %w", err)
		return errors.Join(failure, m.abortPlatformStart(instance, container, process, logFile, failure))
	}
	// init 已进入容器、尚未执行应用；恢复并保存真实路径，供 Exec/终端和恢复使用。
	runningConfig := container.Config()
	runningConfig.Rootfs = rootfs
	if err := container.Set(runningConfig); err != nil {
		failure := fmt.Errorf("保存容器运行配置失败: %w", err)
		return errors.Join(failure, m.abortPlatformStart(instance, container, process, logFile, failure))
	}
	if err := container.Exec(); err != nil {
		failure := fmt.Errorf("执行容器入口失败: %w", err)
		return errors.Join(failure, m.abortPlatformStart(instance, container, process, logFile, failure))
	}
	cleanupErr := cleanupStaging()
	cleanupPending = cleanupErr != nil
	if cleanupErr != nil {
		failure := fmt.Errorf("清理临时挂载源失败: %w", cleanupErr)
		return errors.Join(failure, m.abortPlatformStart(instance, container, process, logFile, failure))
	}
	if err := cleanupRootfs(); err != nil {
		return errors.Join(err, m.abortPlatformStart(instance, container, process, logFile, err))
	}
	if err := restoreNamespace(); err != nil {
		return errors.Join(err, m.abortPlatformStart(instance, container, process, logFile, err))
	}
	pid, err := process.Pid()
	if err != nil || pid <= 0 {
		_ = process.Signal(unix.SIGKILL)
		_, _ = process.Wait()
		if err == nil {
			err = errors.New("容器进程 PID 无效")
		}
		failure := fmt.Errorf("读取容器进程 PID 失败: %w", err)
		return errors.Join(failure, m.abortPlatformStart(instance, container, process, logFile, failure))
	}
	instance.PID = pid
	runtime := &linuxRuntime{
		container: container,
		process:   process,
		logFile:   logFile,
		logPath:   instance.LogPath,
	}
	m.mu.Lock()
	m.runtime[instance.ID] = runtime
	m.mu.Unlock()
	m.runBackground(func(ctx context.Context) {
		m.watchRuntimeLog(ctx, instance.ID, instance.Generation, runtime)
	})
	m.runBackground(func(ctx context.Context) {
		id, generation := instance.ID, instance.Generation
		// 等待子进程不绑定管理器生命周期；面板退出时容器继续运行。
		// 单独回收 Wait，避免同进程重新创建 Manager 后留下僵尸子进程。
		done := make(chan struct{})
		var state *os.ProcessState
		var waitErr error
		go func() {
			state, waitErr = runtime.process.Wait()
			close(done)
		}()
		select {
		case <-ctx.Done():
			return
		case <-done:
		}
		if ctx.Err() != nil {
			return
		}
		exitCode := -1
		if state != nil {
			exitCode = state.ExitCode()
		}
		m.finishPlatformRuntime(id, generation, runtime, exitCode, waitErr)
	})
	return nil
}

// preparePlatformRootfs 以镜像为只读 lower 层，为实例准备独立 overlay 写层。
func (m *Manager) preparePlatformRootfs(instance *Info, image *Image, writable bool) error {
	if !writable {
		instance.Rootfs = image.Rootfs
		instance.WritableLayer = false
		return nil
	}
	instanceDir := filepath.Join(m.instancesRoot, instance.ID)
	upperDir := filepath.Join(instanceDir, "upper")
	workDir := filepath.Join(instanceDir, "work")
	mergedDir := filepath.Join(instanceDir, "rootfs")
	for _, dir := range []string{upperDir, workDir, mergedDir} {
		if err := m.ensureManagedDirectory(dir); err != nil {
			return fmt.Errorf("创建实例可写层失败: %w", err)
		}
		if err := m.normalizePlatformPathOwnership(dir, dir); err != nil {
			return fmt.Errorf("映射实例可写层所有者失败: %w", err)
		}
	}
	// 显式修复宿主机 umask，避免非 root 容器用户无法进入 merged rootfs。
	if err := os.Chmod(upperDir, 0755); err != nil {
		return fmt.Errorf("设置可写层目录权限失败: %w", err)
	}
	if err := os.Chmod(workDir, 0755); err != nil {
		return fmt.Errorf("设置 overlay 工作目录权限失败: %w", err)
	}
	if err := os.Chmod(mergedDir, 0755); err != nil {
		return fmt.Errorf("设置实例 rootfs 权限失败: %w", err)
	}
	data := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", image.Rootfs, upperDir, workDir)
	if err := unix.Mount("overlay", mergedDir, "overlay", 0, data); err != nil {
		return fmt.Errorf("创建实例可写层失败，宿主机可能不支持 overlayfs: %w", err)
	}
	if err := os.Chmod(mergedDir, 0755); err != nil {
		_ = unix.Unmount(mergedDir, unix.MNT_DETACH)
		return fmt.Errorf("设置实例 rootfs 权限失败: %w", err)
	}
	instance.Rootfs = mergedDir
	instance.UpperDir = upperDir
	instance.WorkDir = workDir
	instance.WritableLayer = true
	return nil
}

// cleanupPlatformRootfs 只卸载 merged 目录，保留 upper 层供下次启动使用。
func (m *Manager) cleanupPlatformRootfs(instance *Info) error {
	if !instance.WritableLayer || instance.Rootfs == "" || instance.UpperDir == "" {
		return nil
	}
	mounted, err := mountinfo.Mounted(instance.Rootfs)
	if err != nil {
		return fmt.Errorf("检查实例 rootfs 挂载状态失败: %w", err)
	}
	if !mounted {
		return nil
	}
	if err := unix.Unmount(instance.Rootfs, unix.MNT_DETACH); err != nil && !errors.Is(err, unix.ENOENT) && !errors.Is(err, unix.EINVAL) {
		return fmt.Errorf("卸载实例 rootfs 失败: %w", err)
	}
	return nil
}

// cleanupOrphanPlatformRuntime 只通过可验证的 libcontainer 状态停止并销毁孤立容器。
func (m *Manager) cleanupOrphanPlatformRuntime(id string) error {
	container, err := libcontainer.Load(m.runtimeRoot, id)
	if err != nil {
		return fmt.Errorf("加载容器运行状态失败: %w", err)
	}
	status, err := container.Status()
	if err != nil {
		return fmt.Errorf("读取容器运行状态失败: %w", err)
	}
	if status == libcontainer.Paused {
		if err := container.Resume(); err != nil {
			return fmt.Errorf("恢复孤立容器失败: %w", err)
		}
		status = libcontainer.Running
	}
	if status == libcontainer.Running {
		if err := container.Signal(unix.SIGTERM); err != nil && !errors.Is(err, libcontainer.ErrNotRunning) {
			return fmt.Errorf("停止孤立容器失败: %w", err)
		}
		if !m.waitContainerStopped(container, m.managerOptions().StopGracePeriod) {
			if err := container.Signal(unix.SIGKILL); err != nil && !errors.Is(err, libcontainer.ErrNotRunning) {
				return fmt.Errorf("强制停止孤立容器失败: %w", err)
			}
			if !m.waitContainerStopped(container, m.managerOptions().ForceStopPeriod) {
				return errors.New("等待孤立容器停止超时")
			}
		}
	} else if status == libcontainer.Created {
		if err := container.Signal(unix.SIGKILL); err != nil && !errors.Is(err, libcontainer.ErrNotRunning) {
			return fmt.Errorf("清理未完成启动的孤立容器失败: %w", err)
		}
		if !m.waitContainerStopped(container, m.managerOptions().ForceStopPeriod) {
			return errors.New("等待未完成启动的孤立容器退出超时")
		}
	}
	if err := container.Destroy(); err != nil {
		if processErr := m.killPlatformProcesses(container); processErr != nil {
			return errors.Join(fmt.Errorf("销毁孤立容器失败: %w", err), processErr)
		}
		if err := container.Destroy(); err != nil {
			return fmt.Errorf("销毁孤立容器失败: %w", err)
		}
	}
	return nil
}

// recoverPlatformRuntime 从 libcontainer 状态目录恢复仍在运行的实例。
// libcontainer 会校验进程启动时间，因此不会把复用的 PID 当成原容器。
func (m *Manager) recoverPlatformRuntime(instance *Info) (bool, error) {
	container, err := libcontainer.Load(m.runtimeRoot, instance.ID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if instance.Status == StatusStarting || instance.Status == StatusStopping {
				return false, nil
			}
			if m.isActiveStatus(instance.Status) || instance.PID > 0 {
				return false, errors.New("容器运行状态目录不存在，无法安全确认进程是否退出")
			}
			return false, nil
		}
		return false, fmt.Errorf("加载容器运行状态失败: %w", err)
	}
	status, err := container.Status()
	if err != nil {
		return false, fmt.Errorf("读取容器运行状态失败: %w", err)
	}
	runtime := &linuxRuntime{
		container: container,
		logPath:   instance.LogPath,
		exitCode:  instance.ExitCode,
	}
	if instance.Status == StatusFailed && instance.Error != "" {
		runtime.failure = errors.New(instance.Error)
	}
	if status == libcontainer.Created {
		runtime.exitCode = -1
		runtime.exitErr = errors.New("容器启动未完成")
		if runtime.failure == nil {
			runtime.failure = runtime.exitErr
		}
		if err := container.Signal(unix.SIGKILL); err != nil && !errors.Is(err, libcontainer.ErrNotRunning) {
			return false, fmt.Errorf("清理未完成启动的容器失败: %w", err)
		}
		if !m.waitContainerStopped(container, m.managerOptions().ForceStopPeriod) {
			return false, errors.New("等待未完成启动的容器退出超时")
		}
		m.mu.Lock()
		m.runtime[instance.ID] = runtime
		m.mu.Unlock()
		if err := m.markRuntimeCleanupLocked(instance.ID, instance.Generation, runtime.exitCode, runtime.exitErr); err != nil {
			m.retryPlatformCleanup(instance.ID, instance.Generation, runtime)
			return false, err
		}
		if err := m.releasePlatformRuntime(instance.ID, runtime, container); err != nil {
			return false, err
		}
		return false, nil
	}
	if status != libcontainer.Running && status != libcontainer.Paused {
		m.mu.Lock()
		m.runtime[instance.ID] = runtime
		m.mu.Unlock()
		if err := m.markRuntimeCleanupLocked(instance.ID, instance.Generation, instance.ExitCode, nil); err != nil {
			m.retryPlatformCleanup(instance.ID, instance.Generation, runtime)
			return false, err
		}
		if err := m.releasePlatformRuntime(instance.ID, runtime, container); err != nil {
			return false, err
		}
		return false, nil
	}
	state, err := container.State()
	if err != nil {
		return false, fmt.Errorf("读取容器进程状态失败: %w", err)
	}
	if state.ID != instance.ID || state.InitProcessPid <= 0 {
		return false, errors.New("容器运行状态无效")
	}
	instance.PID = state.InitProcessPid
	m.mu.Lock()
	m.runtime[instance.ID] = runtime
	m.mu.Unlock()
	m.runBackground(func(ctx context.Context) {
		m.watchRuntimeLog(ctx, instance.ID, instance.Generation, runtime)
	})
	m.runBackground(func(ctx context.Context) {
		m.watchRecoveredRuntime(ctx, instance.ID, instance.Generation, runtime)
	})
	return true, nil
}

// refreshPlatformRuntime 根据 libcontainer 状态刷新实例 PID。
func (m *Manager) refreshPlatformRuntime(instance *Info) (bool, error) {
	m.mu.RLock()
	runtime, ok := m.runtime[instance.ID].(*linuxRuntime)
	m.mu.RUnlock()
	if !ok || runtime == nil || runtime.container == nil {
		return m.recoverPlatformRuntime(instance)
	}
	status, err := runtime.container.Status()
	if err != nil {
		return false, fmt.Errorf("读取容器运行状态失败: %w", err)
	}
	if status != libcontainer.Running && status != libcontainer.Created && status != libcontainer.Paused {
		if err := m.markRuntimeCleanupLocked(instance.ID, instance.Generation, instance.ExitCode, nil); err != nil {
			m.retryPlatformCleanup(instance.ID, instance.Generation, runtime)
			return false, err
		}
		if err := m.releasePlatformRuntime(instance.ID, runtime, runtime.container); err != nil {
			return false, err
		}
		return false, nil
	}
	if status == libcontainer.Created {
		return false, errors.New("容器启动状态未完成")
	}
	state, err := runtime.container.State()
	if err != nil {
		return false, fmt.Errorf("读取容器进程状态失败: %w", err)
	}
	instance.PID = state.InitProcessPid
	return instance.PID > 0, nil
}

func (m *Manager) watchRecoveredRuntime(ctx context.Context, id string, generation uint64, runtime *linuxRuntime) {
	ticker := time.NewTicker(m.managerOptions().RuntimePollInterval)
	defer ticker.Stop()
	failures := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if ctx.Err() != nil {
			return
		}
		m.mu.RLock()
		instance := m.instances[id]
		current, _ := m.runtime[id].(*linuxRuntime)
		valid := instance != nil && instance.Generation == generation && current == runtime
		m.mu.RUnlock()
		if !valid {
			return
		}
		status, err := runtime.container.Status()
		if err != nil {
			failures++
			if failures == m.managerOptions().RuntimeFailureLimit {
				m.markPlatformRuntimeUnknown(id, generation, runtime, err)
			}
			continue
		}
		failures = 0
		if status == libcontainer.Stopped {
			m.finishPlatformRuntime(id, generation, runtime, 0, nil)
			return
		}
	}
}

func (m *Manager) finishPlatformRuntime(id string, generation uint64, runtime *linuxRuntime, exitCode int, waitErr error) {
	unlock := m.lockInstance(id)
	defer unlock()
	m.mu.RLock()
	instance := m.instances[id]
	current, _ := m.runtime[id].(*linuxRuntime)
	valid := instance != nil && instance.Generation == generation && current == runtime
	m.mu.RUnlock()
	if !valid {
		return
	}
	runtime.exitCode = exitCode
	runtime.exitErr = waitErr
	if err := m.markRuntimeCleanupLocked(id, generation, exitCode, waitErr); err != nil {
		m.retryPlatformCleanup(id, generation, runtime)
		return
	}
	releaseErr := m.releasePlatformRuntime(id, runtime, runtime.container)
	m.markExitedLocked(id, generation, exitCode, errors.Join(waitErr, releaseErr), releaseErr == nil, runtime.failure)
	if releaseErr == nil && runtime.failure == nil {
		m.scheduleAutoRestart(id, generation)
	}
}

func (m *Manager) markPlatformRuntimeUnknown(id string, generation uint64, runtime *linuxRuntime, statusErr error) {
	unlock := m.lockInstance(id)
	defer unlock()
	m.mu.RLock()
	instance := m.cloneInstance(m.instances[id])
	current, _ := m.runtime[id].(*linuxRuntime)
	m.mu.RUnlock()
	if instance == nil || instance.Generation != generation || current != runtime {
		return
	}
	instance.Status = StatusUnknown
	instance.Error = "读取容器运行状态失败: " + statusErr.Error()
	m.storeInstance(instance)
	_ = m.writeJSON(filepath.Join(m.instancesRoot, id, "instance.json"), instance)
}

func (m *Manager) watchRuntimeLog(ctx context.Context, id string, generation uint64, runtime *linuxRuntime) {
	defer func() {
		// 容器持有独立的文件描述符，关闭面板这端不影响后续日志写入。
		runtime.logMu.Lock()
		if runtime.logFile != nil {
			_ = runtime.logFile.Close()
			runtime.logFile = nil
		}
		runtime.logMu.Unlock()
	}()
	ticker := time.NewTicker(m.managerOptions().RuntimeLogInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if ctx.Err() != nil {
			return
		}
		m.mu.RLock()
		instance := m.instances[id]
		current, _ := m.runtime[id].(*linuxRuntime)
		valid := instance != nil && instance.Generation == generation && current == runtime
		m.mu.RUnlock()
		if !valid {
			return
		}
		_ = m.rotatePlatformLog(id, generation, runtime)
	}
}

func (m *Manager) rotatePlatformLog(id string, generation uint64, runtime *linuxRuntime) error {
	runtime.logMu.Lock()
	defer runtime.logMu.Unlock()
	m.mu.RLock()
	instance := m.instances[id]
	current, _ := m.runtime[id].(*linuxRuntime)
	valid := instance != nil && instance.Generation == generation && current == runtime
	m.mu.RUnlock()
	if !valid || runtime.logPath == "" {
		return nil
	}
	return m.rotateLogFile(runtime.logPath)
}

func (m *Manager) rotateLogFile(path string) error {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return err
	}
	maxSize := m.managerOptions().ContainerLogMaxSize
	if info.Size() < maxSize {
		return file.Close()
	}
	if _, err := file.Seek(-maxSize, io.SeekEnd); err != nil {
		_ = file.Close()
		return err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, maxSize))
	closeErr := file.Close()
	if readErr != nil || closeErr != nil {
		return errors.Join(readErr, closeErr)
	}
	temporary := path + ".1.tmp"
	if err := os.WriteFile(temporary, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(temporary, path+".1"); err != nil {
		_ = os.Remove(temporary)
		return err
	}
	return os.Truncate(path, 0)
}

func (m *Manager) statusPlatform(id string) (*Status, error) {
	m.mu.RLock()
	runtime, _ := m.runtime[id].(*linuxRuntime)
	m.mu.RUnlock()
	var container *libcontainer.Container
	if runtime != nil && runtime.container != nil {
		container = runtime.container
	} else {
		loaded, err := libcontainer.Load(m.runtimeRoot, id)
		if err != nil {
			return nil, fmt.Errorf("加载容器运行状态失败: %w", err)
		}
		container = loaded
	}
	if container == nil {
		return nil, errors.New("实例未运行")
	}
	data, err := container.Stats()
	if err != nil {
		return nil, fmt.Errorf("读取容器资源统计失败: %w", err)
	}
	if data == nil || data.CgroupStats == nil {
		return nil, errors.New("容器未提供 cgroup 资源统计")
	}
	stats := data.CgroupStats
	memory := stats.MemoryStats
	result := &Status{
		ReadTime:            time.Now(),
		CPUUsage:            stats.CpuStats.CpuUsage.TotalUsage,
		CPUUserUsage:        stats.CpuStats.CpuUsage.UsageInUsermode,
		CPUSystemUsage:      stats.CpuStats.CpuUsage.UsageInKernelmode,
		CPUPeriods:          stats.CpuStats.ThrottlingData.Periods,
		CPUThrottledPeriods: stats.CpuStats.ThrottlingData.ThrottledPeriods,
		CPUThrottledTime:    stats.CpuStats.ThrottlingData.ThrottledTime,
		MemoryUsage:         memory.Usage.Usage,
		MemoryLimit:         memory.Usage.Limit,
		MemoryPeak:          memory.Usage.MaxUsage,
		MemoryCache:         memory.Cache,
		Pids:                stats.PidsStats.Current,
		PidsLimit:           stats.PidsStats.Limit,
	}
	// v1 的无限内存上限是按页对齐的 MaxInt64，v2 则使用 MaxUint64。
	if result.MemoryLimit >= uint64(1<<63)-uint64(os.Getpagesize()) {
		result.MemoryLimit = 0
	}
	// v1 优先使用包含子 cgroup 的 total_*，v2 对应 file/anon/inactive_file。
	if value, ok := memory.Stats["total_cache"]; ok {
		result.MemoryCache = value
	}
	if value, ok := memory.Stats["total_rss"]; ok {
		result.MemoryRSS = value
	} else if value, ok := memory.Stats["rss"]; ok {
		result.MemoryRSS = value
	} else {
		result.MemoryRSS = memory.Stats["anon"]
	}
	inactiveFile, ok := memory.Stats["total_inactive_file"]
	if !ok {
		inactiveFile = memory.Stats["inactive_file"]
	}
	result.MemoryWorkingSet = result.MemoryUsage - min(inactiveFile, result.MemoryUsage)
	// runc 的 SwapUsage 同时包含内存和 Swap；未启用 Swap 统计时避免无符号下溢。
	if memory.SwapUsage.Usage >= result.MemoryUsage {
		result.SwapUsage = memory.SwapUsage.Usage - result.MemoryUsage
	}
	// 同一数组还可能包含 Total、Sync、Async，只累加 Read/Write 避免重复统计。
	for _, entry := range stats.BlkioStats.IoServiceBytesRecursive {
		switch entry.Op {
		case "Read":
			result.IOReadBytes += entry.Value
		case "Write":
			result.IOWriteBytes += entry.Value
		}
	}
	for _, entry := range stats.BlkioStats.IoServicedRecursive {
		switch entry.Op {
		case "Read":
			result.IOReadOperations += entry.Value
		case "Write":
			result.IOWriteOperations += entry.Value
		}
	}
	return result, nil
}

// execPlatformCommand 加入现有容器的命名空间执行一次性命令，不会在宿主机执行。
func (m *Manager) execPlatformCommand(id string, options ExecOptions) (*ExecResult, error) {
	release, err := m.lockOpen()
	if err != nil {
		return nil, err
	}
	unlockInstance := m.lockInstance(id)
	defer func() {
		if unlockInstance != nil {
			unlockInstance()
		}
		if release != nil {
			release()
		}
	}()
	instance, err := m.refreshInstanceLocked(id)
	if err != nil {
		return nil, err
	}
	if instance.Status != StatusRunning {
		return nil, errors.New("实例未运行")
	}
	m.mu.RLock()
	image := m.cloneImage(m.images[instance.ImageID])
	var runtime *linuxRuntime
	if value, ok := m.runtime[id].(*linuxRuntime); ok {
		runtime = value
	}
	m.mu.RUnlock()
	if image == nil {
		return nil, errors.New("实例镜像不存在")
	}
	var container *libcontainer.Container
	if runtime != nil && runtime.container != nil {
		container = runtime.container
	} else {
		loaded, err := libcontainer.Load(m.runtimeRoot, id)
		if err != nil {
			return nil, fmt.Errorf("加载容器运行状态失败: %w", err)
		}
		container = loaded
	}
	env := m.mergeEnv(
		[]string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
		image.Env,
		m.timezoneEnv(container.Config()),
		instance.Env,
		options.Env,
	)
	workingDir := options.WorkingDir
	if workingDir == "" {
		workingDir = instance.WorkingDir
	}
	if workingDir == "" {
		workingDir = image.WorkingDir
	}
	if workingDir == "" {
		workingDir = "/"
	}
	rootfs := instance.Rootfs
	if rootfs == "" {
		rootfs = image.Rootfs
	}
	user := instance.User
	if user == "" {
		user = image.User
	}
	uid, gid, groups, err := m.resolveUser(rootfs, user)
	if err != nil {
		return nil, fmt.Errorf("解析容器运行用户失败: %w", err)
	}
	output := &limitedBuffer{max: options.MaxOutput}
	outputReader, outputWriter, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("创建容器命令输出管道失败: %w", err)
	}
	stderrWriter := outputWriter
	var stderrReader *os.File
	if file, ok := options.Stderr.(*os.File); ok && file != nil {
		// 文件由调用方持有，直接继承，保留原来的终端/文件属性。
		stderrWriter = file
	} else if options.Stderr != nil {
		// 使用可主动关闭的管道，避免子孙进程继承 stderr 后让命令一直等待。
		stderrReader, stderrWriter, err = os.Pipe()
		if err != nil {
			_ = outputReader.Close()
			_ = outputWriter.Close()
			return nil, fmt.Errorf("创建容器错误输出管道失败: %w", err)
		}
	}
	closeReaders := func() {
		_ = outputReader.Close()
		if stderrReader != nil {
			_ = stderrReader.Close()
		}
	}
	closeWriters := func() {
		_ = outputWriter.Close()
		if stderrReader != nil {
			_ = stderrWriter.Close()
		}
	}
	copyDone := make(chan struct{})
	var stderrErr error
	if !m.runBackground(func(ctx context.Context) {
		defer close(copyDone)
		closeDone := make(chan struct{})
		stopClose := context.AfterFunc(ctx, func() {
			defer close(closeDone)
			closeReaders()
		})
		var stderrCopy sync.WaitGroup
		if stderrReader != nil {
			stderrCopy.Add(1)
			go func() {
				defer stderrCopy.Done()
				defer stderrReader.Close()
				if _, err := io.Copy(options.Stderr, stderrReader); err != nil && !errors.Is(err, os.ErrClosed) {
					stderrErr = err
				}
			}()
		}
		_, _ = io.Copy(output, outputReader)
		stderrCopy.Wait()
		if !stopClose() {
			<-closeDone
		}
	}) {
		closeReaders()
		closeWriters()
		return nil, errors.New("容器管理器已关闭")
	}
	closeOutput := func() {
		closeWriters()
		timer := time.NewTimer(time.Second)
		defer timer.Stop()
		select {
		case <-copyDone:
		case <-timer.C:
			closeReaders()
			<-copyDone
		}
		closeReaders()
	}
	defer closeOutput()
	process := &libcontainer.Process{
		Args:             append([]string(nil), options.Command...),
		Env:              env,
		UID:              uid,
		GID:              gid,
		AdditionalGroups: groups,
		Capabilities:     m.processCapabilities(uid),
		Cwd:              workingDir,
		Stdin:            options.Stdin,
		Stdout:           outputWriter,
		Stderr:           stderrWriter,
	}
	flushInitOutput, err := m.prepareInitOutput(process)
	if err != nil {
		return nil, err
	}
	runErr := container.Run(process)
	if err := errors.Join(runErr, flushInitOutput()); err != nil {
		if runErr == nil {
			_ = process.Signal(unix.SIGKILL)
			_, _ = process.Wait()
		}
		return nil, fmt.Errorf("执行容器命令失败: %w", err)
	}
	closeWriters()
	wait := make(chan waitResult, 1)
	if !m.runBackground(func(ctx context.Context) {
		state, err := m.waitPlatformProcess(ctx, process, m.managerOptions().StopGracePeriod)
		wait <- waitResult{
			state: state,
			err:   err,
		}
	}) {
		_ = process.Signal(unix.SIGKILL)
		_, _ = process.Wait()
		return nil, errors.New("容器管理器已关闭")
	}
	m.mu.RLock()
	managerContext := m.ctx
	m.mu.RUnlock()
	unlockInstance()
	unlockInstance = nil
	release()
	release = nil
	timer := time.NewTimer(options.Timeout)
	defer timer.Stop()
	result := &ExecResult{ExitCode: -1}
	select {
	case value := <-wait:
		if value.err != nil {
			// 命令自身失败仍需返回退出码和输出，不能当作运行时等待失败而丢弃诊断信息。
			var exitError *exec.ExitError
			if !errors.As(value.err, &exitError) {
				return nil, fmt.Errorf("等待容器命令结束失败: %w", value.err)
			}
			if value.state == nil {
				value.state = exitError.ProcessState
			}
		}
		if value.state != nil {
			result.ExitCode = value.state.ExitCode()
		}
	case <-timer.C:
		result.TimedOut = true
		_ = process.Signal(unix.SIGKILL)
		forceTimer := time.NewTimer(m.managerOptions().ForceStopPeriod)
		select {
		case value := <-wait:
			if value.state != nil {
				result.ExitCode = value.state.ExitCode()
			}
		case <-forceTimer.C:
		}
		forceTimer.Stop()
	case <-managerContext.Done():
		value := <-wait
		if value.state != nil {
			result.ExitCode = value.state.ExitCode()
		}
		return nil, errors.New("容器管理器已关闭")
	}
	closeOutput()
	result.Output, result.Truncated = output.snapshot()
	if stderrErr != nil {
		return result, fmt.Errorf("转发容器命令错误输出失败: %w", stderrErr)
	}
	return result, nil
}

// waitPlatformProcess 在管理器退出时先优雅终止进程，超时后强制结束，并始终回收进程状态。
func (m *Manager) waitPlatformProcess(ctx context.Context, process *libcontainer.Process, grace time.Duration) (*os.ProcessState, error) {
	done := make(chan struct{})
	stopDone := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		defer close(stopDone)
		_ = process.Signal(unix.SIGTERM)
		timer := time.NewTimer(grace)
		defer timer.Stop()
		select {
		case <-done:
			return
		case <-timer.C:
			_ = process.Signal(unix.SIGKILL)
		}
	})
	state, err := process.Wait()
	close(done)
	if !stop() {
		<-stopDone
	}
	return state, err
}

type limitedBuffer struct {
	mu        sync.Mutex
	buffer    bytes.Buffer
	max       int64
	truncated bool
}

func (b *limitedBuffer) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	originalLength := len(data)
	if b.max <= 0 {
		b.truncated = b.truncated || originalLength > 0
		return originalLength, nil
	}
	left := b.max - int64(b.buffer.Len())
	if left > 0 {
		if int64(len(data)) > left {
			b.truncated = true
			data = data[:left]
		}
		_, _ = b.buffer.Write(data)
	} else if originalLength > 0 {
		b.truncated = true
	}
	return originalLength, nil
}

func (b *limitedBuffer) snapshot() (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.String(), b.truncated
}

func (m *Manager) openRuntimeLog(path string) (*os.File, error) {
	if err := m.rotateLogFile(path); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

func (m *Manager) clearPlatformLog(id, path string) error {
	m.mu.RLock()
	runtime, _ := m.runtime[id].(*linuxRuntime)
	m.mu.RUnlock()
	if runtime != nil {
		runtime.logMu.Lock()
		defer runtime.logMu.Unlock()
	}
	clearErr := fileLog.Clear(path)
	var removeErr error
	for _, name := range []string{path + ".1", path + ".1.tmp"} {
		if err := os.Remove(name); err != nil && !errors.Is(err, os.ErrNotExist) {
			removeErr = errors.Join(removeErr, err)
		}
	}
	return errors.Join(clearErr, removeErr)
}

func (m *Manager) stopPlatformRuntime(instance *Info) error {
	m.mu.RLock()
	runtime, _ := m.runtime[instance.ID].(*linuxRuntime)
	m.mu.RUnlock()
	var container *libcontainer.Container
	if runtime != nil && runtime.container != nil {
		container = runtime.container
	} else {
		loaded, err := libcontainer.Load(m.runtimeRoot, instance.ID)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if m.isActiveStatus(instance.Status) || instance.PID > 0 {
					return errors.New("容器运行状态目录不存在，无法安全停止进程")
				}
				return nil
			}
			return fmt.Errorf("加载容器运行状态失败: %w", err)
		}
		container = loaded
	}
	if container == nil {
		return nil
	}
	status, err := container.Status()
	if err != nil {
		return fmt.Errorf("读取容器运行状态失败: %w", err)
	}
	if status == libcontainer.Stopped {
		return m.releasePlatformRuntime(instance.ID, runtime, container)
	}
	if err := container.Signal(unix.SIGTERM); err != nil && !errors.Is(err, libcontainer.ErrNotRunning) {
		return fmt.Errorf("停止容器失败: %w", err)
	}
	if m.waitContainerStopped(container, m.managerOptions().StopGracePeriod) {
		return m.releasePlatformRuntime(instance.ID, runtime, container)
	}
	if err := container.Signal(unix.SIGKILL); err != nil && !errors.Is(err, libcontainer.ErrNotRunning) {
		return fmt.Errorf("强制停止容器失败: %w", err)
	}
	if !m.waitContainerStopped(container, m.managerOptions().ForceStopPeriod) {
		return errors.New("等待容器停止超时")
	}
	return m.releasePlatformRuntime(instance.ID, runtime, container)
}

func (m *Manager) releasePlatformRuntime(id string, runtime *linuxRuntime, container *libcontainer.Container) error {
	if runtime == nil {
		runtime = &linuxRuntime{container: container}
	}
	m.mu.RLock()
	if instance := m.instances[id]; instance != nil {
		runtime.logPath = instance.LogPath
	}
	m.mu.RUnlock()
	runtime.logMu.Lock()
	if runtime.logFile != nil {
		_ = runtime.logFile.Close()
		runtime.logFile = nil
	}
	runtime.logMu.Unlock()
	destroyErr := container.Destroy()
	if destroyErr != nil {
		// 私有 PID namespace 理论上会随 init 一起退出；异常残留只能用 pidfd 精确清理。
		if processErr := m.killPlatformProcesses(container); processErr != nil {
			destroyErr = errors.Join(destroyErr, processErr)
		} else {
			destroyErr = container.Destroy()
		}
	}
	cleanupGeneration := uint64(0)
	retained := false
	m.mu.Lock()
	current, _ := m.runtime[id].(*linuxRuntime)
	if destroyErr != nil {
		if current == nil {
			m.runtime[id] = runtime
			current = runtime
		}
		retained = current == runtime || current != nil && current.container == container
		if retained {
			if instance := m.instances[id]; instance != nil {
				instance.Status = StatusStopping
				instance.PID = 0
				cleanupGeneration = instance.Generation
			}
		}
	} else if current == runtime || current != nil && current.container == container {
		delete(m.runtime, id)
	}
	m.mu.Unlock()
	if destroyErr != nil {
		destroyErr = errors.Join(errors.New("容器运行状态等待清理"), fmt.Errorf("清理容器运行状态失败: %w", destroyErr))
		if retained && cleanupGeneration > 0 {
			m.retryPlatformCleanup(id, cleanupGeneration, runtime)
		}
	}
	return destroyErr
}

func (m *Manager) abortPlatformStart(instance *Info, container *libcontainer.Container, process *libcontainer.Process, logFile *os.File, failure error) error {
	runtime := &linuxRuntime{
		container: container,
		process:   process,
		logFile:   logFile,
		logPath:   instance.LogPath,
		exitCode:  -1,
		failure:   failure,
	}
	return m.releasePlatformRuntime(instance.ID, runtime, container)
}

// setPlatformFailure 单独保留启动事务错误，避免临时清理错误污染最终状态。
func (m *Manager) setPlatformFailure(id string, failure error) {
	if failure == nil {
		return
	}
	m.mu.RLock()
	runtime, _ := m.runtime[id].(*linuxRuntime)
	m.mu.RUnlock()
	if runtime != nil && runtime.failure == nil {
		runtime.failure = failure
	}
}

// killPlatformProcesses 通过 pidfd 二次确认进程归属后清理异常残留的 cgroup 进程。
func (m *Manager) killPlatformProcesses(container *libcontainer.Container) error {
	pids, err := container.Processes()
	if err != nil {
		return fmt.Errorf("读取容器残留进程失败: %w", err)
	}
	self := os.Getpid()
	for _, pid := range pids {
		if pid <= 1 || pid == self {
			return fmt.Errorf("拒绝清理不安全的容器进程 PID: %d", pid)
		}
	}
	processes := make([]process, 0, len(pids))
	defer func() {
		for _, item := range processes {
			_ = unix.Close(item.fd)
		}
	}()
	for _, pid := range pids {
		fd, openErr := unix.PidfdOpen(pid, 0)
		if errors.Is(openErr, unix.ESRCH) {
			continue
		}
		if openErr != nil {
			return fmt.Errorf("打开容器残留进程 pidfd 失败: PID=%d: %w", pid, openErr)
		}
		processes = append(processes, process{
			pid: pid,
			fd:  fd,
		})
	}
	confirmedPIDs, err := container.Processes()
	if err != nil {
		return fmt.Errorf("确认容器残留进程失败: %w", err)
	}
	confirmed := make(map[int]struct{}, len(confirmedPIDs))
	for _, pid := range confirmedPIDs {
		if pid <= 1 || pid == self {
			return fmt.Errorf("拒绝清理不安全的容器进程 PID: %d", pid)
		}
		confirmed[pid] = struct{}{}
	}
	var signalErr error
	for _, item := range processes {
		if _, exists := confirmed[item.pid]; !exists {
			continue
		}
		if err := unix.PidfdSendSignal(item.fd, unix.SIGKILL, nil, 0); err != nil && !errors.Is(err, unix.ESRCH) {
			signalErr = errors.Join(signalErr, fmt.Errorf("终止容器残留进程失败: PID=%d: %w", item.pid, err))
		}
	}
	return signalErr
}

func (m *Manager) lockPlatformLog(id string) func() {
	m.mu.RLock()
	runtime, _ := m.runtime[id].(*linuxRuntime)
	m.mu.RUnlock()
	if runtime == nil {
		return func() {}
	}
	runtime.logMu.Lock()
	return runtime.logMu.Unlock
}

// retryPlatformCleanup 以退避方式重试持久化状态、残留进程和 libcontainer 状态清理。
func (m *Manager) retryPlatformCleanup(id string, generation uint64, runtime *linuxRuntime) {
	runtime.cleanup.Do(func() {
		m.runBackground(func(ctx context.Context) {
			delay := m.managerOptions().RuntimeCleanupDelay
			for {
				timer := time.NewTimer(delay)
				select {
				case <-ctx.Done():
					timer.Stop()
					return
				case <-timer.C:
				}
				if ctx.Err() != nil {
					return
				}
				unlock := m.lockInstance(id)
				m.mu.RLock()
				instance := m.cloneInstance(m.instances[id])
				current, _ := m.runtime[id].(*linuxRuntime)
				m.mu.RUnlock()
				if instance == nil || instance.Generation != generation || current != runtime {
					unlock()
					return
				}
				if err := m.markRuntimeCleanupLocked(id, generation, runtime.exitCode, runtime.exitErr); err != nil {
					unlock()
					if delay < 30*time.Second {
						delay *= 2
						if delay > 30*time.Second {
							delay = 30 * time.Second
						}
					}
					continue
				}
				err := m.stopPlatformRuntime(instance)
				if err == nil {
					m.markExitedLocked(id, generation, runtime.exitCode, runtime.exitErr, true, runtime.failure)
					if runtime.failure == nil {
						m.scheduleAutoRestart(id, generation)
					}
					unlock()
					return
				}
				m.markExitedLocked(id, generation, runtime.exitCode, errors.Join(runtime.exitErr, err), false, nil)
				unlock()
				if delay < 30*time.Second {
					delay *= 2
					if delay > 30*time.Second {
						delay = 30 * time.Second
					}
				}
			}
		})
	})
}

func (m *Manager) waitContainerStopped(container *libcontainer.Container, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		status, err := container.Status()
		if err == nil && status == libcontainer.Stopped {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// prepareTimezoneMount 只准备实例内的时区文件，不沿叶子链接覆盖镜像的其他时区数据。
func (m *Manager) prepareTimezoneMount(rootfs, destination string) (string, error) {
	// 调用线程已 LockOSThread + Unshare(CLONE_NEWNS)，后者隐含 CLONE_FS，
	// 因而这里只调整线程自己的 umask，新建目录保持可遍历，现有目录权限不变。
	umask := unix.Umask(0)
	defer unix.Umask(umask)
	root, err := os.OpenFile(rootfs, unix.O_PATH|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return "", err
	}
	defer root.Close()
	dir, err := pathrs.MkdirAllHandle(root, strings.TrimPrefix(filepath.Dir(destination), "/"), 0755)
	if err != nil {
		return "", fmt.Errorf("准备容器时区目录失败: %w", err)
	}
	defer dir.Close()
	// 通过目录句柄操作最终文件，父目录中的绝对/相对链接只在 rootfs 内解析。
	directory, err := os.Readlink(fmt.Sprintf("/proc/thread-self/fd/%d", dir.Fd()))
	if err != nil {
		return "", err
	}
	if err := m.normalizePlatformPathOwnership(directory, rootfs); err != nil {
		return "", err
	}
	name := filepath.Base(destination)
	var stat unix.Stat_t
	err = unix.Fstatat(int(dir.Fd()), name, &stat, unix.AT_SYMLINK_NOFOLLOW)
	if err == nil && stat.Mode&unix.S_IFMT == unix.S_IFLNK {
		if err := unix.Unlinkat(int(dir.Fd()), name, 0); err != nil {
			return "", fmt.Errorf("替换容器时区链接失败: %w", err)
		}
		err = unix.ENOENT
	}
	if err != nil && !errors.Is(err, unix.ENOENT) {
		return "", err
	}
	if err == nil && stat.Mode&unix.S_IFMT != unix.S_IFREG {
		return "", fmt.Errorf("容器时区文件不是普通文件或符号链接: %s", destination)
	}
	if errors.Is(err, unix.ENOENT) {
		fd, err := unix.Openat(int(dir.Fd()), name, unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0644)
		if err != nil {
			return "", fmt.Errorf("创建容器时区挂载目标失败: %w", err)
		}
		ownerErr := unix.Fchown(fd, int(m.security.UIDMapStart), int(m.security.GIDMapStart))
		if err := errors.Join(ownerErr, unix.Close(fd)); err != nil {
			return "", err
		}
	}
	relative, err := filepath.Rel(rootfs, filepath.Join(directory, name))
	if err != nil {
		return "", err
	}
	return "/" + filepath.ToSlash(relative), nil
}

// hostTimezoneName 仅采用与本次启动快照完全一致的系统名称，自定义 TZif 保留文件路径方式。
func (m *Manager) hostTimezoneName(data []byte) string {
	name, _, err := platform.TimezoneGet()
	if err != nil {
		return ""
	}
	for _, directory := range []string{"/usr/share/zoneinfo", "/etc/zoneinfo", "/usr/share/lib/zoneinfo", "/usr/lib/zoneinfo"} {
		file := filepath.Join(directory, name)
		if expected, err := os.ReadFile(file); err == nil && bytes.Equal(data, expected) {
			// SQL Server 依赖别名链接解析映射；直接使用宿主解析后的名称。
			root, rootErr := filepath.EvalSymlinks(directory)
			resolved, resolveErr := filepath.EvalSymlinks(file)
			if rootErr == nil && resolveErr == nil {
				if canonical, err := filepath.Rel(root, resolved); err == nil && canonical != ".." && !strings.HasPrefix(canonical, "../") {
					return filepath.ToSlash(canonical)
				}
			}
			return name
		}
	}
	return ""
}

// timezoneEnv 使用实际运行配置；标准文件也供清除 TZ 的进程直接读取。
// 默认值覆盖镜像自带的 TZ，但仍允许实例和 Exec 显式覆盖（包括 TZ=）。
func (m *Manager) timezoneEnv(config configs.Config) []string {
	for _, label := range config.Labels {
		if name, ok := strings.CutPrefix(label, timezoneNameLabel); ok && name != "" {
			return []string{"TZ=" + name}
		}
	}
	for _, label := range config.Labels {
		if strings.HasPrefix(label, timezoneMountLabel) {
			return []string{"TZ=:" + containerTimezonePath}
		}
	}
	for _, mount := range config.Mounts {
		switch filepath.Clean(mount.Destination) {
		case "/etc", "/etc/localtime":
			return []string{"TZ=:/etc/localtime"}
		}
	}
	return nil
}

func (m *Manager) buildConfig(instance *Info, image *Image, options RunOptions, mountsReady bool) (*configs.Config, error) {
	security, ok := m.platformSecurity.(*linuxSecurity)
	if !ok || security == nil || security.capabilities == nil || security.seccomp == nil {
		return nil, errors.New("容器安全配置未初始化")
	}
	if err := m.validateImageSecurity(image); err != nil {
		return nil, err
	}
	if err := m.validateInstanceSecurity(instance); err != nil {
		return nil, err
	}
	namespaces := configs.Namespaces([]configs.Namespace{
		{Type: configs.NEWNS},
		{Type: configs.NEWUTS},
		{Type: configs.NEWIPC},
		{Type: configs.NEWPID},
		{Type: configs.NEWUSER},
		// 不创建 NEWNET，容器直接使用宿主机网络。
	})
	mounts := []*configs.Mount{
		{
			Source:      "proc",
			Destination: "/proc",
			Device:      "proc",
			Flags:       unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_NODEV,
		},
		{
			Source:      "tmpfs",
			Destination: "/dev",
			Device:      "tmpfs",
			Flags:       unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_STRICTATIME,
			Data:        "mode=755,size=65536k",
		},
		{
			Source:      "devpts",
			Destination: "/dev/pts",
			Device:      "devpts",
			Flags:       unix.MS_NOSUID | unix.MS_NOEXEC,
			Data:        "newinstance,ptmxmode=0666,mode=0620,gid=5",
		},
		{
			Source:      "shm",
			Destination: "/dev/shm",
			Device:      "tmpfs",
			Flags:       unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_NODEV,
			Data:        "mode=1777,size=65536k",
		},
		{
			Source:      "/sys",
			Destination: "/sys",
			Device:      "bind",
			// User Namespace 会锁定继承的子挂载，必须一并绑定，不能剥离 cgroup 等子挂载。
			Flags: unix.MS_BIND | unix.MS_REC | unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_NODEV | unix.MS_RDONLY,
		},
	}
	// 每次启动重新解析宿主时区，保留完整的夏令时规则，不固化成当前 UTC 偏移。
	// 用户挂载 /etc/localtime、/etc 时保留其文件；显式 TZ 仅覆盖进程环境。
	inheritTimezone := true
	for _, mount := range options.Mounts {
		if destination := filepath.Clean(mount.Destination); destination == "/etc" || destination == "/etc/localtime" {
			inheritTimezone = false
			break
		}
	}
	if inheritTimezone {
		mounts = append(mounts, &configs.Mount{
			Source:      "/etc/localtime",
			Destination: containerTimezonePath,
			Device:      "bind",
			Flags:       unix.MS_BIND | unix.MS_RDONLY | unix.MS_NOSUID | unix.MS_NODEV | unix.MS_NOEXEC,
		})
	}
	for _, mount := range options.Mounts {
		if !mountsReady {
			if err := m.validateMount(mount); err != nil {
				return nil, err
			}
			if mount.ReadOnly {
				if err := m.validatePlatformMountContent(mount.Source); err != nil {
					return nil, err
				}
			} else {
				if err := m.validatePlatformMountOwnership(mount.Source); err != nil {
					return nil, err
				}
			}
		}
		flags := unix.MS_BIND | unix.MS_NOSUID | unix.MS_NODEV
		if mount.ReadOnly {
			flags |= unix.MS_RDONLY
		}
		mounts = append(mounts, &configs.Mount{
			Source:      mount.Source,
			Destination: mount.Destination,
			Device:      "bind",
			Flags:       flags,
		})
	}
	defaultDevices := m.defaultDevices()
	deviceRules := []*devices.Rule{{
		Type:        devices.WildcardDevice,
		Major:       devices.Wildcard,
		Minor:       devices.Wildcard,
		Permissions: "rwm",
		Allow:       false,
	}}
	for _, device := range defaultDevices {
		rule := device.Rule
		deviceRules = append(deviceRules, &rule)
	}
	deviceRules = append(deviceRules,
		&devices.Rule{
			Type:        devices.CharDevice,
			Major:       5,
			Minor:       2,
			Permissions: "rwm",
			Allow:       true,
		},
		&devices.Rule{
			Type:        devices.CharDevice,
			Major:       136,
			Minor:       devices.Wildcard,
			Permissions: "rwm",
			Allow:       true,
		},
	)
	resources := &cgroups.Resources{
		Devices:   deviceRules,
		Memory:    instance.Resources.Memory,
		CpuQuota:  instance.Resources.CPUQuota,
		CpuPeriod: instance.Resources.CPUPeriod,
	}
	pidsLimit := instance.Resources.PidsLimit
	if pidsLimit == 0 && m.managerOptions().DefaultPidsLimit > 0 {
		pidsLimit = m.managerOptions().DefaultPidsLimit
	}
	if pidsLimit > 0 {
		limit := pidsLimit
		resources.PidsLimit = &limit
	}
	return &configs.Config{
		Rootfs: func() string {
			if instance.Rootfs != "" {
				return instance.Rootfs
			}
			return image.Rootfs
		}(),
		Readonlyfs:      !instance.WritableLayer,
		Hostname:        instance.ID,
		Namespaces:      namespaces,
		RootPropagation: unix.MS_PRIVATE | unix.MS_REC,
		UIDMappings: []configs.IDMap{{
			ContainerID: 0,
			HostID:      m.security.UIDMapStart,
			Size:        m.security.IDMapSize,
		}},
		GIDMappings: []configs.IDMap{{
			ContainerID: 0,
			HostID:      m.security.GIDMapStart,
			Size:        m.security.IDMapSize,
		}},
		Capabilities:    security.capabilities,
		Seccomp:         security.seccomp,
		NoNewPrivileges: true,
		// 安装目录哈希隔离不同 Manager，实例 ID 隔离同一 Manager 内的容器。
		Cgroups: &cgroups.Cgroup{
			Path:      filepath.Join("/sinking-cloud", m.installation, instance.ID),
			Resources: resources,
		},
		Devices: defaultDevices,
		Mounts:  mounts,
		MaskPaths: []string{
			"/proc/acpi",
			"/proc/asound",
			"/proc/kcore",
			"/proc/keys",
			"/proc/latency_stats",
			"/proc/sched_debug",
			"/proc/scsi",
			"/proc/timer_list",
			"/proc/timer_stats",
			"/sys/firmware",
			"/sys/devices/system/cpu/cpu0/thermal_throttle",
			"/sys/devices/virtual/powercap",
		},
		ReadonlyPaths: []string{
			"/proc/bus",
			"/proc/fs",
			"/proc/irq",
			"/proc/kcore",
			"/proc/keys",
			"/proc/latency_stats",
			"/proc/sys",
			"/proc/sysrq-trigger",
			"/proc/timer_list",
		},
	}, nil
}

func (m *Manager) resolveUser(rootfs, value string) (int, int, []int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, 0, nil, nil
	}
	var files [2]io.Reader
	for index, name := range []string{"etc/passwd", "etc/group"} {
		// /etc 可以是镜像内的链接；固定文件句柄，避免链接被解析到宿主机。
		handle, err := pathrs.OpenInRoot(rootfs, name)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return 0, 0, nil, err
		}
		info, err := handle.Stat()
		if err != nil || !info.Mode().IsRegular() {
			_ = handle.Close()
			return 0, 0, nil, fmt.Errorf("容器用户文件不是可读的普通文件: %s", name)
		}
		file, err := pathrs.Reopen(handle, unix.O_RDONLY|unix.O_CLOEXEC)
		_ = handle.Close()
		if err != nil {
			return 0, 0, nil, err
		}
		defer file.Close()
		files[index] = file
	}
	user, err := userUtil.GetExecUser(
		value,
		&userUtil.ExecUser{
			Uid: 0,
			Gid: 0,
		},
		files[0],
		files[1],
	)
	if err != nil {
		return 0, 0, nil, err
	}
	return user.Uid, user.Gid, append([]int(nil), user.Sgids...), nil
}

func (m *Manager) normalizeArchitecture(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "x86_64", "x64":
		return "amd64"
	case "aarch64":
		return "arm64"
	case "x86", "i386", "i686":
		return "386"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func (m *Manager) rootfsPathInfo(rootfs, path string) (os.FileInfo, error) {
	if err := m.ensureInside(rootfs, path); err != nil {
		return nil, err
	}
	rootInfo, err := os.Lstat(rootfs)
	if err != nil {
		return nil, err
	}
	if !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("容器 rootfs 必须是普通目录")
	}
	relative, err := filepath.Rel(rootfs, path)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(relative, string(filepath.Separator))
	current := rootfs
	for index, part := range parts {
		if part == "." || part == "" {
			continue
		}
		current = filepath.Join(current, part)
		info, statErr := os.Lstat(current)
		if errors.Is(statErr, os.ErrNotExist) {
			return nil, fmt.Errorf("rootfs 路径不存在: %w", os.ErrNotExist)
		}
		if statErr != nil {
			return nil, statErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("rootfs 路径不能经过符号链接: %s", current)
		}
		if index < len(parts)-1 && !info.IsDir() {
			return nil, fmt.Errorf("rootfs 父路径不是目录: %s", current)
		}
		if index == len(parts)-1 {
			return info, nil
		}
	}
	return rootInfo, nil
}

func (m *Manager) defaultDevices() []*devices.Device {
	return []*devices.Device{
		{
			Path:     "/dev/null",
			FileMode: 0666,
			Rule: devices.Rule{
				Type:        devices.CharDevice,
				Major:       1,
				Minor:       3,
				Permissions: "rwm",
				Allow:       true,
			},
		},
		{
			Path:     "/dev/zero",
			FileMode: 0666,
			Rule: devices.Rule{
				Type:        devices.CharDevice,
				Major:       1,
				Minor:       5,
				Permissions: "rwm",
				Allow:       true,
			},
		},
		{
			Path:     "/dev/full",
			FileMode: 0666,
			Rule: devices.Rule{
				Type:        devices.CharDevice,
				Major:       1,
				Minor:       7,
				Permissions: "rwm",
				Allow:       true,
			},
		},
		{
			Path:     "/dev/random",
			FileMode: 0666,
			Rule: devices.Rule{
				Type:        devices.CharDevice,
				Major:       1,
				Minor:       8,
				Permissions: "rwm",
				Allow:       true,
			},
		},
		{
			Path:     "/dev/urandom",
			FileMode: 0666,
			Rule: devices.Rule{
				Type:        devices.CharDevice,
				Major:       1,
				Minor:       9,
				Permissions: "rwm",
				Allow:       true,
			},
		},
		{
			Path:     "/dev/tty",
			FileMode: 0666,
			Rule: devices.Rule{
				Type:        devices.CharDevice,
				Major:       5,
				Minor:       0,
				Permissions: "rwm",
				Allow:       true,
			},
		},
	}
}

func (m *Manager) prepareMountTargets(rootfs string, mounts []Mount, create bool) error {
	for _, mount := range mounts {
		destinationName := strings.TrimPrefix(filepath.ToSlash(mount.Destination), "/")
		destination, err := m.safeJoin(rootfs, filepath.FromSlash(destinationName))
		if err != nil {
			return fmt.Errorf("挂载目标不安全: %w", err)
		}
		if create {
			if err := m.ensureParent(destination, rootfs); err != nil {
				return fmt.Errorf("创建挂载父目录失败: %w", err)
			}
		}
		sourceInfo, err := os.Lstat(mount.Source)
		if err != nil {
			return fmt.Errorf("读取挂载源失败: %w", err)
		}
		var destinationInfo os.FileInfo
		var destinationErr error
		if create {
			destinationInfo, destinationErr = os.Lstat(destination)
		} else {
			destinationInfo, destinationErr = m.rootfsPathInfo(rootfs, destination)
		}
		if !create && errors.Is(destinationErr, os.ErrNotExist) {
			return fmt.Errorf("只读 rootfs 中不存在挂载目标: %s", mount.Destination)
		}
		if destinationErr == nil && destinationInfo.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("挂载目标不能是符号链接: %s", mount.Destination)
		}
		if sourceInfo.IsDir() {
			if destinationErr == nil && !destinationInfo.IsDir() {
				return fmt.Errorf("目录挂载目标不是目录: %s", mount.Destination)
			}
			if destinationErr != nil && !errors.Is(destinationErr, os.ErrNotExist) {
				return destinationErr
			}
			if destinationErr != nil {
				if !create {
					return destinationErr
				}
				if err := os.MkdirAll(destination, 0755); err != nil {
					return err
				}
			}
			if create {
				if err := m.normalizePlatformPathOwnership(destination, rootfs); err != nil {
					return fmt.Errorf("映射挂载目标所有者失败: %w", err)
				}
			}
			continue
		}
		if destinationErr == nil && destinationInfo.IsDir() {
			return fmt.Errorf("文件挂载目标是目录: %s", mount.Destination)
		}
		if destinationErr != nil && !errors.Is(destinationErr, os.ErrNotExist) {
			return destinationErr
		}
		if create && errors.Is(destinationErr, os.ErrNotExist) {
			file, createErr := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY, 0644)
			if createErr != nil {
				return createErr
			}
			_ = file.Close()
		}
		if create {
			if err := m.normalizePlatformPathOwnership(destination, rootfs); err != nil {
				return fmt.Errorf("映射挂载目标所有者失败: %w", err)
			}
		}
	}
	return nil
}
