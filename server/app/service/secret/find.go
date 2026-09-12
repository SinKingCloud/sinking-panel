package secret

import (
	"errors"
)

// FindById 查询供密钥管理使用的脱敏凭据。
func (s *service) FindById(id int64) (*Secret, error) {
	if id <= 0 {
		return nil, errors.New("密钥 ID 不合法")
	}
	data, err := s.repositorySecret.FindById(id)
	if err != nil {
		return nil, err
	}
	fields, err := s.formatData(data.Data, data.Provider)
	if err != nil {
		return nil, err
	}
	for key, value := range fields {
		fields[key] = s.maskValue(value)
	}
	return &Secret{
		Id:         data.Id,
		Name:       data.Name,
		Provider:   data.Provider,
		Data:       fields,
		CreateTime: data.CreateTime,
		UpdateTime: data.UpdateTime,
	}, nil
}
