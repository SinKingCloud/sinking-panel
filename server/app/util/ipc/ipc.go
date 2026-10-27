// Package ipc 使用本机套接字或命名管道通信，不监听 TCP 端口。
package ipc

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrUnavailable 表示没有进程监听；权限错误不归入此类，防止客户端绕过限制。
var ErrUnavailable = errors.New("本机通信服务未运行")

// endpoint 在通信目录尚未创建时也解析已有父目录的软链接。
func endpoint(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	parent, suffix := absolute, ""
	for {
		resolved, err := filepath.EvalSymlinks(parent)
		if err == nil {
			return filepath.Join(resolved, suffix), nil
		}
		if !errors.Is(err, os.ErrNotExist) || filepath.Dir(parent) == parent {
			return "", err
		}
		suffix = filepath.Join(filepath.Base(parent), suffix)
		parent = filepath.Dir(parent)
	}
}
