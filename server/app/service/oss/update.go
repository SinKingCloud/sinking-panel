package oss

import (
	"encoding/json"
	"errors"
	repositoryOss "server/app/repository/oss"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
)

// Update 仅更新提交的字段，连接配置按字段合并后校验。
func (s *service) Update(id int64, data *repositoryOss.UpdateOss) error {
	if id <= 0 {
		return errors.New("存储 ID 不合法")
	}
	if data == nil {
		return errors.New("存储更新数据不能为空")
	}
	if data.Name != nil {
		name := strings.TrimSpace(*data.Name)
		if name == "" || utf8.RuneCountInString(name) > 100 {
			return errors.New("存储名称应为 1 至 100 个字符")
		}
		data.Name = &name
	}
	var fields map[string]json.RawMessage
	if data.Config != nil {
		content := *data.Config
		if len(content) > 1<<20 || json.Unmarshal([]byte(content), &fields) != nil || fields == nil {
			return errors.New("存储配置必须是 JSON 对象，且不超过 1 MB")
		}
	}
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	return s.database.Transaction(func(tx *gorm.DB) error {
		previous, err := s.repository.FindById(id, tx)
		if err != nil {
			return err
		}
		if previous == nil {
			return errors.New("对象存储不存在")
		}
		if data.Config != nil {
			var values map[string]json.RawMessage
			if json.Unmarshal([]byte(previous.Config), &values) != nil || values == nil {
				return errors.New("已保存的存储配置格式不正确")
			}
			for key, value := range fields {
				values[key] = value
			}
			merged, err := json.Marshal(values)
			if err != nil {
				return err
			}
			_, config, err := s.newClient(previous.SecretId, string(merged), tx)
			if err != nil {
				return err
			}
			content, err := json.Marshal(config)
			if err != nil {
				return err
			}
			value := string(content)
			data.Config = &value
		}
		return s.repository.UpdateById(id, data, tx)
	})
}
