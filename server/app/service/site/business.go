package site

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"server/app/constant"
	"server/app/enum/site_status"
	"server/app/enum/site_type"
	"server/app/model"
	siteRepository "server/app/repository/site"
	fileLog "server/app/util/log"
	"server/app/util/page"
	webServer "server/app/util/server"
	"server/app/util/str"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type siteMutation struct {
	Name    *string
	Status  *int
	Root    *string
	RunPath *string
	Config  *string
	Domains *[]Domain
}

// Create 创建网站及其域名，并同步运行时配置。
func (s *service) Create(data *CreateSite) (*Site, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if data == nil {
		return nil, errors.New("网站数据不能为空")
	}
	if err := s.validateSiteEnums(data.Type, data.Status); err != nil {
		return nil, err
	}
	id := str.GetSnowWorkIns().GetId()
	root := strings.TrimSpace(data.Root)
	autoRoot := false
	createdRoot := false
	if data.Type != site_type.Proxy && root == "" {
		if len(data.Domains) == 0 {
			return nil, errors.New("网站至少需要绑定一个域名")
		}
		domain, normalizeErr := s.normalizeDomain(data.Domains[0])
		if normalizeErr != nil {
			return nil, normalizeErr
		}
		directory := domain
		if strings.HasPrefix(domain, "*.") {
			directory = "_wildcard." + strings.TrimPrefix(domain, "*.")
		} else if strings.Contains(domain, ":") {
			directory = "_ip6-" + strings.NewReplacer(":", "-", "%", "-").Replace(domain)
		}
		base, pathErr := filepath.Abs(constant.SiteRootPath)
		if pathErr != nil {
			return nil, fmt.Errorf("解析网站根目录失败: %w", pathErr)
		}
		if info, statErr := os.Lstat(base); statErr == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
				return nil, errors.New("网站根目录必须是普通目录")
			}
		} else if os.IsNotExist(statErr) {
			if pathErr = os.MkdirAll(base, 0755); pathErr != nil {
				return nil, fmt.Errorf("创建网站根目录失败: %w", pathErr)
			}
		} else {
			return nil, fmt.Errorf("检查网站根目录失败: %w", statErr)
		}
		if info, statErr := os.Lstat(base); statErr != nil {
			return nil, fmt.Errorf("检查网站根目录失败: %w", statErr)
		} else if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return nil, errors.New("网站根目录必须是普通目录")
		}
		base, pathErr = filepath.EvalSymlinks(base)
		if pathErr != nil {
			return nil, fmt.Errorf("解析网站根目录失败: %w", pathErr)
		}
		root = filepath.Join(base, directory)
		relative, pathErr := filepath.Rel(base, root)
		if pathErr != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return nil, errors.New("自动网站目录超出网站根目录")
		}
		autoRoot = true
	}
	config, err := s.createConfig(data)
	if err != nil {
		return nil, err
	}
	if autoRoot {
		if err = os.Mkdir(root, 0755); err == nil {
			createdRoot = true
		} else if !os.IsExist(err) {
			return nil, fmt.Errorf("创建网站根目录失败: %w", err)
		}
		info, statErr := os.Lstat(root)
		if statErr != nil {
			if createdRoot {
				_ = os.Remove(root)
			}
			return nil, fmt.Errorf("检查网站目录失败: %w", statErr)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			if createdRoot {
				_ = os.Remove(root)
			}
			return nil, errors.New("自动网站目录必须是普通目录")
		}
		resolved, resolveErr := filepath.EvalSymlinks(root)
		if resolveErr != nil {
			if createdRoot {
				_ = os.Remove(root)
			}
			return nil, fmt.Errorf("解析网站目录失败: %w", resolveErr)
		}
		base := filepath.Dir(root)
		relative, relativeErr := filepath.Rel(base, resolved)
		if relativeErr != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			if createdRoot {
				_ = os.Remove(root)
			}
			return nil, errors.New("自动网站目录超出网站根目录")
		}
		root = resolved
	}
	cleanupRoot := func() error {
		if !createdRoot {
			return nil
		}
		if removeErr := os.RemoveAll(root); removeErr != nil {
			return fmt.Errorf("清理自动网站目录失败: %w", removeErr)
		}
		createdRoot = false
		return nil
	}
	domainInput := make([]Domain, 0, len(data.Domains))
	for _, domain := range data.Domains {
		domainInput = append(domainInput, Domain{Domain: domain})
	}

	record := &model.Site{
		Id:      id,
		Name:    data.Name,
		Type:    data.Type,
		Status:  data.Status,
		Root:    root,
		RunPath: data.RunPath,
		Config:  config,
	}
	var domains []*model.Domain
	err = s.database.Transaction(func(tx *gorm.DB) error {
		var prepareErr error
		_, domains, prepareErr = s.prepareSite(record, domainInput, nil, false, tx)
		if prepareErr != nil {
			return prepareErr
		}
		if createErr := s.repositorySite.Create(record, tx); createErr != nil {
			return fmt.Errorf("创建网站失败: %w", createErr)
		}
		if createErr := s.repositoryDomain.CreateBatch(domains, tx); createErr != nil {
			return fmt.Errorf("创建网站域名失败: %w", createErr)
		}
		return nil
	})
	if err != nil {
		return nil, errors.Join(err, cleanupRoot())
	}
	s.cache.Delete(constant.CacheNameWithSiteNameEnum)
	if syncErr := s.syncLocked(); syncErr != nil {
		compensateErr := s.removeSiteRecords(record.Id)
		restoreRuntimeErr := s.syncLocked()
		var cleanupErr error
		if compensateErr == nil && restoreRuntimeErr == nil {
			cleanupErr = cleanupRoot()
		}
		return nil, errors.Join(s.mutationSyncError("创建网站", syncErr, compensateErr, restoreRuntimeErr), cleanupErr)
	}
	return s.findByIdLocked(record.Id)
}

