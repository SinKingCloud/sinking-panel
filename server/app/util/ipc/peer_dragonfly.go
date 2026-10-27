package ipc

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/unix"
)

func socketPeerUID(fd int) (uint32, error) {
	// LOCAL_PEERCRED=1，只读取 xucred 的固定前缀；组列表不参与身份校验。
	var credential struct {
		Version uint32
		UID     uint32
	}
	size := uint32(unsafe.Sizeof(credential))
	_, _, errno := unix.Syscall6(unix.SYS_GETSOCKOPT, uintptr(fd), 0, 1,
		uintptr(unsafe.Pointer(&credential)), uintptr(unsafe.Pointer(&size)), 0)
	if errno != 0 {
		return 0, errno
	}
	if size != uint32(unsafe.Sizeof(credential)) || credential.Version != 0 {
		return 0, errors.New("IPC 对端凭据格式不支持")
	}
	return credential.UID, nil
}
