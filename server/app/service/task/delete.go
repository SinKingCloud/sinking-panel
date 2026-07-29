package task

// deleteById 通过ID删除
func (s *service) deleteById(id int64) error {
	return s.repositoryTask.DeleteById(id)
}
