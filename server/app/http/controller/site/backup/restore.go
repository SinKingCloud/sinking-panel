package backup

import (
	"fmt"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"

	"github.com/SinKingCloud/sinking-go/sinking-web"
)

// Restore 提交网站备份恢复任务。
func Restore(c *context.Context) {
	var form struct {
		Id           int64 `json:"id" default:"0" validate:"required,min=1" label:"备份ID"`
		TargetSiteId int64 `json:"target_site_id" default:"0" validate:"min=0" label:"目标网站ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	taskId, err := service.Site.RestoreBackup(form.Id, form.TargetSiteId)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "恢复网站备份", fmt.Sprintf("恢复备份[%d]，目标网站[%d]", form.Id, form.TargetSiteId))
	c.SuccessWithData("恢复任务已提交", sinking_web.H{"task_id": taskId})
}
