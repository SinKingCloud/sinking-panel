package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/certmagic"
)

func (m *Manager) setOptions(options Options) error {
	if options.HTTPListen == nil {
		options.HTTPListen = []string{":80"}
	} else {
		options.HTTPListen = append([]string{}, options.HTTPListen...)
	}
	if options.HTTPSListen == nil {
		options.HTTPSListen = []string{":443"}
	} else {
		options.HTTPSListen = append([]string{}, options.HTTPSListen...)
	}
	if options.Protocols == nil {
		options.Protocols = []string{"h1", "h2", "h3"}
	} else {
		options.Protocols = append([]string{}, options.Protocols...)
	}
	allowedProtocols := map[string]bool{"h1": true, "h2": true, "h2c": true, "h3": true}
	protocolSet := make(map[string]struct{}, len(options.Protocols))
	protocols := make([]string, 0, len(options.Protocols))
	for _, protocol := range options.Protocols {
		protocol = strings.ToLower(strings.TrimSpace(protocol))
		if !allowedProtocols[protocol] {
			return fmt.Errorf("不支持的 HTTP 协议: %s", protocol)
		}
		if _, exists := protocolSet[protocol]; !exists {
			protocolSet[protocol] = struct{}{}
			protocols = append(protocols, protocol)
		}
	}
	if _, h1 := protocolSet["h1"]; !h1 {
		if _, h2 := protocolSet["h2"]; h2 {
			return errors.New("启用 h2 时必须同时启用 h1")
		}
		if _, h2c := protocolSet["h2c"]; h2c {
			return errors.New("启用 h2c 时必须同时启用 h1")
		}
	}
	options.Protocols = protocols
	options.DefaultSite = strings.TrimSpace(options.DefaultSite)
	if options.DefaultSite != "" && !m.validID(options.DefaultSite) {
		return errors.New("默认站点 ID 不合法")
	}
	normalizePage := func(page *ResponseOptions, defaultStatus int, body string) error {
		if page.Status == 0 {
			page.Status = defaultStatus
		}
		if page.Status < 100 || page.Status > 599 {
			return errors.New("页面响应状态码必须在 100 到 599 之间")
		}
		if strings.TrimSpace(page.Body) == "" {
			page.Body = body
		}
		if err := m.validateHeaderMap(page.Headers); err != nil {
			return err
		}
		headers := make(map[string][]string, len(page.Headers)+2)
		hasContentType := false
		hasCacheControl := false
		for name, values := range page.Headers {
			headers[name] = append([]string(nil), values...)
			hasContentType = hasContentType || strings.EqualFold(name, "Content-Type")
			hasCacheControl = hasCacheControl || strings.EqualFold(name, "Cache-Control")
		}
		if !hasContentType {
			headers["Content-Type"] = []string{"text/html; charset=utf-8"}
		}
		if !hasCacheControl {
			headers["Cache-Control"] = []string{"private, no-store"}
		}
		page.Headers = headers
		return nil
	}
	for _, page := range []struct {
		value  *ResponseOptions
		status int
		body   string
	}{
		{value: &options.NotFoundPage, status: 404, body: DefaultNotFoundPage},
		{value: &options.SiteNotFoundPage, status: 404, body: DefaultSiteNotFoundPage},
		{value: &options.SiteDisabledPage, status: 503, body: DefaultSiteDisabledPage},
	} {
		if err := normalizePage(page.value, page.status, page.body); err != nil {
			return err
		}
	}
	for index := range options.HTTPListen {
		options.HTTPListen[index] = strings.TrimSpace(options.HTTPListen[index])
	}
	for index := range options.HTTPSListen {
		options.HTTPSListen[index] = strings.TrimSpace(options.HTTPSListen[index])
	}
	for _, group := range []struct {
		addresses   []string
		defaultPort uint
	}{
		{addresses: options.HTTPListen, defaultPort: 80},
		{addresses: options.HTTPSListen, defaultPort: 443},
	} {
		for index, address := range group.addresses {
			if address == "" {
				return fmt.Errorf("监听地址 %d 不能为空", index+1)
			}
			parsed, parseErr := caddy.ParseNetworkAddressWithDefaults(address, "tcp", group.defaultPort)
			if parseErr != nil {
				return fmt.Errorf("监听地址 %s 无效: %w", address, parseErr)
			}
			group.addresses[index] = parsed.String()
		}
	}
	if options.ReadTimeout < 0 || options.ReadHeaderTimeout < 0 || options.WriteTimeout < 0 ||
		options.IdleTimeout < 0 || options.GracePeriod < 0 {
		return errors.New("http 超时时间不能小于 0")
	}
	if options.ReadHeaderTimeout == 0 {
		options.ReadHeaderTimeout = 10 * time.Second
	}
	if options.IdleTimeout == 0 {
		options.IdleTimeout = 5 * time.Minute
	}
	if options.GracePeriod == 0 {
		options.GracePeriod = 10 * time.Second
	}
	if options.MaxHeaderBytes < 0 {
		return errors.New("请求头大小不能小于 0")
	}
	if options.MaxHeaderBytes == 0 {
		options.MaxHeaderBytes = 1 << 20
	}
	if options.HTTPChallengePort < 0 || options.HTTPChallengePort > 65535 {
		return errors.New("ACME HTTP 验证端口无效")
	}
	if options.HTTPChallengePort == 0 {
		options.HTTPChallengePort = 80
	}
	for _, address := range options.HTTPSListen {
		parsed, _ := caddy.ParseNetworkAddressWithDefaults(address, "tcp", 443)
		if parsed.StartPort > 0 && uint(options.HTTPChallengePort) >= parsed.StartPort && uint(options.HTTPChallengePort) <= parsed.EndPort {
			return fmt.Errorf("ACME HTTP 验证端口 %d 不能与 HTTPS 监听地址 %s 重叠", options.HTTPChallengePort, address)
		}
	}
	options.LogLevel = strings.ToUpper(strings.TrimSpace(options.LogLevel))
	if options.LogLevel == "" {
		options.LogLevel = "INFO"
	}
	if !map[string]bool{"DEBUG": true, "INFO": true, "WARN": true, "ERROR": true, "PANIC": true, "FATAL": true}[options.LogLevel] {
		return fmt.Errorf("http 日志级别无效: %s", options.LogLevel)
	}
	for _, value := range options.TrustedProxies {
		if net.ParseIP(value) == nil {
			if _, _, err := net.ParseCIDR(value); err != nil {
				return fmt.Errorf("可信代理地址无效: %s", value)
			}
		}
	}
	if options.TrustedProxies != nil {
		options.TrustedProxies = append([]string{}, options.TrustedProxies...)
	}
	if options.ClientIPHeaders != nil {
		options.ClientIPHeaders = append([]string{}, options.ClientIPHeaders...)
	}

	var err error
	options.DataPath = strings.TrimSpace(options.DataPath)
	options.ACMEPath = strings.TrimSpace(options.ACMEPath)
	options.CachePath = strings.TrimSpace(options.CachePath)
	options.LogPath = strings.TrimSpace(options.LogPath)
	options.WAFLogPath = strings.TrimSpace(options.WAFLogPath)
	options.ConfigPath = strings.TrimSpace(options.ConfigPath)
	if options.DataPath == "" {
		options.DataPath = filepath.Join(m.root, "data")
	}
	if options.ACMEPath == "" {
		options.ACMEPath = filepath.Join(m.root, "acme")
	}
	if options.CachePath == "" {
		options.CachePath = filepath.Join(m.root, "cache")
	}
	if options.LogPath == "" {
		options.LogPath = filepath.Join(m.root, "logs", "http.log")
	}
	if options.WAFLogPath == "" {
		options.WAFLogPath = filepath.Join(m.root, "logs", "waf.log")
	}
	for target, destination := range map[*string]string{
		&options.DataPath:   options.DataPath,
		&options.ACMEPath:   options.ACMEPath,
		&options.CachePath:  options.CachePath,
		&options.WAFLogPath: options.WAFLogPath,
	} {
		*target, err = m.resolvePath(destination)
		if err != nil {
			return err
		}
	}
	if options.LogPath != "-" {
		options.LogPath, err = m.resolvePath(options.LogPath)
		if err != nil {
			return err
		}
	}
	if options.ConfigPath != "" {
		options.ConfigPath, err = m.resolvePath(options.ConfigPath)
		if err != nil {
			return err
		}
	}
	pathInside := func(parent, child string) bool {
		relative, relativeErr := filepath.Rel(parent, child)
		return relativeErr == nil && (relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))))
	}
	for _, directory := range []string{m.root, options.DataPath, options.ACMEPath, options.CachePath, filepath.Dir(options.WAFLogPath)} {
		if err = os.MkdirAll(directory, 0700); err != nil {
			return fmt.Errorf("创建 http 目录失败: %w", err)
		}
	}
	if options.LogPath != "-" {
		if err = os.MkdirAll(filepath.Dir(options.LogPath), 0700); err != nil {
			return fmt.Errorf("创建 http 日志目录失败: %w", err)
		}
	}
	if options.ConfigPath != "" {
		if err = os.MkdirAll(filepath.Dir(options.ConfigPath), 0700); err != nil {
			return fmt.Errorf("创建 http 配置目录失败: %w", err)
		}
	}
	for _, target := range []*string{&options.DataPath, &options.ACMEPath, &options.CachePath} {
		*target, err = filepath.EvalSymlinks(*target)
		if err != nil {
			return fmt.Errorf("解析 http 目录失败: %w", err)
		}
		*target = filepath.Clean(*target)
	}
	for _, target := range []*string{&options.WAFLogPath, &options.LogPath, &options.ConfigPath} {
		if *target == "" || *target == "-" {
			continue
		}
		parent, evalErr := filepath.EvalSymlinks(filepath.Dir(*target))
		if evalErr != nil {
			return fmt.Errorf("解析 http 文件目录失败: %w", evalErr)
		}
		*target = filepath.Join(parent, filepath.Base(*target))
	}
	files := map[string]string{}
	for name, path := range map[string]string{
		"http 日志": options.LogPath,
		"WAF 日志":  options.WAFLogPath,
		"配置文件":    options.ConfigPath,
	} {
		if path == "" || path == "-" {
			continue
		}
		if previous := files[path]; previous != "" {
			return fmt.Errorf("%s不能与%s使用同一个文件", name, previous)
		}
		files[path] = name
	}
	volumeRoot := filepath.Clean(filepath.VolumeName(options.CachePath) + string(filepath.Separator))
	if options.CachePath == volumeRoot || options.CachePath == m.root {
		return errors.New("CachePath 不能是磁盘根目录或 Manager 根目录")
	}
	acmeVolumeRoot := filepath.Clean(filepath.VolumeName(options.ACMEPath) + string(filepath.Separator))
	if options.ACMEPath == acmeVolumeRoot || pathInside(options.ACMEPath, m.root) {
		return errors.New("ACMEPath 不能是磁盘根目录、Manager 根目录或其父目录")
	}
	if err = os.Chmod(options.ACMEPath, 0700); err != nil {
		return fmt.Errorf("设置 ACME 目录权限失败: %w", err)
	}
	if pathInside(options.CachePath, m.root) || pathInside(options.CachePath, options.DataPath) || pathInside(options.DataPath, options.CachePath) ||
		pathInside(options.CachePath, options.ACMEPath) || pathInside(options.ACMEPath, options.CachePath) {
		return errors.New("CachePath 不能包含 Manager 根目录或与 DataPath、ACMEPath 重叠")
	}
	if pathInside(options.DataPath, options.ACMEPath) || pathInside(options.ACMEPath, options.DataPath) {
		return errors.New("DataPath 不能与 ACMEPath 重叠")
	}
	for _, file := range []string{options.LogPath, options.WAFLogPath, options.ConfigPath} {
		if file == "" || file == "-" {
			continue
		}
		if pathInside(options.CachePath, file) {
			return errors.New("日志或配置文件不能位于 CachePath 内")
		}
		if pathInside(options.DataPath, file) {
			return errors.New("日志或配置文件不能位于 DataPath 内")
		}
		if pathInside(options.ACMEPath, file) {
			return errors.New("日志或配置文件不能位于 ACMEPath 内")
		}
	}
	wafLog, err := os.OpenFile(options.WAFLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("创建 WAF 审计日志失败: %w", err)
	}
	if err = wafLog.Chmod(0o600); err != nil {
		_ = wafLog.Close()
		return fmt.Errorf("设置 WAF 审计日志权限失败: %w", err)
	}
	if err = wafLog.Close(); err != nil {
		return fmt.Errorf("关闭 WAF 审计日志失败: %w", err)
	}
	m.options = options
	m.dataPath = options.DataPath
	m.acmePath = options.ACMEPath
	m.cachePath = options.CachePath
	m.logPath = options.LogPath
	m.wafLogPath = options.WAFLogPath
	m.configPath = options.ConfigPath
	m.acmeStorage = &certmagic.FileStorage{Path: m.acmePath}
	return nil
}

