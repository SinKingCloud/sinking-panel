package sqlite

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// SaveObject 保存数据库对象。
func (Controller) SaveObject(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Schema       string `json:"schema" validate:"omitempty,max=1024" label:"数据库架构"`
		Kind         string `json:"kind" default:"view" validate:"required,oneof=view routine trigger" label:"对象类型"`
		Name         string `json:"name" validate:"required,max=4096" label:"对象名称"`
		Definition   string `json:"definition" validate:"required,max=1048576" label:"对象定义"`
		Replace      bool   `json:"replace" validate:"omitempty" label:"替换已有对象"`
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
	data, err := m.SaveObject(ctx, form.Schema, form.Kind, form.Name, form.Definition, form.Replace)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "Sqlite 保存数据库对象", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("操作成功", data)
}
