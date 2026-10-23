package database

import (
	"errors"
)

// DeleteByIds 只删除面板保存的实例连接，不操作远端数据库。
func (s *service) DeleteByIds(ids []int64) error {
	if len(ids) == 0 {
		return errors.New("请选择实例")
	}
	for _, id := range ids {
		if id <= 0 {
			return errors.New("实例 ID 不合法")
		}
	}
	return s.repository.DeleteByIds(ids)
}
