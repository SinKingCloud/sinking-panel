package script

import (
	"server/app/enum/log_type"
	repositoryScript "server/app/repository/script"
	"server/app/service"
	"server/app/util/context"
	"strings"
)

// Update 修改常用脚本
func Update(c *context.Context) {
	var form struct {
		Ids    []int64 `json:"ids" default:"" validate:"required,min=1,max=1000,unique" label:"ID列表"`
		Name   string  `json:"name" default:"" validate:"omitempty,max=100" label:"脚本名称"`
		Script string  `json:"script" default:"" validate:"omitempty,max=131072" label:"脚本内容"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data := &repositoryScript.UpdateScript{}
	if form.Name != "" {
		if strings.TrimSpace(form.Name) == "" {
			c.Error("脚本名称不能为空")
			return
		}
		data.Name = &form.Name
	}
	if form.Script != "" {
		if strings.TrimSpace(form.Script) == "" {
			c.Error("脚本内容不能为空")
			return
		}
		data.Script = &form.Script
	}
	err := service.Script.UpdateByIds(form.Ids, data)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改常用脚本", "修改常用脚本数据")
		c.Success("修改成功")
	}
}
