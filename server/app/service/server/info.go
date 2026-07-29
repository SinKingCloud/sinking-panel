package server

import "server/app/model"

// FindById 查询信息
func (s *service) FindById(id int64) (*model.Server, error) {
	return s.repositoryServer.FindById(id)
}
