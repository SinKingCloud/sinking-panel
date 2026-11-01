package oss

import (
	"context"
	"server/app/service"
	appContext "server/app/util/context"
	"time"
)

// Test 使用当前填写的密钥和连接配置测试，不创建或修改存储记录。
func Test(c *appContext.Context) {
	var form struct {
		SecretId int64  `json:"secret_id" validate:"required,min=1" label:"密钥ID"`
		Config   string `json:"config" validate:"required,max=1048576" label:"连接配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	if err := service.Oss.Test(ctx, form.SecretId, form.Config); err != nil {
		c.Error(err.Error())
		return
	}
	c.Success("连接成功")
}
