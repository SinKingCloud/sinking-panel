package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	serverCache "server/app/util/server/cache"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/idna"
)

func (m *Manager) cloneSite(site *Site) *Site {
	if site == nil {
		return nil
	}
	data, _ := json.Marshal(site)
	clone := new(Site)
	_ = json.Unmarshal(data, clone)
	for index := range clone.TLS.Certificates {
		clone.TLS.Certificates[index].PrivateKeyPEM = site.TLS.Certificates[index].PrivateKeyPEM
	}
	clone.TLS.coveredDomains = append([]string(nil), site.TLS.coveredDomains...)
	return clone
}

func (m *Manager) cloneSites(sites map[string]*Site) map[string]*Site {
	clone := make(map[string]*Site, len(sites))
	for id, site := range sites {
		clone[id] = m.cloneSite(site)
	}
	return clone
}

func (m *Manager) prepareSites(sites []Site) (map[string]*Site, error) {
	result := make(map[string]*Site, len(sites))
	for index := range sites {
		site, err := m.normalizeSite(&sites[index])
		if err != nil {
			return nil, err
		}
		if result[site.ID] != nil {
			return nil, fmt.Errorf("站点 ID 重复: %s", site.ID)
		}
		result[site.ID] = site
	}
	if err := m.validateSiteSet(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *Manager) normalizeSite(input *Site) (*Site, error) {
	site := m.cloneSite(input)
	site.ID = strings.TrimSpace(site.ID)
	if !m.validID(site.ID) {
		return nil, errors.New("站点 ID 只能包含字母、数字、点、下划线和短横线")
	}
	site.Name = strings.TrimSpace(site.Name)
	if site.Name == "" {
		site.Name = site.ID
	}
	if len(site.Domains) == 0 {
		return nil, fmt.Errorf("站点 %s 至少需要一个域名", site.ID)
	}
	domainSet := make(map[string]struct{}, len(site.Domains))
	site.Domains = site.Domains[:0]
	for _, value := range input.Domains {
		domain, err := m.normalizeDomain(value)
		if err != nil {
			return nil, fmt.Errorf("站点 %s 域名无效: %w", site.ID, err)
		}
		if _, exists := domainSet[domain]; exists {
			continue
		}
		domainSet[domain] = struct{}{}
		site.Domains = append(site.Domains, domain)
	}
	sort.Strings(site.Domains)
	redirectNames := make(map[string]struct{}, len(site.Redirects))
	for index := range site.Redirects {
		redirect := &site.Redirects[index]
		redirect.Name = strings.TrimSpace(redirect.Name)
		if redirect.Name == "" {
			return nil, fmt.Errorf("站点 %s 重定向 %d 名称不能为空", site.ID, index+1)
		}
		if utf8.RuneCountInString(redirect.Name) > 100 {
			return nil, fmt.Errorf("站点 %s 重定向 %d 名称不能超过 100 个字符", site.ID, index+1)
		}
		name := strings.ToLower(redirect.Name)
		if _, exists := redirectNames[name]; exists {
			return nil, fmt.Errorf("站点 %s 重定向名称重复: %s", site.ID, redirect.Name)
		}
		redirectNames[name] = struct{}{}
		domains := make([]string, 0, len(redirect.Domains))
		seenDomains := make(map[string]struct{}, len(redirect.Domains))
		for _, value := range redirect.Domains {
			domain, err := m.normalizeDomain(value)
			if err != nil {
				return nil, fmt.Errorf("站点 %s 重定向 %d 来源域名无效: %w", site.ID, index+1, err)
			}
			if _, exists := domainSet[domain]; !exists {
				return nil, fmt.Errorf("站点 %s 重定向 %d 来源域名不属于当前站点: %s", site.ID, index+1, domain)
			}
			if _, exists := seenDomains[domain]; exists {
				continue
			}
			seenDomains[domain] = struct{}{}
			domains = append(domains, domain)
		}
		redirect.Domains = domains
		paths := make([]string, 0, len(redirect.Paths))
		seenPaths := make(map[string]struct{}, len(redirect.Paths))
		for _, value := range redirect.Paths {
			if strings.ContainsAny(value, "\r\n") {
				return nil, fmt.Errorf("站点 %s 重定向 %d 来源路径不能包含换行符", site.ID, index+1)
			}
			path := strings.TrimSpace(value)
			if !strings.HasPrefix(path, "/") {
				return nil, fmt.Errorf("站点 %s 重定向 %d 来源路径必须以 / 开头: %s", site.ID, index+1, path)
			}
			if _, exists := seenPaths[path]; exists {
				continue
			}
			seenPaths[path] = struct{}{}
			paths = append(paths, path)
		}
		if len(paths) == 0 {
			paths = []string{"/*"}
		}
		redirect.Paths = paths
		if strings.ContainsAny(redirect.Target, "\r\n{}") {
			return nil, fmt.Errorf("站点 %s 重定向 %d 目标地址无效", site.ID, index+1)
		}
		redirect.Target = strings.TrimSpace(redirect.Target)
		if redirect.Target == "" {
			return nil, fmt.Errorf("站点 %s 重定向 %d 目标地址无效", site.ID, index+1)
		}
		target, err := url.Parse(redirect.Target)
		relative := strings.HasPrefix(redirect.Target, "/")
		if relative {
			if strings.HasPrefix(redirect.Target, "//") || strings.ContainsRune(redirect.Target, '\\') ||
				err != nil || target.Scheme != "" || target.Host != "" || target.User != nil {
				return nil, fmt.Errorf("站点 %s 重定向 %d 站内目标必须以单个 / 开头", site.ID, index+1)
			}
		} else if err != nil || target.Host == "" || target.User != nil || target.Hostname() == "" ||
			!strings.EqualFold(target.Scheme, "http") && !strings.EqualFold(target.Scheme, "https") {
			return nil, fmt.Errorf("站点 %s 重定向 %d 目标地址必须是绝对 HTTP、HTTPS 地址或站内路径", site.ID, index+1)
		}
		if redirect.PreserveURI && strings.ContainsAny(redirect.Target, "?#") {
			return nil, fmt.Errorf("站点 %s 重定向 %d 保留 URI 时目标地址不能包含查询参数或片段", site.ID, index+1)
		}
		if redirect.PreserveURI && relative && redirect.Target == "/" {
			return nil, fmt.Errorf("站点 %s 重定向 %d 目标为 / 时保留 URI 会造成循环", site.ID, index+1)
		}
		switch redirect.Status {
		case 301, 302, 307, 308:
		default:
			return nil, fmt.Errorf("站点 %s 重定向 %d 状态码只支持 301、302、307 或 308", site.ID, index+1)
		}
	}
	if !site.Enabled {
		site.TLS.RedirectHTTP = false
		if err := m.normalizeTLS(&site.TLS, site.Domains); err != nil {
			return nil, fmt.Errorf("站点 %s TLS 配置无效: %w", site.ID, err)
		}
		return site, nil
	}

	var err error
	if site.Root != "" {
		site.Root, err = m.resolvePath(site.Root)
		if err != nil {
			return nil, fmt.Errorf("站点 %s 根目录无效: %w", site.ID, err)
		}
	}
	if site.Proxy != nil && site.Root != "" {
		return nil, fmt.Errorf("站点 %s 不能同时配置默认根目录和反向代理", site.ID)
	}
	if site.Proxy != nil {
		if err = m.normalizeProxy(site.Proxy, site.Root); err != nil {
			return nil, fmt.Errorf("站点 %s 反向代理无效: %w", site.ID, err)
		}
	}
	if site.Root != "" {
		m.normalizeStatic(&site.Index, &site.TryFiles, &site.Hide, &site.Precompressed)
	} else if len(site.Index) > 0 || site.Browse || len(site.TryFiles) > 0 || len(site.Hide) > 0 || len(site.Precompressed) > 0 {
		return nil, fmt.Errorf("站点 %s 配置静态文件选项前必须设置根目录", site.ID)
	}
	for index := range site.Routes {
		if err = m.normalizeRoute(&site.Routes[index], site.Root); err != nil {
			return nil, fmt.Errorf("站点 %s 路由 %d 无效: %w", site.ID, index+1, err)
		}
		if index < len(site.Routes)-1 && len(site.Routes[index].Paths) == 0 && len(site.Routes[index].Methods) == 0 {
			return nil, fmt.Errorf("站点 %s 的无条件路由只能放在最后", site.ID)
		}
	}
	if site.Root == "" && site.Proxy == nil && len(site.Routes) == 0 && len(site.Handlers) == 0 {
		return nil, fmt.Errorf("站点 %s 没有可执行的处理器", site.ID)
	}
	if err = m.normalizeHeaders(&site.Headers); err != nil {
		return nil, fmt.Errorf("站点 %s 响应头配置无效: %w", site.ID, err)
	}
	if err = m.normalizeCompression(&site.Compression); err != nil {
		return nil, fmt.Errorf("站点 %s 压缩配置无效: %w", site.ID, err)
	}
	if err = m.normalizeWAF(&site.WAF); err != nil {
		return nil, fmt.Errorf("站点 %s WAF 配置无效: %w", site.ID, err)
	}
	if err = m.normalizeCache(&site.Cache, site.ID); err != nil {
		return nil, fmt.Errorf("站点 %s 缓存配置无效: %w", site.ID, err)
	}
	if err = m.normalizeRateLimit(&site.RateLimit); err != nil {
		return nil, fmt.Errorf("站点 %s 限流配置无效: %w", site.ID, err)
	}
	if err = m.normalizeTrafficLimit(&site.TrafficLimit); err != nil {
		return nil, fmt.Errorf("站点 %s 流量限制配置无效: %w", site.ID, err)
	}
	if len(site.HandlerOrder) == 0 {
		site.HandlerOrder = []Module{
			HandlerWAF,
			HandlerRateLimit,
			HandlerCache,
			HandlerSubroute,
		}
	} else {
		allowed := map[Module]struct{}{
			HandlerWAF:       {},
			HandlerRateLimit: {},
			HandlerCache:     {},
			HandlerSubroute:  {},
		}
		seen := make(map[Module]struct{}, len(site.HandlerOrder))
		order := make([]Module, 0, len(site.HandlerOrder))
		for _, name := range site.HandlerOrder {
			name = Module(strings.ToLower(strings.TrimSpace(string(name))))
			if _, exists := allowed[name]; !exists {
				return nil, fmt.Errorf("站点 %s 处理器顺序包含未知模块: %s", site.ID, name)
			}
			if _, exists := seen[name]; exists {
				return nil, fmt.Errorf("站点 %s 处理器顺序包含重复模块: %s", site.ID, name)
			}
			seen[name] = struct{}{}
			order = append(order, name)
		}
		for _, name := range []Module{HandlerWAF, HandlerRateLimit, HandlerCache, HandlerSubroute} {
			if _, exists := seen[name]; !exists {
				return nil, fmt.Errorf("站点 %s 处理器顺序缺少模块: %s", site.ID, name)
			}
		}
		if order[len(order)-1] != HandlerSubroute {
			return nil, fmt.Errorf("站点 %s 的 %s 处理器必须放在最后", site.ID, HandlerSubroute)
		}
		site.HandlerOrder = order
	}
	if err = m.normalizeTLS(&site.TLS, site.Domains); err != nil {
		return nil, fmt.Errorf("站点 %s TLS 配置无效: %w", site.ID, err)
	}
	if err = m.validateHandlers(site.Handlers); err != nil {
		return nil, fmt.Errorf("站点 %s 自定义处理器无效: %w", site.ID, err)
	}
	return site, nil
}

func (m *Manager) normalizeRoute(route *Route, siteRoot string) error {
	route.Name = strings.TrimSpace(route.Name)
	for index, path := range route.Paths {
		path = strings.TrimSpace(path)
		if !strings.HasPrefix(path, "/") {
			return fmt.Errorf("路径必须以 / 开头: %s", path)
		}
		route.Paths[index] = path
	}
	methods, err := m.normalizeMethods(route.Methods)
	if err != nil {
		return err
	}
	route.Methods = methods
	route.StripPrefix = strings.TrimSpace(route.StripPrefix)
	if route.StripPrefix != "" && !strings.HasPrefix(route.StripPrefix, "/") {
		return errors.New("移除的路径前缀必须以 / 开头")
	}
	route.Rewrite = strings.TrimSpace(route.Rewrite)
	if route.Root != "" {
		route.Root, err = m.resolvePath(route.Root)
		if err != nil {
			return err
		}
		m.normalizeStatic(&route.Index, &route.TryFiles, &route.Hide, &route.Precompressed)
	} else if len(route.Index) > 0 || route.Browse || len(route.TryFiles) > 0 || len(route.Hide) > 0 || len(route.Precompressed) > 0 {
		return errors.New("配置静态文件选项前必须设置路由根目录")
	}
	if route.Proxy != nil && route.Root != "" {
		return errors.New("同一路由不能同时配置根目录和反向代理")
	}
	if route.Proxy != nil {
		if err = m.normalizeProxy(route.Proxy, siteRoot); err != nil {
			return err
		}
	}
	origins := 0
	if route.Root != "" {
		origins++
	}
	if route.Proxy != nil {
		origins++
	}
	if route.Response != nil {
		origins++
		if route.Response.Status == 0 {
			route.Response.Status = 200
		}
		if route.Response.Status < 100 || route.Response.Status > 599 {
			return errors.New("固定响应状态码必须在 100 到 599 之间")
		}
		if err = m.validateHeaderMap(route.Response.Headers); err != nil {
			return err
		}
	}
	if origins > 1 {
		return errors.New("同一路由只能配置一种最终处理方式")
	}
	if origins == 0 && len(route.Handlers) == 0 {
		return errors.New("路由没有可执行的处理器")
	}
	return m.validateHandlers(route.Handlers)
}

func (m *Manager) normalizeProxy(proxy *ProxyOptions, siteRoot string) error {
	proxy.Transport = ProxyTransport(strings.ToLower(strings.TrimSpace(string(proxy.Transport))))
	proxy.Scheme = ProxyScheme(strings.ToLower(strings.TrimSpace(string(proxy.Scheme))))
	proxy.TLSServerName = strings.TrimSpace(proxy.TLSServerName)
	if proxy.Transport == "" {
		proxy.Transport = ProxyTransportHTTP
	}
	if proxy.Transport != ProxyTransportHTTP && proxy.Transport != ProxyTransportFastCGI {
		return errors.New("代理传输只支持 http 或 fastcgi")
	}
	if len(proxy.Upstreams) == 0 {
		return errors.New("至少需要一个上游")
	}
	detectedScheme := ProxyScheme("")
	for index := range proxy.Upstreams {
		proxy.Upstreams[index].Dial = strings.TrimSpace(proxy.Upstreams[index].Dial)
		if proxy.Upstreams[index].Dial == "" || strings.ContainsAny(proxy.Upstreams[index].Dial, "\r\n") {
			return fmt.Errorf("上游地址无效: %s", proxy.Upstreams[index].Dial)
		}
		if strings.Contains(proxy.Upstreams[index].Dial, "://") {
			if proxy.Transport != ProxyTransportHTTP {
				return errors.New("FastCGI 上游地址不能包含 URL 协议")
			}
			address, parseErr := url.Parse(proxy.Upstreams[index].Dial)
			if parseErr != nil || address == nil {
				return fmt.Errorf("上游地址无效: %s", proxy.Upstreams[index].Dial)
			}
			scheme := ProxyScheme(strings.ToLower(address.Scheme))
			if address.Hostname() == "" || address.User != nil || address.Opaque != "" ||
				(address.Path != "" && address.Path != "/") || address.RawQuery != "" || address.Fragment != "" ||
				(scheme != ProxySchemeHTTP && scheme != ProxySchemeHTTPS) {
				return fmt.Errorf("上游地址无效: %s", proxy.Upstreams[index].Dial)
			}
			if proxy.Scheme != "" && proxy.Scheme != scheme || detectedScheme != "" && detectedScheme != scheme {
				return errors.New("同一个反向代理的上游协议必须一致")
			}
			detectedScheme = scheme
			if address.Port() == "" {
				port := "80"
				if scheme == ProxySchemeHTTPS {
					port = "443"
				}
				proxy.Upstreams[index].Dial = net.JoinHostPort(address.Hostname(), port)
			} else {
				proxy.Upstreams[index].Dial = address.Host
			}
		}
		if proxy.Upstreams[index].MaxRequests < 0 {
			return errors.New("上游最大请求数不能小于 0")
		}
	}
	if proxy.Scheme == "" {
		proxy.Scheme = detectedScheme
	}
	if detectedScheme != "" {
		hasHost := false
		for name := range proxy.Headers.RequestSet {
			if strings.EqualFold(name, "Host") {
				hasHost = true
				break
			}
		}
		if !hasHost {
			if proxy.Headers.RequestSet == nil {
				proxy.Headers.RequestSet = make(map[string][]string)
			}
			proxy.Headers.RequestSet["Host"] = []string{"{http.reverse_proxy.upstream.hostport}"}
		}
	}
	proxy.Policy = strings.ToLower(strings.TrimSpace(proxy.Policy))
	if proxy.Policy == "" {
		proxy.Policy = "random"
	}
	allowedPolicies := map[string]bool{
		"random": true, "round_robin": true, "least_conn": true, "first": true,
		"ip_hash": true, "client_ip_hash": true, "uri_hash": true,
	}
	if !allowedPolicies[proxy.Policy] {
		return fmt.Errorf("不支持的负载均衡策略: %s", proxy.Policy)
	}
	for _, duration := range []time.Duration{
		proxy.TryDuration, proxy.TryInterval, proxy.DialTimeout, proxy.ReadTimeout,
		proxy.WriteTimeout, proxy.ResponseHeaderTimeout, proxy.StreamTimeout,
		proxy.StreamCloseDelay, proxy.HealthInterval, proxy.HealthTimeout,
	} {
		if duration < 0 {
			return errors.New("代理超时时间不能小于 0")
		}
	}
	if proxy.Retries < 0 {
		return errors.New("代理重试次数不能小于 0")
	}
	if proxy.HealthURI != "" && !strings.HasPrefix(proxy.HealthURI, "/") {
		return errors.New("健康检查 URI 必须以 / 开头")
	}
	if proxy.HealthStatus != 0 && (proxy.HealthStatus < 100 || proxy.HealthStatus > 599) {
		return errors.New("健康检查状态码必须在 100 到 599 之间")
	}
	if proxy.Transport == ProxyTransportFastCGI {
		if proxy.Scheme != "" || proxy.TLSServerName != "" || proxy.TLSInsecureSkipVerify != nil || len(proxy.Versions) > 0 {
			return errors.New("FastCGI 不支持 HTTP 版本或 TLS 传输选项")
		}
		if proxy.Root == "" {
			proxy.Root = siteRoot
		}
		if proxy.Root == "" {
			return errors.New("FastCGI 必须配置站点根目录")
		}
		root, err := m.resolvePath(proxy.Root)
		if err != nil {
			return err
		}
		proxy.Root = root
		if len(proxy.SplitPath) == 0 {
			proxy.SplitPath = []string{".php"}
		}
		extensionSet := make(map[string]struct{}, len(proxy.SplitPath))
		extensions := make([]string, 0, len(proxy.SplitPath))
		for _, extension := range proxy.SplitPath {
			extension = strings.ToLower(strings.TrimSpace(extension))
			if len(extension) < 2 || !strings.HasPrefix(extension, ".") || strings.ContainsAny(extension, `*?[]{}\/`) {
				return fmt.Errorf("FastCGI 脚本扩展名无效: %s", extension)
			}
			for _, character := range extension[1:] {
				if !(character >= 'a' && character <= 'z') && !(character >= '0' && character <= '9') && character != '.' && character != '-' && character != '_' {
					return fmt.Errorf("FastCGI 脚本扩展名无效: %s", extension)
				}
			}
			if _, exists := extensionSet[extension]; !exists {
				extensionSet[extension] = struct{}{}
				extensions = append(extensions, extension)
			}
		}
		proxy.SplitPath = extensions
		proxy.Index = strings.TrimSpace(proxy.Index)
		if proxy.Index == "" {
			proxy.Index = "index.php"
		}
		if proxy.Index == "." || proxy.Index == ".." || strings.ContainsAny(proxy.Index, "/\\\x00\r\n") {
			return errors.New("FastCGI 入口文件必须是根目录中的文件名")
		}
		matchesExtension := false
		for _, extension := range proxy.SplitPath {
			if strings.HasSuffix(strings.ToLower(proxy.Index), extension) {
				matchesExtension = true
				break
			}
		}
		if !matchesExtension {
			return errors.New("FastCGI 入口文件扩展名不在 SplitPath 中")
		}
		if len(proxy.TryFiles) == 0 {
			proxy.TryFiles = []string{
				"{http.request.uri.path}/" + proxy.Index,
				"{http.request.uri.path}",
				"/" + proxy.Index + "{http.request.uri.path}",
			}
		} else {
			for index, candidate := range proxy.TryFiles {
				candidate = strings.TrimSpace(candidate)
				if candidate == "" || strings.ContainsAny(candidate, "\x00\r\n") {
					return fmt.Errorf("FastCGI TryFiles 第 %d 项无效", index+1)
				}
				proxy.TryFiles[index] = candidate
			}
		}
		proxy.TryPolicy = strings.ToLower(strings.TrimSpace(proxy.TryPolicy))
		if proxy.TryPolicy == "" {
			proxy.TryPolicy = "first_exist"
		}
		if proxy.TryPolicy != "first_exist" && proxy.TryPolicy != "first_exist_fallback" {
			return errors.New("FastCGI TryPolicy 只支持 first_exist 或 first_exist_fallback")
		}
		hideSet := make(map[string]struct{}, len(proxy.Hide)+9)
		proxy.Hide = append([]string{".git", ".svn", ".hg", ".env", ".env.*", "*.sql", "*.bak", "*.backup", "*.log"}, proxy.Hide...)
		hide := make([]string, 0, len(proxy.Hide))
		for _, pattern := range proxy.Hide {
			pattern = strings.TrimSpace(pattern)
			if pattern == "" || strings.ContainsAny(pattern, "\x00\r\n") {
				return errors.New("FastCGI 静态文件隐藏规则无效")
			}
			if _, exists := hideSet[pattern]; !exists {
				hideSet[pattern] = struct{}{}
				hide = append(hide, pattern)
			}
		}
		proxy.Hide = hide
		if proxy.ResponseHeaderTimeout != 0 {
			return errors.New("ResponseHeaderTimeout 仅适用于 HTTP 代理")
		}
		for name, value := range proxy.Env {
			if strings.TrimSpace(name) == "" || strings.ContainsAny(name, "=\x00\r\n") || strings.ContainsRune(value, '\x00') {
				return fmt.Errorf("FastCGI 环境变量无效: %s", name)
			}
		}
	} else if proxy.Root != "" || len(proxy.SplitPath) > 0 || proxy.Index != "" || len(proxy.TryFiles) > 0 ||
		proxy.TryPolicy != "" || proxy.ResolveRootSymlink || len(proxy.Hide) > 0 || len(proxy.Env) > 0 || proxy.CaptureStderr {
		return errors.New("Root、SplitPath、Index、TryFiles、TryPolicy、ResolveRootSymlink、Hide、Env 和 CaptureStderr 仅适用于 FastCGI")
	} else {
		if proxy.Scheme == "" {
			proxy.Scheme = ProxySchemeHTTP
		}
		if proxy.Scheme != ProxySchemeHTTP && proxy.Scheme != ProxySchemeHTTPS {
			return fmt.Errorf("不支持的上游协议: %s", proxy.Scheme)
		}
		if proxy.Scheme == ProxySchemeHTTP && (proxy.TLSServerName != "" || proxy.TLSInsecureSkipVerify != nil && *proxy.TLSInsecureSkipVerify) {
			return errors.New("HTTP 上游不能配置 TLS 选项")
		}
		if proxy.Scheme == ProxySchemeHTTP {
			proxy.TLSInsecureSkipVerify = nil
		}
		seen := make(map[string]struct{}, len(proxy.Versions))
		versions := make([]string, 0, len(proxy.Versions))
		for _, version := range proxy.Versions {
			version = strings.ToLower(strings.TrimSpace(version))
			if version != "1.1" && version != "2" && version != "h2c" && version != "3" {
				return fmt.Errorf("不支持的上游 HTTP 版本: %s", version)
			}
			if _, exists := seen[version]; !exists {
				seen[version] = struct{}{}
				versions = append(versions, version)
			}
		}
		if len(versions) > 1 && slices.Contains(versions, "3") {
			return errors.New("上游 HTTP/3 必须单独启用")
		}
		if slices.Contains(versions, "3") && proxy.Scheme != ProxySchemeHTTPS {
			return errors.New("上游 HTTP/3 必须使用 HTTPS")
		}
		if slices.Contains(versions, "h2c") && proxy.Scheme != ProxySchemeHTTP {
			return errors.New("上游 h2c 只能使用 HTTP")
		}
		proxy.Versions = versions
		if proxy.Scheme == ProxySchemeHTTPS && proxy.TLSInsecureSkipVerify == nil {
			value := true
			proxy.TLSInsecureSkipVerify = &value
		}
	}
	if proxy.TLSServerName != "" {
		name, err := m.normalizeDomain(proxy.TLSServerName)
		if err != nil || strings.HasPrefix(name, "*.") {
			return errors.New("TLS 上游服务器名称无效")
		}
		proxy.TLSServerName = name
	}
	if err := m.normalizeHeaders(&proxy.Headers); err != nil {
		return err
	}
	return nil
}

func (m *Manager) normalizeStatic(index, tryFiles, hide, precompressed *[]string) {
	if *index == nil {
		*index = []string{"index.html", "index.htm"}
	}
	for position, value := range *index {
		(*index)[position] = strings.TrimSpace(value)
	}
	for position, value := range *tryFiles {
		(*tryFiles)[position] = strings.TrimSpace(value)
	}
	for position, value := range *hide {
		(*hide)[position] = strings.TrimSpace(value)
	}
	seen := make(map[string]struct{}, len(*precompressed))
	result := make([]string, 0, len(*precompressed))
	for _, value := range *precompressed {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "br" && value != "zstd" && value != "gzip" {
			continue
		}
		if _, exists := seen[value]; !exists {
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	*precompressed = result
}

func (m *Manager) normalizeCompression(compression *CompressionOptions) error {
	if !compression.Enabled {
		return nil
	}
	if compression.MinLength < 0 {
		return errors.New("最小压缩长度不能小于 0")
	}
	if len(compression.Algorithms) == 0 {
		compression.Algorithms = []string{"zstd", "gzip"}
	}
	seen := make(map[string]struct{}, len(compression.Algorithms))
	algorithms := make([]string, 0, len(compression.Algorithms))
	for _, algorithm := range compression.Algorithms {
		algorithm = strings.ToLower(strings.TrimSpace(algorithm))
		if algorithm != "zstd" && algorithm != "gzip" {
			return fmt.Errorf("不支持的压缩算法: %s", algorithm)
		}
		if _, exists := seen[algorithm]; !exists {
			seen[algorithm] = struct{}{}
			algorithms = append(algorithms, algorithm)
		}
	}
	compression.Algorithms = algorithms
	return nil
}

func (m *Manager) normalizeWAF(waf *WAFOptions) error {
	waf.Mode = WAFMode(strings.ToLower(strings.TrimSpace(string(waf.Mode))))
	if waf.Mode == "" {
		waf.Mode = WAFModeDetection
	}
	if waf.Mode != WAFModeDetection && waf.Mode != WAFModeBlock {
		return errors.New("WAF 模式只支持 detection 或 block")
	}
	if !waf.Enabled {
		return nil
	}
	if waf.RequestBodyLimit < 0 {
		return errors.New("请求体限制不能小于 0")
	}
	if waf.OWASP.Enabled {
		if waf.OWASP.ParanoiaLevel == 0 {
			waf.OWASP.ParanoiaLevel = 1
		}
		if waf.OWASP.ParanoiaLevel < 1 || waf.OWASP.ParanoiaLevel > 4 {
			return errors.New("OWASP CRS 防护级别只能为 1 到 4")
		}
		if waf.OWASP.DetectionParanoiaLevel == 0 {
			waf.OWASP.DetectionParanoiaLevel = waf.OWASP.ParanoiaLevel
		}
		if waf.OWASP.DetectionParanoiaLevel < waf.OWASP.ParanoiaLevel || waf.OWASP.DetectionParanoiaLevel > 4 {
			return errors.New("OWASP CRS 检测级别不能低于防护级别且不能大于 4")
		}
		if waf.OWASP.InboundAnomalyScoreThreshold == 0 {
			waf.OWASP.InboundAnomalyScoreThreshold = 5
		}
		if waf.OWASP.OutboundAnomalyScoreThreshold == 0 {
			waf.OWASP.OutboundAnomalyScoreThreshold = 4
		}
		if waf.OWASP.InboundAnomalyScoreThreshold < 1 || waf.OWASP.OutboundAnomalyScoreThreshold < 1 {
			return errors.New("OWASP CRS 异常分数阈值必须大于 0")
		}
		if waf.OWASP.ReportingLevel != nil && (*waf.OWASP.ReportingLevel < 0 || *waf.OWASP.ReportingLevel > 5) {
			return errors.New("OWASP CRS 日志级别只能为 0 到 5")
		}
		if waf.OWASP.SamplingPercentage == 0 {
			waf.OWASP.SamplingPercentage = 100
		}
		if waf.OWASP.SamplingPercentage < 1 || waf.OWASP.SamplingPercentage > 100 {
			return errors.New("OWASP CRS 采样比例只能为 1 到 100")
		}
		for _, limit := range []struct {
			name  string
			value int64
		}{
			{name: "参数数量", value: int64(waf.OWASP.MaxArguments)},
			{name: "参数名长度", value: int64(waf.OWASP.MaxArgumentNameLength)},
			{name: "参数长度", value: int64(waf.OWASP.MaxArgumentLength)},
			{name: "参数总长度", value: int64(waf.OWASP.TotalArgumentLength)},
			{name: "单文件大小", value: waf.OWASP.MaxFileSize},
			{name: "文件总大小", value: waf.OWASP.CombinedFileSize},
		} {
			if limit.value < 0 {
				return fmt.Errorf("OWASP CRS %s限制不能小于 0", limit.name)
			}
		}
		methods, err := m.normalizeMethods(waf.OWASP.AllowedMethods)
		if err != nil {
			return fmt.Errorf("OWASP CRS 允许方法无效: %w", err)
		}
		waf.OWASP.AllowedMethods = methods

		seen := make(map[string]struct{})
		values := make([]string, 0, len(waf.OWASP.AllowedContentTypes))
		for _, value := range waf.OWASP.AllowedContentTypes {
			value = strings.ToLower(strings.TrimSpace(value))
			mediaType, parameters, err := mime.ParseMediaType(value)
			if err != nil || len(parameters) != 0 || mediaType != value {
				return fmt.Errorf("OWASP CRS Content-Type 无效: %s", value)
			}
			if _, exists := seen[value]; !exists {
				seen[value] = struct{}{}
				values = append(values, value)
			}
		}
		waf.OWASP.AllowedContentTypes = values

		seen = make(map[string]struct{})
		values = make([]string, 0, len(waf.OWASP.AllowedHTTPVersions))
		versions := map[string]struct{}{"HTTP/1.0": {}, "HTTP/1.1": {}, "HTTP/2": {}, "HTTP/2.0": {}, "HTTP/3": {}, "HTTP/3.0": {}}
		for _, value := range waf.OWASP.AllowedHTTPVersions {
			value = strings.ToUpper(strings.TrimSpace(value))
			if _, exists := versions[value]; !exists {
				return fmt.Errorf("OWASP CRS HTTP 版本无效: %s", value)
			}
			if _, exists := seen[value]; !exists {
				seen[value] = struct{}{}
				values = append(values, value)
			}
		}
		waf.OWASP.AllowedHTTPVersions = values

		seen = make(map[string]struct{})
		values = make([]string, 0, len(waf.OWASP.AllowedCharsets))
		for _, value := range waf.OWASP.AllowedCharsets {
			value = strings.ToLower(strings.TrimSpace(value))
			if value == "" || strings.ContainsAny(value, " |/'\"\\\t\r\n") {
				return fmt.Errorf("OWASP CRS 字符集无效: %s", value)
			}
			if _, exists := seen[value]; !exists {
				seen[value] = struct{}{}
				values = append(values, value)
			}
		}
		waf.OWASP.AllowedCharsets = values

		seen = make(map[string]struct{})
		values = make([]string, 0, len(waf.OWASP.RestrictedExtensions))
		for _, value := range waf.OWASP.RestrictedExtensions {
			value = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(value), "/"))
			if len(value) < 2 || !strings.HasPrefix(value, ".") || strings.ContainsAny(value, " /'\"\\\t\r\n") {
				return fmt.Errorf("OWASP CRS 受限扩展名无效: %s", value)
			}
			if _, exists := seen[value]; !exists {
				seen[value] = struct{}{}
				values = append(values, value)
			}
		}
		waf.OWASP.RestrictedExtensions = values

		normalizeHeaders := func(headers []string) ([]string, error) {
			seenHeaders := make(map[string]struct{}, len(headers))
			result := make([]string, 0, len(headers))
			for _, header := range headers {
				header = strings.ToLower(strings.Trim(strings.TrimSpace(header), "/"))
				if !httpguts.ValidHeaderFieldName(header) {
					return nil, fmt.Errorf("请求头名称无效: %s", header)
				}
				if _, exists := seenHeaders[header]; !exists {
					seenHeaders[header] = struct{}{}
					result = append(result, header)
				}
			}
			return result, nil
		}
		waf.OWASP.RestrictedHeaders, err = normalizeHeaders(waf.OWASP.RestrictedHeaders)
		if err != nil {
			return fmt.Errorf("OWASP CRS 受限请求头无效: %w", err)
		}
		waf.OWASP.RestrictedHeadersExtended, err = normalizeHeaders(waf.OWASP.RestrictedHeadersExtended)
		if err != nil {
			return fmt.Errorf("OWASP CRS 扩展受限请求头无效: %w", err)
		}
		setupDirectives := make([]string, 0, len(waf.OWASP.SetupDirectives))
		for _, directive := range waf.OWASP.SetupDirectives {
			if directive = strings.TrimSpace(directive); directive != "" {
				setupDirectives = append(setupDirectives, directive)
			}
		}
		waf.OWASP.SetupDirectives = setupDirectives
	}
	hasRule := false
	targets := map[string]string{
		"REQUEST_URI":                     "REQUEST_URI",
		"REQUEST_FILENAME":                "REQUEST_FILENAME",
		"REQUEST_BASENAME":                "REQUEST_BASENAME",
		"QUERY_STRING":                    "QUERY_STRING",
		"ARGS":                            "ARGS",
		"ARGS_NAMES":                      "ARGS_NAMES",
		"ARGS_COMBINED_SIZE":              "ARGS_COMBINED_SIZE",
		"ARGS_GET":                        "ARGS_GET",
		"ARGS_GET_NAMES":                  "ARGS_GET_NAMES",
		"ARGS_POST":                       "ARGS_POST",
		"ARGS_POST_NAMES":                 "ARGS_POST_NAMES",
		"REQUEST_BODY":                    "REQUEST_BODY",
		"FILES":                           "FILES",
		"FILES_NAMES":                     "FILES_NAMES",
		"REQUEST_METHOD":                  "REQUEST_METHOD",
		"REQUEST_PROTOCOL":                "REQUEST_PROTOCOL",
		"REQUEST_HEADERS":                 "REQUEST_HEADERS",
		"REQUEST_HEADERS_NAMES":           "REQUEST_HEADERS_NAMES",
		"REQUEST_HEADERS:HOST":            "REQUEST_HEADERS:Host",
		"REQUEST_HEADERS:USER-AGENT":      "REQUEST_HEADERS:User-Agent",
		"REQUEST_HEADERS:REFERER":         "REQUEST_HEADERS:Referer",
		"REQUEST_HEADERS:CONTENT-TYPE":    "REQUEST_HEADERS:Content-Type",
		"REQUEST_HEADERS:ORIGIN":          "REQUEST_HEADERS:Origin",
		"REQUEST_HEADERS:AUTHORIZATION":   "REQUEST_HEADERS:Authorization",
		"REQUEST_HEADERS:X-FORWARDED-FOR": "REQUEST_HEADERS:X-Forwarded-For",
		"REQUEST_COOKIES":                 "REQUEST_COOKIES",
		"REQUEST_COOKIES_NAMES":           "REQUEST_COOKIES_NAMES",
		"REMOTE_ADDR":                     "REMOTE_ADDR",
	}
	normalizeConditions := func(ruleID, location string, conditions []WAFCondition) error {
		for conditionIndex := range conditions {
			condition := &conditions[conditionIndex]
			target, exists := targets[strings.ToUpper(strings.TrimSpace(condition.Target))]
			if !exists {
				return fmt.Errorf("WAF 规则 %s %s第 %d 个检查对象无效", ruleID, location, conditionIndex+1)
			}
			condition.Target = target
			condition.Operator = WAFOperator(strings.ToLower(strings.TrimSpace(string(condition.Operator))))
			switch condition.Operator {
			case WAFOperatorContains, WAFOperatorEquals, WAFOperatorStartsWith, WAFOperatorEndsWith, WAFOperatorRegex, WAFOperatorIP, WAFOperatorExists, WAFOperatorNotEmpty, WAFOperatorGreater, WAFOperatorLess:
			default:
				return fmt.Errorf("WAF 规则 %s %s第 %d 个匹配方式无效", ruleID, location, conditionIndex+1)
			}
			if condition.IgnoreCase {
				switch condition.Operator {
				case WAFOperatorContains, WAFOperatorEquals, WAFOperatorStartsWith, WAFOperatorEndsWith, WAFOperatorRegex:
				default:
					return fmt.Errorf("WAF 规则 %s %s第 %d 个匹配方式不支持忽略大小写", ruleID, location, conditionIndex+1)
				}
			}
			if strings.ContainsAny(condition.Value, "\r\n\x00") {
				return fmt.Errorf("WAF 规则 %s %s第 %d 个匹配内容包含非法字符", ruleID, location, conditionIndex+1)
			}
			condition.Value = strings.TrimSpace(condition.Value)
			if condition.Operator == WAFOperatorExists || condition.Operator == WAFOperatorNotEmpty {
				if condition.Value != "" {
					return fmt.Errorf("WAF 规则 %s %s第 %d 个匹配方式不需要匹配内容", ruleID, location, conditionIndex+1)
				}
				continue
			}
			if condition.Value == "" {
				return fmt.Errorf("WAF 规则 %s %s第 %d 个匹配内容不能为空", ruleID, location, conditionIndex+1)
			}
			if len(condition.Value) > 2048 {
				return fmt.Errorf("WAF 规则 %s %s第 %d 个匹配内容不能超过 2048 个字符", ruleID, location, conditionIndex+1)
			}
			backslashes := 0
			for valueIndex := 0; valueIndex < len(condition.Value); valueIndex++ {
				if condition.Value[valueIndex] == '\\' {
					backslashes++
					continue
				}
				if condition.Value[valueIndex] == '"' && backslashes%2 != 0 {
					return fmt.Errorf("WAF 规则 %s %s第 %d 个匹配内容中的双引号前不能包含奇数个连续反斜线", ruleID, location, conditionIndex+1)
				}
				backslashes = 0
			}
			if backslashes%2 != 0 {
				return fmt.Errorf("WAF 规则 %s %s第 %d 个匹配内容不能以奇数个连续反斜线结尾", ruleID, location, conditionIndex+1)
			}
			if condition.Operator == WAFOperatorRegex {
				if _, err := regexp.Compile(condition.Value); err != nil {
					return fmt.Errorf("WAF 规则 %s %s第 %d 个正则表达式无效: %w", ruleID, location, conditionIndex+1, err)
				}
			}
			if condition.Target == "ARGS_COMBINED_SIZE" {
				switch condition.Operator {
				case WAFOperatorEquals, WAFOperatorGreater, WAFOperatorLess:
				default:
					return fmt.Errorf("WAF 规则 %s %s第 %d 个请求参数大小关系无效", ruleID, location, conditionIndex+1)
				}
				if condition.IgnoreCase {
					return fmt.Errorf("WAF 规则 %s %s第 %d 个请求参数大小不支持忽略大小写", ruleID, location, conditionIndex+1)
				}
				value, err := strconv.ParseUint(condition.Value, 10, 64)
				if err != nil {
					return fmt.Errorf("WAF 规则 %s %s第 %d 个比较值必须是非负整数", ruleID, location, conditionIndex+1)
				}
				condition.Value = strconv.FormatUint(value, 10)
			} else if condition.Operator == WAFOperatorGreater || condition.Operator == WAFOperatorLess {
				return fmt.Errorf("WAF 规则 %s %s第 %d 个数值比较仅支持请求参数大小", ruleID, location, conditionIndex+1)
			}
			if condition.Operator == WAFOperatorIP {
				values := strings.Fields(strings.ReplaceAll(condition.Value, ",", " "))
				if len(values) == 0 {
					return fmt.Errorf("WAF 规则 %s %s第 %d 个 IP 或 CIDR 不能为空", ruleID, location, conditionIndex+1)
				}
				for _, value := range values {
					if strings.Contains(value, "/") {
						if _, err := netip.ParsePrefix(value); err != nil {
							return fmt.Errorf("WAF 规则 %s %s第 %d 个 IP 或 CIDR 无效: %s", ruleID, location, conditionIndex+1, value)
						}
					} else if _, err := netip.ParseAddr(value); err != nil {
						return fmt.Errorf("WAF 规则 %s %s第 %d 个 IP 或 CIDR 无效: %s", ruleID, location, conditionIndex+1, value)
					}
				}
				condition.Value = strings.Join(values, ",")
			}
		}
		return nil
	}
	ruleIDs := make(map[string]struct{}, len(waf.Rules))
	for index := range waf.Rules {
		rule := &waf.Rules[index]
		rule.ID = strings.TrimSpace(rule.ID)
		rule.Name = strings.TrimSpace(rule.Name)
		rule.Directive = strings.TrimSpace(rule.Directive)
		if rule.ID == "" {
			return errors.New("WAF 规则 ID 不能为空")
		}
		if strings.ContainsAny(rule.ID, "\r\n\x00") {
			return errors.New("WAF 规则 ID 包含非法字符")
		}
		if _, exists := ruleIDs[rule.ID]; exists {
			return fmt.Errorf("WAF 规则 ID 重复: %s", rule.ID)
		}
		if strings.ContainsAny(rule.Name, "\r\n\x00") {
			return fmt.Errorf("WAF 规则 %s 的名称包含非法字符", rule.ID)
		}
		ruleIDs[rule.ID] = struct{}{}
		structured := len(rule.Conditions) > 0 || len(rule.Groups) > 0
		if structured {
			if len(rule.Conditions) > 0 && len(rule.Groups) > 0 {
				return fmt.Errorf("WAF 规则 %s 不能同时配置条件和条件组", rule.ID)
			}
			rule.Match = WAFMatch(strings.ToLower(strings.TrimSpace(string(rule.Match))))
			if rule.Match == "" {
				rule.Match = WAFMatchAny
			}
			if rule.Match != WAFMatchAny && rule.Match != WAFMatchAll {
				return fmt.Errorf("WAF 规则 %s 的匹配逻辑无效", rule.ID)
			}
			rule.Action = WAFAction(strings.ToLower(strings.TrimSpace(string(rule.Action))))
			if rule.Action == "" {
				rule.Action = WAFActionBlock
			}
			if rule.Action != WAFActionBlock && rule.Action != WAFActionLog {
				return fmt.Errorf("WAF 规则 %s 的命中动作无效", rule.ID)
			}
			if len(rule.Conditions) > 0 {
				if len(rule.Conditions) > 20 {
					return fmt.Errorf("WAF 规则 %s 最多只能配置 20 个条件", rule.ID)
				}
				if err := normalizeConditions(rule.ID, "", rule.Conditions); err != nil {
					return err
				}
			} else {
				if len(rule.Groups) > 10 {
					return fmt.Errorf("WAF 规则 %s 最多只能配置 10 个条件组", rule.ID)
				}
				totalConditions := 0
				for groupIndex := range rule.Groups {
					group := &rule.Groups[groupIndex]
					group.Match = WAFMatch(strings.ToLower(strings.TrimSpace(string(group.Match))))
					if group.Match == "" {
						group.Match = WAFMatchAll
					}
					if group.Match != WAFMatchAny && group.Match != WAFMatchAll {
						return fmt.Errorf("WAF 规则 %s 的第 %d 个条件组匹配逻辑无效", rule.ID, groupIndex+1)
					}
					if len(group.Conditions) == 0 {
						return fmt.Errorf("WAF 规则 %s 的第 %d 个条件组不能为空", rule.ID, groupIndex+1)
					}
					totalConditions += len(group.Conditions)
					if totalConditions > 20 {
						return fmt.Errorf("WAF 规则 %s 最多只能配置 20 个条件", rule.ID)
					}
					if err := normalizeConditions(rule.ID, fmt.Sprintf("第 %d 个条件组的", groupIndex+1), group.Conditions); err != nil {
						return err
					}
				}
			}
			rule.Directive = ""
		} else if rule.Enabled && rule.Directive == "" {
			return fmt.Errorf("WAF 规则 %s 内容不能为空", rule.ID)
		}
		if rule.Enabled {
			hasRule = true
		}
	}
	for index := range waf.Directives {
		waf.Directives[index] = strings.TrimSpace(waf.Directives[index])
		if waf.Directives[index] != "" {
			hasRule = true
		}
	}
	if !waf.OWASP.Enabled && !hasRule {
		return errors.New("WAF 至少需要启用 OWASP CRS 或一条自定义规则")
	}
	return nil
}

func (m *Manager) normalizeCache(cache *CacheOptions, siteID string) error {
	if !cache.Enabled {
		return nil
	}
	if cache.TTL < 0 || cache.Stale < 0 {
		return errors.New("缓存时间不能小于 0")
	}
	if cache.TTL == 0 {
		cache.TTL = 2 * time.Minute
	}
	if cache.MaxSize < 0 {
		return errors.New("缓存容量不能小于 0")
	}
	if cache.MaxSize == 0 {
		cache.MaxSize = 256 << 20
	}
	methods, err := m.normalizeMethods(cache.Methods)
	if err != nil {
		return err
	}
	if len(methods) == 0 {
		methods = []string{"GET", "HEAD"}
	}
	cache.Methods = methods
	if strings.TrimSpace(cache.DefaultCacheControl) == "" {
		cache.DefaultCacheControl = "no-store"
	} else {
		cache.DefaultCacheControl = strings.TrimSpace(cache.DefaultCacheControl)
		if strings.ContainsAny(cache.DefaultCacheControl, "\r\n") {
			return errors.New("默认缓存策略不能包含换行符")
		}
	}
	cache.Name = strings.TrimSpace(cache.Name)
	if cache.Name == "" {
		cache.Name = siteID
	}
	if !m.validID(cache.Name) {
		return errors.New("缓存名称只能包含字母、数字、点、下划线和短横线")
	}
	headerSet := make(map[string]struct{}, len(cache.KeyHeaders))
	headers := make([]string, 0, len(cache.KeyHeaders))
	for _, header := range cache.KeyHeaders {
		header = strings.TrimSpace(header)
		if header == "" || strings.ContainsAny(header, "\r\n:") {
			return errors.New("缓存键请求头名称无效")
		}
		key := strings.ToLower(header)
		if _, exists := headerSet[key]; !exists {
			headerSet[key] = struct{}{}
			headers = append(headers, header)
		}
	}
	cache.KeyHeaders = headers
	return nil
}

func (m *Manager) normalizeRateLimit(limit *RateLimitOptions) error {
	limit.Key = strings.TrimSpace(limit.Key)
	if limit.Key == "" || limit.Key == "{http.request.client_ip}" {
		limit.Key = "{http.vars.client_ip}"
	}
	if !limit.Enabled {
		return nil
	}
	if limit.Window < 0 || limit.MaxEvents < 0 {
		return errors.New("限流窗口和请求数不能小于 0")
	}
	if limit.Window == 0 {
		limit.Window = time.Minute
	}
	if limit.MaxEvents == 0 {
		limit.MaxEvents = 100
	}
	if limit.IPv4Prefix < 0 || limit.IPv4Prefix > 32 || limit.IPv6Prefix < 0 || limit.IPv6Prefix > 128 {
		return errors.New("限流 IP 前缀范围无效")
	}
	for index, path := range limit.Paths {
		path = strings.TrimSpace(path)
		if !strings.HasPrefix(path, "/") {
			return fmt.Errorf("限流路径必须以 / 开头: %s", path)
		}
		limit.Paths[index] = path
	}
	methods, err := m.normalizeMethods(limit.Methods)
	if err != nil {
		return err
	}
	limit.Methods = methods
	return nil
}

func (m *Manager) normalizeTrafficLimit(limit *TrafficLimitOptions) error {
	if !limit.Enabled {
		return nil
	}
	if limit.MaxConnections < 0 || limit.MaxConnectionsPerIP < 0 || limit.RatePerRequest < 0 {
		return errors.New("并发数和响应速度不能小于 0")
	}
	if limit.MaxConnections == 0 && limit.MaxConnectionsPerIP == 0 && limit.RatePerRequest == 0 {
		return errors.New("至少需要配置一项流量限制")
	}
	return nil
}

func (m *Manager) normalizeTLS(options *TLSOptions, domains []string) error {
	options.coveredDomains = nil
	options.MinVersion = strings.ToLower(strings.TrimSpace(options.MinVersion))
	options.MaxVersion = strings.ToLower(strings.TrimSpace(options.MaxVersion))
	if options.MinVersion != "" && options.MinVersion != "tls1.2" && options.MinVersion != "tls1.3" {
		return errors.New("TLS 最低版本只支持 tls1.2 或 tls1.3")
	}
	if options.MaxVersion != "" && options.MaxVersion != "tls1.2" && options.MaxVersion != "tls1.3" {
		return errors.New("TLS 最高版本只支持 tls1.2 或 tls1.3")
	}
	if options.MinVersion == "tls1.3" && options.MaxVersion == "tls1.2" {
		return errors.New("TLS 最高版本不能低于最低版本")
	}
	if !options.Enabled {
		if options.RedirectHTTP {
			return errors.New("启用 HTTPS 后才能开启 HTTP 跳转")
		}
		return nil
	}
	if len(options.Certificates) == 0 {
		return errors.New("启用 TLS 后必须配置证书")
	}
	covered := make(map[string]bool, len(domains))
	siteDomains := make(map[string]struct{}, len(domains))
	domainCertificates := make(map[string]int, len(domains))
	for _, domain := range domains {
		siteDomains[domain] = struct{}{}
	}
	for index := range options.Certificates {
		certificate := &options.Certificates[index]
		certificate.CertificateFile = strings.TrimSpace(certificate.CertificateFile)
		certificate.KeyFile = strings.TrimSpace(certificate.KeyFile)
		certificate.CertificatePEM = strings.TrimSpace(certificate.CertificatePEM)
		certificate.PrivateKeyPEM = strings.TrimSpace(certificate.PrivateKeyPEM)
		explicitDomains := len(certificate.Domains) > 0
		targetDomains := make([]string, 0, len(certificate.Domains))
		if explicitDomains {
			seen := make(map[string]struct{}, len(certificate.Domains))
			for _, value := range certificate.Domains {
				domain, err := m.normalizeDomain(value)
				if err != nil {
					return fmt.Errorf("第 %d 组证书域名无效: %w", index+1, err)
				}
				if _, exists := siteDomains[domain]; !exists {
					return fmt.Errorf("第 %d 组证书域名不属于当前站点: %s", index+1, domain)
				}
				if _, exists := seen[domain]; exists {
					continue
				}
				seen[domain] = struct{}{}
				targetDomains = append(targetDomains, domain)
			}
			sort.Strings(targetDomains)
			certificate.Domains = append([]string(nil), targetDomains...)
		}
		usesFiles := certificate.CertificateFile != "" || certificate.KeyFile != ""
		usesPEM := certificate.CertificatePEM != "" || certificate.PrivateKeyPEM != ""
		if usesFiles == usesPEM {
			return fmt.Errorf("第 %d 组证书必须且只能配置文件路径或 PEM 内容", index+1)
		}
		if usesFiles {
			if certificate.CertificateFile == "" || certificate.KeyFile == "" {
				return fmt.Errorf("第 %d 组证书文件和私钥文件必须同时配置", index+1)
			}
			certificateFile, err := m.resolvePath(certificate.CertificateFile)
			if err != nil {
				return fmt.Errorf("证书路径无效: %w", err)
			}
			keyFile, err := m.resolvePath(certificate.KeyFile)
			if err != nil {
				return fmt.Errorf("私钥路径无效: %w", err)
			}
			certificate.CertificateFile = certificateFile
			certificate.KeyFile = keyFile
		} else if certificate.CertificatePEM == "" || certificate.PrivateKeyPEM == "" {
			return fmt.Errorf("第 %d 组证书内容和私钥内容必须同时配置", index+1)
		}
		certificatePEM := []byte(certificate.CertificatePEM)
		privateKeyPEM := []byte(certificate.PrivateKeyPEM)
		if usesFiles {
			var err error
			certificatePEM, err = os.ReadFile(certificate.CertificateFile)
			if err != nil {
				return fmt.Errorf("读取证书失败: %w", err)
			}
			privateKeyPEM, err = os.ReadFile(certificate.KeyFile)
			if err != nil {
				return fmt.Errorf("读取私钥失败: %w", err)
			}
		}
		pair, err := tls.X509KeyPair(certificatePEM, privateKeyPEM)
		if err != nil {
			return fmt.Errorf("证书和私钥不匹配: %w", err)
		}
		leaf, err := x509.ParseCertificate(pair.Certificate[0])
		if err != nil {
			return fmt.Errorf("解析证书失败: %w", err)
		}
		candidateDomains := targetDomains
		if !explicitDomains {
			candidateDomains = domains
		}
		pairCovered := make([]string, 0, len(candidateDomains))
		for _, domain := range candidateDomains {
			if net.ParseIP(domain) != nil {
				if explicitDomains {
					return fmt.Errorf("第 %d 组证书不覆盖绑定域名: %s", index+1, domain)
				}
				continue
			}
			matched := false
			if strings.HasPrefix(domain, "*.") {
				for _, name := range leaf.DNSNames {
					if strings.EqualFold(name, domain) {
						matched = true
						break
					}
				}
			} else if leaf.VerifyHostname(domain) == nil {
				matched = true
			}
			if matched {
				pairCovered = append(pairCovered, domain)
				continue
			}
			if explicitDomains {
				return fmt.Errorf("第 %d 组证书不覆盖绑定域名: %s", index+1, domain)
			}
		}
		if !explicitDomains {
			if len(pairCovered) == 0 {
				return fmt.Errorf("第 %d 组证书不覆盖当前站点的任何域名", index+1)
			}
			targetDomains = pairCovered
			certificate.Domains = append([]string(nil), targetDomains...)
		}
		for _, domain := range targetDomains {
			if owner, exists := domainCertificates[domain]; exists {
				return fmt.Errorf("域名 %s 同时绑定了第 %d 和第 %d 组证书", domain, owner+1, index+1)
			}
			domainCertificates[domain] = index
			covered[domain] = true
		}
	}
	for _, domain := range domains {
		if covered[domain] {
			options.coveredDomains = append(options.coveredDomains, domain)
		}
	}
	if len(options.coveredDomains) == 0 {
		return errors.New("没有证书覆盖站点域名")
	}
	return nil
}

func (m *Manager) normalizeHeaders(headers *HeaderOptions) error {
	for _, values := range []map[string][]string{headers.RequestAdd, headers.RequestSet, headers.ResponseAdd, headers.ResponseSet} {
		if err := m.validateHeaderMap(values); err != nil {
			return err
		}
	}
	for _, values := range [][]string{headers.RequestDelete, headers.ResponseDelete} {
		for _, value := range values {
			if strings.TrimSpace(value) == "" || strings.ContainsAny(value, "\r\n") {
				return errors.New("请求头名称无效")
			}
		}
	}
	return nil
}

func (m *Manager) validateHeaderMap(headers map[string][]string) error {
	for name, values := range headers {
		if strings.TrimSpace(name) == "" || strings.ContainsAny(name, "\r\n:") {
			return fmt.Errorf("请求头名称无效: %s", name)
		}
		for _, value := range values {
			if strings.ContainsAny(value, "\r\n") {
				return fmt.Errorf("请求头 %s 包含换行符", name)
			}
		}
	}
	return nil
}

func (m *Manager) validateHandlers(handlers []json.RawMessage) error {
	for index, raw := range handlers {
		var handler map[string]interface{}
		if err := json.Unmarshal(raw, &handler); err != nil {
			return fmt.Errorf("处理器 %d JSON 无效: %w", index+1, err)
		}
		name, ok := handler["handler"].(string)
		if !ok || strings.TrimSpace(name) == "" {
			return fmt.Errorf("处理器 %d 缺少 handler", index+1)
		}
	}
	return nil
}

func (m *Manager) normalizeMethods(methods []string) ([]string, error) {
	seen := make(map[string]struct{}, len(methods))
	result := make([]string, 0, len(methods))
	for _, method := range methods {
		method = strings.ToUpper(strings.TrimSpace(method))
		if method == "" {
			continue
		}
		for _, character := range method {
			if !(character >= 'A' && character <= 'Z') && character != '-' {
				return nil, fmt.Errorf("HTTP 方法无效: %s", method)
			}
		}
		if _, exists := seen[method]; !exists {
			seen[method] = struct{}{}
			result = append(result, method)
		}
	}
	return result, nil
}

func (m *Manager) normalizeDomain(value string) (string, error) {
	domain := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(value, ".")))
	if strings.HasPrefix(domain, "[") && strings.HasSuffix(domain, "]") {
		domain = strings.TrimSuffix(strings.TrimPrefix(domain, "["), "]")
	}
	if domain == "" || strings.ContainsAny(domain, "/@?#[]\r\n") {
		return "", errors.New("域名格式错误")
	}
	wildcard := strings.HasPrefix(domain, "*.")
	if wildcard {
		domain = strings.TrimPrefix(domain, "*.")
	}
	if address := net.ParseIP(domain); address != nil {
		if wildcard {
			return "", errors.New("IP 地址不支持通配符")
		}
		return address.String(), nil
	}
	if strings.Contains(domain, ":") {
		return "", errors.New("域名格式错误")
	}
	ascii, err := idna.Lookup.ToASCII(domain)
	if err != nil {
		return "", err
	}
	domain = ascii
	if len(domain) > 253 {
		return "", errors.New("域名过长")
	}
	for _, label := range strings.Split(domain, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", errors.New("域名标签格式错误")
		}
		for _, character := range label {
			if !(character >= 'a' && character <= 'z') && !(character >= '0' && character <= '9') && character != '-' {
				return "", errors.New("域名包含无效字符")
			}
		}
	}
	if wildcard {
		return "*." + domain, nil
	}
	return domain, nil
}

