package sdk

import (
	"path/filepath"
	"server/app/util/plugin"
)

// NewManager 为每个插件注入面板基础资源与同一组业务服务。
func NewManager(root string, app App) *plugin.Manager {
	packageRoot, err := filepath.Abs(root)
	if err != nil {
		packageRoot = root
	}
	return plugin.NewManager(root, plugin.Options{
		ManifestName:    "software.json",
		MaxArchiveSize:  (16 << 30) + (80 << 20),
		MaxExpandedSize: (16 << 30) + (64 << 20),
		App: func(code string) interface{} {
			instance := app
			instance.InstanceID = code
			instance.PackagePath = filepath.Join(packageRoot, code)
			return &instance
		},
	})
}
