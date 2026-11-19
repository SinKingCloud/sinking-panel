package backup

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Download 获取本地备份的文件签名或云端备份的签名 URL。
func Download(c *context.Context) {
	var form struct {
		Id      int64  `json:"id" default:"0" validate:"required,min=1" label:"备份ID"`
		Storage string `json:"storage" default:"" validate:"omitempty,oneof=local oss" label:"备份副本"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	sign, err := service.Site.DownloadBackup(form.Id, form.Storage)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "下载网站备份", "下载备份["+strconv.FormatInt(form.Id, 10)+"]")
		c.SuccessWithData("获取成功", sign)
	}
}
