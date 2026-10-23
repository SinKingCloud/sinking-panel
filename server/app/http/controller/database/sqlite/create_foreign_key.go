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

// CreateForeignKey 创建外键。
func (Controller) CreateForeignKey(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Schema       string `json:"schema" validate:"omitempty,max=1024" label:"数据库架构"`
		Table        string `json:"table" validate:"required,max=1024" label:"数据表"`
		ForeignKey   string `json:"foreign_key" validate:"required,max=65536" label:"外键配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	var foreignKey engine.ForeignKeyDefinition
	if json.Unmarshal([]byte(form.ForeignKey), &foreignKey) != nil {
		c.Error("外键配置格式不正确")
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
	data, err := m.CreateForeignKey(ctx, form.Schema, form.Table, foreignKey)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "Sqlite 创建外键", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
