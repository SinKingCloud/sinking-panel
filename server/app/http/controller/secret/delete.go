package secret

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Delete 删除未被对象存储或证书引用的密钥。
func Delete(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"密钥ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	err := service.Secret.Delete(form.Id)
	if err == nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "删除密钥", "删除密钥["+strconv.FormatInt(form.Id, 10)+"]")
		c.Success("删除成功")
	} else {
		c.Error(err.Error())
	}
}
