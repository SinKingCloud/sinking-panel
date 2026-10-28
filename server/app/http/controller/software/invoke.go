package software

import (
	"net/http"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"time"
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
	// 容器镜像准备和数据库首次启动可能超过面板默认的一分钟写超时。
	switch route.Method {
	case "setup", "start", "stop", "restart", "reload":
		if err := http.NewResponseController(c.Writer).SetWriteDeadline(time.Now().Add(11 * time.Minute)); err != nil {
			c.Error("当前连接不支持长时间软件操作: " + err.Error())
			return
		}
	}
	if err := service.Software.Invoke(c, route.Code, route.Method); err != nil {
		c.Error(err.Error())
		return
	}
	// 不记录业务参数，避免日志泄漏数据库密码等敏感内容。
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "调用软件接口", "调用软件["+route.Code+"]接口["+route.Method+"]")
}
