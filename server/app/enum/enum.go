package enum

import (
	"server/app/enum/cert_type"
	"server/app/enum/file_format"
	"server/app/enum/log_type"
	"server/app/enum/script_type"
	"server/app/enum/secret_name"
	"server/app/enum/secret_provider"
	"server/app/enum/server_auth_type"
	"server/app/enum/site_name"
	"server/app/enum/site_status"
	"server/app/enum/site_type"
	"server/app/enum/system_task_status"
	"server/app/enum/task_exec_type"
	"server/app/enum/task_status"
	"server/app/enum/task_type"
	"server/app/enum/type_module"
	"server/app/service"
	serviceSecret "server/app/service/secret"
)

// Data 枚举信息
var Data = map[string]interface{}{
	"cert": func() interface{} {
		data := service.Secret.GetData()
		providers := secret_provider.Map()
		return map[string]interface{}{
			"type": cert_type.Map(), //证书类型
			"dns": map[string]interface{}{
				"tencentcloud": map[string]interface{}{"name": providers[secret_provider.TencentCloud], "data": data[secret_provider.TencentCloud]},
				"alidns":       map[string]interface{}{"name": providers[secret_provider.Aliyun], "data": data[secret_provider.Aliyun]},
				"huaweicloud":  map[string]interface{}{"name": providers[secret_provider.HuaweiCloud], "data": data[secret_provider.HuaweiCloud]},
				"volcengine":   map[string]interface{}{"name": providers[secret_provider.Volcengine], "data": data[secret_provider.Volcengine]},
				"baiducloud":   map[string]interface{}{"name": providers[secret_provider.BaiduCloud], "data": data[secret_provider.BaiduCloud]},
				"dnspod":       map[string]interface{}{"name": "DNSPod", "data": &serviceSecret.DNSPod{}},
			},
		}
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
			"type":   site_type.Map(),   //网站类型
			"status": site_status.Map(), //网站状态
			"name":   site_name.Map(),   //可选默认站点
		}
	},
	"script": func() interface{} {
		return map[string]interface{}{
			"type": script_type.Map(), //脚本类型
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
	"system": map[string]interface{}{
		"task_status": system_task_status.Map(), //任务状态
	},
	"file": map[string]interface{}{
		"formats": file_format.Map(), //支持的压缩格式
	},
}
