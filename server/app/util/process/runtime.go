package process

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/DeRuina/timberjack"
)

func (m *Manager) normalizeConfig(config Config) (Config, error) {
	config.ID = strings.TrimSpace(config.ID)
	config.Command = strings.TrimSpace(config.Command)
	config.WorkingDir = strings.TrimSpace(config.WorkingDir)
	config.LogPath = strings.TrimSpace(config.LogPath)
	config.Env = append([]string(nil), config.Env...)
	if config.ID == "" {
		return Config{}, errors.New("进程 ID 不能为空")
	}
	if config.Command == "" {
		return Config{}, fmt.Errorf("进程命令不能为空: %s", config.ID)
	}
	if config.RestartDelay < 0 {
		return Config{}, fmt.Errorf("进程重启延迟不能小于零: %s", config.ID)
	}
	if config.RestartDelay == 0 {
		config.RestartDelay = defaultRestartDelay
	}
	if config.StopTimeout < 0 {
		return Config{}, fmt.Errorf("进程停止超时不能小于零: %s", config.ID)
	}
	if config.StopTimeout == 0 {
		config.StopTimeout = defaultStopTimeout
	}
	if config.WorkingDir != "" {
		info, err := os.Stat(config.WorkingDir)
		if err != nil {
			return Config{}, fmt.Errorf("读取进程工作目录失败: %w", err)
		}
		if !info.IsDir() {
			return Config{}, fmt.Errorf("进程工作路径不是目录: %s", config.WorkingDir)
		}
	}
	for _, value := range config.Env {
		key, _, ok := strings.Cut(value, "=")
		if !ok || key == "" || strings.ContainsRune(value, '\x00') {
			return Config{}, fmt.Errorf("进程环境变量不合法: %q", value)
		}
	}
	return config, nil
}

func (m *Manager) cloneConfig(config Config) Config {
	config.Env = append([]string(nil), config.Env...)
	return config
}

func (m *Manager) configsEqual(left, right Config) bool {
	return left.ID == right.ID && left.Command == right.Command && left.WorkingDir == right.WorkingDir &&
		left.AutoRestart == right.AutoRestart && left.RestartDelay == right.RestartDelay &&
		left.StopTimeout == right.StopTimeout && left.LogPath == right.LogPath && slices.Equal(left.Env, right.Env)
}

func (m *Manager) statusOf(entry *managedProcess) *Status {
	if entry == nil {
		return nil
	}
	return &Status{
		ID:           entry.config.ID,
		State:        entry.state,
		PID:          entry.pid,
		Command:      entry.config.Command,
		WorkingDir:   entry.config.WorkingDir,
		AutoRestart:  entry.config.AutoRestart,
		LogPath:      entry.config.LogPath,
		StartedAt:    entry.startedAt,
		ExitedAt:     entry.exitedAt,
		ExitCode:     entry.exitCode,
		RestartCount: entry.restartCount,
		Error:        entry.lastError,
	}
}

func (m *Manager) statusByID(id string) (*Status, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry := m.processes[id]
	if entry == nil {
		return nil, fmt.Errorf("进程不存在: %s", id)
	}
	return m.statusOf(entry), nil
}

func (m *Manager) prepareStartLocked(entry *managedProcess, config Config) {
	if entry.restartTimer != nil {
		entry.restartTimer.Stop()
		entry.restartTimer = nil
	}
	entry.config = m.cloneConfig(config)
	entry.desired = true
	entry.generation++
	entry.pid = 0
	entry.exitCode = -1
	entry.startedAt = time.Time{}
	entry.exitedAt = time.Time{}
	entry.restartCount = 0
	entry.lastError = ""
	entry.stopAt = time.Time{}
	entry.state = StateStarting
}

