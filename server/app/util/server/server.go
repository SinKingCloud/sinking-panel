package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	serverCache "server/app/util/server/cache"
	"sort"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
)

// NewManager 创建一个内存站点管理器。整个进程只应创建一个 Manager。
func NewManager(root string, options ...Options) (*Manager, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("http 数据目录不能为空")
	}
	if len(options) > 1 {
		return nil, errors.New("http 运行参数最多只能传入一组")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("解析 http 数据目录失败: %w", err)
	}
	if err = os.MkdirAll(absRoot, 0700); err != nil {
		return nil, fmt.Errorf("创建 http 数据目录失败: %w", err)
	}
	resolvedRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return nil, fmt.Errorf("解析 http 数据目录失败: %w", err)
	}
	m := &Manager{
		root:  filepath.Clean(resolvedRoot),
		sites: make(map[string]*Site),
	}
	managerOptions := Options{}
	if len(options) > 0 {
		managerOptions = options[0]
	}
	if err = m.setOptions(managerOptions); err != nil {
		return nil, err
	}
	m.config, err = m.buildConfig(m.sites)
	if err != nil {
		return nil, err
	}
	if m.configPath != "" {
		stored, readErr := os.ReadFile(m.configPath)
		if readErr == nil {
			if err = m.validateConfig(stored); err != nil {
				return nil, fmt.Errorf("读取已有 http 配置失败: %w", err)
			}
			if err = os.Chmod(m.configPath, 0600); err != nil {
				return nil, fmt.Errorf("设置 http 配置权限失败: %w", err)
			}
			m.config = append([]byte(nil), stored...)
			m.configMode = true
		} else if !os.IsNotExist(readErr) {
			return nil, fmt.Errorf("读取已有 http 配置失败: %w", readErr)
		}
	}
	return m, nil
}

// Start 启动 HTTP 服务。传入站点时会先替换当前内存站点，空参数则使用已有配置。
func (m *Manager) Start(sites ...Site) error {
	m.operationMu.Lock()
	defer m.operationMu.Unlock()

	m.mu.RLock()
	if m.running {
		m.mu.RUnlock()
		return errors.New("http 已经启动")
	}
	configMode := m.configMode
	currentConfig := append([]byte(nil), m.config...)
	storedSites := m.cloneSites(m.sites)
	m.mu.RUnlock()
	var candidate map[string]*Site
	var config []byte
	var err error
	nextConfigMode := false
	if len(sites) > 0 {
		candidate, err = m.prepareSites(sites)
		if err == nil {
			config, err = m.buildConfig(candidate)
		}
	} else if configMode {
		config = currentConfig
		candidate = make(map[string]*Site)
		nextConfigMode = true
		err = m.validateConfig(config)
	} else {
		source := make([]Site, 0, len(storedSites))
		for _, site := range storedSites {
			source = append(source, *site)
		}
		candidate, err = m.prepareSites(source)
		if err == nil {
			config, err = m.buildConfig(candidate)
		}
	}
	if err != nil {
		return err
	}
	httpRuntime.Lock()
	defer httpRuntime.Unlock()
	if httpRuntime.owner != nil && httpRuntime.owner != m {
		return errors.New("http 运行时已由另一个 Manager 持有")
	}
	if err = caddy.Load(config, false); err != nil {
		return fmt.Errorf("启动 http 失败: %w", errors.Join(err, serverCache.Registry.CloseAll()))
	}
	httpRuntime.owner = m
	if err = m.persistConfig(config); err != nil {
		if stopErr := caddy.Stop(); stopErr != nil {
			m.mu.Lock()
			m.sites = candidate
			m.config = append([]byte(nil), config...)
			m.configMode = nextConfigMode
			m.running = true
			m.mu.Unlock()
			return fmt.Errorf("http 已启动，但保存配置和回滚均失败: %w", errors.Join(err, stopErr))
		}
		httpRuntime.owner = nil
		return fmt.Errorf("保存配置失败，http 启动已回滚: %w", errors.Join(err, serverCache.Registry.CloseAll()))
	}
	m.mu.Lock()
	m.sites = candidate
	m.config = append([]byte(nil), config...)
	m.configMode = nextConfigMode
	m.running = true
	m.mu.Unlock()
	return nil
}

