package file

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"mime"
	"os"
	"path/filepath"
	"server/app/constant"
	"server/app/util/file"
	"strings"
	"time"
)

// CreatePreviewSign 检查文件并生成预览签名，符号链接解析为实际文件路径。
func (s *service) CreatePreviewSign(path string, download bool) (string, error) {
	f := file.NewDisk("")
	if !f.Exists(path) {
		return "", errors.New("文件不存在")
	}
	if f.IsDir(path) {
		return "", errors.New("不能预览目录")
	}
	path = f.Path(path, true, true)
	fileName := filepath.Base(path)
	path, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", errors.New("文件不存在")
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", errors.New("文件不存在")
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("不支持访问该文件类型")
	}
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", errors.New("生成预览签名失败")
	}
	hash := md5.Sum(random)
	key := hex.EncodeToString(hash[:])
	value := PreviewSign{
		Path:      path,
		FileName:  fileName,
		Download:  download,
		ExpiresAt: time.Now().Add(constant.CacheTimeWithFilePreview),
	}
	s.cache.SetWithExpire(constant.CacheNameWithFilePreview+key, value, constant.CacheTimeWithFilePreview)
	return key, nil
}

// CheckPreviewSign 校验文件预览签名
func (s *service) CheckPreviewSign(key string) (PreviewSign, error) {
	if len(key) != 32 {
		return PreviewSign{}, errors.New("预览签名无效或已过期")
	}
	decoded, err := hex.DecodeString(key)
	if err != nil || len(decoded) != md5.Size {
		return PreviewSign{}, errors.New("预览签名无效或已过期")
	}
	cacheKey := constant.CacheNameWithFilePreview + key
	value := s.cache.Get(cacheKey)
	sign, ok := value.(PreviewSign)
	if !ok || sign.Path == "" || !sign.ExpiresAt.After(time.Now()) {
		s.cache.Delete(cacheKey)
		return PreviewSign{}, errors.New("预览签名无效或已过期")
	}
	if time.Until(sign.ExpiresAt) < constant.CacheTimeWithFilePreviewRenewBefore {
		sign.ExpiresAt = time.Now().Add(constant.CacheTimeWithFilePreviewRenew)
		s.cache.SetWithExpire(cacheKey, sign, constant.CacheTimeWithFilePreviewRenew)
	}
	return sign, nil
}

// OpenPreview 校验预览签名并打开普通文件，调用方在响应结束后关闭文件。
func (s *service) OpenPreview(key string) (*PreviewFile, error) {
	sign, err := s.CheckPreviewSign(key)
	if err != nil {
		return nil, errors.New("预览链接无效或已过期")
	}
	info, err := os.Stat(sign.Path)
	if err != nil {
		return nil, errors.New("文件不存在")
	}
	if info.IsDir() {
		return nil, errors.New("不允许访问目录")
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("不支持访问该文件类型")
	}
	handle, err := os.Open(sign.Path)
	if err != nil {
		return nil, errors.New("无法访问文件")
	}
	info, err = handle.Stat()
	if err != nil {
		handle.Close()
		return nil, errors.New("无法访问文件")
	}
	if info.IsDir() {
		handle.Close()
		return nil, errors.New("不允许访问目录")
	}
	if !info.Mode().IsRegular() {
		handle.Close()
		return nil, errors.New("不支持访问该文件类型")
	}
	name := filepath.Base(sign.FileName)
	if name == "." || name == string(filepath.Separator) {
		name = filepath.Base(sign.Path)
	}
	return &PreviewFile{File: handle, Name: name, ModTime: info.ModTime(), Download: sign.Download}, nil
}

// IsViewableInBrowser 判断是否能在浏览器预览
func (s *service) IsViewableInBrowser(contentType string) bool {
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
		"video/quicktime":        true,
		"video/webm":             true,
	}
	return viewableTypes[contentType]
}

// GetContentType 获取文件的 Content-Type
func (s *service) GetContentType(fileName string) string {
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
