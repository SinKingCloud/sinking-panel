package container

import (
	"server/app/enum/container_status"
	"server/app/enum/log_type"
	"server/app/service"
	serviceContainer "server/app/service/container"
	"server/app/util/container"
	"server/app/util/context"
	"time"
)

// List 筛选并分页查询容器摘要。
func List(c *context.Context) {
	query := c.ValidatePage("create_time", "desc", "id,name,create_time")
	var form struct {
		Id              string `json:"id" default:"" validate:"omitempty,max=64" label:"容器ID"`
		Keyword         string `json:"keyword" default:"" validate:"omitempty,max=100" label:"关键词"`
		Name            string `json:"name" default:"" validate:"omitempty,max=100" label:"容器名称"`
		ImageId         string `json:"image_id" default:"" validate:"omitempty,max=128" label:"镜像ID"`
		ImageName       string `json:"image_name" default:"" validate:"omitempty,max=256" label:"镜像名称"`
		Status          string `json:"status" default:"" validate:"omitempty" label:"运行状态"`
		AutoRestart     *bool  `json:"auto_restart" validate:"omitempty" label:"自动重启"`
		ManualStop      *bool  `json:"manual_stop" validate:"omitempty" label:"手动停止"`
		CreateTimeStart string `json:"create_time_start" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"创建起始时间"`
		CreateTimeEnd   string `json:"create_time_end" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"创建结束时间"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	where := &serviceContainer.SelectContainer{
		AutoRestart: form.AutoRestart, ManualStop: form.ManualStop,
	}
	if form.CreateTimeStart != "" {
		value, err := time.ParseInLocation(time.DateTime, form.CreateTimeStart, time.Local)
		if err != nil {
			c.Error("创建起始时间格式错误")
			return
		}
		timestamp := value.Unix()
		where.CreateTimeStart = &timestamp
	}
	if form.CreateTimeEnd != "" {
		value, err := time.ParseInLocation(time.DateTime, form.CreateTimeEnd, time.Local)
		if err != nil {
			c.Error("创建结束时间格式错误")
			return
		}
		timestamp := value.Unix()
		where.CreateTimeEnd = &timestamp
	}
	if form.Id != "" {
		where.ID = &form.Id
	}
	if form.Keyword != "" {
		where.Keyword = &form.Keyword
	}
	if form.Name != "" {
		where.Name = &form.Name
	}
	if form.ImageId != "" {
		where.ImageID = &form.ImageId
	}
	if form.ImageName != "" {
		where.ImageName = &form.ImageName
	}
	if form.Status != "" {
		status := container.State(form.Status)
		if _, ok := container_status.Map()[status]; !ok {
			c.Error("运行状态无效")
			return
		}
		where.Status = &status
	}
	data, err := service.Container.List(where, query)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看容器列表", "查看容器列表数据")
		c.SuccessWithData("获取成功", data)
	}
}
