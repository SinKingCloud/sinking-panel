package mysql

import (
	stdContext "context"
	"server/app/service"
	"server/app/util/context"
	"time"
)

// Charsets 读取实例支持的字符集、排序规则及默认值。
func (Controller) Charsets(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
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
	data, err := m.Charsets(ctx)
	if err != nil {
		c.Error(err.Error())
		return
	}
	c.SuccessWithData("获取成功", data)
}
