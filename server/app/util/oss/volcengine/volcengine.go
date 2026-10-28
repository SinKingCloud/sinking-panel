package volcengine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"server/app/util/oss/types"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

// Client 提供火山引擎 TOS 的对象存储实现。
type Client struct {
	config types.Config
	client *tos.ClientV2
}

// New 使用已标准化的配置创建火山引擎 TOS 客户端。
func New(c types.Config) (*Client, error) {
	opts := []tos.ClientOption{
		tos.WithRegion(c.Region),
		tos.WithCredentialsProvider(tos.NewStaticCredentialsProvider(c.AccessKeyID, c.AccessKeySecret, c.SecurityToken)),
	}
	if c.HTTPClient != nil {
		// 传入完整 http.Client，保留超时、代理、取消和禁止重定向配置。
		opts = append(opts, tos.WithTransport(tos.NewDefaultTransportWithClient(*c.HTTPClient)))
	}
	client, err := tos.NewClientV2(c.Endpoint, opts...)
	if err != nil {
		return nil, types.WrapError(c.Provider, "New", "", err)
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
		return &types.Error{Provider: b.config.Provider, Op: op, Key: key, StatusCode: serverErr.StatusCode, Code: serverErr.Code, Message: serverErr.Message, Err: err}
	}
	// SDK 的 TosClientError 未实现 Unwrap，显式恢复底层取消/超时链。
	var clientErr *tos.TosClientError
	if errors.As(err, &clientErr) && clientErr.Cause != nil {
		return types.WrapError(b.config.Provider, op, key, clientErr.Cause)
	}
	return types.WrapError(b.config.Provider, op, key, err)
}

// ListBuckets 读取账户下的存储桶列表。
func (b *Client) ListBuckets(ctx context.Context) ([]types.Bucket, error) {
	out, err := b.client.ListBuckets(ctx, &tos.ListBucketsInput{})
	if err != nil {
		return nil, b.wrap("ListBuckets", "", err)
	}
	result := make([]types.Bucket, 0, len(out.Buckets))
	for _, bucket := range out.Buckets {
		created, _ := time.Parse(time.RFC3339Nano, bucket.CreationDate)
		result = append(result, types.Bucket{Name: bucket.Name, Region: bucket.Location, CreatedAt: created})
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

// ListObjects 读取单页对象和目录前缀，保留原始续页游标。
func (b *Client) ListObjects(ctx context.Context, opt types.ListOptions) (*types.ObjectPage, error) {
	out, err := b.client.ListObjectsType2(ctx, &tos.ListObjectsType2Input{
		Bucket: b.config.Bucket, Prefix: opt.Prefix, Delimiter: opt.Delimiter,
		ContinuationToken: opt.Token, MaxKeys: opt.Limit, ListOnlyOnce: true,
	})
	if err != nil {
		return nil, b.wrap("ListObjects", opt.Prefix, err)
	}
	page := &types.ObjectPage{
		Objects:        make([]types.Object, 0, len(out.Contents)),
		CommonPrefixes: make([]string, 0, len(out.CommonPrefixes)),
		NextToken:      out.NextContinuationToken,
		HasMore:        out.IsTruncated,
	}
	for _, obj := range out.Contents {
		page.Objects = append(page.Objects, types.Object{Key: obj.Key, Size: obj.Size, ETag: obj.ETag, LastModified: obj.LastModified, StorageClass: string(obj.StorageClass), Metadata: volcengineMetadata(obj.Meta)})
	}
	for _, prefix := range out.CommonPrefixes {
		page.CommonPrefixes = append(page.CommonPrefixes, prefix.Prefix)
	}
	return page, nil
}

// volcengineMetadata 将 SDK 元数据转换为统一的字符串映射。
func volcengineMetadata(meta tos.Metadata) map[string]string {
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

// volcengineObject 将 TOS 对象元数据转换为统一文件信息。
func volcengineObject(key string, meta tos.ObjectMetaV2) types.Object {
	return types.Object{
		Key:                key,
		Size:               meta.ContentLength,
		ETag:               meta.ETag,
		LastModified:       meta.LastModified,
		ContentType:        meta.ContentType,
		ContentEncoding:    meta.ContentEncoding,
		ContentDisposition: meta.ContentDisposition,
		CacheControl:       meta.CacheControl,
		StorageClass:       string(meta.StorageClass),
		Metadata:           volcengineMetadata(meta.Meta),
	}
}

// StatObject 读取对象元数据，不下载正文。
func (b *Client) StatObject(ctx context.Context, key string) (*types.Object, error) {
	out, err := b.client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{Bucket: b.config.Bucket, Key: key})
	if err != nil {
		return nil, b.wrap("StatObject", key, err)
	}
	obj := volcengineObject(key, out.ObjectMetaV2)
	return &obj, nil
}

// GetObject 流式读取对象，支持范围和 ETag 条件。
func (b *Client) GetObject(ctx context.Context, key string, opt types.GetOptions) (*types.Download, error) {
	out, err := b.client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: b.config.Bucket, Key: key, Range: opt.Range, IfMatch: opt.IfMatch})
	if err != nil {
		return nil, b.wrap("GetObject", key, err)
	}
	return &types.Download{Object: volcengineObject(key, out.ObjectMetaV2), Body: out.Content, ContentRange: out.ContentRange}, nil
}

