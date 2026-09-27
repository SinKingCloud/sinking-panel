package container

import (
	"strings"
	"time"

	"server/app/enum/log_type"
	"server/app/service"
	runtimeContainer "server/app/util/container"
	"server/app/util/context"
)

// Exec 执行一次命令并返回退出状态；命令、环境变量和输出不写入操作日志。
func Exec(c *context.Context) {
	var form struct {
		Id         string   `json:"id" validate:"required,max=64" label:"容器ID"`
		Command    []string `json:"command" validate:"required,min=1" label:"执行命令"`
		Env        []string `json:"env" label:"环境变量"`
		WorkingDir string   `json:"working_dir" label:"工作目录"`
		Timeout    int      `json:"timeout" default:"30" validate:"min=1,max=3600" label:"超时秒数"`
		MaxOutput  int64    `json:"max_output" default:"524288" validate:"min=1,max=16777216" label:"输出字节上限"`
		Stdin      string   `json:"stdin" label:"标准输入"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Container.Exec(form.Id, runtimeContainer.ExecOptions{
		Command: form.Command, Env: form.Env, WorkingDir: form.WorkingDir,
		Timeout: time.Duration(form.Timeout) * time.Second, MaxOutput: form.MaxOutput,
		Stdin: strings.NewReader(form.Stdin),
	})
	if err != nil && data == nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "执行容器命令", "在容器["+form.Id+"]中执行命令")
	if err != nil {
		c.SuccessWithData(err.Error(), data)
	} else {
		c.SuccessWithData("执行完成", data)
	}
}
