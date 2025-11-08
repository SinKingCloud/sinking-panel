package task

import (
	"server/app/service"
	"server/app/util/context"
)

// Info 获取详情
func Info(c *context.Context) {
	type Form struct {
		Id int `json:"id" default:"" validate:"numeric,min=1" label:"记录ID"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Task.FindById(form.Id)
	if err == nil && data != nil {
		c.SuccessWithData("获取成功", data)
	} else {
		c.Error("获取失败")
	}
}
