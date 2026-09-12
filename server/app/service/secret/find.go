package secret

import (
	"errors"
	"server/app/model"
)

// FindById 查询供密钥管理使用的完整凭据。
func (s *service) FindById(id int64) (*model.Secret, error) {
	if id <= 0 {
		return nil, errors.New("密钥 ID 不合法")
	}
	return s.repositorySecret.FindById(id)
}
