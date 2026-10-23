package database

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Info 数据库实例详情。
func Info(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Database.FindById(form.Id)
	if err != nil {
		c.Error(err.Error())
		return
	}
	if data == nil {
		c.Error("实例不存在")
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看数据库实例详情", "查看数据库实例["+strconv.FormatInt(form.Id, 10)+"]详情")
	c.SuccessWithData("获取成功", data)
}
