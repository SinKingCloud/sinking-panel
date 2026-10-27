//go:build linux && cgo

package command

import "server/app/util/container"

// executeInit 处理 libcontainer 重新进入容器命名空间时使用的 init 命令。
// 该判断必须发生在创建守护进程和加载应用之前，避免子进程误走面板启动流程。
func (s *Server) executeInit(args []string) bool {
	if len(args) == 0 || args[0] != "init" {
		return false
	}
	container.Init()
	return true
}
