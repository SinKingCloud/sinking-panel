# baidu：百度云 BOS适配器

导入路径：`server/app/util/oss/baidu`。本子包将官方 `github.com/baidubce/bce-sdk-go/services/bos` 的请求和结果转换为 [oss/types](../types/README.md) 的公共数据结构。

业务代码优先从 [oss.NewManager](../README.md) 创建对象存储管理器。根包负责密钥校验、默认 Endpoint、HTTP 传输配置、参数校验、大文件自动分片、失败清理及文件操作组合；本子包是对应厂商的底层适配实现。

## 配置对照

| 项目 | 本厂商配置 |
| --- | --- |
| 密钥枚举 | `secret_provider.BaiduCloud` |
| 密钥管理字段 | `access_key_id / secret_access_key` |
| Region 示例 | `bj` |
| Endpoint 示例 | `https://bj.bcebos.com` |
| Bucket 示例 | `panel-backups` |
| 临时凭据 | `Config.SecurityToken`，和 AccessKeyID / AccessKeySecret 一起传入 |

Endpoint 是区域服务地址，不包含桶名、对象路径或查询参数。表中的桶名为示例，必须替换成自己有权限访问的桶。密钥由配置或现有密钥仓储注入，不在代码中写入真实凭据。

## 推荐使用示例

下面是完整函数，可以放入 `server` 模块中的业务包；它只列出一页备份对象，不创建或修改云资源。

```go
package example

import (
    "context"
    "server/app/enum/secret_provider"
    "server/app/util/oss"
)

func ListBackups(ctx context.Context, keyID, keySecret, token string) (*oss.ObjectPage, error) {
    manager, err := oss.NewManager(oss.Config{
        Provider: secret_provider.BaiduCloud,
        Region: "bj",
        Bucket: "panel-backups",
        AccessKeyID: keyID,
        AccessKeySecret: keySecret,
        SecurityToken: token,
    })
    if err != nil {
        return nil, err
    }
    return manager.ListObjects(ctx, oss.ListOptions{
        Prefix: "backups/",
        Delimiter: "/",
        Limit: 100,
    })
}
```

实际长期服务应复用 Manager。下一页把结果的 `NextToken` 原样赋给 `ListOptions.Token`，保持 Prefix 和 Delimiter 不变。上传、下载、续传、签名 URL 的完整用法见 [统一 OSS 文档](../README.md)。

## 直接调用子包

构造签名为 `New(config types.Config) (*Client, error)`，不发起网络请求。它接受**已经标准化的配置**，不会完整执行根包的校验和默认配置逻辑。直接调用时至少需要提供正确的厂商编号、Region、Endpoint、密钥、桶名和 HTTPClient；HTTPClient 应禁用重定向和透明 gzip 解压，避免签名请求被转发以及压缩备份被改变。

Client 提供以下底层操作，所有方法的首个参数均为非 nil 的 `context.Context`：

| 分类 | 方法及关键约定 |
| --- | --- |
| 桶 | `ListBuckets`、`HeadBucket`、`CreateBucket`、`DeleteBucket`；绑定桶来自 Config.Bucket |
| 查询 | `ListObjects(ctx, types.ListOptions)`、`StatObject(ctx, key)`；列表不保证含全部元数据 |
| 读写 | `GetObject(ctx, key, types.GetOptions)`、`PutObject(ctx, key, reader, size, types.PutOptions)`；下载 Body 必须关闭 |
| 删除与复制 | `DeleteObject`、`CopyObject`；普通复制仍受厂商大小限制，大对象应走根包自动分片复制 |
| 上传会话 | `InitiateMultipartUpload`、`UploadPart`、`ListParts`、`CompleteMultipartUpload`、`AbortMultipartUpload` |
| 分片复制 | `UploadPartCopy`；Offset 从 0 开始，范围终点为 Offset + Size - 1，IfMatch 固定源对象 |
| 会话浏览 | `ListMultipartUploads`，返回未完成上传及两个后续游标 |
| 临时访问 | `Presign(ctx, key, method, ttl)`，使用返回的 URL 与 Headers |

直接 Client 没有 `Upload`、`UploadFile`、`DownloadFile`、`MoveObject`、`Exists` 等组合方法；这些由根包 Manager 提供。直接上传要求准确声明长度，调用方负责失败会话清理和分片大小校验。对外业务不应依赖官方 SDK 的结果类型。

## 本适配器的行为与维护要点

### 区域

Region 示例为 bj，默认 Endpoint 是 bj.bcebos.com；不要将其他厂商的 cn-beijing 直接套用为 BOS 区域。也可显式提供有效服务 Endpoint。

### 流式请求

使用 NewBodyFromReader 保持数据流传输，避免 NewBodyFromSizedReader 把整个请求读入内存。关闭 SDK 自动重试，避免重放不可 seek 的流。

### 上下文

优先使用 WithContext API；未提供包装的方法通过官方 api 层传递独立 BosContext。取消和超时错误由适配器补齐错误链。

### 双游标

官方 ListMultipartUploadsResult 缺少 NextUploadIDMarker 字段，因此使用 SDK 签名请求层解析补充字段。翻页时必须保留两个游标。

### HTTP 200 内嵌错误

合并分片时仍检查正文 code/message 和 ETag，不把 HTTP 200 一律视作成功。响应元数据键统一转为小写。

## 错误和资源管理

- 使用 `errors.Is(err, oss.ErrNotFound)` 或 `errors.As(err, &detail)` 检查统一错误；不要仅匹配中文错误文本。
- 流式下载必须 `defer result.Body.Close()`；调用方应为网络操作设置 context 期限。
- 预签名并不校验对象是否存在，也不刷新临时凭据。不要记录包含认证参数的完整 URL。
- 写入和复制通常覆盖已有目标；删除和批量管理的业务权限应在上层检查。
- Client 没有独立 Close 方法；不要为每个文件反复创建新 HTTP 连接池。

源码入口：[实现文件](baidu.go)。返回类型与错误分类：[types](../types/README.md)。整体能力与使用入口：[oss](../README.md)。
