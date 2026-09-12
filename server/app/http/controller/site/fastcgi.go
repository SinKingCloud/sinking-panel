package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	serviceSite "server/app/service/site"
	"server/app/util/context"
)

// FastCGI 获取或更新 PHP 网站的 PHP-FPM 配置。
func FastCGI(c *context.Context) {
	var form struct {
		Action string                     `json:"action" default:"get" validate:"required,oneof=get set" label:"操作类型"`
		Id     int64                      `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Config *serviceSite.FastCGIUpdate `json:"config" validate:"omitempty" label:"PHP-FPM配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "get" {
		data, err := service.Site.GetFastCGI(form.Id)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站PHP-FPM", "查看网站["+strconv.FormatInt(form.Id, 10)+"]PHP-FPM配置")
			c.SuccessWithData("获取成功", data)
		}
	} else {
		if form.Config == nil {
			c.Error("PHP-FPM配置不能为空")
			return
		}
		err := service.Site.UpdateFastCGI(form.Id, form.Config)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站PHP-FPM", "修改网站["+strconv.FormatInt(form.Id, 10)+"]PHP-FPM配置")
			c.Success("修改成功")
		}
	}
}
