package postgres

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// DatabaseInfo 读取真实数据库的当前属性和占用空间。
func (Controller) DatabaseInfo(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" default:"" validate:"required,max=1024" label:"数据库名称"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	m, err := service.Database.Postgres(ctx, form.Id, "postgres")
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	data, err := m.DatabaseInfo(ctx, form.DatabaseName)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "Postgres 数据库详情", "查看实例["+strconv.FormatInt(form.Id, 10)+"]的数据库["+form.DatabaseName+"]")
	c.SuccessWithData("获取成功", data)
}
