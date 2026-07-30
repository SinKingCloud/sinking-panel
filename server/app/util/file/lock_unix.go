//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package file

import (
	"os"
	"syscall"
)

func (d *Disk) lockFile(file *os.File, exclusive bool) (func(), error) {
	mode := syscall.LOCK_SH
	if exclusive {
		mode = syscall.LOCK_EX
	}
	if err := syscall.Flock(int(file.Fd()), mode); err != nil {
		return nil, err
	}
	return func() {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	}, nil
}
