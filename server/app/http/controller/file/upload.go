package file

import (
	"fmt"
	"path/filepath"
	"server/app/enum/log_type"
	"server/app/service"
	serviceFile "server/app/service/file"
	"server/app/util/context"
	"strconv"
	"strings"
)

// Upload 文件上传处理
func Upload(c *context.Context) {
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

func handleFileUpload(c *context.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.Error("获取上传文件失败: " + err.Error())
		return
	}

	if c.DefaultForm("upload_id", "") == "" {
		data, uploadErr := service.File.Upload(c.Request.Context(), fileHeader, c.DefaultForm("path", "/"))
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

func handleClearUpload(c *context.Context) {
	err := service.File.ClearUpload(c.DefaultForm("upload_id", ""))
	if err != nil {
		c.Error(err.Error())
	} else {
		c.Success("清理上传缓存成功")
	}
}

func parseUploadMeta(uploadID, path, fileName, totalSizeValue, chunkSizeValue, totalChunksValue, fileHash string) (serviceFile.UploadMeta, error) {
	totalSize, err := strconv.ParseInt(totalSizeValue, 10, 64)
	if err != nil {
		return serviceFile.UploadMeta{}, fmt.Errorf("total_size参数不合法")
	}
	chunkSize, err := strconv.ParseInt(chunkSizeValue, 10, 64)
	if err != nil {
		return serviceFile.UploadMeta{}, fmt.Errorf("chunk_size参数不合法")
	}
	totalChunks, err := strconv.Atoi(totalChunksValue)
	if err != nil {
		return serviceFile.UploadMeta{}, fmt.Errorf("total_chunks参数不合法")
	}
	return serviceFile.UploadMeta{
		UploadID:    uploadID,
		Path:        filepath.Clean(path),
		FileName:    fileName,
		TotalSize:   totalSize,
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
		FileHash:    strings.ToLower(fileHash),
	}, nil
}
