package script

import repositoryScript "server/app/repository/script"

// UpdateByIds 通过ID列表更新常用脚本
func (s *service) UpdateByIds(ids []int64, data *repositoryScript.UpdateScript) error {
	return s.repositoryScript.UpdateByIds(ids, data)
}
