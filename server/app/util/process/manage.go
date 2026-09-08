package process

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// NewManager 创建一个空的进程保活管理器。
func NewManager() *Manager {
	return &Manager{processes: make(map[string]*managedProcess)}
}

// Start 注册并启动一个进程。已处于期望运行状态的同名进程不会被覆盖。
func (m *Manager) Start(config Config) (*Status, error) {
	value, err := m.normalizeConfig(config)
	if err != nil {
		return nil, err
	}
	m.operationMu.Lock()
	defer m.operationMu.Unlock()

	m.mu.Lock()
	defer m.mu.Unlock()
	entry := m.processes[value.ID]
	if entry != nil && (entry.desired || entry.command != nil || entry.restartTimer != nil) {
		return m.statusOf(entry), fmt.Errorf("进程已在管理中: %s", value.ID)
	}
	if entry == nil {
		entry = &managedProcess{exitCode: -1}
		m.processes[value.ID] = entry
	}
	m.prepareStartLocked(entry, value)
	err = m.startProcessLocked(entry)
	return m.statusOf(entry), err
}

// Stop 停止指定进程；尚未同步的 ID 也会保留停止意图，防止后续同步自动启动。
func (m *Manager) Stop(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("进程 ID 不能为空")
	}
	m.operationMu.Lock()
	defer m.operationMu.Unlock()
	m.mu.Lock()
	if m.processes[id] == nil {
		m.processes[id] = &managedProcess{config: Config{ID: id}, state: StateStopped, exitCode: -1}
	}
	m.mu.Unlock()
	return m.stopByID(id)
}

// Restart 使用已保存的配置停止并重新启动指定进程。
func (m *Manager) Restart(id string) (*Status, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("进程 ID 不能为空")
	}
	m.operationMu.Lock()
	defer m.operationMu.Unlock()

	m.mu.RLock()
	entry := m.processes[id]
	if entry == nil {
		m.mu.RUnlock()
		return nil, fmt.Errorf("进程不存在: %s", id)
	}
	config := m.cloneConfig(entry.config)
	m.mu.RUnlock()
	if config.Command == "" {
		return nil, fmt.Errorf("进程尚未配置启动命令: %s", id)
	}
	if err := m.stopEntry(entry); err != nil {
		status, _ := m.statusByID(id)
		return status, err
	}

	m.mu.Lock()
	m.prepareStartLocked(entry, config)
	err := m.startProcessLocked(entry)
	status := m.statusOf(entry)
	m.mu.Unlock()
	return status, err
}

// Sync 将管理器收敛到 configs 描述的期望集合。
// 手动停止或重试耗尽的进程只更新配置，需显式 Start 或 Restart 才会重新启动。
func (m *Manager) Sync(configs []Config) error {
	desired := make(map[string]Config, len(configs))
	for _, config := range configs {
		value, err := m.normalizeConfig(config)
		if err != nil {
			return err
		}
		if _, exists := desired[value.ID]; exists {
			return fmt.Errorf("进程 ID 重复: %s", value.ID)
		}
		desired[value.ID] = value
	}

	m.operationMu.Lock()
	defer m.operationMu.Unlock()
	var result error

	m.mu.RLock()
	existingIDs := make([]string, 0, len(m.processes))
	for id := range m.processes {
		existingIDs = append(existingIDs, id)
	}
	m.mu.RUnlock()
	sort.Strings(existingIDs)
	for _, id := range existingIDs {
		if _, keep := desired[id]; keep {
			continue
		}
		if err := m.stopByID(id); err != nil {
			result = errors.Join(result, fmt.Errorf("停止进程 %s 失败: %w", id, err))
			continue
		}
		m.mu.Lock()
		delete(m.processes, id)
		m.mu.Unlock()
	}

	desiredIDs := make([]string, 0, len(desired))
	for id := range desired {
		desiredIDs = append(desiredIDs, id)
	}
	sort.Strings(desiredIDs)
	for _, id := range desiredIDs {
		config := desired[id]
		m.mu.Lock()
		entry := m.processes[id]
		if entry != nil && entry.suspended {
			if entry.command == nil {
				entry.config = m.cloneConfig(config)
			}
			m.mu.Unlock()
			continue
		}
		same := entry != nil && m.configsEqual(entry.config, config)
		healthy := entry != nil && entry.desired && (entry.command != nil || entry.restartTimer != nil)
		occupied := entry != nil && (entry.desired || entry.command != nil || entry.restartTimer != nil)
		if occupied && !(same && healthy) {
			// 在锁内确定配置重启，避免退出回收同时耗尽重试后又被本次同步唤醒。
			entry.desired = false
		}
		m.mu.Unlock()
		if same && healthy {
			continue
		}
		if occupied {
			if err := m.stopEntry(entry); err != nil {
				result = errors.Join(result, fmt.Errorf("重启前停止进程 %s 失败: %w", id, err))
				continue
			}
		}
		m.mu.Lock()
		if entry == nil {
			entry = &managedProcess{exitCode: -1}
			m.processes[id] = entry
		}
		m.prepareStartLocked(entry, config)
		err := m.startProcessLocked(entry)
		m.mu.Unlock()
		if err != nil && !config.AutoRestart {
			result = errors.Join(result, fmt.Errorf("启动进程 %s 失败: %w", id, err))
		}
	}
	return result
}

// StopAll 停止全部受管进程并取消所有自动重启。
func (m *Manager) StopAll() error {
	m.operationMu.Lock()
	defer m.operationMu.Unlock()

	m.mu.Lock()
	entries := make([]*managedProcess, 0, len(m.processes))
	for _, entry := range m.processes {
		if entry.suspended && entry.command == nil && entry.restartTimer == nil {
			continue
		}
		entry.desired = false
		if entry.restartTimer != nil {
			entry.restartTimer.Stop()
			entry.restartTimer = nil
		}
		entries = append(entries, entry)
	}
	m.mu.Unlock()
	sort.Slice(entries, func(i, j int) bool { return entries[i].config.ID < entries[j].config.ID })
	results := make([]error, len(entries))
	group := sync.WaitGroup{}
	group.Add(len(entries))
	for index, entry := range entries {
		go func() {
			defer group.Done()
			if err := m.stopEntry(entry); err != nil {
				results[index] = fmt.Errorf("停止进程 %s 失败: %w", entry.config.ID, err)
			}
		}()
	}
	group.Wait()
	return errors.Join(results...)
}

// Status 返回指定进程的状态快照。
func (m *Manager) Status(id string) (*Status, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("进程 ID 不能为空")
	}
	return m.statusByID(id)
}
