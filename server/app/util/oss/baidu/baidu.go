package baidu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"server/app/util/oss/types"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
)

// Client 提供百度云 BOS 的对象存储实现。
type Client struct {
	config types.Config
	client *bos.Client
}

// New 使用已标准化的配置创建百度云 BOS 客户端。
func New(c types.Config) (*Client, error) {
	conf := bos.NewBosClientConfig(c.AccessKeyID, c.AccessKeySecret, c.Endpoint).WithRetryPolicy(bce.NewNoRetryPolicy())
	conf.HTTPClient = c.HTTPClient
	client, err := bos.NewClientWithConfig(conf)
	if err != nil {
		return nil, err
	}
	client.Config.Credentials.SessionToken = c.SecurityToken
	return &Client{config: c, client: client}, nil
}

// contextError 转换厂商错误，并保留取消和超时原因供 errors.Is 判断。
func (b *Client) contextError(ctx context.Context, op, key string, err error) error {
	if err == nil {
		return nil
	}
	// 部分 SDK 将传输错误转换为字符串；恢复调用方可判定的取消原因。
	if cause := ctx.Err(); cause != nil {
		err = errors.Join(cause, err)
	}
	var serviceErr *bce.BceServiceError
	if errors.As(err, &serviceErr) {
		return &types.Error{Provider: b.config.Provider, Op: op, Key: key, StatusCode: serviceErr.StatusCode, Code: serviceErr.Code, Message: serviceErr.Message, Err: err}
	}
	return types.WrapError(b.config.Provider, op, key, err)
}

// baiduTime 兼容 BOS 响应中的 RFC3339 和 HTTP 时间格式。
func baiduTime(s string) time.Time {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t
	}
	t, _ := http.ParseTime(s)
	return t
}

// ListBuckets 读取账户下的存储桶列表。
func (b *Client) ListBuckets(ctx context.Context) ([]types.Bucket, error) {
	v, err := b.client.ListBucketsWithContext(ctx)
	if err != nil {
		return nil, b.contextError(ctx, "ListBuckets", "", err)
	}
	out := make([]types.Bucket, 0, len(v.Buckets))
	for _, x := range v.Buckets {
		out = append(out, types.Bucket{Name: x.Name, Region: x.Location, CreatedAt: baiduTime(x.CreationDate)})
	}
	return out, nil
}

// HeadBucket 检查配置的存储桶是否存在且可访问。
func (b *Client) HeadBucket(ctx context.Context) error {
	return b.contextError(ctx, "HeadBucket", "", b.client.HeadBucketWithContext(ctx, b.config.Bucket))
}

// CreateBucket 在配置的区域创建存储桶，使用官方 api 层传递独立调用上下文。
func (b *Client) CreateBucket(ctx context.Context) error {
	_, err := api.PutBucket(b.client, b.config.Bucket, nil, b.client.NewBosContext(ctx))
	return b.contextError(ctx, "CreateBucket", "", err)
}

// DeleteBucket 删除配置的空存储桶。
func (b *Client) DeleteBucket(ctx context.Context) error {
	return b.contextError(ctx, "DeleteBucket", "", api.DeleteBucket(b.client, b.config.Bucket, b.client.NewBosContext(ctx)))
}

// ListObjects 按前缀和游标分页列出对象。
func (b *Client) ListObjects(ctx context.Context, o types.ListOptions) (*types.ObjectPage, error) {
	v, err := b.client.ListObjectsWithContext(ctx, b.config.Bucket, &api.ListObjectsArgs{Prefix: o.Prefix, Delimiter: o.Delimiter, Marker: o.Token, MaxKeys: o.Limit})
	if err != nil {
		return nil, b.contextError(ctx, "ListObjects", "", err)
	}
	page := &types.ObjectPage{
		Objects:        slices.Grow([]types.Object(nil), len(v.Contents)),
		CommonPrefixes: slices.Grow([]string(nil), len(v.CommonPrefixes)),
		NextToken:      v.NextMarker,
		HasMore:        v.IsTruncated,
	}
	for _, x := range v.Contents {
		page.Objects = append(page.Objects, types.Object{Key: x.Key, Size: int64(x.Size), ETag: x.ETag, LastModified: baiduTime(x.LastModified), StorageClass: x.StorageClass})
	}
	for _, p := range v.CommonPrefixes {
		page.CommonPrefixes = append(page.CommonPrefixes, p.Prefix)
	}
	return page, nil
}

