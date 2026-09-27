package container

import (
	"server/app/enum/log_type"
	"server/app/service"
	runtimeContainer "server/app/util/container"
	"server/app/util/context"
)

// Run 创建并启动容器，未覆盖的启动参数由镜像提供。
func Run(c *context.Context) {
	var form struct {
		Id          string                          `json:"id" validate:"omitempty,max=64" label:"容器ID"`
		Name        string                          `json:"name" validate:"omitempty,max=100" label:"容器名称"`
		ImageId     string                          `json:"image_id" validate:"required" label:"镜像"`
		Mounts      []runtimeContainer.Mount        `json:"mounts" label:"目录挂载"`
		Env         []string                        `json:"env" label:"环境变量"`
		Command     []string                        `json:"command" label:"启动命令"`
		WorkingDir  string                          `json:"working_dir" label:"工作目录"`
		User        string                          `json:"user" label:"运行用户"`
		Resources   runtimeContainer.ResourceLimits `json:"resources" label:"资源限制"`
		AutoRestart bool                            `json:"auto_restart" label:"自动重启"`
		ReadOnly    bool                            `json:"read_only" label:"只读文件系统"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Container.Run(runtimeContainer.RunOptions{
		ID: form.Id, Name: form.Name, ImageID: form.ImageId,
		Mounts: form.Mounts, Env: form.Env, Command: form.Command,
		WorkingDir: form.WorkingDir, User: form.User, Resources: form.Resources,
		AutoRestart: form.AutoRestart, ReadOnly: form.ReadOnly, CreateOnly: true,
	})
	if err != nil {
		c.Error(err.Error())
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "创建容器", "创建容器["+data.ID+"]")
		c.SuccessWithData("创建成功", data)
	}
}
