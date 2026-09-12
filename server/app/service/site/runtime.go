package site

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"server/app/constant"
	"server/app/enum/site_status"
	"server/app/enum/site_type"
	"server/app/model"
	processManager "server/app/util/process"
	webServer "server/app/util/server"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// Boot 启动网站服务，并返回负责停止服务的清理函数。
func (s *service) Boot() func() {
	s.operationMu.Lock()
	var err error
	if s.active {
		err = s.syncLocked()
	}
	s.operationMu.Unlock()
	if err != nil {
		log.Printf("启动网站服务失败: %v", err)
	}
	return func() {
		s.operationMu.Lock()
		err := errors.Join(s.http.Stop(), s.process.StopAll())
		s.processes = nil
		s.operationMu.Unlock()
		if err != nil {
			log.Printf("停止网站服务失败: %v", err)
		}
	}
}

// Start 启动网站 HTTP 服务和通用网站进程。
func (s *service) Start() error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	return s.setActiveLocked(true)
}

// Stop 停止网站 HTTP 服务和全部通用网站进程。
func (s *service) Stop() error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	return s.setActiveLocked(false)
}

func (s *service) setActiveLocked(active bool) error {
	previous := s.active
	s.active = active
	var err error
	if active {
		err = s.syncLocked()
	} else {
		err = errors.Join(s.http.Stop(), s.process.StopAll())
		s.processes = nil
	}
	if err == nil {
		if err = s.configService.Set(constant.SiteHTTPEnabled, strconv.FormatBool(active)); err != nil {
			err = fmt.Errorf("保存 HTTP 服务状态失败: %w", err)
		}
	}
	if err == nil {
		return nil
	}
	if previous == active {
		return err
	}
	s.active = previous
	if rollbackErr := s.syncLocked(); rollbackErr != nil {
		return errors.Join(err, fmt.Errorf("恢复网站服务状态失败: %w", rollbackErr))
	}
	return err
}

// Restart 使用数据库中的最新配置重启网站服务。
func (s *service) Restart() error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	sites, processes, err := s.runtime()
	if err != nil {
		return err
	}
	if err = s.http.ValidateSites(sites); err != nil {
		return fmt.Errorf("网站配置验证失败: %w", err)
	}
	wasActive := s.active
	wasRunning := s.http.Running()
	previousProcesses := make([]processManager.Config, len(s.processes))
	for index, config := range s.processes {
		previousProcesses[index] = config
		previousProcesses[index].Env = append([]string(nil), config.Env...)
	}
	loadedSites := s.http.Sites()
	previousSites := make([]webServer.Site, 0, len(loadedSites))
	for _, site := range loadedSites {
		if site != nil {
			previousSites = append(previousSites, *site)
		}
	}
	restore := func(cause error) error {
		s.active = wasActive
		httpErr := s.http.SyncSites(previousSites)
		if httpErr == nil {
			if wasRunning {
				if !s.http.Running() {
					httpErr = s.http.Start()
				}
			} else {
				httpErr = s.http.Stop()
			}
		}
		var processErr error
		if wasActive {
			processErr = s.process.Sync(previousProcesses)
		} else {
			processErr = s.process.StopAll()
		}
		if httpErr == nil && processErr == nil {
			s.processes = make([]processManager.Config, len(previousProcesses))
			for index, config := range previousProcesses {
				s.processes[index] = config
				s.processes[index].Env = append([]string(nil), config.Env...)
			}
			return fmt.Errorf("%w，原运行状态已恢复", cause)
		}
		if httpErr != nil {
			httpErr = fmt.Errorf("恢复原 HTTP 服务失败: %w", httpErr)
		}
		if processErr != nil {
			processErr = fmt.Errorf("恢复原网站进程失败: %w", processErr)
		}
		return errors.Join(cause, httpErr, processErr)
	}
	if err = errors.Join(s.http.Stop(), s.process.StopAll()); err != nil {
		return restore(fmt.Errorf("停止网站服务失败: %w", err))
	}
	s.processes = nil
	s.active = true
	if err = s.http.SyncSites(sites); err == nil {
		err = s.http.Start()
	}
	if err == nil {
		err = s.process.Sync(processes)
	}
	if err != nil {
		return restore(fmt.Errorf("重启网站服务失败: %w", err))
	}
	s.processes = make([]processManager.Config, len(processes))
	for index, config := range processes {
		s.processes[index] = config
		s.processes[index].Env = append([]string(nil), config.Env...)
	}
	if err = s.configService.Set(constant.SiteHTTPEnabled, strconv.FormatBool(true)); err != nil {
		return restore(fmt.Errorf("保存 HTTP 服务状态失败: %w", err))
	}
	return nil
}

