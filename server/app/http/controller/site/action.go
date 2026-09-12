package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Enable 启用网站。
func Enable(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if err := service.Site.Enable(form.Id); err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "启用网站", "启用网站["+strconv.FormatInt(form.Id, 10)+"]")
		c.Success("启用成功")
	}
}

// Disable 停用网站。
func Disable(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if err := service.Site.Disable(form.Id); err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "停用网站", "停用网站["+strconv.FormatInt(form.Id, 10)+"]")
		c.Success("停用成功")
	}
}
