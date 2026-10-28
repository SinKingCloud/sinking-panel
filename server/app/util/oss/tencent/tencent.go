package tencent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"server/app/enum/secret_provider"
	"server/app/util/oss/types"
	"strconv"
	"strings"
	"time"

	cos "github.com/tencentyun/cos-go-sdk-v5"
)

const metadataPrefix = "x-cos-meta-"

// 腾讯 SDK 会把源键内的 ?versionId= 当成版本参数；在签名前设置完整源键，保证复制的是指定对象。
type copySourceKey struct{}

type authorizationTransport struct {
	*cos.AuthorizationTransport
}

// RoundTrip 在签名前恢复完整复制源，避免 SDK 将对象键中的版本参数片段误解析。
func (t *authorizationTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if source, ok := request.Context().Value(copySourceKey{}).(string); ok {
		request = request.Clone(request.Context())
		request.Header.Set("x-cos-copy-source", source)
	}
	return t.AuthorizationTransport.RoundTrip(request)
}

// Client 提供腾讯云 COS 的对象存储实现。
type Client struct {
	client   *cos.Client
	bucket   string
	endpoint *url.URL
}

// New 使用已标准化的配置创建腾讯云 COS 客户端。
func New(c types.Config) (*Client, error) {
	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = "https://cos." + c.Region + ".myqcloud.com"
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	bucketURL := *u
	if c.Bucket != "" {
		bucketURL.Host = c.Bucket + "." + u.Host
	}
	serviceURL := *u
	if strings.HasSuffix(u.Hostname(), ".myqcloud.com") {
		serviceURL.Host = "service.cos.myqcloud.com"
	}
	httpClient := &http.Client{}
	if c.HTTPClient != nil {
		*httpClient = *c.HTTPClient
	}
	httpClient.Transport = &authorizationTransport{AuthorizationTransport: &cos.AuthorizationTransport{SecretID: c.AccessKeyID, SecretKey: c.AccessKeySecret, SessionToken: c.SecurityToken, Transport: httpClient.Transport}}
	client := cos.NewClient(&cos.BaseURL{BucketURL: &bucketURL, ServiceURL: &serviceURL}, httpClient)
	// 对象键是原样的存储标识，不进行本地路径归一化。
	client.Conf.ObjectKeySimplifyCheck = false
	return &Client{client: client, bucket: c.Bucket, endpoint: u}, nil
}

// wrapError 提取 COS 错误码和状态码，并保留原始错误链。
func wrapError(op, key string, err error) error {
	if err == nil {
		return nil
	}
	var serviceErr *cos.ErrorResponse
	if errors.As(err, &serviceErr) {
		status := 0
		if serviceErr.Response != nil {
			status = serviceErr.Response.StatusCode
		}
		return &types.Error{Provider: secret_provider.TencentCloud, Op: op, Key: key, StatusCode: status, Code: serviceErr.Code, Message: serviceErr.Message, Err: err}
	}
	return types.WrapError(secret_provider.TencentCloud, op, key, err)
}

// parseTime 兼容响应正文的 RFC3339 时间与 HTTP 响应头时间。
func parseTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t, _ = http.ParseTime(value)
	}
	return t
}

// metadataHeaders 将用户元数据转换为 COS 请求头。
func metadataHeaders(values map[string]string) *http.Header {
	headers := make(http.Header, len(values))
	for key, value := range values {
		headers.Set(metadataPrefix+key, value)
	}
	return &headers
}

// putHeaders 组装简单上传和分片初始化共用的对象属性。
func putHeaders(options types.PutOptions) *cos.ObjectPutHeaderOptions {
	return &cos.ObjectPutHeaderOptions{
		ContentType:        options.ContentType,
		CacheControl:       options.CacheControl,
		ContentDisposition: options.ContentDisposition,
		ContentEncoding:    options.ContentEncoding,
		XCosStorageClass:   options.StorageClass,
		XCosMetaXXX:        metadataHeaders(options.Metadata),
	}
}

// objectFromResponse 从响应头提取对象信息，元数据键统一为小写。
func objectFromResponse(key string, result *cos.Response) types.Object {
	size, _ := strconv.ParseInt(result.Header.Get("Content-Length"), 10, 64)
	metadata := map[string]string{}
	for name, values := range result.Header {
		if len(values) > 0 && len(name) >= len(metadataPrefix) && strings.EqualFold(name[:len(metadataPrefix)], metadataPrefix) {
			metadata[strings.ToLower(name[len(metadataPrefix):])] = values[0]
		}
	}
	return types.Object{
		Key:                key,
		Size:               size,
		ETag:               result.Header.Get("ETag"),
		LastModified:       parseTime(result.Header.Get("Last-Modified")),
		ContentType:        result.Header.Get("Content-Type"),
		CacheControl:       result.Header.Get("Cache-Control"),
		ContentDisposition: result.Header.Get("Content-Disposition"),
		ContentEncoding:    result.Header.Get("Content-Encoding"),
		StorageClass:       result.Header.Get("x-cos-storage-class"),
		Metadata:           metadata,
	}
}

