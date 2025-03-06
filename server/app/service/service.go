package service

import (
	"server/app/service/auth"
	"server/app/service/config"
	"server/app/service/cron"
	"server/app/service/log"
	"server/app/service/server"
)

// service实例
var (
	Config = config.GetIns()
	Auth   = auth.GetIns()
	Server = server.GetIns()
	Log    = log.GetIns()
	Cron   = cron.GetIns()
)

// Enum 枚举信息
var Enum = map[string]interface{}{
	"log.type":         Log.Types(),        //日志类型
	"server.auth_type": Server.AuthTypes(), //服务器验证类型
	"cron.type":        Cron.Types(),       //计划任务类型
	"cron.status":      Cron.Status(),      //计划任务状态
}
