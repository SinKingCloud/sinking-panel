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
		name := strings.TrimSpace(*data.Name)
		if name == "" {
			return errors.New("密钥名称不能为空")
		}
		data.Name = &name
	}
	if data.Provider != nil {
		if _, ok := secret_provider.Map()[*data.Provider]; !ok {
			return errors.New("服务商不合法")
		}
	}
	var fields map[string]string
	if data.Data != nil {
		content := *data.Data
		if len(content) > 1<<20 {
			return errors.New("密钥内容不能超过 1 MB")
		}
		if json.Unmarshal([]byte(content), &fields) != nil || fields == nil {
			return errors.New("密钥内容必须是字符串字段组成的 JSON 对象")
		}
	}
	s.enumMu.Lock()
	defer s.enumMu.Unlock()
	err := s.database.Transaction(func(tx *gorm.DB) error {
		previous, err := s.repositorySecret.FindById(id, tx)
		if err != nil {
			return err
		}
		providerChanged := data.Provider != nil && *data.Provider != previous.Provider
		if data.Data != nil || providerChanged {
			provider := previous.Provider
			value := previous.Data
			if providerChanged {
				provider = *data.Provider
				value = "{}"
			}
			credentials, err := s.formatData(value, provider)
			if err != nil {
				return err
			}
			for key, previousValue := range credentials {
				if value, exists := fields[key]; exists {
					value = strings.TrimSpace(value)
					// 原样回传的当前掩码与省略字段一样，保留数据库中的原值。
					if providerChanged || previousValue == "" || value != s.maskValue(previousValue) {
						credentials[key] = value
					}
				}
				if credentials[key] == "" {
					return errors.New(key + "不能为空")
				}
				if strings.Contains(credentials[key], "****") {
					return errors.New(key + "请填写完整密钥，未修改时无需提交")
				}
			}
			content, _ := json.Marshal(credentials)
			value = string(content)
			data.Data = &value
		}
		if providerChanged {
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
