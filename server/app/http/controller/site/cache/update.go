package cache

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	serviceSite "server/app/service/site"
	"server/app/util/context"
)

// Update 更新网站缓存策略。
func Update(c *context.Context) {
	var form struct {
		Id     int64                    `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Config *serviceSite.CacheUpdate `json:"config" validate:"omitempty" label:"缓存配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
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
