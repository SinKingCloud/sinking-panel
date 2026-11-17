package config

import (
	"strings"

	"server/app/constant"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Set 修改配置
func Set(c *context.Context) {
	var form struct {
		Configs []*Config `json:"configs" default:"" validate:"required,min=1,max=1000,dive,required" label:"配置标识"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	configs := make(map[string]string, len(form.Configs))
	for _, v := range form.Configs {
		group := v.Key
		if index := strings.IndexByte(group, '.'); index > 0 {
			group = group[:index]
		}
		policy, exists := constant.SensitiveGroups[group]
		if !exists {
			c.Error("配置不存在")
			return
		}
		if policy.Write {
			c.Error("该配置不允许通过通用接口修改")
			return
		}
		configs[v.Key] = v.Value
	}
	if err := service.Config.Sets(configs); err != nil {
		c.Error("修改数据失败")
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改系统配置", "修改系统配置数据")
	c.Success("修改数据成功")
}
