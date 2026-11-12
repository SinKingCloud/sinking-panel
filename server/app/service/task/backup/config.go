package backup

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func (s *Manager) parseConfig(content string) (Config, error) {
	var config *Config
	if err := json.Unmarshal([]byte(content), &config); err != nil || config == nil {
		return Config{}, errors.New("备份配置必须是 JSON 对象")
	}
	return *config, nil
}

// Validate 校验保存的备份计划，并返回规范化后的完整配置。
func (s *Manager) Validate(content string) (string, error) {
	config, err := s.parseConfig(content)
	if err != nil {
		return "", err
	}
	config, err = s.validateStorage(config)
	if err != nil {
		return "", err
	}
	switch config.Type {
	case "file", "directory":
		options, err := s.parsePathOptions(config.Options)
		if err != nil {
			return "", err
		}
		options.Paths, err = s.ValidatePaths(options.Paths, config.Type == "directory")
		if err != nil {
			return "", err
		}
		config.Options, err = json.Marshal(options)
		if err != nil {
			return "", err
		}
	case "site":
		if s.repositorySite == nil {
			return "", errors.New("网站仓储未初始化")
		}
		options, err := s.parseSiteOptions(config.Options)
		if err != nil {
			return "", err
		}
		if !options.AllSites {
			sites, err := s.repositorySite.SelectAll()
			if err != nil {
				return "", err
			}
			roots := make(map[int64]string, len(sites))
			for _, site := range sites {
				roots[site.Id] = site.Root
			}
			for _, id := range options.SiteIds {
				root, exists := roots[id]
				if !exists {
					return "", fmt.Errorf("网站 #%d: 网站不存在", id)
				}
				if strings.TrimSpace(root) == "" {
					return "", fmt.Errorf("网站 #%d 未配置根目录", id)
				}
			}
		}
		config.Options, err = json.Marshal(options)
		if err != nil {
			return "", err
		}
	default:
		return "", errors.New("备份类型不支持")
	}
	result, err := json.Marshal(config)
	return string(result), err
}

// Validate 校验并规范化文件存储参数，不查询业务资源。
func (config Config) Validate() (Config, error) {
	config.Storage = strings.TrimSpace(config.Storage)
	config.Prefix = strings.Trim(config.Prefix, "/")
	if config.Storage != "local" && config.Storage != "oss" {
		return Config{}, errors.New("备份存储类型不支持")
	}
	if config.LocalKeep < 0 || config.LocalKeep > 1000 || (config.Storage == "local" && config.LocalKeep == 0) {
		return Config{}, errors.New("本地副本数须为 1 到 1000，云端备份可设为 0")
	}
	if config.OssKeep < 0 || config.OssKeep > 1000 || (config.Storage == "oss" && config.OssKeep == 0) {
		return Config{}, errors.New("云端副本数须为 1 到 1000")
	}
	if config.OssId < 0 || len(config.Prefix) > 512 || strings.ContainsAny(config.Prefix, "\\\x00\r\n") {
		return Config{}, errors.New("备份存储参数不合法")
	}
	for _, part := range strings.Split(config.Prefix, "/") {
		if part == "." || part == ".." {
			return Config{}, errors.New("备份目录不能包含相对路径")
		}
	}
	if config.Storage == "oss" && config.OssId == 0 {
		return Config{}, errors.New("请选择存储实例")
	}
	return config, nil
}
