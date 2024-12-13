package auth

import (
	"image/png"
	"server/app/service"
	"server/app/util/server"
)

func Captcha(c *server.Context) {
	type Form struct {
		Token string `json:"token" default:"" validate:"required,len=16" label:"验证码标识"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	img, _ := service.Auth.GetCaptcha(form.Token)
	err := png.Encode(c.Writer, img)
	if err != nil {
		c.ErrorWithData("生成验证码失败", err)
	}
}
