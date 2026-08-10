package script

// DeleteByIds 通过ID列表删除常用脚本
func (s *service) DeleteByIds(ids []int64) error {
	return s.repositoryScript.DeleteByIds(ids)
}
