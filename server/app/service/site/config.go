package site

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"server/app/enum/site_type"
	webServer "server/app/util/server"
	"strings"
	"time"
)

// createConfig 根据网站类型生成初始配置，不允许混用其他类型的专属配置。
func (s *service) createConfig(data *CreateSite) (string, error) {
	if data == nil {
		return "", errors.New("网站参数不能为空")
	}
	if data.Type != site_type.Static && data.Static != nil {
		return "", errors.New("当前网站类型不支持静态网站配置")
	}
	if data.Type != site_type.Proxy && data.Type != site_type.General && data.Proxy != nil {
		return "", errors.New("当前网站类型不支持反向代理配置")
	}
	if data.Type != site_type.PHP && data.FastCGI != nil {
		return "", errors.New("当前网站类型不支持 FastCGI 配置")
	}
	if data.Type != site_type.General && data.Process != nil {
		return "", errors.New("当前网站类型不支持进程配置")
	}

	var config interface{}
	switch data.Type {
	case site_type.Static:
		if data.Static == nil {
			return "", errors.New("静态网站配置不能为空")
		}
		config = StaticConfig{
			Index:         data.Static.Index,
			Browse:        data.Static.Browse,
			TryFiles:      data.Static.TryFiles,
			Hide:          data.Static.Hide,
			Precompressed: data.Static.Precompressed,
		}
	case site_type.Proxy:
		if data.Proxy == nil {
			return "", errors.New("反向代理配置不能为空")
		}
		config = ProxyConfig{Proxy: *data.Proxy}
	case site_type.PHP:
		if data.FastCGI == nil {
			return "", errors.New("FastCGI 配置不能为空")
		}
		config = PHPConfig{FastCGI: *data.FastCGI}
	case site_type.General:
		if data.Proxy == nil {
			return "", errors.New("通用网站反向代理配置不能为空")
		}
		if data.Process == nil {
			return "", errors.New("通用网站进程配置不能为空")
		}
		config = GeneralConfig{Proxy: *data.Proxy, Process: *data.Process}
	default:
		return "", errors.New("网站类型不合法")
	}
	encoded, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("网站配置格式化失败: %w", err)
	}
	return string(encoded), nil
}

// formatConfig 将数据库 JSON 转换为对应网站类型的结构体。
func (s *service) formatConfig(raw string, siteType int) (interface{}, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = "{}"
	}
	var target interface{}
	switch siteType {
	case site_type.Static:
		target = &StaticConfig{}
	case site_type.Proxy:
		target = &ProxyConfig{}
	case site_type.General:
		target = &GeneralConfig{}
	case site_type.PHP:
		target = &PHPConfig{}
	default:
		return nil, errors.New("网站类型不合法")
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return nil, fmt.Errorf("网站配置格式错误: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("网站配置只能包含一个 JSON 对象")
		}
		return nil, fmt.Errorf("网站配置包含多余内容: %w", err)
	}
	switch data := target.(type) {
	case *StaticConfig:
		return *data, nil
	case *ProxyConfig:
		return *data, nil
	case *GeneralConfig:
		return *data, nil
	case *PHPConfig:
		return *data, nil
	default:
		return nil, errors.New("网站配置类型错误")
	}
}