// baiduObject 转换对象元数据，并统一用户元数据键的大小写。
func baiduObject(key string, v api.ObjectMeta) types.Object {
	metadata := make(map[string]string, len(v.UserMeta))
	for key, value := range v.UserMeta {
		metadata[strings.ToLower(key)] = value
	}
	return types.Object{
		Key:                key,
		Size:               v.ContentLength,
		ETag:               v.ETag,
		LastModified:       baiduTime(v.LastModified),
		ContentType:        v.ContentType,
		CacheControl:       v.CacheControl,
		ContentDisposition: v.ContentDisposition,
		ContentEncoding:    v.ContentEncoding,
		StorageClass:       v.StorageClass,
		Metadata:           metadata,
	}
}

// StatObject 读取对象元数据，不下载正文。
func (b *Client) StatObject(ctx context.Context, key string) (*types.Object, error) {
	v, err := b.client.GetObjectMetaWithContext(ctx, b.config.Bucket, key)
	if err != nil {
		return nil, b.contextError(ctx, "StatObject", key, err)
	}
	o := baiduObject(key, v.ObjectMeta)
	return &o, nil
}

// GetObject 流式读取对象，支持范围和 ETag 条件。
func (b *Client) GetObject(ctx context.Context, key string, o types.GetOptions) (*types.Download, error) {
	var options []api.Option
	if o.Range != "" {
		options = append(options, api.SetHeader("Range", o.Range))
	}
	v, err := b.client.GetObjectWithArgs(ctx, b.config.Bucket, key, &api.GetObjectArgs{IfMatch: o.IfMatch}, options...)
	if err != nil {
		return nil, b.contextError(ctx, "GetObject", key, err)
	}
	return &types.Download{Object: baiduObject(key, v.ObjectMeta), Body: v.Body, ContentRange: v.ContentRange}, nil
}

// PutObject 上传指定长度的数据流。
func (b *Client) PutObject(ctx context.Context, key string, r io.Reader, size int64, o types.PutOptions) (*types.Object, error) {
	// NewBodyFromSizedReader 会读取整个文件；NewBodyFromReader 保持流式传输。
	body, err := bce.NewBodyFromReader(r, size)
	if err != nil {
		return nil, err
	}
	tag, err := b.client.PutObjectWithContext(ctx, b.config.Bucket, key, body, &api.PutObjectArgs{ContentLength: size, ContentType: o.ContentType, CacheControl: o.CacheControl, ContentDisposition: o.ContentDisposition, ContentEncoding: o.ContentEncoding, StorageClass: o.StorageClass, UserMeta: o.Metadata})
	if err != nil {
		return nil, b.contextError(ctx, "PutObject", key, err)
	}
	return &types.Object{Key: key, Size: size, ETag: tag, ContentType: o.ContentType, Metadata: o.Metadata}, nil
}

// DeleteObject 删除指定对象。
func (b *Client) DeleteObject(ctx context.Context, key string) error {
	return b.contextError(ctx, "DeleteObject", key, api.DeleteObject(b.client, b.config.Bucket, key, "", b.client.NewBosContext(ctx)))
}

// CopyObject 在服务端复制对象，可同时替换元数据。
func (b *Client) CopyObject(ctx context.Context, source, target string, o types.CopyOptions) (*types.Object, error) {
	bucket := o.SourceBucket
	if bucket == "" {
		bucket = b.config.Bucket
	}
	args := &api.CopyObjectArgs{MetadataDirective: "copy"}
	if o.ReplaceMetadata {
		args.MetadataDirective = "replace"
		args.ObjectMeta = api.ObjectMeta{ContentType: o.PutOptions.ContentType, CacheControl: o.PutOptions.CacheControl, ContentDisposition: o.PutOptions.ContentDisposition, ContentEncoding: o.PutOptions.ContentEncoding, StorageClass: o.PutOptions.StorageClass, UserMeta: o.PutOptions.Metadata}
	}
	v, err := b.client.CopyObjectWithContext(ctx, b.config.Bucket, target, bucket, source, args)
	if err != nil {
		return nil, b.contextError(ctx, "CopyObject", target, err)
	}
	return &types.Object{Key: target, ETag: v.ETag, LastModified: baiduTime(v.LastModified)}, nil
}

