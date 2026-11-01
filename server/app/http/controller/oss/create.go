package oss

import (
	"server/app/enum/log_type"
	"server/app/model"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Create 保存名称和关联密钥，连接配置在创建后填写。
func Create(c *context.Context) {
	var form struct {
		Name     string `json:"name" validate:"required,max=100" label:"存储名称"`
		SecretId int64  `json:"secret_id" validate:"required,min=1" label:"密钥ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &model.Oss{Name: form.Name, SecretId: form.SecretId}
	if err := service.Oss.Create(data); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "添加 OSS 存储", "添加 OSS 存储["+strconv.FormatInt(data.Id, 10)+"]")
	c.SuccessWithData("添加成功", map[string]int64{"id": data.Id})
}
