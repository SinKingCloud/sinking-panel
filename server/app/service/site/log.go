package site

import (
	"errors"
	"fmt"
	"strconv"

	"server/app/enum/site_type"
	fileLog "server/app/util/log"
	webServer "server/app/util/server"

	"gorm.io/gorm"
)

// ReadServerLog 按字节游标读取 HTTP 服务运行日志。
func (s *service) ReadServerLog(after int64, before int64, pageSize int) (map[string]interface{}, error) {
	s.logMu.RLock()
	defer s.logMu.RUnlock()
	path, err := s.http.LogPath("", webServer.LogServer)
	if err != nil {
		return nil, err
	}
	return fileLog.Read(path, after, before, pageSize)
}

// ClearServerLog 清空 HTTP 服务运行日志和已滚动的历史文件。
func (s *service) ClearServerLog() error {
	s.logMu.Lock()
	defer s.logMu.Unlock()
	path, err := s.http.LogPath("", webServer.LogServer)
	if err != nil {
		return err
	}
	if err = fileLog.Clear(path); err != nil {
		return fmt.Errorf("清理 HTTP 服务运行日志失败: %w", err)
	}
	if s.http.Running() {
		if err = s.http.Reload(); err != nil {
			return fmt.Errorf("日志已清理，但重新加载 HTTP 日志写入器失败: %w", err)
		}
	}
	return nil
}

// ReadLog 按字节游标读取网站日志。
func (s *service) ReadLog(id int64, logType webServer.LogType, after int64, before int64, pageSize int) (map[string]interface{}, error) {
	s.logMu.RLock()
	defer s.logMu.RUnlock()
	path, err := s.logPath(id, logType)
	if err != nil {
		return nil, err
	}
	return fileLog.Read(path, after, before, pageSize)
}

// ClearLog 清空网站日志和已滚动的历史文件。
func (s *service) ClearLog(id int64, logType webServer.LogType) error {
	s.logMu.Lock()
	defer s.logMu.Unlock()
	path, err := s.logPath(id, logType)
	if err != nil {
		return err
	}
	if err = fileLog.Clear(path); err != nil {
		return fmt.Errorf("清理网站日志失败: %w", err)
	}
	if logType != webServer.LogProcess && s.http.Running() {
		if err = s.http.Reload(); err != nil {
			return fmt.Errorf("日志已清理，但重新加载 http 日志写入器失败: %w", err)
		}
	}
	return nil
}

func (s *service) logPath(id int64, logType webServer.LogType) (string, error) {
	if id <= 0 {
		return "", errors.New("网站 ID 不合法")
	}
	record, err := s.repositorySite.FindById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("网站不存在")
		}
		return "", fmt.Errorf("查询网站失败: %w", err)
	}
	if logType == webServer.LogProcess && record.Type != site_type.General {
		return "", errors.New("当前网站没有进程日志")
	}
	path, err := s.http.LogPath(strconv.FormatInt(id, 10), logType)
	if err != nil {
		return "", err
	}
	return path, nil
}
