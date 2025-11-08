package task

import (
	"server/app/model"
	"server/app/service"
	"server/app/util/context"
)

// Create 添加信息
func Create(c *context.Context) {
	type Form struct {
		Name   string `json:"name" default:"" validate:"required" label:"任务名称"`
		Type   int    `json:"type" default:"0" validate:"required,oneof=0" label:"任务类型"`
		Spec   string `json:"spec" default:"" validate:"required" label:"任务表达式"`
		Script string `json:"script" default:"" validate:"required" label:"任务内容"`
		Status int    `json:"status" default:"0" validate:"required,oneof=0 1" label:"任务状态"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	if service.Task.ValidateCron(form.Spec) {
		c.Error("任务表达式不正确")
		return
	}
	err := service.Task.Add(&model.Task{
		Type:   form.Type,
		Name:   form.Name,
		Spec:   form.Spec,
		Script: form.Script,
		Status: form.Status,
	})
	if err == nil {
		c.Success("添加成功")
	} else {
		c.Error("添加失败")
	}
}
