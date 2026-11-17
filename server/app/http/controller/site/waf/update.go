package waf

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	serviceSite "server/app/service/site"
	"server/app/util/context"
)

// Update 更新网站防火墙配置。
func Update(c *context.Context) {
	var form struct {
		Id     int64                  `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Config *serviceSite.WAFUpdate `json:"config" validate:"omitempty" label:"WAF配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	err := service.Site.UpdateWAF(form.Id, form.Config)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站WAF", "修改网站["+strconv.FormatInt(form.Id, 10)+"]WAF配置")
		c.Success("修改成功")
	}
}
