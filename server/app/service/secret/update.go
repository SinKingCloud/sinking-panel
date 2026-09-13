package secret

import (
	"encoding/json"
	"errors"
	"server/app/constant"
	repositorySecret "server/app/repository/secret"
	"strings"

	"gorm.io/gorm"
)

// Update 更新密钥名称及凭据。
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
		if data.Data != nil {
			credentials, err := s.formatData(previous.Data, previous.Provider)
			if err != nil {
				return err
			}
			for key, previousValue := range credentials {
				if value, exists := fields[key]; exists {
					value = strings.TrimSpace(value)
					// 原样回传的当前掩码与省略字段一样，保留数据库中的原值。
					if previousValue == "" || value != s.maskValue(previousValue) {
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
			value := string(content)
			data.Data = &value
		}
		return s.repositorySecret.UpdateById(id, data, tx)
	})
	if err != nil {
		return err
	}
	s.cache.Delete(constant.CacheNameWithSecretNameEnum)
	return nil
}
