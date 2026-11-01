package oss

import (
	"net/http"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	utilOss "server/app/util/oss"
	"strconv"
	"time"
)

// transferDeadline 为文件传输单独放宽超时，不改变其他接口的超时设置。
func transferDeadline(c *context.Context) {
	deadline := time.Now().Add(15 * time.Minute)
	controller := http.NewResponseController(c.Writer)
	_ = controller.SetReadDeadline(deadline)
	_ = controller.SetWriteDeadline(deadline)
}

// Upload 统一处理本机上传、分片上传和服务器后台上传，action 从查询参数读取。
func Upload(c *context.Context) {
	query := c.Request.URL.Query()
	action := query.Get("action")
	switch action {
	case "server":
		var form struct {
			Id        int64       `json:"id" validate:"required,min=1" label:"存储ID"`
			Key       string      `json:"key" validate:"required,max=1024" label:"文件名称"`
			Path      string      `json:"path" validate:"required,max=4096" label:"服务器文件路径"`
			Overwrite bool        `json:"overwrite" default:"false" label:"覆盖同名文件"`
			Options   fileOptions `json:"options"`
		}
		if ok, msg := c.ValidatorAll(&form); !ok {
			c.Error(msg)
			return
		}
		taskID, err := service.Oss.Upload(form.Id, form.Key, form.Path, form.Overwrite, form.Options.putOptions())
		if err != nil {
			c.Error(err.Error())
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "创建 OSS 传输任务", "OSS["+strconv.FormatInt(form.Id, 10)+"] upload "+form.Key+"，文件["+form.Path+"]")
		c.SuccessWithData("已加入后台任务", taskID)
		return
	case "init", "complete", "abort", "parts", "uploads":
		uploadMultipart(c, action)
		return
	case "", "file", "part":
	default:
		c.Error("上传操作类型不合法")
		return
	}

	transferDeadline(c)
	id, err := strconv.ParseInt(query.Get("id"), 10, 64)
	key := query.Get("key")
	if err != nil || id <= 0 || key == "" {
		c.Error("存储 ID 或文件名称不合法")
		return
	}
	var number int
	if action == "part" {
		number, err = strconv.Atoi(query.Get("number"))
		if err != nil || number < 1 || number > 10000 || query.Get("upload_id") == "" {
			c.Error("上传会话或分片编号不合法")
			return
		}
		if c.Request.ContentLength <= 0 || c.Request.ContentLength > utilOss.MaxPartSize {
			c.Error("分片必须提供大于 0 且不超过 5 GiB 的 Content-Length")
			return
		}
	} else if c.Request.ContentLength < 0 || c.Request.ContentLength > utilOss.MaxFileSize {
		c.Error("上传文件必须提供 Content-Length，大小不能超过 5 TiB")
		return
	}
	client, err := service.Oss.Client(id)
	if err != nil {
		c.Error(err.Error())
		return
	}
	defer c.Request.Body.Close()
	var data interface{}
	if action == "part" {
		data, err = client.UploadPart(c.Request.Context(), key, query.Get("upload_id"), number, c.Request.Body, c.Request.ContentLength)
	} else {
		data, err = client.Upload(c.Request.Context(), key, c.Request.Body, c.Request.ContentLength, utilOss.UploadOptions{
			PutOptions: utilOss.PutOptions{ContentType: c.Request.Header.Get("Content-Type")},
		})
	}
	if err != nil {
		c.Error(err.Error())
		return
	}
	if action == "part" {
		c.SuccessWithData("分片上传成功", data)
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "上传 OSS 文件", "OSS["+strconv.FormatInt(id, 10)+"] "+key)
	c.SuccessWithData("上传成功", data)
}
