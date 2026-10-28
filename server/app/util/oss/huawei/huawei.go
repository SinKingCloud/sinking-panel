package huawei

import (
	"context"
	"errors"
	"fmt"
	"io"
	"server/app/util/oss/types"
	"slices"
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

// Client 提供华为云 OBS 的对象存储实现。
type Client struct {
	config types.Config
}

// New 使用已标准化的配置创建华为云 OBS 客户端。
func New(c types.Config) (*Client, error) {
	h := &Client{config: c}
	_, err := h.client(context.Background())
	return h, err
}

// client 为每次调用绑定独立 context，并复用 HTTP 连接池，避免并发调用相互覆盖取消状态。
func (h *Client) client(ctx context.Context) (*obs.ObsClient, error) {
	return obs.New(h.config.AccessKeyID, h.config.AccessKeySecret, h.config.Endpoint,
		obs.WithSecurityToken(h.config.SecurityToken), obs.WithRegion(h.config.Region), obs.WithSignature(obs.SignatureObs),
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
		return &types.Error{Provider: h.config.Provider, Op: op, Key: key, StatusCode: serviceErr.StatusCode, Code: serviceErr.Code, Message: serviceErr.Message, Err: err}
	}
	return types.WrapError(h.config.Provider, op, key, err)
}

// ListBuckets 读取账户下的存储桶列表。
func (h *Client) ListBuckets(ctx context.Context) ([]types.Bucket, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	var buckets []types.Bucket
	marker := ""
	for {
		out, err := c.ListBuckets(&obs.ListBucketsInput{QueryLocation: true, MaxKeys: 1000, Marker: marker})
		if err != nil {
			return nil, h.contextError(ctx, "ListBuckets", "", err)
		}
		buckets = slices.Grow(buckets, len(out.Buckets))
		for _, b := range out.Buckets {
			buckets = append(buckets, types.Bucket{Name: b.Name, Region: b.Location, CreatedAt: b.CreationDate})
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

// ListObjects 按前缀和游标分页列出对象。
func (h *Client) ListObjects(ctx context.Context, o types.ListOptions) (*types.ObjectPage, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	in := &obs.ListObjectsInput{Bucket: h.config.Bucket, Marker: o.Token}
	in.Prefix = o.Prefix
	in.Delimiter = o.Delimiter
	in.MaxKeys = o.Limit
	out, err := c.ListObjects(in)
	if err != nil {
		return nil, h.contextError(ctx, "ListObjects", "", err)
	}
	page := &types.ObjectPage{
		Objects:        slices.Grow([]types.Object(nil), len(out.Contents)),
		CommonPrefixes: out.CommonPrefixes,
		NextToken:      out.NextMarker,
		HasMore:        out.IsTruncated,
	}
	for _, v := range out.Contents {
		page.Objects = append(page.Objects, types.Object{Key: v.Key, Size: v.Size, ETag: v.ETag, LastModified: v.LastModified, StorageClass: string(v.StorageClass)})
	}
	return page, nil
}

// huaweiObject 将 OBS 对象元数据转换为统一文件信息。
func huaweiObject(key string, v obs.GetObjectMetadataOutput) types.Object {
	return types.Object{
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

// StatObject 读取对象元数据，不下载正文。
func (h *Client) StatObject(ctx context.Context, key string) (*types.Object, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	v, err := c.GetObjectMetadata(&obs.GetObjectMetadataInput{Bucket: h.config.Bucket, Key: key})
	if err != nil {
		return nil, h.contextError(ctx, "StatObject", key, err)
	}
	o := huaweiObject(key, *v)
	return &o, nil
}

// GetObject 流式读取对象，支持范围和 ETag 条件。
func (h *Client) GetObject(ctx context.Context, key string, o types.GetOptions) (*types.Download, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	in := &obs.GetObjectInput{GetObjectMetadataInput: obs.GetObjectMetadataInput{Bucket: h.config.Bucket, Key: key}, Range: o.Range, IfMatch: o.IfMatch}
	v, err := c.GetObject(in)
	if err != nil {
		return nil, h.contextError(ctx, "GetObject", key, err)
	}
	contentRange := ""
	for key, values := range v.ResponseHeaders {
		if strings.EqualFold(key, "Content-Range") && len(values) > 0 {
			contentRange = values[0]
			break
		}
	}
	return &types.Download{Object: huaweiObject(key, v.GetObjectMetadataOutput), Body: v.Body, ContentRange: contentRange}, nil
}

// huaweiHeaders 提取上传和复制共用的内容头。
func huaweiHeaders(o types.PutOptions) obs.HttpHeader {
	return obs.HttpHeader{
		ContentType:        o.ContentType,
		CacheControl:       o.CacheControl,
		ContentDisposition: o.ContentDisposition,
		ContentEncoding:    o.ContentEncoding,
	}
}

// operation 组装当前桶的对象标识、存储类型和用户元数据。
func (h *Client) operation(key string, o types.PutOptions) obs.ObjectOperationInput {
	return obs.ObjectOperationInput{Bucket: h.config.Bucket, Key: key, StorageClass: obs.StorageClassType(o.StorageClass), Metadata: o.Metadata}
}

// PutObject 上传指定长度的数据流。
func (h *Client) PutObject(ctx context.Context, key string, r io.Reader, size int64, o types.PutOptions) (*types.Object, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	v, err := c.PutObject(&obs.PutObjectInput{PutObjectBasicInput: obs.PutObjectBasicInput{ObjectOperationInput: h.operation(key, o), HttpHeader: huaweiHeaders(o), ContentLength: size}, Body: r})
	if err != nil {
		return nil, h.contextError(ctx, "PutObject", key, err)
	}
	return &types.Object{Key: key, Size: size, ETag: v.ETag, ContentType: o.ContentType, Metadata: o.Metadata}, nil
}

// DeleteObject 删除指定对象。
func (h *Client) DeleteObject(ctx context.Context, key string) error {
	c, err := h.client(ctx)
	if err != nil {
		return err
	}
	_, err = c.DeleteObject(&obs.DeleteObjectInput{Bucket: h.config.Bucket, Key: key})
	return h.contextError(ctx, "DeleteObject", key, err)
}

// CopyObject 在服务端复制对象，可同时替换元数据。
func (h *Client) CopyObject(ctx context.Context, source, target string, o types.CopyOptions) (*types.Object, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	bucket := o.SourceBucket
	if bucket == "" {
		bucket = h.config.Bucket
	}
	in := &obs.CopyObjectInput{ObjectOperationInput: h.operation(target, types.PutOptions{}), CopySourceBucket: bucket, CopySourceKey: source, MetadataDirective: obs.CopyMetadata}
	if o.ReplaceMetadata {
		in.ObjectOperationInput = h.operation(target, o.PutOptions)
		in.MetadataDirective = obs.ReplaceMetadata
		in.HttpHeader = huaweiHeaders(o.PutOptions)
	}
	v, err := c.CopyObject(in)
	if err != nil {
		return nil, h.contextError(ctx, "CopyObject", target, err)
	}
	return &types.Object{Key: target, ETag: v.ETag, LastModified: v.LastModified}, nil
}

// InitiateMultipartUpload 初始化分片上传并返回可持久化的上传 ID。
func (h *Client) InitiateMultipartUpload(ctx context.Context, key string, o types.PutOptions) (*types.MultipartUpload, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	v, err := c.InitiateMultipartUpload(&obs.InitiateMultipartUploadInput{ObjectOperationInput: h.operation(key, o), HttpHeader: huaweiHeaders(o)})
	if err != nil {
		return nil, h.contextError(ctx, "InitiateMultipartUpload", key, err)
	}
	return &types.MultipartUpload{Key: key, UploadID: v.UploadId}, nil
}

// UploadPart 上传一个分片，同序号分片可重新上传。
func (h *Client) UploadPart(ctx context.Context, key, id string, n int, r io.Reader, size int64) (*types.Part, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	v, err := c.UploadPart(&obs.UploadPartInput{Bucket: h.config.Bucket, Key: key, UploadId: id, PartNumber: n, PartSize: size, Body: r})
	if err != nil {
		return nil, h.contextError(ctx, "UploadPart", key, err)
	}
	return &types.Part{Number: n, ETag: v.ETag, Size: size}, nil
}

// CompleteMultipartUpload 合并已上传分片。
func (h *Client) CompleteMultipartUpload(ctx context.Context, key, id string, parts []types.Part) (*types.Object, error) {
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
		return nil, &types.Error{Provider: h.config.Provider, Op: "CompleteMultipartUpload", Key: key, StatusCode: 200, Code: "InvalidResponse", Message: "合并响应缺少 ETag"}
	}
	return &types.Object{Key: key, ETag: v.ETag, Size: size}, nil
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
func (h *Client) ListParts(ctx context.Context, key, id string, o types.PartsOptions) (*types.PartsPage, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	v, err := c.ListParts(&obs.ListPartsInput{Bucket: h.config.Bucket, Key: key, UploadId: id, MaxParts: o.Limit, PartNumberMarker: o.Marker})
	if err != nil {
		return nil, h.contextError(ctx, "ListParts", key, err)
	}
	page := &types.PartsPage{
		Parts:      slices.Grow([]types.Part(nil), len(v.Parts)),
		NextMarker: v.NextPartNumberMarker,
		HasMore:    v.IsTruncated,
	}
	for _, p := range v.Parts {
		page.Parts = append(page.Parts, types.Part{Number: p.PartNumber, ETag: p.ETag, Size: p.Size, LastModified: p.LastModified})
	}
	return page, nil
}

// ListMultipartUploads 分页查询尚未完成的分片上传。
func (h *Client) ListMultipartUploads(ctx context.Context, o types.UploadsOptions) (*types.UploadsPage, error) {
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	v, err := c.ListMultipartUploads(&obs.ListMultipartUploadsInput{Bucket: h.config.Bucket, Prefix: o.Prefix, KeyMarker: o.KeyMarker, UploadIdMarker: o.UploadIDMarker, MaxUploads: o.Limit})
	if err != nil {
		return nil, h.contextError(ctx, "ListMultipartUploads", "", err)
	}
	page := &types.UploadsPage{
		Uploads:            slices.Grow([]types.MultipartUpload(nil), len(v.Uploads)),
		NextKeyMarker:      v.NextKeyMarker,
		NextUploadIDMarker: v.NextUploadIdMarker,
		HasMore:            v.IsTruncated,
	}
	for _, u := range v.Uploads {
		page.Uploads = append(page.Uploads, types.MultipartUpload{Key: u.Key, UploadID: u.UploadId, Initiated: u.Initiated})
	}
	return page, nil
}

// Presign 生成有有效期的签名访问地址。
func (h *Client) Presign(ctx context.Context, key, method string, ttl time.Duration) (*types.SignedURL, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	v, err := c.CreateSignedUrl(&obs.CreateSignedUrlInput{Bucket: h.config.Bucket, Key: key, Method: obs.HttpMethodType(method), Expires: int(ttl / time.Second)})
	if err != nil {
		return nil, h.contextError(ctx, "Presign", key, err)
	}
	return &types.SignedURL{URL: v.SignedUrl, Headers: v.ActualSignedRequestHeaders, ExpiresAt: time.Now().Add(ttl)}, nil
}

// UploadPartCopy 在服务端按字节范围复制分片，源 ETag 条件防止复制期间对象改变。
func (h *Client) UploadPartCopy(ctx context.Context, key, uploadID string, number int, source string, o types.CopyPartOptions) (*types.Part, error) {
	client, err := h.client(ctx)
	if err != nil {
		return nil, err
	}
	bucket := o.SourceBucket
	if bucket == "" {
		bucket = h.config.Bucket
	}
	input := &obs.CopyPartInput{
		Bucket:           h.config.Bucket,
		Key:              key,
		UploadId:         uploadID,
		PartNumber:       number,
		CopySourceBucket: bucket,
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
	return &types.Part{Number: number, ETag: result.ETag, Size: o.Size, LastModified: result.LastModified}, nil
}
