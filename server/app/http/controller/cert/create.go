package cert

import (
	"server/app/enum/cert_type"
	"server/app/enum/log_type"
	"server/app/model"
	"server/app/service"
	"server/app/util/context"
)

// Create 导入证书。
func Create(c *context.Context) {
	var form struct {
		Name        string `json:"name" default:"" validate:"required,max=100" label:"证书名称"`
		Certificate string `json:"certificate" default:"" validate:"required,max=4194304" label:"证书内容"`
		PrivateKey  string `json:"private_key" default:"" validate:"required,max=1048576" label:"证书私钥"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &model.Cert{
		Name:        form.Name,
		Type:        cert_type.Manual,
		Certificate: form.Certificate,
		PrivateKey:  form.PrivateKey,
	}
	err := service.Site.CreateCert(data)
	if err == nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "导入网站证书", "导入网站证书["+data.Name+"]")
		c.SuccessWithData("导入成功", data)
	} else {
		c.Error(err.Error())
	}
}
