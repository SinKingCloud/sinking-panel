package cert

import (
	stdContext "context"
	"net/http"
	"server/app/enum/cert_type"
	"server/app/enum/log_type"
	"server/app/model"
	"server/app/service"
	serviceSite "server/app/service/site"
	"server/app/util/context"
	webServer "server/app/util/server"
	"strconv"
	"time"
)

// Obtain 通过 ACME 申请证书，type 控制自动或手动验证。
func Obtain(c *context.Context) {
	var form struct {
		Type      int    `json:"type" default:"2" validate:"required,oneof=1 2" label:"证书类型"`
		Action    string `json:"action" default:"start" validate:"required,oneof=start submit cancel" label:"验证操作"`
		SessionId string `json:"session_id" default:"" validate:"omitempty,max=128" label:"验证会话"`
		Name      string `json:"name" default:"" validate:"required_if=Action start,max=100" label:"证书名称"`
		Domain    string `json:"domain" default:"" validate:"required_if=Action start,max=253" label:"证书域名"`
		CA        string `json:"ca" default:"" validate:"omitempty,oneof=production staging" label:"签发环境"`
		Challenge string `json:"challenge" default:"http" validate:"required,oneof=http dns" label:"验证方式"`
		SecretId  *int64 `json:"secret_id" validate:"omitempty,min=0" label:"密钥ID"`
		AutoRenew *int   `json:"auto_renew" validate:"omitempty,oneof=0 1" label:"自动续签"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	manual := form.Type == cert_type.Manual
	if !manual && (form.Action != "start" || form.SessionId != "") {
		c.Error("自动申请不支持分步验证")
		return
	}
	if form.Action == "start" && form.SessionId != "" {
		c.Error("开始验证时不能指定已有会话")
		return
	}
	if form.Action != "start" && form.SessionId == "" {
		c.Error("验证会话不能为空")
		return
	}
	request := serviceSite.CertificateRequest{
		Action:    form.Action,
		SessionId: form.SessionId,
		Domain:    form.Domain,
		CA:        webServer.CertificateCA(form.CA),
		Type:      &form.Type,
		Challenge: webServer.CertificateChallenge(form.Challenge),
		SecretId:  form.SecretId,
		AutoRenew: form.AutoRenew,
	}
	timeout := webServer.CertificateRequestTimeout
	writeTimeout := timeout + 5*time.Second
	if manual {
		timeout = serviceSite.CertificateRequestTimeout
		writeTimeout = timeout + time.Minute
	}
	_ = http.NewResponseController(c.Writer).SetWriteDeadline(time.Now().Add(writeTimeout))
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), timeout)
	defer cancel()
	var data interface{}
	var err error
	if manual {
		data, err = service.Site.ManualCert(ctx, 0, form.Name, request)
	} else {
		data, err = service.Site.ObtainCert(ctx, form.Name, request)
	}
	if err != nil {
		c.Error(err.Error())
		return
	}
	msg := "验证信息获取成功"
	if form.Action == "cancel" {
		msg = "已取消验证"
	}
	if certificate, ok := data.(*model.Cert); ok && certificate != nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "申请网站证书", "申请网站证书["+certificate.Name+"]")
		msg = "申请成功"
		if manual {
			if certificate.Challenge == string(webServer.CertificateChallengeDNS) {
				msg += "，可以删除本次验证的 TXT 记录"
			} else {
				msg += "，可以删除本次验证文件"
			}
		}
	}
	c.SuccessWithData(msg, data)
}

// Renew 续签 ACME 证书，type 控制自动或手动验证。
func Renew(c *context.Context) {
	var form struct {
		Type      int    `json:"type" default:"2" validate:"required,oneof=1 2" label:"证书类型"`
		Action    string `json:"action" default:"start" validate:"required,oneof=start submit cancel" label:"验证操作"`
		SessionId string `json:"session_id" default:"" validate:"omitempty,max=128" label:"验证会话"`
		Id        int64  `json:"id" default:"0" validate:"required,min=1" label:"证书ID"`
		Domain    string `json:"domain" default:"" validate:"required_if=Action start,max=253" label:"续签域名"`
		CA        string `json:"ca" default:"" validate:"omitempty,oneof=production staging" label:"签发环境"`
		Challenge string `json:"challenge" validate:"omitempty,oneof=http dns" label:"验证方式"`
		SecretId  *int64 `json:"secret_id" validate:"omitempty,min=0" label:"密钥ID"`
		AutoRenew *int   `json:"auto_renew" validate:"omitempty,oneof=0 1" label:"自动续签"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	manual := form.Type == cert_type.Manual
	if !manual && (form.Action != "start" || form.SessionId != "") {
		c.Error("自动续签不支持分步验证")
		return
	}
	if form.Action == "start" && form.SessionId != "" {
		c.Error("开始验证时不能指定已有会话")
		return
	}
	if form.Action != "start" && form.SessionId == "" {
		c.Error("验证会话不能为空")
		return
	}
	request := serviceSite.CertificateRequest{
		Action:    form.Action,
		SessionId: form.SessionId,
		Domain:    form.Domain,
		CA:        webServer.CertificateCA(form.CA),
		Type:      &form.Type,
		Challenge: webServer.CertificateChallenge(form.Challenge),
		SecretId:  form.SecretId,
		AutoRenew: form.AutoRenew,
	}
	timeout := webServer.CertificateRequestTimeout
	writeTimeout := timeout + 5*time.Second
	if manual {
		timeout = serviceSite.CertificateRequestTimeout
		writeTimeout = timeout + time.Minute
	}
	_ = http.NewResponseController(c.Writer).SetWriteDeadline(time.Now().Add(writeTimeout))
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), timeout)
	defer cancel()
	var data interface{}
	var err error
	if manual {
		data, err = service.Site.ManualCert(ctx, form.Id, "", request)
	} else {
		data, err = service.Site.RenewCert(ctx, form.Id, request)
	}
	if err != nil {
		c.Error(err.Error())
		return
	}
	msg := "验证信息获取成功"
	if form.Action == "cancel" {
		msg = "已取消验证"
	}
	if certificate, ok := data.(*model.Cert); ok && certificate != nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "续签网站证书", "续签网站证书["+strconv.FormatInt(form.Id, 10)+"]")
		msg = "续签成功"
		if manual {
			if certificate.Challenge == string(webServer.CertificateChallengeDNS) {
				msg += "，可以删除本次验证的 TXT 记录"
			} else {
				msg += "，可以删除本次验证文件"
			}
		}
	}
	c.SuccessWithData(msg, data)
}
