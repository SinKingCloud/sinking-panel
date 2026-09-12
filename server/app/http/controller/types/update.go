package types

import (
	"server/app/enum/log_type"
	repositoryTypes "server/app/repository/types"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Update 修改信息
func Update(c *context.Context) {
	var form struct {
		Ids    []int64 `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
		Module string  `json:"module" default:"" validate:"omitempty,max=50" label:"所属模块"`
		Name   string  `json:"name" default:"" validate:"omitempty,max=50" label:"类型名称"`
		Sort   string  `json:"sort" default:"" validate:"omitempty,numeric" label:"排序"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &repositoryTypes.UpdateType{}
	if form.Module != "" {
		data.Module = &form.Module
	}
	if form.Name != "" {
		data.Name = &form.Name
	}
	if form.Sort != "" {
		sort, err := strconv.ParseInt(form.Sort, 10, 64)
		if err != nil {
			c.Error("排序参数错误")
			return
		}
		data.Sort = &sort
	}
	err := service.Type.UpdateByIds(form.Ids, data)
	if err == nil {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改类型", "修改类型数据")
		c.Success("修改成功")
	} else {
		c.Error("修改失败")
	}
}
