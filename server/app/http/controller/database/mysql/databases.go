package mysql

import (
	stdContext "context"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// Databases 分页读取当前账户可访问的数据库列表。
func (Controller) Databases(c *context.Context) {
	var form struct {
		Id          int64  `json:"id" default:"0" validate:"required,min=1" label:"实例ID"`
		Keyword     string `json:"keyword" validate:"omitempty,max=1024" label:"数据库名称关键词"`
		ShowSystem  bool   `json:"show_system" default:"false" validate:"omitempty" label:"显示系统数据库"`
		OrderByType string `json:"order_by_type" default:"asc" validate:"omitempty,oneof=asc desc" label:"排序类型"`
		Page        int    `json:"page" default:"1" validate:"min=1,max=1000000" label:"页码"`
		PageSize    int    `json:"page_size" default:"20" validate:"min=1,max=1000" label:"每页数量"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	ctx, cancel := stdContext.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	m, err := service.Database.Mysql(ctx, form.Id, "")
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer m.Close()
	data, err := m.Databases(ctx, form.Keyword, form.Page, form.PageSize, form.ShowSystem, form.OrderByType)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "MySQL 数据库列表", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("获取成功", data)
}
