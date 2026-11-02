# 对象存储工具

`oss` 位于 `server/app/util/oss`，按系统工具包习惯使用 `Manager`、`NewManager`、`Interface`、中文注释和按职责拆分的 Go 文件。包不依赖数据库、service 或 HTTP controller，备份服务和文件管理服务可以直接依赖 `oss.Interface`。

## 包结构

```text
oss/
├── oss.go                 # Interface、Backend、厂商注册及实例化入口
├── config.go              # Config、ConfigSource、配置转换与校验
├── secret.go              # Secret、SecretSource、凭据转换与校验
├── entities.go            # 请求、结果和传输进度类型
├── errors.go              # 统一错误及错误分类
├── bucket.go              # 桶管理的通用处理
├── file.go                # 文件管理的通用处理
├── multipart.go           # 分片上传与复制的通用处理
├── transfer.go            # 文件和数据流传输
├── progress.go            # 进度计数、重试去重及回调节流
├── aliyun/                # 阿里 OSS 适配
├── tencent/               # 腾讯 COS 适配
├── huawei/                # 华为 OBS 适配
├── volcengine/            # 火山 TOS 适配
└── baidu/                 # 百度 BOS 适配
```

各厂商子包在 `config.go` 中分别定义 `Secret` 和 `Config`：Secret 使用本厂商的凭据名称，Config 描述该厂商的存储连接。业务将两者传给 `oss.NewManager(secret, config)`，获得统一文件管理 API；也可调用子包的 `New(secret, config)` 创建底层客户端。根包负责公共校验、自动分片、文件操作组合和传输进度。厂商子包转换 SDK 请求和结果，并通过 `init` 注册构造函数。根包不导入厂商，实现层单向依赖功能层，避免循环导入。

## 厂商与配置

已知厂商时推荐使用对应子包的 Secret 和 Config，不必填写 Provider，IDE 会提示各自需要的字段：

| 厂商结构 | Secret 的密钥字段 | Secret 的临时令牌 | Config.Region 示例 | Region 规则 |
| --- | --- | --- | --- | --- |
| [tencent.Secret / Config](tencent/README.md) | `SecretID`、`SecretKey` | `SessionToken` | `ap-guangzhou` | 与 Endpoint 至少填写一个 |
| [aliyun.Secret / Config](aliyun/README.md) | `AccessKeyID`、`AccessKeySecret` | `SecurityToken` | `cn-hangzhou` | 始终必填，用于签名 |
| [huawei.Secret / Config](huawei/README.md) | `AccessKeyID`、`SecretAccessKey` | `SecurityToken` | `cn-north-4` | 与 Endpoint 至少填写一个；创建桶时用于指定目标区域 |
| [volcengine.Secret / Config](volcengine/README.md) | `AccessKeyID`、`AccessKeySecret` | `SecurityToken` | `cn-beijing` | 始终必填，用于签名 |
| [baidu.Secret / Config](baidu/README.md) | `AccessKeyID`、`SecretAccessKey` | `SecurityToken` | `bj` | 与 Endpoint 至少填写一个 |

Secret 只保存凭据和临时令牌，支持本厂商字段名称的 JSON 编解码。Config 只保存 `Region`、`Endpoint`、`Bucket`、`HTTPClient`，序列化后不会包含密钥；HTTPClient 不参与 JSON 编解码。现有五家实现都使用这些连接字段，但必填条件和默认地址不同，由厂商构造函数处理。

使用原始密钥，不能传脱敏值。临时凭据还需要在 Secret 中提供对应令牌；长期密钥时令牌留空。每个 Manager 绑定一个桶并支持并发调用。腾讯桶名必须包含 APPID，例如 `backup-1250000000`；仅列出账户的桶时可以不指定 Bucket。

Endpoint 留空时按 Region 生成公网地址，也可指定区域内网地址；地址不带桶名、路径或查询参数。即使配置了 Endpoint，阿里和火山仍需要 Region。特殊地域或网络地址需与实际桶所在地域匹配，并支持所需 API。

### 使用厂商 Secret 和 Config

下面的完整函数创建统一 Manager，后续文件管理和传输操作都面向 `oss.Interface`：

```go
package example

import (
    "server/app/util/oss"
    "server/app/util/oss/aliyun"
)

func OpenBackupStore(accessKeyID, accessKeySecret string) (oss.Interface, error) {
    return oss.NewManager(aliyun.Secret{
        AccessKeyID: accessKeyID,
        AccessKeySecret: accessKeySecret,
    }, aliyun.Config{
        Region: "cn-hangzhou",
        Bucket: "panel-backups",
    })
}
```

腾讯云保留自身的密钥字段名称，临时令牌也放在 Secret 中：

```go
package example

import (
    "server/app/util/oss"
    "server/app/util/oss/tencent"
)

func OpenTencentStore(secretID, secretKey, sessionToken string) (oss.Interface, error) {
    return oss.NewManager(tencent.Secret{
        SecretID: secretID,
        SecretKey: secretKey,
        SessionToken: sessionToken,
    }, tencent.Config{
        Region: "ap-guangzhou",
        Bucket: "backup-1250000000",
    })
}
```

