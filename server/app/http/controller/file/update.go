package file

import (
	"os"
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
		file, err := f.OpenFile(currentPath, os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			c.Error("打开文件失败: " + err.Error())
			return
		}
		defer file.Close()
		_, err = file.WriteString(*form.Content)
		if err != nil {
			c.Error("写入文件失败: " + err.Error())
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
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改文件", "修改文件["+currentPath+"]")
	c.SuccessWithData("修改成功", data)
}
