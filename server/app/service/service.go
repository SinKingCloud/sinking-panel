package service

import (
	"server/app/service/auth"
	"server/app/service/config"
	"server/app/service/cron"
	"server/app/service/file"
	"server/app/service/log"
	"server/app/service/recycle"
	"server/app/service/server"
	"server/app/service/system"
	"server/app/service/task"
)

// service实例
var (
	Config  = config.GetIns()
	Auth    = auth.GetIns()
	Server  = server.GetIns()
	Log     = log.GetIns()
	Cron    = cron.GetIns()
	File    = file.GetIns()
	Recycle = recycle.GetIns()
	Task    = task.GetIns()
	System  = system.GetIns()
)

// Enum 枚举信息
var Enum = map[string]interface{}{
	"log": map[string]interface{}{
		"type": Log.Types(), //日志类型
	},
	"server": map[string]interface{}{
		"auth_type": Server.AuthTypes(), //服务器验证类型
	},
	"cron": map[string]interface{}{
		"type":   Cron.Types(),  //计划任务类型
		"status": Cron.Status(), //计划任务状态
	},
	"task": map[string]interface{}{
		"status": Task.Status(), //任务状态
	},
	"file": map[string]interface{}{
		"formats": File.Formats(), //支持的压缩格式
	},
}
