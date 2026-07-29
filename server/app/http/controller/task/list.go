package task

import (
	repositoryTask "server/app/repository/task"
	"server/app/service"
	"server/app/util/context"
)

// List 获取计划任务列表
func List(c *context.Context) {
	query := c.ValidatePage("id", "desc", "id", "id,run_time,create_time,update_time")
	var form struct {
		Name            string `json:"name" default:"" validate:"omitempty" label:"任务名称"`
		Type            string `json:"type" default:"" validate:"omitempty,numeric" label:"任务类型"`
		Status          string `json:"status" default:"" validate:"omitempty,oneof=0 1" label:"任务状态"`
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
	if form.Type != "" {
		where.Type = form.Type
	}
	if form.Name != "" {
		where.Name = form.Name
	}
	if form.Status != "" {
		where.Status = form.Status
	}
	if form.RunTimeStart != "" {
		where.RunTimeStart = form.RunTimeStart
	}
	if form.RunTimeEnd != "" {
		where.RunTimeEnd = form.RunTimeEnd
	}
	if form.CreateTimeStart != "" {
		where.CreateTimeStart = form.CreateTimeStart
	}
	if form.CreateTimeEnd != "" {
		where.CreateTimeEnd = form.CreateTimeEnd
	}
	if form.UpdateTimeStart != "" {
		where.UpdateTimeStart = form.UpdateTimeStart
	}
	if form.UpdateTimeEnd != "" {
		where.UpdateTimeEnd = form.UpdateTimeEnd
	}
	data, err := service.Task.Select(where, query)
	if err != nil {
		c.Error("获取失败")
	} else {
		c.SuccessWithData("获取成功", data)
	}
}