// Stop 停止 Manager 管理的 HTTP 运行时，内存配置会被保留。
func (m *Manager) Stop() error {
	m.operationMu.Lock()
	defer m.operationMu.Unlock()

	m.mu.RLock()
	running := m.running
	m.mu.RUnlock()
	if !running {
		return nil
	}
	httpRuntime.Lock()
	defer httpRuntime.Unlock()
	if httpRuntime.owner != m {
		return errors.New("当前 Manager 已失去 http 运行时所有权")
	}
	if err := caddy.Stop(); err != nil {
		return fmt.Errorf("停止 http 失败: %w", err)
	}
	m.mu.Lock()
	m.running = false
	m.mu.Unlock()
	httpRuntime.owner = nil
	if err := serverCache.Registry.CloseAll(); err != nil {
		return fmt.Errorf("http 已停止，但关闭缓存失败: %w", err)
	}
	return nil
}

// Reload 强制重新加载当前 JSON，适用于证书文件原地更新等配置文本未变化的场景。
func (m *Manager) Reload() error {
	m.operationMu.Lock()
	defer m.operationMu.Unlock()
	m.mu.RLock()
	running := m.running
	configMode := m.configMode
	currentConfig := append([]byte(nil), m.config...)
	storedSites := m.cloneSites(m.sites)
	m.mu.RUnlock()
	if !running {
		return errors.New("http 尚未启动")
	}
	prepared := storedSites
	config := currentConfig
	var err error
	if configMode {
		err = m.validateConfig(config)
	} else {
		input := make([]Site, 0, len(storedSites))
		for _, site := range storedSites {
			input = append(input, *site)
		}
		prepared, err = m.prepareSites(input)
		if err == nil {
			config, err = m.buildConfig(prepared)
		}
		if err == nil && !bytes.Equal(config, currentConfig) {
			err = errors.New("站点配置已变化，请使用 SyncSites 更新后再重新加载")
		}
	}
	if err != nil {
		return err
	}
	httpRuntime.Lock()
	defer httpRuntime.Unlock()
	if httpRuntime.owner != m {
		return errors.New("当前 Manager 已失去 http 运行时所有权")
	}
	if err = caddy.Load(config, true); err != nil {
		return fmt.Errorf("重新加载 http 配置失败: %w", err)
	}
	m.mu.Lock()
	if !configMode {
		m.sites = prepared
	}
	m.config = append([]byte(nil), config...)
	m.mu.Unlock()
	return nil
}

// Running 返回 HTTP 服务当前是否由此 Manager 启动。
func (m *Manager) Running() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// Sites 返回全部站点的只读副本。
func (m *Manager) Sites() []*Site {
	m.mu.RLock()
	sites := make([]*Site, 0, len(m.sites))
	for _, site := range m.sites {
		sites = append(sites, m.cloneSite(site))
	}
	m.mu.RUnlock()
	sort.Slice(sites, func(i, j int) bool {
		if sites[i].Name == sites[j].Name {
			return sites[i].ID < sites[j].ID
		}
		return sites[i].Name < sites[j].Name
	})
	return sites
}

// Site 返回指定站点的只读副本。
func (m *Manager) Site(id string) (*Site, error) {
	id = strings.TrimSpace(id)
	m.mu.RLock()
	site := m.sites[id]
	m.mu.RUnlock()
	if site == nil {
		return nil, errors.New("站点不存在")
	}
	return m.cloneSite(site), nil
}

// SyncSites 用数据库或其他配置源的完整结果替换内存站点。
func (m *Manager) SyncSites(sites []Site) error {
	m.operationMu.Lock()
	defer m.operationMu.Unlock()
	candidate, err := m.prepareSites(sites)
	if err != nil {
		return err
	}
	return m.applySites(candidate)
}

// AddSite 添加站点并在运行时热加载。
func (m *Manager) AddSite(site Site) error {
	m.operationMu.Lock()
	defer m.operationMu.Unlock()

	normalized, err := m.normalizeSite(&site)
	if err != nil {
		return err
	}
	m.mu.RLock()
	if m.configMode {
		m.mu.RUnlock()
		return errors.New("配置文件模式下请使用 SyncSites 切换为站点模式")
	}
	candidate := m.cloneSites(m.sites)
	m.mu.RUnlock()
	if candidate[normalized.ID] != nil {
		return errors.New("站点 ID 已存在")
	}
	candidate[normalized.ID] = normalized
	return m.applySites(candidate)
}

