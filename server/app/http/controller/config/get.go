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
	configKey := form.Group
	if form.Key != "" {
		configKey += "." + form.Key
	}
	isSensitive := false
	for _, group := range constant.SensitiveGroups {
		if configKey == group || strings.HasPrefix(configKey, group+".") {
			isSensitive = true
			break
		}
	}
	if isSensitive {
		c.Error("配置不存在")
		return
	}
	content := "查看系统配置[" + form.Group + "]数据"
	if form.Key != "" {
		content = "查看系统配置[" + form.Group + "." + form.Key + "]数据"
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看系统配置", content)
	if form.Key != "" {
		c.SuccessWithData("获取数据成功", sinking_web.H{
			form.Key: service.Config.Get(form.Group, form.Key),
		})
	} else {
		c.SuccessWithData("获取数据成功", service.Config.Group(form.Group))
	}
}
