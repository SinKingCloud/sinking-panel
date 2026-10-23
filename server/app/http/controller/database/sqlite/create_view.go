package sqlite

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// CreateView 创建视图。
func (Controller) CreateView(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Schema       string `json:"schema" validate:"omitempty,max=1024" label:"数据库架构"`
		Table        string `json:"table" validate:"required,max=1024" label:"数据表"`
		Statement    string `json:"statement" validate:"required,max=1048576" label:"查询语句"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	m, err := service.Database.Sqlite(ctx, form.Id, form.DatabaseName)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	data, err := m.CreateView(ctx, form.Schema, form.Table, form.Statement)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "Sqlite 创建视图", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