func (m *Manager) startProcessLocked(entry *managedProcess) error {
	if !entry.desired {
		return errors.New("进程已被停止")
	}
	if entry.command != nil {
		return errors.New("进程仍在运行")
	}
	entry.state = StateStarting
	logFile, err := m.openProcessLog(entry.config.LogPath)
	if err != nil {
		return m.handleStartFailureLocked(entry, fmt.Errorf("打开进程日志失败: %w", err))
	}
	command := m.systemShellCommand(entry.config.Command)
	command.Dir = entry.config.WorkingDir
	command.Env = m.mergeEnvironment(os.Environ(), entry.config.Env)
	if logFile != nil {
		command.Stdout = logFile
		command.Stderr = logFile
	}
	grouped := m.configureProcessGroup(command)
	if err := command.Start(); err != nil {
		if logFile != nil {
			_ = logFile.Close()
		}
		return m.handleStartFailureLocked(entry, fmt.Errorf("启动进程失败: %w", err))
	}
	done := make(chan struct{})
	entry.command = command
	entry.done = done
	entry.grouped = grouped
	entry.pid = command.Process.Pid
	entry.state = StateRunning
	entry.startedAt = time.Now()
	entry.exitedAt = time.Time{}
	entry.exitCode = -1
	entry.lastError = ""
	generation := entry.generation
	go m.waitProcess(entry, generation, command, logFile, done, grouped)
	return nil
}

func (m *Manager) handleStartFailureLocked(entry *managedProcess, err error) error {
	entry.pid = 0
	entry.command = nil
	entry.done = nil
	entry.grouped = false
	entry.exitCode = -1
	entry.exitedAt = time.Now()
	entry.lastError = err.Error()
	if entry.desired && entry.config.AutoRestart {
		m.scheduleRestartLocked(entry)
	} else {
		entry.desired = false
		entry.state = StateFailed
	}
	return err
}

func (m *Manager) waitProcess(entry *managedProcess, generation uint64, command *exec.Cmd, logFile io.WriteCloser, done chan struct{}, grouped bool) {
	waitErr := command.Wait()
	var cleanupErr error
	if grouped {
		m.mu.RLock()
		stopAt := entry.stopAt
		timeout := entry.config.StopTimeout
		m.mu.RUnlock()
		cleanupErr = m.cleanupProcessGroup(command.Process, timeout, stopAt)
	}
	if logFile != nil {
		_ = logFile.Close()
	}
	exitCode := -1
	if command.ProcessState != nil {
		exitCode = command.ProcessState.ExitCode()
	}

	m.mu.Lock()
	current := m.processes[entry.config.ID]
	if current == entry && entry.generation == generation && entry.command == command {
		entry.command = nil
		entry.done = nil
		entry.grouped = false
		entry.pid = 0
		entry.exitCode = exitCode
		entry.exitedAt = time.Now()
		if !entry.desired {
			if cleanupErr != nil {
				entry.state = StateFailed
				entry.lastError = cleanupErr.Error()
			} else {
				entry.state = StateStopped
				entry.lastError = ""
			}
		} else if entry.config.AutoRestart {
			if err := errors.Join(waitErr, cleanupErr); err != nil {
				entry.lastError = err.Error()
			} else {
				entry.lastError = ""
			}
			m.scheduleRestartLocked(entry)
		} else {
			entry.desired = false
			if waitErr != nil || cleanupErr != nil || exitCode != 0 {
				entry.state = StateFailed
				if err := errors.Join(waitErr, cleanupErr); err != nil {
					entry.lastError = err.Error()
				}
			} else {
				entry.state = StateStopped
				entry.lastError = ""
			}
		}
	}
	m.mu.Unlock()
	close(done)
}

func (m *Manager) scheduleRestartLocked(entry *managedProcess) {
	if !entry.desired || !entry.config.AutoRestart {
		return
	}
	if entry.restartTimer != nil {
		entry.restartTimer.Stop()
	}
	entry.state = StateRestarting
	generation := entry.generation
	id := entry.config.ID
	entry.restartTimer = time.AfterFunc(entry.config.RestartDelay, func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		current := m.processes[id]
		if current != entry || entry.generation != generation || !entry.desired || entry.command != nil {
			return
		}
		entry.restartTimer = nil
		entry.restartCount++
		_ = m.startProcessLocked(entry)
	})
}

func (m *Manager) stopByID(id string) error {
	m.mu.RLock()
	entry := m.processes[id]
	m.mu.RUnlock()
	if entry == nil {
		return fmt.Errorf("进程不存在: %s", id)
	}
	return m.stopEntry(entry)
}

