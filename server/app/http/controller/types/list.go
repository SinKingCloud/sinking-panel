package types

import (
	"server/app/enum/log_type"
	repositoryTypes "server/app/repository/types"
	"server/app/service"
	"server/app/util/context"
)

// List 获取类型列表
func List(c *context.Context) {
	query := c.ValidatePage("sort", "asc", "id,sort", "id,module,name,sort,create_time,update_time")
	var form struct {
		Module string `json:"module" default:"" validate:"omitempty,max=50" label:"所属模块"`
		Name   string `json:"name" default:"" validate:"omitempty,max=50" label:"类型名称"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	where := &repositoryTypes.SelectType{}
	if form.Module != "" {
		where.Module = form.Module
	}
	if form.Name != "" {
		where.Name = form.Name
	}
	data, err := service.Type.Select(where, query)
	if err != nil {
		c.Error("获取失败")
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看类型列表", "查看类型列表数据")
		c.SuccessWithData("获取成功", data)
	}
}
