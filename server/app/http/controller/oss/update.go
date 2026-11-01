package oss

import (
	"server/app/enum/log_type"
	repositoryOss "server/app/repository/oss"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Update 修改 OSS 存储配置，不搬迁存储桶中的文件。
func Update(c *context.Context) {
	var form struct {
		Id     int64  `json:"id" default:"0" validate:"required,min=1" label:"存储ID"`
		Name   string `json:"name" default:"" validate:"omitempty,max=100" label:"存储名称"`
		Config string `json:"config" default:"" validate:"omitempty,max=1048576" label:"连接配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &repositoryOss.UpdateOss{}
	if form.Name != "" {
		data.Name = &form.Name
	}
	if form.Config != "" {
		data.Config = &form.Config
	}
	if err := service.Oss.Update(form.Id, data); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改 OSS 存储", "修改 OSS 存储["+strconv.FormatInt(form.Id, 10)+"]")
	c.Success("修改成功")
}