func (m *Manager) updateOptions(options Options) error {
	m.mu.RLock()
	if m.configMode {
		m.mu.RUnlock()
		return errors.New("配置文件模式下无法修改 http 运行参数")
	}
	running := m.running
	currentConfig := append([]byte(nil), m.config...)
	storedSites := m.cloneSites(m.sites)
	if strings.TrimSpace(options.ACMEPath) == "" {
		options.ACMEPath = m.acmePath
	}
	m.mu.RUnlock()

	candidate := &Manager{root: m.root, sites: make(map[string]*Site)}
	if err := candidate.setOptions(options); err != nil {
		return err
	}
	input := make([]Site, 0, len(storedSites))
	for _, site := range storedSites {
		input = append(input, *site)
	}
	prepared, err := candidate.prepareSites(input)
	if err != nil {
		return err
	}
	config, err := candidate.buildConfig(prepared)
	if err != nil {
		return err
	}
	if err = candidate.ensureSiteLogDirectories(prepared); err != nil {
		return err
	}
	changed := !bytes.Equal(config, currentConfig)
	if !running {
		if err = candidate.validateConfig(config); err != nil {
			return err
		}
	}
	commit := func() {
		m.mu.Lock()
		m.options = candidate.options
		m.dataPath = candidate.dataPath
		m.acmePath = candidate.acmePath
		m.cachePath = candidate.cachePath
		m.logPath = candidate.logPath
		m.wafLogPath = candidate.wafLogPath
		m.configPath = candidate.configPath
		m.acmeStorage = candidate.acmeStorage
		m.sites = candidate.cloneSites(prepared)
		m.config = append([]byte(nil), config...)
		m.configMode = false
		m.mu.Unlock()
	}
	if running {
		httpRuntime.Lock()
		defer httpRuntime.Unlock()
		if httpRuntime.owner != m {
			return errors.New("当前 Manager 已失去 http 运行时所有权")
		}
		if changed {
			if err = caddy.Load(config, false); err != nil {
				return fmt.Errorf("加载 http 配置失败: %w", err)
			}
		}
	}
	if err = candidate.persistConfig(config); err != nil {
		if running && changed {
			if rollbackErr := caddy.Load(currentConfig, false); rollbackErr != nil {
				commit()
				return fmt.Errorf("新运行参数已生效，但保存和回滚均失败: %w", errors.Join(err, rollbackErr))
			}
			return fmt.Errorf("保存配置失败，http 热更新已回滚: %w", err)
		}
		return fmt.Errorf("保存 http 配置失败: %w", err)
	}
	commit()
	return nil
}

