# types：对象存储公共数据模型

导入路径：`server/app/util/oss/types`。本包只定义五家厂商适配器共享的配置、请求、结果和错误；不初始化 SDK，不发起网络请求，不依赖根 `oss` 包。

根包通过类型别名暴露这些类型，因此业务通常直接使用 `oss.Config`、`oss.Object`；适配器使用 `types.Config`、`types.Object`，避免 `oss → 厂商 → oss` 循环依赖。根包 `oss.Error` 与 `types.Error` 是同一个类型，错误哨兵也共享。

## 配置

| Config 字段 | 说明 |
| --- | --- |
| `Provider` | 现有 `secret_provider` 的整数枚举，不另行定义厂商编号 |
| `Region` | 厂商区域标识；阿里、火山用于签名，必须提供 |
| `Endpoint` | 不含桶名的区域 HTTP(S) 地址；根包可按 Region 填默认值 |
| `Bucket` | Manager 绑定的桶；只列桶时可空 |
| `AccessKeyID` / `AccessKeySecret` | 原始凭据，不能传脱敏值 |
| `SecurityToken` | 可选临时令牌，不会自动刷新 |
| `HTTPClient` | 自定义代理、Transport、期限；根包复制并标准化配置 |

密钥、临时令牌和 HTTPClient 的 JSON 标签均为 `-`。JSON 序列化不会输出它们，但不能据此把完整 Config 当作任意格式的安全日志对象；调用方仍应避免打印凭据字段。

## 查询与返回值

| 类型 | 用途与字段约定 |
| --- | --- |
| `Bucket` | 名称、Region、CreatedAt；不同服务可能不提供所有信息 |
| `Object` | Key、Size（字节）、ETag、LastModified、内容头、StorageClass、Metadata |
| `ListOptions` | Prefix、Delimiter、Token、Limit；Limit=0 在根包采用 1000 |
| `ObjectPage` | Objects、CommonPrefixes、HasMore、NextToken；没有总数或任意跳页 |
| `GetOptions` | 单个 HTTP Range 和可选 IfMatch ETag |
| `Download` | Object、Body、ContentRange；Body 是需关闭的 `io.ReadCloser` |
| `SignedURL` | URL、必须携带的 Headers、ExpiresAt |
| `DeleteResult` | 单个删除 Key 和 Err；用于批量操作部分成功结果 |

写入、复制返回的 Object 仅包含当次响应及请求中已知的字段，需要完整信息时再调用 `StatObject`。范围下载时 Size 对应响应正文长度，完整对象范围可从 ContentRange 判断。ETag 按厂商原样保存，不应假定总是文件 MD5。

## 写入、复制和分片

`PutOptions` 包括 ContentType、CacheControl、ContentDisposition、ContentEncoding、StorageClass 和用户 Metadata。StorageClass 仍使用各厂商原生取值。`CopyOptions.ReplaceMetadata` 默认 false；true 时用 PutOptions 整体替换，未传字段不会自动保留。

`CopyPartOptions` 包括可选 SourceBucket、从 0 开始的 Offset、Size 和 IfMatch。范围终点是 `Offset + Size - 1`。例如 Offset=0、Size=1 表示第一个字节，不能当作未设置范围。

`Part` 的 Number、ETag、Size 是合并所必需的信息；Number 从 1 开始。`PartsOptions.Marker` 是最后一个分片编号，`PartsPage.NextMarker` 用于后续分页。

`MultipartUpload` 中的 Key、UploadID 可以持久化以便恢复，但恢复前还应由业务核对原文件身份、长度及分片策略。`UploadsOptions` 和 `UploadsPage` 使用 KeyMarker、UploadIDMarker 双游标，两者一起回传。

`UploadOptions` 嵌入 PutOptions，并提供 PartSize、TempDir、Progress。Progress 接收已确认字节和总字节，总长度未知时 total=-1；回调在当前上传 goroutine 内串行执行，不应阻塞过久。缓冲、分片上限和失败清理策略由 [根包传输实现](../README.md) 定义。

## 统一错误

| 哨兵 | 判定 |
| --- | --- |
| `ErrInvalidArgument` | 根包校验失败时通过错误链返回 |
| `ErrNotFound` | HTTP 404 |
| `ErrPermissionDenied` | HTTP 401 或 403 |
| `ErrConflict` | HTTP 409 |
| `ErrPreconditionFailed` | HTTP 412 |

`Error` 保存 Provider、Op、Key、StatusCode、Code、Message、Err；`Error()` 输出操作摘要而不直接打印 Message/Err。`Unwrap()` 保留底层错误，`Is()` 按状态匹配上述分类。`WrapError` 对 nil 返回 nil，已有统一错误直接保留，其余错误包装为未指定状态的 Error。

```go
package example

import (
    "context"
    "errors"
    "fmt"
    "server/app/util/oss"
)

func FindObject(ctx context.Context, storage oss.Interface, key string) (*oss.Object, error) {
    object, err := storage.StatObject(ctx, key)
    if errors.Is(err, oss.ErrNotFound) {
        return nil, nil
    }
    if err != nil {
        var detail *oss.Error
        if errors.As(err, &detail) {
            return nil, fmt.Errorf("读取对象失败，厂商码 %s: %w", detail.Code, err)
        }
        return nil, err
    }
    return object, nil
}
```

不能把权限错误、网络超时或 context 取消当作“不存在”。适配器可能报告 HTTP 200 正文中的业务错误，此时应查看返回的 error 和 Code，而非只依赖状态码。

源码：[entities.go](entities.go)、[errors.go](errors.go)。公共调用入口：[oss](../README.md)。
