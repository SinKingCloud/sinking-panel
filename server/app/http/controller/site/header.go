package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	siteService "server/app/service/site"
	"server/app/util/context"
)

// Header 获取或更新网站请求头和响应头配置。
func Header(c *context.Context) {
	var form struct {
		Action string                    `json:"action" default:"get" validate:"required,oneof=get set" label:"操作类型"`
		Id     int64                     `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Config *siteService.HeaderUpdate `json:"config" validate:"omitempty" label:"请求头配置"`
	}
	if ok, message := c.ValidatorAll(&form); !ok {
		c.Error(message)
		return
	}
	if form.Action == "get" {
		result, err := service.Site.GetHeaders(form.Id)
		if err != nil {
			c.Error(err.Error())
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站请求头", "查看网站["+strconv.FormatInt(form.Id, 10)+"]请求头配置")
		c.SuccessWithData("获取成功", result)
		return
	}
	if form.Config == nil {
		c.Error("请求头配置不能为空")
		return
	}
	if err := service.Site.UpdateHeaders(form.Id, form.Config); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站请求头", "修改网站["+strconv.FormatInt(form.Id, 10)+"]请求头配置")
	c.Success("修改成功")
}
