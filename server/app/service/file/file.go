package file

import (
	"context"
	"mime/multipart"
)

// Service service接口
type Service interface {
	GetDisks() ([]Disk, error)
	GetDiskPaths() ([]string, error)
	CopyWithContext(ctx context.Context, src, destDir string, callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error
	MoveWithContext(ctx context.Context, src, destDir string, callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error
	CompressFiles(ctx context.Context, srcPaths []string, destPath, format string, callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error
	Extract(ctx context.Context, srcPath, destPath string, callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error
	GetFormatByExt(filename string) string
	IsSupportedFormat(format string) bool
	DownloadWithProgress(ctx context.Context, url string, destPath string, progressCallback func(current, total int64, speed float64) bool) error
	FormatSize(bytes int64) string
	IsViewableInBrowser(contentType string) bool
	GetContentType(fileName string) string
	Upload(ctx context.Context, fileHeader *multipart.FileHeader, path string) (*UploadedFile, error)
	UploadChunk(fileHeader *multipart.FileHeader, meta UploadMeta, chunkIndex int) (*UploadedFile, bool, error)
	CheckUpload(meta UploadMeta) ([]int, *UploadedFile, error)
	MergeUpload(ctx context.Context, meta UploadMeta) (*UploadedFile, bool, error)
	ClearUpload(uploadID string) error
}

// service 注入结构
type service struct {
	upload uploadState
}

// NewService 实例化service
func NewService() *service {
	return &service{upload: uploadState{locks: make(map[string]*uploadSessionLock)}}
}
