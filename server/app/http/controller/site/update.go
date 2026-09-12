package site

import (
	"strconv"

	"server/app/enum/log_type"
	"server/app/service"
	serviceSite "server/app/service/site"
	"server/app/util/context"
)

// Update 更新网站基础信息，未传入的字段保持原值。
func Update(c *context.Context) {
	var form struct {
		Id      int64   `json:"id" default:"0" validate:"required,min=1" label:"网站ID"`
		Name    *string `json:"name" validate:"omitempty,max=100" label:"网站名称"`
		TypeId  *int64  `json:"type_id" validate:"omitempty,min=0" label:"网站分类ID"`
		Root    *string `json:"root" validate:"omitempty,max=4096" label:"网站根目录"`
		RunPath *string `json:"run_path" validate:"omitempty,max=4096" label:"网站运行目录"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &serviceSite.UpdateSite{}
	if form.Name != nil {
		data.Name = form.Name
	}
	if form.TypeId != nil {
		data.TypeId = form.TypeId
	}
	if form.Root != nil {
		data.Root = form.Root
	}
	if form.RunPath != nil {
		data.RunPath = form.RunPath
	}
	err := service.Site.Update(form.Id, data)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改网站", "修改网站["+strconv.FormatInt(form.Id, 10)+"]")
		c.Success("修改成功")
	}
}