func (m *Manager) applySites(candidate map[string]*Site) error {
	input := make([]Site, 0, len(candidate))
	for _, site := range candidate {
		input = append(input, *site)
	}
	prepared, err := m.prepareSites(input)
	if err != nil {
		return err
	}
	candidate = prepared
	config, err := m.buildConfig(candidate)
	if err != nil {
		return err
	}
	if err = m.ensureSiteLogDirectories(candidate); err != nil {
		return err
	}
	m.mu.RLock()
	running := m.running
	currentConfig := append([]byte(nil), m.config...)
	m.mu.RUnlock()
	changed := !bytes.Equal(config, currentConfig)
	if changed {
		httpRuntime.Lock()
		defer httpRuntime.Unlock()
		if httpRuntime.owner != nil && httpRuntime.owner != m {
			return errors.New("http 运行时已由另一个 Manager 持有")
		}
		if running && httpRuntime.owner != m {
			return errors.New("当前 Manager 已失去 http 运行时所有权")
		}
	}
	if changed {
		if running {
			if err = caddy.Load(config, false); err != nil {
				return fmt.Errorf("加载 http 配置失败: %w", err)
			}
		} else if err = m.validateConfig(config); err != nil {
			return err
		}
	}
	if err = m.persistConfig(config); err != nil {
		if running && changed {
			if rollbackErr := caddy.Load(currentConfig, false); rollbackErr != nil {
				m.mu.Lock()
				m.sites = m.cloneSites(candidate)
				m.config = append([]byte(nil), config...)
				m.configMode = false
				m.mu.Unlock()
				return fmt.Errorf("新站点配置已生效，但保存和回滚均失败: %w", errors.Join(err, rollbackErr))
			}
			return fmt.Errorf("保存配置失败，站点热更新已回滚: %w", err)
		}
		return fmt.Errorf("保存 http 配置失败: %w", err)
	}
	m.mu.Lock()
	m.sites = m.cloneSites(candidate)
	m.config = append([]byte(nil), config...)
	m.configMode = false
	m.mu.Unlock()
	return nil
}

