package file

import (
	"server/app/service"
	"server/app/util/context"
)

func Disk(c *context.Context) {
	paths, err := service.File.GetDiskPaths()
	if err != nil {
		c.Error("获取失败")
	} else {
		c.SuccessWithData("获取成功", paths)
	}
}
