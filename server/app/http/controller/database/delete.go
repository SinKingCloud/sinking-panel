package database

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Delete 删除保存的数据库实例，不修改远端数据库。
func Delete(c *context.Context) {
	var form struct {
		Ids []int64 `json:"ids" validate:"required,min=1,dive,min=1" label:"实例ID列表"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if err := service.Database.DeleteByIds(form.Ids); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "删除数据库实例", "删除数据库实例")
	c.Success("删除成功")
}
