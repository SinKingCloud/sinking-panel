package middleware

import (
	"github.com/SinKingCloud/sinking-go/sinking-web"
	"server/app/constant"
	"server/app/service"
	"server/app/util"
	"server/app/util/jwt"
	"server/app/util/server"
	"time"
)

// CheckLogin 判断登录
func CheckLogin() sinking_web.HandlerFunc {
	return server.HandleFunc(func(c *server.Context) {
		token := c.Request.Header.Get(constant.JwtTokenName)
		types := c.Request.Header.Get(constant.JwtDeviceName)
		if token == "" || types == "" {
			c.NotLogin("您还未登陆,请先登陆账户", nil)
			c.Abort()
			return
		}
		key := jwt.CheckToken(token)
		if key == nil || (key.ExpiresAt > 0 && time.Now().Unix() > key.ExpiresAt) || key.User == nil {
			c.TokenError("登陆超时,请重新登陆", nil)
			c.Abort()
			return
		}
		//获取用户最新信息
		user, err := service.User.FindByIdWithCache(key.User.Id, false)
		if err != nil || user == nil {
			c.TokenError("登陆超时,请重新登陆", nil)
			c.Abort()
			return
		}
		webToken := jwt.GetLoginToken(user.LoginToken)
		//判断用户token一致性
		if webToken[types] == "" {
			c.TokenError("您的账户已注销登陆,请重新登陆", nil)
			c.Abort()
			return
		}
		jwtToken := jwt.GetLoginToken(key.User.LoginToken)
		if jwtToken[types] != webToken[types] && util.Conf.GetString(constant.ServerMode) != "dev" {
			c.TokenError("您的账户已在其他设备登陆,请重新登陆", nil)
			c.Abort()
			return
		}
		if !service.User.CheckAdmin(user.Id) && user.Status != 0 {
			c.TokenError("您的账户已被禁止登陆,请联系管理员", nil)
			c.Abort()
			return
		}
		c.SetUserInfo(key.User)
		c.Next()
	})
}

// CheckAdmin 判断是否管理员
func CheckAdmin() sinking_web.HandlerFunc {
	return server.HandleFunc(func(c *server.Context) {
		if !c.IsAdmin() {
			c.NotAllow("您的账户权限不足", nil)
			c.Abort()
			return
		}
		c.Next()
	})
}