// ListBuckets 分页读取账户下的全部存储桶。
func (c *Client) ListBuckets(ctx context.Context) ([]types.Bucket, error) {
	out := []types.Bucket{}
	marker := ""
	for {
		result, _, err := c.client.Service.Get(ctx, &cos.ServiceGetOptions{MaxKeys: 1000, Marker: marker})
		if err != nil {
			return nil, wrapError("ListBuckets", "", err)
		}
		for _, v := range result.Buckets {
			out = append(out, types.Bucket{Name: v.Name, Region: v.Region, CreatedAt: parseTime(v.CreationDate)})
		}
		if !result.IsTruncated {
			return out, nil
		}
		if result.NextMarker == "" || result.NextMarker == marker {
			return nil, wrapError("ListBuckets", "", errors.New("服务端未返回有效分页游标"))
		}
		marker = result.NextMarker
	}
}

// HeadBucket 检查绑定的存储桶是否存在且可访问。
func (c *Client) HeadBucket(ctx context.Context) error {
	_, err := c.client.Bucket.Head(ctx)
	return wrapError("HeadBucket", "", err)
}

// CreateBucket 在配置区域创建绑定的存储桶。
func (c *Client) CreateBucket(ctx context.Context) error {
	_, err := c.client.Bucket.Put(ctx, nil)
	return wrapError("CreateBucket", "", err)
}

// DeleteBucket 删除绑定的空存储桶。
func (c *Client) DeleteBucket(ctx context.Context) error {
	_, err := c.client.Bucket.Delete(ctx)
	return wrapError("DeleteBucket", "", err)
}

// ListObjects 按前缀、分隔符和游标查询一页对象。
func (c *Client) ListObjects(ctx context.Context, options types.ListOptions) (*types.ObjectPage, error) {
	result, _, err := c.client.Bucket.Get(ctx, &cos.BucketGetOptions{
		Prefix:       options.Prefix,
		Delimiter:    options.Delimiter,
		Marker:       options.Token,
		MaxKeys:      options.Limit,
		EncodingType: "url",
	})
	if err != nil {
		return nil, wrapError("ListObjects", "", err)
	}
	decode := func(value string) string {
		if result.EncodingType == "url" {
			if decoded, e := url.PathUnescape(value); e == nil {
				return decoded
			}
		}
		return value
	}
	out := &types.ObjectPage{
		Objects:        make([]types.Object, 0, len(result.Contents)),
		CommonPrefixes: make([]string, 0, len(result.CommonPrefixes)),
		NextToken:      decode(result.NextMarker),
		HasMore:        result.IsTruncated,
	}
	for _, v := range result.Contents {
		out.Objects = append(out.Objects, types.Object{
			Key:          decode(v.Key),
			Size:         v.Size,
			ETag:         v.ETag,
			LastModified: parseTime(v.LastModified),
			StorageClass: v.StorageClass,
		})
	}
	for _, v := range result.CommonPrefixes {
		out.CommonPrefixes = append(out.CommonPrefixes, decode(v))
	}
	// 未指定分隔符时部分兼容节点省略 NextMarker，最后一个对象键就是后续游标。
	if out.HasMore && out.NextToken == "" && len(out.Objects) > 0 {
		out.NextToken = out.Objects[len(out.Objects)-1].Key
	}
	return out, nil
}

// StatObject 读取对象大小、时间和元数据，不下载正文。
func (c *Client) StatObject(ctx context.Context, key string) (*types.Object, error) {
	result, err := c.client.Object.Head(ctx, key, nil)
	if err != nil {
		return nil, wrapError("StatObject", key, err)
	}
	out := objectFromResponse(key, result)
	return &out, nil
}

