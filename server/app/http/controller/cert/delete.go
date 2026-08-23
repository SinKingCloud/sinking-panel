package cert

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Delete 删除未被域名使用的证书。
func Delete(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"证书ID"`
	}
	if ok, message := c.ValidatorAll(&form); !ok {
		c.Error(message)
		return
	}
	if err := service.Site.DeleteCert(form.Id); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "删除网站证书", "删除网站证书["+strconv.FormatInt(form.Id, 10)+"]")
	c.Success("删除成功")
}
