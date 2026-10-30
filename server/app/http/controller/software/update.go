package software

import (
	"server/app/enum/log_type"
	repositorySoftware "server/app/repository/software"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Update 修改软件名称和启用状态。
func Update(c *context.Context) {
	var form struct {
		Id     int64  `json:"id" default:"0" validate:"required,min=1" label:"软件ID"`
		Status *int   `json:"status" validate:"omitempty" label:"启用状态"`
		Name   string `json:"name" default:"" validate:"omitempty,max=100" label:"软件名称"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &repositorySoftware.UpdateSoftware{}
	if form.Status != nil {
		data.Status = form.Status
	}
	if form.Name != "" {
		data.Name = &form.Name
	}
	if err := service.Software.Update(form.Id, data); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改软件", "修改软件["+strconv.FormatInt(form.Id, 10)+"]")
	c.Success("修改成功")
}
