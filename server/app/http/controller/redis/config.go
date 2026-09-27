package redis

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"

	"github.com/SinKingCloud/sinking-go/sinking-web"
)

// Config 读取或修改 Redis 配置文件。
func Config(c *context.Context) {
	var form struct {
		Action  string `json:"action" default:"get" validate:"required,oneof=get set" label:"操作类型"`
		Content string `json:"content" default:"" validate:"omitempty" label:"配置内容"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "get" {
		content, err := service.Container.Redis().GetConfig()
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看 Redis 配置", "查看 Redis 配置文件")
			c.SuccessWithData("获取成功", sinking_web.H{"content": content})
		}
	} else {
		if form.Content == "" {
			c.Error("配置内容不能为空")
			return
		}
		data, err := service.Container.Redis().UpdateConfig(c.Request.Context(), form.Content)
		if err != nil {
			c.Error(err.Error())
		} else {
			service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改 Redis 配置", "修改 Redis 配置文件")
			c.SuccessWithData("修改成功", data)
		}
	}
}