func (m *Manager) validateSiteSet(sites map[string]*Site) error {
	domains := make(map[string]string)
	identifiers := make(map[string]string, len(sites))
	contentRoots := make([]string, 0, len(sites))
	certificateFiles := make([]string, 0)
	canonicalPath := func(path string) string {
		path = filepath.Clean(path)
		if resolved, err := filepath.EvalSymlinks(path); err == nil {
			return filepath.Clean(resolved)
		}
		if parent, err := filepath.EvalSymlinks(filepath.Dir(path)); err == nil {
			return filepath.Join(parent, filepath.Base(path))
		}
		return path
	}
	siteLogDirectories := make([]string, 0, len(sites))
	siteLogDirectorySet := make(map[string]struct{}, len(sites))
	for _, site := range sites {
		identifier := strings.ToLower(site.ID)
		if owner := identifiers[identifier]; owner != "" && owner != site.ID {
			return fmt.Errorf("站点 ID %s 和 %s 在当前文件系统上可能冲突", owner, site.ID)
		}
		identifiers[identifier] = site.ID
		logPath, err := m.siteLogPath(site.ID, LogAccess)
		if err != nil {
			return err
		}
		logDirectory := canonicalPath(filepath.Dir(logPath))
		if _, exists := siteLogDirectorySet[logDirectory]; !exists {
			siteLogDirectorySet[logDirectory] = struct{}{}
			siteLogDirectories = append(siteLogDirectories, logDirectory)
		}
		for _, domain := range site.Domains {
			if owner := domains[domain]; owner != "" && owner != site.ID {
				return fmt.Errorf("域名 %s 同时属于站点 %s 和 %s", domain, owner, site.ID)
			}
			domains[domain] = site.ID
		}
		if site.Root != "" {
			contentRoots = append(contentRoots, canonicalPath(site.Root))
		}
		if site.Proxy != nil && site.Proxy.Root != "" {
			contentRoots = append(contentRoots, canonicalPath(site.Proxy.Root))
		}
		for _, route := range site.Routes {
			if route.Root != "" {
				contentRoots = append(contentRoots, canonicalPath(route.Root))
			}
			if route.Proxy != nil && route.Proxy.Root != "" {
				contentRoots = append(contentRoots, canonicalPath(route.Proxy.Root))
			}
		}
		for _, pair := range site.TLS.Certificates {
			for _, file := range []string{pair.CertificateFile, pair.KeyFile} {
				if file != "" {
					certificateFiles = append(certificateFiles, canonicalPath(file))
				}
			}
		}
	}
	pathInside := func(parent, child string) bool {
		relative, err := filepath.Rel(parent, child)
		return err == nil && (relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))))
	}
	managedFiles := make([]string, 0, 3)
	for _, file := range []string{m.logPath, m.wafLogPath, m.configPath} {
		if file != "" && file != "-" {
			managedFiles = append(managedFiles, canonicalPath(file))
		}
	}
	for _, directory := range siteLogDirectories {
		if pathInside(directory, m.cachePath) || pathInside(m.cachePath, directory) ||
			pathInside(directory, m.dataPath) || pathInside(m.dataPath, directory) ||
			pathInside(directory, m.acmePath) || pathInside(m.acmePath, directory) {
			return errors.New("站点日志目录不能与 http 缓存、运行数据或 ACME 数据目录重叠")
		}
		for _, file := range managedFiles {
			if pathInside(directory, file) || pathInside(file, directory) {
				return errors.New("站点日志目录不能与 http 日志或配置文件重叠")
			}
		}
		for _, file := range certificateFiles {
			if pathInside(directory, file) || pathInside(file, directory) {
				return errors.New("站点日志目录不能与 TLS 证书或私钥重叠")
			}
		}
	}
	for _, root := range contentRoots {
		if pathInside(root, m.cachePath) || pathInside(m.cachePath, root) || pathInside(root, m.dataPath) || pathInside(m.dataPath, root) ||
			pathInside(root, m.acmePath) || pathInside(m.acmePath, root) {
			return fmt.Errorf("站点内容目录 %s 不能与 http 缓存、运行数据或 ACME 数据目录重叠", root)
		}
		for _, directory := range siteLogDirectories {
			if pathInside(root, directory) || pathInside(directory, root) {
				return fmt.Errorf("站点内容目录 %s 不能与站点日志目录重叠", root)
			}
		}
		for _, file := range managedFiles {
			if pathInside(root, file) {
				return fmt.Errorf("站点内容目录 %s 不能包含 http 日志或配置文件", root)
			}
		}
		for _, file := range certificateFiles {
			if file != "" && pathInside(root, file) {
				return fmt.Errorf("站点内容目录 %s 不能包含 TLS 证书或私钥", root)
			}
		}
	}
	for _, file := range certificateFiles {
		if pathInside(m.cachePath, file) {
			return errors.New("TLS 证书和私钥不能位于 CachePath 内")
		}
		for _, managed := range managedFiles {
			if pathInside(file, managed) || pathInside(managed, file) {
				return errors.New("TLS 证书或私钥不能与 http 日志或配置文件重叠")
			}
		}
	}
	return nil
}

