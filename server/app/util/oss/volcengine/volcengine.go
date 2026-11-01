package volcengine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"server/app/enum/secret_provider"
	"server/app/util/oss"
	"strings"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

// Client 提供火山引擎 TOS 的对象存储实现。
type Client struct {
	config oss.Config
	client *tos.ClientV2
}

var _ oss.Backend = (*Client)(nil)

// init 注册火山引擎 TOS实现，供统一入口按厂商创建客户端。
func init() {
	oss.Register(secret_provider.Volcengine, func(secret oss.SecretSource, config oss.ConfigSource) (oss.Backend, error) {
		return New(secret, config)
	})
}

// New 使用独立的密钥和连接配置创建客户端，不发起网络请求。
func New(secret oss.SecretSource, source oss.ConfigSource) (*Client, error) {
	key, c, err := oss.NormalizeConfig(secret, source, secret_provider.Volcengine)
	if err != nil {
		return nil, err
	}
	if c.Region == "" {
		return nil, fmt.Errorf("%w: 火山引擎必须提供 Region", oss.ErrInvalidArgument)
	}
	if c.Endpoint == "" {
		c.Endpoint = "https://tos-" + c.Region + ".volces.com"
	}
	opts := []tos.ClientOption{
		tos.WithRegion(c.Region),
		tos.WithCredentialsProvider(tos.NewStaticCredentialsProvider(key.AccessKeyID, key.AccessKeySecret, key.SecurityToken)),
	}
	if c.HTTPClient != nil {
		// 传入完整 http.Client，保留超时、代理、取消和禁止重定向配置。
		opts = append(opts, tos.WithTransport(tos.NewDefaultTransportWithClient(*c.HTTPClient)))
	}
	client, err := tos.NewClientV2(c.Endpoint, opts...)
	if err != nil {
		return nil, oss.WrapError(c.Provider, "New", "", err)
	}
	return &Client{config: c, client: client}, nil
}

// wrap 转换 TOS 服务端错误，并展开 SDK 封装的取消或超时原因。
func (b *Client) wrap(op, key string, err error) error {
	if err == nil {
		return nil
	}
	var serverErr *tos.TosServerError
	if errors.As(err, &serverErr) {
		return &oss.Error{Provider: b.config.Provider, Op: op, Key: key, StatusCode: serverErr.StatusCode, Code: serverErr.Code, Message: serverErr.Message, Err: err}
	}
	// SDK 的 TosClientError 未实现 Unwrap，显式恢复底层取消/超时链。
	var clientErr *tos.TosClientError
	if errors.As(err, &clientErr) && clientErr.Cause != nil {
		return oss.WrapError(b.config.Provider, op, key, clientErr.Cause)
	}
	return oss.WrapError(b.config.Provider, op, key, err)
}

// ListBuckets 读取账户下的存储桶列表。
func (b *Client) ListBuckets(ctx context.Context) ([]oss.Bucket, error) {
	out, err := b.client.ListBuckets(ctx, &tos.ListBucketsInput{})
	if err != nil {
		return nil, b.wrap("ListBuckets", "", err)
	}
	result := make([]oss.Bucket, 0, len(out.Buckets))
	for _, bucket := range out.Buckets {
		created, _ := time.Parse(time.RFC3339Nano, bucket.CreationDate)
		result = append(result, oss.Bucket{Name: bucket.Name, Region: bucket.Location, CreatedAt: created})
	}
	return result, nil
}

// HeadBucket 检查配置的存储桶是否存在且可访问。
func (b *Client) HeadBucket(ctx context.Context) error {
	_, err := b.client.HeadBucket(ctx, &tos.HeadBucketInput{Bucket: b.config.Bucket})
	return b.wrap("HeadBucket", "", err)
}

// CreateBucket 在配置的区域创建存储桶。
func (b *Client) CreateBucket(ctx context.Context) error {
	_, err := b.client.CreateBucketV2(ctx, &tos.CreateBucketV2Input{Bucket: b.config.Bucket})
	return b.wrap("CreateBucket", "", err)
}

// DeleteBucket 删除配置的空存储桶。
func (b *Client) DeleteBucket(ctx context.Context) error {
	_, err := b.client.DeleteBucket(ctx, &tos.DeleteBucketInput{Bucket: b.config.Bucket})
	return b.wrap("DeleteBucket", "", err)
}

