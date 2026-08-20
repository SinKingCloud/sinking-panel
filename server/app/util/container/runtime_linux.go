//go:build linux && cgo

package container

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	standardRuntime "runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	userUtil "github.com/moby/sys/user"
	"github.com/opencontainers/cgroups"
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

func (m *Manager) startPlatformRuntime(instance *Instance, image *Image, options RunOptions) error {
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
	rootfs := instance.Rootfs
	if rootfs == "" {
		rootfs = image.Rootfs
	}
	if err := m.prepareMountTargets(rootfs, options.Mounts, instance.WritableLayer); err != nil {
		return err
	}
	config := m.buildConfig(instance, image, options)
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
		options.Env,
	)
	workingDir := options.WorkingDir
	if workingDir == "" {
		workingDir = image.WorkingDir
	}
	if workingDir == "" {
		workingDir = "/"
	}
	uid, gid, groups, err := m.resolveUser(rootfs, image.User)
	if err != nil {
		failure := fmt.Errorf("解析镜像运行用户失败: %w", err)
		return errors.Join(failure, m.abortPlatformStart(instance, container, nil, logFile, failure))
	}
	process := &libcontainer.Process{
		Args:             args,
		Env:              env,
		UID:              uid,
		GID:              gid,
		AdditionalGroups: groups,
		Cwd:              workingDir,
		Stdout:           logFile,
		Stderr:           logFile,
		Init:             true,
	}
	if err := container.Run(process); err != nil {
		failure := fmt.Errorf("启动容器进程失败: %w", err)
		return errors.Join(failure, m.abortPlatformStart(instance, container, process, logFile, failure))
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
	runtime := &linuxRuntime{container: container, process: process, logFile: logFile, logPath: instance.LogPath}
	m.mu.Lock()
	m.runtime[instance.ID] = runtime
	m.mu.Unlock()
	go m.watchRuntimeLog(instance.ID, instance.Generation, runtime)
	go func(id string, generation uint64, runtime *linuxRuntime) {
		state, waitErr := runtime.process.Wait()
		exitCode := -1
		if state != nil {
			exitCode = state.ExitCode()
		}
		m.finishPlatformRuntime(id, generation, runtime, exitCode, waitErr)
	}(instance.ID, instance.Generation, runtime)
	return nil
}

