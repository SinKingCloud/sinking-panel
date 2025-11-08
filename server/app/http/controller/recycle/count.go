package recycle

import (
	"server/app/service"
	"server/app/util/context"
)

// Count 统计信息
func Count(c *context.Context) {
	totalSize, fileCount, dirCount, err := service.Recycle.Count()
	if err != nil {
		c.Error("获取失败")
		return
	}
	c.SuccessWithData("获取成功", map[string]int64{
		"size": totalSize,
		"file": fileCount,
		"dir":  dirCount,
	})
}
