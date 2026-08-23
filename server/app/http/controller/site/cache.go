package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	siteService "server/app/service/site"
	"server/app/util/context"
)

// Cache 获取或更新网站缓存策略。
func Cache(c *context.Context) {
	var form struct {
		Action string                   `json:"action" default:"get" validate:"required,oneof=get set clear" label:"操作类型"`
		Id     int64                    `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Config *siteService.CacheUpdate `json:"config" validate:"omitempty" label:"缓存配置"`
	}
	if ok, message := c.ValidatorAll(&form); !ok {
		c.Error(message)
		return
	}
	if form.Action == "get" {
		result, err := service.Site.GetCache(form.Id)
		if err != nil {
			c.Error(err.Error())
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站缓存", "查看网站["+strconv.FormatInt(form.Id, 10)+"]缓存配置")
		c.SuccessWithData("获取成功", result)
		return
	}
	if form.Action == "clear" {
		if err := service.Site.ClearCache(form.Id); err != nil {
			c.Error(err.Error())
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "清理网站缓存", "清理网站["+strconv.FormatInt(form.Id, 10)+"]缓存")
		c.Success("清理成功")
		return
	}
	if form.Config == nil {
		c.Error("缓存配置不能为空")
		return
	}
	if err := service.Site.UpdateCache(form.Id, form.Config); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站缓存", "修改网站["+strconv.FormatInt(form.Id, 10)+"]缓存配置")
	c.Success("修改成功")
}
