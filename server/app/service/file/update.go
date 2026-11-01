package file

import (
	"errors"
	"fmt"
	"path/filepath"
	"server/app/util/file"
	"strconv"
)

// Update 修改文件名称、权限或内容，更新内容时保留游标与版本校验。
func (s *service) Update(path string, data *UpdateFile) (*UpdatedFile, error) {
	f := file.NewDisk("")
	if !f.Exists(path) {
		return nil, errors.New("文件或目录不存在")
	}
	if data == nil || (data.Name == "" && data.Permissions == "" && data.Content == nil) {
		return nil, errors.New("请指定要执行的操作：重命名、修改权限或更新内容")
	}
	var permissions uint64
	if data.Permissions != "" {
		value, err := strconv.ParseUint(data.Permissions, 8, 32)
		if err != nil {
			return nil, errors.New("无效的权限值，请提供有效的八进制数字")
		}
		if value > 0777 {
			return nil, errors.New("权限值超出范围，有效范围为 0-777")
		}
		permissions = value
	}
	if data.Content != nil && !f.IsFile(path) {
		return nil, errors.New("只能更新文件内容，不能更新目录")
	}
	result := &UpdatedFile{Path: path}
	if data.Name != "" {
		newPath := filepath.Join(filepath.Dir(path), data.Name)
		if f.Exists(newPath) {
			return nil, errors.New("目标文件或目录已存在")
		}
		if err := f.Rename(path, newPath); err != nil {
			return nil, fmt.Errorf("重命名失败: %w", err)
		}
		result.Operations = append(result.Operations, "重命名")
		result.Path = newPath
	}
	if data.Permissions != "" {
		if err := f.SetFileMode(result.Path, uint32(permissions)); err != nil {
			return nil, fmt.Errorf("修改权限失败: %w", err)
		}
		result.Operations = append(result.Operations, "修改权限")
	}
	if data.Content != nil {
		nextCursor, eof, version, err := f.UpdateFileContent(result.Path, *data.Content, data.Cursor, data.NextCursor, data.Version)
		if err != nil {
			return nil, fmt.Errorf("更新文件内容失败: %w", err)
		}
		cursor := int64(0)
		if data.Cursor != nil {
			cursor = *data.Cursor
		}
		result.Cursor, result.NextCursor, result.Eof, result.Version = &cursor, &nextCursor, &eof, &version
		result.Operations = append(result.Operations, "更新内容")
	}
	return result, nil
}
