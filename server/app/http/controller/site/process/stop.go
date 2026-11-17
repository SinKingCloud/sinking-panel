package process

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Stop 停止网站进程。
func Stop(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Site.Process(form.Id, "stop")
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "停止项目进程", "停止网站["+strconv.FormatInt(form.Id, 10)+"]进程")
	c.SuccessWithData("停止成功", data)
}