// InitiateMultipartUpload 初始化分片上传并返回可持久化的上传 ID。
func (b *Client) InitiateMultipartUpload(ctx context.Context, key string, o types.PutOptions) (*types.MultipartUpload, error) {
	options := make([]api.Option, 0, len(o.Metadata))
	for k, v := range o.Metadata {
		options = append(options, api.UserMeta(k, v))
	}
	v, err := api.InitiateMultipartUpload(b.client, b.config.Bucket, key, o.ContentType, &api.InitiateMultipartUploadArgs{CacheControl: o.CacheControl, ContentDisposition: o.ContentDisposition, ContentEncoding: o.ContentEncoding, StorageClass: o.StorageClass}, b.client.NewBosContext(ctx), options...)
	if err != nil {
		return nil, b.contextError(ctx, "InitiateMultipartUpload", key, err)
	}
	return &types.MultipartUpload{Key: key, UploadID: v.UploadId}, nil
}

// UploadPart 上传一个分片，同序号分片可重新上传。
func (b *Client) UploadPart(ctx context.Context, key, id string, n int, r io.Reader, size int64) (*types.Part, error) {
	body, err := bce.NewBodyFromReader(r, size)
	if err != nil {
		return nil, err
	}
	tag, err := b.client.UploadPartWithContext(ctx, b.config.Bucket, key, id, n, body, nil)
	if err != nil {
		return nil, b.contextError(ctx, "UploadPart", key, err)
	}
	return &types.Part{Number: n, Size: size, ETag: tag}, nil
}

