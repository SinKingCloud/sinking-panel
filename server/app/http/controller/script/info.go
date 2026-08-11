package script

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Info 获取常用脚本详情
func Info(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"" validate:"numeric,min=1" label:"记录ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Script.FindById(form.Id)
	if err == nil && data != nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看脚本详情", "查看脚本["+data.Name+"]详情")
		c.SuccessWithData("获取成功", data)
	} else {
		c.Error("获取失败")
	}
}
