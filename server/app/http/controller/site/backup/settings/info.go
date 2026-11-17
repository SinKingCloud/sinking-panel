package settings

import (
	"server/app/constant"
	"server/app/service"
	"server/app/util/context"
	"strconv"

	sinking_web "github.com/SinKingCloud/sinking-go/sinking-web"
)

// Info 获取网站自动备份设置。
func Info(c *context.Context) {
	configs, err := service.Site.GetBackupSettings()
	if err != nil {
		c.Error(err.Error())
		return
	}
	enabled, _ := strconv.ParseBool(configs[constant.SiteBackupEnabled])
	retries, _ := strconv.Atoi(configs[constant.SiteBackupRetries])
	localKeep, _ := strconv.Atoi(configs[constant.SiteBackupLocalKeep])
	uploadOss, _ := strconv.ParseBool(configs[constant.SiteBackupUploadOss])
	ossId, _ := strconv.ParseInt(configs[constant.SiteBackupOssId], 10, 64)
	ossKeep, _ := strconv.Atoi(configs[constant.SiteBackupOssKeep])
	c.SuccessWithData("获取成功", sinking_web.H{
		"enabled":    enabled,
		"spec":       configs[constant.SiteBackupSpec],
		"retries":    retries,
		"local_keep": localKeep,
		"upload_oss": uploadOss,
		"oss_id":     ossId,
		"oss_keep":   ossKeep,
		"prefix":     configs[constant.SiteBackupPrefix],
	})
}
