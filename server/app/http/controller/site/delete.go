package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Delete 删除网站。
func Delete(c *context.Context) {
	var form struct {
		Id         int64 `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		DeleteRoot bool  `json:"delete_root" default:"false" label:"删除网站根目录"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	err := service.Site.Delete(form.Id, form.DeleteRoot)
	if err != nil {
		c.Error(err.Error())
	} else {
		detail := "删除网站[" + strconv.FormatInt(form.Id, 10) + "]"
		if form.DeleteRoot {
			detail += "，同时删除网站根目录"
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "删除网站", detail)
		c.Success("删除成功")
	}
}
