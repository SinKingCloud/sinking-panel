package process

import (
	"server/app/service"
	"server/app/util/context"
)

// Status 查询网站进程状态。
func Status(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Site.Process(form.Id, "status")
	if err != nil {
		c.Error(err.Error())
		return
	}
	c.SuccessWithData("获取成功", data)
}