func (m *Manager) validID(id string) bool {
	if id == "" || len(id) > 128 || id == "." || id == ".." {
		return false
	}
	for _, character := range id {
		if !(character >= 'a' && character <= 'z') && !(character >= 'A' && character <= 'Z') &&
			!(character >= '0' && character <= '9') && character != '.' && character != '_' && character != '-' {
			return false
		}
	}
	return true
}

func (m *Manager) siteLogPath(id string, logType LogType) (string, error) {
	id = strings.TrimSpace(id)
	if !m.validID(id) {
		return "", errors.New("站点 ID 无效")
	}
	fileName := ""
	switch logType {
	case LogAccess:
		fileName = "http.log"
	case LogWAF:
		fileName = "waf.log"
	case LogProcess:
		fileName = "process.log"
	default:
		return "", errors.New("站点日志类型无效")
	}
	return filepath.Join(m.root, id, fileName), nil
}

func (m *Manager) ensureSiteLogDirectories(sites map[string]*Site) error {
	for id := range sites {
		path, err := m.siteLogPath(id, LogAccess)
		if err != nil {
			return err
		}
		directory := filepath.Dir(path)
		if err = os.MkdirAll(directory, 0700); err != nil {
			return fmt.Errorf("创建站点 %s 日志目录失败: %w", id, err)
		}
		info, err := os.Lstat(directory)
		if err != nil {
			return fmt.Errorf("检查站点 %s 日志目录失败: %w", id, err)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("站点 %s 日志目录无效", id)
		}
		if err = os.Chmod(directory, 0700); err != nil {
			return fmt.Errorf("设置站点 %s 日志目录权限失败: %w", id, err)
		}
	}
	return nil
}

