package site

import (
	"errors"
	"fmt"
	"strings"

	"server/app/model"
	webServer "server/app/util/server"
)

// GetDomains 读取网站当前域名和证书绑定。
func (s *service) GetDomains(id int64) ([]*model.SiteDomain, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if id <= 0 {
		return nil, errors.New("网站 ID 不合法")
	}
	if _, err := s.repositorySite.FindById(id); err != nil {
		return nil, s.nilIfNotFound("查询网站失败", err)
	}
	domains, err := s.repositorySiteDomain.SelectBySiteId(id)
	if err != nil {
		return nil, fmt.Errorf("查询网站域名失败: %w", err)
	}
	return s.cloneSiteDomainRecords(domains), nil
}

// GetSSL 读取网站 TLS 策略和域名证书绑定。
func (s *service) GetSSL(id int64) (*SSLSettings, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	config, err := s.loadHTTPConfigLocked(id)
	if err != nil {
		return nil, err
	}
	domains, err := s.repositorySiteDomain.SelectBySiteId(id)
	if err != nil {
		return nil, fmt.Errorf("查询网站域名失败: %w", err)
	}
	return &SSLSettings{Config: config.TLS, Domains: s.cloneSiteDomainRecords(domains)}, nil
}

// GetWAF 读取网站 WAF 设置。
func (s *service) GetWAF(id int64) (*webServer.WAFOptions, error) {
	config, err := s.getHTTPConfig(id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(config.WAF.BlockPage) == "" {
		config.WAF.BlockPage = webServer.DefaultWAFBlockPage
	}
	return &config.WAF, nil
}

// GetCache 读取不包含内部实例名称的网站缓存设置。
func (s *service) GetCache(id int64) (*CacheConfig, error) {
	config, err := s.getHTTPConfig(id)
	if err != nil {
		return nil, err
	}
	return &CacheConfig{
		Enabled:             config.Cache.Enabled,
		TTL:                 config.Cache.TTL,
		Stale:               config.Cache.Stale,
		Methods:             config.Cache.Methods,
		KeyHeaders:          config.Cache.KeyHeaders,
		DefaultCacheControl: config.Cache.DefaultCacheControl,
		MaxSize:             config.Cache.MaxSize,
	}, nil
}

// GetRateLimit 读取网站访问频率限制。
func (s *service) GetRateLimit(id int64) (*webServer.RateLimitOptions, error) {
	config, err := s.getHTTPConfig(id)
	if err != nil {
		return nil, err
	}
	return &config.RateLimit, nil
}

// GetTrafficLimit 读取网站并发和响应速度限制。
func (s *service) GetTrafficLimit(id int64) (*webServer.TrafficLimitOptions, error) {
	config, err := s.getHTTPConfig(id)
	if err != nil {
		return nil, err
	}
	return &config.TrafficLimit, nil
}

// GetHeaders 读取网站请求和响应头设置。
func (s *service) GetHeaders(id int64) (*webServer.HeaderOptions, error) {
	config, err := s.getHTTPConfig(id)
	if err != nil {
		return nil, err
	}
	return &config.Headers, nil
}

// GetCompression 读取网站响应压缩设置。
func (s *service) GetCompression(id int64) (*webServer.CompressionOptions, error) {
	config, err := s.getHTTPConfig(id)
	if err != nil {
		return nil, err
	}
	return &config.Compression, nil
}

// GetRedirects 读取网站重定向规则。
func (s *service) GetRedirects(id int64) ([]RedirectConfig, error) {
	config, err := s.getHTTPConfig(id)
	if err != nil {
		return nil, err
	}
	return config.Redirects, nil
}

// GetRoutes 读取网站自定义路由。
func (s *service) GetRoutes(id int64) ([]RouteConfig, error) {
	config, err := s.getHTTPConfig(id)
	if err != nil {
		return nil, err
	}
	return config.Routes, nil
}

// GetStatic 读取静态网站专属设置。
func (s *service) GetStatic(id int64) (*StaticOptions, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	value, err := s.loadConfigLocked(id)
	if err != nil {
		return nil, err
	}
	config, ok := value.(StaticConfig)
	if !ok {
		return nil, errors.New("当前网站不是静态网站")
	}
	return &StaticOptions{
		Index:         config.Index,
		Browse:        config.Browse,
		TryFiles:      config.TryFiles,
		Hide:          config.Hide,
		Precompressed: config.Precompressed,
	}, nil
}

// GetProxy 读取反向代理或通用网站的上游设置。
func (s *service) GetProxy(id int64) (*webServer.ProxyOptions, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	value, err := s.loadConfigLocked(id)
	if err != nil {
		return nil, err
	}
	switch config := value.(type) {
	case ProxyConfig:
		return &config.Proxy, nil
	case GeneralConfig:
		return &config.Proxy, nil
	default:
		return nil, errors.New("当前网站不支持反向代理设置")
	}
}

// GetFastCGI 读取 PHP 网站 FastCGI 设置。
func (s *service) GetFastCGI(id int64) (*webServer.ProxyOptions, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	value, err := s.loadConfigLocked(id)
	if err != nil {
		return nil, err
	}
	config, ok := value.(PHPConfig)
	if !ok {
		return nil, errors.New("当前网站不是 PHP 网站")
	}
	return &config.FastCGI, nil
}

// GetProcess 读取通用网站进程保活设置。
func (s *service) GetProcess(id int64) (*ProcessConfig, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	value, err := s.loadConfigLocked(id)
	if err != nil {
		return nil, err
	}
	config, ok := value.(GeneralConfig)
	if !ok {
		return nil, errors.New("当前网站不是通用网站")
	}
	return &config.Process, nil
}

func (s *service) getHTTPConfig(id int64) (*HTTPConfig, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	return s.loadHTTPConfigLocked(id)
}

func (s *service) loadHTTPConfigLocked(id int64) (*HTTPConfig, error) {
	value, err := s.loadConfigLocked(id)
	if err != nil {
		return nil, err
	}
	switch config := value.(type) {
	case StaticConfig:
		return &config.HTTPConfig, nil
	case ProxyConfig:
		return &config.HTTPConfig, nil
	case PHPConfig:
		return &config.HTTPConfig, nil
	case GeneralConfig:
		return &config.HTTPConfig, nil
	default:
		return nil, errors.New("网站配置类型错误")
	}
}
