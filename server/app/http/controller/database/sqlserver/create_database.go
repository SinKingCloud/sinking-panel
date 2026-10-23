package sqlserver

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// CreateDatabase 创建真实数据库。
func (Controller) CreateDatabase(c *context.Context) {
	var form struct {
		Id   int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		Name string `json:"name" default:"" validate:"required,max=1024" label:"数据库名称"`
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
	if err = m.CreateDatabase(ctx, form.Name); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "SqlServer 创建数据库", "在实例["+strconv.FormatInt(form.Id, 10)+"]创建数据库["+form.Name+"]")
	c.Success("创建成功")
}
