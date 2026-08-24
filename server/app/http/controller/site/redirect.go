package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	siteService "server/app/service/site"
	"server/app/util/context"
)

// Redirect 获取或更新网站重定向规则。
func Redirect(c *context.Context) {
	var form struct {
		Action string                        `json:"action" default:"get" validate:"required,oneof=get set" label:"操作类型"`
		Id     int64                         `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Config *[]siteService.RedirectConfig `json:"config" validate:"omitempty,max=1000,dive" label:"重定向配置"`
	}
	if ok, message := c.ValidatorAll(&form); !ok {
		c.Error(message)
		return
	}
	if form.Action == "get" {
		result, err := service.Site.GetRedirects(form.Id)
		if err != nil {
			c.Error(err.Error())
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站重定向", "查看网站["+strconv.FormatInt(form.Id, 10)+"]重定向配置")
		c.SuccessWithData("获取成功", result)
		return
	}
	if form.Config == nil {
		c.Error("重定向配置不能为空")
		return
	}
	if err := service.Site.UpdateRedirects(form.Id, *form.Config); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站重定向", "修改网站["+strconv.FormatInt(form.Id, 10)+"]重定向配置")
	c.Success("修改成功")
}
