package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Domain 获取或更新网站域名，已有域名的证书绑定保持不变。
func Domain(c *context.Context) {
	var form struct {
		Action  string    `json:"action" default:"get" validate:"required,oneof=get set" label:"操作类型"`
		Id      int64     `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Domains *[]string `json:"domains" validate:"omitempty,max=1000,unique,dive,required,max=253" label:"网站域名"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "get" {
		data, err := service.Site.GetDomains(form.Id)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站域名", "查看网站["+strconv.FormatInt(form.Id, 10)+"]域名配置")
			c.SuccessWithData("获取成功", data)
		}
	} else {
		if form.Domains == nil {
			c.Error("网站域名不能为空")
			return
		}
		err := service.Site.UpdateDomains(form.Id, *form.Domains)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站域名", "修改网站["+strconv.FormatInt(form.Id, 10)+"]域名配置")
			c.Success("修改成功")
		}
	}
}
