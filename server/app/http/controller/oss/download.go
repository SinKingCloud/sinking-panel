package oss

import (
	"io"
	"mime"
	"net/http"
	"path"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	utilOss "server/app/util/oss"
	"strconv"
)

// Download 统一处理本机流式下载和服务器后台下载，action 从查询参数读取。
func Download(c *context.Context) {
	query := c.Request.URL.Query()
	switch query.Get("action") {
	case "server":
		var form struct {
			Id        int64  `json:"id" validate:"required,min=1" label:"存储ID"`
			Key       string `json:"key" validate:"required,max=1024" label:"文件名称"`
			Path      string `json:"path" validate:"required,max=4096" label:"服务器文件路径"`
			Overwrite bool   `json:"overwrite" default:"false" label:"覆盖同名文件"`
		}
		if ok, msg := c.ValidatorAll(&form); !ok {
			c.Error(msg)
			return
		}
		taskID, err := service.Oss.Download(form.Id, form.Key, form.Path, form.Overwrite)
		if err != nil {
			c.Error(err.Error())
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "创建 OSS 传输任务", "OSS["+strconv.FormatInt(form.Id, 10)+"] download "+form.Key+"，文件["+form.Path+"]")
		c.SuccessWithData("已加入后台任务", taskID)
		return
	case "", "file":
	default:
		c.Error("下载操作类型不合法")
		return
	}

	transferDeadline(c)
	id, err := strconv.ParseInt(query.Get("id"), 10, 64)
	key := query.Get("key")
	if err != nil || id <= 0 || key == "" {
		c.Error("存储 ID 或文件名称不合法")
		return
	}
	client, err := service.Oss.Client(id)
	if err != nil {
		c.Error(err.Error())
		return
	}
	data, err := client.GetFile(c.Request.Context(), key, utilOss.GetOptions{Range: c.Request.Header.Get("Range")})
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer data.Body.Close()
	c.SetHeader("Content-Type", "application/octet-stream")
	c.SetHeader("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": path.Base(key)}))
	c.SetHeader("Cache-Control", "private, no-store")
	c.SetHeader("X-Content-Type-Options", "nosniff")
	if data.File.Size >= 0 {
		c.SetHeader("Content-Length", strconv.FormatInt(data.File.Size, 10))
	}
	status := http.StatusOK
	if data.ContentRange != "" {
		c.SetHeader("Content-Range", data.ContentRange)
		status = http.StatusPartialContent
	}
	c.Writer.WriteHeader(status)
	if _, err = io.Copy(c.Writer, data.Body); err != nil {
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "下载 OSS 文件", "OSS["+strconv.FormatInt(id, 10)+"] "+key)
}
