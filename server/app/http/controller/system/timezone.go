package system

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strings"
)

// Timezone 获取或设置系统时区。
func Timezone(c *context.Context) {
	var form struct {
		Action   string `json:"action" default:"get" validate:"required,oneof=get set" label:"操作类型"`
		Timezone string `json:"timezone" default:"" validate:"required_if=Action set,max=128" label:"时区"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	switch form.Action {
	case "get":
		data, err := service.System.GetTimezone()
		if err != nil {
			c.Error(err.Error())
		} else {
			c.SuccessWithData("获取成功", data)
		}
	case "set":
		if err := service.System.SetTimezone(form.Timezone); err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改系统时区", "系统时区设置为 "+strings.TrimSpace(form.Timezone))
			c.Success("设置成功")
		}
	default:
		c.Error("不支持该操作")
	}
}
