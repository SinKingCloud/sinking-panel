package mongo

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"strings"
	"time"
)

// DatabaseInfo 查询 MongoDB 数据库的集合、文档和存储统计。
func (Controller) DatabaseInfo(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if strings.TrimSpace(form.DatabaseName) == "" {
		c.Error("请选择数据库")
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	m, err := service.Database.Mongo(ctx, form.Id, form.DatabaseName)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	data, err := m.DatabaseInfo(ctx)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "Mongo 数据库详情", "查看实例["+strconv.FormatInt(form.Id, 10)+"]的数据库["+form.DatabaseName+"]")
	c.SuccessWithData("获取成功", data)
}
