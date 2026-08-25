package site

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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
		}
	}
	mergePage := func(target *webServer.ResponseOptions, update *HTTPPageUpdate) {
		if update == nil {
			return
		}
		if update.Status != nil {
			target.Status = *update.Status
		}
		if update.Body != nil {
			target.Body = *update.Body
		}
		if update.Headers != nil {
			target.Headers = make(map[string][]string, len(*update.Headers))
			for name, values := range *update.Headers {
				target.Headers[name] = slices.Clone(values)
			}
		}
	}
	mergePage(&candidate.NotFoundPage, config.NotFoundPage)
	mergePage(&candidate.SiteNotFoundPage, config.SiteNotFoundPage)
	mergePage(&candidate.SiteDisabledPage, config.SiteDisabledPage)
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
	configs, err := s.httpConfigs(candidate, config)
	if err != nil {
		return err
	}
	if len(configs) == 0 {
		return nil
	}
	effective := candidate
	if effective.LogPath == "" {
		effective.LogPath, err = filepath.Abs(constant.ServerLogPath)
		if err != nil {
			return fmt.Errorf("解析 HTTP 服务日志路径失败: %w", err)
		}
		effective.LogPath = filepath.Clean(effective.LogPath)
	}
	if effective.ConfigPath == "" {
		effective.ConfigPath = filepath.Join(s.root, "http", "config.json")
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
	err = s.config.Sets(configs)
	if err == nil {
		return nil
	}
	return rollback(fmt.Errorf("保存 HTTP 配置失败: %w", err))
}

func (s *service) httpConfigs(config webServer.Options, update *HTTPUpdate) (map[string]string, error) {
	configs := make(map[string]string)
	for _, item := range []struct {
		key     string
		changed bool
		value   interface{}
	}{
		{constant.SiteHTTPListen, update == nil || update.HTTPListen != nil, config.HTTPListen},
		{constant.SiteHTTPSListen, update == nil || update.HTTPSListen != nil, config.HTTPSListen},
		{constant.SiteHTTPProtocols, update == nil || update.Protocols != nil, config.Protocols},
		{constant.SiteHTTPDefaultSite, update == nil || update.DefaultSite != nil, config.DefaultSite},
		{constant.SiteHTTPNotFoundPage, update == nil || update.NotFoundPage != nil, config.NotFoundPage},
		{constant.SiteHTTPSiteNotFoundPage, update == nil || update.SiteNotFoundPage != nil, config.SiteNotFoundPage},
		{constant.SiteHTTPSiteDisabledPage, update == nil || update.SiteDisabledPage != nil, config.SiteDisabledPage},
		{constant.SiteHTTPDataPath, update == nil || update.DataPath != nil, config.DataPath},
		{constant.SiteHTTPCachePath, update == nil || update.CachePath != nil, config.CachePath},
		{constant.SiteHTTPLogPath, update == nil || update.LogPath != nil, config.LogPath},
		{constant.SiteHTTPWAFLogPath, update == nil || update.WAFLogPath != nil, config.WAFLogPath},
		{constant.SiteHTTPConfigPath, update == nil || update.ConfigPath != nil, config.ConfigPath},
		{constant.SiteHTTPLogLevel, update == nil || update.LogLevel != nil, config.LogLevel},
		{constant.SiteHTTPTrustedProxies, update == nil || update.TrustedProxies != nil, config.TrustedProxies},
		{constant.SiteHTTPClientIPHeaders, update == nil || update.ClientIPHeaders != nil, config.ClientIPHeaders},
		{constant.SiteHTTPTrustedProxiesStrict, update == nil || update.TrustedProxiesStrict != nil, config.TrustedProxiesStrict},
		{constant.SiteHTTPReadTimeout, update == nil || update.ReadTimeout != nil, config.ReadTimeout},
		{constant.SiteHTTPReadHeaderTimeout, update == nil || update.ReadHeaderTimeout != nil, config.ReadHeaderTimeout},
		{constant.SiteHTTPWriteTimeout, update == nil || update.WriteTimeout != nil, config.WriteTimeout},
		{constant.SiteHTTPIdleTimeout, update == nil || update.IdleTimeout != nil, config.IdleTimeout},
		{constant.SiteHTTPGracePeriod, update == nil || update.GracePeriod != nil, config.GracePeriod},
		{constant.SiteHTTPMaxHeaderBytes, update == nil || update.MaxHeaderBytes != nil, config.MaxHeaderBytes},
		{constant.SiteHTTPChallengeHost, update == nil || update.HTTPChallengeHost != nil, config.HTTPChallengeHost},
		{constant.SiteHTTPChallengePort, update == nil || update.HTTPChallengePort != nil, config.HTTPChallengePort},
		{constant.SiteHTTPACMEEmail, update == nil || update.ACMEEmail != nil, config.ACMEEmail},
	} {
		if !item.changed {
			continue
		}
		content, err := json.Marshal(item.value)
		if err != nil {
			return nil, fmt.Errorf("序列化 HTTP 配置 %s 失败: %w", item.key, err)
		}
		configs[item.key] = string(content)
	}
	return configs, nil
}

