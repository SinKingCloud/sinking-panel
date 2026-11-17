package backup

import (
	"errors"
	"path/filepath"
	"regexp"
	"server/app/constant"
	repositorySite "server/app/repository/site"
	repositorySiteBackup "server/app/repository/site_backup"
	serviceOss "server/app/service/oss"
	"strings"
	"sync"
)

// Manager 分发备份任务，保存文件、上传 OSS 并管理副本。
type Manager struct {
	repositorySite       repositorySite.Interface
	repositorySiteBackup repositorySiteBackup.Interface
	ossService           serviceOss.Service
	root                 string
	siteRoot             string
	gate                 chan struct{}
	pathRuns             sync.Map
	siteRuns             sync.Map
}

// NewManager 创建备份执行器；仅使用本地存储时 ossService 可以为 nil。
// 调用者决定根目录和生命周期，本执行器不创建后台协程。
func NewManager(repositorySite repositorySite.Interface, repositorySiteBackup repositorySiteBackup.Interface, ossService serviceOss.Service, root, siteRoot string) *Manager {
	if strings.TrimSpace(root) == "" {
		root = constant.BackupPath
	}
	root, err := filepath.Abs(root)
	if err != nil {
		root = constant.BackupPath
	}
	if strings.TrimSpace(siteRoot) == "" {
		siteRoot = constant.SiteBackupPath
	}
	siteRoot, err = filepath.Abs(siteRoot)
	if err != nil {
		siteRoot = constant.SiteBackupPath
	}
	return &Manager{
		repositorySite:       repositorySite,
		repositorySiteBackup: repositorySiteBackup,
		ossService:           ossService,
		root:                 filepath.Clean(root),
		siteRoot:             filepath.Clean(siteRoot),
		gate:                 make(chan struct{}, 1),
	}
}

// Path 返回备份根目录，生成文件的业务可据此避免递归备份自己的输出。
func (s *Manager) Path() string {
	return s.root
}

// validateTarget 校验文件归属及扩展名，不解释具体业务。
func (s *Manager) validateTarget(target Target) (Target, error) {
	if err := s.validateBackupPath(target.Key, target.ID); err != nil {
		return Target{}, err
	}
	if target.Extension == "" {
		target.Extension = ".bak"
	}
	if len(target.Extension) > 32 || !regexp.MustCompile(`^\.[a-z0-9]+(?:\.[a-z0-9]+)*$`).MatchString(target.Extension) {
		return Target{}, errors.New("备份扩展名不合法")
	}
	if target.ContentType == "" {
		target.ContentType = "application/octet-stream"
	}
	return target, nil
}

// validateStorage 校验并规范化存储参数；仅云端备份检查 OSS 依赖和存储实例。
func (s *Manager) validateStorage(config Config) (Config, error) {
	config, err := config.Validate()
	if err != nil {
		return Config{}, err
	}
	if config.Storage == "oss" {
		if s.ossService == nil {
			return Config{}, errors.New("对象存储服务未初始化")
		}
		if err := s.ossService.Check(config.OssId); err != nil {
			return Config{}, err
		}
	}
	return config, nil
}
