package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
)

// Info 读取包信息并检查系统与依赖声明，不解释或执行插件代码。
func (m *Manager) Info(directory string) (*Package, error) {
	return m.info(directory)
}

func (m *Manager) info(directory string) (*Package, error) {
	manifest, err := m.Manifest(directory)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(m.root)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	code, err := root.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	defer code.Close()
	server, err := code.Stat("server")
	if err != nil || !server.IsDir() {
		return nil, errors.New("插件后端入口目录不存在")
	}
	if manifest.Frontend != "" {
		entry, err := code.Stat(manifest.Frontend)
		if err != nil || !entry.Mode().IsRegular() {
			return nil, errors.New("插件前端入口不存在或类型不正确")
		}
	}
	if manifest.Icon != "" {
		icon, err := code.Stat(manifest.Icon)
		if err != nil || !icon.Mode().IsRegular() || icon.Size() > 8<<20 {
			return nil, errors.New("插件图标必须是存在且不超过8MB的普通文件")
		}
	}
	result := &Package{
		Directory: directory,
		Manifest:  manifest,
		Available: true,
	}
	if err = m.checkSupport(manifest); err != nil {
		result.Available, result.UnavailableReason = false, err.Error()
	}
	return result, nil
}

// Manifest 只读取并校验清单，依赖检查不受插件代码或页面文件缺失影响。
func (m *Manager) Manifest(directory string) (*Manifest, error) {
	if m.ctx.Err() != nil {
		return nil, errors.New("插件管理器已关闭")
	}
	if !filepath.IsLocal(directory) || directory == "." || strings.ContainsAny(directory, `\:`) || filepath.ToSlash(filepath.Clean(directory)) != directory {
		return nil, errors.New("插件目录必须是包内相对路径")
	}
	root, err := os.OpenRoot(m.root)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	code, err := root.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	defer code.Close()
	file, err := code.Open(m.options.ManifestName)
	if err != nil {
		return nil, fmt.Errorf("读取 %s 失败: %w", m.options.ManifestName, err)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, (256<<10)+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 256<<10 {
		return nil, errors.New("插件描述文件不能超过256KB")
	}
	var manifest Manifest
	if err = json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("插件描述无效: %w", err)
	}
	identifier := regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
	if !identifier.MatchString(manifest.Id) || strings.TrimSpace(manifest.Name) == "" || len(manifest.Name) > 100 || strings.TrimSpace(manifest.Version) == "" || len(manifest.Version) > 50 {
		return nil, errors.New("插件标识、名称或版本不合法")
	}
	requires := make(map[string]bool, len(manifest.Requires))
	for _, id := range manifest.Requires {
		if !identifier.MatchString(id) {
			return nil, fmt.Errorf("依赖软件标识不合法: %s", id)
		}
		if id == manifest.Id {
			return nil, errors.New("软件不能依赖自身")
		}
		if requires[id] {
			return nil, fmt.Errorf("依赖软件重复: %s", id)
		}
		requires[id] = true
	}
	if manifest.Frontend != "" {
		path := manifest.Frontend
		if !filepath.IsLocal(path) || path == "." || strings.ContainsAny(path, `\:`) || filepath.ToSlash(filepath.Clean(path)) != path {
			return nil, errors.New("插件前端入口必须是包内相对路径")
		}
		if !strings.HasPrefix(manifest.Frontend, "web/") {
			return nil, errors.New("插件前端入口必须位于 web 目录")
		}
	}
	if manifest.Icon != "" {
		path := manifest.Icon
		if !filepath.IsLocal(path) || strings.ContainsAny(path, `\:`) || filepath.ToSlash(filepath.Clean(path)) != path || !strings.HasPrefix(path, "web/") {
			return nil, errors.New("插件图标必须是 web 目录内的相对路径")
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".svg", ".png", ".jpg", ".jpeg", ".webp", ".ico", ".gif":
		default:
			return nil, errors.New("插件图标仅支持 SVG、PNG、JPG、WebP、ICO 或 GIF 图片")
		}
	}
	for _, system := range manifest.Platforms {
		switch system {
		case "aix", "android", "darwin", "dragonfly", "freebsd", "illumos", "ios", "js", "linux", "netbsd", "openbsd", "plan9", "solaris", "wasip1", "windows":
		default:
			return nil, fmt.Errorf("插件声明的系统不合法: %s，请使用 linux、windows、darwin 等 Go 系统标识", system)
		}
	}
	return &manifest, nil
}

func (m *Manager) checkSupport(manifest *Manifest) error {
	if len(manifest.Platforms) > 0 && !slices.Contains(manifest.Platforms, runtime.GOOS) {
		return fmt.Errorf("此插件仅支持 %s，当前服务器系统为 %s", strings.Join(manifest.Platforms, "、"), runtime.GOOS)
	}
	return nil
}
