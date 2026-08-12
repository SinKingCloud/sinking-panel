package middleware

import (
	"server/app/constant"
	"server/app/service"
	"server/app/util/context"
	"server/app/util/jwt"
	"strings"

	"github.com/gorilla/websocket"
)

// CheckLogin 判断登录
func CheckLogin(c *context.Context) {
	token := c.Request.Header.Get(constant.JwtTokenName)
	types := c.Request.Header.Get(constant.JwtDeviceName)
	isWebSocket := websocket.IsWebSocketUpgrade(c.Request)
	if isWebSocket {
		protocols := websocket.Subprotocols(c.Request)
		if token == "" && len(protocols) > 0 {
			token = protocols[0]
		}
		if types == "" && len(protocols) > 1 {
			types = protocols[1]
		}
	}
	if token == "" || types == "" {
		var form struct {
			Token  string `json:"token"`
			Device string `json:"device"`
		}
		_ = c.BindAll(&form)
		if token == "" && form.Token != "" {
			token = form.Token
		}
		if types == "" && form.Device != "" {
			types = form.Device
		}
	}
	if token == "" || (types == "" && !isWebSocket) {
		c.NotLogin("您还未登陆,请先登陆账户", nil)
		c.Abort()
		return
	}
	key := jwt.CheckToken(token)
	if key == nil || key.User == nil {
		c.TokenError("登陆超时,请重新登陆", nil)
		c.Abort()
		return
	}
	if types == "" {
		prefix := constant.LoginToken + "."
		for configKey, configValue := range service.Config.Group(constant.LoginGroup) {
			if configValue == key.User.LoginToken && strings.HasPrefix(configKey, prefix) {
				types = strings.TrimPrefix(configKey, prefix)
				break
			}
		}
	}
	loginToken := service.Config.Get(constant.LoginGroup, constant.LoginToken+"."+types)
	if loginToken == "" {
		c.TokenError("您的账户已注销登陆,请重新登陆", nil)
		c.Abort()
		return
	}
	if key.User.LoginToken != loginToken {
		c.TokenError("您的账户已在其他设备登陆,请重新登陆", nil)
		c.Abort()
		return
	}
	c.SetUserInfo(key.User)
	c.Next()
}