// Running 返回网站 HTTP 服务是否正在运行。
func (s *service) Running() bool {
	return s.http.Running()
}

// Sync 将数据库配置同步到 HTTP 服务和通用网站进程。
func (s *service) Sync() error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	config, exists, err := s.loadHTTP()
	if err != nil {
		return fmt.Errorf("读取 HTTP 配置失败: %w", err)
	}
	if exists {
		config.ACMEPath = s.http.Options().ACMEPath
		if err = s.http.UpdateOptions(config); err != nil {
			return fmt.Errorf("应用 HTTP 配置失败: %w", err)
		}
	}
	return s.syncLocked()
}

func (s *service) syncLocked() error {
	sites, processes, err := s.runtime()
	if err != nil {
		return err
	}
	if err = s.http.ValidateSites(sites); err != nil {
		return fmt.Errorf("网站配置验证失败: %w", err)
	}
	if !s.active {
		if err = errors.Join(s.http.Stop(), s.process.StopAll()); err != nil {
			return err
		}
		s.processes = nil
		return s.http.SyncSites(sites)
	}
	if err = s.http.SyncSites(sites); err != nil {
		return fmt.Errorf("同步网站 HTTP 配置失败: %w", err)
	}
	if !s.http.Running() {
		if err = s.http.Start(); err != nil {
			return err
		}
	}
	if err = s.process.Sync(processes); err != nil {
		rollbackErr := s.process.Sync(s.processes)
		if rollbackErr != nil {
			rollbackErr = fmt.Errorf("恢复原网站进程失败: %w", rollbackErr)
		}
		return errors.Join(fmt.Errorf("同步网站进程失败: %w", err), rollbackErr)
	}
	s.processes = make([]processManager.Config, len(processes))
	for index, config := range processes {
		s.processes[index] = config
		s.processes[index].Env = append([]string(nil), config.Env...)
	}
	return nil
}

