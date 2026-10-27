package sdk

import (
	"server/app/util/plugin"
	"server/global"
)

// NewManager 为每个插件注入面板基础资源与同一组业务服务。
func NewManager(root string, service *Service) *plugin.Manager {
	return plugin.NewManager(root, plugin.Options{
		ManifestName: "software.json",
		App: func(code string) interface{} {
			return &App{
				Service:    service,
				Database:   global.App.Database,
				Cache:      global.App.Cache,
				Config:     global.App.Config,
				Log:        global.App.Log,
				InstanceID: code,
			}
		},
	})
}
