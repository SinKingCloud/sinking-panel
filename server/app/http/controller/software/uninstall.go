package software

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"

	"github.com/SinKingCloud/sinking-go/sinking-web"
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
	taskId, err := service.Software.Uninstall(strconv.FormatInt(form.Id, 10))
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "卸载软件", "卸载软件并清理插件数据["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("卸载任务已提交", sinking_web.H{"task_id": taskId})
}
