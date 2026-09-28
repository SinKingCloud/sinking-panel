package mysql

import (
	"path/filepath"
	"strings"

	"server/app/enum/log_type"
	"server/app/service"
	serviceSoftware "server/app/service/software/mysql"
	"server/app/util/context"
)

// Image 解析本机镜像，返回数据库安装参数，不启动容器。
func Image(c *context.Context) {
	m := service.Software.Mysql()
	var form struct {
		ImagePath string `json:"image_path" validate:"required" label:"镜像路径"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	source := strings.TrimSpace(form.ImagePath)
	if filepath.IsAbs(source) {
		image, err := service.Container.Import(source)
		if err != nil {
			c.Error(err.Error())
			return
		}
		source = image.Id
	}
	data, err := m.Image(source)
	if err != nil {
		c.Error(err.Error())
		return
	}
	c.SuccessWithData("解析成功", data)
}

// Install 安装数据库，初始化账号通过专用文件交给官方镜像。
func Install(c *context.Context) {
	m := service.Software.Mysql()
	var form struct {
		ImagePath string `json:"image_path" default:"" label:"镜像路径"`
		ImageID   string `json:"image_id" default:"" label:"镜像 ID"`
		Username  string `json:"username" default:"" label:"管理员"`
		Password  string `json:"password" default:"" label:"密码"`
		Bind      string `json:"bind" default:"127.0.0.1" label:"绑定地址"`
		Port      int    `json:"port" default:"0" validate:"min=0,max=65535" label:"端口"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	source := form.ImageID
	if source == "" {
		source = form.ImagePath
	}
	data, err := m.Install(c.Request.Context(), source, &serviceSoftware.InstallOptions{
		Username: form.Username, Password: form.Password, Bind: form.Bind, Port: form.Port,
	})
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "安装数据库服务", m.InstanceID)
	c.SuccessWithData("安装成功", data)
}

// Status 返回容器状态、数据库就绪情况和负载指标。
func Status(c *context.Context) {
	m := service.Software.Mysql()
	data, err := m.Status(c.Request.Context())
	if err != nil {
		c.Error(err.Error())
		return
	}
	c.SuccessWithData("获取成功", data)
}

// Start 使用统一容器服务启动数据库，连接就绪情况由状态接口查询。
func Start(c *context.Context) {
	m := service.Software.Mysql()
	if _, err := service.Container.Start(c.Request.Context(), m.InstanceID); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "启动数据库服务", m.InstanceID)
	c.Success("启动成功")
}

// Stop 停止数据库并暂停自动恢复。
func Stop(c *context.Context) {
	m := service.Software.Mysql()
	if err := service.Container.Stop(c.Request.Context(), m.InstanceID); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "停止数据库服务", m.InstanceID)
	c.Success("停止成功")
}

// Restart 使用统一容器服务重启数据库，进程启动时读取已保存的配置。
func Restart(c *context.Context) {
	m := service.Software.Mysql()
	if err := service.Container.Restart(c.Request.Context(), m.InstanceID); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "重启数据库服务", m.InstanceID)
	c.Success("重启成功")
}

// Uninstall 默认保留数据库文件，只有显式选择时一并删除数据。
func Uninstall(c *context.Context) {
	m := service.Software.Mysql()
	var form struct {
		RemoveData bool `json:"remove_data" default:"false" label:"删除数据"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if err := m.Uninstall(c.Request.Context(), form.RemoveData); err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "卸载数据库服务", m.InstanceID)
	c.Success("卸载成功")
}

// Performance 共用表单协议，各软件自行定义和校验配置字段。
func Performance(c *context.Context) {
	m := service.Software.Mysql()
	var form struct {
		Action string                 `json:"action" default:"get" validate:"oneof=get set" label:"操作"`
		Config map[string]interface{} `json:"config" label:"配置"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "get" {
		data, err := m.GetPerformance()
		if err != nil {
			c.Error(err.Error())
			return
		}
		c.SuccessWithData("获取成功", data)
		return
	}
	if len(form.Config) == 0 {
		c.Error("配置不能为空")
		return
	}
	data, err := m.UpdatePerformance(c.Request.Context(), form.Config)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改数据库性能配置", m.InstanceID)
	c.SuccessWithData("配置已保存", data)
}

// Config 读取或保存原始配置文件，与可视化表单使用同一份配置。
func Config(c *context.Context) {
	m := service.Software.Mysql()
	var form struct {
		Action  string `json:"action" default:"get" validate:"oneof=get set" label:"操作"`
		Content string `json:"content" default:"" label:"配置内容"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "get" {
		content, err := m.GetConfig()
		if err != nil {
			c.Error(err.Error())
			return
		}
		c.SuccessWithData("获取成功", map[string]interface{}{"content": content})
		return
	}
	data, err := m.UpdateConfig(c.Request.Context(), form.Content)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改数据库配置文件", m.InstanceID)
	c.SuccessWithData("配置已保存", data)
}

// Reload 尝试重载当前软件支持的配置，必须重启的项会明确返回。
func Reload(c *context.Context) {
	m := service.Software.Mysql()
	data, err := m.Reload(c.Request.Context())
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "重载数据库配置", m.InstanceID)
	c.SuccessWithData("处理成功", data)
}

// Log 直接复用容器日志的游标读取和清理能力。
func Log(c *context.Context) {
	m := service.Software.Mysql()
	var form struct {
		Action   string `json:"action" default:"read" validate:"oneof=read clear" label:"操作"`
		After    int64  `json:"after" default:"0" validate:"numeric,min=0" label:"增量游标"`
		Before   int64  `json:"before" default:"0" validate:"numeric,min=0" label:"历史游标"`
		PageSize int    `json:"page_size" default:"300" validate:"min=1,max=10000" label:"读取行数"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Action == "clear" {
		if err := service.Container.ClearLog(m.InstanceID); err != nil {
			c.Error(err.Error())
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "清理数据库日志", m.InstanceID)
		c.Success("清理成功")
		return
	}
	data, err := service.Container.ReadLog(m.InstanceID, form.After, form.Before, form.PageSize)
	if err != nil {
		c.Error(err.Error())
		return
	}
	c.SuccessWithData("获取成功", data)
}
