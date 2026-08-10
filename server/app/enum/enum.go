package enum

import (
	"server/app/enum/file_format"
	"server/app/enum/log_type"
	"server/app/enum/script_type"
	"server/app/enum/server_auth_type"
	"server/app/enum/system_task_status"
	"server/app/enum/task_status"
	"server/app/enum/task_type"
	"server/app/enum/type_module"
)

// Data 枚举信息
var Data = map[string]interface{}{
	"log": map[string]interface{}{
		"type": log_type.Map(), //日志类型
	},
	"server": map[string]interface{}{
		"auth_type": server_auth_type.Map(), //服务器验证类型
	},
	"script": func() interface{} {
		return map[string]interface{}{
			"type": script_type.Map(), //脚本类型
		}
	},
	"task": map[string]interface{}{
		"type":   task_type.Map(),   //计划任务类型
		"status": task_status.Map(), //计划任务状态
	},
	"type": map[string]interface{}{
		"module": type_module.Map(), //类型所属模块
	},
	"system": map[string]interface{}{
		"task_status": system_task_status.Map(), //任务状态
	},
	"file": map[string]interface{}{
		"formats": file_format.Map(), //支持的压缩格式
	},
}
