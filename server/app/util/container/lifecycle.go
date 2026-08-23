package container

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func (m *Manager) loadMetadata() error {
	imageEntries, err := os.ReadDir(m.imagesRoot)
	if err != nil {
		return err
	}
	for _, entry := range imageEntries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		metadataPath := filepath.Join(m.imagesRoot, entry.Name())
		if m.validateManagedNode(metadataPath, false, false) != nil {
			continue
		}
		var image Image
		if err := m.readJSON(metadataPath, &image); err != nil {
			continue
		}
		if m.validateImageID(image.ID) != nil {
			continue
		}
		digest := strings.TrimPrefix(image.ID, "sha256:")
		if entry.Name() != digest+".json" {
			continue
		}
		image.Rootfs = filepath.Join(m.imagesRoot, digest)
		info, statErr := os.Lstat(image.Rootfs)
		if statErr == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			m.images[image.ID] = &image
		}
	}
	instanceEntries, err := os.ReadDir(m.instancesRoot)
	if err != nil {
		return err
	}
instances:
	for _, entry := range instanceEntries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".deleted-") {
			_ = os.RemoveAll(filepath.Join(m.instancesRoot, entry.Name()))
			continue
		}
		instanceDir := filepath.Join(m.instancesRoot, entry.Name())
		metadataPath := filepath.Join(instanceDir, "instance.json")
		if m.validateManagedNode(instanceDir, true, false) != nil || m.validateManagedNode(metadataPath, false, false) != nil {
			continue
		}
		var instance Instance
		if err := m.readJSON(metadataPath, &instance); err != nil {
			continue
		}
		if instance.ID == "" || instance.ID != entry.Name() || m.validateID(instance.ID) != nil {
			continue
		}
		if m.ensureInside(m.instancesRoot, instanceDir) != nil {
			continue
		}
		instance.LogPath = filepath.Join(instanceDir, "container.log")
		if m.validateManagedNode(instance.LogPath, false, true) != nil {
			continue
		}
		if instance.WritableLayer {
			instance.Rootfs = filepath.Join(instanceDir, "rootfs")
			instance.UpperDir = filepath.Join(instanceDir, "upper")
			instance.WorkDir = filepath.Join(instanceDir, "work")
			for _, path := range []string{instance.Rootfs, instance.UpperDir, instance.WorkDir} {
				if m.validateManagedNode(path, true, true) != nil {
					continue instances
				}
			}
		} else {
			image := m.images[instance.ImageID]
			if image == nil {
				continue
			}
			instance.Rootfs = image.Rootfs
			instance.UpperDir = ""
			instance.WorkDir = ""
		}
		if instance.Generation > m.generation {
			m.generation = instance.Generation
		}
		m.instances[instance.ID] = &instance
	}
	return nil
}

