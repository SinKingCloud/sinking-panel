package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	siteService "server/app/service/site"
	"server/app/util/context"
)

// RateLimit 获取或更新网站访问频率限制。
func RateLimit(c *context.Context) {
	var form struct {
		Action string                       `json:"action" default:"get" validate:"required,oneof=get set" label:"操作类型"`
		Id     int64                        `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Config *siteService.RateLimitUpdate `json:"config" validate:"omitempty" label:"访问频率限制"`
	}
	if ok, message := c.ValidatorAll(&form); !ok {
		c.Error(message)
		return
	}
	if form.Action == "get" {
		result, err := service.Site.GetRateLimit(form.Id)
		if err != nil {
			c.Error(err.Error())
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站访问频率限制", "查看网站["+strconv.FormatInt(form.Id, 10)+"]访问频率限制配置")
		c.SuccessWithData("获取成功", result)
		return
	}
	if form.Config == nil {
		c.Error("访问频率限制不能为空")
		return
	}
	if err := service.Site.UpdateRateLimit(form.Id, form.Config); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站访问频率限制", "修改网站["+strconv.FormatInt(form.Id, 10)+"]访问频率限制配置")
	c.Success("修改成功")
}