// GetObject 打开对象下载流，支持字节范围和 ETag 条件，调用方负责关闭 Body。
func (c *Client) GetObject(ctx context.Context, key string, options types.GetOptions) (*types.Download, error) {
	h := http.Header{}
	if options.IfMatch != "" {
		h.Set("If-Match", options.IfMatch)
	}
	result, err := c.client.Object.Get(ctx, key, &cos.ObjectGetOptions{Range: options.Range, XOptionHeader: &h})
	if err != nil {
		return nil, wrapError("GetObject", key, err)
	}
	return &types.Download{Object: objectFromResponse(key, result), Body: result.Body, ContentRange: result.Header.Get("Content-Range")}, nil
}

// PutObject 上传指定长度的数据流，并返回已确认的对象信息。
func (c *Client) PutObject(ctx context.Context, key string, body io.Reader, size int64, options types.PutOptions) (*types.Object, error) {
	headers := putHeaders(options)
	headers.ContentLength = size
	if size == 0 {
		body = strings.NewReader("")
	}
	result, err := c.client.Object.Put(ctx, key, body, &cos.ObjectPutOptions{ObjectPutHeaderOptions: headers})
	if err != nil {
		return nil, wrapError("PutObject", key, err)
	}
	return &types.Object{
		Key:                key,
		Size:               size,
		ETag:               result.Header.Get("ETag"),
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
	_, err := c.client.Object.Delete(ctx, key)
	return wrapError("DeleteObject", key, err)
}

// CopyObject 在服务端复制对象，按选项保留或替换元数据。
func (c *Client) CopyObject(ctx context.Context, source, key string, options types.CopyOptions) (*types.Object, error) {
	headers := &cos.ObjectCopyHeaderOptions{}
	if options.ReplaceMetadata {
		p := options.PutOptions
		headers.XCosMetadataDirective = "Replaced"
		headers.ContentType = p.ContentType
		headers.CacheControl = p.CacheControl
		headers.ContentDisposition = p.ContentDisposition
		headers.ContentEncoding = p.ContentEncoding
		headers.XCosStorageClass = p.StorageClass
		headers.XCosMetaXXX = metadataHeaders(p.Metadata)
	}
	ctx, copySource := c.copySource(ctx, options.SourceBucket, source)
	result, _, err := c.client.Object.Copy(ctx, key, copySource, &cos.ObjectCopyOptions{ObjectCopyHeaderOptions: headers})
	if err != nil {
		return nil, wrapError("CopyObject", key, err)
	}
	return &types.Object{Key: key, ETag: result.ETag, LastModified: parseTime(result.LastModified)}, nil
}

// InitiateMultipartUpload 初始化分片上传会话，并设置最终对象的元数据。
func (c *Client) InitiateMultipartUpload(ctx context.Context, key string, options types.PutOptions) (*types.MultipartUpload, error) {
	result, _, err := c.client.Object.InitiateMultipartUpload(ctx, key, &cos.InitiateMultipartUploadOptions{ObjectPutHeaderOptions: putHeaders(options)})
	if err != nil {
		return nil, wrapError("InitiateMultipartUpload", key, err)
	}
	return &types.MultipartUpload{Key: key, UploadID: result.UploadID, Initiated: time.Now()}, nil
}

// UploadPart 上传指定编号的分片，同编号重传会替换原分片。
// 腾讯 SDK 直接拼接 uploadId，必须先转义，避免临时会话中的 +、& 等字符改变请求参数。
func (c *Client) UploadPart(ctx context.Context, key, id string, number int, body io.Reader, size int64) (*types.Part, error) {
	result, err := c.client.Object.UploadPart(ctx, key, url.QueryEscape(id), number, body, &cos.ObjectUploadPartOptions{ContentLength: size})
	if err != nil {
		return nil, wrapError("UploadPart", key, err)
	}
	return &types.Part{Number: number, Size: size, ETag: result.Header.Get("ETag")}, nil
}

// CompleteMultipartUpload 按传入顺序合并分片，并返回最终对象信息。
func (c *Client) CompleteMultipartUpload(ctx context.Context, key, id string, parts []types.Part) (*types.Object, error) {
	list := make([]cos.Object, 0, len(parts))
	var size int64
	for _, p := range parts {
		list = append(list, cos.Object{PartNumber: p.Number, ETag: p.ETag})
		size += p.Size
	}
	result, _, err := c.client.Object.CompleteMultipartUpload(ctx, key, url.QueryEscape(id), &cos.CompleteMultipartUploadOptions{Parts: list})
	if err != nil {
		return nil, wrapError("CompleteMultipartUpload", key, err)
	}
	return &types.Object{Key: key, Size: size, ETag: result.ETag}, nil
}

// AbortMultipartUpload 中止分片会话并清理尚未合并的分片。
func (c *Client) AbortMultipartUpload(ctx context.Context, key, id string) error {
	_, err := c.client.Object.AbortMultipartUpload(ctx, key, url.QueryEscape(id))
	return wrapError("AbortMultipartUpload", key, err)
}

// ListParts 按分片编号游标查询已上传分片。
func (c *Client) ListParts(ctx context.Context, key, id string, options types.PartsOptions) (*types.PartsPage, error) {
	result, _, err := c.client.Object.ListParts(ctx, key, url.QueryEscape(id), &cos.ObjectListPartsOptions{
		MaxParts:         strconv.Itoa(options.Limit),
		PartNumberMarker: strconv.Itoa(options.Marker),
	})
	if err != nil {
		return nil, wrapError("ListParts", key, err)
	}
	marker, _ := strconv.Atoi(result.NextPartNumberMarker)
	out := &types.PartsPage{
		Parts:      make([]types.Part, 0, len(result.Parts)),
		NextMarker: marker,
		HasMore:    result.IsTruncated,
	}
	for _, v := range result.Parts {
		out.Parts = append(out.Parts, types.Part{Number: v.PartNumber, Size: v.Size, ETag: v.ETag, LastModified: parseTime(v.LastModified)})
	}
	return out, nil
}

// ListMultipartUploads 按对象和上传 ID 双游标查询未完成会话。
func (c *Client) ListMultipartUploads(ctx context.Context, options types.UploadsOptions) (*types.UploadsPage, error) {
	result, _, err := c.client.Bucket.ListMultipartUploads(ctx, &cos.ListMultipartUploadsOptions{
		Prefix:         options.Prefix,
		KeyMarker:      options.KeyMarker,
		UploadIDMarker: options.UploadIDMarker,
		MaxUploads:     options.Limit,
	})
	if err != nil {
		return nil, wrapError("ListMultipartUploads", "", err)
	}
	out := &types.UploadsPage{
		Uploads:            make([]types.MultipartUpload, 0, len(result.Uploads)),
		HasMore:            result.IsTruncated,
		NextKeyMarker:      result.NextKeyMarker,
		NextUploadIDMarker: result.NextUploadIDMarker,
	}
	for _, v := range result.Uploads {
		out.Uploads = append(out.Uploads, types.MultipartUpload{Key: v.Key, UploadID: v.UploadID, Initiated: parseTime(v.Initiated)})
	}
	return out, nil
}

// Presign 生成包含认证参数的临时访问地址。
func (c *Client) Presign(ctx context.Context, key, method string, expiry time.Duration) (*types.SignedURL, error) {
	if err := ctx.Err(); err != nil {
		return nil, wrapError("Presign", key, err)
	}
	result, err := c.client.Object.GetPresignedURL2(ctx, method, key, expiry, nil)
	if err != nil {
		return nil, wrapError("Presign", key, err)
	}
	return &types.SignedURL{URL: result.String(), ExpiresAt: time.Now().Add(expiry)}, nil
}

// UploadPartCopy 使用源 ETag 约束每个分片，防止复制期间源对象变化。
func (c *Client) UploadPartCopy(ctx context.Context, key, id string, number int, source string, options types.CopyPartOptions) (*types.Part, error) {
	ctx, copySource := c.copySource(ctx, options.SourceBucket, source)
	result, _, err := c.client.Object.CopyPart(ctx, key, url.QueryEscape(id), number, copySource, &cos.ObjectCopyPartOptions{
		XCosCopySourceRange:   fmt.Sprintf("bytes=%d-%d", options.Offset, options.Offset+options.Size-1),
		XCosCopySourceIfMatch: options.IfMatch,
	})
	if err != nil {
		return nil, wrapError("UploadPartCopy", key, err)
	}
	return &types.Part{
		Number:       number,
		ETag:         result.ETag,
		Size:         options.Size,
		LastModified: parseTime(result.LastModified),
	}, nil
}

// copySource 生成 SDK 复制源，同时在 context 中保留逐字节转义的签名头。
// 整体转义后仅还原路径分隔符，避免为每一级目录分配切片和临时字符串。
func (c *Client) copySource(ctx context.Context, bucket, key string) (context.Context, string) {
	if bucket == "" {
		bucket = c.bucket
	}
	prefix := bucket + "." + c.endpoint.Host + "/"
	escapedKey := strings.ReplaceAll(url.QueryEscape(key), "+", "%20")
	escapedKey = strings.ReplaceAll(escapedKey, "%2F", "/")
	return context.WithValue(ctx, copySourceKey{}, prefix+escapedKey), prefix + key
}
