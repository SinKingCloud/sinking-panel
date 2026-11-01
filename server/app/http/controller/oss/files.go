package oss

import (
	"server/app/service"
	"server/app/util/context"
	"server/app/util/page"
)

// Files 使用云厂商游标分页读取文件和目录前缀。
func Files(c *context.Context) {
	var form struct {
		Id        int64  `json:"id" validate:"required,min=1" label:"存储ID"`
		Prefix    string `json:"prefix" validate:"omitempty,max=1024" label:"目录前缀"`
		Delimiter string `json:"delimiter" default:"/" validate:"omitempty,oneof=/" label:"目录分隔符"`
		CursorId  string `json:"cursor_id" validate:"omitempty,max=32768" label:"分页游标"`
		PageSize  int    `json:"page_size" default:"20" validate:"min=1,max=1000" label:"分页容量"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Oss.Files(c.Request.Context(), form.Id, form.Prefix, form.Delimiter, &page.Query{
		CursorId: form.CursorId,
		PageSize: form.PageSize,
	})
	if err != nil {
		c.Error(err.Error())
		return
	}
	c.SuccessWithData("获取成功", data)
}
