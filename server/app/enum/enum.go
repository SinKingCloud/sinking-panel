package enum

import (
	"server/app/enum/cert_auto_renew"
	"server/app/enum/cert_type"
	"server/app/enum/container_image"
	"server/app/enum/container_status"
	"server/app/enum/database_kind"
	"server/app/enum/database_type"
	"server/app/enum/file_format"
	"server/app/enum/log_type"
	"server/app/enum/script_module"
	"server/app/enum/secret_name"
	"server/app/enum/secret_provider"
	"server/app/enum/server_auth_type"
	"server/app/enum/site_kind"
	"server/app/enum/site_name"
	"server/app/enum/site_status"
	"server/app/enum/site_type"
	"server/app/enum/software_status"
	"server/app/enum/system_task_status"
	"server/app/enum/task_exec_type"
	"server/app/enum/task_status"
	"server/app/enum/task_type"
	"server/app/enum/type_module"
)

// Data 枚举信息
var Data = map[string]interface{}{
	"software": map[string]interface{}{
		"status": software_status.Map(), //软件安装状态
	},
	"database": func() interface{} {
		return map[string]interface{}{
			"kind": database_kind.Map(), //数据库类型
			"type": database_type.Map(), //数据库分类
		}
	},
	"container": func() interface{} {
		return map[string]interface{}{
			"status": container_status.Map(), //容器运行状态
			"image":  container_image.Map(),  //镜像名称及标签
		}
	},
	"cert": map[string]interface{}{
		"type":       cert_type.Map(),       //证书类型
		"auto_renew": cert_auto_renew.Map(), //自动续签状态
	},
	"secret": func() interface{} {
		return map[string]interface{}{
			"provider": secret_provider.Map(), //密钥服务商
			"name":     secret_name.Map(),     //密钥名称
		}
	},
	"log": map[string]interface{}{
		"type": log_type.Map(), //日志类型
	},
	"server": map[string]interface{}{
		"auth_type": server_auth_type.Map(), //服务器验证类型
	},
	"site": func() interface{} {
		return map[string]interface{}{
			"kind":   site_kind.Map(),   //网站类型
			"type":   site_type.Map(),   //网站分类
			"status": site_status.Map(), //网站状态
			"name":   site_name.Map(),   //可选默认站点
		}
	},
	"task": func() interface{} {
		return map[string]interface{}{
			"exec_type": task_exec_type.Map(), //计划任务执行类型
			"type":      task_type.Map(),      //计划任务分类
			"status":    task_status.Map(),    //计划任务状态
		}
	},
	"type": map[string]interface{}{
		"module": type_module.Map(), //类型所属模块
	},
	"script": map[string]interface{}{
		"module": script_module.Map(), //脚本所属模块
	},
	"system": map[string]interface{}{
		"task_status": system_task_status.Map(), //任务状态
	},
	"file": map[string]interface{}{
		"formats": file_format.Map(), //支持的压缩格式
	},
}