// Update 更新网站基础信息，并同步运行时配置。
func (s *service) Update(id int64, data *UpdateSite) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if data == nil {
		return s.updateLocked(id, nil)
	}
	return s.updateLocked(id, &siteMutation{
		Name:    data.Name,
		Root:    data.Root,
		RunPath: data.RunPath,
	})
}

func (s *service) updateLocked(id int64, data *siteMutation) error {
	if id <= 0 {
		return errors.New("网站 ID 不合法")
	}
	if data == nil {
		return errors.New("网站更新数据不能为空")
	}
	previous, err := s.repositorySite.FindById(id)
	if err != nil {
		return s.nilIfNotFound("查询网站失败", err)
	}
	previous = s.cloneSiteRecord(previous)
	if data.Name == nil && data.Status == nil && data.Root == nil &&
		data.RunPath == nil && data.Config == nil && data.Domains == nil {
		return nil
	}
	previousDomains, err := s.repositoryDomain.SelectBySiteId(id)
	if err != nil {
		return fmt.Errorf("查询网站域名失败: %w", err)
	}
	previousDomains = s.cloneDomainRecords(previousDomains)

	candidate := *previous
	if data.Name != nil {
		candidate.Name = *data.Name
	}
	if data.Status != nil {
		candidate.Status = *data.Status
	}
	if data.Root != nil {
		candidate.Root = *data.Root
	}
	if data.RunPath != nil {
		candidate.RunPath = *data.RunPath
	}
	if data.Config != nil {
		candidate.Config = *data.Config
	}
	if err = s.validateSiteEnums(candidate.Type, candidate.Status); err != nil {
		return err
	}
	domainsChanged := data.Domains != nil
	domainInput := s.domainValues(previousDomains)
	if domainsChanged {
		domainInput = append([]Domain(nil), (*data.Domains)...)
	}

	var domains []*model.Domain
	err = s.database.Transaction(func(tx *gorm.DB) error {
		var prepareErr error
		validateCertificates := data.Status != nil && previous.Status != candidate.Status && candidate.Status == site_status.Enabled
		_, domains, prepareErr = s.prepareSite(&candidate, domainInput, previousDomains, validateCertificates, tx)
		if prepareErr != nil {
			return prepareErr
		}
		if updateErr := s.repositorySite.UpdateById(id, s.completeSiteUpdate(&candidate), tx); updateErr != nil {
			return fmt.Errorf("更新网站失败: %w", updateErr)
		}
		if domainsChanged {
			if deleteErr := s.repositoryDomain.DeleteBySiteId(id, tx); deleteErr != nil {
				return fmt.Errorf("更新网站域名前清理旧数据失败: %w", deleteErr)
			}
			if createErr := s.repositoryDomain.CreateBatch(domains, tx); createErr != nil {
				return fmt.Errorf("更新网站域名失败: %w", createErr)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.cache.Delete(constant.CacheNameWithSiteNameEnum)
	if syncErr := s.syncLocked(); syncErr != nil {
		compensateErr := s.restoreSiteRecords(previous, previousDomains)
		restoreRuntimeErr := s.syncLocked()
		return s.mutationSyncError("更新网站", syncErr, compensateErr, restoreRuntimeErr)
	}
	return nil
}

// Delete 删除网站和域名，运行时同步成功后不再保留其缓存。
func (s *service) Delete(id int64) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if id <= 0 {
		return errors.New("网站 ID 不合法")
	}
	previous, err := s.repositorySite.FindById(id)
	if err != nil {
		return s.nilIfNotFound("查询网站失败", err)
	}
	previousHTTP := s.http.Options()
	clearDefault := previousHTTP.DefaultSite == strconv.FormatInt(id, 10)
	restoreDefault := func() error {
		if !clearDefault {
			return nil
		}
		return s.updateHTTPLocked(&HTTPUpdate{DefaultSite: &previousHTTP.DefaultSite})
	}
	previous = s.cloneSiteRecord(previous)
	previousDomains, err := s.repositoryDomain.SelectBySiteId(id)
	if err != nil {
		return fmt.Errorf("查询网站域名失败: %w", err)
	}
	previousDomains = s.cloneDomainRecords(previousDomains)
	logPaths := make([]string, 0, 3)
	for _, logType := range []webServer.LogType{webServer.LogAccess, webServer.LogWAF, webServer.LogProcess} {
		path, pathErr := s.http.LogPath(strconv.FormatInt(id, 10), logType)
		if pathErr != nil {
			return fmt.Errorf("解析网站日志路径失败: %w", pathErr)
		}
		logPaths = append(logPaths, path)
	}
	if clearDefault {
		value := ""
		if err = s.updateHTTPLocked(&HTTPUpdate{DefaultSite: &value}); err != nil {
			return fmt.Errorf("取消默认站点失败: %w", err)
		}
	}

	err = s.database.Transaction(func(tx *gorm.DB) error {
		if deleteErr := s.repositoryDomain.DeleteBySiteId(id, tx); deleteErr != nil {
			return fmt.Errorf("删除网站域名失败: %w", deleteErr)
		}
		if deleteErr := s.repositorySite.DeleteById(id, tx); deleteErr != nil {
			return fmt.Errorf("删除网站失败: %w", deleteErr)
		}
		return nil
	})
	if err != nil {
		if restoreErr := restoreDefault(); restoreErr != nil {
			return errors.Join(err, fmt.Errorf("恢复默认站点失败: %w", restoreErr))
		}
		return err
	}
	s.cache.Delete(constant.CacheNameWithSiteNameEnum)

	syncErr := s.syncLocked()
	if syncErr != nil {
		compensateErr := s.restoreSiteRecords(previous, previousDomains)
		var restoreRuntimeErr error
		if compensateErr == nil && clearDefault {
			restoreRuntimeErr = restoreDefault()
		} else {
			restoreRuntimeErr = s.syncLocked()
		}
		return s.mutationSyncError("删除网站", syncErr, compensateErr, restoreRuntimeErr)
	}
	cacheErr := s.http.DeleteSiteCache(strconv.FormatInt(id, 10))
	s.logMu.Lock()
	var logErr error
	logDirectory := filepath.Dir(logPaths[0])
	directoryInfo, directoryErr := os.Lstat(logDirectory)
	if directoryErr != nil && !os.IsNotExist(directoryErr) {
		logErr = directoryErr
	} else if directoryErr == nil {
		if directoryInfo.Mode()&os.ModeSymlink != 0 || !directoryInfo.IsDir() {
			logErr = errors.New("网站日志目录无效")
		} else {
			for _, path := range logPaths {
				if filepath.Dir(path) != logDirectory {
					logErr = errors.Join(logErr, errors.New("网站日志文件不在同一目录"))
					continue
				}
				if clearErr := fileLog.Clear(path); clearErr != nil {
					logErr = errors.Join(logErr, clearErr)
					continue
				}
				if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
					logErr = errors.Join(logErr, removeErr)
				}
			}
			if logErr == nil {
				if removeErr := os.Remove(logDirectory); removeErr != nil && !os.IsNotExist(removeErr) {
					logErr = removeErr
				}
			}
		}
	}
	s.logMu.Unlock()
	if cleanupErr := errors.Join(cacheErr, logErr); cleanupErr != nil {
		return fmt.Errorf("网站已删除，但清理运行数据失败: %w", cleanupErr)
	}
	return nil
}

// Enable 启用网站。
func (s *service) Enable(id int64) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	status := site_status.Enabled
	return s.updateLocked(id, &siteMutation{Status: &status})
}

// Disable 停用网站。
func (s *service) Disable(id int64) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	status := site_status.Disabled
	return s.updateLocked(id, &siteMutation{Status: &status})
}

