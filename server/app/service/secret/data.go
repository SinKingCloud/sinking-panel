package secret

import (
	"encoding/json"
	"errors"
	"server/app/enum/secret_provider"
	"strings"
)

// formatData 按厂商结构体读取配置，不暴露未声明的字段。
func (s *service) formatData(value string, provider int) (map[string]string, error) {
	var data interface{}
	switch provider {
	case secret_provider.TencentCloud:
		data = &TencentCloud{}
	case secret_provider.Aliyun:
		data = &Aliyun{}
	case secret_provider.HuaweiCloud:
		data = &HuaweiCloud{}
	case secret_provider.Volcengine:
		data = &Volcengine{}
	case secret_provider.BaiduCloud:
		data = &BaiduCloud{}
	default:
		return nil, errors.New("服务商不合法")
	}
	if err := json.Unmarshal([]byte(value), data); err != nil {
		return nil, errors.New("密钥内容格式不正确")
	}
	content, _ := json.Marshal(data)
	var fields map[string]string
	_ = json.Unmarshal(content, &fields)
	return fields, nil
}

// maskValue 保留首尾字符，中间使用星号，显示长度至少为六位。
func (s *service) maskValue(value string) string {
	characters := []rune(value)
	if len(characters) == 0 {
		return ""
	}
	return string(characters[0]) + strings.Repeat("*", max(4, len(characters)-2)) + string(characters[len(characters)-1])
}
