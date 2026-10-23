package file

import (
	"server/app/service"
	"server/app/util/context"
)

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
		c.SuccessWithData("文件已上传完成", map[string]interface{}{
			"upload_id":       meta.UploadID,
			"total_chunks":    meta.TotalChunks,
			"uploaded_chunks": uploadedChunks,
			"completed":       true,
			"file":            completed,
		})
		return
	}
	c.SuccessWithData("获取分片状态成功", map[string]interface{}{
		"upload_id":       meta.UploadID,
		"total_chunks":    meta.TotalChunks,
		"uploaded_chunks": uploadedChunks,
	})
}
