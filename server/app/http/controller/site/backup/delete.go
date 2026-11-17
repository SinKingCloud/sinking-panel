package backup

import (
	"fmt"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Delete 删除网站备份及其副本。
func Delete(c *context.Context) {
	var form struct {
		Ids []int64 `json:"ids" validate:"required,min=1,max=1000,unique,dive,min=1" label:"备份ID列表"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if err := service.Site.DeleteBackup(form.Ids); err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "删除网站备份", fmt.Sprintf("删除备份%v", form.Ids))
		c.Success("删除成功")
	}
}
