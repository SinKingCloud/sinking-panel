package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Enable 启用网站。
func Enable(c *context.Context) {
	changeStatus(c, true)
}

// Disable 停用网站。
func Disable(c *context.Context) {
	changeStatus(c, false)
}

func changeStatus(c *context.Context, enabled bool) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	action := "停用"
	err := service.Site.Disable(form.Id)
	if enabled {
		action = "启用"
		err = service.Site.Enable(form.Id)
	}
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, action+"网站", action+"网站["+strconv.FormatInt(form.Id, 10)+"]")
		c.Success(action + "成功")
	}
}
