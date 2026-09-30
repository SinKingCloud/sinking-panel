package server

// DeleteByIds 通过ID列表删除
func (s *service) DeleteByIds(ids []int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repositoryServer.DeleteByIds(ids)
}
