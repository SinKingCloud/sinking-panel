package site

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"server/app/constant"
	"server/app/enum/site_status"
	webServer "server/app/util/server"
)

// GetHTTP 返回当前实际生效的 HTTP 服务全局参数。
func (s *service) GetHTTP() webServer.Options {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	return s.http.Options()
}

// UpdateHTTP 合并并应用 HTTP 服务全局参数，失败时恢复原运行参数和进程配置。
func (s *service) UpdateHTTP(config *HTTPUpdate) error {
	if config == nil {
		return errors.New("HTTP 配置不能为空")
	}
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	return s.updateHTTPLocked(config)
}

func (s *service) updateHTTPLocked(config *HTTPUpdate) error {
	previous := s.http.Options()
	candidate := previous
	var err error
	if config.HTTPListen != nil {
		candidate.HTTPListen = slices.Clone(*config.HTTPListen)
	}
	if config.HTTPSListen != nil {
		candidate.HTTPSListen = slices.Clone(*config.HTTPSListen)
	}
	if config.Protocols != nil {
		candidate.Protocols = slices.Clone(*config.Protocols)
	}
	if config.DefaultSite != nil {
		candidate.DefaultSite = strings.TrimSpace(*config.DefaultSite)
		if candidate.DefaultSite != "" {
			id, parseErr := strconv.ParseInt(candidate.DefaultSite, 10, 64)
			if parseErr != nil || id <= 0 {
				return errors.New("默认站点 ID 不合法")
			}
			record, findErr := s.repositorySite.FindById(id)
			if findErr != nil {
				return fmt.Errorf("默认站点不存在: %w", findErr)
			}
			if record.Status != site_status.Enabled {
				return errors.New("默认站点必须处于启用状态")
			}
			domains, queryErr := s.repositorySiteDomain.SelectBySiteId(id)
			if queryErr != nil {
				return fmt.Errorf("查询默认站点域名失败: %w", queryErr)
			}
			if len(domains) == 0 {
				return errors.New("默认站点至少需要绑定一个域名")
			}
		}
	}
	for _, page := range []struct {
		target *webServer.ResponseOptions
		update *HTTPPageUpdate
	}{
		{&candidate.NotFoundPage, config.NotFoundPage},
		{&candidate.SiteNotFoundPage, config.SiteNotFoundPage},
		{&candidate.SiteDisabledPage, config.SiteDisabledPage},
	} {
		if page.update == nil {
			continue
		}
		if page.update.Status != nil {
			page.target.Status = *page.update.Status
		}
		if page.update.Body != nil {
			page.target.Body = *page.update.Body
		}
		if page.update.Headers != nil {
			page.target.Headers = make(map[string][]string, len(*page.update.Headers))
			for name, values := range *page.update.Headers {
				page.target.Headers[name] = slices.Clone(values)
			}
		}
	}
	if config.DataPath != nil {
		candidate.DataPath = strings.TrimSpace(*config.DataPath)
	}
	if config.CachePath != nil {
		candidate.CachePath = strings.TrimSpace(*config.CachePath)
	}
	if config.LogPath != nil {
		candidate.LogPath = strings.TrimSpace(*config.LogPath)
	}
	if config.WAFLogPath != nil {
		candidate.WAFLogPath = strings.TrimSpace(*config.WAFLogPath)
	}
	if config.ConfigPath != nil {
		candidate.ConfigPath = strings.TrimSpace(*config.ConfigPath)
	}
	if config.LogLevel != nil {
		candidate.LogLevel = *config.LogLevel
	}
	if config.TrustedProxies != nil {
		candidate.TrustedProxies = slices.Clone(*config.TrustedProxies)
	}
	if config.ClientIPHeaders != nil {
		candidate.ClientIPHeaders = slices.Clone(*config.ClientIPHeaders)
	}
	if config.TrustedProxiesStrict != nil {
		candidate.TrustedProxiesStrict = *config.TrustedProxiesStrict
	}
	if config.ReadTimeout != nil {
		candidate.ReadTimeout = *config.ReadTimeout
	}
	if config.ReadHeaderTimeout != nil {
		candidate.ReadHeaderTimeout = *config.ReadHeaderTimeout
	}
	if config.WriteTimeout != nil {
		candidate.WriteTimeout = *config.WriteTimeout
	}
	if config.IdleTimeout != nil {
		candidate.IdleTimeout = *config.IdleTimeout
	}
	if config.GracePeriod != nil {
		candidate.GracePeriod = *config.GracePeriod
	}
	if config.MaxHeaderBytes != nil {
		candidate.MaxHeaderBytes = *config.MaxHeaderBytes
	}
	if config.HTTPChallengeHost != nil {
		candidate.HTTPChallengeHost = *config.HTTPChallengeHost
	}
	if config.HTTPChallengePort != nil {
		candidate.HTTPChallengePort = *config.HTTPChallengePort
	}
	if config.ACMEEmail != nil {
		candidate.ACMEEmail = *config.ACMEEmail
	}
	configs := make(map[string]string)
	for _, item := range []struct {
		key     string
		changed bool
		value   interface{}
	}{
		{constant.SiteHTTPListen, config.HTTPListen != nil, candidate.HTTPListen},
		{constant.SiteHTTPSListen, config.HTTPSListen != nil, candidate.HTTPSListen},
		{constant.SiteHTTPProtocols, config.Protocols != nil, candidate.Protocols},
		{constant.SiteHTTPDefaultSite, config.DefaultSite != nil, candidate.DefaultSite},
		{constant.SiteHTTPNotFoundPage, config.NotFoundPage != nil, candidate.NotFoundPage},
		{constant.SiteHTTPSiteNotFoundPage, config.SiteNotFoundPage != nil, candidate.SiteNotFoundPage},
		{constant.SiteHTTPSiteDisabledPage, config.SiteDisabledPage != nil, candidate.SiteDisabledPage},
		{constant.SiteHTTPDataPath, config.DataPath != nil, candidate.DataPath},
		{constant.SiteHTTPCachePath, config.CachePath != nil, candidate.CachePath},
		{constant.SiteHTTPLogPath, config.LogPath != nil, candidate.LogPath},
		{constant.SiteHTTPWAFLogPath, config.WAFLogPath != nil, candidate.WAFLogPath},
		{constant.SiteHTTPConfigPath, config.ConfigPath != nil, candidate.ConfigPath},
		{constant.SiteHTTPLogLevel, config.LogLevel != nil, candidate.LogLevel},
		{constant.SiteHTTPTrustedProxies, config.TrustedProxies != nil, candidate.TrustedProxies},
		{constant.SiteHTTPClientIPHeaders, config.ClientIPHeaders != nil, candidate.ClientIPHeaders},
		{constant.SiteHTTPTrustedProxiesStrict, config.TrustedProxiesStrict != nil, candidate.TrustedProxiesStrict},
		{constant.SiteHTTPReadTimeout, config.ReadTimeout != nil, candidate.ReadTimeout},
		{constant.SiteHTTPReadHeaderTimeout, config.ReadHeaderTimeout != nil, candidate.ReadHeaderTimeout},
		{constant.SiteHTTPWriteTimeout, config.WriteTimeout != nil, candidate.WriteTimeout},
		{constant.SiteHTTPIdleTimeout, config.IdleTimeout != nil, candidate.IdleTimeout},
		{constant.SiteHTTPGracePeriod, config.GracePeriod != nil, candidate.GracePeriod},
		{constant.SiteHTTPMaxHeaderBytes, config.MaxHeaderBytes != nil, candidate.MaxHeaderBytes},
		{constant.SiteHTTPChallengeHost, config.HTTPChallengeHost != nil, candidate.HTTPChallengeHost},
		{constant.SiteHTTPChallengePort, config.HTTPChallengePort != nil, candidate.HTTPChallengePort},
		{constant.SiteHTTPACMEEmail, config.ACMEEmail != nil, candidate.ACMEEmail},
	} {
		if !item.changed {
			continue
		}
		content, err := json.Marshal(item.value)
		if err != nil {
			return fmt.Errorf("序列化 HTTP 配置 %s 失败: %w", item.key, err)
		}
		configs[item.key] = string(content)
	}
	if len(configs) == 0 {
		return nil
	}
	effective := candidate
	for _, item := range []struct {
		target *string
		path   string
	}{
		{&effective.DataPath, constant.ServerDataPath},
		{&effective.ACMEPath, constant.AcmePath},
		{&effective.CachePath, constant.ServerCachePath},
		{&effective.LogPath, constant.ServerLogPath},
		{&effective.WAFLogPath, constant.ServerWAFLogPath},
		{&effective.ConfigPath, constant.ServerConfigPath},
	} {
		if *item.target != "" {
			continue
		}
		*item.target, err = filepath.Abs(item.path)
		if err != nil {
			return fmt.Errorf("解析 HTTP 服务默认路径失败: %w", err)
		}
		*item.target = filepath.Clean(*item.target)
	}
	if err = s.http.UpdateOptions(effective); err != nil {
		return err
	}
	rollback := func(cause error) error {
		optionsErr := s.http.UpdateOptions(previous)
		var syncErr error
		if optionsErr == nil {
			syncErr = s.syncLocked()
		}
		if optionsErr == nil && syncErr == nil {
			return fmt.Errorf("%w，运行参数已恢复", cause)
		}
		if optionsErr != nil {
			optionsErr = fmt.Errorf("恢复原 HTTP 配置失败: %w", optionsErr)
		}
		if syncErr != nil {
			syncErr = fmt.Errorf("恢复原网站运行时失败: %w", syncErr)
		}
		return errors.Join(cause, optionsErr, syncErr)
	}
	if err = s.syncLocked(); err != nil {
		return rollback(fmt.Errorf("同步网站运行时失败: %w", err))
	}
	err = s.configService.Sets(configs)
	if err == nil {
		return nil
	}
	return rollback(fmt.Errorf("保存 HTTP 配置失败: %w", err))
}

