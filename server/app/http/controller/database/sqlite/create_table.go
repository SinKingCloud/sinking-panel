package sqlite

import (
	stdContext "context"
	"encoding/json"
	"server/app/enum/log_type"
	"server/app/service"
	engine "server/app/service/database/sqlite"
	"server/app/util/context"
	"strconv"
	"time"
)

// CreateTable 创建数据表。
func (Controller) CreateTable(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Schema       string `json:"schema" validate:"omitempty,max=1024" label:"数据库架构"`
		Table        string `json:"table" validate:"required,max=1024" label:"数据表"`
		Columns      string `json:"columns" validate:"required,max=1048576" label:"字段配置"`
		PrimaryKeys  string `json:"primary_keys" validate:"omitempty,max=65536" label:"主键字段"`
		Indexes      string `json:"indexes" validate:"omitempty,max=1048576" label:"索引配置"`
		ForeignKeys  string `json:"foreign_keys" validate:"omitempty,max=1048576" label:"外键配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	var columns []engine.ColumnDefinition
	if json.Unmarshal([]byte(form.Columns), &columns) != nil {
		c.Error("字段配置格式不正确")
		return
	}
	var primaryKeys []string
	if form.PrimaryKeys != "" {
		if json.Unmarshal([]byte(form.PrimaryKeys), &primaryKeys) != nil {
			c.Error("主键字段格式不正确")
			return
		}
	}
	var foreignKeys []engine.ForeignKeyDefinition
	if form.ForeignKeys != "" {
		if json.Unmarshal([]byte(form.ForeignKeys), &foreignKeys) != nil {
			c.Error("外键配置格式不正确")
			return
		}
	}
	var indexes []engine.IndexDefinition
	if form.Indexes != "" && json.Unmarshal([]byte(form.Indexes), &indexes) != nil {
		c.Error("索引配置格式不正确")
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
	data, err := m.CreateTable(ctx, form.Schema, form.Table, columns, primaryKeys, indexes, foreignKeys...)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "Sqlite 创建数据表", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