### 使用根包统一类型

`oss.Secret` 和 `oss.Config` 支持动态选择厂商。Config.Provider 使用 `server/app/enum/secret_provider` 枚举；独立使用根包时需要导入对应实现：

```go
package example

import (
    "server/app/enum/secret_provider"
    "server/app/util/oss"
    _ "server/app/util/oss/aliyun"
)

func OpenCommonStore(accessKeyID, accessKeySecret string) (oss.Interface, error) {
    return oss.NewManager(oss.Secret{
        AccessKeyID: accessKeyID,
        AccessKeySecret: accessKeySecret,
    }, oss.Config{
        Provider: secret_provider.Aliyun,
        Region: "cn-hangzhou",
        Bucket: "panel-backups",
    })
}
```

统一 Secret 的三个 JSON 字段为 `access_key_id`、`access_key_secret`、`security_token`。例如腾讯 SecretID 对应统一 AccessKeyID，SecretKey 对应 AccessKeySecret，SessionToken 对应 SecurityToken。使用厂商 Secret 时，其 `Secret()` 方法负责转换，业务无需手工映射。

### 直接调用厂商子包

同一组密钥与配置可以创建底层客户端，仍然执行配置校验、默认地址及 HTTP 传输设置：

```go
package example

import "server/app/util/oss/aliyun"

func OpenClient(accessKeyID, accessKeySecret string) (*aliyun.Client, error) {
    secret := aliyun.Secret{AccessKeyID: accessKeyID, AccessKeySecret: accessKeySecret}
    config := aliyun.Config{Region: "cn-hangzhou", Bucket: "panel-backups"}
    return aliyun.New(&secret, &config)
}
```

子包入口也接受根包的 Secret、Config；使用统一 Config 时，Provider 必须与子包一致。底层 Client 提供文件、桶和显式分片 API，自动 Upload、UploadFile、DownloadFile、MoveFile、RenameFile 等组合能力由 Manager 提供。

### 实例化方式对照

| 调用方式 | 返回值 | 适用场景 |
| --- | --- | --- |
| `oss.NewManager(aliyun.Secret{...}, aliyun.Config{...})` | `*oss.Manager` | 使用厂商字段和完整统一 API |
| `oss.NewManager(oss.Secret{...}, oss.Config{...})` | `*oss.Manager` | 使用统一字段动态选择厂商 |
| `aliyun.New(aliyun.Secret{...}, aliyun.Config{...})` | `*aliyun.Client` | 直接使用厂商底层操作 |
| `aliyun.New(oss.Secret{...}, oss.Config{...})` | `*aliyun.Client` | 使用统一字段，Provider 与子包一致 |
| `oss.NewFromSecret(provider, data, config)` | `*oss.Manager` | 从原始密钥 JSON 读取凭据，config 可为公共或厂商类型 |

五家入口都支持结构体值及非 nil 指针。空密钥或配置、nil 指针、配置厂商不匹配及缺少必填字段等错误，可以用 `errors.Is(err, oss.ErrInvalidArgument)` 判断。构造过程不发网络请求，实际凭据和访问权限通过具体操作验证。

SecretSource 提供 `Secret() oss.Secret`；ConfigSource 提供 `Config() oss.Config` 和 `WithHTTPClient(*http.Client) oss.ConfigSource`。WithHTTPClient 返回使用指定 HTTP 客户端的配置副本，保留其他厂商参数。根入口把完整的厂商配置传给构造器，并复用已经准备好的 HTTP 连接池。业务通常只需传入结构体，无需直接调用这些方法。

### 对接密钥管理

`NewFromSecret` 接收厂商枚举、原始密钥 JSON、连接配置。密钥 JSON 的字段与对应厂商 Secret 的 JSON 标签一致：

| 厂商枚举 | 必填密钥 JSON 字段 | 临时令牌 JSON 字段 |
| --- | --- | --- |
| `secret_provider.TencentCloud` | `secret_id`、`secret_key` | `session_token` |
| `secret_provider.Aliyun` | `access_key_id`、`access_key_secret` | `security_token` |
| `secret_provider.HuaweiCloud` | `access_key_id`、`secret_access_key` | `security_token` |
| `secret_provider.Volcengine` | `access_key_id`、`access_key_secret` | `security_token` |
| `secret_provider.BaiduCloud` | `access_key_id`、`secret_access_key` | `security_token` |

从仓储读取时，前两个参数使用原始记录的 Provider、Data。不能使用 `service.Secret.FindById` 返回的脱敏字段。长期密钥可以省略令牌字段；临时令牌从密钥 JSON 读取，不放在 Config 中。工具包不自动刷新临时凭据，过期后使用新凭据创建 Manager。

下面的完整函数展示如何用厂商 Secret 生成密钥 JSON，并使用同一份连接配置初始化：

```go
package example

import (
    "encoding/json"
    "server/app/enum/secret_provider"
    "server/app/util/oss"
    "server/app/util/oss/aliyun"
)

func OpenEncodedSecret(secret aliyun.Secret, config aliyun.Config) (*oss.Manager, error) {
    data, err := json.Marshal(secret)
    if err != nil {
        return nil, err
    }
    return oss.NewFromSecret(secret_provider.Aliyun, string(data), config)
}
```