// loadHTTP 读取并合并已保存的 HTTP 服务全局参数。
func (s *service) loadHTTP(includeLegacy bool) (webServer.Options, bool, error) {
	values := s.config.Group(constant.SiteGroup)
	config := webServer.Options{}
	exists := false
	decode := func(key string, target interface{}) error {
		raw, ok := values[key]
		if !ok {
			return nil
		}
		decoder := json.NewDecoder(strings.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(target); err != nil {
			return fmt.Errorf("HTTP 配置 %s 格式错误: %w", key, err)
		}
		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
			if err == nil {
				return fmt.Errorf("HTTP 配置 %s 只能包含一个 JSON 值", key)
			}
			return fmt.Errorf("HTTP 配置 %s 包含多余内容: %w", key, err)
		}
		exists = true
		return nil
	}
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
	splitCount := 0
	for _, item := range fields {
		if _, ok := values[item.key]; ok {
			splitCount++
		}
	}
	legacyRaw := strings.TrimSpace(values[constant.SiteHTTPGroup])
	if includeLegacy && legacyRaw != "" && splitCount < len(fields) {
		var legacy *webServer.Options
		legacyErr := func() error {
			decoder := json.NewDecoder(strings.NewReader(legacyRaw))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&legacy); err != nil {
				return fmt.Errorf("旧版 HTTP 配置格式错误: %w", err)
			}
			if legacy == nil {
				return errors.New("旧版 HTTP 配置必须是 JSON 对象")
			}
			if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
				if err == nil {
					return errors.New("旧版 HTTP 配置只能包含一个 JSON 对象")
				}
				return fmt.Errorf("旧版 HTTP 配置包含多余内容: %w", err)
			}
			return nil
		}()
		if legacyErr != nil {
			if splitCount == 0 {
				return webServer.Options{}, false, legacyErr
			}
			log.Printf("%v，已忽略旧配置并使用拆分配置", legacyErr)
		} else {
			config = *legacy
			exists = true
			legacyRoot, pathErr := filepath.Abs(filepath.Join(constant.BasePath, "data", "site", "http"))
			if pathErr != nil {
				return webServer.Options{}, false, fmt.Errorf("解析旧版 HTTP 目录失败: %w", pathErr)
			}
			for target, path := range map[*string]string{
				&config.DataPath:   filepath.Join(legacyRoot, "data"),
				&config.CachePath:  filepath.Join(legacyRoot, "cache"),
				&config.LogPath:    filepath.Join(legacyRoot, "logs", "http.log"),
				&config.WAFLogPath: filepath.Join(legacyRoot, "logs", "waf.log"),
			} {
				if filepath.Clean(*target) == filepath.Clean(path) {
					*target = ""
				}
			}
		}
	}
	for _, item := range fields {
		if err := decode(item.key, item.target); err != nil {
			return webServer.Options{}, false, err
		}
	}
	if !exists {
		return webServer.Options{}, false, nil
	}
	return config, true, nil
}
