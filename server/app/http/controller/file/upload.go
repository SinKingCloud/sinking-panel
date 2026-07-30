package file

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"runtime"
	"server/app/constant"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"server/app/util/file"
	"strconv"
	"strings"
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
func Upload(c *context.Context) {
	// 获取请求参数
	action := c.DefaultQuery("action", "upload")

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
func handleFileUpload(c *context.Context) {
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
		if info, err := os.Stat(chunkPath); err == nil && info.Mode().IsRegular() && info.Size() == fileHeader.Size {
			data := map[string]interface{}{
				"upload_id":   uploadID,
				"chunk_index": chunkIndexInt,
				"file_name":   fileName,
				"exists":      true,
			}
			c.SuccessWithData("分片已存在", data)
			return
		}
		_ = os.Remove(chunkPath)
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
	if err := saveUploadedFile(fileHeader, filePath); err != nil {
		c.Error("保存文件失败: " + err.Error())
		return
	}
	data := map[string]interface{}{
		"name": fileName,
		"path": filePath,
		"size": fileHeader.Size,
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "上传文件", "上传文件["+filePath+"]")
	c.SuccessWithData("文件上传成功", data)
}

// handleCheckChunks 检查分片上传状态
func handleCheckChunks(c *context.Context) {
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
func handleMergeChunks(c *context.Context) {
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
	totalSize, err := strconv.ParseInt(c.DefaultForm("total_size", "0"), 10, 64)
	if err != nil || totalSize < 0 {
		c.Error("total_size参数不合法")
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
	destFile, err := os.CreateTemp(filePath, "."+filepath.Base(fileName)+".merge-*")
	if err != nil {
		c.Error("打开目标文件失败: " + err.Error())
		return
	}
	tempPath := destFile.Name()
	completed := false
	defer func() {
		_ = destFile.Close()
		if !completed {
			_ = os.Remove(tempPath)
		}
	}()
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
	if totalSize > 0 && fileSize != totalSize {
		c.Error(fmt.Sprintf("合并文件大小不一致: %d/%d", fileSize, totalSize))
		return
	}
	if err = destFile.Sync(); err != nil {
		c.Error("保存合并文件失败: " + err.Error())
		return
	}
	if err = destFile.Chmod(0644); err != nil {
		c.Error("设置合并文件权限失败: " + err.Error())
		return
	}
	if err = destFile.Close(); err != nil {
		c.Error("关闭合并文件失败: " + err.Error())
		return
	}
	if runtime.GOOS == "windows" {
		if err = os.Remove(destPath); err != nil && !os.IsNotExist(err) {
			c.Error("替换目标文件失败: " + err.Error())
			return
		}
	}
	if err = os.Rename(tempPath, destPath); err != nil {
		c.Error("保存合并文件失败: " + err.Error())
		return
	}
	completed = true
	_ = os.RemoveAll(chunkDir)
	data := map[string]interface{}{
		"name": fileName,
		"path": destPath,
		"size": fileSize,
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventCreate, "合并上传文件", "上传文件["+destPath+"]")
	c.SuccessWithData("文件合并成功", data)
}

// saveUploadedFile 保存上传的文件到指定路径
func saveUploadedFile(fileHeader *multipart.FileHeader, destPath string) error {
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.CreateTemp(filepath.Dir(destPath), "."+filepath.Base(destPath)+".upload-*")
	if err != nil {
		return err
	}
	tempPath := dest.Name()
	defer func() {
		_ = dest.Close()
		_ = os.Remove(tempPath)
	}()
	written, err := io.Copy(dest, src)
	if err != nil {
		return err
	}
	if fileHeader.Size >= 0 && written != fileHeader.Size {
		return fmt.Errorf("上传文件大小不一致: %d/%d", written, fileHeader.Size)
	}
	if err = dest.Sync(); err != nil {
		return err
	}
	if err = dest.Chmod(0644); err != nil {
		return err
	}
	if err = dest.Close(); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		if err = os.Remove(destPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return os.Rename(tempPath, destPath)
}