// loadHTTP 读取并合并已保存的 HTTP 服务全局参数。
func (s *service) loadHTTP() (webServer.Options, bool, error) {
	values := s.configService.Group(constant.SiteGroup)
	config := webServer.Options{}
	exists := false
	fields := []struct {
		key    string
		target interface{}
	}{
		{constant.SiteHTTPListen, &config.HTTPListen},
		{constant.SiteHTTPSListen, &config.HTTPSListen},
		{constant.SiteHTTPProtocols, &config.Protocols},
		{constant.SiteHTTPDefaultSite, &config.DefaultSite},
		{constant.SiteHTTPNotFoundPage, &config.NotFoundPage},
		{constant.SiteHTTPSiteNotFoundPage, &config.SiteNotFoundPage},
		{constant.SiteHTTPSiteDisabledPage, &config.SiteDisabledPage},
		{constant.SiteHTTPDataPath, &config.DataPath},
		{constant.SiteHTTPCachePath, &config.CachePath},
		{constant.SiteHTTPLogPath, &config.LogPath},
		{constant.SiteHTTPWAFLogPath, &config.WAFLogPath},
		{constant.SiteHTTPConfigPath, &config.ConfigPath},
		{constant.SiteHTTPLogLevel, &config.LogLevel},
		{constant.SiteHTTPTrustedProxies, &config.TrustedProxies},
		{constant.SiteHTTPClientIPHeaders, &config.ClientIPHeaders},
		{constant.SiteHTTPTrustedProxiesStrict, &config.TrustedProxiesStrict},
		{constant.SiteHTTPReadTimeout, &config.ReadTimeout},
		{constant.SiteHTTPReadHeaderTimeout, &config.ReadHeaderTimeout},
		{constant.SiteHTTPWriteTimeout, &config.WriteTimeout},
		{constant.SiteHTTPIdleTimeout, &config.IdleTimeout},
		{constant.SiteHTTPGracePeriod, &config.GracePeriod},
		{constant.SiteHTTPMaxHeaderBytes, &config.MaxHeaderBytes},
		{constant.SiteHTTPChallengeHost, &config.HTTPChallengeHost},
		{constant.SiteHTTPChallengePort, &config.HTTPChallengePort},
		{constant.SiteHTTPACMEEmail, &config.ACMEEmail},
	}
	for _, item := range fields {
		raw, ok := values[item.key]
		if !ok {
			continue
		}
		decoder := json.NewDecoder(strings.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(item.target); err != nil {
			return webServer.Options{}, false, fmt.Errorf("HTTP 配置 %s 格式错误: %w", item.key, err)
		}
		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
			if err == nil {
				return webServer.Options{}, false, fmt.Errorf("HTTP 配置 %s 只能包含一个 JSON 值", item.key)
			}
			return webServer.Options{}, false, fmt.Errorf("HTTP 配置 %s 包含多余内容: %w", item.key, err)
		}
		exists = true
	}
	return config, exists, nil
}
