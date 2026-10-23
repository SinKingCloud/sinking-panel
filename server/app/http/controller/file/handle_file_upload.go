package file

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

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
		c.SuccessWithData("文件上传成功", map[string]interface{}{
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
	c.SuccessWithData("分片上传成功", map[string]interface{}{
		"upload_id":   meta.UploadID,
		"chunk_index": chunkIndex,
		"file_name":   meta.FileName,
		"exists":      exists,
	})
}
