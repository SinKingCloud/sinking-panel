package file

import (
	"path/filepath"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"server/app/util/file"
	"strconv"
)

// Update 更新文件
func Update(c *context.Context) {
	var form struct {
		Path        string  `json:"path" default:"" validate:"required" label:"文件路径"`
		Name        string  `json:"name" default:"" label:"新文件名"`
		Permissions string  `json:"permissions" default:"" label:"权限"`
		Content     *string `json:"content" default:"" validate:"omitempty" label:"文件内容"`
		Cursor      *int64  `json:"cursor" default:"" validate:"omitempty,min=0" label:"文件游标"`
		NextCursor  *int64  `json:"next_cursor" default:"" validate:"omitempty,min=0" label:"下一文件游标"`
		Version     string  `json:"version" default:"" validate:"omitempty" label:"文件版本"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	f := file.NewDisk("")
	if !f.Exists(form.Path) {
		c.Error("文件或目录不存在")
		return
	}
	var operations []string
	var nextCursor int64
	var eof bool
	var version string
	var err error
	currentPath := form.Path
	if form.Name != "" {
		dir := filepath.Dir(currentPath)
		newPath := filepath.Join(dir, form.Name)
		if f.Exists(newPath) {
			c.Error("目标文件或目录已存在")
			return
		}
		err := f.Rename(currentPath, newPath)
		if err != nil {
			c.Error("重命名失败: " + err.Error())
			return
		}
		operations = append(operations, "重命名")
		currentPath = newPath
	}
	if form.Permissions != "" {
		permissions, err := strconv.ParseUint(form.Permissions, 8, 32)
		if err != nil {
			c.Error("无效的权限值，请提供有效的八进制数字")
			return
		}
		if permissions > 0777 {
			c.Error("权限值超出范围，有效范围为 0-777")
			return
		}
		err = f.SetFileMode(currentPath, uint32(permissions))
		if err != nil {
			c.Error("修改权限失败: " + err.Error())
			return
		}
		operations = append(operations, "修改权限")
	}

	if form.Content != nil {
		if !f.IsFile(currentPath) {
			c.Error("只能更新文件内容，不能更新目录")
			return
		}
		nextCursor, eof, version, err = f.UpdateFileContent(currentPath, *form.Content, form.Cursor, form.NextCursor, form.Version)
		if err != nil {
			c.Error("更新文件内容失败: " + err.Error())
			return
		}
		operations = append(operations, "更新内容")
	}
	if len(operations) == 0 {
		c.Error("请指定要执行的操作：重命名、修改权限或更新内容")
		return
	}
	data := map[string]interface{}{
		"operations": operations,
		"path":       currentPath,
	}
	if form.Content != nil {
		data["cursor"] = int64(0)
		if form.Cursor != nil {
			data["cursor"] = *form.Cursor
		}
		data["next_cursor"] = nextCursor
		data["eof"] = eof
		data["version"] = version
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改文件", "修改文件["+currentPath+"]")
	c.SuccessWithData("修改成功", data)
}
