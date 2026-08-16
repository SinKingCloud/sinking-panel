package log

import (
	"time"
)

// Clear 清理操作日志，至少保留最近一天的日志。
func (s *service) Clear(day int) (int64, error) {
	if day < 1 {
		day = 1
	}
	ids, err := s.repositoryLog.SelectIdByCreateTime(time.Now().AddDate(0, 0, -day), 10000)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	if err = s.repositoryLog.Delete(ids); err != nil {
		return 0, err
	}
	return int64(len(ids)), nil
}