func (m *Manager) applyConfig(config []byte) error {
	m.mu.RLock()
	running := m.running
	currentConfig := append([]byte(nil), m.config...)
	m.mu.RUnlock()
	httpRuntime.Lock()
	defer httpRuntime.Unlock()
	if httpRuntime.owner != nil && httpRuntime.owner != m {
		return errors.New("http 运行时已由另一个 Manager 持有")
	}
	if running && httpRuntime.owner != m {
		return errors.New("当前 Manager 已失去 http 运行时所有权")
	}
	if err := m.validateConfig(config); err != nil {
		return err
	}
	changed := !bytes.Equal(config, currentConfig)
	if running && changed {
		if err := caddy.Load(config, false); err != nil {
			return fmt.Errorf("加载 http 配置失败: %w", err)
		}
	}
	if err := m.persistConfig(config); err != nil {
		if running && changed {
			if rollbackErr := caddy.Load(currentConfig, false); rollbackErr != nil {
				m.mu.Lock()
				m.sites = make(map[string]*Site)
				m.config = append([]byte(nil), config...)
				m.configMode = true
				m.mu.Unlock()
				return fmt.Errorf("新配置已生效，但保存和回滚均失败: %w", errors.Join(err, rollbackErr))
			}
			return fmt.Errorf("保存配置失败，http 热更新已回滚: %w", err)
		}
		return fmt.Errorf("保存 http 配置失败: %w", err)
	}
	m.mu.Lock()
	m.sites = make(map[string]*Site)
	m.config = append([]byte(nil), config...)
	m.configMode = true
	m.mu.Unlock()
	return nil
}

