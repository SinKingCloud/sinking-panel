package server

import (
	"server/app/enum/log_type"
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
		data.Ip = &form.Ip
	}
	if form.Port != "" {
		port, err := strconv.Atoi(form.Port)
		if err != nil || port < 1 || port > 65535 {
			c.Error("端口参数错误")
			return
		}
		data.Port = &port
	}
	if form.User != "" {
		data.User = &form.User
	}
	if form.AuthType != "" {
		authType, err := strconv.Atoi(form.AuthType)
		if err != nil {
			c.Error("验证方式参数错误")
			return
		}
		data.AuthType = &authType
	}
	if form.Password != "" {
		data.Password = &form.Password
	}
	if form.Name != "" {
		data.Name = &form.Name
	}
	err := service.Server.UpdateByIds(form.Ids, data)
	if err == nil {
		title := "修改服务器"
		content := "修改服务器数据"
		if len(form.Ids) == 1 && form.Ids[0] == 0 {
			title = "修改本机连接"
			content = "修改本机SSH连接配置"
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, title, content)
		c.Success("修改成功")
	} else {
		c.Error("修改失败")
	}
}
