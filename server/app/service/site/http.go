package site

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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
	candidate, exists, err := s.loadHTTP()
	if err != nil {
		log.Printf("网站 HTTP 配置损坏，本次修改将以当前运行参数为基础覆盖保存: %v", err)
		candidate = previous
		exists = false
	}
	if !exists {
		candidate = previous
	}
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
		candidate.DataPath = *config.DataPath
	}
	if config.CachePath != nil {
		candidate.CachePath = *config.CachePath
	}
	if config.LogPath != nil {
		candidate.LogPath = *config.LogPath
	}
	if config.WAFLogPath != nil {
		candidate.WAFLogPath = *config.WAFLogPath
	}
	if config.ConfigPath != nil {
		candidate.ConfigPath = *config.ConfigPath
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
	if err = s.http.UpdateOptions(candidate); err != nil {
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
	content, err := json.Marshal(s.http.Options())
	if err == nil {
		err = s.config.Set(constant.SiteHTTPOptions, string(content))
	}
	if err == nil {
		return nil
	}
	return rollback(fmt.Errorf("保存 HTTP 配置失败: %w", err))
}

// loadHTTP 绕过缓存读取已保存的 HTTP 服务全局参数。
func (s *service) loadHTTP() (webServer.Options, bool, error) {
	raw, err := s.config.Load(constant.SiteGroup, constant.SiteHTTPOptions)
	if err != nil {
		return webServer.Options{}, false, fmt.Errorf("读取 HTTP 配置失败: %w", err)
	}
	if strings.TrimSpace(raw) == "" {
		return webServer.Options{}, false, nil
	}
	var config *webServer.Options
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&config); err != nil {
		return webServer.Options{}, false, fmt.Errorf("HTTP 配置格式错误: %w", err)
	}
	if config == nil {
		return webServer.Options{}, false, errors.New("HTTP 配置必须是 JSON 对象")
	}
	if err = decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return webServer.Options{}, false, errors.New("HTTP 配置只能包含一个 JSON 对象")
		}
		return webServer.Options{}, false, fmt.Errorf("HTTP 配置包含多余内容: %w", err)
	}
	return *config, true, nil
}
