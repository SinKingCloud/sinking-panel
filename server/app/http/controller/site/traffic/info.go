package traffic

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Info 获取网站并发和响应速度限制。
func Info(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Site.GetTrafficLimit(form.Id)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站流量限制", "查看网站["+strconv.FormatInt(form.Id, 10)+"]流量限制配置")
		c.SuccessWithData("获取成功", data)
	}
}
