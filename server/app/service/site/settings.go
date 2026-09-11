package site

import (
	"encoding/json"
	"errors"
	"fmt"

	webServer "server/app/util/server"
)

// UpdateDomains 更新网站域名，同名域名保留原证书绑定。
func (s *service) UpdateDomains(id int64, domains []string) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if id <= 0 {
		return errors.New("网站 ID 不合法")
	}
	previous, err := s.repositorySiteDomain.SelectBySiteId(id)
	if err != nil {
		return fmt.Errorf("查询网站域名失败: %w", err)
	}
	certificateByDomain := make(map[string]int64, len(previous))
	for _, domain := range previous {
		if domain == nil {
			continue
		}
		normalized, normalizeErr := s.normalizeDomain(domain.Domain)
		if normalizeErr != nil {
			return fmt.Errorf("现有网站域名无效: %w", normalizeErr)
		}
		certificateByDomain[normalized] = domain.CertId
	}
	input := make([]SiteDomain, 0, len(domains))
	for _, value := range domains {
		domain, normalizeErr := s.normalizeDomain(value)
		if normalizeErr != nil {
			return normalizeErr
		}
		input = append(input, SiteDomain{Domain: domain, CertId: certificateByDomain[domain]})
	}
	return s.updateLocked(id, &siteMutation{Domains: &input})
}

// UpdateSSL 更新 TLS 策略和列出的域名证书绑定。
func (s *service) UpdateSSL(id int64, config *TLSUpdate, bindings []SiteDomainCertificate) error {
	if config == nil && len(bindings) == 0 {
		return errors.New("SSL 配置和证书绑定不能同时为空")
	}
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	value, err := s.loadConfigLocked(id)
	if err != nil {
		return err
	}
	if config != nil {
		value, err = s.replaceHTTPConfig(value, func(httpConfig *HTTPConfig) {
			if config.RedirectHTTP != nil {
				httpConfig.TLS.RedirectHTTP = *config.RedirectHTTP
			}
			if config.MinVersion != nil {
				httpConfig.TLS.MinVersion = *config.MinVersion
			}
			if config.MaxVersion != nil {
				httpConfig.TLS.MaxVersion = *config.MaxVersion
			}
		})
		if err != nil {
			return err
		}
	}
	var domainInput *[]SiteDomain
	if len(bindings) > 0 {
		domains, queryErr := s.repositorySiteDomain.SelectBySiteId(id)
		if queryErr != nil {
			return fmt.Errorf("查询网站域名失败: %w", queryErr)
		}
		indices := make(map[int64]int, len(domains))
		updated := make([]SiteDomain, 0, len(domains))
		for _, domain := range domains {
			if domain == nil {
				continue
			}
			indices[domain.Id] = len(updated)
			updated = append(updated, SiteDomain{Domain: domain.Domain, CertId: domain.CertId})
		}
		seen := make(map[int64]struct{}, len(bindings))
		for _, binding := range bindings {
			if binding.DomainId <= 0 {
				return errors.New("域名 ID 不合法")
			}
			if binding.CertId < 0 {
				return errors.New("证书 ID 不能小于 0")
			}
			if _, exists := seen[binding.DomainId]; exists {
				return fmt.Errorf("域名证书绑定重复: %d", binding.DomainId)
			}
			seen[binding.DomainId] = struct{}{}
			index, exists := indices[binding.DomainId]
			if !exists {
				return fmt.Errorf("域名不属于当前网站: %d", binding.DomainId)
			}
			updated[index].CertId = binding.CertId
		}
		domainInput = &updated
	}
	return s.saveConfigLocked(id, value, domainInput)
}

