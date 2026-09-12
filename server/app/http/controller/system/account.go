package system

import (
	"fmt"
	"server/app/constant"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	ip2 "server/app/util/ip"
	"strings"

	"github.com/SinKingCloud/sinking-go/sinking-web"
)

// Account 账户管理
func Account(c *context.Context) {
	var form struct {
		Action   string `json:"action" default:"account" validate:"required,oneof=account update" label:"操作类型"`
		Account  string `json:"account" default:"" validate:"omitempty,alphanum,max=20" label:"登录账号"`
		Password string `json:"password" default:"" validate:"omitempty,min=6,max=20" label:"登录密码"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	switch form.Action {
	case "account":
		user := c.GetUserInfo()
		location := "未知"
		ipInfo, err := ip2.Query(user.LoginIp)
		if err == nil && ipInfo != nil {
			location = strings.ReplaceAll(fmt.Sprintf("%s %s %s %s", ipInfo.Country, ipInfo.Province, ipInfo.City, ipInfo.Isp), "  ", "")
		}
		c.SuccessWithData("获取成功", sinking_web.H{
			"account":        service.Config.Get(constant.LoginGroup, constant.LoginAccount),
			"login_ip":       user.LoginIp,
			"login_location": location,
			"login_time":     user.LoginTime,
		})
	case "update":
		if form.Account == "" && form.Password == "" {
			c.Error("登录账号和密码不能同时为空")
			return
		}
		err := service.Auth.UpdateAccount(form.Account, form.Password)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改登录信息", "修改登录账号或密码")
			c.Success("修改成功")
		}
	}
}
