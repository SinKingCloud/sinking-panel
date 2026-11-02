# huawei：华为云 OBS适配器

导入路径：`server/app/util/oss/huawei`。本子包将官方 `github.com/huaweicloud/huaweicloud-sdk-go-obs/obs` 的请求和结果转换为 [oss](../README.md#公共数据结构) 的公共数据结构。

使用 `huawei.Secret` 提供凭据，使用 `huawei.Config` 提供存储连接参数，再调用 `oss.NewManager(secret, config)` 获得统一文件管理、自动分片和进度回调。仅需要底层厂商操作时调用 `huawei.New(secret, config)`。两个入口都校验配置、补齐默认服务地址并准备 HTTP 客户端，构造时不发送网络请求。

导入本子包时会通过 `init` 调用 `oss.Register` 注册实现，无需手工注册。仅使用根包统一类型的独立业务，可以使用 `_ "server/app/util/oss/huawei"` 加载本实现。

## 密钥与连接配置

两个结构体均定义在 [config.go](config.go)，职责分开：`Secret` 只包含凭据和临时令牌，`Config` 只包含存储连接参数。序列化 Config 不会包含密钥。

### Secret

| 字段 | JSON 字段 | 是否必需 | 说明 |
| --- | --- | --- | --- |
| `AccessKeyID` | `access_key_id` | 必填 | 原始访问密钥 ID，不能使用脱敏值 |
| `SecretAccessKey` | `secret_access_key` | 必填 | 与访问密钥 ID 配对的原始密钥，不能使用脱敏值 |
| `SecurityToken` | `security_token` | 使用临时凭据时必填 | 与临时访问密钥一起传入；长期密钥留空，不自动刷新 |

Secret 支持 JSON 序列化和反序列化，保留本厂商的字段名称。`Secret()` 将凭据转换为统一 `oss.Secret`，临时令牌仍由 Secret 提供，不放入 Config。不要将 Secret 或包含凭据的 JSON 打印到日志或作为配置详情返回。

### Config

| 字段 | JSON 字段 | 是否必需 | 说明与示例 |
| --- | --- | --- | --- |
| `Region` | `region` | 与 Endpoint 至少填写一项；创建桶时用于指定目标区域 | 例如 `cn-north-4` |
| `Endpoint` | `endpoint` | 选填 | 留空按 Region 生成 `https://obs.cn-north-4.myhuaweicloud.com`；可以指定区域内网服务地址 |
| `Bucket` | `bucket` | 按操作决定 | 操作文件时必填；仅列出账户存储桶时可省略 |
| `HTTPClient` | 不参与 JSON | 选填 | 自定义代理、超时和连接池；构造时复制客户端并设置统一传输行为 |

Config 无需填写 Provider，`Config()` 固定返回 `secret_provider.HuaweiCloud` 的公共连接信息。Endpoint 是区域服务地址，不包含桶名、对象路径、查询参数或用户名密码；省略协议时补全 `https://`。表中的桶名为示例，需替换为自己有权限访问的桶。

`WithHTTPClient` 返回使用指定 HTTP 客户端的完整配置副本，保留其他厂商字段，不修改原配置。统一入口将原厂商配置传给对应构造器，并复用已准备的 HTTP 连接池。

## 推荐使用示例

下面的完整函数使用独立的 Secret、Config 创建统一 Manager，并列出一页备份文件：

```go
package example

import (
    "context"
    "server/app/util/oss"
    "server/app/util/oss/huawei"
)

func ListBackups(ctx context.Context, keyID, keySecret, token string) (*oss.FilePage, error) {
    manager, err := oss.NewManager(huawei.Secret{
        AccessKeyID: keyID,
        SecretAccessKey: keySecret,
        SecurityToken: token,
    }, huawei.Config{
        Region: "cn-north-4",
        Bucket: "panel-backups",
    })
    if err != nil {
        return nil, err
    }
    return manager.ListFiles(ctx, oss.ListOptions{
        Prefix: "backups/",
        Delimiter: "/",
        Limit: 100,
    })
}
```

长期服务应复用 Manager。密钥和配置都支持结构体值及非 nil 指针，例如 `oss.NewManager(&secret, &config)`。下一页把结果的 `NextToken` 原样赋给 `ListOptions.Token`，保持 Prefix 和 Delimiter 不变。上传、下载、续传、签名 URL 的完整用法见 [统一 OSS 文档](../README.md)。

## 直接调用子包

构造签名为 `New(secret oss.SecretSource, config oss.ConfigSource) (*Client, error)`。下例使用本厂商的两个结构体创建底层客户端：

```go
package example

import "server/app/util/oss/huawei"

func OpenClient(keyID, keySecret, token string) (*huawei.Client, error) {
    return huawei.New(huawei.Secret{
        AccessKeyID: keyID,
        SecretAccessKey: keySecret,
        SecurityToken: token,
    }, huawei.Config{
        Region: "cn-north-4",
        Bucket: "panel-backups",
    })
}
```

也可使用根包的 `oss.Secret` 和 `oss.Config`。统一密钥使用 `AccessKeyID`、`AccessKeySecret` 和 `SecurityToken`；统一配置必须指定正确的 Provider：

```go
package example

import (
    "server/app/enum/secret_provider"
    "server/app/util/oss"
    "server/app/util/oss/huawei"
)

func OpenClientFromCommon(keyID, keySecret, token string) (*huawei.Client, error) {
    return huawei.New(oss.Secret{
        AccessKeyID: keyID,
        AccessKeySecret: keySecret,
        SecurityToken: token,
    }, oss.Config{
        Provider: secret_provider.HuaweiCloud,
        Region: "cn-north-4",
        Bucket: "panel-backups",
    })
}
```

传入其他厂商的配置、nil 密钥或配置、缺失必需字段或不合法地址时，构造返回可由 `errors.Is(err, oss.ErrInvalidArgument)` 判断的错误。构造成功仅表示本地配置有效，云端凭据、桶是否存在及权限需通过具体操作验证。

直接调用子包也会浅拷贝 HTTPClient 并禁止自动重定向。标准 `*http.Transport` 尚未禁用透明 gzip 解压时，会先克隆再关闭自动解压；已经关闭时直接复用连接池。自定义 Transport 应自行保证不自动解压响应。无需为子包入口手工准备标准化配置。

## 从 JSON 加载密钥

密钥 JSON 使用 `access_key_id`、`secret_access_key`，临时凭据额外包含 `security_token`。`NewFromSecret` 从原始密钥 JSON 中读取这些字段，第三个参数仅传连接配置：

```go
package example

import (
    "server/app/enum/secret_provider"
    "server/app/util/oss"
    "server/app/util/oss/huawei"
)

func OpenFromSecret(rawSecret string) (*oss.Manager, error) {
    return oss.NewFromSecret(secret_provider.HuaweiCloud, rawSecret, huawei.Config{
        Region: "cn-north-4",
        Bucket: "panel-backups",
    })
}
```

provider 参数必须与本厂商匹配，密钥 JSON 必须包含原始凭据，不能使用脱敏值。可以用 `json.Marshal(huawei.Secret{...})` 生成符合字段名称的密钥 JSON，构造器不会从 Config 中读取或保留凭据。

## 底层 API

Client 提供以下底层操作，所有方法的首个参数均为非 nil 的 `context.Context`：

| 分类 | 方法及关键约定 |
| --- | --- |
| 桶 | `ListBuckets`、`HeadBucket`、`CreateBucket`、`DeleteBucket`；绑定桶来自 Config.Bucket |
| 查询 | `ListFiles(ctx, oss.ListOptions)`、`StatFile(ctx, key)`；列表不保证含全部元数据 |
| 读写 | `GetFile(ctx, key, oss.GetOptions)`、`PutFile(ctx, key, reader, size, oss.PutOptions)`；下载 Body 必须关闭 |
| 删除与复制 | `DeleteFile`、`CopyFile`；复制限当前桶内，普通复制仍受厂商大小限制，大对象应走根包自动分片复制 |
| 上传会话 | `InitiateMultipartUpload`、`UploadPart`、`ListParts`、`CompleteMultipartUpload`、`AbortMultipartUpload` |
| 分片复制 | `UploadPartCopy`；读取当前桶内的源对象，Offset 从 0 开始，范围终点为 Offset + Size - 1，IfMatch 固定源对象 |
| 会话浏览 | `ListMultipartUploads`，返回未完成上传及两个后续游标 |
| 临时访问 | `Sign(ctx, key, method, ttl)`，使用返回的 URL 与 Headers |

直接 Client 实现根包的 `oss.Backend`，没有 `Upload`、`UploadFile`、`DownloadFile`、`MoveFile`、`RenameFile`、`Exists` 等组合方法；这些由根包 Manager 提供。上传与下载进度也由 Manager 统一处理，需要业务进度显示时请使用 `oss.NewManager`，不要依赖底层 Client 处理 `GetOptions.Progress`。具体见[进度回调示例](../README.md#传输进度回调)。直接上传要求准确声明长度，调用方负责失败会话清理和分片大小校验。对外业务不应依赖官方 SDK 的结果类型。

## 本适配器的行为与维护要点

### 调用上下文

OBS 将 context 放在客户端配置中；每次操作创建轻量 SDK 客户端并复用 HTTPClient，使并发请求的取消状态互不覆盖。不要缓存后反复修改同一个请求 context。

### 签名与重试

使用原生 SignatureObs，启用 TLS 校验，关闭 SDK 自动重试和自动重定向，避免重放不能回退的输入流或将认证发往其他节点。

### 分页

对象列表使用 Marker；桶列表内部循环读取全部页，分片会话需要同时保存 KeyMarker 与 UploadIDMarker。

### 复制条件

分片复制显式发送源范围和 x-obs-copy-source-if-match。零偏移的单字节范围同样有效。

### 完成与取消

合并响应缺少 ETag 会作为无效响应报错；取消或超时会保留 context 错误供 errors.Is 判断。

## 错误和资源管理

- 使用 `errors.Is(err, oss.ErrNotFound)` 或 `errors.As(err, &detail)` 检查统一错误；不要仅匹配中文错误文本。
- 流式下载必须 `defer result.Body.Close()`；调用方应为网络操作设置 context 期限。
- 预签名并不校验对象是否存在，也不刷新临时凭据。不要记录包含认证参数的完整 URL。
- 写入和复制通常覆盖已有目标；删除和批量管理的业务权限应在上层检查。
- Client 没有独立 Close 方法；不要为每个文件反复创建新 HTTP 连接池。

源码入口：[配置结构](config.go)、[实现文件](huawei.go)。返回类型与错误分类：[公共类型与错误](../README.md#公共数据结构)。整体能力与使用入口：[oss](../README.md)。
