package ssl

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	serviceSite "server/app/service/site"
	"server/app/util/context"
)

// Update 更新网站 TLS 策略和域名证书绑定。
func Update(c *context.Context) {
	var form struct {
		Id       int64                               `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Config   *serviceSite.TLSUpdate              `json:"config" validate:"omitempty" label:"SSL配置"`
		Bindings []serviceSite.SiteDomainCertificate `json:"bindings" validate:"omitempty,max=1000,dive" label:"证书绑定"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	err := service.Site.UpdateSSL(form.Id, form.Config, form.Bindings)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站SSL", "修改网站["+strconv.FormatInt(form.Id, 10)+"]SSL配置")
		c.Success("修改成功")
	}
}
