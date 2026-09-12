package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	serviceSite "server/app/service/site"
	"server/app/util/context"
)

// Cache 获取或更新网站缓存策略。
func Cache(c *context.Context) {
	var form struct {
		Action string                   `json:"action" default:"get" validate:"required,oneof=get set clear" label:"操作类型"`
		Id     int64                    `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Config *serviceSite.CacheUpdate `json:"config" validate:"omitempty" label:"缓存配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "get" {
		data, err := service.Site.GetCache(form.Id)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站缓存", "查看网站["+strconv.FormatInt(form.Id, 10)+"]缓存配置")
			c.SuccessWithData("获取成功", data)
		}
	} else if form.Action == "clear" {
		err := service.Site.ClearCache(form.Id)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "清理网站缓存", "清理网站["+strconv.FormatInt(form.Id, 10)+"]缓存")
			c.Success("清理成功")
		}
	} else {
		if form.Config == nil {
			c.Error("缓存配置不能为空")
			return
		}
		err := service.Site.UpdateCache(form.Id, form.Config)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站缓存", "修改网站["+strconv.FormatInt(form.Id, 10)+"]缓存配置")
			c.Success("修改成功")
		}
	}
}
