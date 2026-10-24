package redis

import (
	stdContext "context"
	"server/app/enum/database_kind"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Export 创建导出后台任务，文件内容仅在服务器流式处理。
func (Controller) Export(c *context.Context) {
	var form struct {
		Id           int64   `json:"id" validate:"required,min=1" label:"实例ID"`
		DatabaseName string  `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Path         string  `json:"path" validate:"required,max=4096" label:"文件路径"`
		Match        string  `json:"match" validate:"omitempty,max=65536" label:"键名匹配"`
		Key          *string `json:"key" validate:"omitempty,max=65536" label:"键名"`
		Overwrite    bool    `json:"overwrite" default:"false" label:"覆盖同名文件"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	taskID, err := service.Database.Transfer(form.Id, database_kind.Redis, "export", form.Path, form.DatabaseName, func(ctx stdContext.Context, progress func(float64, string)) (map[string]interface{}, error) {
		manager, err := service.Database.Redis(ctx, form.Id, form.DatabaseName)
		if err != nil {
			return nil, err
		}
		defer manager.Close()
		return manager.Export(ctx, form.Path, form.Match, form.Key, form.Overwrite, progress)
	})
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "Redis 导出数据", "实例["+strconv.FormatInt(form.Id, 10)+"]，文件["+form.Path+"]")
	c.SuccessWithData("已加入后台任务", taskID)
}
