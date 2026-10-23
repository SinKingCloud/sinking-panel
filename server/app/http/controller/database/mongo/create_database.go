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

// CreateDatabase 通过初始集合创建 MongoDB 数据库。
func (Controller) CreateDatabase(c *context.Context) {
	var form struct {
		Id         int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		Name       string `json:"name" validate:"required,max=255" label:"数据库名称"`
		Collection string `json:"collection" validate:"required,max=255" label:"初始集合"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	form.Name = strings.TrimSpace(form.Name)
	if form.Name == "" {
		c.Error("数据库名称应为 1 至 255 个字符，且不包含空字符")
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	m, err := service.Database.Mongo(ctx, form.Id, form.Name)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	if err = m.CreateDatabase(ctx, form.Collection); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "Mongo 创建数据库", "在实例["+strconv.FormatInt(form.Id, 10)+"]创建数据库["+form.Name+"]")
	c.Success("创建成功")
}
