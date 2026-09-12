package secret

import (
	"encoding/json"
	"errors"
	"server/app/constant"
	"server/app/enum/secret_provider"
	repositorySecret "server/app/repository/secret"
	"strings"

	"gorm.io/gorm"
)

// Update 更新密钥，服务商变化时解除证书关联。
func (s *service) Update(id int64, data *repositorySecret.UpdateSecret) error {
	if id <= 0 {
		return errors.New("密钥 ID 不合法")
	}
	if data == nil {
		return errors.New("密钥更新数据不能为空")
	}
	if data.Name != nil {
		*data.Name = strings.TrimSpace(*data.Name)
		if *data.Name == "" {
			return errors.New("密钥名称不能为空")
		}
	}
	if data.Provider != nil {
		if _, ok := secret_provider.Map()[*data.Provider]; !ok {
			return errors.New("服务商不合法")
		}
	}
	if data.Data != nil {
		if len(*data.Data) > 1<<20 {
			return errors.New("密钥内容不能超过 1 MB")
		}
		var fields map[string]json.RawMessage
		if json.Unmarshal([]byte(*data.Data), &fields) != nil || fields == nil {
			return errors.New("密钥内容必须是 JSON 对象")
		}
	}
	s.enumMu.Lock()
	defer s.enumMu.Unlock()
	err := s.database.Transaction(func(tx *gorm.DB) error {
		previous, err := s.repositorySecret.FindById(id, tx)
		if err != nil {
			return err
		}
		if data.Provider != nil && *data.Provider != previous.Provider {
			if err = s.repositoryCert.ClearSecretId(id, tx); err != nil {
				return err
			}
		}
		return s.repositorySecret.UpdateById(id, data, tx)
	})
	if err != nil {
		return err
	}
	s.cache.Delete(constant.CacheNameWithSecretNameEnum)
	return nil
}