func (m *Manager) stopEntry(entry *managedProcess) error {
	m.mu.Lock()
	entry.desired = false
	if entry.restartTimer != nil {
		entry.restartTimer.Stop()
		entry.restartTimer = nil
	}
	command := entry.command
	done := entry.done
	grouped := entry.grouped
	timeout := entry.config.StopTimeout
	if command == nil || command.Process == nil || done == nil {
		entry.pid = 0
		entry.state = StateStopped
		entry.lastError = ""
		entry.exitedAt = time.Now()
		m.mu.Unlock()
		return nil
	}
	entry.state = StateStopping
	entry.stopAt = time.Now()
	m.mu.Unlock()

	_ = m.signalProcessTree(command.Process, grouped, false)
	wait := func(timeout time.Duration) bool {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		select {
		case <-done:
			return true
		case <-timer.C:
			return false
		}
	}
	result := func() error {
		m.mu.RLock()
		defer m.mu.RUnlock()
		if entry.command == nil && entry.state == StateFailed && entry.lastError != "" {
			return errors.New(entry.lastError)
		}
		return nil
	}
	if wait(timeout) {
		return result()
	}
	killErr := m.signalProcessTree(command.Process, grouped, true)
	if wait(forceWaitTimeout) {
		return result()
	}
	m.mu.Lock()
	if entry.command == command {
		entry.state = StateFailed
		entry.lastError = "强制停止进程超时"
	}
	m.mu.Unlock()
	return errors.Join(errors.New("强制停止进程超时"), killErr)
}

func (m *Manager) cleanupProcessGroup(process *os.Process, timeout time.Duration, stopAt time.Time) error {
	exited, err := m.waitProcessGroup(process.Pid, 0)
	if err != nil || exited {
		return err
	}
	stopping := !stopAt.IsZero()
	if !stopping {
		stopAt = time.Now()
		_ = m.signalProcessTree(process, true, false)
	}
	remaining := timeout - time.Since(stopAt)
	if remaining > 0 {
		exited, err = m.waitProcessGroup(process.Pid, remaining)
		if err != nil || exited {
			return err
		}
	}
	if stopping {
		exited, err = m.waitProcessGroup(process.Pid, forceWaitTimeout)
		if exited {
			return nil
		}
		return errors.Join(errors.New("清理进程组超时"), err)
	}
	killErr := m.signalProcessTree(process, true, true)
	exited, err = m.waitProcessGroup(process.Pid, forceWaitTimeout)
	if exited {
		return nil
	}
	return errors.Join(errors.New("清理进程组超时"), killErr, err)
}

func (m *Manager) waitProcessGroup(groupID int, timeout time.Duration) (bool, error) {
	deadline := time.Now().Add(timeout)
	for {
		members, err := m.processGroupMembers(groupID)
		if err != nil || len(members) == 0 {
			return len(members) == 0, err
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return false, nil
		}
		if remaining > 50*time.Millisecond {
			remaining = 50 * time.Millisecond
		}
		time.Sleep(remaining)
	}
}

func (m *Manager) systemShellCommand(line string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd.exe", "/D", "/S", "/C", line)
	}
	return exec.Command("/bin/sh", "-c", line)
}

