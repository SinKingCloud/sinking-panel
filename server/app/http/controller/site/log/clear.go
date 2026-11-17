package log

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/caddy"
	"server/app/util/context"
)

// Clear 清理网站访问、WAF 和进程日志。
func Clear(c *context.Context) {
	var form struct {
		Id   int64  `json:"id" default:"0" validate:"required,numeric,min=1" label:"网站ID"`
		Type string `json:"type" default:"access" validate:"required,oneof=access waf process" label:"日志类型"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	logType := caddy.LogType(form.Type)
	err := service.Site.ClearLog(form.Id, logType)
	if err != nil {
		c.Error(err.Error())
		return
	}
	typeName := map[caddy.LogType]string{
		caddy.LogAccess:  "访问",
		caddy.LogWAF:     "WAF",
		caddy.LogProcess: "进程",
	}[logType]
	service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "清理网站日志", "清理网站["+strconv.FormatInt(form.Id, 10)+"]"+typeName+"日志")
	c.Success("清理成功")
}