func (s *service) runtime() ([]webServer.Site, []processManager.Config, error) {
	var records []*model.Site
	var domains []*model.SiteDomain
	var certificates []*model.Cert
	err := s.database.Transaction(func(tx *gorm.DB) error {
		var queryErr error
		records, queryErr = s.repositorySite.SelectAll(tx)
		if queryErr != nil {
			return fmt.Errorf("查询网站失败: %w", queryErr)
		}
		siteIds := make([]int64, 0, len(records))
		for _, record := range records {
			siteIds = append(siteIds, record.Id)
		}
		domains, queryErr = s.repositorySiteDomain.SelectBySiteIds(siteIds, tx)
		if queryErr != nil {
			return fmt.Errorf("查询网站域名失败: %w", queryErr)
		}
		certSet := make(map[int64]struct{})
		for _, domain := range domains {
			if domain.CertId > 0 {
				certSet[domain.CertId] = struct{}{}
			}
		}
		certIds := make([]int64, 0, len(certSet))
		for id := range certSet {
			certIds = append(certIds, id)
		}
		certificates, queryErr = s.repositoryCert.SelectByIds(certIds, tx)
		if queryErr != nil {
			return fmt.Errorf("查询网站证书失败: %w", queryErr)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	domainsBySite := make(map[int64][]*model.SiteDomain, len(records))
	for _, domain := range domains {
		domainsBySite[domain.SiteId] = append(domainsBySite[domain.SiteId], domain)
	}
	certById := make(map[int64]*model.Cert, len(certificates))
	for _, certificate := range certificates {
		certById[certificate.Id] = certificate
	}

	result := make([]webServer.Site, 0, len(records))
	processes := make([]processManager.Config, 0)
	for _, record := range records {
		var config interface{}
		if record.Status == site_status.Enabled {
			_, config, err = s.checkConfig(record.Config, record.Type)
			if err != nil {
				return nil, nil, fmt.Errorf("网站 %s 配置无效: %w", record.Name, err)
			}
		}
		runtimeSite, runtimeProcess, err := s.runtimeSite(record, domainsBySite[record.Id], certById, config)
		if err != nil {
			return nil, nil, err
		}
		if len(runtimeSite.Domains) > 0 {
			result = append(result, runtimeSite)
		}
		if runtimeProcess != nil {
			processes = append(processes, *runtimeProcess)
		}
	}
	return result, processes, nil
}

func (s *service) runtimeSite(record *model.Site, domains []*model.SiteDomain, certificates map[int64]*model.Cert, config interface{}) (webServer.Site, *processManager.Config, error) {
	if len(domains) == 0 && record.Type != site_type.General {
		return webServer.Site{}, nil, fmt.Errorf("网站 %s 没有绑定域名", record.Name)
	}
	runtimeSite := webServer.Site{
		ID:      strconv.FormatInt(record.Id, 10),
		Name:    record.Name,
		Enabled: record.Status == site_status.Enabled,
		Domains: make([]string, 0, len(domains)),
	}
	if len(domains) == 0 {
		if !runtimeSite.Enabled {
			return runtimeSite, nil, nil
		}
		value, ok := config.(GeneralConfig)
		if !ok {
			return webServer.Site{}, nil, fmt.Errorf("网站 %s 配置类型错误", record.Name)
		}
		runtimeProcess, err := s.runtimeProcess(record, &value.Process)
		return runtimeSite, runtimeProcess, err
	}
	certDomains := make(map[int64][]string)
	for _, domain := range domains {
		runtimeSite.Domains = append(runtimeSite.Domains, domain.Domain)
		if domain.CertId > 0 {
			certDomains[domain.CertId] = append(certDomains[domain.CertId], domain.Domain)
		}
	}
	ids := make([]int64, 0, len(certDomains))
	for id := range certDomains {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		certificate := certificates[id]
		if certificate == nil {
			return webServer.Site{}, nil, fmt.Errorf("网站 %s 使用的证书 %d 不存在", record.Name, id)
		}
		runtimeSite.TLS.Certificates = append(runtimeSite.TLS.Certificates, webServer.CertificatePair{
			CertificatePEM: certificate.Certificate,
			PrivateKeyPEM:  certificate.PrivateKey,
			Domains:        certDomains[id],
		})
	}
	runtimeSite.TLS.Enabled = len(runtimeSite.TLS.Certificates) > 0
	if !runtimeSite.Enabled {
		return runtimeSite, nil, nil
	}
	runRoot := ""
	if record.Type != site_type.Proxy && record.Type != site_type.General {
		var err error
		runRoot, err = s.runRoot(record)
		if err != nil {
			return webServer.Site{}, nil, fmt.Errorf("网站 %s 运行目录无效: %w", record.Name, err)
		}
	}
	var common HTTPConfig
	var runtimeProcess *processManager.Config
	switch value := config.(type) {
	case StaticConfig:
		common = value.HTTPConfig
		runtimeSite.Root = runRoot
		runtimeSite.Index = value.Index
		runtimeSite.Browse = value.Browse
		runtimeSite.TryFiles = value.TryFiles
		runtimeSite.Hide = value.Hide
		runtimeSite.Precompressed = value.Precompressed
	case ProxyConfig:
		common = value.HTTPConfig
		proxy := value.Proxy
		runtimeSite.Proxy = &proxy
	case PHPConfig:
		common = value.HTTPConfig
		proxy := value.FastCGI
		proxy.Root = runRoot
		runtimeSite.Proxy = &proxy
	case GeneralConfig:
		common = value.HTTPConfig
		proxy := value.Proxy
		runtimeSite.Proxy = &proxy
		var err error
		runtimeProcess, err = s.runtimeProcess(record, &value.Process)
		if err != nil {
			return webServer.Site{}, nil, err
		}
	default:
		return webServer.Site{}, nil, fmt.Errorf("网站 %s 配置类型错误", record.Name)
	}
	runtimeSite.Redirects = make([]webServer.RedirectOptions, 0, len(common.Redirects))
	for _, redirect := range common.Redirects {
		runtimeSite.Redirects = append(runtimeSite.Redirects, webServer.RedirectOptions{
			Name:        redirect.Name,
			Enabled:     redirect.Enabled,
			Domains:     append([]string(nil), redirect.Domains...),
			Paths:       append([]string(nil), redirect.Paths...),
			Target:      redirect.Target,
			Status:      redirect.Status,
			PreserveURI: redirect.PreserveURI,
		})
	}
	runtimeSite.Routes = make([]webServer.Route, 0, len(common.Routes))
	for _, route := range common.Routes {
		runtimeRoute := webServer.Route{
			Name:          route.Name,
			Paths:         append([]string(nil), route.Paths...),
			Methods:       append([]string(nil), route.Methods...),
			StripPrefix:   route.StripPrefix,
			Rewrite:       route.Rewrite,
			Root:          route.Root,
			Index:         append([]string(nil), route.Index...),
			Browse:        route.Browse,
			TryFiles:      append([]string(nil), route.TryFiles...),
			Hide:          append([]string(nil), route.Hide...),
			Precompressed: append([]string(nil), route.Precompressed...),
		}
		if route.Proxy != nil {
			proxy := *route.Proxy
			runtimeRoute.Proxy = &proxy
		}
		if route.Response != nil {
			response := *route.Response
			runtimeRoute.Response = &response
		}
		runtimeSite.Routes = append(runtimeSite.Routes, runtimeRoute)
	}
	runtimeSite.Headers = common.Headers
	runtimeSite.Compression = common.Compression
	runtimeSite.WAF = common.WAF
	runtimeSite.Cache = common.Cache
	runtimeSite.RateLimit = common.RateLimit
	runtimeSite.TrafficLimit = common.TrafficLimit
	runtimeSite.TLS.RedirectHTTP = common.TLS.RedirectHTTP
	runtimeSite.TLS.MinVersion = common.TLS.MinVersion
	runtimeSite.TLS.MaxVersion = common.TLS.MaxVersion

	if !runtimeSite.TLS.Enabled && runtimeSite.TLS.RedirectHTTP {
		return webServer.Site{}, nil, fmt.Errorf("网站 %s 未绑定证书，不能开启 HTTPS 跳转", record.Name)
	}
	return runtimeSite, runtimeProcess, nil
}

func (s *service) runtimeProcess(record *model.Site, config *ProcessConfig) (*processManager.Config, error) {
	runRoot, err := s.runRoot(record)
	if err != nil {
		return nil, fmt.Errorf("网站 %s 运行目录无效: %w", record.Name, err)
	}
	id := strconv.FormatInt(record.Id, 10)
	processLogPath, err := s.http.LogPath(id, webServer.LogProcess)
	if err != nil {
		return nil, fmt.Errorf("网站 %s 进程日志路径无效: %w", record.Name, err)
	}
	environment := make([]string, 0, len(config.Environment))
	for name, content := range config.Environment {
		environment = append(environment, name+"="+content)
	}
	sort.Strings(environment)
	return &processManager.Config{
		ID:           id,
		Command:      config.Command,
		WorkingDir:   runRoot,
		Env:          environment,
		AutoRestart:  true,
		MaxRetries:   config.MaxRetries,
		RestartDelay: config.RestartDelay,
		StopTimeout:  config.StopTimeout,
		LogPath:      processLogPath,
	}, nil
}

func (s *service) runRoot(record *model.Site) (string, error) {
	root := filepath.Clean(strings.TrimSpace(record.Root))
	if root == "" || root == "." || !filepath.IsAbs(root) {
		return "", errors.New("网站根目录必须是绝对路径")
	}
	runPath := filepath.Clean(filepath.FromSlash(strings.TrimSpace(record.RunPath)))
	if runPath == "." {
		runPath = ""
	}
	if filepath.IsAbs(runPath) || filepath.VolumeName(runPath) != "" || runPath == ".." || strings.HasPrefix(runPath, ".."+string(filepath.Separator)) {
		return "", errors.New("网站运行目录不能超出网站根目录")
	}
	target := filepath.Clean(filepath.Join(root, runPath))
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("网站运行目录不能超出网站根目录")
	}
	return target, nil
}