// UpdateSite 修改站点并在运行时热加载。
func (m *Manager) UpdateSite(site Site) error {
	m.operationMu.Lock()
	defer m.operationMu.Unlock()

	normalized, err := m.normalizeSite(&site)
	if err != nil {
		return err
	}
	m.mu.RLock()
	if m.configMode {
		m.mu.RUnlock()
		return errors.New("配置文件模式下请使用 SyncSites 切换为站点模式")
	}
	candidate := m.cloneSites(m.sites)
	m.mu.RUnlock()
	if candidate[normalized.ID] == nil {
		return errors.New("站点不存在")
	}
	candidate[normalized.ID] = normalized
	return m.applySites(candidate)
}

// DeleteSite 删除站点及其缓存，不会删除站点目录和证书。
func (m *Manager) DeleteSite(id string) error {
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
	if candidate[id] == nil {
		return errors.New("站点不存在")
	}
	delete(candidate, id)
	if err := m.applySites(candidate); err != nil {
		return err
	}
	if err := m.removeSiteCache(id); err != nil {
		caddy.Log().Warn("站点已删除，但清理缓存失败", zap.String("site", id), zap.Error(err))
	}
	return nil
}

// EnableSite 启用站点并在运行时热加载。
func (m *Manager) EnableSite(id string) error {
	return m.setSiteEnabled(id, true)
}

// DisableSite 停用站点并在运行时热加载。
func (m *Manager) DisableSite(id string) error {
	return m.setSiteEnabled(id, false)
}

// ClearSiteCache 清空指定站点的 HTTP 和 HTTPS 缓存，运行中的站点无需重载。
func (m *Manager) ClearSiteCache(id string) error {
	m.operationMu.Lock()
	defer m.operationMu.Unlock()

	id = strings.TrimSpace(id)
	m.mu.RLock()
	if m.configMode {
		m.mu.RUnlock()
		return errors.New("配置文件模式下无法按站点清理缓存")
	}
	site := m.sites[id]
	m.mu.RUnlock()
	if site == nil {
		return errors.New("站点不存在")
	}
	target, err := m.siteCachePath(id)
	if err != nil {
		return err
	}
	if err = serverCache.Registry.PurgeTree(m.cachePath, target); err != nil {
		return fmt.Errorf("清空站点缓存失败: %w", err)
	}
	return nil
}

// Config 返回当前 JSON 配置副本，内存证书模式下内容包含私钥。
func (m *Manager) Config() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]byte(nil), m.config...)
}

// SaveConfig 校验、保存并热加载原始 JSON，成功后进入配置文件模式。
func (m *Manager) SaveConfig(config []byte) error {
	m.operationMu.Lock()
	defer m.operationMu.Unlock()
	return m.applyConfig(append([]byte(nil), config...))
}

// ObtainCertificate 显式申请并返回证书，不会自动修改或启用站点 TLS。
func (m *Manager) ObtainCertificate(ctx context.Context, request CertificateRequest) (*Certificate, error) {
	m.operationMu.Lock()
	defer m.operationMu.Unlock()
	httpRuntime.Lock()
	defer httpRuntime.Unlock()
	if httpRuntime.owner != nil && httpRuntime.owner != m {
		return nil, errors.New("http 运行时已由另一个 Manager 持有")
	}
	return m.obtainCertificate(ctx, request, false)
}

// RenewCertificate 强制续签已有证书并返回新证书，不会自动部署。
func (m *Manager) RenewCertificate(ctx context.Context, request CertificateRequest) (*Certificate, error) {
	m.operationMu.Lock()
	defer m.operationMu.Unlock()
	httpRuntime.Lock()
	defer httpRuntime.Unlock()
	if httpRuntime.owner != nil && httpRuntime.owner != m {
		return nil, errors.New("http 运行时已由另一个 Manager 持有")
	}
	return m.obtainCertificate(ctx, request, true)
}
