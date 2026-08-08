package config

import (
	"server/app/constant"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strings"
)

// Set 修改配置
func Set(c *context.Context) {
	type Config struct {
		Key   string `json:"key" default:"" validate:"required,max=100" label:"配置标识"`
		Value string `json:"value" default:"" validate:"omitempty" label:"配置内容"`
	}
	var form struct {
		Configs []*Config `json:"configs" default:"" validate:"gte=1" label:"配置标识"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	configs := make(map[string]string)
	for _, v := range form.Configs {
		if v == nil || v.Key == constant.LoginGroup || strings.HasPrefix(v.Key, constant.LoginGroup+".") {
			continue
		}
		configs[v.Key] = v.Value
	}
	if len(configs) > 0 {
		if err := service.Config.Sets(configs); err != nil {
			c.Error("修改数据失败")
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改系统配置", "修改系统配置数据")
	}
	c.Success("修改数据成功")
}