// FindById 查询网站基础详情。
func (s *service) FindById(id int64) (*Site, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	return s.findByIdLocked(id)
}

func (s *service) findByIdLocked(id int64) (*Site, error) {
	if id <= 0 {
		return nil, errors.New("网站 ID 不合法")
	}
	record, err := s.repositorySite.FindById(id)
	if err != nil {
		return nil, s.nilIfNotFound("查询网站失败", err)
	}
	return &Site{
		Id:         record.Id,
		Name:       record.Name,
		Type:       record.Type,
		Status:     record.Status,
		Root:       record.Root,
		RunPath:    record.RunPath,
		CreateTime: record.CreateTime,
		UpdateTime: record.UpdateTime,
	}, nil
}

// Select 分页查询网站。
func (s *service) Select(where *siteRepository.SelectSite, queryPage *page.Query) (*page.Result[*siteRepository.Site], error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	condition := where
	if where != nil {
		copyWhere := *where
		condition = &copyWhere
		if copyWhere.Type != "" {
			value, err := strconv.Atoi(copyWhere.Type)
			if err != nil {
				return nil, errors.New("网站类型参数错误")
			}
			if _, ok := site_type.Map()[value]; !ok {
				return nil, errors.New("网站类型参数不合法")
			}
			copyWhere.Type = strconv.Itoa(value)
		}
		if copyWhere.Status != "" {
			value, err := strconv.Atoi(copyWhere.Status)
			if err != nil {
				return nil, errors.New("网站状态参数错误")
			}
			if _, ok := site_status.Map()[value]; !ok {
				return nil, errors.New("网站状态参数不合法")
			}
			copyWhere.Status = strconv.Itoa(value)
		}
	}
	return s.repositorySite.Select(condition, queryPage)
}

