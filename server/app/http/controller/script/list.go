package script

import (
	"server/app/enum/log_type"
	repositoryScript "server/app/repository/script"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"strings"
)

// List 获取常用脚本列表
func List(c *context.Context) {
	query := c.ValidatePage("id", "desc", "id", "id,module,type_id,name,create_time,update_time")
	var form struct {
		Module  string `json:"module" default:"script_ssh" validate:"required,max=50" label:"所属模块"`
		TypeId  string `json:"type_id" default:"" validate:"omitempty,numeric" label:"类型ID"`
		Keyword string `json:"keyword" default:"" validate:"omitempty" label:"关键词"`
		Name    string `json:"name" default:"" validate:"omitempty,max=100" label:"脚本名称"`
		Script  string `json:"script" default:"" validate:"omitempty,max=131072" label:"脚本内容"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if strings.TrimSpace(form.Module) == "" {
		c.Error("所属模块不能为空")
		return
	}
	where := &repositoryScript.SelectScript{Module: &form.Module}
	if form.TypeId != "" {
		typeId, err := strconv.ParseInt(form.TypeId, 10, 64)
		if err != nil || typeId < 0 {
			c.Error("类型ID参数错误")
			return
		}
		where.TypeId = &typeId
	}
	if form.Keyword != "" {
		where.Keyword = &form.Keyword
	}
	if form.Name != "" {
		where.Name = &form.Name
	}
	if form.Script != "" {
		where.Script = &form.Script
	}
	data, err := service.Script.Select(where, query)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看常用脚本", "查看常用脚本列表")
		c.SuccessWithData("获取成功", data)
	}
}
