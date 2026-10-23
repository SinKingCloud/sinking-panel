package elasticsearch

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// RefreshFields 后台重新处理历史文档，使新增子字段的映射生效。
func (Controller) RefreshFields(c *context.Context) {
	var form struct {
		Id           int64  `json:"id" validate:"required,min=1" label:"实例ID"`
		DatabaseName string `json:"database_name" validate:"required,max=255" label:"索引名称"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	taskID, err := service.Database.RefreshElasticsearchFields(form.Id, form.DatabaseName)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "Elasticsearch 回填文档", "实例["+strconv.FormatInt(form.Id, 10)+"] 索引["+form.DatabaseName+"]")
	c.SuccessWithData("已加入后台任务", taskID)
}