// ListFiles 读取单页文件和目录前缀，保留原始续页游标。
func (b *Client) ListFiles(ctx context.Context, opt oss.ListOptions) (*oss.FilePage, error) {
	out, err := b.client.ListObjectsType2(ctx, &tos.ListObjectsType2Input{
		Bucket: b.config.Bucket, Prefix: opt.Prefix, Delimiter: opt.Delimiter,
		ContinuationToken: opt.Token, MaxKeys: opt.Limit, ListOnlyOnce: true,
	})
	if err != nil {
		return nil, b.wrap("ListFiles", opt.Prefix, err)
	}
	page := &oss.FilePage{
		Files:          make([]oss.File, 0, len(out.Contents)),
		CommonPrefixes: make([]string, 0, len(out.CommonPrefixes)),
		NextToken:      out.NextContinuationToken,
		HasMore:        out.IsTruncated,
	}
	for _, obj := range out.Contents {
		page.Files = append(page.Files, oss.File{Key: obj.Key, Size: obj.Size, ETag: obj.ETag, LastModified: obj.LastModified, StorageClass: string(obj.StorageClass), Metadata: b.volcengineMetadata(obj.Meta)})
	}
	for _, prefix := range out.CommonPrefixes {
		page.CommonPrefixes = append(page.CommonPrefixes, prefix.Prefix)
	}
	return page, nil
}

// volcengineMetadata 将 SDK 元数据转换为统一的字符串映射。
func (b *Client) volcengineMetadata(meta tos.Metadata) map[string]string {
	if meta == nil {
		return nil
	}
	result := make(map[string]string)
	meta.Range(func(k, v string) bool {
		result[k] = v
		return true
	})
	return result
}

// volcengineFile 将 TOS 文件元数据转换为统一文件信息。
func (b *Client) volcengineFile(key string, meta tos.ObjectMetaV2) oss.File {
	return oss.File{
		Key:                key,
		Size:               meta.ContentLength,
		ETag:               meta.ETag,
		LastModified:       meta.LastModified,
		ContentType:        meta.ContentType,
		ContentEncoding:    meta.ContentEncoding,
		ContentDisposition: meta.ContentDisposition,
		CacheControl:       meta.CacheControl,
		StorageClass:       string(meta.StorageClass),
		Metadata:           b.volcengineMetadata(meta.Meta),
	}
}

// StatFile 读取文件元数据，不下载正文。
func (b *Client) StatFile(ctx context.Context, key string) (*oss.File, error) {
	out, err := b.client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{Bucket: b.config.Bucket, Key: key})
	if err != nil {
		return nil, b.wrap("StatFile", key, err)
	}
	obj := b.volcengineFile(key, out.ObjectMetaV2)
	return &obj, nil
}

// GetFile 流式读取文件，支持范围和 ETag 条件。
func (b *Client) GetFile(ctx context.Context, key string, opt oss.GetOptions) (*oss.Download, error) {
	out, err := b.client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: b.config.Bucket, Key: key, Range: opt.Range, IfMatch: opt.IfMatch})
	if err != nil {
		return nil, b.wrap("GetFile", key, err)
	}
	file := b.volcengineFile(key, out.ObjectMetaV2)
	if out.Header.Get("Content-Length") == "" {
		file.Size = -1
	}
	return &oss.Download{File: file, Body: out.Content, ContentRange: out.ContentRange}, nil
}

// PutFile 上传指定长度的数据流和用户元数据。
func (b *Client) PutFile(ctx context.Context, key string, body io.Reader, size int64, opt oss.PutOptions) (*oss.File, error) {
	out, err := b.client.PutObjectV2(ctx, &tos.PutObjectV2Input{PutObjectBasicInput: tos.PutObjectBasicInput{
		Bucket: b.config.Bucket, Key: key, ContentLength: size, ContentType: opt.ContentType,
		CacheControl: opt.CacheControl, ContentDisposition: opt.ContentDisposition, ContentEncoding: opt.ContentEncoding,
		StorageClass: enum.StorageClassType(opt.StorageClass), Meta: opt.Metadata,
	}, Content: body})
	if err != nil {
		return nil, b.wrap("PutFile", key, err)
	}
	return &oss.File{Key: key, Size: size, ETag: out.ETag, ContentType: opt.ContentType, CacheControl: opt.CacheControl,
		ContentDisposition: opt.ContentDisposition, ContentEncoding: opt.ContentEncoding, StorageClass: opt.StorageClass, Metadata: opt.Metadata}, nil
}

