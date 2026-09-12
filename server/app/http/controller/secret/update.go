package secret

import (
	"server/app/enum/log_type"
	repositorySecret "server/app/repository/secret"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Update 修改密钥凭据。
func Update(c *context.Context) {
	var form struct {
		Id       int64  `json:"id" default:"0" validate:"required,min=1" label:"密钥ID"`
		Name     string `json:"name" default:"" validate:"omitempty,max=100" label:"密钥名称"`
		Provider string `json:"provider" default:"" validate:"omitempty,numeric" label:"服务商"`
		Data     string `json:"data" default:"" validate:"omitempty,max=1048576" label:"密钥内容"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &repositorySecret.UpdateSecret{}
	if form.Name != "" {
		data.Name = &form.Name
	}
	if form.Provider != "" {
		provider, err := strconv.Atoi(form.Provider)
		if err != nil || provider < 0 {
			c.Error("服务商参数错误")
			return
		}
		data.Provider = &provider
	}
	if form.Data != "" {
		data.Data = &form.Data
	}
	err := service.Secret.Update(form.Id, data)
	if err == nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改密钥", "修改密钥["+strconv.FormatInt(form.Id, 10)+"]")
		c.Success("修改成功")
	} else {
		c.Error(err.Error())
	}
}