传入厂商 Config 时，记录的 Provider 必须与其一致；传入统一 Config 时，Provider 由记录指定。配置与密钥分开解析，详情接口只返回当前厂商的连接字段，不返回 Secret。

动态业务可以保留调用方提供的厂商 ConfigSource，不必逐个取出 Region、Endpoint、Bucket 再拼成另一份配置。下面的独立业务加载所有已支持厂商；只接入一家时可仅保留对应导入。面板的 `service/oss` 已集中加载这些实现。

```go
package example

import (
    "server/app/util/oss"
    _ "server/app/util/oss/aliyun"
    _ "server/app/util/oss/baidu"
    _ "server/app/util/oss/huawei"
    _ "server/app/util/oss/tencent"
    _ "server/app/util/oss/volcengine"
)

func OpenFromSecret(provider int, rawSecret string, config oss.ConfigSource) (oss.Interface, error) {
    return oss.NewFromSecret(provider, rawSecret, config)
}
```

未加载厂商实现时，构造返回“厂商实现未注册”的参数错误。扩展适配器时实现 `oss.Backend`，并在子包的 `init` 中调用 `oss.Register(provider, factory)`，工厂接收 `SecretSource` 和 `ConfigSource`，返回 Backend。注册不创建客户端、不发送网络请求；空工厂或重复注册会立即 panic。新增厂商字段时，在该厂商 Config 中声明并在其构造器内处理，不要在 service 中固定挑选字段。

## 统一 API

| 能力 | 方法 |
| --- | --- |
| 桶管理 | `ListBuckets`、`HeadBucket`、`CreateBucket`、`DeleteBucket` |
| 分页与信息 | `ListFiles`、`StatFile`、`Exists` |
| 上传 | `PutFile`、`Upload`、`UploadFile` |
| 下载 | `GetFile`、`DownloadFile` |
| 文件操作 | `CopyFile`、`MoveFile`、`RenameFile`、`SetMetadata`、`DeleteFile`、`DeleteFiles` |
| 虚拟目录 | `CreateDirectory`，配合 `ListFiles` 的 `Prefix`、`Delimiter` |
| 分片会话 | `InitiateMultipartUpload`、`UploadPart`、`UploadPartCopy`、`ListParts`、`CompleteMultipartUpload`、`AbortMultipartUpload`、`ListMultipartUploads` |
| 临时访问 | `Sign`，支持 `GET`、`HEAD`、`PUT` |

工具包提供 Go API；面板的配置持久化和 HTTP 接口由上层 [OSS 服务](../../service/oss/oss.go) 和 [HTTP controller](../../http/controller/oss) 提供。桶策略、生命周期、版本管理、跨区域复制配置等云平台治理功能不属于当前文件管理接口。

面板添加 OSS 时，`/oss/create` 只提交 `name` 和 `secret_id`，返回新记录的 `id`。创建成功后调用 `/oss/info`，根据返回的 `name` 和当前厂商的 `config` 字段生成编辑表单，再通过 `/oss/update` 保存连接配置。`config` 在详情中是字段对象，更新时提交 JSON 字符串；关联密钥创建后不能修改。初始配置字段为空，填写完整后才能保存连接配置、测试连接和操作云端文件。

备份等业务选择存储时，通过 `/system/enum?name=oss` 获取 `data.name` 中的存储 ID 与名称映射，例如 `{"name":{"1":"网站备份"}}`。名称映射缓存 10 分钟，新增、修改或删除 OSS 配置成功后立即清理缓存；只返回名称，不包含存储配置或密钥内容。

### 面板上传、下载与签名

面板将上传操作统一到 `/oss/upload`，下载操作统一到 `/oss/download`，临时访问签名单独使用 `/oss/sign`。这三个接口均需面板登录凭据，`id` 表示面板中保存的 OSS 配置 ID，`key` 表示存储桶内的文件名。

| 接口 | URL 查询参数 action | 请求正文与用途 |
| --- | --- | --- |
| `/oss/upload` | `file`，省略时默认使用 | 原始文件二进制；`id`、`key` 放在 URL 查询参数，后端自动选择普通或分片上传 |
| `/oss/upload` | `part` | 原始分片二进制；URL 还需 `id`、`key`、`upload_id`、`number`，返回分片编号、ETag 和长度 |
| `/oss/upload` | `server` | JSON：`id`、`key`、`path`、可选 `options`，将服务器文件上传加入后台任务 |
| `/oss/upload` | `init` / `complete` / `abort` / `parts` / `uploads` | JSON：初始化、合并、取消、列出已上传分片或未完成会话 |
| `/oss/download` | `file`，省略时默认使用 | 无文件正文；`id`、`key` 放在 URL，返回下载流，支持 `Range` 请求头 |
| `/oss/download` | `server` | JSON：`id`、`key`、`path`、`overwrite`，将 OSS 文件下载到服务器的任务加入后台队列 |
| `/oss/sign` | 不使用 action | JSON：`id`、`key`、`method`、`expires`、`download`，返回临时 URL、必要请求头和过期时间 |

