package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Count 统计信息
func Count(c *context.Context) {
	var form struct {
		Path string `json:"path" default:"" validate:"required" label:"文件路径"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.File.Count(form.Path)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "统计文件数据", "统计文件或目录["+form.Path+"]数据")
		c.SuccessWithData("获取成功", data)
	}
}
