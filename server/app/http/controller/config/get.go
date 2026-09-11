package config

import (
	"strings"

	"server/app/constant"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"

	"github.com/SinKingCloud/sinking-go/sinking-web"
)

// Get 获取配置
func Get(c *context.Context) {
	var form struct {
		Group string `form:"group" json:"group" default:"" validate:"required,alphanum,max=100" label:"组ID"`
		Key   string `form:"key" json:"key" default:"" validate:"omitempty,max=100" label:"配置标识"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	policy := constant.SensitiveGroups[form.Group]
	mask := func(value string) string {
		characters := []rune(value)
		if len(characters) == 0 {
			return value
		}
		if len(characters) == 1 {
			return string(characters[0]) + strings.Repeat("*", 4) + string(characters[0])
		}
		stars := len(characters) - 2
		if stars < 4 {
			stars = 4
		}
		return string(characters[0]) + strings.Repeat("*", stars) + string(characters[len(characters)-1])
	}
	content := "查看系统配置[" + form.Group + "]数据"
	if form.Key != "" {
		content = "查看系统配置[" + form.Group + "." + form.Key + "]数据"
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看系统配置", content)
	if form.Key != "" {
		value := service.Config.Get(form.Group, form.Key)
		if policy.Read {
			value = mask(value)
		}
		c.SuccessWithData("获取数据成功", sinking_web.H{
			form.Key: value,
		})
	} else {
		values := service.Config.Group(form.Group)
		if policy.Read {
			masked := make(map[string]string, len(values))
			for key, value := range values {
				masked[key] = mask(value)
			}
			values = masked
		}
		c.SuccessWithData("获取数据成功", values)
	}
}
