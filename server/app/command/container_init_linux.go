//go:build linux && cgo

package command

import (
	"github.com/opencontainers/runc/libcontainer"
	_ "github.com/opencontainers/runc/libcontainer/nsenter"
)

// ExecuteInit 处理 libcontainer 重新进入容器命名空间时使用的 init 命令。
// 该判断必须发生在创建守护进程和加载应用之前，避免子进程误走面板启动流程。
func ExecuteInit(args []string) bool {
	if len(args) == 0 || args[0] != "init" {
		return false
	}
	libcontainer.Init()
	return true
}
