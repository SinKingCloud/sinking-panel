package log

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// List 查看网站备份任务日志。
func List(c *context.Context) {
	var form struct {
		After    int64 `json:"after" default:"0" validate:"numeric,min=0" label:"增量游标"`
		Before   int64 `json:"before" default:"0" validate:"numeric,min=0" label:"历史游标"`
		PageSize int   `json:"page_size" default:"300" validate:"numeric,min=1,max=10000" label:"读取行数"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	logs, err := service.Site.ReadBackupLog(form.After, form.Before, form.PageSize)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站备份日志", "查看网站自动备份任务日志")
	c.SuccessWithData("获取成功", logs)
}
