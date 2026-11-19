package recycle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"server/app/util/file"
	"strings"
)

// Create 检查整批路径后将文件或目录移入回收站。
func (s *service) Create(names ...string) error {
	if len(names) == 0 {
		return errors.New("文件路径列表不能为空")
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	f := file.NewDisk("")
	root := f.Path(s.path, true, true)
	if err := os.MkdirAll(root, 0755); err != nil {
		return err
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	contains := func(dir, path string) bool {
		for {
			if dir == path {
				return true
			}
			// 同一机器上的卷也可能有不同的大小写规则，以实际文件身份为准。
			if strings.EqualFold(dir, path) {
				a, _ := os.Lstat(dir)
				b, _ := os.Lstat(path)
				return os.SameFile(a, b)
			}
			parent := filepath.Dir(path)
			if parent == path {
				return false
			}
			path = parent
		}
	}
	resolvedParents := make(map[string]string)
	for _, name := range names {
		path := f.Path(name, true, true)
		info, err := os.Lstat(path)
		if name == "" || err != nil {
			return fmt.Errorf("该目录或文件不存在: %s", name)
		}
		// 移动的是文件或链接本身，只解析父目录，不跟随末级链接。
		dir := filepath.Dir(path)
		parent, ok := resolvedParents[dir]
		if !ok {
			parent, err = filepath.EvalSymlinks(dir)
			if err != nil {
				return err
			}
			resolvedParents[dir] = parent
		}
		realPath := filepath.Join(parent, filepath.Base(path))
		alias := path != realPath || root != realRoot
		if contains(root, path) || alias && contains(realRoot, realPath) {
			return errors.New("回收站目录及其内容必须选择彻底删除")
		}
		if (info.IsDir() || info.Mode()&os.ModeSymlink != 0) && (contains(path, root) || alias && contains(realPath, realRoot)) {
			return errors.New("不能将包含回收站的目录移入回收站")
		}
	}
	for _, name := range names {
		newName := s.encodeName(f.Path(name, true, true), 0, 0)
		if err := f.Rename(name, filepath.Join(root, newName)); err != nil {
			return fmt.Errorf("移动至回收站失败: %w", err)
		}
	}
	return nil
}
