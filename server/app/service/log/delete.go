package log

import (
	"time"
)

// Clear 每批清理一万条操作日志，直到清理完成，至少保留最近一天的日志。
func (s *service) Clear(day int) (int64, error) {
	if day < 1 {
		day = 1
	}
	before := time.Now().AddDate(0, 0, -day)
	var total int64
	for {
		ids, err := s.repositoryLog.SelectIdByCreateTime(before, 10000)
		if err != nil {
			return total, err
		}
		if len(ids) == 0 {
			return total, nil
		}
		if err = s.repositoryLog.Delete(ids); err != nil {
			return total, err
		}
		total += int64(len(ids))
	}
}