// preparePlatformRootfs 以镜像为只读 lower 层，为实例准备独立 overlay 写层。
func (m *Manager) preparePlatformRootfs(instance *Instance, image *Image, writable bool) error {
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
func (m *Manager) cleanupPlatformRootfs(instance *Instance) error {
	if !instance.WritableLayer || instance.Rootfs == "" || instance.UpperDir == "" {
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
	if err := m.cleanupOrphanInstanceStorage(id); err != nil {
		return err
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
func (m *Manager) recoverPlatformRuntime(instance *Instance) (bool, error) {
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
	runtime := &linuxRuntime{container: container, logPath: instance.LogPath, exitCode: instance.ExitCode}
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
	go m.watchRuntimeLog(instance.ID, instance.Generation, runtime)
	go m.watchRecoveredRuntime(instance.ID, instance.Generation, runtime)
	return true, nil
}

// refreshPlatformRuntime 根据 libcontainer 状态刷新实例 PID。
func (m *Manager) refreshPlatformRuntime(instance *Instance) (bool, error) {
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

func (m *Manager) watchRecoveredRuntime(id string, generation uint64, runtime *linuxRuntime) {
	ticker := time.NewTicker(m.managerOptions().RuntimePollInterval)
	defer ticker.Stop()
	failures := 0
	for range ticker.C {
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
	m.markExitedLocked(id, generation, exitCode, errors.Join(waitErr, releaseErr), !errors.Is(releaseErr, errRuntimeCleanupPending), runtime.failure)
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

func (m *Manager) watchRuntimeLog(id string, generation uint64, runtime *linuxRuntime) {
	ticker := time.NewTicker(m.managerOptions().RuntimeLogInterval)
	defer ticker.Stop()
	for range ticker.C {
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

func (m *Manager) statsPlatform(id string) (*Stats, error) {
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
	result := &Stats{}
	if data.CgroupStats != nil {
		result.CPUUsage = data.CgroupStats.CpuStats.CpuUsage.TotalUsage
		result.MemoryUsage = data.CgroupStats.MemoryStats.Usage.Usage
		// cgroup 未设置内存上限时，runc 使用 uint64 最大值表示无限制。
		memoryLimit := data.CgroupStats.MemoryStats.Usage.Limit
		if memoryLimit == ^uint64(0) {
			memoryLimit = 0
		}
		result.MemoryLimit = memoryLimit
		result.Pids = data.CgroupStats.PidsStats.Current
	}
	return result, nil
}

// execPlatformCommand 加入现有容器的命名空间执行一次性命令，不会在宿主机执行。
func (m *Manager) execPlatformCommand(id string, options ExecOptions) (*ExecResult, error) {
	m.mu.RLock()
	instance := m.cloneInstance(m.instances[id])
	if instance == nil {
		m.mu.RUnlock()
		return nil, errors.New("实例不存在")
	}
	if (instance.Status != "running" && instance.Status != "unknown") || instance.PID <= 0 {
		m.mu.RUnlock()
		return nil, errors.New("实例未运行")
	}
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
	uid, gid, groups, err := m.resolveUser(rootfs, image.User)
	if err != nil {
		return nil, fmt.Errorf("解析镜像运行用户失败: %w", err)
	}
	output := &limitedBuffer{max: options.MaxOutput}
	outputReader, outputWriter, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("创建容器命令输出管道失败: %w", err)
	}
	process := &libcontainer.Process{
		Args:             append([]string(nil), options.Command...),
		Env:              env,
		UID:              uid,
		GID:              gid,
		AdditionalGroups: groups,
		Cwd:              workingDir,
		Stdout:           outputWriter,
		Stderr:           outputWriter,
	}
	if err := container.Run(process); err != nil {
		_ = outputReader.Close()
		_ = outputWriter.Close()
		return nil, fmt.Errorf("执行容器命令失败: %w", err)
	}
	_ = outputWriter.Close()
	copyDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(output, outputReader)
		close(copyDone)
	}()
	closeOutput := func() {
		select {
		case <-copyDone:
		case <-time.After(time.Second):
			_ = outputReader.Close()
			<-copyDone
		}
		_ = outputReader.Close()
	}
	defer closeOutput()
	type waitResult struct {
		state *os.ProcessState
		err   error
	}
	wait := make(chan waitResult, 1)
	go func() {
		state, err := process.Wait()
		wait <- waitResult{state: state, err: err}
	}()
	timer := time.NewTimer(options.Timeout)
	defer timer.Stop()
	result := &ExecResult{ExitCode: -1}
	select {
	case value := <-wait:
		if value.err != nil {
			return nil, fmt.Errorf("等待容器命令结束失败: %w", value.err)
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
	}
	closeOutput()
	result.Output, result.Truncated = output.snapshot()
	return result, nil
}

type limitedBuffer struct {
	mu sync.Mutex
	bytes.Buffer
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
	left := b.max - int64(b.Len())
	if left > 0 {
		if int64(len(data)) > left {
			b.truncated = true
			data = data[:left]
		}
		_, _ = b.Buffer.Write(data)
	} else if originalLength > 0 {
		b.truncated = true
	}
	return originalLength, nil
}

func (b *limitedBuffer) snapshot() (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Buffer.String(), b.truncated
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
	var truncateErr error
	if runtime != nil && runtime.logFile != nil {
		truncateErr = runtime.logFile.Truncate(0)
	} else {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			truncateErr = err
		} else {
			truncateErr = file.Close()
		}
	}
	var removeErr error
	for _, name := range []string{path + ".1", path + ".1.tmp"} {
		if err := os.Remove(name); err != nil && !errors.Is(err, os.ErrNotExist) {
			removeErr = errors.Join(removeErr, err)
		}
	}
	return errors.Join(truncateErr, removeErr)
}

func (m *Manager) stopPlatformRuntime(instance *Instance) error {
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
				cleanupGeneration = instance.Generation
			}
		}
	} else if current == runtime || current != nil && current.container == container {
		delete(m.runtime, id)
	}
	m.mu.Unlock()
	if destroyErr != nil {
		destroyErr = errors.Join(errRuntimeCleanupPending, fmt.Errorf("清理容器运行状态失败: %w", destroyErr))
		if retained && cleanupGeneration > 0 {
			m.retryPlatformCleanup(id, cleanupGeneration, runtime)
		}
	}
	return destroyErr
}

func (m *Manager) abortPlatformStart(instance *Instance, container *libcontainer.Container, process *libcontainer.Process, logFile *os.File, failure error) error {
	runtime := &linuxRuntime{container: container, process: process, logFile: logFile, logPath: instance.LogPath, exitCode: -1, failure: failure}
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
	type process struct {
		pid int
		fd  int
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
		processes = append(processes, process{pid: pid, fd: fd})
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
		go func() {
			delay := m.managerOptions().RuntimeCleanupDelay
			for {
				time.Sleep(delay)
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
		}()
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

func (m *Manager) buildConfig(instance *Instance, image *Image, options RunOptions) *configs.Config {
	namespaces := configs.Namespaces([]configs.Namespace{
		{Type: configs.NEWNS},
		{Type: configs.NEWUTS},
		{Type: configs.NEWIPC},
		{Type: configs.NEWPID},
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
			Flags:       unix.MS_NOSUID | unix.MS_STRICTATIME,
			Data:        "mode=755",
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
			Source:      "sysfs",
			Destination: "/sys",
			Device:      "sysfs",
			Flags:       unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_NODEV | unix.MS_RDONLY,
		},
	}
	for _, mount := range options.Mounts {
		flags := unix.MS_BIND | unix.MS_REC
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
	resources := &cgroups.Resources{
		Memory:    instance.Resources.Memory,
		CpuQuota:  instance.Resources.CPUQuota,
		CpuPeriod: instance.Resources.CPUPeriod,
	}
	if instance.Resources.PidsLimit > 0 {
		limit := instance.Resources.PidsLimit
		resources.PidsLimit = &limit
	}
	return &configs.Config{
		Rootfs: func() string {
			if instance.Rootfs != "" {
				return instance.Rootfs
			}
			return image.Rootfs
		}(),
		Readonlyfs: !instance.WritableLayer,
		Hostname:   instance.ID,
		Namespaces: namespaces,
		// 安装目录哈希隔离不同 Manager，实例 ID 隔离同一 Manager 内的容器。
		Cgroups: &cgroups.Cgroup{
			Path:      filepath.Join("/sinking-cloud", m.installation, instance.ID),
			Resources: resources,
		},
		Devices: m.defaultDevices(),
		Mounts:  mounts,
		MaskPaths: []string{
			"/proc/kcore",
			"/proc/keys",
			"/proc/latency_stats",
			"/proc/timer_list",
			"/sys/firmware",
			"/sys/devices/system/cpu/cpu0/thermal_throttle",
		},
		ReadonlyPaths: []string{
			"/proc/asound",
			"/proc/acpi",
			"/proc/kcore",
			"/proc/keys",
			"/proc/latency_stats",
			"/proc/timer_list",
			"/proc/sys",
			"/proc/sysrq-trigger",
			"/proc/irq",
			"/proc/bus",
		},
	}
}

func (m *Manager) resolveUser(rootfs, value string) (int, int, []int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, 0, nil, nil
	}
	passwdPath, err := m.safeRootfsFile(rootfs, filepath.Join("etc", "passwd"))
	if err != nil {
		return 0, 0, nil, err
	}
	groupPath, err := m.safeRootfsFile(rootfs, filepath.Join("etc", "group"))
	if err != nil {
		return 0, 0, nil, err
	}
	user, err := userUtil.GetExecUserPath(
		value,
		&userUtil.ExecUser{Uid: 0, Gid: 0},
		passwdPath,
		groupPath,
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

func (m *Manager) safeRootfsFile(rootfs, name string) (string, error) {
	path, err := m.safeJoin(rootfs, name)
	if err != nil {
		return "", err
	}
	info, err := m.rootfsPathInfo(rootfs, path)
	if errors.Is(err, os.ErrNotExist) {
		return path, nil
	}
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("容器用户文件不是普通文件: %s", path)
	}
	return path, nil
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

// mergeEnv 按组依次覆盖同名变量，保持 image < instance < exec 的优先级。
func (m *Manager) mergeEnv(groups ...[]string) []string {
	result := make([]string, 0)
	positions := make(map[string]int)
	for _, group := range groups {
		for _, value := range group {
			key, _, ok := strings.Cut(value, "=")
			if !ok || key == "" {
				continue
			}
			if position, exists := positions[key]; exists {
				result[position] = value
				continue
			}
			positions[key] = len(result)
			result = append(result, value)
		}
	}
	return result
}

func (m *Manager) defaultDevices() []*devices.Device {
	return []*devices.Device{
		{Path: "/dev/null", FileMode: 0666, Rule: devices.Rule{Type: devices.CharDevice, Major: 1, Minor: 3, Permissions: "rwm", Allow: true}},
		{Path: "/dev/zero", FileMode: 0666, Rule: devices.Rule{Type: devices.CharDevice, Major: 1, Minor: 5, Permissions: "rwm", Allow: true}},
		{Path: "/dev/full", FileMode: 0666, Rule: devices.Rule{Type: devices.CharDevice, Major: 1, Minor: 7, Permissions: "rwm", Allow: true}},
		{Path: "/dev/random", FileMode: 0666, Rule: devices.Rule{Type: devices.CharDevice, Major: 1, Minor: 8, Permissions: "rwm", Allow: true}},
		{Path: "/dev/urandom", FileMode: 0666, Rule: devices.Rule{Type: devices.CharDevice, Major: 1, Minor: 9, Permissions: "rwm", Allow: true}},
		{Path: "/dev/tty", FileMode: 0666, Rule: devices.Rule{Type: devices.CharDevice, Major: 5, Minor: 0, Permissions: "rwm", Allow: true}},
	}
}

func (m *Manager) prepareMountTargets(rootfs string, mounts []Mount, create bool) error {
	for _, mount := range mounts {
		if err := m.applyMountPermissions(mount); err != nil {
			return err
		}
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
	}
	return nil
}

// applyMountPermissions 按需统一 bind mount 源权限；空权限不修改源目录。
// 遍历时不跟随符号链接，防止站点路径逃逸。
func (m *Manager) applyMountPermissions(mount Mount) error {
	if mount.DirectoryMode == "" && mount.FileMode == "" {
		return nil
	}
	directoryMode, err := m.parseMountMode(mount.DirectoryMode)
	if err != nil {
		return fmt.Errorf("解析挂载目录权限失败: %w", err)
	}
	fileMode, err := m.parseMountMode(mount.FileMode)
	if err != nil {
		return fmt.Errorf("解析挂载文件权限失败: %w", err)
	}
	info, err := os.Lstat(mount.Source)
	if err != nil {
		return fmt.Errorf("读取挂载源失败: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("挂载源不能是符号链接: %s", mount.Source)
	}
	if !info.IsDir() {
		if fileMode != 0 && info.Mode().IsRegular() {
			if err := os.Chmod(mount.Source, fileMode); err != nil {
				return fmt.Errorf("设置挂载文件权限失败: %w", err)
			}
		}
		return nil
	}
	return filepath.WalkDir(mount.Source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		mode := fileMode
		if entry.IsDir() {
			mode = directoryMode
		}
		if mode == 0 {
			return nil
		}
		if err := os.Chmod(path, mode); err != nil {
			return fmt.Errorf("设置挂载路径权限失败 %s: %w", path, err)
		}
		return nil
	})
}

func (m *Manager) parseMountMode(value string) (os.FileMode, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	value = strings.TrimPrefix(value, "0o")
	value = strings.TrimPrefix(value, "0O")
	value = strings.TrimPrefix(value, "0")
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseUint(value, 8, 12)
	if err != nil || parsed > 07777 {
		return 0, fmt.Errorf("权限必须是 0000-07777 的八进制值")
	}
	return os.FileMode(parsed), nil
}
