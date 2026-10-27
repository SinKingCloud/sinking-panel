//go:build !windows && !darwin && !linux && !freebsd && !openbsd && !netbsd && !dragonfly

package ipc

import (
	"context"
	"fmt"
	"net"
	"runtime"
)

// Listen 返回不支持的平台错误，避免退回不安全的网络监听。
func Listen(string) (net.Listener, error) {
	return nil, fmt.Errorf("当前系统 %s 不支持本机插件通信", runtime.GOOS)
}

// Dial 保留不支持的平台错误，不将其误认为面板已停止。
func Dial(context.Context, string) (net.Conn, error) {
	return nil, fmt.Errorf("当前系统 %s 不支持本机插件通信", runtime.GOOS)
}
