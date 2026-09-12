package secret

import (
	"errors"
	"server/app/constant"

	"gorm.io/gorm"
)

// Delete 解除证书关联后删除密钥。
func (s *service) Delete(id int64) error {
	if id <= 0 {
		return errors.New("密钥 ID 不合法")
	}
	s.enumMu.Lock()
	defer s.enumMu.Unlock()
	err := s.database.Transaction(func(tx *gorm.DB) error {
		if err := s.repositoryCert.ClearSecretId(id, tx); err != nil {
			return err
		}
		return s.repositorySecret.DeleteById(id, tx)
	})
	if err != nil {
		return err
	}
	s.cache.Delete(constant.CacheNameWithSecretNameEnum)
	return nil
}