func (m *Manager) buildConfig(sites map[string]*Site) ([]byte, error) {
	if err := m.validateSiteSet(sites); err != nil {
		return nil, err
	}
	httpsPort := 0
	httpsPorts := make(map[uint]struct{})
	ambiguousHTTPSPort := false
	for _, address := range m.options.HTTPSListen {
		parsed, _ := caddy.ParseNetworkAddressWithDefaults(address, "tcp", 443)
		if parsed.StartPort <= 443 && parsed.EndPort >= 443 {
			httpsPort = 443
			break
		}
		if parsed.StartPort == 0 || parsed.StartPort != parsed.EndPort {
			ambiguousHTTPSPort = true
			continue
		}
		httpsPorts[parsed.StartPort] = struct{}{}
	}
	if httpsPort == 0 && !ambiguousHTTPSPort && len(httpsPorts) == 1 {
		for port := range httpsPorts {
			httpsPort = int(port)
		}
	}
	all := make([]*Site, 0, len(sites))
	enabled := make([]*Site, 0, len(sites))
	disabled := make([]*Site, 0, len(sites))
	for _, site := range sites {
		all = append(all, site)
		if site.Enabled {
			enabled = append(enabled, site)
		} else {
			disabled = append(disabled, site)
		}
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	sort.SliceStable(enabled, func(i, j int) bool { return enabled[i].ID < enabled[j].ID })
	sort.SliceStable(disabled, func(i, j int) bool { return disabled[i].ID < disabled[j].ID })
	logEncoder := map[string]interface{}{
		"format":       "console",
		"time_format":  "2006-01-02 15:04:05.000",
		"time_local":   true,
		"level_format": "upper",
	}
	accessLoggerNames := make(map[string][]string)
	accessLoggerBySite := make(map[string]string, len(all))
	accessLogs := make(map[string]interface{}, len(all))
	wafLogNamespace := "http.handlers." + string(HandlerWAF)
	excludedLogNamespaces := []string{wafLogNamespace}
	for _, site := range all {
		fingerprint := sha256.Sum256([]byte(site.ID))
		loggerName := fmt.Sprintf("site_%x", fingerprint)
		accessLoggerBySite[site.ID] = loggerName
		logPath, err := m.siteLogPath(site.ID, LogAccess)
		if err != nil {
			return nil, err
		}
		for _, domain := range site.Domains {
			accessLoggerNames[domain] = []string{loggerName}
		}
		namespace := "http.log.access." + loggerName
		excludedLogNamespaces = append(excludedLogNamespaces, namespace)
		accessLogs[loggerName] = map[string]interface{}{
			"include": []string{namespace},
			"encoder": logEncoder,
			"writer": map[string]interface{}{
				"output":           "file",
				"filename":         logPath,
				"mode":             "0600",
				"dir_mode":         "0700",
				"roll_size_mb":     50,
				"roll_keep":        10,
				"roll_keep_days":   30,
				"roll_compression": "gzip",
			},
		}
	}
	exactDomains := make(map[string]string)
	for _, site := range all {
		for _, domain := range site.Domains {
			if !strings.HasPrefix(domain, "*.") {
				exactDomains[domain] = site.ID
			}
		}
	}

	httpRoutes := make([]interface{}, 0, len(all)+1)
	httpsRoutes := make([]interface{}, 0, len(all)+1)
	exactTLSPolicies := make([]interface{}, 0)
	wildcardTLSPolicies := make([]interface{}, 0)
	ipTLSPolicies := make([]map[string]interface{}, 0)
	type certificateFileConfig struct {
		Certificate string   `json:"certificate"`
		Key         string   `json:"key"`
		Tags        []string `json:"tags"`
	}
	type certificatePEMConfig struct {
		Certificate string   `json:"certificate"`
		Key         string   `json:"key"`
		Tags        []string `json:"tags"`
	}
	certificateFiles := make(map[string]*certificateFileConfig)
	certificatePEMs := make(map[string]*certificatePEMConfig)
	excludedFor := func(site *Site) []string {
		excludedDomains := make([]string, 0)
		for domain, owner := range exactDomains {
			if owner == site.ID {
				continue
			}
			for _, pattern := range site.Domains {
				if strings.HasPrefix(pattern, "*.") && strings.Count(pattern, ".") == strings.Count(domain, ".") &&
					strings.HasSuffix(domain, strings.TrimPrefix(pattern, "*")) {
					excludedDomains = append(excludedDomains, domain)
					break
				}
			}
		}
		sort.Strings(excludedDomains)
		return excludedDomains
	}
	registerTLS := func(site *Site, tlsDomains []string) error {
		for _, pair := range site.TLS.Certificates {
			certificateIdentity := pair.CertificatePEM + "\x00" + pair.PrivateKeyPEM
			if pair.CertificatePEM == "" {
				certificateContent, err := os.ReadFile(pair.CertificateFile)
				if err != nil {
					return fmt.Errorf("读取站点 %s 证书失败: %w", site.ID, err)
				}
				keyContent, err := os.ReadFile(pair.KeyFile)
				if err != nil {
					return fmt.Errorf("读取站点 %s 私钥失败: %w", site.ID, err)
				}
				certificateIdentity = string(certificateContent) + "\x00" + string(keyContent)
			}
			fingerprint := sha256.Sum256([]byte(certificateIdentity))
			tag := fmt.Sprintf("site:%s:certificate:%x", site.ID, fingerprint)
			pairDomains := pair.Domains
			if len(pairDomains) == 0 {
				pairDomains = tlsDomains
			}
			exact, wildcard := make([]string, 0, len(pairDomains)), make([]string, 0, len(pairDomains))
			ipDomains := make([]string, 0)
			for _, domain := range pairDomains {
				if net.ParseIP(domain) != nil {
					ipDomains = append(ipDomains, domain)
				}
				if strings.HasPrefix(domain, "*.") {
					wildcard = append(wildcard, domain)
				} else {
					exact = append(exact, domain)
				}
			}
			for policyIndex, domains := range [][]string{exact, wildcard, ipDomains} {
				if len(domains) == 0 {
					continue
				}
				policy := map[string]interface{}{
					"match":                 map[string]interface{}{"sni": domains},
					"certificate_selection": map[string]interface{}{"any_tag": []string{tag}},
				}
				if site.TLS.MinVersion != "" {
					policy["protocol_min"] = site.TLS.MinVersion
				}
				if site.TLS.MaxVersion != "" {
					policy["protocol_max"] = site.TLS.MaxVersion
				}
				if policyIndex == 2 {
					// IP 访问通常不发送 SNI，通过接收连接的本地地址选择绑定证书。
					policy["default_sni"] = domains[0]
					policy["match"] = map[string]interface{}{
						"sni":      []string{""},
						"local_ip": map[string]interface{}{"ranges": domains},
					}
					ipTLSPolicies = append(ipTLSPolicies, policy)
				} else if policyIndex == 0 {
					exactTLSPolicies = append(exactTLSPolicies, policy)
				} else {
					wildcardTLSPolicies = append(wildcardTLSPolicies, policy)
				}
			}
			if pair.CertificatePEM != "" {
				key := string(fingerprint[:])
				certificate := certificatePEMs[key]
				if certificate == nil {
					certificate = &certificatePEMConfig{Certificate: pair.CertificatePEM, Key: pair.PrivateKeyPEM}
					certificatePEMs[key] = certificate
				}
				certificate.Tags = append(certificate.Tags, tag)
				continue
			}
			key := pair.CertificateFile + "\x00" + pair.KeyFile
			certificate := certificateFiles[key]
			if certificate == nil {
				certificate = &certificateFileConfig{Certificate: pair.CertificateFile, Key: pair.KeyFile}
				certificateFiles[key] = certificate
			}
			certificate.Tags = append(certificate.Tags, tag)
		}
		return nil
	}
	for _, site := range enabled {
		excludedDomains := excludedFor(site)
		if site.TLS.Enabled {
			tlsDomains := append([]string(nil), site.TLS.coveredDomains...)
			covered := make(map[string]struct{}, len(tlsDomains))
			for _, domain := range tlsDomains {
				covered[domain] = struct{}{}
			}
			plainDomains := make([]string, 0, len(site.Domains)-len(tlsDomains))
			for _, domain := range site.Domains {
				if _, exists := covered[domain]; !exists {
					plainDomains = append(plainDomains, domain)
				}
			}
			httpsRoute, err := m.buildSiteRoute(site, tlsDomains, excludedDomains, "https")
			if err != nil {
				return nil, err
			}
			httpsRoutes = append(httpsRoutes, httpsRoute)
			if err := registerTLS(site, tlsDomains); err != nil {
				return nil, err
			}
			if site.TLS.RedirectHTTP {
				if len(m.options.HTTPSListen) == 0 {
					return nil, fmt.Errorf("站点 %s 开启了 HTTP 跳转，但没有配置 HTTPS 监听地址", site.ID)
				}
				if httpsPort == 0 {
					return nil, errors.New("存在多个 HTTPS 监听端口，无法确定 HTTP 跳转目标")
				}
				type redirectTarget struct {
					domains []string
					host    string
				}
				redirectTargets := []redirectTarget{{host: "{http.request.host}"}}
				for _, domain := range tlsDomains {
					if address := net.ParseIP(domain); address != nil && address.To4() == nil {
						// Host 带端口时占位符会去掉 IPv6 方括号，使用规范地址生成合法 URL。
						redirectTargets = append(redirectTargets, redirectTarget{domains: []string{domain}, host: "[" + domain + "]"})
					} else {
						redirectTargets[0].domains = append(redirectTargets[0].domains, domain)
					}
				}
				for _, target := range redirectTargets {
					if len(target.domains) == 0 {
						continue
					}
					location := "https://" + target.host
					if httpsPort != 443 {
						location += ":" + strconv.Itoa(httpsPort)
					}
					location += "{http.request.uri}"
					redirectMatcher := map[string]interface{}{"host": target.domains}
					if len(excludedDomains) > 0 {
						redirectMatcher["not"] = []interface{}{map[string]interface{}{"host": excludedDomains}}
					}
					redirectHandlers := make([]interface{}, 0, 2)
					if handler := m.buildTrafficLimitHandler(site); handler != nil {
						redirectHandlers = append(redirectHandlers, handler)
					}
					redirectHandlers = append(redirectHandlers, map[string]interface{}{
						"handler": HandlerStaticResponse, "status_code": 308,
						"headers": map[string][]string{"Location": {location}},
					})
					httpRoutes = append(httpRoutes, map[string]interface{}{
						"match":    []interface{}{redirectMatcher},
						"handle":   redirectHandlers,
						"terminal": true,
					})
				}
				if len(plainDomains) > 0 {
					httpRoute, err := m.buildSiteRoute(site, plainDomains, excludedDomains, "http")
					if err != nil {
						return nil, err
					}
					httpRoutes = append(httpRoutes, httpRoute)
				}
			} else {
				httpRoute, err := m.buildSiteRoute(site, site.Domains, excludedDomains, "http")
				if err != nil {
					return nil, err
				}
				httpRoutes = append(httpRoutes, httpRoute)
			}
		} else {
			httpRoute, err := m.buildSiteRoute(site, site.Domains, excludedDomains, "http")
			if err != nil {
				return nil, err
			}
			httpRoutes = append(httpRoutes, httpRoute)
		}
	}
	for _, site := range disabled {
		excludedDomains := excludedFor(site)
		httpRoutes = append(httpRoutes, m.buildResponseRoute(m.options.SiteDisabledPage, site.Domains, excludedDomains))
		if site.TLS.Enabled && len(site.TLS.coveredDomains) > 0 {
			tlsDomains := append([]string(nil), site.TLS.coveredDomains...)
			httpsRoutes = append(httpsRoutes, m.buildResponseRoute(m.options.SiteDisabledPage, tlsDomains, excludedDomains))
			if err := registerTLS(site, tlsDomains); err != nil {
				return nil, err
			}
		}
	}
	var defaultSite *Site
	if m.options.DefaultSite != "" {
		defaultSite = sites[m.options.DefaultSite]
		if defaultSite == nil && len(sites) > 0 {
			return nil, fmt.Errorf("默认站点不存在: %s", m.options.DefaultSite)
		}
		if defaultSite != nil && !defaultSite.Enabled {
			defaultSite = nil
		}
	}
	if defaultSite != nil {
		if len(m.options.HTTPListen) == 0 {
			return nil, errors.New("配置默认站点时必须启用 HTTP 监听")
		}
		defaultRoute, err := m.buildSiteRoute(defaultSite, nil, nil, "http-default")
		if err != nil {
			return nil, fmt.Errorf("生成默认站点路由失败: %w", err)
		}
		httpRoutes = append(httpRoutes, defaultRoute)
	} else {
		httpRoutes = append(httpRoutes, m.buildResponseRoute(m.options.SiteNotFoundPage, nil, nil))
	}
	if len(httpsRoutes) > 0 || len(m.options.HTTPSListen) > 0 {
		httpsRoutes = append(httpsRoutes, m.buildResponseRoute(m.options.SiteNotFoundPage, nil, nil))
	}
	tlsPolicies := append(exactTLSPolicies, wildcardTLSPolicies...)
	for _, policy := range ipTLSPolicies {
		if len(ipTLSPolicies) == 1 {
			// 只有一组 IP 证书时也支持公网地址映射到内网地址的 NAT 部署。
			delete(policy["match"].(map[string]interface{}), "local_ip")
		}
		tlsPolicies = append(tlsPolicies, policy)
	}
	if len(httpsRoutes) > 0 && len(tlsPolicies) == 0 {
		tlsPolicies = []interface{}{map[string]interface{}{}}
	}

	httpApp := map[string]interface{}{
		"http_port":    m.options.HTTPChallengePort,
		"grace_period": m.options.GracePeriod.String(),
		"servers":      map[string]interface{}{},
	}
	servers := httpApp["servers"].(map[string]interface{})
	if len(httpRoutes) > 0 {
		if len(m.options.HTTPListen) == 0 {
			for _, site := range enabled {
				if !site.TLS.Enabled || site.TLS.RedirectHTTP || len(site.TLS.coveredDomains) < len(site.Domains) {
					return nil, fmt.Errorf("站点 %s 需要 HTTP，但没有配置 HTTP 监听地址", site.ID)
				}
			}
		} else {
			hasPlainHTTP := len(m.options.Protocols) == 0
			for _, protocol := range m.options.Protocols {
				if protocol == "h1" || protocol == "h2c" {
					hasPlainHTTP = true
					break
				}
			}
			if !hasPlainHTTP {
				return nil, errors.New("HTTP 监听至少需要启用 h1 或 h2c")
			}
			defaultLoggerName := ""
			if defaultSite != nil {
				defaultLoggerName = accessLoggerBySite[defaultSite.ID]
			}
			servers["http"] = m.buildHTTPServer(m.options.HTTPListen, httpRoutes, nil, accessLoggerNames, defaultLoggerName, enabled)
		}
	}
	if len(httpsRoutes) > 0 {
		if len(m.options.HTTPSListen) == 0 {
			return nil, errors.New("存在 TLS 站点，但没有配置 HTTPS 监听地址")
		}
		servers["https"] = m.buildHTTPServer(m.options.HTTPSListen, httpsRoutes, tlsPolicies, accessLoggerNames, "", enabled)
	}

	loadFiles := make([]certificateFileConfig, 0, len(certificateFiles))
	certificateKeys := make([]string, 0, len(certificateFiles))
	for key := range certificateFiles {
		certificateKeys = append(certificateKeys, key)
	}
	sort.Strings(certificateKeys)
	for _, key := range certificateKeys {
		certificate := certificateFiles[key]
		sort.Strings(certificate.Tags)
		loadFiles = append(loadFiles, *certificate)
	}
	loadPEMs := make([]certificatePEMConfig, 0, len(certificatePEMs))
	certificateKeys = certificateKeys[:0]
	for key := range certificatePEMs {
		certificateKeys = append(certificateKeys, key)
	}
	sort.Strings(certificateKeys)
	for _, key := range certificateKeys {
		certificate := certificatePEMs[key]
		sort.Strings(certificate.Tags)
		loadPEMs = append(loadPEMs, *certificate)
	}
	tlsApp := map[string]interface{}{}
	certificateLoaders := map[string]interface{}{}
	if len(loadFiles) != 0 {
		certificateLoaders["load_files"] = loadFiles
	}
	if len(loadPEMs) != 0 {
		certificateLoaders["load_pem"] = loadPEMs
	}
	if len(certificateLoaders) != 0 {
		tlsApp["certificates"] = certificateLoaders
	}
	apps := map[string]interface{}{
		"http":   httpApp,
		"tls":    tlsApp,
		"events": map[string]interface{}{},
		"cache":  map[string]interface{}{},
	}
	persist := false
	config := map[string]interface{}{
		"admin": map[string]interface{}{
			"disabled": true,
			"config":   map[string]interface{}{"persist": persist},
		},
		"storage": map[string]interface{}{"module": "file_system", "root": m.dataPath},
		"apps":    apps,
	}
	logConfig := map[string]interface{}{"level": m.options.LogLevel, "encoder": logEncoder}
	logConfig["exclude"] = excludedLogNamespaces
	if m.logPath != "-" {
		logConfig["writer"] = map[string]interface{}{
			"output":           "file",
			"filename":         m.logPath,
			"mode":             "0600",
			"dir_mode":         "0700",
			"roll_size_mb":     50,
			"roll_keep":        10,
			"roll_keep_days":   30,
			"roll_compression": "gzip",
		}
	} else {
		logConfig["writer"] = map[string]interface{}{"output": "discard"}
	}
	logs := map[string]interface{}{
		"default": logConfig,
		"waf": map[string]interface{}{
			"include": []string{wafLogNamespace},
			"level":   m.options.LogLevel,
			"encoder": logEncoder,
			"writer": map[string]interface{}{
				"output":           "file",
				"filename":         m.wafLogPath,
				"mode":             "0600",
				"dir_mode":         "0700",
				"roll_size_mb":     50,
				"roll_keep":        10,
				"roll_keep_days":   30,
				"roll_compression": "gzip",
			},
		},
	}
	for name, accessLog := range accessLogs {
		logs[name] = accessLog
	}
	config["logging"] = map[string]interface{}{"logs": logs}
	result, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("生成 http 配置失败: %w", err)
	}
	return result, nil
}

