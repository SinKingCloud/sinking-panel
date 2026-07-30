package server

import (
	"server/app/enum/log_type"
	repositoryServer "server/app/repository/server"
	"server/app/service"
	"server/app/util/context"
)

// Update 修改信息
func Update(c *context.Context) {
	var form struct {
		Ids      []int64 `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
		Ip       string  `json:"ip" default:"" validate:"omitempty,ip" label:"IP地址"`
		Port     *int    `json:"port" default:"" validate:"omitempty,min=1,max=65535" label:"端口"`
		User     string  `json:"user" default:"" validate:"omitempty" label:"账号"`
		AuthType *int    `json:"auth_type" default:"" validate:"omitempty,oneof=0 1" label:"验证方式"`
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
	if form.Port != nil {
		data.Port = *form.Port
	}
	if form.User != "" {
		data.User = form.User
	}
	if form.AuthType != nil {
		data.AuthType = *form.AuthType
	}
	if form.Password != "" {
		data.Password = form.Password
	}
	if form.Name != "" {
		data.Name = form.Name
	}
	err := service.Server.UpdateByIds(form.Ids, data)
	if err == nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改服务器", "修改服务器数据")
		c.Success("修改成功")
	} else {
		c.Error("修改失败")
	}
}
