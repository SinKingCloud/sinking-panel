package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	siteService "server/app/service/site"
	"server/app/util/context"
)

// SSL 获取或更新网站 TLS 策略和域名证书绑定。
func SSL(c *context.Context) {
	var form struct {
		Action   string                          `json:"action" default:"get" validate:"required,oneof=get set" label:"操作类型"`
		Id       int64                           `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Config   *siteService.TLSUpdate          `json:"config" validate:"omitempty" label:"SSL配置"`
		Bindings []siteService.DomainCertificate `json:"bindings" validate:"omitempty,max=1000,dive" label:"证书绑定"`
	}
	if ok, message := c.ValidatorAll(&form); !ok {
		c.Error(message)
		return
	}
	if form.Action == "get" {
		result, err := service.Site.GetSSL(form.Id)
		if err != nil {
			c.Error(err.Error())
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站SSL", "查看网站["+strconv.FormatInt(form.Id, 10)+"]SSL配置")
		c.SuccessWithData("获取成功", result)
		return
	}
	if form.Config == nil && len(form.Bindings) == 0 {
		c.Error("更新内容不能为空")
		return
	}
	if err := service.Site.UpdateSSL(form.Id, form.Config, form.Bindings); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站SSL", "修改网站["+strconv.FormatInt(form.Id, 10)+"]SSL配置")
	c.Success("修改成功")
}
