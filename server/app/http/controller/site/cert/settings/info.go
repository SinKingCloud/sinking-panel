package settings

import (
	"server/app/constant"
	"server/app/service"
	"server/app/util/context"
	"strconv"

	"github.com/SinKingCloud/sinking-go/sinking-web"
)

// Info 获取证书自动续签任务设置。
func Info(c *context.Context) {
	settings, err := service.Site.GetCertSettings()
	if err != nil {
		c.Error(err.Error())
		return
	}
	enabled, _ := strconv.ParseBool(settings[constant.CertAutoRenewEnabled])
	retries, _ := strconv.Atoi(settings[constant.CertAutoRenewRetries])
	c.SuccessWithData("获取成功", sinking_web.H{
		"enabled": enabled,
		"spec":    settings[constant.CertAutoRenewSpec],
		"retries": retries,
	})
}
