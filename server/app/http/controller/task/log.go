package task

import (
	"server/app/service"
	"server/app/util/context"
)

func Log(c *context.Context) {
	query := c.ValidatePage("id", "asc", "id")
	if query.CursorId != "" || query.CursorLastId != "" {
		c.Error("任务日志不支持游标分页")
		return
	}
	if query.IsCursor() {
		query.Page = 1
	}
	var form struct {
		Id int64 `json:"id" default:"" validate:"numeric,min=1" label:"记录ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Task.FindById(form.Id)
	if err != nil || data == nil {
		c.Error("获取失败")
	} else {
		c.SuccessWithData("获取成功", service.Task.ReadLog(data.Id, query.Page, query.PageSize))
	}
}
