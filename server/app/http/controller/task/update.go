package task

import (
	"server/app/enum/log_type"
	repositoryTask "server/app/repository/task"
	"server/app/service"
	"server/app/util/context"
)

// Update 修改信息
func Update(c *context.Context) {
	var form struct {
		Ids      []int64 `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
		TypeId   *int64  `json:"type_id" default:"" validate:"omitempty,min=0" label:"任务分类ID"`
		ExecType *int    `json:"exec_type" default:"" validate:"omitempty,numeric" label:"任务执行类型"`
		Name     string  `json:"name" default:"" validate:"omitempty" label:"任务名称"`
		Spec     string  `json:"spec" default:"" validate:"omitempty" label:"任务表达式"`
		Script   string  `json:"script" default:"" validate:"omitempty" label:"任务内容"`
		Status   *int    `json:"status" default:"" validate:"omitempty,numeric" label:"任务状态"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &repositoryTask.UpdateTask{}
	if form.TypeId != nil {
		data.TypeId = *form.TypeId
	}
	if form.ExecType != nil {
		data.ExecType = *form.ExecType
	}
	if form.Name != "" {
		data.Name = form.Name
	}
	if form.Spec != "" {
		data.Spec = form.Spec
	}
	if form.Script != "" {
		data.Script = form.Script
	}
	if form.Status != nil {
		data.Status = *form.Status
	}
	err := service.Task.UpdateByIds(form.Ids, data)
	if err == nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改计划任务", "修改计划任务数据")
		c.Success("修改成功")
	} else {
		c.Error(err.Error())
	}
}
