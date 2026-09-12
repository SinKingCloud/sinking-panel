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
		Id          int64  `json:"id" default:"0" validate:"required,min=1" label:"证书ID"`
		Name        string `json:"name" default:"" validate:"omitempty,max=100" label:"证书名称"`
		Certificate string `json:"certificate" default:"" validate:"omitempty,max=4194304" label:"证书内容"`
		PrivateKey  string `json:"private_key" default:"" validate:"omitempty,max=1048576" label:"证书私钥"`
		Challenge   string `json:"challenge" default:"" validate:"omitempty,oneof=http dns" label:"验证方式"`
		SecretId    string `json:"secret_id" default:"" validate:"omitempty,numeric" label:"密钥ID"`
		AutoRenew   string `json:"auto_renew" default:"" validate:"omitempty,oneof=0 1" label:"自动续签"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &repositoryCert.UpdateCert{}
	if form.Name != "" {
		data.Name = &form.Name
	}
	if form.Certificate != "" {
		data.Certificate = &form.Certificate
	}
	if form.PrivateKey != "" {
		data.PrivateKey = &form.PrivateKey
	}
	if form.Challenge != "" {
		data.Challenge = &form.Challenge
	}
	if form.SecretId != "" {
		secretId, err := strconv.ParseInt(form.SecretId, 10, 64)
		if err != nil || secretId < 0 {
			c.Error("密钥ID参数错误")
			return
		}
		data.SecretId = &secretId
	}
	if form.AutoRenew != "" {
		autoRenew, err := strconv.Atoi(form.AutoRenew)
		if err != nil || (autoRenew != 0 && autoRenew != 1) {
			c.Error("自动续签参数错误")
			return
		}
		data.AutoRenew = &autoRenew
	}
	err := service.Site.UpdateCert(form.Id, data)
	if err == nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站证书", "修改网站证书["+strconv.FormatInt(form.Id, 10)+"]")
		c.Success("修改成功")
	} else {
		c.Error(err.Error())
	}
}
