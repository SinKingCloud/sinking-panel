package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Compress 创建压缩任务。
func Compress(c *context.Context) {
	var form struct {
		Paths  []string `json:"paths" default:"" validate:"required" label:"文件路径"`
		Dir    string   `json:"dir" default:"" validate:"required" label:"目标目录"`
		Format string   `json:"format" default:"zip" validate:"required" label:"压缩格式"`
		Name   string   `json:"name" default:"" validate:"required" label:"文件名"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	taskID, err := service.File.CompressTask(form.Paths, form.Dir, form.Name, form.Format)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "创建压缩任务", "压缩文件: "+form.Name+"."+form.Format)
	c.SuccessWithData("创建压缩任务成功", taskID)
}
