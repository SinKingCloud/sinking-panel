package cert

import (
	"fmt"
	"server/app/enum/log_type"
	"server/app/service"
	serviceSite "server/app/service/site"
	"server/app/util/context"
	"strings"
)

// Settings 获取或修改证书自动续签任务设置。
func Settings(c *context.Context) {
	var form struct {
		Action  string `json:"action" default:"get" validate:"required,oneof=get set" label:"操作类型"`
		Enabled *bool  `json:"enabled" label:"自动续签"`
		Spec    string `json:"spec" validate:"omitempty,max=200" label:"执行周期"`
		Retries *int   `json:"retries" validate:"omitempty,min=0,max=10" label:"重试次数"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "set" {
		if form.Enabled == nil || form.Spec == "" || form.Retries == nil {
			c.Error("请填写自动续签开关、执行周期和重试次数")
			return
		}
		settings := serviceSite.CertSettings{
			Enabled: *form.Enabled,
			Spec:    strings.TrimSpace(form.Spec),
			Retries: *form.Retries,
		}
		if err := service.Site.SetCertSettings(settings); err != nil {
			c.Error(err.Error())
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改证书续签设置", fmt.Sprintf("自动续签：%t，执行周期：%s，重试次数：%d", settings.Enabled, settings.Spec, settings.Retries))
		c.SuccessWithData("保存成功", settings)
		return
	}
	settings, err := service.Site.GetCertSettings()
	if err != nil {
		c.Error(err.Error())
	} else {
		c.SuccessWithData("获取成功", settings)
	}
}
