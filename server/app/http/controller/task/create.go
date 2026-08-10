package task

import (
	"server/app/enum/log_type"
	"server/app/model"
	"server/app/service"
	"server/app/util/context"
)

// Create 添加信息
func Create(c *context.Context) {
	var form struct {
		TypeId   int64  `json:"type_id" default:"0" validate:"min=0" label:"任务分类ID"`
		ExecType int    `json:"exec_type" default:"0" validate:"numeric" label:"任务执行类型"`
		Name     string `json:"name" default:"" validate:"required" label:"任务名称"`
		Spec     string `json:"spec" default:"" validate:"required" label:"任务表达式"`
		Script   string `json:"script" default:"" validate:"required" label:"任务内容"`
		Status   int    `json:"status" default:"0" validate:"numeric" label:"任务状态"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	err := service.Task.Add(&model.Task{
		TypeId:   form.TypeId,
		ExecType: form.ExecType,
		Name:     form.Name,
		Spec:     form.Spec,
		Script:   form.Script,
		Status:   form.Status,
	})
	if err == nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "创建计划任务", "创建计划任务["+form.Name+"]")
		c.Success("添加成功")
	} else {
		c.Error(err.Error())
	}
}
