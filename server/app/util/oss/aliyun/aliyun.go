package aliyun

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"server/app/enum/secret_provider"
	"server/app/util/oss/types"
	"time"

	ali "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// Client 提供阿里云 OSS 的对象存储实现。
type Client struct {
	client *ali.Client
	bucket string
}

// New 使用已标准化的配置创建阿里云 OSS 客户端。
func New(c types.Config) (*Client, error) {
	cfg := ali.LoadDefaultConfig().WithRegion(c.Region).WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.AccessKeySecret, c.SecurityToken))
	if c.Endpoint != "" {
		cfg.WithEndpoint(c.Endpoint)
	}
	if c.HTTPClient != nil {
		cfg.WithHttpClient(c.HTTPClient)
	}
	return &Client{client: ali.NewClient(cfg), bucket: c.Bucket}, nil
}

// wrapError 提取 OSS 错误码和状态码，并保留原始错误链。
func wrapError(op, key string, err error) error {
	if err == nil {
		return nil
	}
	var serviceErr *ali.ServiceError
	if errors.As(err, &serviceErr) {
		return &types.Error{Provider: secret_provider.Aliyun, Op: op, Key: key, StatusCode: serviceErr.StatusCode, Code: serviceErr.Code, Message: serviceErr.Message, Err: err}
	}
	return types.WrapError(secret_provider.Aliyun, op, key, err)
}

// optionalString 将空字符串转换为未设置的 SDK 可选字段。
func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// ListBuckets 分页读取账户下的全部存储桶。
func (c *Client) ListBuckets(ctx context.Context) ([]types.Bucket, error) {
	out := []types.Bucket{}
	var marker *string
	for {
		result, err := c.client.ListBuckets(ctx, &ali.ListBucketsRequest{Marker: marker, MaxKeys: 1000})
		if err != nil {
			return nil, wrapError("ListBuckets", "", err)
		}
		for _, v := range result.Buckets {
			out = append(out, types.Bucket{Name: ali.ToString(v.Name), Region: ali.ToString(v.Region), CreatedAt: ali.ToTime(v.CreationDate)})
		}
		if !result.IsTruncated {
			return out, nil
		}
		if ali.ToString(result.NextMarker) == "" || ali.ToString(result.NextMarker) == ali.ToString(marker) {
			return nil, wrapError("ListBuckets", "", errors.New("服务端未返回有效分页游标"))
		}
		marker = result.NextMarker
	}
}

// HeadBucket 通过 GetBucketInfo 检查桶是否可访问，需要读取桶信息的权限。
func (c *Client) HeadBucket(ctx context.Context) error {
	_, err := c.client.GetBucketInfo(ctx, &ali.GetBucketInfoRequest{Bucket: &c.bucket})
	return wrapError("HeadBucket", "", err)
}

// CreateBucket 在配置区域创建绑定的存储桶。
func (c *Client) CreateBucket(ctx context.Context) error {
	_, err := c.client.PutBucket(ctx, &ali.PutBucketRequest{Bucket: &c.bucket})
	return wrapError("CreateBucket", "", err)
}

// DeleteBucket 删除绑定的空存储桶。
func (c *Client) DeleteBucket(ctx context.Context) error {
	_, err := c.client.DeleteBucket(ctx, &ali.DeleteBucketRequest{Bucket: &c.bucket})
	return wrapError("DeleteBucket", "", err)
}

// ListObjects 按前缀、分隔符和游标查询一页对象。
func (c *Client) ListObjects(ctx context.Context, options types.ListOptions) (*types.ObjectPage, error) {
	result, err := c.client.ListObjectsV2(ctx, &ali.ListObjectsV2Request{
		Bucket:            &c.bucket,
		Prefix:            optionalString(options.Prefix),
		Delimiter:         optionalString(options.Delimiter),
		ContinuationToken: optionalString(options.Token),
		MaxKeys:           int32(options.Limit),
		EncodingType:      ali.Ptr("url"),
	})
	if err != nil {
		return nil, wrapError("ListObjects", "", err)
	}
	out := &types.ObjectPage{
		Objects:        make([]types.Object, 0, len(result.Contents)),
		CommonPrefixes: make([]string, 0, len(result.CommonPrefixes)),
		NextToken:      ali.ToString(result.NextContinuationToken),
		HasMore:        result.IsTruncated,
	}
	for _, v := range result.Contents {
		out.Objects = append(out.Objects, types.Object{
			Key:          ali.ToString(v.Key),
			Size:         v.Size,
			ETag:         ali.ToString(v.ETag),
			LastModified: ali.ToTime(v.LastModified),
			StorageClass: ali.ToString(v.StorageClass),
		})
	}
	for _, v := range result.CommonPrefixes {
		out.CommonPrefixes = append(out.CommonPrefixes, ali.ToString(v.Prefix))
	}
	return out, nil
}

