package sqlserver

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// DropDatabase 批量删除真实数据库及其数据。
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
	m, err := service.Database.Sqlserver(ctx, form.Id, "master")
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	if err = m.DropDatabases(ctx, form.Names); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "SqlServer 删除数据库", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.Success("删除成功")
}
