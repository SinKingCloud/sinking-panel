package volcengine

import (
	"net/http"
	"server/app/enum/secret_provider"
	"server/app/util/oss"
)

// Secret 配置火山引擎 TOS凭据，支持 JSON 编解码，不能传入脱敏值。
type Secret struct {
	AccessKeyID     string `json:"access_key_id"`     // 必填，火山引擎 Access Key ID。
	AccessKeySecret string `json:"access_key_secret"` // 必填，火山引擎 Secret Access Key，不能使用脱敏值。
	SecurityToken   string `json:"security_token"`    // 使用 STS 临时凭据时必填；长期密钥留空。
}

// Secret 将厂商凭据转换为统一密钥，保留临时令牌。
func (s Secret) Secret() oss.Secret {
	return oss.Secret{AccessKeyID: s.AccessKeyID, AccessKeySecret: s.AccessKeySecret, SecurityToken: s.SecurityToken}
}

// Config 配置火山引擎 TOS连接，不包含凭据，无需填写 Provider。
type Config struct {
	Region     string       `json:"region"`   // 必填，签名区域，例如 cn-beijing；指定 Endpoint 时仍需填写。
	Endpoint   string       `json:"endpoint"` // 选填，区域服务地址，例如 https://tos-cn-beijing.volces.com；留空按 Region 生成。
	Bucket     string       `json:"bucket"`   // 对象操作必填，例如 panel-backups；仅列举账户存储桶时可留空。
	HTTPClient *http.Client `json:"-"`        // 选填，自定义超时和传输；构造时复制，不修改原客户端。
}

// Config 将连接参数转换为统一配置，保留自定义 HTTP 客户端。
func (c Config) Config() oss.Config {
	return oss.Config{
		Provider: secret_provider.Volcengine, Region: c.Region, Endpoint: c.Endpoint, Bucket: c.Bucket,
		HTTPClient: c.HTTPClient,
	}
}

// WithHTTPClient 返回使用指定 HTTP 客户端的配置副本，保留其余厂商参数。
func (c Config) WithHTTPClient(client *http.Client) oss.ConfigSource {
	c.HTTPClient = client
	return c
}
