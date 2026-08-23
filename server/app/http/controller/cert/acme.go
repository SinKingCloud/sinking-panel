package cert

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	webServer "server/app/util/server"
)

// Obtain 通过 ACME 申请证书。
func Obtain(c *context.Context) {
	var form struct {
		Name   string `json:"name" default:"" validate:"required,max=100" label:"证书名称"`
		Domain string `json:"domain" default:"" validate:"required,max=253" label:"证书域名"`
		Email  string `json:"email" default:"" validate:"omitempty,email,max=254" label:"ACME邮箱"`
		CA     string `json:"ca" default:"" validate:"omitempty,oneof=production staging" label:"签发环境"`
	}
	if ok, message := c.ValidatorAll(&form); !ok {
		c.Error(message)
		return
	}
	result, err := service.Site.ObtainCert(c.Request.Context(), form.Name, webServer.CertificateRequest{
		Domain: form.Domain,
		Email:  form.Email,
		CA:     webServer.CertificateCA(form.CA),
	})
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "申请网站证书", "申请网站证书["+result.Name+"]")
	c.SuccessWithData("申请成功", result)
}

// Renew 续签 ACME 证书。
func Renew(c *context.Context) {
	var form struct {
		Id     int64  `json:"id" default:"0" validate:"required,min=1" label:"证书ID"`
		Domain string `json:"domain" default:"" validate:"required,max=253" label:"续签域名"`
		Email  string `json:"email" default:"" validate:"omitempty,email,max=254" label:"ACME邮箱"`
		CA     string `json:"ca" default:"" validate:"omitempty,oneof=production staging" label:"签发环境"`
	}
	if ok, message := c.ValidatorAll(&form); !ok {
		c.Error(message)
		return
	}
	result, err := service.Site.RenewCert(c.Request.Context(), form.Id, webServer.CertificateRequest{
		Domain: form.Domain,
		Email:  form.Email,
		CA:     webServer.CertificateCA(form.CA),
	})
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "续签网站证书", "续签网站证书["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("续签成功", result)
}
