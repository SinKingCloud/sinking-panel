package file

import (
	"net/http"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"time"

	"github.com/SinKingCloud/sinking-go/sinking-web"
)

// Upload 文件上传处理
func Upload(c *context.Context) {
	// 大文件上传和合并单独放宽超时，不影响其他接口的一分钟限制。
	deadline := time.Now().Add(15 * time.Minute)
	controller := http.NewResponseController(c.Writer)
	_ = controller.SetReadDeadline(deadline)
	_ = controller.SetWriteDeadline(deadline)
	switch c.DefaultQuery("action", "upload") {
	case "upload":
		handleFileUpload(c)
	case "merge":
		handleMergeChunks(c)
	case "check":
		handleCheckChunks(c)
	case "clear":
		handleClearUpload(c)
	default:
		c.Error("不支持的操作类型")
	}
}

// handleFileUpload 处理普通文件和分片上传。
func handleFileUpload(c *context.Context) {
	// 超出 256 KB 的文件内容写入临时文件，避免每个分片在内存中反复扩容。
	if err := c.Request.ParseMultipartForm(256 << 10); err != nil {
		c.Error("读取上传文件失败: " + err.Error())
		return
	}
	defer c.Request.MultipartForm.RemoveAll()
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.Error("获取上传文件失败: " + err.Error())
		return
	}

	if c.DefaultForm("upload_id", "") == "" {
		var form struct {
			Path      string `json:"path" default:"/" validate:"required" label:"上传目录"`
			Overwrite bool   `json:"overwrite" default:"true" validate:"omitempty" label:"覆盖同名文件"`
		}
		if ok, msg := c.ValidatorPost(&form); !ok {
			c.Error(msg)
			return
		}
		data, uploadErr := service.File.Upload(c.Request.Context(), fileHeader, form.Path, form.Overwrite)
		if uploadErr != nil {
			c.Error(uploadErr.Error())
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "上传文件", "上传文件["+data.Path+"]")
		c.SuccessWithData("文件上传成功", sinking_web.H{
			"name": data.Name,
			"path": data.Path,
			"size": data.Size,
		})
		return
	}

	meta, err := parseUploadMeta(
		c.DefaultForm("upload_id", ""),
		c.DefaultForm("path", "/"),
		c.DefaultForm("file_name", ""),
		c.DefaultForm("total_size", ""),
		c.DefaultForm("chunk_size", ""),
		c.DefaultForm("total_chunks", ""),
		c.DefaultForm("file_hash", ""),
	)
	if err != nil {
		c.Error(err.Error())
		return
	}
	chunkIndex, err := strconv.Atoi(c.DefaultForm("chunk_index", ""))
	if err != nil {
		c.Error("chunk_index参数不合法")
		return
	}
	completed, exists, err := service.File.UploadChunk(fileHeader, meta, chunkIndex)
	if err != nil {
		c.Error(err.Error())
		return
	}
	if completed != nil {
		c.SuccessWithData("文件已上传完成", completed)
		return
	}
	c.SuccessWithData("分片上传成功", sinking_web.H{
		"upload_id":   meta.UploadID,
		"chunk_index": chunkIndex,
		"file_name":   meta.FileName,
		"exists":      exists,
	})
}

// handleMergeChunks 合并上传分片。
func handleMergeChunks(c *context.Context) {
	meta, err := parseUploadMeta(
		c.DefaultForm("upload_id", ""),
		c.DefaultForm("path", "/"),
		c.DefaultForm("file_name", ""),
		c.DefaultForm("total_size", ""),
		c.DefaultForm("chunk_size", ""),
		c.DefaultForm("total_chunks", ""),
		c.DefaultForm("file_hash", ""),
	)
	if err != nil {
		c.Error(err.Error())
		return
	}
	completed, existed, err := service.File.MergeUpload(c.Request.Context(), meta)
	if err != nil {
		c.Error(err.Error())
	} else {
		if !existed {
			service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "合并上传文件", "上传文件["+completed.Path+"]")
		}
		c.SuccessWithData("文件合并成功", completed)
	}
}

// handleCheckChunks 查询分片上传状态。
func handleCheckChunks(c *context.Context) {
	meta, err := parseUploadMeta(
		c.DefaultQuery("upload_id", ""),
		c.DefaultQuery("path", "/"),
		c.DefaultQuery("file_name", ""),
		c.DefaultQuery("total_size", ""),
		c.DefaultQuery("chunk_size", ""),
		c.DefaultQuery("total_chunks", ""),
		c.DefaultQuery("file_hash", ""),
	)
	if err != nil {
		c.Error(err.Error())
		return
	}
	uploadedChunks, completed, err := service.File.CheckUpload(meta)
	if err != nil {
		c.Error(err.Error())
		return
	}
	if completed != nil {
		c.SuccessWithData("文件已上传完成", sinking_web.H{
			"upload_id":       meta.UploadID,
			"total_chunks":    meta.TotalChunks,
			"uploaded_chunks": uploadedChunks,
			"completed":       true,
			"file":            completed,
		})
		return
	}
	c.SuccessWithData("获取分片状态成功", sinking_web.H{
		"upload_id":       meta.UploadID,
		"total_chunks":    meta.TotalChunks,
		"uploaded_chunks": uploadedChunks,
	})
}

// handleClearUpload 清理上传会话。
func handleClearUpload(c *context.Context) {
	err := service.File.ClearUpload(c.DefaultForm("upload_id", ""))
	if err != nil {
		c.Error(err.Error())
	} else {
		c.Success("清理上传缓存成功")
	}
}