// checkConfig 校验并返回规范化 JSON。
func (s *service) checkConfig(raw string, siteType int) (string, interface{}, error) {
	value, err := s.formatConfig(raw, siteType)
	if err != nil {
		return "", nil, err
	}
	normalizeHTTP := func(config *HTTPConfig) {
		config.Cache.Name = ""
	}
	switch config := value.(type) {
	case StaticConfig:
		normalizeHTTP(&config.HTTPConfig)
		value = config
	case ProxyConfig:
		normalizeHTTP(&config.HTTPConfig)
		if len(config.Proxy.Upstreams) == 0 {
			return "", nil, errors.New("反向代理至少需要一个上游")
		}
		if config.Proxy.Transport != "" && config.Proxy.Transport != webServer.ProxyTransportHTTP {
			return "", nil, errors.New("反向代理网站只支持 HTTP 传输")
		}
		config.Proxy.Transport = webServer.ProxyTransportHTTP
		value = config
	case PHPConfig:
		normalizeHTTP(&config.HTTPConfig)
		if len(config.FastCGI.Upstreams) == 0 {
			return "", nil, errors.New("PHP 网站至少需要一个 PHP-FPM 上游")
		}
		config.FastCGI.Transport = webServer.ProxyTransportFastCGI
		config.FastCGI.Scheme = ""
		config.FastCGI.Root = ""
		config.FastCGI.TLSServerName = ""
		config.FastCGI.TLSInsecureSkipVerify = nil
		config.FastCGI.Versions = nil
		value = config
	case GeneralConfig:
		normalizeHTTP(&config.HTTPConfig)
		config.Process.Command = strings.TrimSpace(config.Process.Command)
		if config.Process.Command == "" {
			return "", nil, errors.New("通用网站启动命令不能为空")
		}
		if len(config.Proxy.Upstreams) == 0 {
			return "", nil, errors.New("通用网站至少需要一个本地 HTTP 上游")
		}
		if config.Proxy.Transport != "" && config.Proxy.Transport != webServer.ProxyTransportHTTP {
			return "", nil, errors.New("通用网站只支持 HTTP 传输")
		}
		if config.Proxy.Scheme != "" && config.Proxy.Scheme != webServer.ProxySchemeHTTP {
			return "", nil, errors.New("通用网站只支持本机 HTTP 上游")
		}
		for _, upstream := range config.Proxy.Upstreams {
			address := strings.TrimSpace(upstream.Dial)
			if strings.HasPrefix(address, "unix/") || strings.HasPrefix(address, "unix+h2c/") {
				continue
			}
			host := ""
			if strings.Contains(address, "://") {
				parsed, parseErr := url.Parse(address)
				if parseErr != nil || parsed == nil || !strings.EqualFold(parsed.Scheme, string(webServer.ProxySchemeHTTP)) {
					return "", nil, fmt.Errorf("通用网站上游必须使用本机 HTTP: %s", address)
				}
				host = parsed.Hostname()
			} else {
				var parseErr error
				host, _, parseErr = net.SplitHostPort(address)
				if parseErr != nil {
					return "", nil, fmt.Errorf("通用网站上游地址无效: %s", address)
				}
			}
			if strings.EqualFold(host, "localhost") || host == "" {
				continue
			}
			if zone := strings.LastIndex(host, "%"); zone >= 0 {
				host = host[:zone]
			}
			ip := net.ParseIP(host)
			if ip == nil || !ip.IsLoopback() && !ip.IsUnspecified() {
				return "", nil, fmt.Errorf("通用网站上游必须指向本机: %s", address)
			}
		}
		if config.Process.RestartDelay < 0 || config.Process.StopTimeout < 0 {
			return "", nil, errors.New("进程重启和停止时间不能小于 0")
		}
		if config.Process.RestartDelay == 0 {
			config.Process.RestartDelay = 2 * time.Second
		}
		if config.Process.StopTimeout == 0 {
			config.Process.StopTimeout = 10 * time.Second
		}
		environment := make(map[string]string, len(config.Process.Environment))
		for name, content := range config.Process.Environment {
			name = strings.TrimSpace(name)
			if name == "" || strings.ContainsAny(name, "=\x00\r\n") || strings.ContainsRune(content, '\x00') {
				return "", nil, fmt.Errorf("进程环境变量无效: %s", name)
			}
			if _, exists := environment[name]; exists {
				return "", nil, fmt.Errorf("进程环境变量重复: %s", name)
			}
			environment[name] = content
		}
		config.Process.Environment = environment
		config.Proxy.Transport = webServer.ProxyTransportHTTP
		config.Proxy.Scheme = webServer.ProxySchemeHTTP
		value = config
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", nil, fmt.Errorf("网站配置格式化失败: %w", err)
	}
	return string(encoded), value, nil
}