// PutObject 上传指定长度的数据流和用户元数据。
func (b *Client) PutObject(ctx context.Context, key string, body io.Reader, size int64, opt types.PutOptions) (*types.Object, error) {
	out, err := b.client.PutObjectV2(ctx, &tos.PutObjectV2Input{PutObjectBasicInput: tos.PutObjectBasicInput{
		Bucket: b.config.Bucket, Key: key, ContentLength: size, ContentType: opt.ContentType,
		CacheControl: opt.CacheControl, ContentDisposition: opt.ContentDisposition, ContentEncoding: opt.ContentEncoding,
		StorageClass: enum.StorageClassType(opt.StorageClass), Meta: opt.Metadata,
	}, Content: body})
	if err != nil {
		return nil, b.wrap("PutObject", key, err)
	}
	return &types.Object{Key: key, Size: size, ETag: out.ETag, ContentType: opt.ContentType, CacheControl: opt.CacheControl,
		ContentDisposition: opt.ContentDisposition, ContentEncoding: opt.ContentEncoding, StorageClass: opt.StorageClass, Metadata: opt.Metadata}, nil
}

// DeleteObject 删除指定对象。
func (b *Client) DeleteObject(ctx context.Context, key string) error {
	_, err := b.client.DeleteObjectV2(ctx, &tos.DeleteObjectV2Input{Bucket: b.config.Bucket, Key: key})
	return b.wrap("DeleteObject", key, err)
}

// CopyObject 在服务端复制对象，可同时替换元数据。
func (b *Client) CopyObject(ctx context.Context, source, destination string, opt types.CopyOptions) (*types.Object, error) {
	sourceBucket := opt.SourceBucket
	if sourceBucket == "" {
		sourceBucket = b.config.Bucket
	}
	input := &tos.CopyObjectInput{Bucket: b.config.Bucket, Key: destination, SrcBucket: sourceBucket, SrcKey: source, MetadataDirective: enum.MetadataDirectiveCopy}
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
		return nil, b.wrap("CopyObject", destination, err)
	}
	modified, _ := time.Parse(time.RFC3339Nano, out.LastModified)
	return &types.Object{Key: destination, ETag: out.ETag, LastModified: modified}, nil
}

// InitiateMultipartUpload 初始化分片上传并返回可持久化的上传 ID。
func (b *Client) InitiateMultipartUpload(ctx context.Context, key string, opt types.PutOptions) (*types.MultipartUpload, error) {
	out, err := b.client.CreateMultipartUploadV2(ctx, &tos.CreateMultipartUploadV2Input{
		Bucket: b.config.Bucket, Key: key, ContentType: opt.ContentType, CacheControl: opt.CacheControl,
		ContentDisposition: opt.ContentDisposition, ContentEncoding: opt.ContentEncoding,
		StorageClass: enum.StorageClassType(opt.StorageClass), Meta: opt.Metadata,
	})
	if err != nil {
		return nil, b.wrap("InitiateMultipartUpload", key, err)
	}
	return &types.MultipartUpload{Key: key, UploadID: out.UploadID}, nil
}

// UploadPart 上传一个分片，同编号分片可以重新上传。
func (b *Client) UploadPart(ctx context.Context, key, uploadID string, number int, body io.Reader, size int64) (*types.Part, error) {
	out, err := b.client.UploadPartV2(ctx, &tos.UploadPartV2Input{UploadPartBasicInput: tos.UploadPartBasicInput{
		Bucket: b.config.Bucket, Key: key, UploadID: uploadID, PartNumber: number,
	}, Content: body, ContentLength: size})
	if err != nil {
		return nil, b.wrap("UploadPart", key, err)
	}
	return &types.Part{Number: number, ETag: out.ETag, Size: size}, nil
}

