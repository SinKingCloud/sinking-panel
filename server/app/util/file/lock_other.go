//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !windows

package file

import "os"

func (d *Disk) lockFile(_ *os.File, _ bool) (func(), error) {
	return func() {}, nil
}
