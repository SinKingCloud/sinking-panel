package recycle

import (
	"server/app/service"
	"server/app/util/context"

	"github.com/SinKingCloud/sinking-go/sinking-web"
)

// Count 统计信息
func Count(c *context.Context) {
	totalSize, fileCount, dirCount, err := service.Recycle.Count()
	if err != nil {
		c.Error("获取失败")
	} else {
		c.SuccessWithData("获取成功", sinking_web.H{
			"size": totalSize,
			"file": fileCount,
			"dir":  dirCount,
		})
	}
}
