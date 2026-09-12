package secret

import (
	"encoding/json"
	"errors"
	"server/app/constant"
	"server/app/enum/secret_provider"
	"server/app/model"
	"server/app/util/str"
	"strings"
)

// Create 创建密钥凭据。
func (s *service) Create(data *model.Secret) error {
	if data == nil {
		return errors.New("密钥数据不能为空")
	}
	data.Name = strings.TrimSpace(data.Name)
	if data.Name == "" {
		return errors.New("密钥名称不能为空")
	}
	if _, ok := secret_provider.Map()[data.Provider]; !ok {
		return errors.New("服务商不合法")
	}
	if len(data.Data) > 1<<20 {
		return errors.New("密钥内容不能超过 1 MB")
	}
	if strings.TrimSpace(data.Data) == "" {
		data.Data = "{}"
	} else {
		fields, err := s.formatData(data.Data, data.Provider)
		if err != nil {
			return err
		}
		for key, value := range fields {
			value = strings.TrimSpace(value)
			if value == "" {
				return errors.New(key + "不能为空")
			}
			if strings.Contains(value, "****") {
				return errors.New(key + "请填写完整密钥")
			}
			fields[key] = value
		}
		content, _ := json.Marshal(fields)
		data.Data = string(content)
	}
	data.Id = str.GetSnowWorkIns().GetId()
	s.enumMu.Lock()
	defer s.enumMu.Unlock()
	if err := s.repositorySecret.Create(data); err != nil {
		return err
	}
	s.cache.Delete(constant.CacheNameWithSecretNameEnum)
	return nil
}
