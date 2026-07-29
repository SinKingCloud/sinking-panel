package server

// DeleteByIds 通过ID列表删除
func (s *service) DeleteByIds(ids []int64) error {
	return s.repositoryServer.DeleteByIds(ids)
}
