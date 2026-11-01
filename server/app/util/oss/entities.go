package oss

import (
	"io"
	"net/http"
	"time"
)

// Bucket 描述存储桶的名称、区域和创建时间。
type Bucket struct {
	Name      string    `json:"name"`
	Region    string    `json:"region,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// File 描述文件信息；写入操作仅返回已知字段，完整元数据通过 StatFile 查询。
type File struct {
	Key                string            `json:"key"`
	Size               int64             `json:"size"`
	ETag               string            `json:"etag"`
	LastModified       time.Time         `json:"last_modified"`
	ContentType        string            `json:"content_type,omitempty"`
	CacheControl       string            `json:"cache_control,omitempty"`
	ContentDisposition string            `json:"content_disposition,omitempty"`
	ContentEncoding    string            `json:"content_encoding,omitempty"`
	StorageClass       string            `json:"storage_class,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
}

// ListOptions 配置前缀、目录分隔符和分页大小。
// Token 是不透明游标，下一页须保持 Prefix/Delimiter 不变；Limit 范围 1..1000，0 使用 1000。
type ListOptions struct {
	Prefix    string
	Delimiter string
	Token     string
	Limit     int
}

// FilePage 返回一页文件、子目录及下一页游标。
type FilePage struct {
	Files          []File   `json:"files"`
	CommonPrefixes []string `json:"common_prefixes"`
	NextToken      string   `json:"next_token"`
	HasMore        bool     `json:"has_more"`
}

// PutOptions 配置写入文件的内容属性、存储类型和用户元数据。
type PutOptions struct {
	ContentType        string
	CacheControl       string
	ContentDisposition string
	ContentEncoding    string
	StorageClass       string
	Metadata           map[string]string
}

// GetOptions 支持单个 HTTP bytes 范围，例如 bytes=0-1023；IfMatch 可避免读取已变更文件。
type GetOptions struct {
	Range    string
	IfMatch  string
	Progress ProgressFunc // Manager 按 Body 读取进度通知；范围下载的 total 为本次响应长度。
}

// Download 包含文件信息和下载流，调用方负责关闭 Body。
type Download struct {
	File         File
	Body         io.ReadCloser
	ContentRange string
}

// CopyOptions 配置当前桶内复制，默认保留源元数据；ReplaceMetadata 为 true 时使用 PutOptions 替换。
type CopyOptions struct {
	ReplaceMetadata bool
	PutOptions      PutOptions
}

// CopyPartOptions 指定当前桶内源文件的一段字节范围，IfMatch 用于固定复制源的 ETag。
type CopyPartOptions struct {
	Offset  int64
	Size    int64
	IfMatch string
}

// Part 描述已上传分片，合并时需要提供编号、ETag 和长度。
type Part struct {
	Number       int       `json:"number"`
	ETag         string    `json:"etag"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
}

// MultipartUpload 标识尚未完成的上传会话，可保存到任务记录供恢复使用。
type MultipartUpload struct {
	Key       string    `json:"key"`
	UploadID  string    `json:"upload_id"`
	Initiated time.Time `json:"initiated"`
}

// PartsOptions 使用最后一个分片编号分页查询已上传分片。
type PartsOptions struct {
	Marker int
	Limit  int
}

// PartsPage 返回已上传分片及下一页编号。
type PartsPage struct {
	Parts      []Part `json:"parts"`
	NextMarker int    `json:"next_marker"`
	HasMore    bool   `json:"has_more"`
}

// UploadsOptions 按前缀筛选会话，分页时同时回传文件名和上传 ID 游标。
type UploadsOptions struct {
	Prefix         string
	KeyMarker      string
	UploadIDMarker string
	Limit          int
}

// UploadsPage 返回未完成会话及下一页的两个游标。
type UploadsPage struct {
	Uploads            []MultipartUpload `json:"uploads"`
	NextKeyMarker      string            `json:"next_key_marker"`
	NextUploadIDMarker string            `json:"next_upload_id_marker"`
	HasMore            bool              `json:"has_more"`
}

// PresignOptions 覆盖 GET 签名地址的响应头，不修改文件元数据。
type PresignOptions struct {
	ContentType        string // 响应内容类型，阿里云不支持覆盖此响应头。
	ContentDisposition string // 响应展示方式；下载可设为 attachment 并携带文件名。
}

// SignedURL 包含临时访问地址和必须同时携带的请求头。
type SignedURL struct {
	URL       string      `json:"url"`
	Headers   http.Header `json:"headers,omitempty"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// DeleteResult 描述批量删除中的单个文件结果。
type DeleteResult struct {
	Key string
	Err error
}

// UploadOptions 配置自动分片上传，Progress 接收 SDK 读取的累计字节数。
// PartSize 为 0 时使用 16 MiB，已知长度的大文件会自动扩大分片，确保不超过 10000 片。
// 超过 64 MiB 的分片使用临时文件；TempDir 为空时使用系统临时目录。
type UploadOptions struct {
	PutOptions
	PartSize int64
	TempDir  string
	Progress ProgressFunc
}

// ProgressFunc 接收已传输字节数与总字节数，未知总长度为 -1。
// 上传按 SDK 消费数据计数，文件下载按落盘计数，成功后才报告最终完成。
// GetFile 的进度表示 Body 完整读取，调用方仍需检查读取和 Close 的错误。
// 同一操作串行回调并自动节流；回调应快速返回，业务以方法返回的错误判断成功与否。
type ProgressFunc func(transferred, total int64)

// DownloadOptions 配置完整文件下载的进度，文件成功替换后才报告最终完成。
type DownloadOptions struct {
	Progress ProgressFunc
}

// Manager 绑定一个桶，可并发调用；所有网络操作都接受 context。
type Manager struct {
	config Config
	driver Backend
}
