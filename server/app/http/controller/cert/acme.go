package cert

import (
	stdContext "context"
	"net/http"
	"server/app/enum/log_type"
	"server/app/service"
	serviceSite "server/app/service/site"
	"server/app/util/context"
	webServer "server/app/util/server"
	"strconv"
	"time"
)

const certificateRequestTimeout = 20 * time.Minute

// Obtain 通过 ACME 申请证书。
func Obtain(c *context.Context) {
	var form struct {
		Name           string                   `json:"name" default:"" validate:"required,max=100" label:"证书名称"`
		Domain         string                   `json:"domain" default:"" validate:"required,max=253" label:"证书域名"`
		Email          string                   `json:"email" default:"" validate:"omitempty,email,max=254" label:"ACME邮箱"`
		CA             string                   `json:"ca" default:"" validate:"omitempty,oneof=production staging" label:"签发环境"`
		Challenge      string                   `json:"challenge" default:"http" validate:"required,oneof=http dns" label:"验证方式"`
		DNSProvider    string                   `json:"dns_provider" default:"" validate:"omitempty,oneof=alidns dnspod tencentcloud huaweicloud volcengine baiducloud" label:"DNS服务商"`
		DNSCredentials webServer.DNSCredentials `json:"dns_credentials" label:"DNS验证凭据"`
		SecretId       *int64                   `json:"secret_id" validate:"omitempty,min=0" label:"密钥ID"`
		AutoRenew      *int                     `json:"auto_renew" validate:"omitempty,oneof=0 1" label:"自动续签"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	request := serviceSite.CertificateRequest{
		SecretId:  form.SecretId,
		AutoRenew: form.AutoRenew,
		CertificateRequest: webServer.CertificateRequest{
			Domain:         form.Domain,
			Email:          form.Email,
			CA:             webServer.CertificateCA(form.CA),
			Challenge:      webServer.CertificateChallenge(form.Challenge),
			DNSProvider:    webServer.DNSProvider(form.DNSProvider),
			DNSCredentials: form.DNSCredentials,
		},
	}
	// DNS 传播可能耗时数分钟，单独延长证书请求的超时。
	_ = http.NewResponseController(c.Writer).SetWriteDeadline(time.Now().Add(certificateRequestTimeout + time.Minute))
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), certificateRequestTimeout)
	defer cancel()
	data, err := service.Site.ObtainCert(ctx, form.Name, request)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "申请网站证书", "申请网站证书["+data.Name+"]")
		c.SuccessWithData("申请成功", data)
	}
}

// Renew 续签 ACME 证书。
func Renew(c *context.Context) {
	var form struct {
		Id             int64                    `json:"id" default:"0" validate:"required,min=1" label:"证书ID"`
		Domain         string                   `json:"domain" default:"" validate:"required,max=253" label:"续签域名"`
		Email          string                   `json:"email" default:"" validate:"omitempty,email,max=254" label:"ACME邮箱"`
		CA             string                   `json:"ca" default:"" validate:"omitempty,oneof=production staging" label:"签发环境"`
		Challenge      string                   `json:"challenge" validate:"omitempty,oneof=http dns" label:"验证方式"`
		DNSProvider    string                   `json:"dns_provider" default:"" validate:"omitempty,oneof=alidns dnspod tencentcloud huaweicloud volcengine baiducloud" label:"DNS服务商"`
		DNSCredentials webServer.DNSCredentials `json:"dns_credentials" label:"DNS验证凭据"`
		SecretId       *int64                   `json:"secret_id" validate:"omitempty,min=0" label:"密钥ID"`
		AutoRenew      *int                     `json:"auto_renew" validate:"omitempty,oneof=0 1" label:"自动续签"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	request := serviceSite.CertificateRequest{
		SecretId:  form.SecretId,
		AutoRenew: form.AutoRenew,
		CertificateRequest: webServer.CertificateRequest{
			Domain:         form.Domain,
			Email:          form.Email,
			CA:             webServer.CertificateCA(form.CA),
			Challenge:      webServer.CertificateChallenge(form.Challenge),
			DNSProvider:    webServer.DNSProvider(form.DNSProvider),
			DNSCredentials: form.DNSCredentials,
		},
	}
	_ = http.NewResponseController(c.Writer).SetWriteDeadline(time.Now().Add(certificateRequestTimeout + time.Minute))
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), certificateRequestTimeout)
	defer cancel()
	data, err := service.Site.RenewCert(ctx, form.Id, request)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "续签网站证书", "续签网站证书["+strconv.FormatInt(form.Id, 10)+"]")
		c.SuccessWithData("续签成功", data)
	}
}
