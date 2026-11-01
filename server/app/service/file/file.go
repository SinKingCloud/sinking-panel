package file

import (
	"context"
	"mime/multipart"
	serviceRecycle "server/app/service/recycle"
	serviceSystem "server/app/service/system"
	"server/app/util/cache"
	"server/app/util/page"
)

// Service service接口
type Service interface {
	List(path, keyword string, query *page.Query) (*page.Result[*File], error)
	Info(path string, read bool, cursor int64, pageSize int, version string) (map[string]interface{}, error)
	Count(path string) (map[string]interface{}, error)
	Create(path, name string, chmod uint32) error
	Update(path string, data *UpdateFile) (*UpdatedFile, error)
	Delete(paths []string, recycle bool) error
	GetDisks() ([]Disk, error)
	GetDiskPaths() ([]string, error)
	CopyTask(paths []string, target string) (string, error)
	MoveTask(paths []string, target string) (string, error)
	CompressTask(paths []string, dir, name, format string) (string, error)
	ExtractTask(path, dir string) (string, error)
	DownloadTask(url, path, name string) (taskID, targetPath string, err error)
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
	Upload(ctx context.Context, fileHeader *multipart.FileHeader, path string, overwrite bool) (*UploadedFile, error)
	UploadChunk(fileHeader *multipart.FileHeader, meta UploadMeta, chunkIndex int) (*UploadedFile, bool, error)
	CheckUpload(meta UploadMeta) ([]int, *UploadedFile, error)
	MergeUpload(ctx context.Context, meta UploadMeta) (*UploadedFile, bool, error)
	ClearUpload(uploadID string) error
	CreatePreviewSign(path string, download bool) (string, error)
	CheckPreviewSign(key string) (PreviewSign, error)
	OpenPreview(key string) (*PreviewFile, error)
}

// service 注入结构
type service struct {
	cache          cache.Interface
	recycleService serviceRecycle.Service
	systemService  serviceSystem.Service
	upload         uploadState
}

// NewService 实例化service
func NewService(recycleService serviceRecycle.Service, systemService serviceSystem.Service, cache cache.Interface) *service {
	return &service{
		cache:          cache,
		recycleService: recycleService,
		systemService:  systemService,
		upload:         uploadState{locks: make(map[string]*uploadSessionLock)},
	}
}
