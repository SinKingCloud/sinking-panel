package task

import (
	"server/app/enum/log_type"
	repositoryTask "server/app/repository/task"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Update 修改信息
func Update(c *context.Context) {
	var form struct {
		Ids      []int64 `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
		TypeId   string  `json:"type_id" default:"" validate:"omitempty,numeric" label:"任务分类ID"`
		ExecType string  `json:"exec_type" default:"" validate:"omitempty,numeric" label:"任务执行类型"`
		Name     string  `json:"name" default:"" validate:"omitempty" label:"任务名称"`
		Spec     string  `json:"spec" default:"" validate:"omitempty" label:"任务表达式"`
		Script   string  `json:"script" default:"" validate:"omitempty" label:"任务内容"`
		Status   string  `json:"status" default:"" validate:"omitempty,numeric" label:"任务状态"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &repositoryTask.UpdateTask{}
	if form.TypeId != "" {
		typeId, err := strconv.ParseInt(form.TypeId, 10, 64)
		if err != nil || typeId < 0 {
			c.Error("任务分类ID参数错误")
			return
		}
		data.TypeId = &typeId
	}
	if form.ExecType != "" {
		execType, err := strconv.Atoi(form.ExecType)
		if err != nil {
			c.Error("任务执行类型参数错误")
			return
		}
		data.ExecType = &execType
	}
	if form.Name != "" {
		data.Name = &form.Name
	}
	if form.Spec != "" {
		data.Spec = &form.Spec
	}
	if form.Script != "" {
		data.Script = &form.Script
	}
	if form.Status != "" {
		status, err := strconv.Atoi(form.Status)
		if err != nil {
			c.Error("任务状态参数错误")
			return
		}
		data.Status = &status
	}
	err := service.Task.UpdateByIds(form.Ids, data)
	if err == nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改计划任务", "修改计划任务数据")
		c.Success("修改成功")
	} else {
		c.Error(err.Error())
	}
}
