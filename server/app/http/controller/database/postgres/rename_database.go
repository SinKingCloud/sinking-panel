package postgres

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// RenameDatabase 重命名真实数据库。
func (Controller) RenameDatabase(c *context.Context) {
	var form struct {
		Id      int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		Name    string `json:"name" default:"" validate:"required,max=1024" label:"数据库名称"`
		NewName string `json:"new_name" default:"" validate:"required,max=1024" label:"新数据库名称"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	m, err := service.Database.Postgres(ctx, form.Id, "postgres")
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	if err = m.RenameDatabase(ctx, form.Name, form.NewName); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "Postgres 重命名数据库", "修改实例["+strconv.FormatInt(form.Id, 10)+"]的数据库["+form.Name+"]，重命名为["+form.NewName+"]")
	c.Success("修改成功")
}