上传、下载的 `action` 只从 URL 查询参数读取，不能仅放在 JSON 正文中。这样上传 JSON 文件时，正文仍按文件二进制处理。`file`、`part` 上传必须提供准确的 `Content-Length`。`server` 操作立即返回系统任务 ID，进度和取消通过系统任务接口处理。

所有上传都直接覆盖同名文件，无需传入覆盖参数；分片上传在成功合并后替换同名文件。从服务器上传时，支持跨目录、跨页多选文件，每个文件独立创建后台任务，页面使用当前云端目录和服务器文件名作为目标 `key`；同一批选择中的同名文件需分批上传。若任务提交失败，会停止后续提交并显示已提交数量，已创建的任务继续执行。下载到服务器页面仅填写保存目录和文件名，使用选中文件的原始 `key`，固定提交 `overwrite: true` 覆盖服务器同名文件。直接调用下载接口时仍可通过 `overwrite` 控制覆盖，省略时默认不覆盖。

服务器上传、下载全部提交成功后，页面关闭填写路径的表单并打开系统任务弹窗，统一查看进度、日志和取消任务。上传完成后，点击文件列表的刷新按钮查看新文件；文件页面不再跟踪后台任务状态。批量上传部分提交成功时仍保留失败说明，并禁止再次提交整批文件，避免重复上传。

例如，从服务器上传备份文件：

```http
POST /oss/upload?action=server
Content-Type: application/json

{"id":1,"key":"backup/site.tar.gz","path":"/data/backup/site.tar.gz"}
```

获取浏览器附件下载链接：

```http
POST /oss/sign
Content-Type: application/json

{"id":1,"key":"backup/site.tar.gz","method":"GET","expires":86400,"download":true}
```

以上示例省略面板鉴权头。浏览器可直接访问返回的签名 URL；需要面板代理下载时使用 `/oss/download?id=1&key=backup/site.tar.gz`。显式分片上传依次调用 `action=init` 获取 `upload_id`、`action=part` 上传各片，最后用 `action=complete` 提交 `parts` 数组；失败或取消时用 `action=abort` 清理会话。`parts`、`uploads` 两个 action 支持按各自游标查询，适合恢复未完成上传。

页面下载和复制下载链接的有效期均为 24 小时，`/oss/sign` 省略 `expires` 时也默认 86400 秒。下载到服务器通过云密钥调用 SDK，不使用这条临时链接，也不额外设置下载总时限；调用方的 context 和自定义 HTTPClient 超时仍然生效。

## 公共数据结构

所有公共类型都直接定义在 `oss` 根包。业务不需要导入厂商 SDK 的请求或结果类型；厂商子包实现相同的 `oss.Backend`。`oss.Interface` 在 Backend 能力上增加自动上传、本地下载、批量删除等业务常用操作。

| 类型 | 关键字段与约定 |
| --- | --- |
| `Secret` | `AccessKeyID`、`AccessKeySecret`、`SecurityToken`；仅包含统一凭据 |
| `SecretSource` | `Secret() Secret`，转换厂商凭据字段 |
| `Config` | `Provider`、`Region`、`Endpoint`、`Bucket`、`HTTPClient`；只包含连接配置 |
| `ConfigSource` | `Config()` 提供公共连接信息，`WithHTTPClient` 返回保留厂商字段的配置副本 |
| `Bucket` | `Name`、`Region`、`CreatedAt`；不同云服务可能不返回全部信息 |
| `File` | `Key`、`Size`、`ETag`、`LastModified`、内容属性、`StorageClass` 和 `Metadata` |
| `ListOptions` | `Prefix`、`Delimiter`、`StartAfter`、`Token`、`Limit`；Manager 中 Limit 为 0 时使用 1000 |
| `FilePage` | `Files`、`CommonPrefixes`、`HasMore`、`NextToken`，没有可靠的总条数或任意跳页 |
| `PutOptions` | `ContentType`、`CacheControl`、`ContentDisposition`、`ContentEncoding`、`StorageClass`、`Metadata` |
| `GetOptions` | `Range`、`IfMatch`、`Progress`；Progress 由 Manager 处理，按实际读取流统计 |
| `Download` | `File`、`Body`、`ContentRange`；调用方负责关闭 Body |
| `UploadOptions` | 嵌入 `PutOptions`，另有 `PartSize`、`TempDir`、`Progress` |
| `DownloadOptions` | `Progress`；用于完整文件下载，不提供 Range，避免将局部内容覆盖为完整文件 |
| `CopyOptions` | `ReplaceMetadata`、`PutOptions`；在当前桶内复制，默认保留源元数据 |
| `CopyPartOptions` | `Offset`、`Size`、`IfMatch`；复制当前桶内源文件的指定范围，Offset 从 0 开始，范围终点为 Offset + Size - 1 |
| `Part` | `Number`、`ETag`、`Size`、`LastModified`；合并时编号、ETag、长度必需，编号从 1 开始 |
| `MultipartUpload` | `Key`、`UploadID`、`Initiated`；持久化续传还需记录本地文件身份和分片策略 |
| `PartsOptions` / `PartsPage` | 通过 `Marker` / `NextMarker` 翻页；`Limit` 设定页大小 |
| `UploadsOptions` / `UploadsPage` | 按 Prefix 过滤，下一页同时传回对象名和上传 ID 游标 |
| `SignedURL` | `URL`、`Headers`、`ExpiresAt`；调用方须携带签名所要求的请求头 |
| `DeleteResult` | 单个对象的 `Key` 和 `Err`；批量操作可能部分成功 |

