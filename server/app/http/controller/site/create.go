package site

import (
	"server/app/enum/log_type"
	"server/app/service"
	serviceSite "server/app/service/site"
	"server/app/util/context"
	webServer "server/app/util/server"
)

// Create 创建网站。
func Create(c *context.Context) {
	var form struct {
		Name    string                     `json:"name" default:"" validate:"required,max=100" label:"网站名称"`
		TypeId  int64                      `json:"type_id" default:"0" validate:"min=0" label:"网站分类ID"`
		Type    int                        `json:"type" default:"0" validate:"oneof=0 1 2 3" label:"网站类型"`
		Status  int                        `json:"status" default:"0" validate:"oneof=0 1" label:"网站状态"`
		Root    string                     `json:"root" default:"" validate:"omitempty,max=4096" label:"网站根目录"`
		RunPath string                     `json:"run_path" default:"" validate:"omitempty,max=4096" label:"网站运行目录"`
		Domains []string                   `json:"domains" default:"" validate:"max=1000,unique,dive,required,max=253" label:"网站域名"`
		Static  *serviceSite.StaticOptions `json:"static" label:"静态网站配置"`
		Proxy   *webServer.ProxyOptions    `json:"proxy" label:"反向代理配置"`
		FastCGI *webServer.ProxyOptions    `json:"fastcgi" label:"FastCGI 配置"`
		Process *serviceSite.ProcessConfig `json:"process" label:"进程配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Site.Create(&serviceSite.CreateSite{
		Name:    form.Name,
		TypeId:  form.TypeId,
		Type:    form.Type,
		Status:  form.Status,
		Root:    form.Root,
		RunPath: form.RunPath,
		Domains: form.Domains,
		Static:  form.Static,
		Proxy:   form.Proxy,
		FastCGI: form.FastCGI,
		Process: form.Process,
	})
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "创建网站", "创建网站["+data.Name+"]")
		c.SuccessWithData("创建成功", data)
	}
}
