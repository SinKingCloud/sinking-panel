package script

import (
	"server/app/enum/log_type"
	repositoryScript "server/app/repository/script"
	"server/app/service"
	"server/app/util/context"
)

// List 获取常用脚本列表
func List(c *context.Context) {
	query := c.ValidatePage("id", "desc", "id", "id,type_id,name,create_time,update_time")
	var form struct {
		TypeId  string `json:"type_id" default:"" validate:"omitempty,numeric" label:"类型ID"`
		Keyword string `json:"keyword" default:"" validate:"omitempty" label:"关键词"`
		Name    string `json:"name" default:"" validate:"omitempty,max=100" label:"脚本名称"`
		Script  string `json:"script" default:"" validate:"omitempty,max=131072" label:"脚本内容"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	where := &repositoryScript.SelectScript{}
	if form.TypeId != "" {
		where.TypeId = form.TypeId
	}
	if form.Keyword != "" {
		where.Keyword = form.Keyword
	}
	if form.Name != "" {
		where.Name = form.Name
	}
	if form.Script != "" {
		where.Script = form.Script
	}
	data, err := service.Script.Select(where, query)
	if err != nil {
		c.Error("获取失败")
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看常用脚本", "查看常用脚本列表")
		c.SuccessWithData("获取成功", data)
	}
}
