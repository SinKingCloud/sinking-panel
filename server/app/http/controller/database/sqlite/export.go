package sqlite

import (
	stdContext "context"
	"server/app/enum/database_type"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Export 创建导出后台任务，文件内容仅在服务器流式处理。
func (Controller) Export(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=1024" label:"数据库名称"`
		Path         string `json:"path" validate:"required,max=4096" label:"文件路径"`
		Schema       string `json:"schema" validate:"omitempty,max=1024" label:"数据库架构"`
		Table        string `json:"table" validate:"omitempty,max=1024" label:"数据表"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	taskID, err := service.Database.Transfer(form.Id, database_type.Sqlite, "export", form.Path, form.DatabaseName, func(ctx stdContext.Context, progress func(float64, string)) (map[string]interface{}, error) {
		manager, err := service.Database.Sqlite(ctx, form.Id, form.DatabaseName)
		if err != nil {
			return nil, err
		}
		defer manager.Close()
		return manager.Export(ctx, form.Path, form.Schema, form.Table, progress)
	})
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "Sqlite 导出数据库", "实例["+strconv.FormatInt(form.Id, 10)+"]")
	c.SuccessWithData("已加入后台任务", taskID)
}
