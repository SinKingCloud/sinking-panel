package oss

import (
	"mime"
	"net/http"
	"path"
	"server/app/service"
	"server/app/util/context"
	utilOss "server/app/util/oss"
	"time"
)

// Sign 生成文件临时访问地址，下载签名使用附件响应头保留文件名。
func Sign(c *context.Context) {
	c.SetHeader("Cache-Control", "no-store")
	var form struct {
		Id       int64  `json:"id" validate:"required,min=1" label:"存储ID"`
		Key      string `json:"key" validate:"required,max=1024" label:"文件名称"`
		Method   string `json:"method" default:"GET" validate:"oneof=GET PUT HEAD" label:"访问方法"`
		Download bool   `json:"download" default:"false" label:"下载文件"`
		Expires  int    `json:"expires" default:"600" validate:"min=1,max=604800" label:"签名有效期"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.Download && form.Method != http.MethodGet {
		c.Error("只有 GET 签名支持下载文件")
		return
	}
	client, err := service.Oss.Client(form.Id)
	if err != nil {
		c.Error(err.Error())
		return
	}
	var options []utilOss.PresignOptions
	if form.Download {
		options = append(options, utilOss.PresignOptions{
			ContentDisposition: mime.FormatMediaType("attachment", map[string]string{"filename": path.Base(form.Key)}),
		})
	}
	data, err := client.Presign(c.Request.Context(), form.Key, form.Method, time.Duration(form.Expires)*time.Second, options...)
	if err != nil {
		c.Error(err.Error())
		return
	}
	c.SuccessWithData("获取成功", data)
}
