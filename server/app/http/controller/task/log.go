package task

import (
	"server/app/service"
	"server/app/util/context"
)

func Log(c *context.Context) {
	var form struct {
		Id       int64 `json:"id" default:"" validate:"required,numeric,min=1" label:"记录ID"`
		Page     int   `json:"page" default:"1" validate:"numeric,min=1" label:"分页页码"`
		PageSize int   `json:"page_size" default:"20" validate:"numeric,min=1,max=1000" label:"分页容量"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Task.FindById(form.Id)
	if err != nil || data == nil {
		c.Error("获取失败")
	} else {
		c.SuccessWithData("获取成功", service.Task.ReadLog(data.Id, form.Page, form.PageSize))
	}
}
