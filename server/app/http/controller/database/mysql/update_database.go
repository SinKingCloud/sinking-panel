package mysql

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// UpdateDatabase 修改数据库的默认字符集和排序规则。
func (Controller) UpdateDatabase(c *context.Context) {
	var form struct {
		Id        int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		Name      string `json:"name" default:"" validate:"required,max=1024" label:"数据库名称"`
		Charset   string `json:"charset" default:"" validate:"omitempty,max=64" label:"字符集"`
		Collation string `json:"collation" default:"" validate:"omitempty,max=128" label:"排序规则"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	m, err := service.Database.Mysql(ctx, form.Id, "information_schema")
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	if err = m.UpdateDatabase(ctx, form.Name, form.Charset, form.Collation); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "MySQL 修改数据库", "修改实例["+strconv.FormatInt(form.Id, 10)+"]的数据库["+form.Name+"]")
	c.Success("修改成功")
}
