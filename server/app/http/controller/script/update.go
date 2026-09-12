package script

import (
	"server/app/enum/log_type"
	repositoryScript "server/app/repository/script"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Update 修改常用脚本
func Update(c *context.Context) {
	var form struct {
		Ids    []int64 `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
		TypeId string  `json:"type_id" default:"" validate:"omitempty,numeric" label:"类型ID"`
		Name   string  `json:"name" default:"" validate:"omitempty,max=100" label:"脚本名称"`
		Script string  `json:"script" default:"" validate:"omitempty,max=131072" label:"脚本内容"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &repositoryScript.UpdateScript{}
	if form.TypeId != "" {
		typeId, err := strconv.ParseInt(form.TypeId, 10, 64)
		if err != nil || typeId < 0 {
			c.Error("类型ID参数错误")
			return
		}
		data.TypeId = &typeId
	}
	if form.Name != "" {
		data.Name = &form.Name
	}
	if form.Script != "" {
		data.Script = &form.Script
	}
	err := service.Script.UpdateByIds(form.Ids, data)
	if err != nil {
		c.Error("修改失败")
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改常用脚本", "修改常用脚本数据")
		c.Success("修改成功")
	}
}
