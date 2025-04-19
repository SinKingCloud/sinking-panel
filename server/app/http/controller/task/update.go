package task

import (
	"server/app/service"
	"server/app/util/server"
)

// Update 修改信息
func Update(c *server.Context) {
	type Form struct {
		Ids    []int  `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
		Name   string `json:"name" default:"" validate:"omitempty" label:"任务名称"`
		Type   string `json:"type" default:"" validate:"omitempty,oneof=0" label:"任务类型"`
		Spec   string `json:"spec" default:"" validate:"omitempty" label:"任务表达式"`
		Script string `json:"script" default:"" validate:"omitempty" label:"任务内容"`
		Status string `json:"status" default:"" validate:"omitempty,oneof=0 1" label:"任务状态"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	data := make(map[string]interface{})
	if form.Name != "" {
		data["name"] = form.Name
	}
	if form.Type != "" {
		data["type"] = form.Type
	}
	if form.Spec != "" {
		if service.Task.ValidateCron(form.Spec) {
			c.Error("任务表达式不正确")
			return
		}
		data["spec"] = form.Spec
	}
	if form.Script != "" {
		data["script"] = form.Script
	}
	if form.Status != "" {
		data["status"] = form.Status
	}
	err := service.Task.UpdateByIds(form.Ids, data)
	if err == nil {
		c.Success("修改成功")
	} else {
		c.Error("修改失败")
	}
}
