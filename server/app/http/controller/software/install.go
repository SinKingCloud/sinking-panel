package software

import (
	"net/http"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"time"
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
	// 为安装、升级及失败回滚预留响应时间。
	if err := http.NewResponseController(c.Writer).SetWriteDeadline(time.Now().Add(11 * time.Minute)); err != nil {
		c.Error("当前连接不支持安装软件: " + err.Error())
		return
	}
	data, err := service.Software.Install(form.Path)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "安装软件", "安装软件["+form.Path+"]")
	c.SuccessWithData("安装成功", data)
}