// CompleteMultipartUpload 合并已上传分片。
func (b *Client) CompleteMultipartUpload(ctx context.Context, key, id string, parts []types.Part) (*types.Object, error) {
	args := &api.CompleteMultipartUploadArgs{Parts: slices.Grow([]api.UploadInfoType(nil), len(parts))}
	var size int64
	for _, p := range parts {
		args.Parts = append(args.Parts, api.UploadInfoType{PartNumber: p.Number, ETag: p.ETag})
		size += p.Size
	}
	data, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	body, err := bce.NewBodyFromBytes(data)
	if err != nil {
		return nil, err
	}
	// BOS 可能在 HTTP 200 的响应正文中返回合并错误；SDK 的结果类型未声明错误字段。
	req := &api.BosRequest{}
	req.SetUri("/" + b.config.Bucket + "/" + key)
	req.SetMethod(http.MethodPost)
	req.SetParam("uploadId", id)
	req.SetBucket(b.config.Bucket)
	req.SetObject(key)
	req.SetIsObjectReq(true)
	req.SetHeader("Content-Type", "application/json")
	req.SetBody(body)
	resp := &api.BosResponse{}
	if err = api.SendRequest(b.client, req, resp, b.client.NewBosContext(ctx)); err != nil {
		return nil, b.contextError(ctx, "CompleteMultipartUpload", key, err)
	}
	if resp.IsFail() {
		return nil, b.contextError(ctx, "CompleteMultipartUpload", key, resp.ServiceError())
	}
	var result struct {
		api.CompleteMultipartUploadResult
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err = resp.ParseJsonBody(&result); err != nil {
		return nil, b.contextError(ctx, "CompleteMultipartUpload", key, err)
	}
	if result.Code != "" {
		return nil, &types.Error{Provider: b.config.Provider, Op: "CompleteMultipartUpload", Key: key, StatusCode: 200, Code: result.Code, Message: result.Message}
	}
	if result.ETag == "" {
		return nil, &types.Error{Provider: b.config.Provider, Op: "CompleteMultipartUpload", Key: key, StatusCode: 200, Code: "InvalidResponse", Message: "合并响应缺少 ETag"}
	}
	return &types.Object{Key: key, Size: size, ETag: result.ETag}, nil
}

// AbortMultipartUpload 终止分片上传并清理服务端分片。
func (b *Client) AbortMultipartUpload(ctx context.Context, key, id string) error {
	return b.contextError(ctx, "AbortMultipartUpload", key, api.AbortMultipartUpload(b.client, b.config.Bucket, key, id, b.client.NewBosContext(ctx)))
}

// ListParts 分页查询已上传分片。
func (b *Client) ListParts(ctx context.Context, key, id string, o types.PartsOptions) (*types.PartsPage, error) {
	marker := ""
	if o.Marker != 0 {
		marker = strconv.Itoa(o.Marker)
	}
	v, err := b.client.ListPartsWithContext(ctx, b.config.Bucket, key, id, &api.ListPartsArgs{MaxParts: o.Limit, PartNumberMarker: marker})
	if err != nil {
		return nil, b.contextError(ctx, "ListParts", key, err)
	}
	page := &types.PartsPage{
		Parts:      slices.Grow([]types.Part(nil), len(v.Parts)),
		NextMarker: v.NextPartNumberMarker,
		HasMore:    v.IsTruncated,
	}
	for _, p := range v.Parts {
		page.Parts = append(page.Parts, types.Part{Number: p.PartNumber, ETag: p.ETag, Size: int64(p.Size), LastModified: baiduTime(p.LastModified)})
	}
	return page, nil
}

// ListMultipartUploads 分页查询尚未完成的分片上传。
func (b *Client) ListMultipartUploads(ctx context.Context, o types.UploadsOptions) (*types.UploadsPage, error) {
	// SDK 的结果结构遗漏 uploadIdMarker，使用 SDK 请求/签名层保留两个分页游标。
	req := &api.BosRequest{}
	req.SetUri("/" + b.config.Bucket)
	req.SetMethod(http.MethodGet)
	req.SetParam("uploads", "")
	req.SetBucket(b.config.Bucket)
	if o.Prefix != "" {
		req.SetParam("prefix", o.Prefix)
	}
	if o.KeyMarker != "" {
		req.SetParam("keyMarker", o.KeyMarker)
	}
	if o.UploadIDMarker != "" {
		req.SetParam("uploadIdMarker", o.UploadIDMarker)
	}
	req.SetParam("maxUploads", strconv.Itoa(o.Limit))
	resp := &api.BosResponse{}
	if err := api.SendRequest(b.client, req, resp, b.client.NewBosContext(ctx)); err != nil {
		return nil, b.contextError(ctx, "ListMultipartUploads", "", err)
	}
	if resp.IsFail() {
		return nil, b.contextError(ctx, "ListMultipartUploads", "", resp.ServiceError())
	}
	var v struct {
		api.ListMultipartUploadsResult
		NextUploadIDMarker string `json:"nextUploadIdMarker"`
	}
	if err := resp.ParseJsonBody(&v); err != nil {
		return nil, b.contextError(ctx, "ListMultipartUploads", "", err)
	}
	page := &types.UploadsPage{
		Uploads:            slices.Grow([]types.MultipartUpload(nil), len(v.Uploads)),
		NextKeyMarker:      v.NextKeyMarker,
		NextUploadIDMarker: v.NextUploadIDMarker,
		HasMore:            v.IsTruncated,
	}
	for _, u := range v.Uploads {
		page.Uploads = append(page.Uploads, types.MultipartUpload{Key: u.Key, UploadID: u.UploadId, Initiated: baiduTime(u.Initiated)})
	}
	return page, nil
}

// Presign 生成有有效期的签名访问地址。
func (b *Client) Presign(ctx context.Context, key, method string, ttl time.Duration) (*types.SignedURL, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	params := map[string]string{}
	if b.config.SecurityToken != "" {
		params["x-bce-security-token"] = b.config.SecurityToken
	}
	return &types.SignedURL{URL: b.client.GeneratePresignedUrl(b.config.Bucket, key, int(ttl/time.Second), method, nil, params), ExpiresAt: time.Now().Add(ttl)}, nil
}

// UploadPartCopy 在服务端按字节范围复制分片，源 ETag 条件防止复制期间对象改变。
func (b *Client) UploadPartCopy(ctx context.Context, key, uploadID string, number int, source string, o types.CopyPartOptions) (*types.Part, error) {
	bucket := o.SourceBucket
	if bucket == "" {
		bucket = b.config.Bucket
	}
	args := &api.UploadPartCopyArgs{
		SourceRange: fmt.Sprintf("bytes=%d-%d", o.Offset, o.Offset+o.Size-1),
		IfMatch:     o.IfMatch,
	}
	result, err := api.UploadPartCopy(b.client, b.config.Bucket, key, "/"+bucket+"/"+source, uploadID, number, args, b.client.NewBosContext(ctx))
	if err != nil {
		return nil, b.contextError(ctx, "UploadPartCopy", key, err)
	}
	return &types.Part{Number: number, ETag: result.ETag, Size: o.Size, LastModified: baiduTime(result.LastModified)}, nil
}
