package redirect

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// List 获取网站重定向规则。
func List(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Site.GetRedirects(form.Id)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看网站重定向", "查看网站["+strconv.FormatInt(form.Id, 10)+"]重定向配置")
		c.SuccessWithData("获取成功", data)
	}
}
