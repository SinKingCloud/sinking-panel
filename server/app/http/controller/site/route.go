package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	serviceSite "server/app/service/site"
	"server/app/util/context"
)

// Route 获取或更新网站自定义路由。
func Route(c *context.Context) {
	var form struct {
		Action string                     `json:"action" default:"get" validate:"required,oneof=get set" label:"操作类型"`
		Id     int64                      `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Config *[]serviceSite.RouteConfig `json:"config" validate:"omitempty,max=1000,dive" label:"路由配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "get" {
		data, err := service.Site.GetRoutes(form.Id)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站路由", "查看网站["+strconv.FormatInt(form.Id, 10)+"]路由配置")
			c.SuccessWithData("获取成功", data)
		}
	} else {
		if form.Config == nil {
			c.Error("路由配置不能为空")
			return
		}
		err := service.Site.UpdateRoutes(form.Id, *form.Config)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站路由", "修改网站["+strconv.FormatInt(form.Id, 10)+"]路由配置")
			c.Success("修改成功")
		}
	}
}
