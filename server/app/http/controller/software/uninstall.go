package software

import (
	"net/http"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"
)

// Uninstall 卸载软件并清理插件数据。
func Uninstall(c *context.Context) {
	var form struct {
		Id int64 `json:"id" default:"0" validate:"required,min=1" label:"软件ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if err := http.NewResponseController(c.Writer).SetWriteDeadline(time.Now().Add(11 * time.Minute)); err != nil {
		c.Error("当前连接不支持卸载软件: " + err.Error())
		return
	}
	taskId, err := service.Software.Uninstall(strconv.FormatInt(form.Id, 10))
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "卸载软件", "卸载软件并清理插件数据["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("卸载成功", map[string]int64{"task_id": taskId})
}