// UpdateWAF 更新网站 WAF 设置。
func (s *service) UpdateWAF(id int64, config *WAFUpdate) error {
	if config == nil {
		return errors.New("WAF 配置不能为空")
	}
	return s.updateHTTPConfig(id, func(httpConfig *HTTPConfig) {
		current := &httpConfig.WAF
		if config.Enabled != nil {
			current.Enabled = *config.Enabled
		}
		if config.Mode != nil {
			current.Mode = *config.Mode
		}
		if config.BlockPage != nil {
			current.BlockPage = *config.BlockPage
		}
		if config.OWASP != nil {
			update := config.OWASP
			owasp := &current.OWASP
			if update.Enabled != nil {
				owasp.Enabled = *update.Enabled
			}
			if update.ParanoiaLevel != nil {
				owasp.ParanoiaLevel = *update.ParanoiaLevel
			}
			if update.DetectionParanoiaLevel != nil {
				owasp.DetectionParanoiaLevel = *update.DetectionParanoiaLevel
			}
			if update.InboundAnomalyScoreThreshold != nil {
				owasp.InboundAnomalyScoreThreshold = *update.InboundAnomalyScoreThreshold
			}
			if update.OutboundAnomalyScoreThreshold != nil {
				owasp.OutboundAnomalyScoreThreshold = *update.OutboundAnomalyScoreThreshold
			}
			if update.ReportingLevel != nil {
				owasp.ReportingLevel = update.ReportingLevel
			}
			if update.EarlyBlocking != nil {
				owasp.EarlyBlocking = *update.EarlyBlocking
			}
			if update.SamplingPercentage != nil {
				owasp.SamplingPercentage = *update.SamplingPercentage
			}
			if update.EnforceBodyProcessor != nil {
				owasp.EnforceBodyProcessor = *update.EnforceBodyProcessor
			}
			if update.ValidateUTF8 != nil {
				owasp.ValidateUTF8 = *update.ValidateUTF8
			}
			if update.SkipResponseAnalysis != nil {
				owasp.SkipResponseAnalysis = *update.SkipResponseAnalysis
			}
			if update.AllowedMethods != nil {
				owasp.AllowedMethods = *update.AllowedMethods
			}
			if update.AllowedContentTypes != nil {
				owasp.AllowedContentTypes = *update.AllowedContentTypes
			}
			if update.AllowedHTTPVersions != nil {
				owasp.AllowedHTTPVersions = *update.AllowedHTTPVersions
			}
			if update.AllowedCharsets != nil {
				owasp.AllowedCharsets = *update.AllowedCharsets
			}
			if update.RestrictedExtensions != nil {
				owasp.RestrictedExtensions = *update.RestrictedExtensions
			}
			if update.RestrictedHeaders != nil {
				owasp.RestrictedHeaders = *update.RestrictedHeaders
			}
			if update.RestrictedHeadersExtended != nil {
				owasp.RestrictedHeadersExtended = *update.RestrictedHeadersExtended
			}
			if update.MaxArguments != nil {
				owasp.MaxArguments = *update.MaxArguments
			}
			if update.MaxArgumentNameLength != nil {
				owasp.MaxArgumentNameLength = *update.MaxArgumentNameLength
			}
			if update.MaxArgumentLength != nil {
				owasp.MaxArgumentLength = *update.MaxArgumentLength
			}
			if update.TotalArgumentLength != nil {
				owasp.TotalArgumentLength = *update.TotalArgumentLength
			}
			if update.MaxFileSize != nil {
				owasp.MaxFileSize = *update.MaxFileSize
			}
			if update.CombinedFileSize != nil {
				owasp.CombinedFileSize = *update.CombinedFileSize
			}
			if update.SetupDirectives != nil {
				owasp.SetupDirectives = *update.SetupDirectives
			}
		}
		if config.AuditLog != nil {
			current.AuditLog = *config.AuditLog
		}
		if config.RequestBodyLimit != nil {
			current.RequestBodyLimit = *config.RequestBodyLimit
		}
		if config.Directives != nil {
			current.Directives = *config.Directives
		}
		if config.Rules != nil {
			current.Rules = *config.Rules
		}
	})
}

// UpdateCache 更新网站响应缓存设置。
func (s *service) UpdateCache(id int64, config *CacheUpdate) error {
	if config == nil {
		return errors.New("缓存配置不能为空")
	}
	return s.updateHTTPConfig(id, func(httpConfig *HTTPConfig) {
		current := &httpConfig.Cache
		if config.Enabled != nil {
			current.Enabled = *config.Enabled
		}
		if config.TTL != nil {
			current.TTL = *config.TTL
		}
		if config.Stale != nil {
			current.Stale = *config.Stale
		}
		if config.Methods != nil {
			current.Methods = *config.Methods
		}
		if config.KeyHeaders != nil {
			current.KeyHeaders = *config.KeyHeaders
		}
		if config.DefaultCacheControl != nil {
			current.DefaultCacheControl = *config.DefaultCacheControl
		}
		if config.MaxSize != nil {
			current.MaxSize = *config.MaxSize
		}
	})
}

