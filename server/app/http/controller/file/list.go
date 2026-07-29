package file

import (
	"server/app/util/context"
	"server/app/util/file"
	"server/app/util/page"
)

func List(c *context.Context) {
	query := c.ValidatePage("name", "asc", "name,size,update_time")
	if query.CursorId != "" || query.CursorLastId != "" {
		c.Error("文件列表不支持游标分页")
		return
	}
	if query.IsCursor() {
		query.Page = 1
	}
	var form struct {
		Path string `json:"path" default:"/" validate:"required" label:"目录"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	f := file.NewDisk("")
	list, total, err := f.FileListWithPage(form.Path, query.Page, query.PageSize, query.OrderByField, query.OrderByType)
	if err != nil {
		c.Error("获取失败")
		return
	}
	c.SuccessWithData("获取成功", page.New(query, total, list))
}
