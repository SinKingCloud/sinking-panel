//go:build darwin || freebsd

package ipc

import (
	"errors"

	"golang.org/x/sys/unix"
)

func socketPeerUID(fd int) (uint32, error) {
	credential, err := unix.GetsockoptXucred(fd, unix.SOL_LOCAL, unix.LOCAL_PEERCRED)
	if err != nil {
		return 0, err
	}
	// Darwin、FreeBSD 的 XUCRED_VERSION 均为 0，未知布局不继续认证。
	if credential.Version != 0 {
		return 0, errors.New("IPC 对端凭据版本不支持")
	}
	return credential.Uid, nil
}
