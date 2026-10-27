package plugin

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
)

// Open 只映射受管包内的路径，不复制源码，也不读取服务器 GOPATH。
func (m moduleFS) Open(name string) (fs.File, error) {
	name = filepath.ToSlash(name)
	if strings.HasPrefix(name, "/") || strings.Contains(name, ":") {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}
	name = path.Clean(name)
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}
	prefix := "src/" + m.module
	if name == prefix || strings.HasPrefix(name, prefix+"/") {
		name = "server" + strings.TrimPrefix(name, prefix)
	} else if strings.HasPrefix(name, "src/") || name == "src" {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrNotExist,
		}
	}
	return m.FS.Open(name)
}
