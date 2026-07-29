package file

import (
	"server/app/util/context"
	"server/app/util/file"
	"server/app/util/page"
)

func List(c *context.Context) {
	var form struct {
		Path         string `json:"path" default:"/" validate:"required" label:"目录"`
		Page         int    `json:"page" default:"1" validate:"numeric,min=1" label:"分页页码"`
		PageSize     int    `json:"page_size" default:"20" validate:"numeric,min=1,max=1000" label:"分页容量"`
		OrderByField string `json:"order_by_field" default:"name" validate:"required,oneof=name size update_time" label:"排序字段"`
		OrderByType  string `json:"order_by_type" default:"asc" validate:"required,oneof=asc desc" label:"排序类型"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	f := file.NewDisk("")
	list, total, err := f.FileListWithPage(form.Path, form.Page, form.PageSize, form.OrderByField, form.OrderByType)
	if err != nil {
		c.Error("获取失败")
		return
	}
	c.SuccessWithData("获取成功", page.New(&page.Query{
		Page:         form.Page,
		PageSize:     form.PageSize,
		OrderByField: form.OrderByField,
		OrderByType:  form.OrderByType,
	}, total, list))
}
