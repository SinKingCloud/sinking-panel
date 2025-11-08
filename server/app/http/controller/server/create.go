package server

import (
	"server/app/model"
	"server/app/service"
	"server/app/util/context"
)

// Create 添加信息
func Create(c *context.Context) {
	type Form struct {
		Ip       string `json:"ip" default:"" validate:"required,ip" label:"IP地址"`
		Port     int    `json:"title" default:"" validate:"required,min=1,max=65535" label:"端口"`
		User     string `json:"user" default:"" validate:"required" label:"账号"`
		AuthType int    `json:"auth_type" default:"" validate:"required,oneof=0 1" label:"验证方式"`
		Password string `json:"password" default:"" validate:"required" label:"密码"`
		Name     string `json:"name" default:"" validate:"required" label:"名称"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	err := service.Server.Create(&model.Server{
		Ip:       form.Ip,
		Port:     form.Port,
		User:     form.User,
		AuthType: form.AuthType,
		Password: form.Password,
		Name:     form.Name,
	})
	if err == nil {
		c.Success("添加成功")
	} else {
		c.Error("添加失败")
	}
}
