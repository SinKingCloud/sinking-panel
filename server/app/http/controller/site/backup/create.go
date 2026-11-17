package backup

import (
	"fmt"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"

	"github.com/SinKingCloud/sinking-go/sinking-web"
)

// Create 将所选网站加入同一个备份任务。
func Create(c *context.Context) {
	var form struct {
		SiteIds []int64 `json:"site_ids" validate:"required,min=1,max=1000,dive,min=1" label:"网站ID列表"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	taskId, err := service.Site.CreateBackup(form.SiteIds)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "创建网站备份", fmt.Sprintf("备份网站%v", form.SiteIds))
	c.SuccessWithData("备份任务已提交", sinking_web.H{"task_id": taskId})
}
