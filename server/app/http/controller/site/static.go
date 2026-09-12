package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	serviceSite "server/app/service/site"
	"server/app/util/context"
)

// Static 获取或更新静态网站文件服务配置。
func Static(c *context.Context) {
	var form struct {
		Action string                    `json:"action" default:"get" validate:"required,oneof=get set" label:"操作类型"`
		Id     int64                     `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Config *serviceSite.StaticUpdate `json:"config" validate:"omitempty" label:"静态网站配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "get" {
		data, err := service.Site.GetStatic(form.Id)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看静态网站配置", "查看网站["+strconv.FormatInt(form.Id, 10)+"]静态网站配置")
			c.SuccessWithData("获取成功", data)
		}
	} else {
		if form.Config == nil {
			c.Error("静态网站配置不能为空")
			return
		}
		err := service.Site.UpdateStatic(form.Id, form.Config)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改静态网站配置", "修改网站["+strconv.FormatInt(form.Id, 10)+"]静态网站配置")
			c.Success("修改成功")
		}
	}
}