// ClearCache 清理网站 HTTP 和 HTTPS 响应缓存。
func (s *service) ClearCache(id int64) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if id <= 0 {
		return errors.New("网站 ID 不合法")
	}
	if _, err := s.repositorySite.FindById(id); err != nil {
		return s.nilIfNotFound("查询网站失败", err)
	}
	identifier := strconv.FormatInt(id, 10)
	if _, err := s.http.Site(identifier); err != nil {
		if syncErr := s.syncLocked(); syncErr != nil {
			return fmt.Errorf("加载网站运行时失败: %w", syncErr)
		}
	}
	if err := s.http.ClearSiteCache(identifier); err != nil {
		return fmt.Errorf("清理网站缓存失败: %w", err)
	}
	return nil
}

func (s *service) removeSiteRecords(id int64) error {
	return s.database.Transaction(func(tx *gorm.DB) error {
		if err := s.repositoryDomain.DeleteBySiteId(id, tx); err != nil {
			return fmt.Errorf("补偿删除网站域名失败: %w", err)
		}
		if err := s.repositorySite.DeleteById(id, tx); err != nil {
			return fmt.Errorf("补偿删除网站失败: %w", err)
		}
		return nil
	})
}

func (s *service) restoreSiteRecords(record *model.Site, domains []*model.Domain) error {
	return s.database.Transaction(func(tx *gorm.DB) error {
		if err := s.repositoryDomain.DeleteBySiteId(record.Id, tx); err != nil {
			return fmt.Errorf("补偿清理网站域名失败: %w", err)
		}
		if err := s.repositorySite.DeleteById(record.Id, tx); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("补偿清理网站失败: %w", err)
		}
		withoutHooks := tx.Session(&gorm.Session{SkipHooks: true})
		restoredSite := *record
		if err := s.repositorySite.Create(&restoredSite, withoutHooks); err != nil {
			return fmt.Errorf("补偿恢复网站失败: %w", err)
		}
		if err := s.repositoryDomain.CreateBatch(s.cloneDomainRecords(domains), withoutHooks); err != nil {
			return fmt.Errorf("补偿恢复网站域名失败: %w", err)
		}
		return nil
	})
}

func (s *service) completeSiteUpdate(record *model.Site) *siteRepository.UpdateSite {
	return &siteRepository.UpdateSite{
		Name:    &record.Name,
		Status:  &record.Status,
		Root:    &record.Root,
		RunPath: &record.RunPath,
		Config:  &record.Config,
	}
}

func (s *service) domainValues(records []*model.Domain) []Domain {
	result := make([]Domain, 0, len(records))
	for _, record := range records {
		if record != nil {
			result = append(result, Domain{Domain: record.Domain, CertId: record.CertId})
		}
	}
	return result
}

func (s *service) cloneSiteRecord(record *model.Site) *model.Site {
	if record == nil {
		return nil
	}
	clone := *record
	return &clone
}

func (s *service) cloneDomainRecords(records []*model.Domain) []*model.Domain {
	result := make([]*model.Domain, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		clone := *record
		result = append(result, &clone)
	}
	return result
}

func (s *service) mutationSyncError(action string, syncErr, compensateErr, restoreRuntimeErr error) error {
	joined := []error{fmt.Errorf("同步网站运行时失败: %w", syncErr)}
	if compensateErr != nil {
		joined = append(joined, fmt.Errorf("恢复数据库失败: %w", compensateErr))
	}
	if restoreRuntimeErr != nil {
		joined = append(joined, fmt.Errorf("恢复网站运行时失败: %w", restoreRuntimeErr))
	}
	return fmt.Errorf("%s失败: %w", action, errors.Join(joined...))
}

func (s *service) nilIfNotFound(message string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("网站不存在")
	}
	return fmt.Errorf("%s: %w", message, err)
}
