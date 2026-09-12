package task

import (
	"server/app/enum/log_type"
	repositoryTask "server/app/repository/task"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// List 获取计划任务列表
func List(c *context.Context) {
	query := c.ValidatePage("id", "desc", "id", "id,type_id,exec_type,status,run_time,create_time,update_time")
	var form struct {
		Name            string `json:"name" default:"" validate:"omitempty" label:"任务名称"`
		TypeId          string `json:"type_id" default:"" validate:"omitempty,numeric" label:"任务分类ID"`
		ExecType        string `json:"exec_type" default:"" validate:"omitempty,numeric" label:"任务执行类型"`
		Status          string `json:"status" default:"" validate:"omitempty,numeric" label:"任务状态"`
		RunTimeStart    string `json:"run_time_start" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"运行起始时间"`
		RunTimeEnd      string `json:"run_time_end" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"运行结束时间"`
		CreateTimeStart string `json:"create_time_start" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"创建起始时间"`
		CreateTimeEnd   string `json:"create_time_end" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"创建结束时间"`
		UpdateTimeStart string `json:"update_time_start" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"更新起始时间"`
		UpdateTimeEnd   string `json:"update_time_end" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"更新结束时间"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	where := &repositoryTask.SelectTask{}
	if form.TypeId != "" {
		typeId, err := strconv.ParseInt(form.TypeId, 10, 64)
		if err != nil || typeId < 0 {
			c.Error("任务分类参数错误")
			return
		}
		if typeId != 0 {
			where.TypeId = &typeId
		}
	}
	if form.ExecType != "" {
		execType, err := strconv.Atoi(form.ExecType)
		if err != nil {
			c.Error("任务执行类型参数错误")
			return
		}
		where.ExecType = &execType
	}
	if form.Name != "" {
		where.Name = &form.Name
	}
	if form.Status != "" {
		status, err := strconv.Atoi(form.Status)
		if err != nil {
			c.Error("任务状态参数错误")
			return
		}
		where.Status = &status
	}
	if form.RunTimeStart != "" {
		where.RunTimeStart = &form.RunTimeStart
	}
	if form.RunTimeEnd != "" {
		where.RunTimeEnd = &form.RunTimeEnd
	}
	if form.CreateTimeStart != "" {
		where.CreateTimeStart = &form.CreateTimeStart
	}
	if form.CreateTimeEnd != "" {
		where.CreateTimeEnd = &form.CreateTimeEnd
	}
	if form.UpdateTimeStart != "" {
		where.UpdateTimeStart = &form.UpdateTimeStart
	}
	if form.UpdateTimeEnd != "" {
		where.UpdateTimeEnd = &form.UpdateTimeEnd
	}
	data, err := service.Task.Select(where, query)
	if err != nil {
		c.Error("获取失败")
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看计划任务", "查看计划任务列表")
		c.SuccessWithData("获取成功", data)
	}
}
