package ipc

import "golang.org/x/sys/unix"

func socketPeerUID(fd int) (uint32, error) {
	// sockpeercred 的首字段为 uint32 UID，getsockopt 按缓冲区长度返回前缀。
	// 使用 libc 包装，避免 OpenBSD 7.5 起禁止的间接系统调用。
	uid, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_PEERCRED)
	return uint32(uid), err
}
