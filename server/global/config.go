package global

import "server/app/constant"

// IsDebug 是否debug模式
func (a *Application) IsDebug() bool {
	return a.Config.GetString(constant.ServerMode) == "dev"
}

// ServerAddr 获取服务监听地址
func (a *Application) ServerAddr() (host string, port string) {
	host = a.Config.GetString(constant.ServerHost)
	port = a.Config.GetString(constant.ServerPort)
	if host == "" {
		host = "0.0.0.0"
	}
	if port == "" {
		port = "5678"
	}
	return
}
