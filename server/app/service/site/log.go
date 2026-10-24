package site

import (
	"errors"
	"fmt"
	"strconv"

	"server/app/enum/site_kind"
	"server/app/util/caddy"
	fileLog "server/app/util/log"
)

// ReadServerLog 按字节游标读取 HTTP 服务运行日志。
func (s *service) ReadServerLog(after int64, before int64, pageSize int) (map[string]interface{}, error) {
	s.logMu.RLock()
	defer s.logMu.RUnlock()
	path, err := s.http.LogPath("", caddy.LogServer)
	if err != nil {
		return nil, err
	}
	return fileLog.Read(path, after, before, pageSize)
}

// ClearServerLog 清空 HTTP 服务运行日志和已滚动的历史文件。
func (s *service) ClearServerLog() error {
	s.logMu.Lock()
	defer s.logMu.Unlock()
	path, err := s.http.LogPath("", caddy.LogServer)
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
func (s *service) ReadLog(id int64, logType caddy.LogType, after, before int64, pageSize int) (map[string]interface{}, error) {
	s.logMu.RLock()
	defer s.logMu.RUnlock()
	path, err := s.logPath(id, logType)
	if err != nil {
		return nil, err
	}
	return fileLog.Read(path, after, before, pageSize)
}

// ClearLog 清空网站日志和已滚动的历史文件。
func (s *service) ClearLog(id int64, logType caddy.LogType) error {
	s.logMu.Lock()
	defer s.logMu.Unlock()
	path, err := s.logPath(id, logType)
	if err != nil {
		return err
	}
	if err = fileLog.Clear(path); err != nil {
		return fmt.Errorf("清理网站日志失败: %w", err)
	}
	if logType != caddy.LogProcess && s.http.Running() {
		if err = s.http.Reload(); err != nil {
			return fmt.Errorf("日志已清理，但重新加载 http 日志写入器失败: %w", err)
		}
	}
	return nil
}

func (s *service) logPath(id int64, logType caddy.LogType) (string, error) {
	if id <= 0 {
		return "", errors.New("网站 ID 不合法")
	}
	record, err := s.repositorySite.FindById(id)
	if err != nil {
		return "", fmt.Errorf("查询网站失败: %w", err)
	}
	if record == nil {
		return "", errors.New("网站不存在")
	}
	if logType == caddy.LogProcess && record.Type != site_kind.General {
		return "", errors.New("当前网站没有进程日志")
	}
	path, err := s.http.LogPath(strconv.FormatInt(id, 10), logType)
	if err != nil {
		return "", err
	}
	return path, nil
}
