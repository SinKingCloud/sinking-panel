package secret

import (
	"server/app/enum/log_type"
	"server/app/model"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Create 添加密钥凭据。
func Create(c *context.Context) {
	var form struct {
		Name     string `json:"name" default:"" validate:"required,max=100" label:"密钥名称"`
		Provider int    `json:"provider" default:"0" validate:"min=0" label:"服务商"`
		Data     string `json:"data" default:"" validate:"required,max=1048576" label:"密钥内容"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &model.Secret{
		Name:     form.Name,
		Provider: form.Provider,
		Data:     form.Data,
	}
	err := service.Secret.Create(data)
	if err == nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "添加密钥", "添加密钥["+strconv.FormatInt(data.Id, 10)+"]")
		c.SuccessWithData("添加成功", map[string]int64{"id": data.Id})
	} else {
		c.Error(err.Error())
	}
}
