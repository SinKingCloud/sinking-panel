package site

import (
	"errors"
	"fmt"
	"strconv"

	"server/app/enum/site_status"
	"server/app/enum/site_type"
	processManager "server/app/util/process"

	"gorm.io/gorm"
)

// Process 查询或控制项目进程，不改变网站状态和反向代理配置。
func (s *service) Process(id int64, action string) (*processManager.Status, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if id <= 0 {
		return nil, errors.New("网站 ID 不合法")
	}
	switch action {
	case "status", "start", "stop", "restart":
	default:
		return nil, errors.New("进程操作不合法")
	}
	record, err := s.repositorySite.FindById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("网站不存在")
		}
		return nil, fmt.Errorf("查询网站失败: %w", err)
	}
	if record.Type != site_type.General {
		return nil, errors.New("当前网站不是通用网站")
	}
	processID := strconv.FormatInt(id, 10)
	current, _ := s.process.Status(processID)
	if current != nil {
		if action == "status" && current.Command != "" {
			return current, nil
		}
		if action == "stop" && current.Command != "" {
			err = s.process.Stop(processID)
			current, _ = s.process.Status(processID)
			return current, err
		}
	}
	if action == "start" || action == "restart" {
		if !s.active {
			return nil, errors.New("请先启用网站 HTTP 服务")
		}
		if record.Status != site_status.Enabled {
			return nil, errors.New("请先启用网站")
		}
		if action == "start" && current != nil &&
			(current.State == processManager.StateRunning || current.State == processManager.StateStarting || current.State == processManager.StateRestarting) {
			return current, nil
		}
	}
	value, err := s.loadConfigLocked(id)
	if err != nil {
		return nil, err
	}
	general := value.(GeneralConfig)
	config, err := s.runtimeProcess(record, &general.Process)
	if err != nil {
		return nil, err
	}
	if action == "status" || action == "stop" {
		if action == "stop" {
			if err = s.process.Stop(processID); err != nil {
				return nil, err
			}
		}
		return &processManager.Status{
			ID:          processID,
			State:       processManager.StateStopped,
			Command:     config.Command,
			WorkingDir:  config.WorkingDir,
			AutoRestart: config.AutoRestart,
			MaxRetries:  config.MaxRetries,
			LogPath:     config.LogPath,
			ExitCode:    -1,
		}, nil
	}
	if action == "restart" && current != nil {
		if err = s.process.Stop(processID); err != nil {
			return nil, err
		}
	}
	status, err := s.process.Start(*config)
	if status != nil {
		for index := range s.processes {
			if s.processes[index].ID == processID {
				s.processes[index] = *config
				return status, err
			}
		}
		s.processes = append(s.processes, *config)
	}
	return status, err
}
