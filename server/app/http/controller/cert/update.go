package cert

import (
	"strconv"

	"server/app/enum/log_type"
	certRepository "server/app/repository/cert"
	"server/app/service"
	"server/app/util/context"
)

// Update 更新证书，未传入的字段保持原值。
func Update(c *context.Context) {
	var form struct {
		Id          int64   `json:"id" default:"0" validate:"required,min=1" label:"证书ID"`
		Name        *string `json:"name" validate:"omitempty,max=100" label:"证书名称"`
		Certificate *string `json:"certificate" validate:"omitempty,max=4194304" label:"证书内容"`
		PrivateKey  *string `json:"private_key" validate:"omitempty,max=1048576" label:"证书私钥"`
	}
	if ok, message := c.ValidatorAll(&form); !ok {
		c.Error(message)
		return
	}
	if err := service.Site.UpdateCert(form.Id, &certRepository.UpdateCert{
		Name:        form.Name,
		Certificate: form.Certificate,
		PrivateKey:  form.PrivateKey,
	}); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站证书", "修改网站证书["+strconv.FormatInt(form.Id, 10)+"]")
	c.Success("修改成功")
}