// DeleteFile 删除指定文件。
func (b *Client) DeleteFile(ctx context.Context, key string) error {
	_, err := b.client.DeleteObjectV2(ctx, &tos.DeleteObjectV2Input{Bucket: b.config.Bucket, Key: key})
	return b.wrap("DeleteFile", key, err)
}

// CopyFile 在服务端复制文件，可同时替换元数据。
func (b *Client) CopyFile(ctx context.Context, source, destination string, opt oss.CopyOptions) (*oss.File, error) {
	input := &tos.CopyObjectInput{Bucket: b.config.Bucket, Key: destination, SrcBucket: b.config.Bucket, SrcKey: source, MetadataDirective: enum.MetadataDirectiveCopy}
	if opt.ReplaceMetadata {
		input.MetadataDirective = enum.MetadataDirectiveReplace
		input.ContentType = opt.PutOptions.ContentType
		input.CacheControl = opt.PutOptions.CacheControl
		input.ContentDisposition = opt.PutOptions.ContentDisposition
		input.ContentEncoding = opt.PutOptions.ContentEncoding
		input.Meta = opt.PutOptions.Metadata
		input.StorageClass = enum.StorageClassType(opt.PutOptions.StorageClass)
	}
	out, err := b.client.CopyObject(ctx, input)
	if err != nil {
		return nil, b.wrap("CopyFile", destination, err)
	}
	modified, _ := time.Parse(time.RFC3339Nano, out.LastModified)
	return &oss.File{Key: destination, ETag: out.ETag, LastModified: modified}, nil
}

// InitiateMultipartUpload 初始化分片上传并返回可持久化的上传 ID。
func (b *Client) InitiateMultipartUpload(ctx context.Context, key string, opt oss.PutOptions) (*oss.MultipartUpload, error) {
	out, err := b.client.CreateMultipartUploadV2(ctx, &tos.CreateMultipartUploadV2Input{
		Bucket: b.config.Bucket, Key: key, ContentType: opt.ContentType, CacheControl: opt.CacheControl,
		ContentDisposition: opt.ContentDisposition, ContentEncoding: opt.ContentEncoding,
		StorageClass: enum.StorageClassType(opt.StorageClass), Meta: opt.Metadata,
	})
	if err != nil {
		return nil, b.wrap("InitiateMultipartUpload", key, err)
	}
	return &oss.MultipartUpload{Key: key, UploadID: out.UploadID}, nil
}

// UploadPart 上传一个分片，同编号分片可以重新上传。
func (b *Client) UploadPart(ctx context.Context, key, uploadID string, number int, body io.Reader, size int64) (*oss.Part, error) {
	out, err := b.client.UploadPartV2(ctx, &tos.UploadPartV2Input{UploadPartBasicInput: tos.UploadPartBasicInput{
		Bucket: b.config.Bucket, Key: key, UploadID: uploadID, PartNumber: number,
	}, Content: body, ContentLength: size})
	if err != nil {
		return nil, b.wrap("UploadPart", key, err)
	}
	return &oss.Part{Number: number, ETag: out.ETag, Size: size}, nil
}

// UploadPartCopy 在服务端复制源文件的指定范围，避免备份文件经过本机下载。
func (b *Client) UploadPartCopy(ctx context.Context, key, uploadID string, number int, source string, options oss.CopyPartOptions) (*oss.Part, error) {
	out, err := b.client.UploadPartCopyV2(ctx, &tos.UploadPartCopyV2Input{
		Bucket: b.config.Bucket, Key: key, UploadID: uploadID, PartNumber: number,
		SrcBucket: b.config.Bucket, SrcKey: source,
		CopySourceRange:   fmt.Sprintf("bytes=%d-%d", options.Offset, options.Offset+options.Size-1),
		CopySourceIfMatch: options.IfMatch,
	})
	if err != nil {
		return nil, b.wrap("UploadPartCopy", key, err)
	}
	return &oss.Part{Number: number, ETag: out.ETag, Size: options.Size, LastModified: out.LastModified}, nil
}

// CompleteMultipartUpload 按传入顺序合并已上传分片。
func (b *Client) CompleteMultipartUpload(ctx context.Context, key, uploadID string, parts []oss.Part) (*oss.File, error) {
	items := make([]tos.UploadedPartV2, len(parts))
	var size int64
	for i, part := range parts {
		items[i] = tos.UploadedPartV2{PartNumber: part.Number, ETag: part.ETag}
		size += part.Size
	}
	out, err := b.client.CompleteMultipartUploadV2(ctx, &tos.CompleteMultipartUploadV2Input{Bucket: b.config.Bucket, Key: key, UploadID: uploadID, Parts: items})
	if err != nil {
		return nil, b.wrap("CompleteMultipartUpload", key, err)
	}
	return &oss.File{Key: key, ETag: out.ETag, Size: size}, nil
}

