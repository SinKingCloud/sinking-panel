//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package ipc

import (
	"errors"
	"fmt"
	"net"
)

// peerUID 从已连接套接字读取内核记录的对端身份，不接受请求中自报的 UID。
func peerUID(conn *net.UnixConn) (uint32, error) {
	raw, err := conn.SyscallConn()
	if err != nil {
		return 0, fmt.Errorf("读取 IPC 对端身份失败: %w", err)
	}
	var uid uint32
	var credentialErr error
	if err = raw.Control(func(fd uintptr) {
		uid, credentialErr = socketPeerUID(int(fd))
	}); err != nil {
		return 0, fmt.Errorf("读取 IPC 对端身份失败: %w", err)
	}
	if credentialErr != nil {
		return 0, fmt.Errorf("读取 IPC 对端身份失败: %w", credentialErr)
	}
	if uid == ^uint32(0) {
		return 0, errors.New("IPC 对端身份无效")
	}
	return uid, nil
}
