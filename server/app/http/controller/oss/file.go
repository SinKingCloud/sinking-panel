package oss

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	utilOss "server/app/util/oss"
	"strconv"
)

// File 查看文件信息、创建目录、删除、复制、移动或修改文件。
func File(c *context.Context) {
	var form struct {
		Id          int64       `json:"id" validate:"required,min=1" label:"存储ID"`
		Action      string      `json:"action" validate:"required,oneof=info mkdir delete copy move metadata" label:"操作类型"`
		Key         string      `json:"key" validate:"omitempty,max=1024" label:"文件名称"`
		Keys        []string    `json:"keys" validate:"omitempty,min=1,max=1000,dive,required,max=1024" label:"文件列表"`
		Destination string      `json:"destination" validate:"omitempty,max=1024" label:"目标名称"`
		Options     fileOptions `json:"options"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Key == "" && (form.Action != "delete" || len(form.Keys) == 0) {
		c.Error("请选择文件")
		return
	}
	if (form.Action == "copy" || form.Action == "move") && form.Destination == "" {
		c.Error("请输入目标文件名称")
		return
	}
	if form.Key != "" && len(form.Keys) > 0 {
		c.Error("文件名称与文件列表不能同时提交")
		return
	}
	var client utilOss.Interface
	var err error
	if form.Action != "delete" && form.Action != "copy" && form.Action != "move" {
		client, err = service.Oss.Client(form.Id)
		if err != nil {
			c.Error(err.Error())
			return
		}
	}
	ctx := c.Request.Context()
	var data interface{}
	switch form.Action {
	case "info":
		data, err = client.StatFile(ctx, form.Key)
	case "mkdir":
		data, err = client.CreateDirectory(ctx, form.Key)
	case "delete":
		keys := form.Keys
		if len(keys) == 0 {
			keys = []string{form.Key}
		}
		var result []utilOss.DeleteResult
		result, err = service.Oss.Delete(ctx, form.Id, keys)
		items := make([]deleteResult, 0, len(result))
		for _, item := range result {
			value := deleteResult{Key: item.Key}
			if item.Err != nil {
				value.Error = item.Err.Error()
			}
			items = append(items, value)
		}
		data = items
	case "copy":
		transferDeadline(c)
		data, err = service.Oss.Copy(ctx, form.Id, form.Key, form.Destination, utilOss.CopyOptions{})
	case "move":
		transferDeadline(c)
		data, err = service.Oss.Move(ctx, form.Id, form.Key, form.Destination)
	case "metadata":
		transferDeadline(c)
		data, err = client.SetMetadata(ctx, form.Key, form.Options.putOptions())
	}
	if err != nil {
		c.Error(err.Error())
		return
	}
	if form.Action != "info" {
		event := log_type.EventUpdate
		if form.Action == "delete" {
			event = log_type.EventDelete
		}
		service.Log.Create(c.GetRequestIp(), event, "操作 OSS 文件", "OSS["+strconv.FormatInt(form.Id, 10)+"] "+form.Action+" "+form.Key)
	}
	c.SuccessWithData("操作成功", data)
}
