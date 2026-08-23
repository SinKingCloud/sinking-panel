package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	siteService "server/app/service/site"
	"server/app/util/context"
)

// Waf 获取或更新网站防火墙配置。
func Waf(c *context.Context) {
	var form struct {
		Action string                 `json:"action" default:"get" validate:"required,oneof=get set" label:"操作类型"`
		Id     int64                  `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Config *siteService.WAFUpdate `json:"config" validate:"omitempty" label:"WAF配置"`
	}
	if ok, message := c.ValidatorAll(&form); !ok {
		c.Error(message)
		return
	}
	if form.Action == "get" {
		result, err := service.Site.GetWAF(form.Id)
		if err != nil {
			c.Error(err.Error())
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站WAF", "查看网站["+strconv.FormatInt(form.Id, 10)+"]WAF配置")
		c.SuccessWithData("获取成功", result)
		return
	}
	if form.Config == nil {
		c.Error("WAF配置不能为空")
		return
	}
	if err := service.Site.UpdateWAF(form.Id, form.Config); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站WAF", "修改网站["+strconv.FormatInt(form.Id, 10)+"]WAF配置")
	c.Success("修改成功")
}
