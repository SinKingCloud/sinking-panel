package software

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
)

// Invoke 将原始请求交给插件控制器，由插件绑定参数、校验并返回响应。
func Invoke(c *context.Context) {
	var route struct {
		Code   string `json:"code" validate:"required,max=100" label:"软件标识"`
		Method string `json:"method" validate:"required,max=100" label:"操作方法"`
	}
	if ok, msg := c.ValidatorPath(&route); !ok {
		c.Error(msg)
		return
	}
	if err := service.Software.Invoke(c, route.Code, route.Method); err != nil {
		c.Error(err.Error())
		return
	}
	// 不记录业务参数，避免日志泄漏数据库密码等敏感内容。
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "调用软件接口", "调用软件["+route.Code+"]接口["+route.Method+"]")
}