写入或复制返回的 `File` 只包含本次请求及响应已知的信息，完整元数据使用 `StatFile` 读取。Size 的单位为字节；范围下载时 `Download.File.Size` 是本次响应体的长度，完整文件范围可从 ContentRange 读取。ETag 按云厂商原样保存，不应总是将其视为文件 MD5。

`ReplaceMetadata=true` 会整体替换元数据，未传字段不会自动保留。StorageClass 使用对应厂商的原生取值。`CopyPartOptions{Offset: 0, Size: 1}` 表示第一个字节，不能把它理解为未设置范围。

根包与各厂商的 Secret 都支持 JSON 编解码，Config 不包含任何凭据，HTTPClient 不参与 JSON 编解码。`ResolveSecret`、`ResolveConfig` 分别处理密钥和连接配置的转换与空值，`NormalizeConfig` 接收两者进行公共校验并准备 HTTP 客户端；默认区域地址和厂商签名要求由各厂商构造函数处理。

## 上传备份

### 小文件：保存 JSON 备份清单

`PutFile` 需要准确的字节长度。下面的函数适合小型清单、配置等数据，先编码到内存，再上传；大型备份使用后面的流式或文件接口，避免将全部内容读入内存。

```go
package example

import (
	"bytes"
	"context"
	"encoding/json"
	"server/app/util/oss"
)

func SaveManifest(ctx context.Context, store oss.Interface, key string, manifest any) (*oss.File, error) {
	data, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	return store.PutFile(ctx, key, bytes.NewReader(data), int64(len(data)), oss.PutOptions{
		ContentType: "application/json",
	})
}
```

### 大文件与未知长度数据流

小文件已知长度时使用 `PutFile`；本地文件推荐 `UploadFile`，它自动选择简单上传或分片上传：

```go
file, err := manager.UploadFile(ctx, "database/2026-10-28.sql.gz", filename, oss.UploadOptions{
    PutOptions: oss.PutOptions{
        ContentType: "application/gzip",
        Metadata: map[string]string{"backup-type": "database"},
    },
    Progress: func(uploaded, total int64) {
        // 更新任务进度；回调在当前上传调用内串行执行。
    },
})
```

流式备份或压缩管道可以直接传入 reader，长度未知时使用 `-1`：

```go
file, err := manager.Upload(ctx, "database/latest.sql.gz", backupReader, -1, oss.UploadOptions{
    PutOptions: oss.PutOptions{ContentType: "application/gzip"},
})
```

默认分片为 16 MiB，按顺序上传，在单次上传内复用同一块分片缓冲，避免逐片分配大块内存。分片超过 64 MiB 时转存临时文件，`TempDir` 可指定备份任务的临时目录；临时文件在调用结束前清理。已知长度的文件会自动扩大分片，避免超过 10000 片。统一限制为：对象最多 5 TiB，单片最多 5 GiB，非末片至少 5 MiB。未知长度时不能提前调整分片，默认分片最多覆盖约 156 GiB；更大的流应预先增大 `PartSize`，达到上限会返回错误并中止上传，不会发布截断对象。

自动上传失败或取消后，会用独立的 30 秒 context 中止分片会话；中止也失败时返回聚合错误，包含可用于后续清理的上传 ID。调用方应设置任务期限。任意 `io.Reader` 无法被工具包强制打断；对于阻塞的 `io.Pipe`，备份任务必须在取消时关闭管道或停止生产者。

上传进度统计 SDK 已消费的数据，分片全部发送后还需服务端确认并成功合并。`PutFile` 和 `Upload` 的已知长度表示要上传的字节数，额外尾随内容不属于本次上传；备份生产者发生错误时应通过 reader 返回该错误，不能把失败包装成正常 EOF。

## 传输进度回调

上传和下载共用 `oss.ProgressFunc`，参数是累计传输字节数和总字节数：

```go
type ProgressFunc func(transferred, total int64)
```

| 入口 | 回调位置 | 统计内容 | 最终通知时机 |
| --- | --- | --- | --- |
| `Upload` / `UploadFile` | `UploadOptions.Progress` | SDK 从当前对象或分片读取的累计字节 | 简单上传成功，或全部分片成功合并并完成清理 |
| `GetFile` | `GetOptions.Progress` | 调用方从返回 Body 读出的字节 | 读到 EOF 且长度校验通过 |
| `DownloadFile` | `DownloadOptions.Progress` | 成功写入临时文件的字节 | 长度校验、关闭、同步和替换目标文件成功 |

回调在同一操作内串行执行，传输过程中约每 100 ms 最多通知一次，开始和完成阶段会单独通知。不要假定每个读写块或每个分片都对应一次回调。上传按当前分片读取位置的最大值计数；SDK 为重试回退读取不会让进度倒退，也不会重复累加该分片。将数据读入本地分片缓冲的过程不计为已经上传。

