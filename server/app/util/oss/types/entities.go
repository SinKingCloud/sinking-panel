// Package types 定义各对象存储厂商共用的配置、请求和返回数据。
package types

import (
	"io"
	"net/http"
	"time"
)

// Config 的 Provider 使用 secret_provider 枚举。Endpoint 为区域服务地址，不含桶名。
// Bucket 可留空，此时只使用桶管理 API。密钥不会通过 JSON 输出。
type Config struct {
	Provider        int          `json:"provider"`
	Region          string       `json:"region"`
	Endpoint        string       `json:"endpoint"`
	Bucket          string       `json:"bucket"`
	AccessKeyID     string       `json:"-"`
	AccessKeySecret string       `json:"-"`
	SecurityToken   string       `json:"-"`
	HTTPClient      *http.Client `json:"-"`
}

// Bucket 描述存储桶的名称、区域和创建时间。
type Bucket struct {
	Name      string    `json:"name"`
	Region    string    `json:"region,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Object 描述对象信息；写入操作仅返回已知字段，完整元数据通过 StatObject 查询。
type Object struct {
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

// ObjectPage 返回一页对象、子目录及下一页游标。
type ObjectPage struct {
	Objects        []Object `json:"objects"`
	CommonPrefixes []string `json:"common_prefixes"`
	NextToken      string   `json:"next_token"`
	HasMore        bool     `json:"has_more"`
}

// PutOptions 配置写入对象的内容属性、存储类型和用户元数据。
type PutOptions struct {
	ContentType        string
	CacheControl       string
	ContentDisposition string
	ContentEncoding    string
	StorageClass       string
	Metadata           map[string]string
}

// GetOptions 支持单个 HTTP bytes 范围，例如 bytes=0-1023；IfMatch 可避免读取已变更对象。
type GetOptions struct {
	Range   string
	IfMatch string
}

// Download 包含对象信息和下载流，调用方负责关闭 Body。
type Download struct {
	Object       Object
	Body         io.ReadCloser
	ContentRange string
}

// CopyOptions 默认保留源元数据；ReplaceMetadata 为 true 时使用 PutOptions 替换。
type CopyOptions struct {
	SourceBucket    string
	ReplaceMetadata bool
	PutOptions      PutOptions
}

// CopyPartOptions 指定源对象的一段字节范围，IfMatch 用于固定复制源的 ETag。
type CopyPartOptions struct {
	SourceBucket string
	Offset       int64
	Size         int64
	IfMatch      string
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

// UploadsOptions 按前缀筛选会话，分页时同时回传对象名和上传 ID 游标。
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

// SignedURL 包含临时访问地址和必须同时携带的请求头。
type SignedURL struct {
	URL       string      `json:"url"`
	Headers   http.Header `json:"headers,omitempty"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// DeleteResult 描述批量删除中的单个对象结果。
type DeleteResult struct {
	Key string
	Err error
}

// UploadOptions 配置自动分片上传。Progress 在当前调用中串行执行，反映已确认上传字节数。
// PartSize 为 0 时使用 16 MiB，已知长度的大文件会自动扩大分片，确保不超过 10000 片。
// 超过 64 MiB 的分片使用临时文件；TempDir 为空时使用系统临时目录。
type UploadOptions struct {
	PutOptions
	PartSize int64
	TempDir  string
	Progress func(uploaded, total int64)
}
