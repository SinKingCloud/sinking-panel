package process

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	serviceSite "server/app/service/site"
	"server/app/util/context"
)

// Update 修改网站进程配置。
func Update(c *context.Context) {
	var form struct {
		Id     int64                      `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Config *serviceSite.ProcessUpdate `json:"config" validate:"omitempty" label:"进程配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if err := service.Site.UpdateProcess(form.Id, form.Config); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站进程", "修改网站["+strconv.FormatInt(form.Id, 10)+"]进程配置")
	c.Success("修改成功")
}