// StatObject 读取对象大小、时间和元数据，不下载正文。
func (c *Client) StatObject(ctx context.Context, key string) (*types.Object, error) {
	result, err := c.client.HeadObject(ctx, &ali.HeadObjectRequest{Bucket: &c.bucket, Key: &key})
	if err != nil {
		return nil, wrapError("StatObject", key, err)
	}
	return &types.Object{
		Key:                key,
		Size:               result.ContentLength,
		ETag:               ali.ToString(result.ETag),
		LastModified:       ali.ToTime(result.LastModified),
		ContentType:        ali.ToString(result.ContentType),
		StorageClass:       ali.ToString(result.StorageClass),
		Metadata:           result.Metadata,
		CacheControl:       ali.ToString(result.CacheControl),
		ContentDisposition: ali.ToString(result.ContentDisposition),
		ContentEncoding:    ali.ToString(result.ContentEncoding),
	}, nil
}

// GetObject 打开对象下载流，支持字节范围和 ETag 条件，调用方负责关闭 Body。
func (c *Client) GetObject(ctx context.Context, key string, options types.GetOptions) (*types.Download, error) {
	result, err := c.client.GetObject(ctx, &ali.GetObjectRequest{Bucket: &c.bucket, Key: &key, Range: optionalString(options.Range), IfMatch: optionalString(options.IfMatch)})
	if err != nil {
		return nil, wrapError("GetObject", key, err)
	}
	return &types.Download{
		Body:         result.Body,
		ContentRange: ali.ToString(result.ContentRange),
		Object: types.Object{
			Key:                key,
			Size:               result.ContentLength,
			ETag:               ali.ToString(result.ETag),
			LastModified:       ali.ToTime(result.LastModified),
			ContentType:        ali.ToString(result.ContentType),
			StorageClass:       ali.ToString(result.StorageClass),
			Metadata:           result.Metadata,
			CacheControl:       result.Headers.Get("Cache-Control"),
			ContentDisposition: result.Headers.Get("Content-Disposition"),
			ContentEncoding:    result.Headers.Get("Content-Encoding"),
		},
	}, nil
}

// PutObject 上传指定长度的数据流，并返回已确认的对象信息。
func (c *Client) PutObject(ctx context.Context, key string, body io.Reader, size int64, options types.PutOptions) (*types.Object, error) {
	result, err := c.client.PutObject(ctx, &ali.PutObjectRequest{
		Bucket:             &c.bucket,
		Key:                &key,
		Body:               body,
		ContentLength:      &size,
		ContentType:        optionalString(options.ContentType),
		CacheControl:       optionalString(options.CacheControl),
		ContentDisposition: optionalString(options.ContentDisposition),
		ContentEncoding:    optionalString(options.ContentEncoding),
		StorageClass:       ali.StorageClassType(options.StorageClass),
		Metadata:           options.Metadata,
	})
	if err != nil {
		return nil, wrapError("PutObject", key, err)
	}
	return &types.Object{
		Key:                key,
		Size:               size,
		ETag:               ali.ToString(result.ETag),
		ContentType:        options.ContentType,
		CacheControl:       options.CacheControl,
		ContentDisposition: options.ContentDisposition,
		ContentEncoding:    options.ContentEncoding,
		StorageClass:       options.StorageClass,
		Metadata:           options.Metadata,
	}, nil
}

// DeleteObject 删除指定对象。
func (c *Client) DeleteObject(ctx context.Context, key string) error {
	_, err := c.client.DeleteObject(ctx, &ali.DeleteObjectRequest{Bucket: &c.bucket, Key: &key})
	return wrapError("DeleteObject", key, err)
}

