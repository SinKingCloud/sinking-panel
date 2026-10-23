package mongo

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// DropDatabase 批量删除同一 MongoDB 实例中的业务数据库。
func (Controller) DropDatabase(c *context.Context) {
	var form struct {
		Id    int64    `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		Names []string `json:"names" validate:"required,min=1,dive,required,max=1024" label:"数据库名称列表"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	m, err := service.Database.Mongo(ctx, form.Id, "")
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	if err = m.DropDatabases(ctx, form.Names); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "Mongo 删除数据库", "删除实例["+strconv.FormatInt(form.Id, 10)+"]的真实数据库")
	c.Success("删除成功")
}
