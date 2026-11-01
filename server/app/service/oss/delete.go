package oss

import "errors"

// DeleteByIds 删除面板中的连接。
func (s *service) DeleteByIds(ids []int64) error {
	if len(ids) == 0 {
		return errors.New("请选择要删除的对象存储")
	}
	for _, id := range ids {
		if id <= 0 {
			return errors.New("存储 ID 不合法")
		}
	}
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	return s.repository.DeleteByIds(ids)
}
