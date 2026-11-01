package tencent

import (
	"net/http"
	"server/app/enum/secret_provider"
	"server/app/util/oss"
)

// Secret 配置腾讯云 COS凭据，支持 JSON 编解码，不能传入脱敏值。
type Secret struct {
	SecretID     string `json:"secret_id"`     // 必填，腾讯云 SecretId。
	SecretKey    string `json:"secret_key"`    // 必填，腾讯云 SecretKey，不能使用脱敏值。
	SessionToken string `json:"session_token"` // 使用 STS 临时凭据时必填；长期密钥留空。
}

// Secret 将厂商凭据转换为统一密钥，保留临时令牌。
func (s Secret) Secret() oss.Secret {
	return oss.Secret{AccessKeyID: s.SecretID, AccessKeySecret: s.SecretKey, SecurityToken: s.SessionToken}
}

// Config 配置腾讯云 COS连接，不包含凭据，无需填写 Provider。
type Config struct {
	Region     string       `json:"region"`   // 区域，例如 ap-guangzhou；与 Endpoint 至少填写一项。
	Endpoint   string       `json:"endpoint"` // 区域服务地址，例如 https://cos.ap-guangzhou.myqcloud.com；留空按 Region 生成。
	Bucket     string       `json:"bucket"`   // 对象操作必填，需包含 APPID，例如 backup-1250000000；仅列桶时可留空。
	HTTPClient *http.Client `json:"-"`        // 选填，自定义超时和传输；构造时复制，不修改原客户端。
}

// Config 将连接参数转换为统一配置，保留自定义 HTTP 客户端。
func (c Config) Config() oss.Config {
	return oss.Config{
		Provider: secret_provider.TencentCloud, Region: c.Region, Endpoint: c.Endpoint, Bucket: c.Bucket,
		HTTPClient: c.HTTPClient,
	}
}

// WithHTTPClient 返回使用指定 HTTP 客户端的配置副本，保留其余厂商参数。
func (c Config) WithHTTPClient(client *http.Client) oss.ConfigSource {
	c.HTTPClient = client
	return c
}