func (m *Manager) buildHTTPServer(listen []string, routes []interface{}, tlsPolicies []interface{}, loggerNames map[string][]string, defaultLoggerName string, sites []*Site) map[string]interface{} {
	routes = append([]interface{}{
		map[string]interface{}{
			"handle": []interface{}{map[string]interface{}{
				"handler": HandlerHeaders,
				"response": map[string]interface{}{
					"delete":   []string{"Server"},
					"deferred": true,
				},
			}},
		},
	}, routes...)
	logOptions := map[string]interface{}{
		"logger_names":        loggerNames,
		"skip_unmapped_hosts": defaultLoggerName == "",
	}
	if defaultLoggerName != "" {
		logOptions["default_logger_name"] = defaultLoggerName
	}
	errorRoutes := make([]interface{}, 0, len(sites)+1)
	for _, site := range sites {
		if !site.WAF.Enabled {
			continue
		}
		body := site.WAF.BlockPage
		if strings.TrimSpace(body) == "" {
			body = DefaultWAFBlockPage
		}
		response := m.buildResponseHandler(ResponseOptions{
			Body: body,
			Headers: map[string][]string{
				"Content-Type":  {"text/html; charset=utf-8"},
				"Cache-Control": {"no-store"},
			},
		})
		response["status_code"] = "{http.error.status_code}"
		errorRoutes = append(errorRoutes, map[string]interface{}{
			"match": []interface{}{map[string]interface{}{
				"vars": map[string][]string{"waf_site_id": {site.ID}},
				// 官方 WAF 的拦截错误携带事务 ID，避免覆盖网站自身的错误响应。
				"expression": "{http.error.message} == 'interruption triggered' && {http.error.id} != '' && {http.error.id} == {http.transaction_id}",
			}},
			"handle":   []interface{}{response},
			"terminal": true,
		})
	}
	errorRoutes = append(errorRoutes, map[string]interface{}{
		"match": []interface{}{map[string]interface{}{
			"expression": "{http.error.status_code} == 404",
		}},
		"handle":   []interface{}{m.buildResponseHandler(m.options.NotFoundPage)},
		"terminal": true,
	})
	server := map[string]interface{}{
		"listen":              append([]string(nil), listen...),
		"routes":              routes,
		"errors":              map[string]interface{}{"routes": errorRoutes},
		"automatic_https":     map[string]interface{}{"disable": true},
		"protocols":           append([]string(nil), m.options.Protocols...),
		"read_header_timeout": m.options.ReadHeaderTimeout.String(),
		"idle_timeout":        m.options.IdleTimeout.String(),
		"max_header_bytes":    m.options.MaxHeaderBytes,
		"logs":                logOptions,
	}
	if m.options.ReadTimeout > 0 {
		server["read_timeout"] = m.options.ReadTimeout.String()
	}
	if m.options.WriteTimeout > 0 {
		server["write_timeout"] = m.options.WriteTimeout.String()
	}
	if len(tlsPolicies) > 0 {
		server["tls_connection_policies"] = tlsPolicies
	}
	if len(m.options.TrustedProxies) > 0 {
		server["trusted_proxies"] = map[string]interface{}{"source": "static", "ranges": m.options.TrustedProxies}
		if len(m.options.ClientIPHeaders) > 0 {
			server["client_ip_headers"] = m.options.ClientIPHeaders
		}
		if m.options.TrustedProxiesStrict {
			server["trusted_proxies_strict"] = 1
		}
	}
	return server
}

func (m *Manager) validateConfig(config []byte) error {
	if len(bytes.TrimSpace(config)) == 0 {
		return errors.New("http 配置不能为空")
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(config, &document); err != nil {
		return fmt.Errorf("解析 http 配置失败: %w", err)
	}
	var admin struct {
		Disabled bool `json:"disabled"`
	}
	if err := json.Unmarshal(document["admin"], &admin); err != nil || !admin.Disabled {
		return errors.New("http 配置必须关闭管理接口")
	}
	var parsed caddy.Config
	if err := caddy.StrictUnmarshalJSON(config, &parsed); err != nil {
		return fmt.Errorf("解析 http 配置失败: %w", err)
	}
	if err := caddy.Validate(&parsed); err != nil {
		return fmt.Errorf("验证 http 配置失败: %w", err)
	}
	return nil
}

func (m *Manager) persistConfig(config []byte) error {
	if m.configPath == "" {
		return nil
	}
	temporary, err := os.CreateTemp(filepath.Dir(m.configPath), ".http-*.json")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err = temporary.Chmod(0600); err == nil {
		_, err = temporary.Write(config)
	}
	if err == nil {
		err = temporary.Sync()
	}
	closeErr := temporary.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(temporaryPath, m.configPath)
}
