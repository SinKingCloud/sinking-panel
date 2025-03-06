package server

import (
	"server/app/service"
	"server/app/util/server"
)

// Update 修改信息
func Update(c *server.Context) {
	type Form struct {
		Ids      []int  `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
		Ip       string `json:"ip" default:"" validate:"omitempty,ip" label:"IP地址"`
		Port     string `json:"title" default:"" validate:"omitempty,min=1,max=65535" label:"端口"`
		User     string `json:"user" default:"" validate:"omitempty" label:"账号"`
		AuthType string `json:"auth_type" default:"" validate:"omitempty,oneof=0 1" label:"验证方式"`
		Password string `json:"password" default:"" validate:"omitempty" label:"密码"`
		Name     string `json:"name" default:"" validate:"omitempty" label:"名称"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	data := make(map[string]interface{})
	if form.Ip != "" {
		data["ip"] = form.Ip
	}
	if form.Port != "" {
		data["port"] = form.Port
	}
	if form.User != "" {
		data["user"] = form.User
	}
	if form.AuthType != "" {
		data["auth_type"] = form.AuthType
	}
	if form.Password != "" {
		data["password"] = form.Password
	}
	err := service.Server.UpdateByIds(form.Ids, data)
	if err == nil {
		c.Success("修改成功")
	} else {
		c.Error("修改失败")
	}
}
