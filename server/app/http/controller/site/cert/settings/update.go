package settings

import (
	"fmt"
	"server/app/constant"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"strings"
)

// Update 修改证书自动续签任务设置。
func Update(c *context.Context) {
	var form struct {
		Enabled *bool  `json:"enabled" label:"自动续签"`
		Spec    string `json:"spec" validate:"omitempty,max=200" label:"执行周期"`
		Retries *int   `json:"retries" validate:"omitempty,min=0,max=10" label:"重试次数"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Enabled == nil || form.Spec == "" || form.Retries == nil {
		c.Error("请填写自动续签开关、执行周期和重试次数")
		return
	}
	settings := map[string]string{
		constant.CertAutoRenewEnabled: strconv.FormatBool(*form.Enabled),
		constant.CertAutoRenewSpec:    strings.TrimSpace(form.Spec),
		constant.CertAutoRenewRetries: strconv.Itoa(*form.Retries),
	}
	if err := service.Site.SetCertSettings(settings); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改证书续签设置", fmt.Sprintf("自动续签：%t，执行周期：%s，重试次数：%d", *form.Enabled, settings[constant.CertAutoRenewSpec], *form.Retries))
	c.Success("保存成功")
}
