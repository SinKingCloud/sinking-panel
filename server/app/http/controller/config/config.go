package config

import (
	"github.com/SinKingCloud/sinking-go/sinking-web"
	"server/app/service"
	"server/app/util/server"
)

// Get 获取配置
func Get(c *server.Context) {
	type Form struct {
		Group string `form:"group" json:"group" default:"" validate:"required,max=100" label:"组ID"`
		Key   string `form:"key" json:"key" default:"" validate:"omitempty,max=100" label:"配置标识"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	if form.Key != "" {
		c.SuccessWithData("获取数据成功", sinking_web.H{
			form.Key: service.Config.Get(form.Group, form.Key),
		})
	} else {
		c.SuccessWithData("获取数据成功", service.Config.Group(form.Group))
	}
}

// Set 修改配置
func Set(c *server.Context) {
	type Config struct {
		Key   string `json:"key" default:"" validate:"required,max=100" label:"配置标识"`
		Value string `json:"value" default:"" validate:"omitempty" label:"配置内容"`
	}
	type Form struct {
		Configs []*Config `json:"configs" default:"" validate:"gte=1" label:"配置标识"`
		Group   string    `json:"group" default:"" validate:"required,max=100" label:"配置组"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	for _, v := range form.Configs {
		if v.Value != "" {
			_ = service.Config.Set(form.Group, v.Key, v.Value)
		}
	}
	c.Success("修改数据成功")
}