- 总长度已知且大于 0 时，成功前的进度最多报告到 `total - 1`；最后一次 `(total, total)` 只在该入口成功完成后发出。
- 总长度未知时，中途的 total 为 `-1`；成功后最后一次回调为 `(实际字节数, 实际字节数)`。
- 空文件成功后回调 `(0, 0)`，业务不能用除法直接计算百分比。
- 传输失败、context 取消、流提前关闭不会发送完成通知。上传分片失败仍会尝试清理未完成会话。
- 回调应快速返回；更新 UI、网络推送或写数据库时，由业务自行将状态派发到自己的队列或界面线程。不要在回调中阻塞等待同一个传输结束。

中间的上传字节数表示 SDK 消费进度，不表示这些字节已经成为云端可用对象。`Upload`、`UploadFile`、`DownloadFile` 的成功以返回 error 为 nil 为准。`GetFile` 返回成功只表示已取得下载流；调用方还要检查读取和 Close 的错误。流式回调只代表 Body 读取过程，不代表后续 writer 的落盘、数据库导入或业务处理完成。

### 上传文件并报告进度

下面的函数保留统一接口，业务只需要提供进度处理函数。它支持小文件直接上传和大文件自动分片；文件在上传期间应保持不变。

```go
package example

import (
    "context"
    "server/app/util/oss"
)

func UploadBackup(ctx context.Context, store oss.Interface, key, filename string, report oss.ProgressFunc) (*oss.File, error) {
    return store.UploadFile(ctx, key, filename, oss.UploadOptions{
        PutOptions: oss.PutOptions{ContentType: "application/gzip"},
        Progress: report,
    })
}
```

未知长度备份流可以使用同样的回调：

```go
package example

import (
    "context"
    "io"
    "server/app/util/oss"
)

func UploadBackupStream(ctx context.Context, store oss.Interface, key string, reader io.Reader, report oss.ProgressFunc) (*oss.File, error) {
    return store.Upload(ctx, key, reader, -1, oss.UploadOptions{
        PutOptions: oss.PutOptions{ContentType: "application/gzip"},
        Progress: report,
    })
}
```

### 下载文件并报告进度

```go
package example

import (
    "context"
    "server/app/util/oss"
)

func DownloadBackup(ctx context.Context, store oss.Interface, key, filename string, report oss.ProgressFunc) error {
    return store.DownloadFile(ctx, key, filename, oss.DownloadOptions{
        Progress: report,
    })
}
```

不需要进度时仍可直接调用 `store.DownloadFile(ctx, key, filename)`。下载先写入同目录临时文件，失败时保留旧目标并清理临时文件；最终进度不会早于目标文件替换完成。

网络中断或临时服务错误会自动重试，单次调用最多额外重试 3 次，分别等待 1、2、4 秒，等待时也可取消。已写入的数据通过 `Range` 从断点继续下载，并携带首次响应的 `ETag` 作为 `IfMatch` 条件；续传响应必须满足原文件范围、总长度和 ETag 校验。进度按累计落盘字节计算，重试不会重复计数。权限错误、文件不存在、源文件变化及本地写入失败不会重试。

续传只覆盖当前调用中的临时网络故障，不保留跨进程或跨任务的下载断点。远端未提供完整长度、未提供 ETag 或只提供弱 ETag 时，仍可正常连续下载；若已经写入部分内容后中断，会报错并清理临时文件，避免拼接无法确认版本的数据。

### 流式读取并报告进度

```go
package example

import (
    "context"
    "errors"
    "io"
    "server/app/util/oss"
)

func StreamBackup(ctx context.Context, store oss.Interface, key string, destination io.Writer, report oss.ProgressFunc) (written int64, err error) {
    download, err := store.GetFile(ctx, key, oss.GetOptions{Progress: report})
    if err != nil {
        return 0, err
    }
    defer func() { err = errors.Join(err, download.Body.Close()) }()
    return io.Copy(destination, download.Body)
}
```

带 Range 时，total 为当前范围响应的长度，transferred 从 0 开始；断点续传业务需要自行加上之前已完成的偏移。应读到 EOF，让流包装完成长度校验；提前停止读取时务必 Close，但不会发送完成进度。

### 将字节数转换为显示文案

下面的回调示例处理未知长度和空文件，业务可以把生成的状态交给自己的界面。展示百分比与任务成功状态分开处理，错误状态来自传输方法的返回值。

```go
package example

import "fmt"

func ProgressText(transferred, total int64) string {
    if total < 0 {
        return fmt.Sprintf("已传输 %d 字节", transferred)
    }
    if total == 0 {
        return "空文件传输完成"
    }
    percent := int(100 * float64(transferred) / float64(total))
    return fmt.Sprintf("%d%%（%d / %d 字节）", percent, transferred, total)
}
```

## 分页浏览

云对象列表采用游标分页，没有可靠的总条数或随机页码。将 `NextToken` 原样传入下一次请求，并保持相同的前缀和分隔符。

