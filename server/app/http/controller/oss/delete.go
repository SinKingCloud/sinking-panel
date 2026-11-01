package oss

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Delete 删除面板中的存储配置，不删除远端文件或存储桶。
func Delete(c *context.Context) {
	var form struct {
		Ids []int64 `json:"ids" validate:"required,min=1,max=1000,dive,min=1" label:"存储ID列表"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if err := service.Oss.DeleteByIds(form.Ids); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "删除 OSS 存储", "删除面板中的 OSS 存储配置")
	c.Success("删除成功")
}