// CopyObject 在服务端复制对象，按选项保留或替换元数据。
func (c *Client) CopyObject(ctx context.Context, source, key string, options types.CopyOptions) (*types.Object, error) {
	bucket := options.SourceBucket
	if bucket == "" {
		bucket = c.bucket
	}
	request := &ali.CopyObjectRequest{Bucket: &c.bucket, Key: &key, SourceBucket: &bucket, SourceKey: &source}
	if options.ReplaceMetadata {
		p := options.PutOptions
		request.MetadataDirective = ali.Ptr("REPLACE")
		request.ContentType = optionalString(p.ContentType)
		request.CacheControl = optionalString(p.CacheControl)
		request.ContentDisposition = optionalString(p.ContentDisposition)
		request.ContentEncoding = optionalString(p.ContentEncoding)
		request.StorageClass = ali.StorageClassType(p.StorageClass)
		request.Metadata = p.Metadata
	}
	result, err := c.client.CopyObject(ctx, request)
	if err != nil {
		return nil, wrapError("CopyObject", key, err)
	}
	return &types.Object{Key: key, ETag: ali.ToString(result.ETag), LastModified: ali.ToTime(result.LastModified)}, nil
}

// InitiateMultipartUpload 初始化分片上传会话，并设置最终对象的元数据。
func (c *Client) InitiateMultipartUpload(ctx context.Context, key string, options types.PutOptions) (*types.MultipartUpload, error) {
	result, err := c.client.InitiateMultipartUpload(ctx, &ali.InitiateMultipartUploadRequest{
		Bucket:             &c.bucket,
		Key:                &key,
		ContentType:        optionalString(options.ContentType),
		CacheControl:       optionalString(options.CacheControl),
		ContentDisposition: optionalString(options.ContentDisposition),
		ContentEncoding:    optionalString(options.ContentEncoding),
		StorageClass:       ali.StorageClassType(options.StorageClass),
		Metadata:           options.Metadata,
	})
	if err != nil {
		return nil, wrapError("InitiateMultipartUpload", key, err)
	}
	return &types.MultipartUpload{Key: key, UploadID: ali.ToString(result.UploadId), Initiated: time.Now()}, nil
}

// UploadPart 上传指定编号的分片，同编号重传会替换原分片。
func (c *Client) UploadPart(ctx context.Context, key, id string, number int, body io.Reader, size int64) (*types.Part, error) {
	result, err := c.client.UploadPart(ctx, &ali.UploadPartRequest{
		Bucket:        &c.bucket,
		Key:           &key,
		UploadId:      &id,
		PartNumber:    int32(number),
		Body:          body,
		ContentLength: &size,
	})
	if err != nil {
		return nil, wrapError("UploadPart", key, err)
	}
	return &types.Part{Number: number, Size: size, ETag: ali.ToString(result.ETag)}, nil
}

// CompleteMultipartUpload 按传入顺序合并分片，并返回最终对象信息。
func (c *Client) CompleteMultipartUpload(ctx context.Context, key, id string, parts []types.Part) (*types.Object, error) {
	list := make([]ali.UploadPart, 0, len(parts))
	var size int64
	for _, p := range parts {
		list = append(list, ali.UploadPart{PartNumber: int32(p.Number), ETag: ali.Ptr(p.ETag)})
		size += p.Size
	}
	result, err := c.client.CompleteMultipartUpload(ctx, &ali.CompleteMultipartUploadRequest{
		Bucket:                  &c.bucket,
		Key:                     &key,
		UploadId:                &id,
		CompleteMultipartUpload: &ali.CompleteMultipartUpload{Parts: list},
	})
	if err != nil {
		return nil, wrapError("CompleteMultipartUpload", key, err)
	}
	return &types.Object{Key: key, Size: size, ETag: ali.ToString(result.ETag)}, nil
}

// AbortMultipartUpload 中止分片会话并清理尚未合并的分片。
func (c *Client) AbortMultipartUpload(ctx context.Context, key, id string) error {
	_, err := c.client.AbortMultipartUpload(ctx, &ali.AbortMultipartUploadRequest{Bucket: &c.bucket, Key: &key, UploadId: &id})
	return wrapError("AbortMultipartUpload", key, err)
}

