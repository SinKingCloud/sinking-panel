package cache

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Clear 清理网站缓存。
func Clear(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	err := service.Site.ClearCache(form.Id)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "清理网站缓存", "清理网站["+strconv.FormatInt(form.Id, 10)+"]缓存")
		c.Success("清理成功")
	}
}
