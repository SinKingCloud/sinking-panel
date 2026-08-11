package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"server/app/util/file"
	"server/app/util/page"
)

func List(c *context.Context) {
	var form struct {
		Path         string `json:"path" default:"/" validate:"required" label:"目录"`
		Keyword      string `json:"keyword" default:"" validate:"omitempty,max=255" label:"关键词"`
		Page         int    `json:"page" default:"1" validate:"numeric,min=1" label:"分页页码"`
		PageSize     int    `json:"page_size" default:"20" validate:"numeric,min=1,max=1000" label:"分页容量"`
		OrderByField string `json:"order_by_field" default:"" validate:"omitempty,oneof=name size update_time" label:"排序字段"`
		OrderByType  string `json:"order_by_type" default:"" validate:"omitempty,oneof=asc desc" label:"排序类型"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	f := file.NewDisk("")
	list, total, err := f.FileListWithPage(form.Path, form.Page, form.PageSize, form.OrderByField, form.OrderByType, form.Keyword)
	if err != nil {
		c.Error("获取失败")
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看文件列表", "查看目录["+form.Path+"]文件列表")
	c.SuccessWithData("获取成功", page.New(&page.Query{
		Page:         form.Page,
		PageSize:     form.PageSize,
		OrderByField: form.OrderByField,
		OrderByType:  form.OrderByType,
	}, total, list))
}
