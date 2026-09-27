package redis

import (
	"server/app/enum/log_type"
	"server/app/service"
	serviceRedis "server/app/service/software/redis"
	"server/app/util/context"
)

// Install 导入本机镜像并创建 Redis 容器。
func Install(c *context.Context) {
	var form struct {
		ImagePath string                          `json:"image_path" default:"" label:"镜像路径"`
		ImageID   string                          `json:"image_id" default:"" label:"镜像 ID"`
		Config    *serviceRedis.PerformanceUpdate `json:"config" validate:"omitempty" label:"性能配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	source := form.ImagePath
	if form.ImageID != "" {
		source = form.ImageID
	}
	data, err := service.Software.Redis().Install(c.Request.Context(), source, form.Config)
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "安装 Redis", "导入本机镜像并创建 Redis 容器")
		c.SuccessWithData("安装成功", data)
	}
}
