package site

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"server/app/constant"
	"server/app/enum/site_kind"
	"server/app/enum/site_status"
	"server/app/model"
	repositorySite "server/app/repository/site"
	"server/app/util/caddy"
	fileLog "server/app/util/log"
	"server/app/util/page"
	"server/app/util/str"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

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
	if err := s.validateTypeId(data.TypeId); err != nil {
		return nil, err
	}
	id := str.GetSnowWorkIns().GetId()
	root := strings.TrimSpace(data.Root)
	autoRoot := false
	createdRoot := false
	var createdPages []string
	if root == "" {
		if len(data.Domains) == 0 && data.Type != site_kind.General {
			return nil, errors.New("网站至少需要绑定一个域名")
		}
		directory := strconv.FormatInt(id, 10)
		if len(data.Domains) > 0 {
			domain, normalizeErr := s.normalizeDomain(data.Domains[0])
			if normalizeErr != nil {
				return nil, normalizeErr
			}
			directory = domain
			if strings.HasPrefix(domain, "*.") {
				directory = "_wildcard." + strings.TrimPrefix(domain, "*.")
			} else if strings.Contains(domain, ":") {
				directory = "_ip6-" + strings.NewReplacer(":", "-", "%", "-").Replace(domain)
			}
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
			var cleanupErr error
			for _, path := range createdPages {
				if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
					cleanupErr = errors.Join(cleanupErr, fmt.Errorf("清理网站默认页面失败: %w", removeErr))
				}
			}
			return cleanupErr
		}
		if removeErr := os.RemoveAll(root); removeErr != nil {
			return fmt.Errorf("清理自动网站目录失败: %w", removeErr)
		}
		createdRoot = false
		return nil
	}
	domainInput := make([]SiteDomain, 0, len(data.Domains))
	for _, domain := range data.Domains {
		domainInput = append(domainInput, SiteDomain{Domain: domain})
	}

	record := &model.Site{
		Id:      id,
		TypeId:  data.TypeId,
		Name:    data.Name,
		Type:    data.Type,
		Status:  data.Status,
		Root:    root,
		RunPath: data.RunPath,
		Config:  config,
	}
	var domains []*model.SiteDomain
	err = s.database.Transaction(func(tx *gorm.DB) error {
		var prepareErr error
		var value interface{}
		value, domains, prepareErr = s.prepareSite(record, domainInput, nil, false, tx)
		if prepareErr != nil {
			return prepareErr
		}
		if record.Type == site_kind.Static || record.Type == site_kind.PHP {
			runRoot, pathErr := s.runRoot(record)
			if pathErr != nil {
				return pathErr
			}
			if pathErr = os.MkdirAll(runRoot, 0755); pathErr != nil {
				return fmt.Errorf("创建网站运行目录失败: %w", pathErr)
			}
			// 已部署的网站保留原首页，避免新增 index.html 遮盖 index.htm 或 PHP 首页。
			indexes := []string{"index.html", "index.htm"}
			switch config := value.(type) {
			case StaticConfig:
				indexes = append(indexes, config.Index...)
			case PHPConfig:
				indexes = append(indexes, "index.php", config.FastCGI.Index)
			}
			hasIndex := false
			for _, index := range indexes {
				if index == "" {
					continue
				}
				if _, statErr := os.Lstat(filepath.Join(runRoot, index)); statErr == nil {
					hasIndex = true
					break
				} else if !os.IsNotExist(statErr) {
					return fmt.Errorf("检查网站首页失败: %w", statErr)
				}
			}
			pages := s.http.Options()
			for _, page := range []struct{ name, body string }{
				{"index.html", pages.InitialPage.Body},
				{"404.html", pages.NotFoundPage.Body},
			} {
				if page.name == "index.html" && hasIndex {
					continue
				}
				path := filepath.Join(runRoot, page.name)
				file, openErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
				if os.IsExist(openErr) {
					continue
				}
				if openErr != nil {
					return fmt.Errorf("创建网站默认页面失败: %w", openErr)
				}
				createdPages = append(createdPages, path)
				_, writeErr := file.WriteString(page.body)
				if writeErr = errors.Join(writeErr, file.Close()); writeErr != nil {
					return fmt.Errorf("写入网站默认页面失败: %w", writeErr)
				}
			}
		}
		if createErr := s.repositorySite.Create(record, tx); createErr != nil {
			return fmt.Errorf("创建网站失败: %w", createErr)
		}
		if createErr := s.repositorySiteDomain.CreateBatch(domains, tx); createErr != nil {
			return fmt.Errorf("创建网站域名失败: %w", createErr)
		}
		return nil
	})
	if err != nil {
		return nil, errors.Join(err, cleanupRoot())
	}
	s.cache.Delete(constant.CacheNameWithSiteNameEnum)
	if syncErr := s.syncLocked(false); syncErr != nil {
		compensateErr := s.database.Transaction(func(tx *gorm.DB) error {
			if err := s.repositorySiteDomain.DeleteBySiteId(record.Id, tx); err != nil {
				return fmt.Errorf("补偿删除网站域名失败: %w", err)
			}
			if err := s.repositorySite.DeleteById(record.Id, tx); err != nil {
				return fmt.Errorf("补偿删除网站失败: %w", err)
			}
			return nil
		})
		restoreRuntimeErr := s.syncLocked(false)
		var cleanupErr error
		if compensateErr == nil && restoreRuntimeErr == nil {
			cleanupErr = cleanupRoot()
		}
		syncErrors := []error{fmt.Errorf("同步网站运行时失败: %w", syncErr)}
		if compensateErr != nil {
			syncErrors = append(syncErrors, fmt.Errorf("恢复数据库失败: %w", compensateErr))
		}
		if restoreRuntimeErr != nil {
			syncErrors = append(syncErrors, fmt.Errorf("恢复网站运行时失败: %w", restoreRuntimeErr))
		}
		return nil, errors.Join(fmt.Errorf("创建网站失败: %w", errors.Join(syncErrors...)), cleanupErr)
	}
	result, err := s.findByIdLocked(record.Id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("网站不存在")
	}
	return result, nil
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
		TypeId:  data.TypeId,
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
		return fmt.Errorf("查询网站失败: %w", err)
	}
	if previous == nil {
		return errors.New("网站不存在")
	}
	previousRecord := *previous
	previous = &previousRecord
	if data.Name == nil && data.TypeId == nil && data.Status == nil && data.Root == nil &&
		data.RunPath == nil && data.Config == nil && data.Domains == nil {
		return nil
	}
	previousDomains, err := s.repositorySiteDomain.SelectBySiteId(id)
	if err != nil {
		return fmt.Errorf("查询网站域名失败: %w", err)
	}
	domainRecords := previousDomains
	previousDomains = make([]*model.SiteDomain, 0, len(domainRecords))
	for _, record := range domainRecords {
		if record != nil {
			domain := *record
			previousDomains = append(previousDomains, &domain)
		}
	}

	candidate := *previous
	if data.Name != nil {
		candidate.Name = *data.Name
	}
	if data.TypeId != nil {
		candidate.TypeId = *data.TypeId
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
	if err = s.validateTypeId(candidate.TypeId); err != nil {
		return err
	}
	domainsChanged := data.Domains != nil
	var domainInput []SiteDomain
	if domainsChanged {
		domainInput = append([]SiteDomain(nil), (*data.Domains)...)
	} else {
		domainInput = make([]SiteDomain, 0, len(previousDomains))
		for _, domain := range previousDomains {
			domainInput = append(domainInput, SiteDomain{Domain: domain.Domain, CertId: domain.CertId})
		}
	}
	previousDefault := s.http.Options().DefaultSite
	clearDefault := domainsChanged && len(domainInput) == 0 && previousDefault == strconv.FormatInt(id, 10)

	var domains []*model.SiteDomain
	err = s.database.Transaction(func(tx *gorm.DB) error {
		var prepareErr error
		validateCertificates := data.Status != nil && previous.Status != candidate.Status && candidate.Status == site_status.Enabled
		_, domains, prepareErr = s.prepareSite(&candidate, domainInput, previousDomains, validateCertificates, tx)
		if prepareErr != nil {
			return prepareErr
		}
		if updateErr := s.repositorySite.UpdateById(id, &repositorySite.UpdateSite{
			Name:    &candidate.Name,
			TypeId:  &candidate.TypeId,
			Status:  &candidate.Status,
			Root:    &candidate.Root,
			RunPath: &candidate.RunPath,
			Config:  &candidate.Config,
		}, tx); updateErr != nil {
			return fmt.Errorf("更新网站失败: %w", updateErr)
		}
		if domainsChanged {
			if deleteErr := s.repositorySiteDomain.DeleteBySiteId(id, tx); deleteErr != nil {
				return fmt.Errorf("更新网站域名前清理旧数据失败: %w", deleteErr)
			}
			if createErr := s.repositorySiteDomain.CreateBatch(domains, tx); createErr != nil {
				return fmt.Errorf("更新网站域名失败: %w", createErr)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.cache.Delete(constant.CacheNameWithSiteNameEnum)
	var syncErr error
	if clearDefault {
		value := ""
		syncErr = s.updateHTTPLocked(&HTTPUpdate{DefaultSite: &value})
	} else {
		syncErr = s.syncLocked(false)
	}
	if syncErr != nil {
		compensateErr := s.restoreSiteRecords(previous, previousDomains)
		restoreRuntimeErr := s.syncLocked(false)
		if clearDefault && compensateErr == nil && restoreRuntimeErr == nil {
			restoreRuntimeErr = s.updateHTTPLocked(&HTTPUpdate{DefaultSite: &previousDefault})
		}
		syncErrors := []error{fmt.Errorf("同步网站运行时失败: %w", syncErr)}
		if compensateErr != nil {
			syncErrors = append(syncErrors, fmt.Errorf("恢复数据库失败: %w", compensateErr))
		}
		if restoreRuntimeErr != nil {
			syncErrors = append(syncErrors, fmt.Errorf("恢复网站运行时失败: %w", restoreRuntimeErr))
		}
		return fmt.Errorf("更新网站失败: %w", errors.Join(syncErrors...))
	}
	return nil
}

// Delete 删除网站和域名，可在运行时同步成功后将网站根目录移至回收站。
func (s *service) Delete(id int64, deleteRoot bool) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if id <= 0 {
		return errors.New("网站 ID 不合法")
	}
	previous, err := s.repositorySite.FindById(id)
	if err != nil {
		return fmt.Errorf("查询网站失败: %w", err)
	}
	if previous == nil {
		return errors.New("网站不存在")
	}
	rootToRecycle := ""
	if deleteRoot {
		rootToRecycle = strings.TrimSpace(previous.Root)
		if rootToRecycle == "" {
			return errors.New("网站根目录为空，无法删除")
		}
		rootToRecycle, err = filepath.Abs(rootToRecycle)
		if err != nil {
			return fmt.Errorf("解析网站根目录失败: %w", err)
		}
		rootToRecycle = filepath.Clean(rootToRecycle)
		pathContains := func(parent, child string) bool {
			relative, relativeErr := filepath.Rel(parent, child)
			return relativeErr == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
		}
		panelRoot, pathErr := filepath.Abs(constant.BasePath)
		if pathErr != nil {
			return fmt.Errorf("解析面板目录失败: %w", pathErr)
		}
		managedRoot, pathErr := filepath.Abs(constant.SiteRootPath)
		if pathErr != nil {
			return fmt.Errorf("解析网站目录失败: %w", pathErr)
		}
		if resolved, resolveErr := filepath.EvalSymlinks(panelRoot); resolveErr == nil {
			panelRoot = filepath.Clean(resolved)
		}
		if resolved, resolveErr := filepath.EvalSymlinks(managedRoot); resolveErr == nil {
			managedRoot = filepath.Clean(resolved)
		}
		validateDeleteRoot := func(root string) error {
			volumeRoot := filepath.VolumeName(root) + string(filepath.Separator)
			if root == filepath.Clean(volumeRoot) {
				return errors.New("不能删除磁盘或文件系统根目录")
			}
			insideManagedRoot := root != managedRoot && pathContains(managedRoot, root)
			if pathContains(root, panelRoot) || pathContains(root, managedRoot) || (pathContains(panelRoot, root) && !insideManagedRoot) {
				return errors.New("网站根目录包含面板数据，不能删除")
			}
			if home, homeErr := os.UserHomeDir(); homeErr == nil && home != "" {
				home, homeErr = filepath.Abs(home)
				if homeErr == nil {
					home = filepath.Clean(home)
					if resolved, resolveErr := filepath.EvalSymlinks(home); resolveErr == nil {
						home = filepath.Clean(resolved)
					}
					if pathContains(root, home) {
						return errors.New("不能删除用户主目录")
					}
				}
			}
			protected := make([]string, 0, 16)
			if runtime.GOOS == "windows" {
				protected = append(protected, os.Getenv("SystemRoot"), os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)"), os.Getenv("ProgramData"))
			} else {
				protected = append(protected, "/bin", "/boot", "/dev", "/etc", "/lib", "/lib64", "/proc", "/run", "/sbin", "/sys", "/usr", "/var/lib", "/var/log", "/var/run", "/var/spool")
				if runtime.GOOS == "darwin" {
					protected = append(protected, "/Applications", "/Library", "/System", "/private/etc", "/private/var")
				}
			}
			for _, path := range protected {
				if path == "" {
					continue
				}
				path, pathErr := filepath.Abs(path)
				if pathErr != nil {
					return fmt.Errorf("解析受保护目录失败: %w", pathErr)
				}
				path = filepath.Clean(path)
				if resolved, resolveErr := filepath.EvalSymlinks(path); resolveErr == nil {
					path = filepath.Clean(resolved)
				}
				if pathContains(root, path) || pathContains(path, root) {
					return errors.New("网站根目录包含系统数据，不能删除")
				}
			}
			protectedRoots := make([]string, 0, 12)
			if runtime.GOOS == "windows" {
				protectedRoots = append(protectedRoots, filepath.Join(filepath.VolumeName(root)+string(filepath.Separator), "Users"))
			} else {
				protectedRoots = append(protectedRoots, "/data", "/home", "/media", "/mnt", "/opt", "/root", "/srv", "/tmp", "/var/tmp", "/var/www", "/www")
				if runtime.GOOS == "darwin" {
					protectedRoots = append(protectedRoots, "/Users", "/Volumes")
				}
			}
			for _, path := range protectedRoots {
				path, pathErr := filepath.Abs(path)
				if pathErr != nil {
					return fmt.Errorf("解析受保护目录失败: %w", pathErr)
				}
				path = filepath.Clean(path)
				if resolved, resolveErr := filepath.EvalSymlinks(path); resolveErr == nil {
					path = filepath.Clean(resolved)
				}
				if pathContains(root, path) {
					return errors.New("不能删除网站公共根目录")
				}
			}
			return nil
		}
		if err = validateDeleteRoot(rootToRecycle); err != nil {
			return err
		}
		if info, statErr := os.Lstat(rootToRecycle); statErr == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return errors.New("网站根目录是符号链接，不能自动删除")
			}
			if !info.IsDir() {
				return errors.New("网站根目录不是目录，不能自动删除")
			}
			resolved, resolveErr := filepath.EvalSymlinks(rootToRecycle)
			if resolveErr != nil {
				return fmt.Errorf("解析网站根目录失败: %w", resolveErr)
			}
			rootToRecycle = filepath.Clean(resolved)
			if err = validateDeleteRoot(rootToRecycle); err != nil {
				return err
			}
		} else if !os.IsNotExist(statErr) {
			return fmt.Errorf("检查网站根目录失败: %w", statErr)
		}
		sites, selectErr := s.repositorySite.SelectAll()
		if selectErr != nil {
			return fmt.Errorf("检查网站目录占用失败: %w", selectErr)
		}
		for _, item := range sites {
			if item.Id == id || strings.TrimSpace(item.Root) == "" {
				continue
			}
			otherRoot, pathErr := filepath.Abs(item.Root)
			if pathErr != nil {
				return fmt.Errorf("解析网站[%s]根目录失败: %w", item.Name, pathErr)
			}
			otherRoot = filepath.Clean(otherRoot)
			if resolved, resolveErr := filepath.EvalSymlinks(otherRoot); resolveErr == nil {
				otherRoot = filepath.Clean(resolved)
			}
			if pathContains(rootToRecycle, otherRoot) || pathContains(otherRoot, rootToRecycle) {
				return fmt.Errorf("网站根目录与网站[%s]共用或嵌套，请保留目录后再删除", item.Name)
			}
		}
	}
	previousHTTP := s.http.Options()
	clearDefault := previousHTTP.DefaultSite == strconv.FormatInt(id, 10)
	restoreDefault := func() error {
		if !clearDefault {
			return nil
		}
		return s.updateHTTPLocked(&HTTPUpdate{DefaultSite: &previousHTTP.DefaultSite})
	}
	previousRecord := *previous
	previous = &previousRecord
	previousDomains, err := s.repositorySiteDomain.SelectBySiteId(id)
	if err != nil {
		return fmt.Errorf("查询网站域名失败: %w", err)
	}
	domainRecords := previousDomains
	previousDomains = make([]*model.SiteDomain, 0, len(domainRecords))
	for _, record := range domainRecords {
		if record != nil {
			domain := *record
			previousDomains = append(previousDomains, &domain)
		}
	}
	logPaths := make([]string, 0, 3)
	for _, logType := range []caddy.LogType{caddy.LogAccess, caddy.LogWAF, caddy.LogProcess} {
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
		if deleteErr := s.repositorySiteDomain.DeleteBySiteId(id, tx); deleteErr != nil {
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

	syncErr := s.syncLocked(false)
	if syncErr != nil {
		compensateErr := s.restoreSiteRecords(previous, previousDomains)
		var restoreRuntimeErr error
		if compensateErr == nil && clearDefault {
			restoreRuntimeErr = restoreDefault()
		} else {
			restoreRuntimeErr = s.syncLocked(false)
		}
		syncErrors := []error{fmt.Errorf("同步网站运行时失败: %w", syncErr)}
		if compensateErr != nil {
			syncErrors = append(syncErrors, fmt.Errorf("恢复数据库失败: %w", compensateErr))
		}
		if restoreRuntimeErr != nil {
			syncErrors = append(syncErrors, fmt.Errorf("恢复网站运行时失败: %w", restoreRuntimeErr))
		}
		return fmt.Errorf("删除网站失败: %w", errors.Join(syncErrors...))
	}
	cacheErr := s.http.DeleteSiteCache(strconv.FormatInt(id, 10))
	var rootErr error
	if rootToRecycle != "" {
		// 目录已不存在时无需回收，保持删除网站操作可正常完成。
		if _, statErr := os.Lstat(rootToRecycle); statErr == nil {
			if recycleErr := s.recycleService.Create(rootToRecycle); recycleErr != nil {
				rootErr = fmt.Errorf("网站根目录移至回收站失败: %w", recycleErr)
			}
		} else if !os.IsNotExist(statErr) {
			rootErr = fmt.Errorf("检查网站根目录失败: %w", statErr)
		}
	}
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
	if cleanupErr := errors.Join(cacheErr, logErr, rootErr); cleanupErr != nil {
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
		return nil, fmt.Errorf("查询网站失败: %w", err)
	}
	if record == nil {
		return nil, nil
	}
	return &Site{
		Id:         record.Id,
		TypeId:     record.TypeId,
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
func (s *service) Select(where *repositorySite.SelectSite, queryPage *page.Query) (*page.Result[*repositorySite.Site], error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if where != nil {
		if where.TypeId != nil && *where.TypeId < 0 {
			return nil, errors.New("网站分类参数错误")
		}
		if where.Type != nil {
			if _, ok := site_kind.Map()[*where.Type]; !ok {
				return nil, errors.New("网站类型参数不合法")
			}
		}
		if where.Status != nil {
			if _, ok := site_status.Map()[*where.Status]; !ok {
				return nil, errors.New("网站状态参数不合法")
			}
		}
	}
	return s.repositorySite.Select(where, queryPage)
}

// ClearCache 清理网站 HTTP 和 HTTPS 响应缓存。
func (s *service) ClearCache(id int64) error {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if id <= 0 {
		return errors.New("网站 ID 不合法")
	}
	record, err := s.repositorySite.FindById(id)
	if err != nil {
		return fmt.Errorf("查询网站失败: %w", err)
	}
	if record == nil {
		return errors.New("网站不存在")
	}
	identifier := strconv.FormatInt(id, 10)
	if _, err := s.http.Site(identifier); err != nil {
		if syncErr := s.syncLocked(false); syncErr != nil {
			return fmt.Errorf("加载网站运行时失败: %w", syncErr)
		}
	}
	if err := s.http.ClearSiteCache(identifier); err != nil {
		return fmt.Errorf("清理网站缓存失败: %w", err)
	}
	return nil
}

func (s *service) restoreSiteRecords(record *model.Site, domains []*model.SiteDomain) error {
	return s.database.Transaction(func(tx *gorm.DB) error {
		if err := s.repositorySiteDomain.DeleteBySiteId(record.Id, tx); err != nil {
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
		restoredDomains := make([]*model.SiteDomain, 0, len(domains))
		for _, record := range domains {
			if record != nil {
				domain := *record
				restoredDomains = append(restoredDomains, &domain)
			}
		}
		if err := s.repositorySiteDomain.CreateBatch(restoredDomains, withoutHooks); err != nil {
			return fmt.Errorf("补偿恢复网站域名失败: %w", err)
		}
		return nil
	})
}