func (m *Manager) openProcessLog(path string) (io.WriteCloser, error) {
	if path == "" {
		return nil, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	logger := &timberjack.Logger{
		Filename:    path,
		MaxSize:     50,
		MaxBackups:  10,
		MaxAge:      30,
		Compression: "gzip",
		FileMode:    0600,
	}
	// 启动前打开日志，避免路径或权限错误延迟到子进程写入时才暴露。
	if _, err := logger.Write(nil); err != nil {
		_ = logger.Close()
		return nil, err
	}
	if err := os.Chmod(path, 0600); err != nil {
		_ = logger.Close()
		return nil, err
	}
	return logger, nil
}

func (m *Manager) mergeEnvironment(base, overrides []string) []string {
	result := append([]string(nil), base...)
	positions := make(map[string]int, len(result))
	normalizeKey := func(key string) string {
		if runtime.GOOS == "windows" {
			return strings.ToUpper(key)
		}
		return key
	}
	for index, value := range result {
		key, _, ok := strings.Cut(value, "=")
		if ok {
			positions[normalizeKey(key)] = index
		}
	}
	for _, value := range overrides {
		key, _, _ := strings.Cut(value, "=")
		key = normalizeKey(key)
		if index, exists := positions[key]; exists {
			result[index] = value
		} else {
			positions[key] = len(result)
			result = append(result, value)
		}
	}
	return result
}

// configureProcessGroup 通过反射配置平台进程属性，避免字段差异影响交叉编译。
func (m *Manager) configureProcessGroup(command *exec.Cmd) bool {
	if runtime.GOOS == "windows" {
		return false
	}
	commandValue := reflect.ValueOf(command).Elem()
	attributeField := commandValue.FieldByName("SysProcAttr")
	if !attributeField.IsValid() || attributeField.Type().Kind() != reflect.Pointer {
		return false
	}
	attribute := reflect.New(attributeField.Type().Elem())
	configured := false
	grouped := false
	setProcessGroup := attribute.Elem().FieldByName("Setpgid")
	if setProcessGroup.IsValid() && setProcessGroup.CanSet() && setProcessGroup.Kind() == reflect.Bool {
		setProcessGroup.SetBool(true)
		configured = true
		grouped = true
	}
	parentDeathSignal := attribute.Elem().FieldByName("Pdeathsig")
	if parentDeathSignal.IsValid() && parentDeathSignal.CanSet() && parentDeathSignal.CanInt() {
		parentDeathSignal.SetInt(9) // Linux SIGKILL
		configured = true
	}
	if configured {
		attributeField.Set(attribute)
	}
	return grouped
}

func (m *Manager) signalProcessTree(process *os.Process, grouped, force bool) error {
	if process == nil {
		return nil
	}
	if runtime.GOOS == "windows" {
		arguments := []string{"/PID", strconv.Itoa(process.Pid), "/T"}
		if force {
			arguments = append(arguments, "/F")
		}
		if err := exec.Command("taskkill", arguments...).Run(); err == nil {
			return nil
		}
	}
	processes := []*os.Process{process}
	if grouped {
		if members, err := m.processGroupMembers(process.Pid); err == nil && len(members) > 0 {
			processes = members
		}
	}
	var result error
	for _, member := range processes {
		var err error
		if force {
			err = member.Kill()
		} else {
			err = member.Signal(syscall.SIGTERM)
		}
		if err != nil && !errors.Is(err, os.ErrProcessDone) {
			result = errors.Join(result, err)
		}
	}
	return result
}

func (m *Manager) processGroupMembers(groupID int) ([]*os.Process, error) {
	if runtime.GOOS == "linux" {
		entries, err := os.ReadDir("/proc")
		if err != nil {
			return nil, err
		}
		result := make([]*os.Process, 0, 4)
		for _, entry := range entries {
			pid, parseErr := strconv.Atoi(entry.Name())
			if parseErr != nil || pid == os.Getpid() {
				continue
			}
			stat, readErr := os.ReadFile(filepath.Join("/proc", entry.Name(), "stat"))
			if readErr != nil {
				continue
			}
			content := string(stat)
			closing := strings.LastIndexByte(content, ')')
			if closing < 0 {
				continue
			}
			fields := strings.Fields(content[closing+1:])
			if len(fields) < 3 || fields[0] == "Z" {
				continue
			}
			processGroup, groupErr := strconv.Atoi(fields[2])
			if groupErr != nil || processGroup != groupID {
				continue
			}
			member, findErr := os.FindProcess(pid)
			if findErr == nil {
				result = append(result, member)
			}
		}
		return result, nil
	}
	output, err := exec.Command("ps", "-e", "-o", "pid=", "-o", "pgid=", "-o", "stat=").Output()
	if err != nil {
		return nil, err
	}
	result := make([]*os.Process, 0, 4)
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 3 || strings.HasPrefix(fields[2], "Z") {
			continue
		}
		pid, pidErr := strconv.Atoi(fields[0])
		pgid, pgidErr := strconv.Atoi(fields[1])
		if pidErr != nil || pgidErr != nil || pgid != groupID || pid == os.Getpid() {
			continue
		}
		member, findErr := os.FindProcess(pid)
		if findErr == nil {
			result = append(result, member)
		}
	}
	return result, nil
}
