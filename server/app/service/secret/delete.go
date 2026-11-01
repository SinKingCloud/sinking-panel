package secret

import (
	"errors"

	"gorm.io/gorm"
)

// Delete 在事务内检查对象存储和证书关联，仅删除未被使用的密钥。
func (s *service) Delete(id int64) error {
	if id <= 0 {
		return errors.New("密钥 ID 不合法")
	}
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	return s.database.Transaction(func(tx *gorm.DB) error {
		count, err := s.repositoryOss.CountBySecretId(id, tx)
		if err != nil {
			return err
		}
		if count > 0 {
			return errors.New("密钥被对象存储使用，请先修改或删除对应存储配置")
		}
		count, err = s.repositoryCert.CountBySecretId(id, tx)
		if err != nil {
			return err
		}
		if count > 0 {
			return errors.New("密钥被证书使用，请先修改或解除对应证书的密钥关联")
		}
		return s.repositorySecret.DeleteById(id, tx)
	})
}
