package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DownloadWithProgress 带进度的文件下载，支持断点续传
func (s *service) DownloadWithProgress(ctx context.Context, url string, destPath string, progressCallback func(current, total int64, speed float64) bool) error {
	tempPath := destPath + ".download"
	dir := filepath.Dir(destPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	var resumeOffset int64 = 0
	fileInfo, err := os.Stat(tempPath)
	if err == nil && fileInfo.Size() > 0 {
		resumeOffset = fileInfo.Size()
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	if resumeOffset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeOffset))
	}
	client := &http.Client{
		Timeout: 0,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP错误: %s", resp.Status)
	}
	var contentLength int64 = -1
	if resp.StatusCode == http.StatusOK {
		contentLength = resp.ContentLength
	} else if resp.StatusCode == http.StatusPartialContent {
		contentRange := resp.Header.Get("Content-Range")
		if contentRange != "" {
			parts := strings.Split(contentRange, "/")
			if len(parts) == 2 {
				contentLength, _ = strconv.ParseInt(parts[1], 10, 64)
			}
		}
	} else {
		return fmt.Errorf("意外的HTTP状态: %s", resp.Status)
	}
	var file *os.File
	if resumeOffset > 0 && resp.StatusCode == http.StatusPartialContent {
		file, err = os.OpenFile(tempPath, os.O_APPEND|os.O_WRONLY, 0644)
	} else {
		file, err = os.Create(tempPath)
		resumeOffset = 0
	}
	if err != nil {
		return err
	}
	fileClosed := false
	defer func() {
		if !fileClosed {
			_ = file.Close()
		}
	}()
	buffer := make([]byte, 32*1024) // 32KB缓冲区
	var lastUpdate = time.Now()
	var lastBytes = resumeOffset
	var currentSpeed float64 = 0
	var updateInterval = time.Second
	var downloaded = resumeOffset
	for {
		if ctx.Err() != nil {
			return errors.New("下载被取消")
		}
		n, readErr := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := file.Write(buffer[:n])
			if writeErr != nil {
				return writeErr
			}
			// 更新进度
			downloaded += int64(n)
			// 计算下载速度
			now := time.Now()
			elapsed := now.Sub(lastUpdate)
			if elapsed >= updateInterval {
				bytesInPeriod := downloaded - lastBytes
				currentSpeed = float64(bytesInPeriod) / elapsed.Seconds()
				lastUpdate = now
				lastBytes = downloaded
				// 回调进度，同时检查是否应该继续
				if progressCallback != nil {
					if !progressCallback(downloaded, contentLength, currentSpeed) {
						return errors.New("下载被取消")
					}
				}
				// 再次检查上下文是否已取消
				if ctx.Err() != nil {
					return errors.New("下载被取消")
				}
			}
		}
		// 处理错误
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break // 正常结束
			}
			if ctx.Err() != nil {
				return errors.New("下载被取消")
			}
			return readErr // 其他错误
		}
	}
	if ctx.Err() != nil {
		return errors.New("下载被取消")
	}
	if err := file.Close(); err != nil {
		return err
	}
	fileClosed = true
	destInfo, err := os.Lstat(destPath)
	if os.IsNotExist(err) {
		return os.Rename(tempPath, destPath)
	}
	if err != nil {
		return err
	}
	if destInfo.IsDir() {
		return fmt.Errorf("目标路径是目录: %s", destPath)
	}
	backup, err := os.CreateTemp(dir, "."+filepath.Base(destPath)+".backup-*")
	if err != nil {
		return err
	}
	backupPath := backup.Name()
	if err = backup.Close(); err != nil {
		_ = os.Remove(backupPath)
		return err
	}
	if err = os.Remove(backupPath); err != nil {
		return err
	}
	if err = os.Rename(destPath, backupPath); err != nil {
		return err
	}
	if err = os.Rename(tempPath, destPath); err != nil {
		if restoreErr := os.Rename(backupPath, destPath); restoreErr != nil {
			return fmt.Errorf("替换下载文件失败: %v，恢复原文件失败: %v", err, restoreErr)
		}
		return err
	}
	if err = os.Remove(backupPath); err != nil {
		return fmt.Errorf("下载完成但清理原文件备份失败: %v", err)
	}
	return nil
}

// FormatSize 格式化文件大小
func (s *service) FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	var size string
	switch {
	case bytes < KB:
		size = fmt.Sprintf("%d B", bytes)
	case bytes < MB:
		size = fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	case bytes < GB:
		size = fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes < TB:
		size = fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	default:
		size = fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	}

	return size
}
