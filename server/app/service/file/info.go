package file

import (
	"errors"
	"fmt"
	"server/app/util/file"
	"time"
)

// Info 获取文件元数据，可按游标读取指定版本的文本内容。
func (s *service) Info(path string, read bool, cursor int64, pageSize int, version string) (map[string]interface{}, error) {
	f := file.NewDisk("")
	if !f.Exists(path) {
		return nil, errors.New("该目录或文件不存在")
	}
	info, err := f.FileInfo(path)
	if err != nil {
		return nil, errors.New("获取失败")
	}
	data := map[string]interface{}{
		"name":        info.Name,
		"size":        info.Size,
		"mode":        info.Mode,
		"is_dir":      info.IsDir,
		"update_time": time.Unix(info.UpdateTime, 0).Format("2006-01-02 15:04:05"),
	}
	if read {
		content, nextCursor, eof, currentVersion, err := f.GetFileContent(path, cursor, pageSize, version)
		if err != nil {
			return nil, fmt.Errorf("读取文件内容失败: %w", err)
		}
		data["content"] = content
		data["cursor"] = cursor
		data["next_cursor"] = nextCursor
		data["eof"] = eof
		data["version"] = currentVersion
	}
	return data, nil
}