// AbortMultipartUpload 终止分片上传并清理服务端分片。
func (b *Client) AbortMultipartUpload(ctx context.Context, key, uploadID string) error {
	_, err := b.client.AbortMultipartUpload(ctx, &tos.AbortMultipartUploadInput{Bucket: b.config.Bucket, Key: key, UploadID: uploadID})
	return b.wrap("AbortMultipartUpload", key, err)
}

// ListParts 分页查询已上传分片及后续分片编号。
func (b *Client) ListParts(ctx context.Context, key, uploadID string, opt oss.PartsOptions) (*oss.PartsPage, error) {
	out, err := b.client.ListParts(ctx, &tos.ListPartsInput{Bucket: b.config.Bucket, Key: key, UploadID: uploadID, PartNumberMarker: opt.Marker, MaxParts: opt.Limit})
	if err != nil {
		return nil, b.wrap("ListParts", key, err)
	}
	page := &oss.PartsPage{Parts: make([]oss.Part, 0, len(out.Parts)), NextMarker: out.NextPartNumberMarker, HasMore: out.IsTruncated}
	for _, part := range out.Parts {
		page.Parts = append(page.Parts, oss.Part{Number: part.PartNumber, ETag: part.ETag, Size: part.Size, LastModified: part.LastModified})
	}
	return page, nil
}

// ListMultipartUploads 使用文件和上传 ID 双游标查询未完成的上传会话。
func (b *Client) ListMultipartUploads(ctx context.Context, opt oss.UploadsOptions) (*oss.UploadsPage, error) {
	out, err := b.client.ListMultipartUploadsV2(ctx, &tos.ListMultipartUploadsV2Input{Bucket: b.config.Bucket, Prefix: opt.Prefix,
		KeyMarker: opt.KeyMarker, UploadIDMarker: opt.UploadIDMarker, MaxUploads: opt.Limit})
	if err != nil {
		return nil, b.wrap("ListMultipartUploads", opt.Prefix, err)
	}
	page := &oss.UploadsPage{Uploads: make([]oss.MultipartUpload, 0, len(out.Uploads)), NextKeyMarker: out.NextKeyMarker, NextUploadIDMarker: out.NextUploadIDMarker, HasMore: out.IsTruncated}
	for _, upload := range out.Uploads {
		page.Uploads = append(page.Uploads, oss.MultipartUpload{Key: upload.Key, UploadID: upload.UploadID, Initiated: upload.Initiated})
	}
	return page, nil
}

// Sign 生成有有效期的签名地址及调用方必须携带的请求头。
func (b *Client) Sign(ctx context.Context, key, method string, expiry time.Duration, options ...oss.SignOptions) (*oss.SignedURL, error) {
	if err := ctx.Err(); err != nil {
		return nil, b.wrap("Sign", key, err)
	}
	query := make(map[string]string)
	if len(options) > 0 {
		if options[0].ContentType != "" {
			query["response-content-type"] = options[0].ContentType
		}
		if options[0].ContentDisposition != "" {
			query["response-content-disposition"] = options[0].ContentDisposition
		}
	}
	out, err := b.client.PreSignedURL(&tos.PreSignedURLInput{Bucket: b.config.Bucket, Key: key, HTTPMethod: enum.HttpMethodType(method), Expires: int64(expiry / time.Second), Query: query})
	if err != nil {
		return nil, b.wrap("Sign", key, err)
	}
	signedURL, err := url.Parse(out.SignedUrl)
	if err != nil {
		return nil, b.wrap("Sign", key, err)
	}
	required := make(map[string]bool)
	for _, name := range strings.Split(signedURL.Query().Get("X-Tos-SignedHeaders"), ";") {
		required[strings.ToLower(name)] = true
	}
	headers := make(http.Header, len(out.SignedHeader))
	for key, value := range out.SignedHeader {
		// SDK 返回所有构造时的请求头，只保留签名实际要求的字段。
		if required[strings.ToLower(key)] {
			headers.Set(key, value)
		}
	}
	return &oss.SignedURL{URL: out.SignedUrl, Headers: headers, ExpiresAt: time.Now().Add(expiry)}, nil
}
