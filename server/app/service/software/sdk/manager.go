package sdk

import (
	"path/filepath"
	"server/app/util/plugin"
	"server/global"
)

// NewManager 为每个插件注入面板基础资源与同一组业务服务。
func NewManager(root string, service *Service) *plugin.Manager {
	packageRoot, err := filepath.Abs(root)
	if err != nil {
		packageRoot = root
	}
	return plugin.NewManager(root, plugin.Options{
		ManifestName:    "software.json",
		MaxArchiveSize:  (16 << 30) + (80 << 20),
		MaxExpandedSize: (16 << 30) + (64 << 20),
		App: func(code string) interface{} {
			return &App{
				Service:     service,
				Database:    global.App.Database,
				Cache:       global.App.Cache,
				Config:      global.App.Config,
				Log:         global.App.Log,
				InstanceID:  code,
				PackagePath: filepath.Join(packageRoot, code),
			}
		},
	})
}
