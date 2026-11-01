package oss

import (
	"encoding/json"
	"errors"
	"server/app/enum/secret_provider"
	utilOss "server/app/util/oss"
	"server/app/util/oss/aliyun"
	"server/app/util/oss/baidu"
	"server/app/util/oss/huawei"
	"server/app/util/oss/tencent"
	"server/app/util/oss/volcengine"
)

// formatConfig 返回厂商配置及表单字段，空字段保留用于编辑，不包含密钥。
func (s *service) formatConfig(content string, provider int) (utilOss.ConfigSource, map[string]string, error) {
	var values map[string]json.RawMessage
	if len(content) > 1<<20 || json.Unmarshal([]byte(content), &values) != nil || values == nil {
		return nil, nil, errors.New("存储配置必须是 JSON 对象，且不超过 1 MB")
	}
	var config utilOss.ConfigSource
	switch provider {
	case secret_provider.Aliyun:
		config = &aliyun.Config{}
	case secret_provider.TencentCloud:
		config = &tencent.Config{}
	case secret_provider.HuaweiCloud:
		config = &huawei.Config{}
	case secret_provider.BaiduCloud:
		config = &baidu.Config{}
	case secret_provider.Volcengine:
		config = &volcengine.Config{}
	default:
		return nil, nil, errors.New("云密钥厂商不支持对象存储")
	}
	if err := json.Unmarshal([]byte(content), config); err != nil {
		return nil, nil, errors.New("存储配置字段格式不正确")
	}
	contentBytes, err := json.Marshal(config)
	if err != nil {
		return nil, nil, err
	}
	var fields map[string]string
	if err = json.Unmarshal(contentBytes, &fields); err != nil {
		return nil, nil, err
	}
	for name := range fields {
		if string(values[name]) == "null" {
			return nil, nil, errors.New("存储配置字段 " + name + " 格式不正确")
		}
	}
	return config, fields, nil
}
