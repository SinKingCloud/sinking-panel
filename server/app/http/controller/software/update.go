package software

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Update 修改软件显示名称。
func Update(c *context.Context) {
	var form struct {
		Id   int64  `json:"id" default:"0" validate:"required,min=1" label:"软件ID"`
		Name string `json:"name" default:"" validate:"required,max=100" label:"软件名称"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if err := service.Software.Update(form.Id, form.Name); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改软件", "修改软件["+strconv.FormatInt(form.Id, 10)+"]名称")
	c.Success("修改成功")
}
