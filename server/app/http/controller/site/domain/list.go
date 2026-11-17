package domain

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// List 获取网站域名，已有域名的证书绑定保持不变。
func List(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Site.GetDomains(form.Id)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站域名", "查看网站["+strconv.FormatInt(form.Id, 10)+"]域名配置")
		c.SuccessWithData("获取成功", data)
	}
}
