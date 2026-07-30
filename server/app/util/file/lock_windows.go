//go:build windows

package file

import (
	"os"

	"golang.org/x/sys/windows"
)

func (d *Disk) lockFile(file *os.File, exclusive bool) (func(), error) {
	var flags uint32
	if exclusive {
		flags = windows.LOCKFILE_EXCLUSIVE_LOCK
	}
	overlapped := &windows.Overlapped{}
	if err := windows.LockFileEx(windows.Handle(file.Fd()), flags, 0, 0xffffffff, 0xffffffff, overlapped); err != nil {
		return nil, err
	}
	return func() {
		_ = windows.UnlockFileEx(windows.Handle(file.Fd()), 0, 0xffffffff, 0xffffffff, overlapped)
	}, nil
}
