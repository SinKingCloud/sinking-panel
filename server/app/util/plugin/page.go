package plugin

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Page 读取页面资源，返回内容、MIME 类型和实际路径；省略后缀时在本地解析模块。
func (m *Manager) Page(directory, path string) ([]byte, string, string, error) {
	if m.ctx.Err() != nil {
		return nil, "", "", errors.New("插件管理器已关闭")
	}
	for _, name := range []string{directory, path} {
		if !filepath.IsLocal(name) || name == "." || strings.ContainsAny(name, `\:`) || filepath.ToSlash(filepath.Clean(name)) != name {
			return nil, "", "", errors.New("插件资源路径不合法")
		}
	}
	if !strings.HasPrefix(path, "web/") {
		return nil, "", "", errors.New("插件资源只能位于 web 目录")
	}
	types := map[string]string{
		".js":   "text/javascript; charset=utf-8",
		".mjs":  "text/javascript; charset=utf-8",
		".css":  "text/css; charset=utf-8",
		".ts":   "text/plain; charset=utf-8",
		".tsx":  "text/plain; charset=utf-8",
		".jsx":  "text/plain; charset=utf-8",
		".json": "application/json; charset=utf-8",
		".png":  "image/png",
		".svg":  "image/svg+xml",
	}
	root, err := os.OpenRoot(m.root)
	if err != nil {
		return nil, "", "", err
	}
	defer root.Close()
	code, err := root.OpenRoot(directory)
	if err != nil {
		return nil, "", "", err
	}
	defer code.Close()
	candidates := []string{path}
	if filepath.Ext(path) == "" {
		candidates = nil
		for _, base := range []string{path, path + "/index"} {
			for _, extension := range []string{".tsx", ".ts", ".jsx", ".js", ".mjs", ".json"} {
				candidates = append(candidates, base+extension)
			}
		}
	}
	for _, candidate := range candidates {
		contentType := types[strings.ToLower(filepath.Ext(candidate))]
		if contentType == "" {
			return nil, "", "", errors.New("不支持的插件资源类型")
		}
		file, err := code.Open(candidate)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, "", "", err
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil {
			return nil, "", "", err
		}
		if !info.Mode().IsRegular() || info.Size() > 8<<20 {
			return nil, "", "", errors.New("插件资源必须是不超过8MB的普通文件")
		}
		data, err := io.ReadAll(io.LimitReader(file, (8<<20)+1))
		if err != nil {
			return nil, "", "", err
		}
		if len(data) > 8<<20 {
			return nil, "", "", errors.New("插件资源不能超过8MB")
		}
		return data, contentType, candidate, nil
	}
	return nil, "", "", &os.PathError{
		Op:   "open",
		Path: path,
		Err:  os.ErrNotExist,
	}
}
