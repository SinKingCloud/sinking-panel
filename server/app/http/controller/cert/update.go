package cert

import (
	"server/app/enum/log_type"
	repositoryCert "server/app/repository/cert"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Update 更新证书，未传入的字段保持原值。
func Update(c *context.Context) {
	var form struct {
		Id          int64   `json:"id" default:"0" validate:"required,min=1" label:"证书ID"`
		Name        *string `json:"name" validate:"omitempty,max=100" label:"证书名称"`
		Certificate *string `json:"certificate" validate:"omitempty,max=4194304" label:"证书内容"`
		PrivateKey  *string `json:"private_key" validate:"omitempty,max=1048576" label:"证书私钥"`
		Challenge   *string `json:"challenge" validate:"omitempty,oneof=http dns" label:"验证方式"`
		SecretId    *int64  `json:"secret_id" validate:"omitempty,min=0" label:"密钥ID"`
		AutoRenew   *int    `json:"auto_renew" validate:"omitempty,oneof=0 1" label:"自动续签"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &repositoryCert.UpdateCert{}
	if form.Name != nil {
		data.Name = form.Name
	}
	if form.Certificate != nil {
		data.Certificate = form.Certificate
	}
	if form.PrivateKey != nil {
		data.PrivateKey = form.PrivateKey
	}
	if form.Challenge != nil {
		data.Challenge = form.Challenge
	}
	if form.SecretId != nil {
		data.SecretId = form.SecretId
	}
	if form.AutoRenew != nil {
		data.AutoRenew = form.AutoRenew
	}
	err := service.Site.UpdateCert(form.Id, data)
	if err == nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站证书", "修改网站证书["+strconv.FormatInt(form.Id, 10)+"]")
		c.Success("修改成功")
	} else {
		c.Error(err.Error())
	}
}
