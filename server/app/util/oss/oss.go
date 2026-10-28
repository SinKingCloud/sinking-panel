// Package oss 提供统一的对象存储接口，并按密钥厂商创建对应实现。
package oss

import (
	"context"
	"encoding/json"
	"io"
	"server/app/enum/secret_provider"
	"server/app/util/oss/aliyun"
	"server/app/util/oss/baidu"
	"server/app/util/oss/huawei"
	"server/app/util/oss/tencent"
	"server/app/util/oss/volcengine"
	"time"
)

// Interface 定义供文件管理和备份服务依赖的统一对象存储接口。
type Interface interface {
	ListBuckets(context.Context) ([]Bucket, error)
	HeadBucket(context.Context) error
	CreateBucket(context.Context) error
	DeleteBucket(context.Context) error
	ListObjects(context.Context, ListOptions) (*ObjectPage, error)
	StatObject(context.Context, string) (*Object, error)
	GetObject(context.Context, string, GetOptions) (*Download, error)
	PutObject(context.Context, string, io.Reader, int64, PutOptions) (*Object, error)
	DeleteObject(context.Context, string) error
	CopyObject(context.Context, string, string, CopyOptions) (*Object, error)
	InitiateMultipartUpload(context.Context, string, PutOptions) (*MultipartUpload, error)
	UploadPart(context.Context, string, string, int, io.Reader, int64) (*Part, error)
	UploadPartCopy(context.Context, string, string, int, string, CopyPartOptions) (*Part, error)
	CompleteMultipartUpload(context.Context, string, string, []Part) (*Object, error)
	AbortMultipartUpload(context.Context, string, string) error
	ListParts(context.Context, string, string, PartsOptions) (*PartsPage, error)
	ListMultipartUploads(context.Context, UploadsOptions) (*UploadsPage, error)
	Presign(context.Context, string, string, time.Duration) (*SignedURL, error)
	Exists(context.Context, string) (bool, error)
	MoveObject(context.Context, string, string) (*Object, error)
	SetMetadata(context.Context, string, PutOptions) (*Object, error)
	CreateDirectory(context.Context, string) (*Object, error)
	DeleteObjects(context.Context, []string) ([]DeleteResult, error)
	Upload(context.Context, string, io.Reader, int64, UploadOptions) (*Object, error)
	UploadFile(context.Context, string, string, UploadOptions) (*Object, error)
	DownloadFile(context.Context, string, string) error
}

// backend 是厂商适配层。所有 reader 都按给定 size 发送；调用者负责关闭下载流。
type backend interface {
	ListBuckets(context.Context) ([]Bucket, error)
	HeadBucket(context.Context) error
	CreateBucket(context.Context) error
	DeleteBucket(context.Context) error
	ListObjects(context.Context, ListOptions) (*ObjectPage, error)
	StatObject(context.Context, string) (*Object, error)
	GetObject(context.Context, string, GetOptions) (*Download, error)
	PutObject(context.Context, string, io.Reader, int64, PutOptions) (*Object, error)
	DeleteObject(context.Context, string) error
	CopyObject(context.Context, string, string, CopyOptions) (*Object, error)
	InitiateMultipartUpload(context.Context, string, PutOptions) (*MultipartUpload, error)
	UploadPart(context.Context, string, string, int, io.Reader, int64) (*Part, error)
	UploadPartCopy(context.Context, string, string, int, string, CopyPartOptions) (*Part, error)
	CompleteMultipartUpload(context.Context, string, string, []Part) (*Object, error)
	AbortMultipartUpload(context.Context, string, string) error
	ListParts(context.Context, string, string, PartsOptions) (*PartsPage, error)
	ListMultipartUploads(context.Context, UploadsOptions) (*UploadsPage, error)
	Presign(context.Context, string, string, time.Duration) (*SignedURL, error)
}

var _ Interface = (*Manager)(nil)

// NewManager 创建客户端，不发网络请求。更换桶时创建新的 Manager。
func NewManager(c Config) (*Manager, error) {
	c, err := normalizeConfig(c)
	if err != nil {
		return nil, err
	}
	return newManager(c)
}

// newManager 使用已校验的配置创建厂商客户端，复用配置中的 HTTP 连接池。
func newManager(c Config) (*Manager, error) {
	var driver backend
	var err error
	switch c.Provider {
	case secret_provider.TencentCloud:
		driver, err = tencent.New(c)
	case secret_provider.Aliyun:
		driver, err = aliyun.New(c)
	case secret_provider.HuaweiCloud:
		driver, err = huawei.New(c)
	case secret_provider.Volcengine:
		driver, err = volcengine.New(c)
	case secret_provider.BaiduCloud:
		driver, err = baidu.New(c)
	default:
		return nil, invalid("不支持的密钥厂商")
	}
	if err != nil {
		return nil, providerError(c.Provider, "NewManager", "", err)
	}
	return &Manager{config: c, driver: driver}, nil
}

// NewFromSecret 使用 repositorySecret.FindById 返回的 Provider 和原始 Data。
// 不依赖数据库或 service 层；Config 的密钥会被凭据记录覆盖。
func NewFromSecret(provider int, data string, c Config) (*Manager, error) {
	var fields map[string]string
	if err := json.Unmarshal([]byte(data), &fields); err != nil || fields == nil {
		return nil, invalid("密钥内容必须是 JSON 字符串字段对象")
	}
	c.Provider = provider
	c.AccessKeyID = fields["access_key_id"]
	switch provider {
	case secret_provider.TencentCloud:
		c.AccessKeyID, c.AccessKeySecret = fields["secret_id"], fields["secret_key"]
	case secret_provider.Aliyun, secret_provider.Volcengine:
		c.AccessKeySecret = fields["access_key_secret"]
	case secret_provider.HuaweiCloud, secret_provider.BaiduCloud:
		c.AccessKeySecret = fields["secret_access_key"]
	default:
		return nil, invalid("不支持的密钥厂商")
	}
	return NewManager(c)
}
