package cache

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Info 获取网站缓存策略。
func Info(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Site.GetCache(form.Id)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站缓存", "查看网站["+strconv.FormatInt(form.Id, 10)+"]缓存配置")
		c.SuccessWithData("获取成功", data)
	}
}
