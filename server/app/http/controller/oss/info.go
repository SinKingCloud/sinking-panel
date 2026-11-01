package oss

import (
	"server/app/service"
	"server/app/util/context"
)

// Info 获取对象存储详情。
func Info(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"存储ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Oss.FindById(form.Id)
	if err != nil {
		c.Error(err.Error())
		return
	}
	c.SuccessWithData("获取成功", data)
}
