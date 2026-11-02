package huawei

import (
	"context"
	"errors"
	"fmt"
	"io"
	"server/app/enum/secret_provider"
	"server/app/util/oss"
	"slices"
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

// Client 提供华为云 OBS 的对象存储实现。
type Client struct {
	secret oss.Secret
	config oss.Config
}

var _ oss.Backend = (*Client)(nil)

// init 注册华为云 OBS实现，供统一入口按厂商创建客户端。
func init() {
	oss.Register(secret_provider.HuaweiCloud, func(secret oss.SecretSource, config oss.ConfigSource) (oss.Backend, error) {
		return New(secret, config)
	})
}

// New 使用独立的密钥和连接配置创建客户端，不发起网络请求。
func New(secret oss.SecretSource, source oss.ConfigSource) (*Client, error) {
	key, c, err := oss.NormalizeConfig(secret, source, secret_provider.HuaweiCloud)
	if err != nil {
		return nil, err
	}
	if c.Region == "" && c.Endpoint == "" {
		return nil, fmt.Errorf("%w: 华为云必须提供 Region 或 Endpoint", oss.ErrInvalidArgument)
	}
	if c.Endpoint == "" {
		c.Endpoint = "https://obs." + c.Region + ".myhuaweicloud.com"
	}
	h := &Client{secret: key, config: c}
	if _, err = h.client(context.Background()); err != nil {
		return nil, err
	}
	return h, nil
}

// client 为每次调用绑定独立 context，并复用 HTTP 连接池，避免并发调用相互覆盖取消状态。
func (h *Client) client(ctx context.Context) (*obs.ObsClient, error) {
	return obs.New(h.secret.AccessKeyID, h.secret.AccessKeySecret, h.config.Endpoint,
		obs.WithSecurityToken(h.secret.SecurityToken), obs.WithRegion(h.config.Region), obs.WithSignature(obs.SignatureObs),
		obs.WithHttpClient(h.config.HTTPClient), obs.WithSslVerify(true), obs.WithRequestContext(ctx), obs.WithMaxRetryCount(0), obs.WithMaxRedirectCount(0))
}

// contextError 转换厂商错误，并保留取消和超时原因供 errors.Is 判断。
func (h *Client) contextError(ctx context.Context, op, key string, err error) error {
	if err == nil {
		return nil
	}
	// 部分 SDK 将传输错误转换为字符串；恢复调用方可判定的取消原因。
	if cause := ctx.Err(); cause != nil {
		err = errors.Join(cause, err)
	}
	var serviceErr obs.ObsError
	if errors.As(err, &serviceErr) {
		return &oss.Error{Provider: h.config.Provider, Op: op, Key: key, StatusCode: serviceErr.StatusCode, Code: serviceErr.Code, Message: serviceErr.Message, Err: err}
	}
	return oss.WrapError(h.config.Provider, op, key, err)
}

// ListBuckets 读取账户下的存储桶列表。
func (h *Client) ListBuckets(ctx context.Context) ([]oss.Bucket, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	var buckets []oss.Bucket
	marker := ""
	for {
		out, err := c.ListBuckets(&obs.ListBucketsInput{QueryLocation: true, MaxKeys: 1000, Marker: marker})
		if err != nil {
			return nil, h.contextError(ctx, "ListBuckets", "", err)
		}
		buckets = slices.Grow(buckets, len(out.Buckets))
		for _, b := range out.Buckets {
			buckets = append(buckets, oss.Bucket{Name: b.Name, Region: b.Location, CreatedAt: b.CreationDate})
		}
		if !out.IsTruncated {
			return buckets, nil
		}
		if out.NextMarker == "" || out.NextMarker == marker {
			return nil, fmt.Errorf("OBS 返回无效的存储桶分页游标")
		}
		marker = out.NextMarker
	}
}

// HeadBucket 检查配置的存储桶是否存在且可访问。
func (h *Client) HeadBucket(ctx context.Context) error {
	c, err := h.client(ctx)
	if err != nil {
		return err
	}
	_, err = c.HeadBucket(h.config.Bucket)
	return h.contextError(ctx, "HeadBucket", "", err)
}

// CreateBucket 在配置的区域创建存储桶。
func (h *Client) CreateBucket(ctx context.Context) error {
	c, err := h.client(ctx)
	if err != nil {
		return err
	}
	in := &obs.CreateBucketInput{Bucket: h.config.Bucket}
	in.Location = h.config.Region
	_, err = c.CreateBucket(in)
	return h.contextError(ctx, "CreateBucket", "", err)
}

// DeleteBucket 删除配置的空存储桶。
func (h *Client) DeleteBucket(ctx context.Context) error {
	c, err := h.client(ctx)
	if err != nil {
		return err
	}
	_, err = c.DeleteBucket(h.config.Bucket)
	return h.contextError(ctx, "DeleteBucket", "", err)
}

// ListFiles 按前缀和游标分页列出文件。
func (h *Client) ListFiles(ctx context.Context, o oss.ListOptions) (*oss.FilePage, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	marker := o.Token
	if marker == "" {
		marker = o.StartAfter
	}
	in := &obs.ListObjectsInput{Bucket: h.config.Bucket, Marker: marker}
	in.Prefix = o.Prefix
	in.Delimiter = o.Delimiter
	in.MaxKeys = o.Limit
	out, err := c.ListObjects(in)
	if err != nil {
		return nil, h.contextError(ctx, "ListFiles", "", err)
	}
	page := &oss.FilePage{
		Files:          slices.Grow([]oss.File(nil), len(out.Contents)),
		CommonPrefixes: out.CommonPrefixes,
		NextToken:      out.NextMarker,
		HasMore:        out.IsTruncated,
	}
	for _, v := range out.Contents {
		page.Files = append(page.Files, oss.File{Key: v.Key, Size: v.Size, ETag: v.ETag, LastModified: v.LastModified, StorageClass: string(v.StorageClass)})
	}
	return page, nil
}

// huaweiFile 将 OBS 文件元数据转换为统一文件信息。
func (h *Client) huaweiFile(key string, v obs.GetObjectMetadataOutput) oss.File {
	return oss.File{
		Key:                key,
		Size:               v.ContentLength,
		ETag:               v.ETag,
		LastModified:       v.LastModified,
		ContentType:        v.ContentType,
		CacheControl:       v.CacheControl,
		ContentDisposition: v.ContentDisposition,
		ContentEncoding:    v.ContentEncoding,
		StorageClass:       string(v.StorageClass),
		Metadata:           v.Metadata,
	}
}

// StatFile 读取文件元数据，不下载正文。
func (h *Client) StatFile(ctx context.Context, key string) (*oss.File, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	v, err := c.GetObjectMetadata(&obs.GetObjectMetadataInput{Bucket: h.config.Bucket, Key: key})
	if err != nil {
		return nil, h.contextError(ctx, "StatFile", key, err)
	}
	o := h.huaweiFile(key, *v)
	return &o, nil
}

// GetFile 流式读取文件，支持范围和 ETag 条件。
func (h *Client) GetFile(ctx context.Context, key string, o oss.GetOptions) (*oss.Download, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	in := &obs.GetObjectInput{GetObjectMetadataInput: obs.GetObjectMetadataInput{Bucket: h.config.Bucket, Key: key}, Range: o.Range, IfMatch: o.IfMatch}
	v, err := c.GetObject(in)
	if err != nil {
		return nil, h.contextError(ctx, "GetFile", key, err)
	}
	file := h.huaweiFile(key, v.GetObjectMetadataOutput)
	file.Size = -1
	contentRange := ""
	for name, values := range v.ResponseHeaders {
		if len(values) == 0 {
			continue
		}
		if strings.EqualFold(name, "Content-Range") {
			contentRange = values[0]
		}
		if strings.EqualFold(name, "Content-Length") && values[0] != "" {
			file.Size = v.ContentLength
		}
	}
	return &oss.Download{File: file, Body: v.Body, ContentRange: contentRange}, nil
}

// huaweiHeaders 提取上传和复制共用的内容头。
func (h *Client) huaweiHeaders(o oss.PutOptions) obs.HttpHeader {
	return obs.HttpHeader{
		ContentType:        o.ContentType,
		CacheControl:       o.CacheControl,
		ContentDisposition: o.ContentDisposition,
		ContentEncoding:    o.ContentEncoding,
	}
}

// operation 组装当前桶的文件标识、存储类型和用户元数据。
func (h *Client) operation(key string, o oss.PutOptions) obs.ObjectOperationInput {
	return obs.ObjectOperationInput{Bucket: h.config.Bucket, Key: key, StorageClass: obs.StorageClassType(o.StorageClass), Metadata: o.Metadata}
}

// PutFile 上传指定长度的数据流。
func (h *Client) PutFile(ctx context.Context, key string, r io.Reader, size int64, o oss.PutOptions) (*oss.File, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	v, err := c.PutObject(&obs.PutObjectInput{PutObjectBasicInput: obs.PutObjectBasicInput{ObjectOperationInput: h.operation(key, o), HttpHeader: h.huaweiHeaders(o), ContentLength: size}, Body: r})
	if err != nil {
		return nil, h.contextError(ctx, "PutFile", key, err)
	}
	return &oss.File{Key: key, Size: size, ETag: v.ETag, ContentType: o.ContentType, Metadata: o.Metadata}, nil
}

// DeleteFile 删除指定文件。
func (h *Client) DeleteFile(ctx context.Context, key string) error {
	c, err := h.client(ctx)
	if err != nil {
		return err
	}
	_, err = c.DeleteObject(&obs.DeleteObjectInput{Bucket: h.config.Bucket, Key: key})
	return h.contextError(ctx, "DeleteFile", key, err)
}

// CopyFile 在服务端复制文件，可同时替换元数据。
func (h *Client) CopyFile(ctx context.Context, source, target string, o oss.CopyOptions) (*oss.File, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	in := &obs.CopyObjectInput{ObjectOperationInput: h.operation(target, oss.PutOptions{}), CopySourceBucket: h.config.Bucket, CopySourceKey: source, MetadataDirective: obs.CopyMetadata}
	if o.ReplaceMetadata {
		in.ObjectOperationInput = h.operation(target, o.PutOptions)
		in.MetadataDirective = obs.ReplaceMetadata
		in.HttpHeader = h.huaweiHeaders(o.PutOptions)
	}
	v, err := c.CopyObject(in)
	if err != nil {
		return nil, h.contextError(ctx, "CopyFile", target, err)
	}
	return &oss.File{Key: target, ETag: v.ETag, LastModified: v.LastModified}, nil
}

// InitiateMultipartUpload 初始化分片上传并返回可持久化的上传 ID。
func (h *Client) InitiateMultipartUpload(ctx context.Context, key string, o oss.PutOptions) (*oss.MultipartUpload, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	v, err := c.InitiateMultipartUpload(&obs.InitiateMultipartUploadInput{ObjectOperationInput: h.operation(key, o), HttpHeader: h.huaweiHeaders(o)})
	if err != nil {
		return nil, h.contextError(ctx, "InitiateMultipartUpload", key, err)
	}
	return &oss.MultipartUpload{Key: key, UploadID: v.UploadId}, nil
}

// UploadPart 上传一个分片，同序号分片可重新上传。
func (h *Client) UploadPart(ctx context.Context, key, id string, n int, r io.Reader, size int64) (*oss.Part, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	v, err := c.UploadPart(&obs.UploadPartInput{Bucket: h.config.Bucket, Key: key, UploadId: id, PartNumber: n, PartSize: size, Body: r})
	if err != nil {
		return nil, h.contextError(ctx, "UploadPart", key, err)
	}
	return &oss.Part{Number: n, ETag: v.ETag, Size: size}, nil
}

// CompleteMultipartUpload 合并已上传分片。
func (h *Client) CompleteMultipartUpload(ctx context.Context, key, id string, parts []oss.Part) (*oss.File, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	in := &obs.CompleteMultipartUploadInput{
		Bucket:   h.config.Bucket,
		Key:      key,
		UploadId: id,
		Parts:    slices.Grow([]obs.Part(nil), len(parts)),
	}
	var size int64
	for _, p := range parts {
		in.Parts = append(in.Parts, obs.Part{PartNumber: p.Number, ETag: p.ETag})
		size += p.Size
	}
	v, err := c.CompleteMultipartUpload(in)
	if err != nil {
		return nil, h.contextError(ctx, "CompleteMultipartUpload", key, err)
	}
	if v.ETag == "" {
		return nil, &oss.Error{Provider: h.config.Provider, Op: "CompleteMultipartUpload", Key: key, StatusCode: 200, Code: "InvalidResponse", Message: "合并响应缺少 ETag"}
	}
	return &oss.File{Key: key, ETag: v.ETag, Size: size}, nil
}

// AbortMultipartUpload 终止分片上传并清理服务端分片。
func (h *Client) AbortMultipartUpload(ctx context.Context, key, id string) error {
	c, err := h.client(ctx)
	if err != nil {
		return err
	}
	_, err = c.AbortMultipartUpload(&obs.AbortMultipartUploadInput{Bucket: h.config.Bucket, Key: key, UploadId: id})
	return h.contextError(ctx, "AbortMultipartUpload", key, err)
}

// ListParts 分页查询已上传分片。
func (h *Client) ListParts(ctx context.Context, key, id string, o oss.PartsOptions) (*oss.PartsPage, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	v, err := c.ListParts(&obs.ListPartsInput{Bucket: h.config.Bucket, Key: key, UploadId: id, MaxParts: o.Limit, PartNumberMarker: o.Marker})
	if err != nil {
		return nil, h.contextError(ctx, "ListParts", key, err)
	}
	page := &oss.PartsPage{
		Parts:      slices.Grow([]oss.Part(nil), len(v.Parts)),
		NextMarker: v.NextPartNumberMarker,
		HasMore:    v.IsTruncated,
	}
	for _, p := range v.Parts {
		page.Parts = append(page.Parts, oss.Part{Number: p.PartNumber, ETag: p.ETag, Size: p.Size, LastModified: p.LastModified})
	}
	return page, nil
}

// ListMultipartUploads 分页查询尚未完成的分片上传。
func (h *Client) ListMultipartUploads(ctx context.Context, o oss.UploadsOptions) (*oss.UploadsPage, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	v, err := c.ListMultipartUploads(&obs.ListMultipartUploadsInput{Bucket: h.config.Bucket, Prefix: o.Prefix, KeyMarker: o.KeyMarker, UploadIdMarker: o.UploadIDMarker, MaxUploads: o.Limit})
	if err != nil {
		return nil, h.contextError(ctx, "ListMultipartUploads", "", err)
	}
	page := &oss.UploadsPage{
		Uploads:            slices.Grow([]oss.MultipartUpload(nil), len(v.Uploads)),
		NextKeyMarker:      v.NextKeyMarker,
		NextUploadIDMarker: v.NextUploadIdMarker,
		HasMore:            v.IsTruncated,
	}
	for _, u := range v.Uploads {
		page.Uploads = append(page.Uploads, oss.MultipartUpload{Key: u.Key, UploadID: u.UploadId, Initiated: u.Initiated})
	}
	return page, nil
}

// Sign 生成有有效期的签名访问地址。
func (h *Client) Sign(ctx context.Context, key, method string, ttl time.Duration, options ...oss.SignOptions) (*oss.SignedURL, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
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
	v, err := c.CreateSignedUrl(&obs.CreateSignedUrlInput{Bucket: h.config.Bucket, Key: key, Method: obs.HttpMethodType(method), Expires: int(ttl / time.Second), QueryParams: query})
	if err != nil {
		return nil, h.contextError(ctx, "Sign", key, err)
	}
	return &oss.SignedURL{URL: v.SignedUrl, Headers: v.ActualSignedRequestHeaders, ExpiresAt: time.Now().Add(ttl)}, nil
}

// UploadPartCopy 在服务端按字节范围复制分片，源 ETag 条件防止复制期间文件改变。
func (h *Client) UploadPartCopy(ctx context.Context, key, uploadID string, number int, source string, o oss.CopyPartOptions) (*oss.Part, error) {
	client, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	input := &obs.CopyPartInput{
		Bucket:           h.config.Bucket,
		Key:              key,
		UploadId:         uploadID,
		PartNumber:       number,
		CopySourceBucket: h.config.Bucket,
		CopySourceKey:    source,
		CopySourceRange:  fmt.Sprintf("bytes=%d-%d", o.Offset, o.Offset+o.Size-1),
	}
	var result *obs.CopyPartOutput
	if o.IfMatch != "" {
		result, err = client.CopyPart(input, obs.WithCustomHeader("x-obs-copy-source-if-match", o.IfMatch))
	} else {
		result, err = client.CopyPart(input)
	}
	if err != nil {
		return nil, h.contextError(ctx, "UploadPartCopy", key, err)
	}
	return &oss.Part{Number: number, ETag: result.ETag, Size: o.Size, LastModified: result.LastModified}, nil
}
