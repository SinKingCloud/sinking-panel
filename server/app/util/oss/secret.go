package oss

import (
	"fmt"
	"reflect"
)

// SecretSource 接收统一密钥或厂商专用密钥，与存储连接配置分开传递。
type SecretSource interface {
	Secret() Secret
}

// Secret 保存原始访问凭据，不能用于存储配置表单或详情响应。
type Secret struct {
	AccessKeyID     string `json:"access_key_id"`     // 原始访问密钥 ID，腾讯云对应 SecretID。
	AccessKeySecret string `json:"access_key_secret"` // 原始访问密钥，腾讯云对应 SecretKey。
	SecurityToken   string `json:"security_token"`    // 临时凭据令牌，长期密钥留空。
}

// Secret 返回统一密钥副本，不修改调用方的数据。
func (s Secret) Secret() Secret {
	return s
}

// ResolveSecret 转换凭据并拒绝空接口和空指针。
func ResolveSecret(source SecretSource) (Secret, error) {
	if source == nil {
		return Secret{}, fmt.Errorf("%w: 密钥不能为空", ErrInvalidArgument)
	}
	value := reflect.ValueOf(source)
	if value.Kind() == reflect.Pointer && value.IsNil() {
		return Secret{}, fmt.Errorf("%w: 密钥不能为空", ErrInvalidArgument)
	}
	return source.Secret(), nil
}
