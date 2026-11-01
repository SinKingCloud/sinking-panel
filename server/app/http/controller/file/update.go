package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	serviceFile "server/app/service/file"
	"server/app/util/context"
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
	data, err := service.File.Update(form.Path, &serviceFile.UpdateFile{
		Name:        form.Name,
		Permissions: form.Permissions,
		Content:     form.Content,
		Cursor:      form.Cursor,
		NextCursor:  form.NextCursor,
		Version:     form.Version,
	})
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改文件", "修改文件["+data.Path+"]")
	c.SuccessWithData("修改成功", data)
}