func (m *Manager) resolvePath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("路径不能为空")
	}
	if !filepath.IsAbs(value) {
		value = filepath.Join(m.root, value)
	}
	path, err := filepath.Abs(value)
	if err != nil {
		return "", err
	}
	return filepath.Clean(path), nil
}

func (m *Manager) setSiteEnabled(id string, enabled bool) error {
	m.operationMu.Lock()
	defer m.operationMu.Unlock()
	id = strings.TrimSpace(id)
	m.mu.RLock()
	if m.configMode {
		m.mu.RUnlock()
		return errors.New("配置文件模式下请使用 SyncSites 切换为站点模式")
	}
	candidate := m.cloneSites(m.sites)
	m.mu.RUnlock()
	site := candidate[id]
	if site == nil {
		return errors.New("站点不存在")
	}
	if site.Enabled == enabled {
		return nil
	}
	site.Enabled = enabled
	return m.applySites(candidate)
}

func (m *Manager) removeSiteCache(id string) error {
	target, err := m.siteCachePath(id)
	if err != nil {
		return err
	}
	return serverCache.Registry.RemoveTree(m.cachePath, target)
}

func (m *Manager) siteCachePath(id string) (string, error) {
	if !m.validID(id) {
		return "", errors.New("站点 ID 无效")
	}
	target := filepath.Join(m.cachePath, id)
	relative, err := filepath.Rel(m.cachePath, target)
	if err != nil || relative == "." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || relative == ".." {
		return "", errors.New("站点缓存路径不安全")
	}
	return target, nil
}