列举目录内容时，可同时设置 `Prefix: "database/"`、`StartAfter: "database/"` 和 `Delimiter: "/"`，由云端跳过目录自身的占位对象，不占用本页条数。阿里云、火山引擎使用原生 `StartAfter`，腾讯云、华为云、百度云使用原生 `Marker`；`Token` 非空时优先按续页游标查询，忽略 `StartAfter`。留空 `StartAfter` 会保留原始列举行为，递归删除等需要处理目录占位对象的操作无需设置。

每次 `ListFiles` 只读取云端一页，不为凑满条数额外请求；云端实际返回数量可以小于 `Limit`，是否继续翻页以 `HasMore` 为准。

下面的示例按游标读取目录内容：

```go
options := oss.ListOptions{Prefix: "database/", Delimiter: "/", StartAfter: "database/", Limit: 100}
for {
    page, err := manager.ListFiles(ctx, options)
    if err != nil {
        return err
    }
    // page.Files 是文件；page.CommonPrefixes 是子目录。
    if !page.HasMore {
        break
    }
    options.Token = page.NextToken
}
```

不设置 `Delimiter` 时递归列出该前缀下的对象。对象 key 保留原始斜杠、空格和点号，不进行本地文件路径归一化。目录是以 `/` 结尾的空对象；删除目录标记不会递归删除其内容。

## 续传和分片管理

调用方持久化 `Key`、`UploadID`、文件身份与长度、分片大小后，可以跨进程恢复：

1. `InitiateMultipartUpload` 初始化会话，同时指定最终对象元数据。
2. `UploadPart` 上传分片，保存返回的 `Part`；同一编号重传会覆盖旧片。
3. 恢复时循环 `ListParts`，用 `NextMarker` 查询下一页，确认源文件未变化后上传缺片。
4. 将全部分片的编号、ETag、长度传给 `CompleteMultipartUpload`。已排序列表直接校验，乱序列表才复制排序，不修改调用方切片。
5. 不再使用的会话调用 `AbortMultipartUpload`。`ListMultipartUploads` 同时返回下一页的 `NextKeyMarker` 和 `NextUploadIDMarker`，两者必须一起回传。

`Upload` 不自动保存断点，会清理失败的会话；需要持久化续传的任务应直接组合上述分片 API。最终合并必须提供从 1 开始的连续分片编号，且最后一片之外的所有分片至少 5 MiB。

## 下载、复制与临时链接

`GetFile` 返回流式 `Body`，调用方负责关闭。`GetOptions.Range` 接受单个 HTTP 范围，例如 `bytes=0-1023`、`bytes=1024-`；通过 `IfMatch` 固定源 ETag，适合断点下载。

下面的完整函数将对象复制到调用方提供的 writer，并处理下载流的关闭错误；writer 的生命周期仍由调用方负责。

```go
package example

import (
	"context"
	"errors"
	"io"
	"server/app/util/oss"
)

func ReadBackup(ctx context.Context, store oss.Interface, key string, destination io.Writer) (written int64, err error) {
	download, err := store.GetFile(ctx, key, oss.GetOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { err = errors.Join(err, download.Body.Close()) }()
	return io.Copy(destination, download.Body)
}
```

`DownloadFile` 先写入目标同目录的临时文件，中断时自动重试续传，校验读取长度并同步成功后替换目标；失败时保留原目标文件。该方法下载完整对象，只获取指定范围请用 `GetFile`。

`CopyFile` 在 Manager 绑定的当前桶内复制，source 和 destination 均为该桶内的文件 key。复制前会读取源元数据，超过 1 GiB 自动使用云端分片复制，每一片携带源 ETag 条件，避免复制过程中源变化产生混合文件。分片复制失败会中止会话。`UploadPartCopy` 同样在当前桶内读取源文件，可直接用于调用方管理的分片复制任务。

`MoveFile` 接收完整目标 key，通过复制加删除移动文件；复制成功后才删除源。它不是原子操作，删除失败时返回目标信息和错误，两份对象都会保留。执行期间调用方应保证源不被并发修改。所有写入遵循云存储覆盖语义；目标已存在时可能被覆盖。`SetMetadata` 整体替换元数据和内容头，应传入需要保留的字段；复制不会保留源对象 ACL 等独立访问策略。

`RenameFile(ctx, key, name)` 只接收单个新文件名，保留源 key 中的原始目录前缀（包括连续斜杠、空格等），不使用本地路径规范化。它会拒绝空名称、`.`、`..`、目录分隔符、目录标记和原名称，并校验最终 key 的 UTF-8 及 1024 字节限制。目标文件已存在时返回 `ErrConflict`；读取目标失败时直接返回错误，不继续复制。底层复用 `MoveFile`，超过 1 GiB 的文件自动分片复制。检查与复制不是原子操作，执行期间需保证源和目标不被并发修改；删除源失败时会返回目标信息和错误，两份文件都保留。

```go
// 将 backup/site.tar.gz 改名为 backup/site-20261102.tar.gz。
file, err := manager.RenameFile(ctx, "backup/site.tar.gz", "site-20261102.tar.gz")
// 面板业务可使用注入的 OSS service，storageID 是已保存的存储配置 ID。
file, err = ossService.Rename(ctx, storageID, "backup/site.tar.gz", "site-20261102.tar.gz")
```

