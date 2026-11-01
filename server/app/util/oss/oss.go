// Package oss 提供统一的对象存储接口，并按密钥厂商创建对应实现。
package oss

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"server/app/enum/secret_provider"
	"sync"
	"time"
)

// Interface 定义供文件管理和备份服务依赖的统一对象存储接口。
type Interface interface {
	Backend
	Exists(context.Context, string) (bool, error)
	MoveFile(context.Context, string, string) (*File, error)
	SetMetadata(context.Context, string, PutOptions) (*File, error)
	CreateDirectory(context.Context, string) (*File, error)
	DeleteFiles(context.Context, []string) ([]DeleteResult, error)
	Upload(context.Context, string, io.Reader, int64, UploadOptions) (*File, error)
	UploadFile(context.Context, string, string, UploadOptions) (*File, error)
	DownloadFile(context.Context, string, string, ...DownloadOptions) error
}

// Backend 定义厂商实现需要提供的底层能力，组合操作与进度处理由 Manager 负责。
// 所有 reader 都按给定 size 发送；调用者负责关闭下载流。
type Backend interface {
	ListBuckets(context.Context) ([]Bucket, error)
	HeadBucket(context.Context) error
	CreateBucket(context.Context) error
	DeleteBucket(context.Context) error
	ListFiles(context.Context, ListOptions) (*FilePage, error)
	StatFile(context.Context, string) (*File, error)
	GetFile(context.Context, string, GetOptions) (*Download, error)
	PutFile(context.Context, string, io.Reader, int64, PutOptions) (*File, error)
	DeleteFile(context.Context, string) error
	CopyFile(context.Context, string, string, CopyOptions) (*File, error)
	InitiateMultipartUpload(context.Context, string, PutOptions) (*MultipartUpload, error)
	UploadPart(context.Context, string, string, int, io.Reader, int64) (*Part, error)
	UploadPartCopy(context.Context, string, string, int, string, CopyPartOptions) (*Part, error)
	CompleteMultipartUpload(context.Context, string, string, []Part) (*File, error)
	AbortMultipartUpload(context.Context, string, string) error
	ListParts(context.Context, string, string, PartsOptions) (*PartsPage, error)
	ListMultipartUploads(context.Context, UploadsOptions) (*UploadsPage, error)
	Presign(context.Context, string, string, time.Duration, ...PresignOptions) (*SignedURL, error)
}

var _ Interface = (*Manager)(nil)

var (
	providersMu sync.RWMutex
	providers   = make(map[int]func(SecretSource, ConfigSource) (Backend, error))
)

// Register 注册厂商构造函数，通常由厂商子包的 init 调用。
// 空函数或重复注册属于接入错误，直接 panic，避免静默覆盖已有实现。
func Register(provider int, factory func(SecretSource, ConfigSource) (Backend, error)) {
	if factory == nil {
		panic("oss: 厂商构造函数不能为空")
	}
	providersMu.Lock()
	defer providersMu.Unlock()
	if _, exists := providers[provider]; exists {
		panic("oss: 厂商已注册")
	}
	providers[provider] = factory
}

// NewManager 分别接收密钥和存储配置，创建绑定存储桶的管理器，不发网络请求。
// 调用方需导入对应厂商子包，由其自动注册具体实现。
func NewManager(credentials SecretSource, source ConfigSource) (*Manager, error) {
	c, err := ResolveConfig(source)
	if err != nil {
		return nil, err
	}
	secret, c, err := NormalizeConfig(credentials, source, c.Provider)
	if err != nil {
		return nil, err
	}
	source = source.WithHTTPClient(c.HTTPClient)
	m := &Manager{config: c}
	providersMu.RLock()
	factory := providers[c.Provider]
	providersMu.RUnlock()
	if factory == nil {
		return nil, m.invalid("厂商实现未注册，请导入对应的 oss 厂商子包")
	}
	driver, err := factory(secret, source)
	if err != nil {
		if errors.Is(err, ErrInvalidArgument) {
			return nil, err
		}
		return nil, m.providerError("NewManager", "", err)
	}
	if driver == nil {
		return nil, m.invalid("厂商构造函数返回了空客户端")
	}
	m.driver = driver
	return m, nil
}

// NewFromSecret 读取原始密钥记录中的凭据及临时令牌，存储配置独立传入。
// 统一配置的 Provider 由记录指定；厂商专用配置必须与记录厂商一致。
func NewFromSecret(provider int, data string, source ConfigSource) (*Manager, error) {
	c, err := ResolveConfig(source)
	if err != nil {
		return nil, err
	}
	switch source.(type) {
	case Config, *Config:
		c.Provider = provider
		source = c
	default:
		if c.Provider != provider {
			return nil, fmt.Errorf("%w: 配置与密钥的厂商不一致", ErrInvalidArgument)
		}
	}
	var fields map[string]string
	if err := json.Unmarshal([]byte(data), &fields); err != nil || fields == nil {
		return nil, fmt.Errorf("%w: 密钥内容必须是 JSON 字符串字段对象", ErrInvalidArgument)
	}
	secret := Secret{AccessKeyID: fields["access_key_id"], SecurityToken: fields["security_token"]}
	switch provider {
	case secret_provider.TencentCloud:
		secret.AccessKeyID, secret.AccessKeySecret, secret.SecurityToken = fields["secret_id"], fields["secret_key"], fields["session_token"]
	case secret_provider.Aliyun, secret_provider.Volcengine:
		secret.AccessKeySecret = fields["access_key_secret"]
	case secret_provider.HuaweiCloud, secret_provider.BaiduCloud:
		secret.AccessKeySecret = fields["secret_access_key"]
	default:
		return nil, fmt.Errorf("%w: 不支持的密钥厂商", ErrInvalidArgument)
	}
	return NewManager(secret, source)
}
