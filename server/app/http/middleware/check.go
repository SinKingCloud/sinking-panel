package middleware

import (
	"encoding/base64"
	"encoding/json"
	"server/app/constant"
	"server/app/service"
	"server/app/util/context"
	"server/app/util/jwt"

	"github.com/gorilla/websocket"
)

// CheckLogin 判断登录
func CheckLogin(c *context.Context) {
	token := c.Request.Header.Get(constant.JwtTokenName)
	types := c.Request.Header.Get(constant.JwtDeviceName)
	if token == "" || types == "" {
		if token == "" {
			if cookie, err := c.Request.Cookie(constant.JwtTokenName); err == nil {
				token = cookie.Value
			}
		}
		if types == "" {
			if cookie, err := c.Request.Cookie(constant.JwtDeviceName); err == nil {
				types = cookie.Value
			}
		}
	}
	if token == "" || types == "" {
		isWebSocket := websocket.IsWebSocketUpgrade(c.Request)
		if isWebSocket {
			protocols := websocket.Subprotocols(c.Request)
			if len(protocols) > 0 {
				var form struct {
					Token  string `json:"token"`
					Device string `json:"device"`
				}
				if data, err := base64.RawURLEncoding.DecodeString(protocols[0]); err == nil {
					_ = json.Unmarshal(data, &form)
				}
				if token == "" && form.Token != "" {
					token = form.Token
				}
				if types == "" && form.Device != "" {
					types = form.Device
				}
			}
		}
	}
	if token == "" || types == "" {
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
	c.Set(constant.JwtTokenName, token)
	c.Set(constant.JwtDeviceName, types)
	c.SetUserInfo(key.User)
	c.Next()
}
