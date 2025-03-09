package file

import (
	"server/app/service"
	"server/app/util/server"
)

func Disk(c *server.Context) {
	paths, err := service.File.GetDiskPaths()
	if err != nil {
		c.Error("获取失败")
		return
	}
	c.SuccessWithData("获取成功", paths)
}
