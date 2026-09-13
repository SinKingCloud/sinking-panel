package secret

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Info 获取密钥详情及供编辑使用的脱敏字段。
func Info(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"密钥ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Secret.FindById(form.Id)
	if err == nil && data != nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看密钥详情", "查看密钥["+strconv.FormatInt(form.Id, 10)+"]详情")
		c.SuccessWithData("获取成功", data)
	} else {
		c.Error("获取失败")
	}
}
