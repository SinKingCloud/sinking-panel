package cert

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Info 获取证书详情。
func Info(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"证书ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Site.FindCert(form.Id)
	if err == nil && data != nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看证书详情", "查看网站证书["+data.Name+"]详情")
		c.SuccessWithData("获取成功", data)
	} else {
		c.Error("获取失败")
	}
}
