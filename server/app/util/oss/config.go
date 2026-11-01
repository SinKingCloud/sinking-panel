package oss

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

// ConfigSource 让统一入口和厂商入口接收统一配置或厂商专用配置。
type ConfigSource interface {
	Config() Config
	WithHTTPClient(*http.Client) ConfigSource
}

// Config 用于动态选择厂商；已知厂商时优先使用对应子包的 Config。
type Config struct {
	Provider   int          `json:"provider"` // 密钥厂商，使用 secret_provider 枚举。
	Region     string       `json:"region"`   // 地域；阿里云和火山引擎签名必须提供。
	Endpoint   string       `json:"endpoint"` // 区域服务地址，不含桶名；为空时根据 Region 生成。
	Bucket     string       `json:"bucket"`   // 存储桶名称；仅列出账户下的桶时可留空。
	HTTPClient *http.Client `json:"-"`        // 可选 HTTP 客户端，保留超时和自定义传输配置。
}

// Config 返回统一配置副本，不修改调用方的数据。
func (c Config) Config() Config {
	return c
}

// WithHTTPClient 返回使用指定 HTTP 客户端的配置副本。
func (c Config) WithHTTPClient(client *http.Client) ConfigSource {
	c.HTTPClient = client
	return c
}

// ResolveConfig 读取公共连接参数并拒绝空接口和空指针，不处理访问凭据。
func ResolveConfig(source ConfigSource) (Config, error) {
	if source == nil {
		return Config{}, fmt.Errorf("%w: 配置不能为空", ErrInvalidArgument)
	}
	value := reflect.ValueOf(source)
	if value.Kind() == reflect.Pointer && value.IsNil() {
		return Config{}, fmt.Errorf("%w: 配置不能为空", ErrInvalidArgument)
	}
	return source.Config(), nil
}

// NormalizeConfig 校验公共配置并准备 HTTP 客户端，不发网络请求。
// 默认服务地址和签名所需的地域约束由对应厂商实现处理。
// 已关闭透明解压的标准 Transport 直接复用，避免重复构造连接池。
func NormalizeConfig(credentials SecretSource, source ConfigSource, provider int) (Secret, Config, error) {
	secret, err := ResolveSecret(credentials)
	if err != nil {
		return Secret{}, Config{}, err
	}
	c, err := ResolveConfig(source)
	if err != nil {
		return Secret{}, Config{}, err
	}
	if c.Provider != provider {
		return Secret{}, Config{}, fmt.Errorf("%w: 配置与客户端的厂商不一致", ErrInvalidArgument)
	}
	if len(c.Region) > 100 || len(c.Endpoint) > 500 || len(c.Bucket) > 255 {
		return Secret{}, Config{}, fmt.Errorf("%w: 存储连接参数过长", ErrInvalidArgument)
	}
	if strings.TrimSpace(secret.AccessKeyID) == "" || strings.TrimSpace(secret.AccessKeySecret) == "" || strings.Contains(secret.AccessKeyID, "*") || strings.Contains(secret.AccessKeySecret, "*") {
		return Secret{}, Config{}, fmt.Errorf("%w: 需要完整的原始访问密钥，不能使用脱敏值", ErrInvalidArgument)
	}
	if c.Bucket != "" && !regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`).MatchString(c.Bucket) {
		return Secret{}, Config{}, fmt.Errorf("%w: 桶名称不合法", ErrInvalidArgument)
	}
	if c.Region != "" && !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(c.Region) {
		return Secret{}, Config{}, fmt.Errorf("%w: 区域名称不合法", ErrInvalidArgument)
	}
	if c.Endpoint != "" {
		if !strings.Contains(c.Endpoint, "://") {
			c.Endpoint = "https://" + c.Endpoint
		}
		u, err := url.Parse(c.Endpoint)
		if err != nil || u.Hostname() == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
			return Secret{}, Config{}, fmt.Errorf("%w: Endpoint 必须是无桶名、路径、凭据和查询参数的 HTTP(S) 服务地址", ErrInvalidArgument)
		}
		c.Endpoint = strings.TrimSuffix(c.Endpoint, "/")
	}
	client := &http.Client{}
	if c.HTTPClient != nil {
		*client = *c.HTTPClient
	}
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	if standard, ok := transport.(*http.Transport); ok && !standard.DisableCompression {
		cloned := standard.Clone()
		cloned.DisableCompression = true
		transport = cloned
	}
	client.Transport = transport
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	c.HTTPClient = client
	return secret, c, nil
}