// UpdateRateLimit 更新网站访问频率限制。
func (s *service) UpdateRateLimit(id int64, config *RateLimitUpdate) error {
	if config == nil {
		return errors.New("访问频率限制配置不能为空")
	}
	return s.updateHTTPConfig(id, func(httpConfig *HTTPConfig) {
		current := &httpConfig.RateLimit
		if config.Enabled != nil {
			current.Enabled = *config.Enabled
		}
		if config.Key != nil {
			current.Key = *config.Key
		}
		if config.Window != nil {
			current.Window = *config.Window
		}
		if config.MaxEvents != nil {
			current.MaxEvents = *config.MaxEvents
		}
		if config.Paths != nil {
			current.Paths = *config.Paths
		}
		if config.Methods != nil {
			current.Methods = *config.Methods
		}
		if config.IPv4Prefix != nil {
			current.IPv4Prefix = *config.IPv4Prefix
		}
		if config.IPv6Prefix != nil {
			current.IPv6Prefix = *config.IPv6Prefix
		}
	})
}

// UpdateTrafficLimit 更新网站并发和响应速度限制。
func (s *service) UpdateTrafficLimit(id int64, config *TrafficLimitUpdate) error {
	if config == nil {
		return errors.New("流量限制配置不能为空")
	}
	return s.updateHTTPConfig(id, func(httpConfig *HTTPConfig) {
		current := &httpConfig.TrafficLimit
		if config.Enabled != nil {
			current.Enabled = *config.Enabled
		}
		if config.MaxConnections != nil {
			current.MaxConnections = *config.MaxConnections
		}
		if config.MaxConnectionsPerIP != nil {
			current.MaxConnectionsPerIP = *config.MaxConnectionsPerIP
		}
		if config.RatePerRequest != nil {
			current.RatePerRequest = *config.RatePerRequest
		}
	})
}

// UpdateHeaders 更新网站请求和响应头设置。
func (s *service) UpdateHeaders(id int64, config *HeaderUpdate) error {
	if config == nil {
		return errors.New("请求头配置不能为空")
	}
	return s.updateHTTPConfig(id, func(httpConfig *HTTPConfig) {
		config.apply(&httpConfig.Headers)
	})
}

// UpdateCompression 更新网站响应压缩设置。
func (s *service) UpdateCompression(id int64, config *CompressionUpdate) error {
	if config == nil {
		return errors.New("压缩配置不能为空")
	}
	return s.updateHTTPConfig(id, func(httpConfig *HTTPConfig) {
		current := &httpConfig.Compression
		if config.Enabled != nil {
			current.Enabled = *config.Enabled
		}
		if config.Algorithms != nil {
			current.Algorithms = *config.Algorithms
		}
		if config.MinLength != nil {
			current.MinLength = *config.MinLength
		}
	})
}

// UpdateRedirects 更新网站重定向规则。
func (s *service) UpdateRedirects(id int64, redirects []RedirectConfig) error {
	return s.updateHTTPConfig(id, func(httpConfig *HTTPConfig) { httpConfig.Redirects = redirects })
}

// UpdateRoutes 更新网站自定义路由。
func (s *service) UpdateRoutes(id int64, routes []RouteConfig) error {
	return s.updateHTTPConfig(id, func(httpConfig *HTTPConfig) { httpConfig.Routes = routes })
}

// UpdateStatic 更新静态网站专属设置。
func (s *service) UpdateStatic(id int64, config *StaticUpdate) error {
	if config == nil {
		return errors.New("静态网站配置不能为空")
	}
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	value, err := s.loadConfigLocked(id)
	if err != nil {
		return err
	}
	static, ok := value.(StaticConfig)
	if !ok {
		return errors.New("当前网站不是静态网站")
	}
	if config.Index != nil {
		static.Index = *config.Index
	}
	if config.Browse != nil {
		static.Browse = *config.Browse
	}
	if config.TryFiles != nil {
		static.TryFiles = *config.TryFiles
	}
	if config.Hide != nil {
		static.Hide = *config.Hide
	}
	if config.Precompressed != nil {
		static.Precompressed = *config.Precompressed
	}
	return s.saveConfigLocked(id, static, nil)
}