// UploadPartCopy 在服务端复制源对象的指定范围，避免备份文件经过本机下载。
func (b *Client) UploadPartCopy(ctx context.Context, key, uploadID string, number int, source string, options types.CopyPartOptions) (*types.Part, error) {
	sourceBucket := options.SourceBucket
	if sourceBucket == "" {
		sourceBucket = b.config.Bucket
	}
	out, err := b.client.UploadPartCopyV2(ctx, &tos.UploadPartCopyV2Input{
		Bucket: b.config.Bucket, Key: key, UploadID: uploadID, PartNumber: number,
		SrcBucket: sourceBucket, SrcKey: source,
		CopySourceRange:   fmt.Sprintf("bytes=%d-%d", options.Offset, options.Offset+options.Size-1),
		CopySourceIfMatch: options.IfMatch,
	})
	if err != nil {
		return nil, b.wrap("UploadPartCopy", key, err)
	}
	return &types.Part{Number: number, ETag: out.ETag, Size: options.Size, LastModified: out.LastModified}, nil
}

// CompleteMultipartUpload 按传入顺序合并已上传分片。
func (b *Client) CompleteMultipartUpload(ctx context.Context, key, uploadID string, parts []types.Part) (*types.Object, error) {
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
	return &types.Object{Key: key, ETag: out.ETag, Size: size}, nil
}

// AbortMultipartUpload 终止分片上传并清理服务端分片。
func (b *Client) AbortMultipartUpload(ctx context.Context, key, uploadID string) error {
	_, err := b.client.AbortMultipartUpload(ctx, &tos.AbortMultipartUploadInput{Bucket: b.config.Bucket, Key: key, UploadID: uploadID})
	return b.wrap("AbortMultipartUpload", key, err)
}

// ListParts 分页查询已上传分片及后续分片编号。
func (b *Client) ListParts(ctx context.Context, key, uploadID string, opt types.PartsOptions) (*types.PartsPage, error) {
	out, err := b.client.ListParts(ctx, &tos.ListPartsInput{Bucket: b.config.Bucket, Key: key, UploadID: uploadID, PartNumberMarker: opt.Marker, MaxParts: opt.Limit})
	if err != nil {
		return nil, b.wrap("ListParts", key, err)
	}
	page := &types.PartsPage{Parts: make([]types.Part, 0, len(out.Parts)), NextMarker: out.NextPartNumberMarker, HasMore: out.IsTruncated}
	for _, part := range out.Parts {
		page.Parts = append(page.Parts, types.Part{Number: part.PartNumber, ETag: part.ETag, Size: part.Size, LastModified: part.LastModified})
	}
	return page, nil
}

// ListMultipartUploads 使用对象和上传 ID 双游标查询未完成的上传会话。
func (b *Client) ListMultipartUploads(ctx context.Context, opt types.UploadsOptions) (*types.UploadsPage, error) {
	out, err := b.client.ListMultipartUploadsV2(ctx, &tos.ListMultipartUploadsV2Input{Bucket: b.config.Bucket, Prefix: opt.Prefix,
		KeyMarker: opt.KeyMarker, UploadIDMarker: opt.UploadIDMarker, MaxUploads: opt.Limit})
	if err != nil {
		return nil, b.wrap("ListMultipartUploads", opt.Prefix, err)
	}
	page := &types.UploadsPage{Uploads: make([]types.MultipartUpload, 0, len(out.Uploads)), NextKeyMarker: out.NextKeyMarker, NextUploadIDMarker: out.NextUploadIDMarker, HasMore: out.IsTruncated}
	for _, upload := range out.Uploads {
		page.Uploads = append(page.Uploads, types.MultipartUpload{Key: upload.Key, UploadID: upload.UploadID, Initiated: upload.Initiated})
	}
	return page, nil
}

// Presign 生成有有效期的签名地址及调用方必须携带的请求头。
func (b *Client) Presign(ctx context.Context, key, method string, expiry time.Duration) (*types.SignedURL, error) {
	if err := ctx.Err(); err != nil {
		return nil, b.wrap("Presign", key, err)
	}
	out, err := b.client.PreSignedURL(&tos.PreSignedURLInput{Bucket: b.config.Bucket, Key: key, HTTPMethod: enum.HttpMethodType(method), Expires: int64(expiry / time.Second)})
	if err != nil {
		return nil, b.wrap("Presign", key, err)
	}
	headers := make(http.Header, len(out.SignedHeader))
	for key, value := range out.SignedHeader {
		headers.Set(key, value)
	}
	return &types.SignedURL{URL: out.SignedUrl, Headers: headers, ExpiresAt: time.Now().Add(expiry)}, nil
}
