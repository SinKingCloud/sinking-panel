package sqlserver

import (
	stdContext "context"
	"encoding/json"
	"server/app/enum/log_type"
	"server/app/service"
	engine "server/app/service/database/sqlserver"
	"server/app/util/context"
	"strconv"
	"time"
)

// AddColumn 添加字段。
func (Controller) AddColumn(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Schema       string `json:"schema" validate:"omitempty,max=1024" label:"数据库架构"`
		Table        string `json:"table" validate:"required,max=1024" label:"数据表"`
		Column       string `json:"column" validate:"required,max=65536" label:"字段配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	var column engine.ColumnDefinition
	if json.Unmarshal([]byte(form.Column), &column) != nil {
		c.Error("字段配置格式不正确")
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	m, err := service.Database.Sqlserver(ctx, form.Id, form.DatabaseName)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	data, err := m.AddColumn(ctx, form.Schema, form.Table, column)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "SqlServer 添加字段", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