func (m *Manager) reconcileRuntimeStates() error {
	entries, err := os.ReadDir(m.runtimeRoot)
	if err != nil {
		return fmt.Errorf("读取容器运行状态目录失败: %w", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(m.runtimeRoot, name)
		if err := m.ensureInside(m.runtimeRoot, path); err != nil {
			return fmt.Errorf("容器运行状态路径不安全: %w", err)
		}
		info, statErr := os.Lstat(path)
		m.mu.RLock()
		instance := m.instances[name]
		m.mu.RUnlock()
		valid := m.validateID(name) == nil && statErr == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0
		if instance != nil && valid {
			continue
		}
		if statErr != nil {
			return fmt.Errorf("读取孤立容器运行状态 %s 失败: %w", name, statErr)
		}
		if !valid {
			return fmt.Errorf("孤立容器运行状态目录无效，已保留等待人工处理: %s", path)
		}
		if cleanupErr := m.cleanupOrphanPlatformRuntime(name); cleanupErr == nil {
			if _, statErr := os.Lstat(path); errors.Is(statErr, os.ErrNotExist) {
				continue
			} else if statErr != nil {
				return fmt.Errorf("确认孤立容器运行状态清理结果失败: %w", statErr)
			}
			return fmt.Errorf("孤立容器运行状态清理完成后目录仍然存在: %s", path)
		} else {
			return fmt.Errorf("无法安全清理孤立容器运行状态 %s，状态已保留等待人工处理: %w", name, cleanupErr)
		}
	}
	return nil
}

func (m *Manager) recoverInstances() error {
	ids := make([]string, 0, len(m.instances))
	for id := range m.instances {
		ids = append(ids, id)
	}
	for _, id := range ids {
		unlock := m.lockInstance(id)
		m.mu.RLock()
		instance := m.cloneInstance(m.instances[id])
		image := m.cloneImage(m.images[instance.ImageID])
		m.mu.RUnlock()
		hasRuntimeState, runtimeErr := m.hasRuntimeState(id)
		securityErr := m.validateInstanceSecurity(instance)
		if securityErr == nil {
			if image == nil {
				securityErr = errors.New("实例引用的镜像不存在")
			} else if err := m.validateImageSecurity(image); err != nil {
				securityErr = fmt.Errorf("实例引用的镜像不安全: %w", err)
			}
		}
		if securityErr == nil {
			options := m.optionsFromInstance(instance)
			if err := m.normalizeRunOptions(&options); err != nil {
				securityErr = fmt.Errorf("实例配置不安全: %w", err)
			} else {
				for _, mount := range options.Mounts {
					if mount.ReadOnly {
						if err := m.validatePlatformMountContent(mount.Source); err != nil {
							securityErr = fmt.Errorf("实例只读挂载不安全: %w", err)
							break
						}
					} else {
						if err := m.validatePlatformMountOwnership(mount.Source); err != nil {
							securityErr = fmt.Errorf("实例可写挂载不安全: %w", err)
							break
						}
					}
				}
				instance.Mounts = options.Mounts
				instance.Resources = options.Resources
			}
		}
		if securityErr != nil {
			if runtimeErr != nil {
				unlock()
				return fmt.Errorf("旧实例 %s 的不安全运行状态无法确认: %w", id, runtimeErr)
			}
			if !hasRuntimeState && (m.isActiveStatus(instance.Status) || instance.PID > 0) {
				unlock()
				return fmt.Errorf("旧实例 %s 标记为运行中但缺少可验证的运行状态，已拒绝继续启动", id)
			}
			if hasRuntimeState {
				if err := m.cleanupOrphanPlatformRuntime(id); err != nil {
					unlock()
					return fmt.Errorf("停止旧实例 %s 失败，已拒绝继续启动: %w", id, err)
				}
			}
			if err := m.cleanupRuntimeRootfs(instance); err != nil {
				unlock()
				return fmt.Errorf("清理旧实例 %s 文件系统失败: %w", id, err)
			}
			instance.Status = StatusFailed
			instance.PID = 0
			instance.EndedAt = time.Now().Unix()
			instance.AutoRestart = false
			instance.Error = securityErr.Error()
			m.storeInstance(instance)
			if err := m.writeJSON(filepath.Join(m.instancesRoot, id, "instance.json"), instance); err != nil {
				unlock()
				return fmt.Errorf("保存旧实例安全状态失败: %w", err)
			}
			unlock()
			continue
		}
		if runtimeErr != nil {
			instance.Status = StatusUnknown
			instance.Error = runtimeErr.Error()
			m.storeInstance(instance)
			if writeErr := m.writeJSON(filepath.Join(m.instancesRoot, instance.ID, "instance.json"), instance); writeErr != nil {
				unlock()
				return fmt.Errorf("保存实例恢复状态失败: %w", writeErr)
			}
			unlock()
			continue
		}
		if instance == nil || !m.isActiveStatus(instance.Status) && instance.PID <= 0 && !hasRuntimeState {
			unlock()
			continue
		}
		previousStatus, previousError := instance.Status, instance.Error
		shouldAutoRestart := false
		running, err := m.recoverPlatformRuntime(instance)
		if errors.Is(err, errRuntimeCleanupPending) {
			m.mu.RLock()
			current := m.cloneInstance(m.instances[id])
			m.mu.RUnlock()
			if current != nil && current.Generation == instance.Generation {
				instance = current
			}
			instance.Status = StatusStopping
			instance.PID = 0
			instance.Error = err.Error()
		} else if err != nil {
			instance.Status = StatusUnknown
			instance.Error = "恢复运行状态失败: " + err.Error()
		} else if running {
			instance.Status = StatusRunning
			instance.Error = ""
		} else {
			shouldAutoRestart = instance.AutoRestart && (previousStatus == StatusRunning || previousStatus == StatusStarting)
			instance.Status = StatusStopped
			instance.PID = 0
			instance.EndedAt = time.Now().Unix()
			instance.Error = ""
			if previousStatus == StatusFailed {
				instance.Status = StatusFailed
				instance.Error = previousError
			}
			if cleanupErr := m.cleanupRuntimeRootfs(instance); cleanupErr != nil {
				shouldAutoRestart = false
				instance.Status = StatusFailed
				if instance.Error != "" {
					cleanupErr = errors.Join(errors.New(instance.Error), cleanupErr)
				}
				instance.Error = cleanupErr.Error()
			}
		}
		m.storeInstance(instance)
		if writeErr := m.writeJSON(filepath.Join(m.instancesRoot, instance.ID, "instance.json"), instance); writeErr != nil {
			unlock()
			return fmt.Errorf("保存实例恢复状态失败: %w", writeErr)
		}
		if shouldAutoRestart && instance.Status == StatusStopped {
			m.scheduleAutoRestart(instance.ID, instance.Generation)
		}
		unlock()
	}
	return nil
}

// Import 将 Docker save 归档导入本地镜像存储。
func (m *Manager) resolveImage(id string) (*Image, error) {
	id = strings.TrimSpace(id)
	if image := m.images[id]; image != nil {
		return image, nil
	}
	keys := make([]string, 0, len(m.images))
	for key := range m.images {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var matched *Image
	for _, key := range keys {
		image := m.images[key]
		match := image.Name == id
		for _, tag := range image.Tags {
			match = match || tag == id
		}
		if !match {
			continue
		}
		if matched != nil && matched.ID != image.ID {
			return nil, fmt.Errorf("镜像名称不唯一，请使用镜像ID: %s", id)
		}
		matched = image
	}
	if matched != nil {
		return matched, nil
	}
	return nil, fmt.Errorf("镜像不存在: %s", id)
}
func (m *Manager) runLocked(options RunOptions) (*Instance, error) {
	m.mountMu.Lock()
	defer m.mountMu.Unlock()
	m.imageMu.RLock()
	defer m.imageMu.RUnlock()
	m.mu.RLock()
	for id, instance := range m.instances {
		_, hasRuntime := m.runtime[id]
		if id == options.ID || !m.isActiveStatus(instance.Status) && instance.PID <= 0 && !hasRuntime {
			continue
		}
		for _, activeMount := range instance.Mounts {
			if activeMount.ReadOnly {
				continue
			}
			for _, mount := range options.Mounts {
				if m.ensureInside(activeMount.Source, mount.Source) == nil || m.ensureInside(mount.Source, activeMount.Source) == nil {
					m.mu.RUnlock()
					return nil, fmt.Errorf("挂载源与运行实例 %s 的可写目录重叠", id)
				}
			}
		}
	}
	image, err := m.resolveImage(options.ImageID)
	if err == nil {
		image = m.cloneImage(image)
	}
	existing := m.cloneInstance(m.instances[options.ID])
	_, hasRuntime := m.runtime[options.ID]
	m.mu.RUnlock()
	hasRuntimeState, stateErr := m.hasRuntimeState(options.ID)
	if stateErr != nil {
		return nil, stateErr
	}
	if err != nil {
		return nil, err
	}
	if err := m.validateImageSecurity(image); err != nil {
		return nil, err
	}
	if existing != nil {
		if err := m.validateInstanceSecurity(existing); err != nil {
			return nil, err
		}
	}
	if existing == nil && hasRuntimeState {
		return nil, fmt.Errorf("实例运行状态已存在但元数据缺失: %s", options.ID)
	}
	if existing != nil && (m.isActiveStatus(existing.Status) || existing.PID > 0 || hasRuntime || hasRuntimeState) {
		existing, err = m.refreshInstanceLocked(options.ID)
		if err != nil {
			return nil, err
		}
	}
	if existing != nil && (m.isActiveStatus(existing.Status) || existing.PID > 0) {
		return nil, fmt.Errorf("实例仍在运行或正在停止: %s", options.ID)
	}

	instanceDir := filepath.Join(m.instancesRoot, options.ID)
	if err := m.ensureInside(m.instancesRoot, instanceDir); err != nil {
		return nil, fmt.Errorf("实例目录不安全: %w", err)
	}
	if existing == nil {
		if _, err := os.Lstat(instanceDir); err == nil {
			return nil, fmt.Errorf("实例目录已存在但缺少有效元数据，请先备份并删除孤立实例: %s", options.ID)
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("读取实例目录失败: %w", err)
		}
	}
	if err := m.ensureManagedDirectory(instanceDir); err != nil {
		return nil, fmt.Errorf("创建实例目录失败: %w", err)
	}
	if err := m.validateManagedNode(filepath.Join(instanceDir, "container.log"), false, true); err != nil {
		return nil, fmt.Errorf("实例日志路径不安全: %w", err)
	}
	for _, name := range []string{"rootfs", "upper", "work"} {
		if err := m.validateManagedNode(filepath.Join(instanceDir, name), true, true); err != nil {
			return nil, fmt.Errorf("实例可写层路径不安全: %w", err)
		}
	}

	createdAt := time.Now().Unix()
	if existing != nil && existing.CreatedAt > 0 {
		createdAt = existing.CreatedAt
		if err := m.cleanupRuntimeRootfs(existing); err != nil {
			return nil, err
		}
	}
	if existing != nil && existing.ImageID != image.ID {
		if err := m.resetWritableLayer(instanceDir); err != nil {
			return nil, err
		}
	}
	instance := &Instance{
		ID:              options.ID,
		Name:            strings.TrimSpace(options.Name),
		ImageID:         image.ID,
		Status:          StatusStarting,
		Mounts:          append([]Mount(nil), options.Mounts...),
		CreatedAt:       createdAt,
		LogPath:         filepath.Join(instanceDir, "container.log"),
		Command:         append([]string(nil), options.Command...),
		Env:             append([]string(nil), options.Env...),
		WorkingDir:      options.WorkingDir,
		Resources:       options.Resources,
		AutoRestart:     options.AutoRestart,
		WritableLayer:   !options.ReadOnly,
		Generation:      m.nextGeneration(),
		SecurityVersion: m.security.Version,
		UIDMapStart:     m.security.UIDMapStart,
		GIDMapStart:     m.security.GIDMapStart,
		IDMapSize:       m.security.IDMapSize,
	}
	if instance.Name == "" {
		instance.Name = image.Name
		if instance.Name == "" && len(image.Tags) > 0 {
			instance.Name = image.Tags[0]
		}
	}
	if existing != nil && existing.ImageID == image.ID && existing.WritableLayer && instance.WritableLayer {
		instance.Rootfs = existing.Rootfs
		instance.UpperDir = existing.UpperDir
		instance.WorkDir = existing.WorkDir
	}
	m.storeInstance(instance)
	if err := m.prepareRuntimeRootfs(instance, image, instance.WritableLayer); err != nil {
		return nil, m.failStart(instance, err)
	}
	m.storeInstance(instance)
	if err := m.writeJSON(filepath.Join(instanceDir, "instance.json"), instance); err != nil {
		return nil, m.failStart(instance, fmt.Errorf("保存实例配置失败: %w", err))
	}
	if err := m.startRuntime(instance, image, options); err != nil {
		return nil, m.failStart(instance, err)
	}
	instance.Status = StatusRunning
	instance.StartedAt = time.Now().Unix()
	instance.EndedAt = 0
	instance.ExitCode = 0
	instance.Error = ""
	if err := m.writeJSON(filepath.Join(instanceDir, "instance.json"), instance); err != nil {
		failure := fmt.Errorf("保存运行状态失败: %w", err)
		m.setPlatformFailure(instance.ID, failure)
		stopErr := m.stopRuntime(instance)
		cleanupErr := m.cleanupRuntimeRootfs(instance)
		instance.PID = 0
		instance.Status = StatusFailed
		instance.Error = failure.Error()
		instance.EndedAt = time.Now().Unix()
		m.storeInstance(instance)
		persistErr := m.writeJSON(filepath.Join(instanceDir, "instance.json"), instance)
		return nil, errors.Join(failure, stopErr, cleanupErr, persistErr)
	}
	m.storeInstance(instance)
	return m.cloneInstance(instance), nil
}

// Start 使用已保存配置启动一个停止的实例。
func (m *Manager) stopLocked(id string) error {
	if _, err := m.refreshInstanceLocked(id); err != nil {
		return err
	}
	m.mu.Lock()
	current := m.instances[id]
	if current == nil {
		m.mu.Unlock()
		return errors.New("实例不存在")
	}
	previousStatus := current.Status
	_, hasRuntime := m.runtime[id]
	if !m.isActiveStatus(current.Status) && current.PID <= 0 && !hasRuntime {
		m.mu.Unlock()
		return nil
	}
	current.Status = StatusStopping
	instance := m.cloneInstance(current)
	generation := current.Generation
	m.mu.Unlock()
	metadataPath := filepath.Join(m.instancesRoot, id, "instance.json")
	if err := m.writeJSON(metadataPath, instance); err != nil {
		m.mu.Lock()
		if current = m.instances[id]; current != nil && current.Generation == generation {
			current.Status = previousStatus
		}
		m.mu.Unlock()
		return fmt.Errorf("保存实例停止状态失败: %w", err)
	}
	if err := m.stopRuntime(instance); err != nil {
		m.mu.Lock()
		if current = m.instances[id]; current != nil && current.Generation == generation {
			if errors.Is(err, errRuntimeCleanupPending) {
				current.Status = StatusStopping
				current.PID = 0
			} else {
				current.Status = previousStatus
			}
			current.Error = err.Error()
			instance = m.cloneInstance(current)
		}
		m.mu.Unlock()
		_ = m.writeJSON(metadataPath, instance)
		return err
	}
	cleanupErr := m.cleanupRuntimeRootfs(instance)
	m.mu.Lock()
	if current = m.instances[id]; current != nil && current.Generation == generation {
		current.PID = 0
		current.EndedAt = time.Now().Unix()
		if cleanupErr != nil {
			current.Status = StatusFailed
			current.Error = cleanupErr.Error()
		} else {
			current.Status = StatusStopped
			current.Error = ""
		}
		delete(m.runtime, id)
		instance = m.cloneInstance(current)
	}
	m.mu.Unlock()
	persistErr := m.writeJSON(metadataPath, instance)
	return errors.Join(cleanupErr, persistErr)
}

// Restart 停止后按原配置重新启动实例。
func (m *Manager) markExitedLocked(id string, generation uint64, exitCode int, waitErr error, cleanup bool, failure error) {
	m.mu.RLock()
	instance := m.cloneInstance(m.instances[id])
	if instance == nil || instance.Generation != generation {
		m.mu.RUnlock()
		return
	}
	m.mu.RUnlock()
	var cleanupErr error
	if cleanup {
		cleanupErr = m.cleanupRuntimeRootfs(instance)
	}
	instance.ExitCode = exitCode
	instance.PID = 0
	if !cleanup {
		instance.Status = StatusStopping
		instance.Error = "容器运行状态清理失败"
		if waitErr != nil {
			instance.Error = waitErr.Error()
		}
	} else if failure != nil {
		// 启动事务的原始失败原因不能被迟到的清理结果覆盖。
		instance.Status = StatusFailed
		instance.Error = errors.Join(failure, cleanupErr).Error()
	} else if cleanupErr != nil {
		instance.Status = StatusFailed
		instance.Error = errors.Join(waitErr, cleanupErr).Error()
	} else if waitErr != nil {
		instance.Error = waitErr.Error()
		instance.Status = StatusFailed
	} else {
		instance.Status = StatusStopped
		instance.Error = ""
	}
	instance.EndedAt = time.Now().Unix()
	m.storeInstance(instance)
	_ = m.writeJSON(filepath.Join(m.instancesRoot, id, "instance.json"), instance)
}

// scheduleAutoRestart 在异常退出后的短暂延迟后重新启动实例。
// 通过实例代次和待重启标记校验，避免旧进程退出事件覆盖用户后续的手动操作。
func (m *Manager) scheduleAutoRestart(id string, generation uint64) {
	m.mu.Lock()
	instance := m.instances[id]
	if instance == nil || instance.Generation != generation || !instance.AutoRestart {
		m.mu.Unlock()
		return
	}
	if m.autoRestart == nil {
		m.autoRestart = make(map[string]uint64)
	}
	m.autoRestart[id] = generation
	delay := m.managerOptions().AutoRestartDelay
	m.mu.Unlock()

	go func() {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		<-timer.C

		unlock := m.lockInstance(id)
		defer unlock()
		m.mu.Lock()
		pendingGeneration, pending := m.autoRestart[id]
		if pending {
			delete(m.autoRestart, id)
		}
		instance = m.cloneInstance(m.instances[id])
		_, hasRuntime := m.runtime[id]
		m.mu.Unlock()
		if !pending || pendingGeneration != generation || instance == nil || instance.Generation != generation ||
			!instance.AutoRestart || (instance.Status != StatusStopped && instance.Status != StatusFailed) || instance.PID > 0 || hasRuntime {
			return
		}
		hasRuntimeState, err := m.hasRuntimeState(id)
		if err != nil || hasRuntimeState {
			return
		}
		options := m.optionsFromInstance(instance)
		if err := m.normalizeRunOptions(&options); err != nil {
			return
		}
		_, _ = m.runLocked(options)
	}()
}

// cancelAutoRestart 取消尚未执行的自动重启，不改变实例的持久化配置。
func (m *Manager) cancelAutoRestart(id string) {
	m.mu.Lock()
	delete(m.autoRestart, id)
	m.mu.Unlock()
}

// markRuntimeCleanupLocked 先持久化清理意图，避免进程崩溃后把已删除的 runtime 误判为仍在运行。
func (m *Manager) markRuntimeCleanupLocked(id string, generation uint64, exitCode int, waitErr error) error {
	m.mu.Lock()
	instance := m.instances[id]
	if instance == nil || instance.Generation != generation {
		m.mu.Unlock()
		return errors.New("实例运行代次已变化")
	}
	updated := m.cloneInstance(instance)
	updated.Status = StatusStopping
	updated.PID = 0
	updated.ExitCode = exitCode
	updated.EndedAt = time.Now().Unix()
	if waitErr != nil {
		updated.Error = waitErr.Error()
	}
	m.instances[id] = m.cloneInstance(updated)
	m.mu.Unlock()
	if err := m.writeJSON(filepath.Join(m.instancesRoot, id, "instance.json"), updated); err != nil {
		m.mu.Lock()
		if current := m.instances[id]; current != nil && current.Generation == generation {
			current.Status = StatusStopping
			current.PID = 0
			current.Error = "保存实例退出状态失败: " + err.Error()
		}
		m.mu.Unlock()
		return errors.Join(errRuntimeCleanupPending, fmt.Errorf("保存实例退出状态失败: %w", err))
	}
	return nil
}

func (m *Manager) refreshInstanceLocked(id string) (*Instance, error) {
	m.mu.RLock()
	instance := m.cloneInstance(m.instances[id])
	_, hasRuntime := m.runtime[id]
	m.mu.RUnlock()
	if instance == nil {
		return nil, errors.New("实例不存在")
	}
	hasRuntimeState, stateErr := m.hasRuntimeState(id)
	if stateErr != nil {
		return nil, stateErr
	}
	if !m.isActiveStatus(instance.Status) && instance.PID <= 0 && !hasRuntime && !hasRuntimeState {
		return instance, nil
	}
	previousStatus, previousPID := instance.Status, instance.PID
	running, err := m.refreshPlatformRuntime(instance)
	if errors.Is(err, errRuntimeCleanupPending) {
		m.mu.RLock()
		current := m.cloneInstance(m.instances[id])
		m.mu.RUnlock()
		if current != nil && current.Generation == instance.Generation {
			instance = current
		}
		instance.Status = StatusStopping
		instance.PID = 0
		instance.Error = err.Error()
	} else if err != nil {
		instance.Status = StatusUnknown
		instance.Error = err.Error()
	} else if running {
		instance.Status = StatusRunning
		instance.Error = ""
	} else {
		instance.Status = StatusStopped
		instance.PID = 0
		instance.EndedAt = time.Now().Unix()
		instance.Error = ""
		if cleanupErr := m.cleanupRuntimeRootfs(instance); cleanupErr != nil {
			instance.Status = StatusFailed
			instance.Error = cleanupErr.Error()
			err = cleanupErr
		}
	}
	m.mu.Lock()
	if current := m.instances[id]; current != nil && current.Generation == instance.Generation {
		m.instances[id] = m.cloneInstance(instance)
		if previousStatus != instance.Status || previousPID != instance.PID || err != nil {
			_ = m.writeJSON(filepath.Join(m.instancesRoot, id, "instance.json"), instance)
		}
	}
	m.mu.Unlock()
	return m.cloneInstance(instance), err
}

func (m *Manager) failStart(instance *Instance, startErr error) error {
	m.setPlatformFailure(instance.ID, startErr)
	var cleanupErr error
	if !errors.Is(startErr, errRuntimeCleanupPending) {
		cleanupErr = m.cleanupRuntimeRootfs(instance)
	}
	result := errors.Join(startErr, cleanupErr)
	instance.Status = StatusFailed
	instance.PID = 0
	instance.Error = result.Error()
	instance.EndedAt = time.Now().Unix()
	m.storeInstance(instance)
	persistErr := m.writeJSON(filepath.Join(m.instancesRoot, instance.ID, "instance.json"), instance)
	return errors.Join(result, persistErr)
}

func (m *Manager) optionsFromInstance(instance *Instance) RunOptions {
	return RunOptions{
		ID:          instance.ID,
		Name:        instance.Name,
		ImageID:     instance.ImageID,
		Mounts:      append([]Mount(nil), instance.Mounts...),
		Env:         append([]string(nil), instance.Env...),
		Command:     append([]string(nil), instance.Command...),
		WorkingDir:  instance.WorkingDir,
		Resources:   instance.Resources,
		AutoRestart: instance.AutoRestart,
		ReadOnly:    !instance.WritableLayer,
	}
}

func (m *Manager) normalizeRunOptions(options *RunOptions) error {
	options.ID = strings.TrimSpace(options.ID)
	options.ImageID = strings.TrimSpace(options.ImageID)
	options.WorkingDir = strings.TrimSpace(options.WorkingDir)
	if options.ImageID == "" {
		return errors.New("镜像不能为空")
	}
	if err := m.validateID(options.ID); err != nil {
		return err
	}
	if options.WorkingDir != "" && !filepath.IsAbs(options.WorkingDir) {
		return errors.New("工作目录必须是容器内绝对路径")
	}
	if len(options.Command) > 0 && strings.TrimSpace(options.Command[0]) == "" {
		return errors.New("容器入口命令不能为空")
	}
	if options.Resources.Memory < 0 || options.Resources.CPUQuota < 0 || options.Resources.PidsLimit < 0 {
		return errors.New("资源限制不能小于0")
	}
	if options.Resources.CPUQuota > 0 && options.Resources.CPUPeriod == 0 {
		options.Resources.CPUPeriod = 100000
	}
	if options.Resources.PidsLimit == 0 && m.managerOptions().DefaultPidsLimit > 0 {
		options.Resources.PidsLimit = m.managerOptions().DefaultPidsLimit
	}
	for _, value := range options.Env {
		key, _, ok := strings.Cut(value, "=")
		if !ok || key == "" || strings.ContainsAny(key, "\x00") {
			return fmt.Errorf("环境变量格式无效: %s", value)
		}
	}
	destinations := make(map[string]struct{}, len(options.Mounts))
	for index := range options.Mounts {
		mount := &options.Mounts[index]
		resolved, err := m.resolveMountSource(mount.Source)
		if err != nil {
			return err
		}
		mount.Source = resolved
		mount.Destination = path.Clean(strings.TrimSpace(mount.Destination))
		if err := m.validateMount(*mount); err != nil {
			return err
		}
		if _, exists := destinations[mount.Destination]; exists {
			return fmt.Errorf("挂载目标重复: %s", mount.Destination)
		}
		destinations[mount.Destination] = struct{}{}
	}
	return nil
}

func (m *Manager) lockInstance(id string) func() {
	hash := uint32(2166136261)
	for index := 0; index < len(id); index++ {
		hash ^= uint32(id[index])
		hash *= 16777619
	}
	lock := &m.operationLocks[hash%instanceLockCount]
	lock.Lock()
	return lock.Unlock
}

func (m *Manager) hasRuntimeState(id string) (bool, error) {
	path := filepath.Join(m.runtimeRoot, id)
	if err := m.ensureInside(m.runtimeRoot, path); err != nil {
		return false, fmt.Errorf("运行状态目录不安全: %w", err)
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("读取运行状态目录失败: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return false, errors.New("运行状态目录无效")
	}
	return true, nil
}

func (m *Manager) nextGeneration() uint64 {
	m.mu.Lock()
	m.generation++
	if m.generation == 0 {
		m.generation++
	}
	generation := m.generation
	m.mu.Unlock()
	return generation
}

func (m *Manager) isActiveStatus(status Status) bool {
	return status == StatusStarting || status == StatusRunning || status == StatusStopping || status == StatusUnknown
}

func (m *Manager) storeInstance(instance *Instance) {
	m.mu.Lock()
	current := m.instances[instance.ID]
	if current == nil || current.Generation <= instance.Generation {
		m.instances[instance.ID] = m.cloneInstance(instance)
	}
	m.mu.Unlock()
}

func (m *Manager) resetWritableLayer(instanceDir string) error {
	if err := m.ensureInside(m.instancesRoot, instanceDir); err != nil {
		return fmt.Errorf("实例目录不安全: %w", err)
	}
	if err := m.validateManagedNode(instanceDir, true, false); err != nil {
		return fmt.Errorf("实例目录不安全: %w", err)
	}
	for _, name := range []string{"rootfs", "upper", "work"} {
		path := filepath.Join(instanceDir, name)
		if err := m.ensureInside(instanceDir, path); err != nil {
			return fmt.Errorf("实例可写层路径不安全: %w", err)
		}
		if err := m.validateManagedNode(path, true, true); err != nil {
			return fmt.Errorf("实例可写层路径不安全: %w", err)
		}
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("清理旧实例可写层失败: %w", err)
		}
	}
	return nil
}

func (m *Manager) cleanupOrphanInstanceStorage(id string) error {
	if err := m.validateID(id); err != nil {
		return err
	}
	instanceDir := filepath.Join(m.instancesRoot, id)
	if err := m.ensureInside(m.instancesRoot, instanceDir); err != nil {
		return fmt.Errorf("孤立实例目录不安全: %w", err)
	}
	info, err := os.Lstat(instanceDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("读取孤立实例目录失败: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("孤立实例目录不是受管普通目录: %s", instanceDir)
	}
	rootfs := filepath.Join(instanceDir, "rootfs")
	rootfsInfo, rootfsErr := os.Lstat(rootfs)
	if rootfsErr != nil && !errors.Is(rootfsErr, os.ErrNotExist) {
		return fmt.Errorf("读取孤立实例 rootfs 失败: %w", rootfsErr)
	}
	if rootfsErr == nil && rootfsInfo.IsDir() && rootfsInfo.Mode()&os.ModeSymlink == 0 {
		instance := &Instance{
			ID:            id,
			WritableLayer: true,
			Rootfs:        rootfs,
			UpperDir:      filepath.Join(instanceDir, "upper"),
			WorkDir:       filepath.Join(instanceDir, "work"),
		}
		if err := m.cleanupRuntimeRootfs(instance); err != nil {
			return fmt.Errorf("卸载孤立实例 rootfs 失败: %w", err)
		}
	}
	tombstone := filepath.Join(m.instancesRoot, fmt.Sprintf(".deleted-%s-%d", id, time.Now().UnixNano()))
	if err := m.ensureInside(m.instancesRoot, tombstone); err != nil {
		return fmt.Errorf("孤立实例删除临时目录不安全: %w", err)
	}
	if err := os.Rename(instanceDir, tombstone); err != nil {
		return fmt.Errorf("标记孤立实例删除失败: %w", err)
	}
	if err := os.RemoveAll(tombstone); err != nil {
		return fmt.Errorf("清理孤立实例目录失败: %w", err)
	}
	return nil
}