// UpdateProxy 更新反向代理或通用网站的上游设置。
func (s *service) UpdateProxy(id int64, config *ProxyUpdate) error {
	if config == nil {
		return errors.New("反向代理配置不能为空")
	}
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	value, err := s.loadConfigLocked(id)
	if err != nil {
		return err
	}
	switch current := value.(type) {
	case ProxyConfig:
		config.apply(&current.Proxy)
		value = current
	case GeneralConfig:
		config.apply(&current.Proxy)
		value = current
	default:
		return errors.New("当前网站不支持反向代理设置")
	}
	return s.saveConfigLocked(id, value, nil)
}

// UpdateFastCGI 更新 PHP 网站的 FastCGI 设置。
func (s *service) UpdateFastCGI(id int64, config *FastCGIUpdate) error {
	if config == nil {
		return errors.New("PHP-FPM 配置不能为空")
	}
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	value, err := s.loadConfigLocked(id)
	if err != nil {
		return err
	}
	php, ok := value.(PHPConfig)
	if !ok {
		return errors.New("当前网站不是 PHP 网站")
	}
	config.apply(&php.FastCGI)
	return s.saveConfigLocked(id, php, nil)
}

// UpdateProcess 更新通用网站的进程保活设置。
func (s *service) UpdateProcess(id int64, config *ProcessUpdate) error {
	if config == nil {
		return errors.New("进程配置不能为空")
	}
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	value, err := s.loadConfigLocked(id)
	if err != nil {
		return err
	}
	general, ok := value.(GeneralConfig)
	if !ok {
		return errors.New("当前网站不是通用网站")
	}
	if config.Command != nil {
		general.Process.Command = *config.Command
	}
	if config.Environment != nil {
		general.Process.Environment = *config.Environment
	}
	if config.MaxRetries != nil {
		general.Process.MaxRetries = *config.MaxRetries
	}
	if config.RestartDelay != nil {
		general.Process.RestartDelay = *config.RestartDelay
	}
	if config.StopTimeout != nil {
		general.Process.StopTimeout = *config.StopTimeout
	}
	return s.saveConfigLocked(id, general, nil)
}

func (config *HeaderUpdate) apply(current *webServer.HeaderOptions) {
	if config.RequestAdd != nil {
		current.RequestAdd = *config.RequestAdd
	}
	if config.RequestSet != nil {
		current.RequestSet = *config.RequestSet
	}
	if config.RequestDelete != nil {
		current.RequestDelete = *config.RequestDelete
	}
	if config.ResponseAdd != nil {
		current.ResponseAdd = *config.ResponseAdd
	}
	if config.ResponseSet != nil {
		current.ResponseSet = *config.ResponseSet
	}
	if config.ResponseDelete != nil {
		current.ResponseDelete = *config.ResponseDelete
	}
}

func (config *ProxyUpdate) apply(current *webServer.ProxyOptions) {
	if config.Upstreams != nil {
		current.Upstreams = *config.Upstreams
	}
	if config.Scheme != nil {
		current.Scheme = *config.Scheme
	}
	if config.Policy != nil {
		current.Policy = *config.Policy
	}
	if config.Retries != nil {
		current.Retries = *config.Retries
	}
	if config.TryDuration != nil {
		current.TryDuration = *config.TryDuration
	}
	if config.TryInterval != nil {
		current.TryInterval = *config.TryInterval
	}
	if config.DialTimeout != nil {
		current.DialTimeout = *config.DialTimeout
	}
	if config.ReadTimeout != nil {
		current.ReadTimeout = *config.ReadTimeout
	}
	if config.WriteTimeout != nil {
		current.WriteTimeout = *config.WriteTimeout
	}
	if config.ResponseHeaderTimeout != nil {
		current.ResponseHeaderTimeout = *config.ResponseHeaderTimeout
	}
	if config.FlushInterval != nil {
		current.FlushInterval = *config.FlushInterval
	}
	if config.StreamTimeout != nil {
		current.StreamTimeout = *config.StreamTimeout
	}
	if config.StreamCloseDelay != nil {
		current.StreamCloseDelay = *config.StreamCloseDelay
	}
	if config.Versions != nil {
		current.Versions = *config.Versions
	}
	if config.TLSServerName != nil {
		current.TLSServerName = *config.TLSServerName
	}
	if config.TLSInsecureSkipVerify != nil {
		current.TLSInsecureSkipVerify = config.TLSInsecureSkipVerify
	}
	if config.HealthURI != nil {
		current.HealthURI = *config.HealthURI
	}
	if config.HealthInterval != nil {
		current.HealthInterval = *config.HealthInterval
	}
	if config.HealthTimeout != nil {
		current.HealthTimeout = *config.HealthTimeout
	}
	if config.HealthStatus != nil {
		current.HealthStatus = *config.HealthStatus
	}
	if config.Headers != nil {
		config.Headers.apply(&current.Headers)
	}
}

