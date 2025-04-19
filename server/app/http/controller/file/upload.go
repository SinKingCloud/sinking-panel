package file

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"server/app/constant"
	"server/app/util/file"
	"server/app/util/server"
	"strconv"
	"strings"
	"time"
)

// ChunkInfo 分片信息
type ChunkInfo struct {
	UploadID    string `json:"upload_id"`    // 上传任务ID
	ChunkIndex  int    `json:"chunk_index"`  // 当前分片索引
	ChunkSize   int64  `json:"chunk_size"`   // 分片大小
	TotalSize   int64  `json:"total_size"`   // 文件总大小
	TotalChunks int    `json:"total_chunks"` // 总分片数
	FileName    string `json:"file_name"`    // 文件名
	FilePath    string `json:"file_path"`    // 保存路径
}

// Upload 文件上传处理
func Upload(c *server.Context) {
	// 获取请求参数
	action := c.DefaultQuery("action", "upload")
	time.Sleep(time.Second)

	switch action {
	case "upload":
		// 普通上传或分片上传
		handleFileUpload(c)
	case "merge":
		// 合并分片
		handleMergeChunks(c)
	case "check":
		// 检查分片状态
		handleCheckChunks(c)
	default:
		c.Error("不支持的操作类型")
	}
}

// handleFileUpload 处理文件上传
func handleFileUpload(c *server.Context) {
	uploadPath := c.DefaultForm("path", "/")
	uploadID := c.DefaultForm("upload_id", "")
	chunkIndex := c.DefaultForm("chunk_index", "-1")
	disk := file.NewDisk("")
	if err := disk.CreateDir(uploadPath); err != nil {
		c.Error("创建目录失败: " + err.Error())
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.Error("获取上传文件失败: " + err.Error())
		return
	}
	fileName := fileHeader.Filename
	chunkIndexInt, _ := strconv.Atoi(chunkIndex)
	if uploadID != "" && chunkIndexInt >= 0 {
		chunkDir := filepath.Join(constant.TempPath, "upload", uploadID)
		if err := os.MkdirAll(chunkDir, 0755); err != nil {
			c.Error("创建临时目录失败: " + err.Error())
			return
		}
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("%d", chunkIndexInt))
		if _, err := os.Stat(chunkPath); err == nil {
			data := map[string]interface{}{
				"upload_id":   uploadID,
				"chunk_index": chunkIndexInt,
				"file_name":   fileName,
				"exists":      true,
			}
			c.SuccessWithData("分片已存在", data)
			return
		}
		if err := saveUploadedFile(fileHeader, chunkPath); err != nil {
			c.Error("保存分片失败: " + err.Error())
			return
		}
		data := map[string]interface{}{
			"upload_id":   uploadID,
			"chunk_index": chunkIndexInt,
			"file_name":   fileName,
			"exists":      false,
		}
		c.SuccessWithData("分片上传成功", data)
		return
	}
	filePath := filepath.Join(uploadPath, fileName)
	if err := disk.AutoCreate(filePath); err != nil {
		c.Error("创建文件失败: " + err.Error())
		return
	}
	destFile, err := disk.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		c.Error("打开目标文件失败: " + err.Error())
		return
	}
	defer destFile.Close()
	srcFile, err := fileHeader.Open()
	if err != nil {
		c.Error("打开源文件失败: " + err.Error())
		return
	}
	defer srcFile.Close()
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		c.Error("保存文件失败: " + err.Error())
		return
	}
	data := map[string]interface{}{
		"name": fileName,
		"path": filePath,
		"size": fileHeader.Size,
	}
	c.SuccessWithData("文件上传成功", data)
}

// handleCheckChunks 检查分片上传状态
func handleCheckChunks(c *server.Context) {
	uploadID := c.DefaultQuery("upload_id", "")
	if uploadID == "" {
		c.Error("缺少upload_id参数")
		return
	}
	totalChunks, _ := strconv.Atoi(c.DefaultQuery("total_chunks", "0"))
	if totalChunks <= 0 {
		c.Error("缺少total_chunks参数")
		return
	}
	chunkDir := filepath.Join(constant.TempPath, "upload", uploadID)
	if _, err := os.Stat(chunkDir); os.IsNotExist(err) {
		data := map[string]interface{}{
			"uploaded_chunks": []int{},
		}
		c.SuccessWithData("尚未上传任何分片", data)
		return
	}
	var uploadedChunks []int
	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("%d", i))
		if _, err := os.Stat(chunkPath); err == nil {
			uploadedChunks = append(uploadedChunks, i)
		}
	}
	data := map[string]interface{}{
		"upload_id":       uploadID,
		"total_chunks":    totalChunks,
		"uploaded_chunks": uploadedChunks,
	}
	c.SuccessWithData("获取分片状态成功", data)
}

// handleMergeChunks 合并文件分片
func handleMergeChunks(c *server.Context) {
	uploadID := c.DefaultForm("upload_id", "")
	if uploadID == "" {
		c.Error("缺少upload_id参数")
		return
	}
	fileName := c.DefaultForm("file_name", "")
	if fileName == "" {
		c.Error("缺少file_name参数")
		return
	}
	filePath := c.DefaultForm("path", "/")
	if !strings.HasSuffix(filePath, "/") {
		filePath += "/"
	}
	totalChunks, _ := strconv.Atoi(c.DefaultForm("total_chunks", "0"))
	if totalChunks <= 0 {
		c.Error("缺少total_chunks参数")
		return
	}
	chunkDir := filepath.Join(constant.TempPath, "upload", uploadID)
	if _, err := os.Stat(chunkDir); os.IsNotExist(err) {
		c.Error("上传ID无效或分片不存在")
		return
	}
	var missingChunks []int
	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("%d", i))
		if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
			missingChunks = append(missingChunks, i)
		}
	}
	if len(missingChunks) > 0 {
		c.Error(fmt.Sprintf("还有 %d 个分片未上传完成", len(missingChunks)))
		return
	}
	disk := file.NewDisk("")
	if err := disk.CreateDir(filePath); err != nil {
		c.Error("创建目录失败: " + err.Error())
		return
	}
	destPath := filepath.Join(filePath, fileName)
	if err := disk.AutoCreate(destPath); err != nil {
		c.Error("创建合并文件失败: " + err.Error())
		return
	}
	destFile, err := disk.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		c.Error("打开目标文件失败: " + err.Error())
		return
	}
	defer destFile.Close()
	var fileSize int64 = 0
	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("%d", i))
		if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
			c.Error(fmt.Sprintf("分片 %d 不存在", i))
			return
		}
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			c.Error(fmt.Sprintf("打开分片 %d 失败: %s", i, err.Error()))
			return
		}
		written, err := io.Copy(destFile, chunkFile)
		chunkFile.Close()
		if err != nil {
			c.Error(fmt.Sprintf("写入分片 %d 失败: %s", i, err.Error()))
			return
		}
		fileSize += written
	}
	_ = os.RemoveAll(chunkDir)
	data := map[string]interface{}{
		"name": fileName,
		"path": destPath,
		"size": fileSize,
	}
	c.SuccessWithData("文件合并成功", data)
}

// saveUploadedFile 保存上传的文件到指定路径
func saveUploadedFile(fileHeader *multipart.FileHeader, destPath string) error {
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	return err
}
