package oss

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	utilOss "server/app/util/oss"
	"strconv"
)

// uploadMultipart 管理云端分片会话，操作由 Upload 的查询参数指定。
func uploadMultipart(c *context.Context, action string) {
	var form struct {
		Id             int64          `json:"id" validate:"required,min=1" label:"存储ID"`
		Key            string         `json:"key" validate:"omitempty,max=1024" label:"文件名称"`
		UploadId       string         `json:"upload_id" validate:"omitempty,max=4096" label:"上传会话"`
		Options        fileOptions    `json:"options"`
		Parts          []utilOss.Part `json:"parts" validate:"omitempty,max=10000" label:"分片列表"`
		Marker         int            `json:"marker" default:"0" validate:"min=0,max=10000" label:"分片游标"`
		Limit          int            `json:"limit" default:"1000" validate:"min=1,max=1000" label:"每页数量"`
		Prefix         string         `json:"prefix" validate:"omitempty,max=1024" label:"文件前缀"`
		KeyMarker      string         `json:"key_marker" validate:"omitempty,max=1024" label:"文件游标"`
		UploadIdMarker string         `json:"upload_id_marker" validate:"omitempty,max=4096" label:"会话游标"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if action != "uploads" && form.Key == "" {
		c.Error("请选择文件")
		return
	}
	if action != "init" && action != "uploads" && form.UploadId == "" {
		c.Error("请选择上传会话")
		return
	}
	client, err := service.Oss.Client(form.Id)
	if err != nil {
		c.Error(err.Error())
		return
	}
	ctx := c.Request.Context()
	var data interface{}
	switch action {
	case "init":
		data, err = client.InitiateMultipartUpload(ctx, form.Key, form.Options.putOptions())
	case "complete":
		transferDeadline(c)
		data, err = client.CompleteMultipartUpload(ctx, form.Key, form.UploadId, form.Parts)
	case "abort":
		err = client.AbortMultipartUpload(ctx, form.Key, form.UploadId)
	case "parts":
		data, err = client.ListParts(ctx, form.Key, form.UploadId, utilOss.PartsOptions{Marker: form.Marker, Limit: form.Limit})
	case "uploads":
		data, err = client.ListMultipartUploads(ctx, utilOss.UploadsOptions{
			Prefix: form.Prefix, KeyMarker: form.KeyMarker, UploadIDMarker: form.UploadIdMarker, Limit: form.Limit,
		})
	}
	if err != nil {
		c.Error(err.Error())
		return
	}
	if action == "complete" || action == "abort" {
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "OSS 分片上传", "OSS["+strconv.FormatInt(form.Id, 10)+"] "+action+" "+form.Key)
	}
	c.SuccessWithData("操作成功", data)
}
