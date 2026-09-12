package secret

import (
	"server/app/enum/log_type"
	repositorySecret "server/app/repository/secret"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Update 更新密钥凭据，未传入的字段保持原值。
func Update(c *context.Context) {
	var form struct {
		Id       int64   `json:"id" default:"0" validate:"required,min=1" label:"密钥ID"`
		Name     *string `json:"name" validate:"omitempty,min=1,max=100" label:"密钥名称"`
		Provider *int    `json:"provider" validate:"omitempty,min=0" label:"服务商"`
		Data     *string `json:"data" validate:"omitempty,min=1,max=1048576" label:"密钥内容"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &repositorySecret.UpdateSecret{}
	if form.Name != nil {
		data.Name = form.Name
	}
	if form.Provider != nil {
		data.Provider = form.Provider
	}
	if form.Data != nil {
		data.Data = form.Data
	}
	err := service.Secret.Update(form.Id, data)
	if err == nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改密钥", "修改密钥["+strconv.FormatInt(form.Id, 10)+"]")
		c.Success("修改成功")
	} else {
		c.Error(err.Error())
	}
}
