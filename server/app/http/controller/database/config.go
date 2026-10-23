package database

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Config 返回数据库类型的默认连接配置。
func Config(c *context.Context) {
	var form struct {
		Type int `json:"type" default:"0" validate:"required" label:"数据库类型"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Database.Config(form.Type, "", 0, "", "")
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看连接配置", "查看数据库默认连接配置")
	c.SuccessWithData("获取成功", data)
}
