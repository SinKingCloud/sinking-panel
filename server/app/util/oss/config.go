package oss

import (
	"net/http"
	"net/url"
	"server/app/enum/secret_provider"
	"strings"
)

// normalizeConfig 校验配置并准备区域地址和独立 HTTP 客户端，不发起网络请求。
func normalizeConfig(c Config) (Config, error) {
	switch c.Provider {
	case secret_provider.TencentCloud, secret_provider.Aliyun, secret_provider.HuaweiCloud, secret_provider.Volcengine, secret_provider.BaiduCloud:
	default:
		return Config{}, invalid("不支持的密钥厂商")
	}
	if strings.TrimSpace(c.AccessKeyID) == "" || strings.TrimSpace(c.AccessKeySecret) == "" || strings.Contains(c.AccessKeyID, "*") || strings.Contains(c.AccessKeySecret, "*") {
		return Config{}, invalid("需要完整的原始访问密钥，不能使用脱敏值")
	}
	if c.Bucket != "" && !bucketNamePattern.MatchString(c.Bucket) {
		return Config{}, invalid("桶名称不合法")
	}
	if c.Region != "" && !regionNamePattern.MatchString(c.Region) {
		return Config{}, invalid("区域名称不合法")
	}
	if c.Endpoint == "" {
		if c.Region == "" {
			return Config{}, invalid("必须提供 Region 或 Endpoint")
		}
		switch c.Provider {
		case secret_provider.TencentCloud:
			c.Endpoint = "https://cos." + c.Region + ".myqcloud.com"
		case secret_provider.Aliyun:
			c.Endpoint = "https://oss-" + c.Region + ".aliyuncs.com"
		case secret_provider.HuaweiCloud:
			c.Endpoint = "https://obs." + c.Region + ".myhuaweicloud.com"
		case secret_provider.Volcengine:
			c.Endpoint = "https://tos-" + c.Region + ".volces.com"
		case secret_provider.BaiduCloud:
			c.Endpoint = "https://" + c.Region + ".bcebos.com"
		}
	}
	if !strings.Contains(c.Endpoint, "://") {
		c.Endpoint = "https://" + c.Endpoint
	}
	u, err := url.Parse(c.Endpoint)
	if err != nil || u.Hostname() == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
		return Config{}, invalid("Endpoint 必须是无桶名、路径、凭据和查询参数的 HTTP(S) 服务地址")
	}
	c.Endpoint = strings.TrimSuffix(c.Endpoint, "/")
	if (c.Provider == secret_provider.Aliyun || c.Provider == secret_provider.Volcengine) && c.Region == "" {
		return Config{}, invalid("该厂商签名需要 Region")
	}
	client := &http.Client{}
	if c.HTTPClient != nil {
		*client = *c.HTTPClient
	}
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	if standardTransport, ok := transport.(*http.Transport); ok {
		// 备份下载必须保留对象原始字节，不能让 Go 自动解压 Content-Encoding: gzip。
		clonedTransport := standardTransport.Clone()
		clonedTransport.DisableCompression = true
		transport = clonedTransport
	}
	client.Transport = transport
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	c.HTTPClient = client
	return c, nil
}