面板文件操作接口通过 JSON 正文中的 `action=rename` 调用重命名，`name` 不包含目录；移动仍使用 `action=move` 和完整的 `destination`：

```http
POST /oss/file
Content-Type: application/json

{"id":1,"action":"rename","key":"backup/site.tar.gz","name":"site-20261102.tar.gz"}
```

成功时 `data` 返回重命名后的文件信息，例如 `key` 为 `backup/site-20261102.tar.gz`。文件列表的操作菜单和右键菜单都提供独立的“重命名”入口。

`DeleteFiles` 返回每个 key 的结果和聚合错误，方便区分部分成功。`DeleteBucket` 只删除空桶。归档对象下载/复制、存储类型和访问权限等仍受厂商规则约束；`StorageClass` 使用相应厂商的原生值。

```go
signed, err := manager.Sign(ctx, "database/latest.sql.gz", "GET", 15*time.Minute)
// 客户端访问 signed.URL，并携带 signed.Headers 中的头。
```

预签名有效期为 1 秒到 7 天，同时受临时凭据剩余有效期限制。签名地址具备临时访问权限，不应写入公开日志。

浏览器下载文件时，可以传入一份 `SignOptions`，通过附件响应头和文件名触发下载，避免图片、文本等直接在页面内预览：

```go
package example

import (
    "context"
    "mime"
    "path"
    "server/app/util/oss"
    "time"
)

func CreateDownloadURL(ctx context.Context, store oss.Interface, key string) (*oss.SignedURL, error) {
    return store.Sign(ctx, key, "GET", 10*time.Minute, oss.SignOptions{
        ContentDisposition: mime.FormatMediaType("attachment", map[string]string{
            "filename": path.Base(key),
        }),
    })
}
```

阿里云 GET 请求不支持 `response-content-type`，因此下载签名不要设置 `ContentType`，否则会返回 `InvalidRequest`（`0017-00000902`）。文件类型应在上传时通过 `PutOptions.ContentType` 设置，详见[阿里云错误说明](https://help.aliyun.com/zh/oss/user-guide/0017-00000902)。

响应覆盖参数仅用于 GET，最多传入一份；普通 GET、HEAD、PUT 签名仍可省略参数。各厂商实现会在签名前转换对应的查询参数，生成 URL 后不要再次拼接或修改参数。`SignedURL.Headers` 只包含请求实际需要携带的签名头；浏览器直接打开链接时会自动发送 Host，其余必需头需由请求客户端设置。

## 错误与验证

| 错误哨兵 | 判定含义 |
| --- | --- |
| `ErrInvalidArgument` | 配置、未加载的实现或操作参数不合法 |
| `ErrNotFound` | HTTP 404 |
| `ErrPermissionDenied` | HTTP 401 或 403 |
| `ErrConflict` | HTTP 409 |
| `ErrPreconditionFailed` | HTTP 412 |

`oss.Error` 保存 `Provider`、`Op`、`Key`、`StatusCode`、`Code`、`Message`、`Err`。`Unwrap()` 保留底层错误链，`Is()` 按响应状态匹配统一错误。`WrapError` 对 nil 返回 nil，保留已有的统一错误，其他错误包装为对应厂商操作错误。有些厂商在 HTTP 200 正文中报告业务错误，应检查返回 error 和 Code。

所有网络操作接受 `context.Context`。使用 `errors.Is` 检查 `ErrInvalidArgument`、`ErrNotFound`、`ErrPermissionDenied`、`ErrConflict`、`ErrPreconditionFailed` 或 context 的取消/超时错误。使用 `errors.As` 提取 `*oss.Error` 获取厂商、操作、状态码和错误码；其 `Error()` 不输出原始 SDK 响应，避免日志泄漏签名内容。`Exists` 仅将 404 视为不存在，不会吞掉权限或网络错误。

可传入 `Config.HTTPClient` 配置代理、Transport 和超时。工具包浅拷贝 client 并禁用自动重定向，避免认证请求被转发到错误的区域或主机；标准 `*http.Transport` 未关闭透明 gzip 解压时会克隆并关闭自动解压，已经关闭时直接复用连接池，确保压缩备份按原始字节下载。自定义 Transport 也应保留响应原始字节，不做自动解压。阿里 `HeadBucket` 通过 `GetBucketInfo` 实现，需要对应读桶信息权限。

使用编译和静态检查验证包及所有厂商子包：

```sh
cd server
go build ./app/util/oss/...
go vet ./app/util/oss/...
```

官方 SDK：[阿里 OSS](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2)、[腾讯 COS](https://github.com/tencentyun/cos-go-sdk-v5)、[华为 OBS](https://github.com/huaweicloud/huaweicloud-sdk-go-obs)、[火山 TOS](https://github.com/volcengine/ve-tos-golang-sdk)、[百度 BOS](https://github.com/baidubce/bce-sdk-go)。

各适配器说明：[阿里](aliyun/README.md)、[腾讯](tencent/README.md)、[华为](huawei/README.md)、[火山](volcengine/README.md)、[百度](baidu/README.md)。公共字段和错误类型见本文说明，其他工具见 [util 索引](../README.md)。
