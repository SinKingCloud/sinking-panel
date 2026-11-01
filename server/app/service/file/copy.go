package file

import (
	"context"
	"errors"
	"server/app/util/file"
)

// CopyTask 创建批量文件复制任务。
func (s *service) CopyTask(paths []string, target string) (string, error) {
	return s.transferTask(paths, target, false)
}

// transferTask 校验并创建复制或移动任务，提交后使用独立的路径列表。
func (s *service) transferTask(paths []string, target string, move bool) (string, error) {
	if len(paths) == 0 || target == "" {
		return "", errors.New("源文件路径和目标路径不能为空")
	}
	f := file.NewDisk("")
	for _, path := range paths {
		if path == "" || !f.Exists(path) {
			return "", errors.New("源文件或目录不存在: " + path)
		}
	}
	paths = append([]string(nil), paths...)
	action, transfer := "复制", s.CopyWithContext
	if move {
		action, transfer = "移动", s.MoveWithContext
	}
	data := map[string]interface{}{"source_paths": paths, "target_path": target}
	return s.createTask(action, action+"文件: "+paths[0]+" -> "+target, data, func(ctx context.Context, update func(float64, string)) error {
		for index, path := range paths {
			err := transfer(ctx, path, target, func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool {
				if ctx.Err() != nil {
					return false
				}
				progress := float64(index) / float64(len(paths)) * 100
				if total > 0 {
					progress = (float64(index) + float64(current)/float64(total)) / float64(len(paths)) * 100
				}
				update(progress, "正在"+action+": "+currentFile)
				return true
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}
