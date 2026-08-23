package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	siteService "server/app/service/site"
	"server/app/util/context"
)

// Update 更新网站基础信息，未传入的字段保持原值。
func Update(c *context.Context) {
	var form struct {
		Id      int64   `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Name    *string `json:"name" validate:"omitempty,max=100" label:"网站名称"`
		Root    *string `json:"root" validate:"omitempty,max=4096" label:"网站根目录"`
		RunPath *string `json:"run_path" validate:"omitempty,max=4096" label:"网站运行目录"`
	}
	if ok, message := c.ValidatorAll(&form); !ok {
		c.Error(message)
		return
	}
	if err := service.Site.Update(form.Id, &siteService.UpdateSite{
		Name:    form.Name,
		Root:    form.Root,
		RunPath: form.RunPath,
	}); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站", "修改网站["+strconv.FormatInt(form.Id, 10)+"]")
	c.Success("修改成功")
}
