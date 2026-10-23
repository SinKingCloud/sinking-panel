package database

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Test 测试尚未保存的实例连接配置。
func Test(c *context.Context) {
	var params struct {
		Id     int64  `json:"id" default:"0" validate:"min=0" label:"实例ID"`
		Type   int    `json:"type" default:"0" validate:"required" label:"数据库类型"`
		Config string `json:"config" default:"" validate:"required,max=1048576" label:"连接配置"`
	}
	if ok, msg := c.ValidatorAll(&params); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Database.Test(c.Request.Context(), params.Id, params.Type, params.Config)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "测试数据库连接", "测试数据库连接参数")
	c.SuccessWithData("连接成功", data)
}
