package file

import (
	"mime"
	"path/filepath"
	"strings"
)

// IsViewableInBrowser 判断是否能在浏览器预览
func (s *Service) IsViewableInBrowser(contentType string) bool {
	viewableTypes := map[string]bool{
		"text/html":              true,
		"text/plain":             true,
		"text/css":               true,
		"text/markdown":          true,
		"text/csv":               true,
		"application/javascript": true,
		"application/json":       true,
		"application/xml":        true,
		"application/pdf":        true,
		"image/png":              true,
		"image/jpeg":             true,
		"image/gif":              true,
		"image/bmp":              true,
		"image/svg+xml":          true,
		"image/webp":             true,
		"audio/mpeg":             true,
		"video/mp4":              true,
		"video/webm":             true,
	}
	return viewableTypes[contentType]
}

// GetContentType 获取文件的 Content-Type
func (s *Service) GetContentType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	// 常用MIME类型映射
	mimeTypes := map[string]string{
		".html":  "text/html",
		".htm":   "text/html",
		".css":   "text/css",
		".js":    "application/javascript",
		".json":  "application/json",
		".txt":   "text/plain",
		".md":    "text/markdown",
		".xml":   "application/xml",
		".pdf":   "application/pdf",
		".png":   "image/png",
		".jpg":   "image/jpeg",
		".jpeg":  "image/jpeg",
		".gif":   "image/gif",
		".bmp":   "image/bmp",
		".svg":   "image/svg+xml",
		".webp":  "image/webp",
		".mp3":   "audio/mpeg",
		".mp4":   "video/mp4",
		".webm":  "video/webm",
		".avi":   "video/x-msvideo",
		".mov":   "video/quicktime",
		".wmv":   "video/x-ms-wmv",
		".flv":   "video/x-flv",
		".doc":   "application/msword",
		".docx":  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":   "application/vnd.ms-excel",
		".xlsx":  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":   "application/vnd.ms-powerpoint",
		".pptx":  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".zip":   "application/zip",
		".rar":   "application/x-rar-compressed",
		".7z":    "application/x-7z-compressed",
		".tar":   "application/x-tar",
		".gz":    "application/gzip",
		".csv":   "text/csv",
		".ttf":   "font/ttf",
		".woff":  "font/woff",
		".woff2": "font/woff2",
	}
	if m, ok := mimeTypes[ext]; ok {
		return m
	}
	mimeType := mime.TypeByExtension(ext)
	if mimeType != "" {
		return mimeType
	}
	return "application/octet-stream"
}
