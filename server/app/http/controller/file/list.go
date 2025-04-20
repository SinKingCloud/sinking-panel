package file

import (
	"server/app/util/file"
	"server/app/util/page"
	"server/app/util/server"
)

func List(c *server.Context) {
	pageInfo := page.ValidatePageDefault(c)
	type Form struct {
		OrderByField string `json:"order_by_field" default:"name" validate:"oneof=name size update_time" label:"排序字段"`
		OrderByType  string `json:"order_by_type" default:"asc" validate:"oneof=desc asc" label:"排序类型"`
		Path         string `json:"path" default:"/" validate:"required" label:"目录"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	f := file.NewDisk("")
	list, total, err := f.FileListWithPage(form.Path, pageInfo.Page, pageInfo.PageSize, form.OrderByField, form.OrderByType)
	if err != nil {
		c.Error("获取失败")
		return
	}
	c.SuccessWithData("获取成功", page.NewPage(total, pageInfo.Page, pageInfo.PageSize, list))
}
