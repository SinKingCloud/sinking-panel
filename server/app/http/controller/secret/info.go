package secret

import (
	"encoding/json"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Info 获取密钥详情及供编辑使用的脱敏字段。
func Info(c *context.Context) {
	var form struct {
		Id       int64  `json:"id" default:"0" validate:"required,min=1" label:"密钥ID"`
		Provider string `json:"provider" default:"" validate:"omitempty,numeric" label:"服务商"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Secret.FindById(form.Id)
	if err == nil && data != nil {
		if form.Provider != "" {
			provider, err := strconv.Atoi(form.Provider)
			if err != nil || provider < 0 {
				c.Error("服务商参数错误")
				return
			}
			fields, ok := service.Secret.GetData()[provider]
			if !ok {
				c.Error("服务商不合法")
				return
			}
			if provider != data.Provider {
				content, _ := json.Marshal(fields)
				data.Provider = provider
				data.Data = map[string]string{}
				_ = json.Unmarshal(content, &data.Data)
			}
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看密钥详情", "查看密钥["+strconv.FormatInt(form.Id, 10)+"]详情")
		c.SuccessWithData("获取成功", data)
	} else {
		c.Error("获取失败")
	}
}
