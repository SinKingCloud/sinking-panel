package site

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Info 获取网站详情。
func Info(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Site.FindById(form.Id)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站详情", "查看网站["+data.Name+"]详情")
		c.SuccessWithData("获取成功", data)
	}
}