func (config *FastCGIUpdate) apply(current *webServer.ProxyOptions) {
	if config.Upstreams != nil {
		current.Upstreams = *config.Upstreams
	}
	if config.Policy != nil {
		current.Policy = *config.Policy
	}
	if config.Retries != nil {
		current.Retries = *config.Retries
	}
	if config.TryDuration != nil {
		current.TryDuration = *config.TryDuration
	}
	if config.TryInterval != nil {
		current.TryInterval = *config.TryInterval
	}
	if config.DialTimeout != nil {
		current.DialTimeout = *config.DialTimeout
	}
	if config.ReadTimeout != nil {
		current.ReadTimeout = *config.ReadTimeout
	}
	if config.WriteTimeout != nil {
		current.WriteTimeout = *config.WriteTimeout
	}
	if config.FlushInterval != nil {
		current.FlushInterval = *config.FlushInterval
	}
	if config.StreamTimeout != nil {
		current.StreamTimeout = *config.StreamTimeout
	}
	if config.StreamCloseDelay != nil {
		current.StreamCloseDelay = *config.StreamCloseDelay
	}
	if config.HealthURI != nil {
		current.HealthURI = *config.HealthURI
	}
	if config.HealthInterval != nil {
		current.HealthInterval = *config.HealthInterval
	}
	if config.HealthTimeout != nil {
		current.HealthTimeout = *config.HealthTimeout
	}
	if config.HealthStatus != nil {
		current.HealthStatus = *config.HealthStatus
	}
	if config.SplitPath != nil {
		current.SplitPath = *config.SplitPath
	}
	if config.Index != nil {
		current.Index = *config.Index
	}
	if config.TryFiles != nil {
		current.TryFiles = *config.TryFiles
	}
	if config.TryPolicy != nil {
		current.TryPolicy = *config.TryPolicy
	}
	if config.ResolveRootSymlink != nil {
		current.ResolveRootSymlink = *config.ResolveRootSymlink
	}
	if config.Hide != nil {
		current.Hide = *config.Hide
	}
	if config.Env != nil {
		current.Env = *config.Env
	}
	if config.CaptureStderr != nil {
		current.CaptureStderr = *config.CaptureStderr
	}
	if config.Headers != nil {
		config.Headers.apply(&current.Headers)
	}
}

func (s *service) updateHTTPConfig(id int64, update func(*HTTPConfig)) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	value, err := s.loadConfigLocked(id)
	if err != nil {
		return err
	}
	value, err = s.replaceHTTPConfig(value, update)
	if err != nil {
		return err
	}
	return s.saveConfigLocked(id, value, nil)
}

func (s *service) replaceHTTPConfig(value interface{}, update func(*HTTPConfig)) (interface{}, error) {
	switch config := value.(type) {
	case StaticConfig:
		update(&config.HTTPConfig)
		return config, nil
	case ProxyConfig:
		update(&config.HTTPConfig)
		return config, nil
	case PHPConfig:
		update(&config.HTTPConfig)
		return config, nil
	case GeneralConfig:
		update(&config.HTTPConfig)
		return config, nil
	default:
		return nil, errors.New("网站配置类型错误")
	}
}

func (s *service) loadConfigLocked(id int64) (interface{}, error) {
	if id <= 0 {
		return nil, errors.New("网站 ID 不合法")
	}
	record, err := s.repositorySite.FindById(id)
	if err != nil {
		return nil, s.nilIfNotFound("查询网站失败", err)
	}
	_, config, err := s.checkConfig(record.Config, record.Type)
	if err != nil {
		return nil, fmt.Errorf("网站配置无效: %w", err)
	}
	return config, nil
}

func (s *service) saveConfigLocked(id int64, config interface{}, domains *[]SiteDomain) error {
	encoded, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("网站配置格式化失败: %w", err)
	}
	raw := string(encoded)
	return s.updateLocked(id, &siteMutation{Config: &raw, Domains: domains})
}
