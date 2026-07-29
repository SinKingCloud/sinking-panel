package auth

import (
	"server/app/constant"
	"server/app/http/middleware"
	"server/app/service"
	"server/app/util/context"
)

// Login 账号登录
func Login(c *context.Context) {
	var form struct {
		Account  string `json:"account" default:"" validate:"required" label:"账户"`
		Password string `json:"password" default:"" validate:"required" label:"密码"`
		Device   string `json:"device" default:"web" validate:"required,oneof=web pc mobile android" label:"登陆设备"`
		Token    string `json:"token" default:"" validate:"required" label:"验证码标识"`
		CaptchaX int    `json:"captcha_x" default:"" validate:"required,numeric" label:"验证码X坐标"`
		CaptchaY int    `json:"captcha_y" default:"" validate:"required,numeric" label:"验证码Y坐标"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if !service.Auth.CheckCaptcha(form.Token, form.CaptchaX, form.CaptchaY) {
		c.Error("验证码验证失败请重试")
		return
	}
	token, err := service.Auth.Login(form.Account, form.Password, form.Device, c.GetRequestIp())
	if err != nil {
		c.Error(err.Error())
		return
	}
	c.SuccessWithData("登录成功", token)
}

// Logout 退出登录
func Logout(c *context.Context) {
	middleware.CheckLogin(c)
	if c.IsAborted() {
		return
	}
	if err := service.Auth.Logout(c.Request.Header.Get(constant.JwtDeviceName)); err != nil {
		c.Error(err.Error())
		return
	}
	c.Success("注销成功")
}
