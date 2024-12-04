package auth

import (
	"github.com/SinKingCloud/sinking-go/sinking-web"
	"server/app/constant"
	"server/app/model"
	"server/app/service"
	"server/app/service/user_log"
	"server/app/util/server"
)

type ControllerWeb struct {
}

func (ControllerWeb) Info(c *server.Context) {
	config := service.Config.GetSysConfigs(constant.WebGroup, false)
	c.SuccessWithData("获取成功", sinking_web.H{
		"title":    c.GetStringWithDefault(config[constant.WebTitle], "默认标题"),
		"name":     c.GetStringWithDefault(config[constant.WebName], "默认名称"),
		"keywords": c.GetStringWithDefault(config[constant.WebKeyWords], "默认名称"),
		"describe": c.GetStringWithDefault(config[constant.WebDescribe], "默认名称"),
		"contact":  c.GetStringWithDefault(config[constant.WebContact], ""),
		"url":      c.GetStringWithDefault(config[constant.WebUrl], "默认标题"),
	})
}

func (ControllerWeb) Login(c *server.Context) {
	type Form struct {
		Phone    string `form:"phone" json:"phone" default:"" validate:"required,numeric,len=11" label:"手机号"`
		Code     string `form:"code" json:"code" default:"" validate:"-" label:"短信验证码"`
		Password string `form:"password" json:"password" default:"" validate:"-" label:"密码"`
		Device   string `form:"device" json:"device" default:"web" validate:"required,oneof=web" label:"登陆设备"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	var user *model.User
	var err error
	if form.Code != "" {
		//短信登陆
		user, err = service.Auth.GetWebUserInfoByPhone(form.Phone, form.Code)
	} else {
		//密码登陆
		user, err = service.Auth.GetWebUserInfoByPwd(form.Phone, form.Password)
	}
	if err != nil {
		c.Error(err.Error())
		return
	}
	token, err := service.Auth.GetJwtToken(form.Device, user, c.GetRequestIp(), constant.JwtExpireTime)
	if err != nil {
		c.Error(err.Error())
		return
	}
	//写入登录日志
	service.UserLog.CreateAsync(user.Id, c.GetRequestIp(), user_log.EventLogin, "账户登录", "账户登陆成功")
	c.SuccessWithData("登陆成功", token)
}

// OutLogin 退出登录
func (ControllerWeb) OutLogin(c *server.Context) {
	user := c.GetUserInfo()
	u, err := service.User.FindByIdWithCache(user.Id, true)
	if err == nil && u != nil {
		service.Auth.OutLogin(c.Request.Header.Get(constant.JwtDeviceName), u)
	}
	service.UserLog.CreateAsync(user.Id, c.GetRequestIp(), user_log.EventLogin, "注销登录", "注销登陆成功")
	c.Success("注销登录成功")
}
