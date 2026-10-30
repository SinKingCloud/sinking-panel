package software

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"

	"github.com/SinKingCloud/sinking-go/sinking-web"
)

// Install 安装服务器上的软件 ZIP 包或源码目录。
func Install(c *context.Context) {
	var form struct {
		Path string `json:"path" default:"" validate:"required,max=4096" label:"软件包路径"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	taskID, err := service.Software.Install(form.Path)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "安装软件", "安装软件["+form.Path+"]")
	c.SuccessWithData("安装任务已提交", sinking_web.H{"task_id": taskID})
}
