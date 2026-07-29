package server

import (
	repositoryServer "server/app/repository/server"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Update 修改信息
func Update(c *context.Context) {
	var form struct {
		Ids      []int64 `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
		Ip       string  `json:"ip" default:"" validate:"omitempty,ip" label:"IP地址"`
		Port     string  `json:"port" default:"" validate:"omitempty,numeric" label:"端口"`
		User     string  `json:"user" default:"" validate:"omitempty" label:"账号"`
		AuthType string  `json:"auth_type" default:"" validate:"omitempty,oneof=0 1" label:"验证方式"`
		Password string  `json:"password" default:"" validate:"omitempty" label:"密码"`
		Name     string  `json:"name" default:"" validate:"omitempty" label:"名称"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &repositoryServer.UpdateServer{}
	if form.Ip != "" {
		data.Ip = form.Ip
	}
	if form.Port != "" {
		port, _ := strconv.Atoi(form.Port)
		if port < 1 || port > 65535 {
			c.Error("端口参数不合法")
			return
		}
		data.Port = port
	}
	if form.User != "" {
		data.User = form.User
	}
	if form.AuthType != "" {
		data.AuthType, _ = strconv.Atoi(form.AuthType)
	}
	if form.Password != "" {
		data.Password = form.Password
	}
	if form.Name != "" {
		data.Name = form.Name
	}
	err := service.Server.UpdateByIds(form.Ids, data)
	if err == nil {
		c.Success("修改成功")
	} else {
		c.Error("修改失败")
	}
}