// ListParts 按分片编号游标查询已上传分片。
func (c *Client) ListParts(ctx context.Context, key, id string, options types.PartsOptions) (*types.PartsPage, error) {
	result, err := c.client.ListParts(ctx, &ali.ListPartsRequest{
		Bucket:           &c.bucket,
		Key:              &key,
		UploadId:         &id,
		MaxParts:         int32(options.Limit),
		PartNumberMarker: int32(options.Marker),
	})
	if err != nil {
		return nil, wrapError("ListParts", key, err)
	}
	out := &types.PartsPage{
		Parts:      make([]types.Part, 0, len(result.Parts)),
		NextMarker: int(result.NextPartNumberMarker),
		HasMore:    result.IsTruncated,
	}
	for _, v := range result.Parts {
		out.Parts = append(out.Parts, types.Part{Number: int(v.PartNumber), ETag: ali.ToString(v.ETag), Size: v.Size, LastModified: ali.ToTime(v.LastModified)})
	}
	return out, nil
}

// ListMultipartUploads 按对象和上传 ID 双游标查询未完成会话。
func (c *Client) ListMultipartUploads(ctx context.Context, options types.UploadsOptions) (*types.UploadsPage, error) {
	result, err := c.client.ListMultipartUploads(ctx, &ali.ListMultipartUploadsRequest{
		Bucket:         &c.bucket,
		Prefix:         optionalString(options.Prefix),
		KeyMarker:      optionalString(options.KeyMarker),
		UploadIdMarker: optionalString(options.UploadIDMarker),
		MaxUploads:     int32(options.Limit),
	})
	if err != nil {
		return nil, wrapError("ListMultipartUploads", "", err)
	}
	out := &types.UploadsPage{
		Uploads:            make([]types.MultipartUpload, 0, len(result.Uploads)),
		HasMore:            result.IsTruncated,
		NextKeyMarker:      ali.ToString(result.NextKeyMarker),
		NextUploadIDMarker: ali.ToString(result.NextUploadIdMarker),
	}
	for _, v := range result.Uploads {
		out.Uploads = append(out.Uploads, types.MultipartUpload{Key: ali.ToString(v.Key), UploadID: ali.ToString(v.UploadId), Initiated: ali.ToTime(v.Initiated)})
	}
	return out, nil
}

// Presign 生成带有效期的访问地址，并返回请求时需要携带的签名头。
func (c *Client) Presign(ctx context.Context, key, method string, expiry time.Duration) (*types.SignedURL, error) {
	var request any
	switch method {
	case http.MethodGet:
		request = &ali.GetObjectRequest{Bucket: &c.bucket, Key: &key}
	case http.MethodPut:
		request = &ali.PutObjectRequest{Bucket: &c.bucket, Key: &key}
	case http.MethodHead:
		request = &ali.HeadObjectRequest{Bucket: &c.bucket, Key: &key}
	default:
		return nil, wrapError("Presign", key, fmt.Errorf("不支持的签名方法: %s", method))
	}
	result, err := c.client.Presign(ctx, request, ali.PresignExpires(expiry))
	if err != nil {
		return nil, wrapError("Presign", key, err)
	}
	headers := make(http.Header, len(result.SignedHeaders))
	for k, v := range result.SignedHeaders {
		headers.Set(k, v)
	}
	return &types.SignedURL{URL: result.URL, ExpiresAt: result.Expiration, Headers: headers}, nil
}

// UploadPartCopy 在云端复制指定范围，备份移动过程中不经过本机下载。
func (c *Client) UploadPartCopy(ctx context.Context, key, id string, number int, source string, options types.CopyPartOptions) (*types.Part, error) {
	bucket := options.SourceBucket
	if bucket == "" {
		bucket = c.bucket
	}
	byteRange := fmt.Sprintf("bytes=%d-%d", options.Offset, options.Offset+options.Size-1)
	result, err := c.client.UploadPartCopy(ctx, &ali.UploadPartCopyRequest{
		Bucket:       &c.bucket,
		Key:          &key,
		UploadId:     &id,
		PartNumber:   int32(number),
		SourceBucket: &bucket,
		SourceKey:    &source,
		Range:        &byteRange,
		IfMatch:      optionalString(options.IfMatch),
	})
	if err != nil {
		return nil, wrapError("UploadPartCopy", key, err)
	}
	return &types.Part{
		Number:       number,
		ETag:         ali.ToString(result.ETag),
		Size:         options.Size,
		LastModified: ali.ToTime(result.LastModified),
	}, nil
}
