package script

import (
	"server/app/enum/log_type"
	repositoryScript "server/app/repository/script"
	"server/app/service"
	"server/app/util/context"
)

// Update 修改常用脚本
func Update(c *context.Context) {
	var form struct {
		Ids    []int64 `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
		TypeId *int64  `json:"type_id" default:"" validate:"omitempty,min=0" label:"类型ID"`
		Name   string  `json:"name" default:"" validate:"omitempty,max=100" label:"脚本名称"`
		Script string  `json:"script" default:"" validate:"omitempty,max=131072" label:"脚本内容"`
		Sort   *int    `json:"sort" default:"" validate:"omitempty,numeric" label:"排序"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &repositoryScript.UpdateScript{}
	if form.TypeId != nil {
		data.TypeId = *form.TypeId
	}
	if form.Name != "" {
		data.Name = form.Name
	}
	if form.Script != "" {
		data.Script = form.Script
	}
	if form.Sort != nil {
		data.Sort = *form.Sort
	}
	if err := service.Script.UpdateByIds(form.Ids, data); err != nil {
		c.Error("修改失败")
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改常用脚本", "修改常用脚本数据")
	c.Success("修改成功")
}
