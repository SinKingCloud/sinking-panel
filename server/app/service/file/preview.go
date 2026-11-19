package file

import (
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

// CreatePreviewSign 生成文件签名；有效期至少一分钟，非正数表示不过期。
func (s *service) CreatePreviewSign(path string, download bool, expires time.Duration) (string, error) {
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
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", errors.New("生成预览签名失败")
	}
	key := hex.EncodeToString(random[:])
	value := PreviewSign{
		Path:     path,
		FileName: fileName,
		Download: download,
	}
	if expires > 0 {
		expires = max(expires, time.Minute)
		value.TTL = expires
		value.ExpiresAt = time.Now().Add(expires)
		if value.ExpiresAt.UnixNano() <= 0 {
			return "", errors.New("预览签名有效期过长")
		}
	} else {
		expires = -1
	}
	s.cache.SetWithExpire(constant.CacheNameWithFilePreview+key, value, expires)
	return key, nil
}

// CheckPreviewSign 校验文件预览签名
func (s *service) CheckPreviewSign(key string) (PreviewSign, error) {
	if len(key) != 32 {
		return PreviewSign{}, errors.New("预览签名无效或已过期")
	}
	if _, err := hex.DecodeString(key); err != nil {
		return PreviewSign{}, errors.New("预览签名无效或已过期")
	}
	cacheKey := constant.CacheNameWithFilePreview + key
	value := s.cache.Get(cacheKey)
	sign, ok := value.(PreviewSign)
	if !ok || sign.Path == "" || (!sign.ExpiresAt.IsZero() && !sign.ExpiresAt.After(time.Now())) {
		return PreviewSign{}, errors.New("预览签名无效或已过期")
	}
	// 保留临期续签，短签名按签发时长缩短续期窗口。
	if sign.TTL > 0 && time.Until(sign.ExpiresAt) < min(10*time.Minute, sign.TTL/6) {
		ttl := min(30*time.Minute, sign.TTL)
		sign.ExpiresAt = time.Now().Add(ttl)
		if !s.cache.ReplaceWithExpire(cacheKey, sign, ttl) {
			return PreviewSign{}, errors.New("预览签名无效或已过期")
		}
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
	return &PreviewFile{
		File:      handle,
		Name:      name,
		Size:      info.Size(),
		ModTime:   info.ModTime(),
		Download:  sign.Download,
		ExpiresAt: sign.ExpiresAt,
	}, nil
}

// IsViewableInBrowser 判断是否能在浏览器预览
func (s *service) IsViewableInBrowser(contentType string) bool {
	switch contentType {
	case "text/plain", "text/css", "text/markdown", "text/csv",
		"application/javascript", "application/json", "application/pdf",
		"image/png", "image/jpeg", "image/gif", "image/bmp", "image/webp",
		"audio/mpeg", "video/mp4", "video/quicktime", "video/webm":
		return true
	}
	return false
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
