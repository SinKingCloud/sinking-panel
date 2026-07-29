package recycle

import (
	"server/app/service"
	"server/app/util/context"
	"server/app/util/page"
)

func List(c *context.Context) {
	query := c.ValidatePage("name", "asc", "name,size,update_time")
	if query.CursorId != "" || query.CursorLastId != "" {
		c.Error("回收站列表不支持游标分页")
		return
	}
	if query.IsCursor() {
		query.Page = 1
	}
	list, total, err := service.Recycle.Select(query.Page, query.PageSize, query.OrderByField, query.OrderByType)
	if err != nil {
		c.Error("获取失败")
		return
	}
	c.SuccessWithData("获取成功", page.New(query, total, list))
}
