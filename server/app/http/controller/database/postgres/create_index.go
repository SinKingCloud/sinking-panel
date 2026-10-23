package postgres

import (
	stdContext "context"
	"encoding/json"
	"server/app/enum/log_type"
	"server/app/service"
	engine "server/app/service/database/postgres"
	"server/app/util/context"
	"strconv"
	"time"
)

// CreateIndex 创建索引。
func (Controller) CreateIndex(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Schema       string `json:"schema" validate:"omitempty,max=1024" label:"数据库架构"`
		Table        string `json:"table" validate:"required,max=1024" label:"数据表"`
		Name         string `json:"name" validate:"required,max=1024" label:"索引名称"`
		Columns      string `json:"columns" validate:"required,max=65536" label:"索引字段"`
		Unique       bool   `json:"unique" validate:"omitempty" label:"唯一索引"`
		Primary      bool   `json:"primary" validate:"omitempty" label:"主键索引"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	var columns []engine.IndexColumn
	if json.Unmarshal([]byte(form.Columns), &columns) != nil {
		c.Error("索引字段格式不正确")
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	m, err := service.Database.Postgres(ctx, form.Id, form.DatabaseName)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	data, err := m.CreateIndex(ctx, form.Schema, form.Table, form.Name, columns, form.Unique, form.Primary)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "Postgres 创建索引", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
