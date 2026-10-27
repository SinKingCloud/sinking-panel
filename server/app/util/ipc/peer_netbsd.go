package ipc

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/unix"
)

func socketPeerUID(fd int) (uint32, error) {
	// NetBSD sys/un.h：SOL_LOCAL=0、LOCAL_PEEREID=3，返回 struct unpcbid。
	var credential struct {
		PID int32
		UID uint32
		GID uint32
	}
	size := uint32(unsafe.Sizeof(credential))
	_, _, errno := unix.Syscall6(unix.SYS_GETSOCKOPT, uintptr(fd), 0, 3,
		uintptr(unsafe.Pointer(&credential)), uintptr(unsafe.Pointer(&size)), 0)
	if errno != 0 {
		return 0, errno
	}
	if size != uint32(unsafe.Sizeof(credential)) {
		return 0, errors.New("IPC 对端凭据长度不正确")
	}
	return credential.UID, nil
}
