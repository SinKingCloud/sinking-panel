package container

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// NewManager 创建镜像、实例和运行时状态目录，并恢复已有实例状态。
// options 可选；未提供或字段为零时使用内置默认配置。
func NewManager(root string, options ...ManagerOptions) (*Manager, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("容器数据目录不能为空")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("解析容器数据目录失败: %w", err)
	}
	if info, statErr := os.Lstat(absRoot); statErr == nil {
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
		for index := len(missing) - 1; index >= 0; index-- {
			resolved = filepath.Join(resolved, missing[index])
		}
		absRoot = filepath.Clean(resolved)
	}
	installation := sha256.Sum256([]byte(absRoot))
	managerOptions := ManagerOptions{}
	if len(options) > 0 {
		managerOptions = options[0]
	}
	m := &Manager{
		imagesRoot:    filepath.Join(absRoot, "images"),
		instancesRoot: filepath.Join(absRoot, "instances"),
		runtimeRoot:   filepath.Join(absRoot, "runtime"),
		installation:  fmt.Sprintf("%x", installation[:16]),
		options:       normalizeManagerOptions(managerOptions),
		images:        make(map[string]*Image),
		instances:     make(map[string]*Instance),
		runtime:       make(map[string]interface{}),
		autoRestart:   make(map[string]uint64),
	}
	for _, path := range []string{absRoot, m.imagesRoot, m.instancesRoot, m.runtimeRoot} {
		if err := m.ensureManagedDirectory(path); err != nil {
			return nil, err
		}
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
	return m, nil
}

func (m *Manager) Import(source string) (*Image, error) {
	m.importMu.Lock()
	defer m.importMu.Unlock()
	m.imageMu.Lock()
	defer m.imageMu.Unlock()
	image, err := m.importDockerSave(source)
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

// Image 返回一个镜像的完整元数据。
func (m *Manager) Image(id string) (*Image, error) {
	m.mu.RLock()
	image, err := m.resolveImage(id)
	if err == nil {
		image = m.cloneImage(image)
	}
	m.mu.RUnlock()
	return image, err
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

func (m *Manager) Run(options RunOptions) (*Instance, error) {
	if options.ID == "" {
		options.ID = fmt.Sprintf("container-%d", time.Now().UnixNano())
	}
	if err := m.normalizeRunOptions(&options); err != nil {
		return nil, err
	}
	unlock := m.lockInstance(options.ID)
	defer unlock()
	m.cancelAutoRestart(options.ID)
	return m.runLocked(options)
}

func (m *Manager) Start(id string) (*Instance, error) {
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

// Stop 优雅停止实例，超时后由平台运行时强制结束。
func (m *Manager) Stop(id string) error {
	if err := m.validateID(id); err != nil {
		return err
	}
	unlock := m.lockInstance(id)
	defer unlock()
	m.cancelAutoRestart(id)
	return m.stopLocked(id)
}

func (m *Manager) Restart(id string) (*Instance, error) {
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
			return nil
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

// Instances 返回全部实例，并同步平台运行状态。
func (m *Manager) Instances() []*Instance {
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
		result = append(result, m.cloneInstance(instance))
	}
	m.mu.RUnlock()
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt > result[j].CreatedAt })
	return result
}

// Instance 返回一个实例的最新状态和完整配置。
func (m *Manager) Instance(id string) (*Instance, error) {
	if err := m.validateID(id); err != nil {
		return nil, err
	}
	unlock := m.lockInstance(id)
	defer unlock()
	return m.refreshInstanceLocked(id)
}

// Stats 返回运行实例当前的 CPU、内存和进程数。
func (m *Manager) Stats(id string) (*Stats, error) {
	if err := m.validateID(id); err != nil {
		return nil, err
	}
	unlock := m.lockInstance(id)
	defer unlock()
	instance, err := m.refreshInstanceLocked(id)
	if err != nil {
		return nil, err
	}
	if instance == nil || instance.Status != StatusRunning {
		return nil, errors.New("实例未运行")
	}
	return m.statsPlatform(id)
}

// Shutdown 并发停止全部运行中或仍有运行时资源待清理的实例。
func (m *Manager) Shutdown() error {
	m.mu.RLock()
	ids := make([]string, 0, len(m.instances))
	for id, instance := range m.instances {
		_, hasRuntime := m.runtime[id]
		if m.isActiveStatus(instance.Status) || instance.PID > 0 || hasRuntime {
			ids = append(ids, id)
		}
	}
	m.mu.RUnlock()
	errorsChannel := make(chan error, len(ids))
	var group sync.WaitGroup
	for _, id := range ids {
		group.Add(1)
		go func(instanceID string) {
			defer group.Done()
			if err := m.Stop(instanceID); err != nil {
				errorsChannel <- fmt.Errorf("停止实例 %s 失败: %w", instanceID, err)
			}
		}(id)
	}
	group.Wait()
	close(errorsChannel)
	var result error
	for err := range errorsChannel {
		result = errors.Join(result, err)
	}
	return result
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
	unlock := m.lockInstance(id)
	defer unlock()
	instance, err := m.refreshInstanceLocked(id)
	if err != nil {
		return nil, err
	}
	if instance.Status != StatusRunning {
		return nil, errors.New("实例未运行")
	}
	return m.execPlatform(id, options)
}

// OpenTerminal 在运行实例内启动交互式 shell。关闭会话不会停止容器主进程。
func (m *Manager) OpenTerminal(id string, height, width int) (TerminalSession, error) {
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
	data, err := m.readLogData([]string{path + ".1", path}, 0)
	if err != nil {
		return nil, err
	}
	return readCursorLog(data, filepath.Base(path), after, before, pageSize), nil
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
