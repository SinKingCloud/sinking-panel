# 对象存储工具

`oss` 位于 `server/app/util/oss`，按系统工具包习惯使用 `Manager`、`NewManager`、`Interface`、中文注释和按职责拆分的 Go 文件。包不依赖数据库、service 或 HTTP controller，备份服务和文件管理服务可以直接依赖 `oss.Interface`。

## 包结构

```text
oss/
├── oss.go                 # 统一 Interface、NewManager 和 NewFromSecret 入口
├── entities.go            # 对外公共类型别名，保持 oss.Config 等调用方式
├── bucket.go              # 桶管理的通用处理
├── object.go              # 文件管理的通用处理
├── multipart.go           # 分片上传与复制的通用处理
├── transfer.go            # 文件和数据流传输
├── types/                 # 厂商共用的数据结构与错误类型
├── aliyun/                # 阿里 OSS 适配
├── tencent/               # 腾讯 COS 适配
├── huawei/                # 华为 OBS 适配
├── volcengine/            # 火山 TOS 适配
└── baidu/                 # 百度 BOS 适配
```

调用方统一从 `oss` 包创建实例。各厂商子包仅负责 SDK 调用与结果转换，依赖独立的 `types` 包，不反向依赖 `oss`，避免循环导入。

## 厂商与配置

复用 `server/app/enum/secret_provider`，不新增另一套厂商编号。

| 厂商 | 枚举 | 密钥管理 JSON 字段 | Region 示例 |
| --- | --- | --- | --- |
| 腾讯 COS | `TencentCloud` | `secret_id`、`secret_key` | `ap-guangzhou` |
| 阿里 OSS | `Aliyun` | `access_key_id`、`access_key_secret` | `cn-hangzhou` |
| 华为 OBS | `HuaweiCloud` | `access_key_id`、`secret_access_key` | `cn-north-4` |
| 火山 TOS | `Volcengine` | `access_key_id`、`access_key_secret` | `cn-beijing` |
| 百度 BOS | `BaiduCloud` | `access_key_id`、`secret_access_key` | `bj` |

每个 Manager 绑定一个桶，支持并发调用。腾讯桶名需要包含 APPID，例如 `backup-1250000000`。只列出账户的桶时可以不指定 `Bucket`。

`Endpoint` 留空时按 `Region` 生成公网地址，也可指定完整区域服务地址用于内网访问；地址不带桶名、路径或查询参数。阿里和火山签名始终需要 `Region`。特殊地域、加速和私有网络应显式传入对应服务地址，并确认该地址支持要调用的 API。

下面是可直接放入业务包的完整初始化函数；更换厂商时使用上表的枚举、区域和对应凭据。构造过程不发网络请求，实际权限在操作时验证。

```go
package example

import (
	"server/app/enum/secret_provider"
	"server/app/util/oss"
)

func OpenBackupStore(accessKeyID, accessKeySecret string) (oss.Interface, error) {
	return oss.NewManager(oss.Config{
		Provider:        secret_provider.Aliyun,
		Region:          "cn-hangzhou",
		Bucket:          "panel-backups",
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
	})
}
```

对接现有密钥管理时，使用仓储读取的原始记录：

```go
manager, err := oss.NewFromSecret(record.Provider, record.Data, oss.Config{
    Region: "cn-hangzhou",
    Bucket: "panel-backups",
})
```

不要使用 `service.Secret.FindById` 返回的脱敏字段。`NewFromSecret` 根据记录的厂商覆盖 Config 中的长期密钥，并拒绝脱敏值；临时令牌通过 `Config.SecurityToken` 传入。该包不自动刷新临时凭据，过期后使用新凭据创建 Manager。

## 统一 API

| 能力 | 方法 |
| --- | --- |
| 桶管理 | `ListBuckets`、`HeadBucket`、`CreateBucket`、`DeleteBucket` |
| 分页与信息 | `ListObjects`、`StatObject`、`Exists` |
| 上传 | `PutObject`、`Upload`、`UploadFile` |
| 下载 | `GetObject`、`DownloadFile` |
| 文件操作 | `CopyObject`、`MoveObject`、`SetMetadata`、`DeleteObject`、`DeleteObjects` |
| 虚拟目录 | `CreateDirectory`，配合 `ListObjects` 的 `Prefix`、`Delimiter` |
| 分片会话 | `InitiateMultipartUpload`、`UploadPart`、`UploadPartCopy`、`ListParts`、`CompleteMultipartUpload`、`AbortMultipartUpload`、`ListMultipartUploads` |
| 临时访问 | `Presign`，支持 `GET`、`HEAD`、`PUT` |

工具包提供 Go API，本次没有增加前端页面或 HTTP 路由。桶策略、生命周期、版本管理、跨区域复制配置等云平台治理功能不属于当前文件管理接口。

## 上传备份

### 小文件：保存 JSON 备份清单

`PutObject` 需要准确的字节长度。下面的函数适合小型清单、配置等数据，先编码到内存，再上传；大型备份使用后面的流式或文件接口，避免将全部内容读入内存。

```go
package example

import (
	"bytes"
	"context"
	"encoding/json"
	"server/app/util/oss"
)

func SaveManifest(ctx context.Context, store oss.Interface, key string, manifest any) (*oss.Object, error) {
	data, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	return store.PutObject(ctx, key, bytes.NewReader(data), int64(len(data)), oss.PutOptions{
		ContentType: "application/json",
	})
}
```

### 大文件与未知长度数据流

小文件已知长度时使用 `PutObject`；本地文件推荐 `UploadFile`，它自动选择简单上传或分片上传：

```go
object, err := manager.UploadFile(ctx, "database/2026-10-28.sql.gz", filename, oss.UploadOptions{
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
object, err := manager.Upload(ctx, "database/latest.sql.gz", backupReader, -1, oss.UploadOptions{
    PutOptions: oss.PutOptions{ContentType: "application/gzip"},
})
```

默认分片为 16 MiB，按顺序上传，在单次上传内复用同一块分片缓冲，避免逐片分配大块内存。分片超过 64 MiB 时转存临时文件，`TempDir` 可指定备份任务的临时目录；临时文件在调用结束前清理。已知长度的文件会自动扩大分片，避免超过 10000 片。统一限制为：对象最多 5 TiB，单片最多 5 GiB，非末片至少 5 MiB。未知长度时不能提前调整分片，默认分片最多覆盖约 156 GiB；更大的流应预先增大 `PartSize`，达到上限会返回错误并中止上传，不会发布截断对象。

自动上传失败或取消后，会用独立的 30 秒 context 中止分片会话；中止也失败时返回聚合错误，包含可用于后续清理的上传 ID。调用方应设置任务期限。任意 `io.Reader` 无法被工具包强制打断；对于阻塞的 `io.Pipe`，备份任务必须在取消时关闭管道或停止生产者。

上传进度代表厂商已确认的字节数，分片全部上传后还需成功合并。`PutObject` 和 `Upload` 的已知长度表示要上传的字节数，额外尾随内容不属于本次上传；备份生产者发生错误时应通过 reader 返回该错误，不能把失败包装成正常 EOF。

## 分页浏览

云对象列表采用游标分页，没有可靠的总条数或随机页码。将 `NextToken` 原样传入下一次请求，并保持相同的前缀和分隔符：

```go
options := oss.ListOptions{Prefix: "database/", Delimiter: "/", Limit: 100}
for {
    page, err := manager.ListObjects(ctx, options)
    if err != nil {
        return err
    }
    // page.Objects 是文件；page.CommonPrefixes 是子目录。
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

`GetObject` 返回流式 `Body`，调用方负责关闭。`GetOptions.Range` 接受单个 HTTP 范围，例如 `bytes=0-1023`、`bytes=1024-`；通过 `IfMatch` 固定源 ETag，适合断点下载。

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
	download, err := store.GetObject(ctx, key, oss.GetOptions{})
	if err != nil {
		return 0, err
	}
	defer func() { err = errors.Join(err, download.Body.Close()) }()
	return io.Copy(destination, download.Body)
}
```

`DownloadFile` 先写入目标同目录的临时文件，校验读取长度并同步成功后替换目标；失败时保留原目标文件。该方法下载完整对象，分段下载请用 `GetObject`。

`CopyObject` 默认在当前桶内复制，`CopyOptions.SourceBucket` 可指定同厂商同区域的另一个桶。复制前会读取源元数据，超过 1 GiB 自动使用云端分片复制，每一片携带源 ETag 条件，避免复制过程中源变化产生混合文件。分片复制失败会中止会话。`UploadPartCopy` 同样可直接用于调用方管理的复制任务。

`MoveObject` 用复制加删除完成重命名；复制成功后才删除源。它不是原子操作，删除失败时返回目标信息和错误，两份对象都会保留。执行期间调用方应保证源不被并发修改。所有写入遵循云存储覆盖语义；目标已存在时可能被覆盖。`SetMetadata` 整体替换元数据和内容头，应传入需要保留的字段；复制不会保留源对象 ACL 等独立访问策略。

`DeleteObjects` 返回每个 key 的结果和聚合错误，方便区分部分成功。`DeleteBucket` 只删除空桶。归档对象下载/复制、存储类型、跨桶权限等仍受厂商规则约束；`StorageClass` 使用相应厂商的原生值。

```go
signed, err := manager.Presign(ctx, "database/latest.sql.gz", "GET", 15*time.Minute)
// 客户端访问 signed.URL，并携带 signed.Headers 中的头。
```

预签名有效期为 1 秒到 7 天，同时受临时凭据剩余有效期限制。签名地址具备临时访问权限，不应写入公开日志。

## 错误与验证

所有网络操作接受 `context.Context`。使用 `errors.Is` 检查 `ErrInvalidArgument`、`ErrNotFound`、`ErrPermissionDenied`、`ErrConflict`、`ErrPreconditionFailed` 或 context 的取消/超时错误。使用 `errors.As` 提取 `*oss.Error` 获取厂商、操作、状态码和错误码；其 `Error()` 不输出原始 SDK 响应，避免日志泄漏签名内容。`Exists` 仅将 404 视为不存在，不会吞掉权限或网络错误。

可传入 `Config.HTTPClient` 配置代理、Transport 和超时。工具包克隆 client 并禁用自动重定向，避免认证请求被转发到错误的区域或主机；对于标准 `*http.Transport` 还会克隆并禁用自动解压，确保压缩备份按原始字节下载。自定义 Transport 也应保留响应原始字节，不做自动解压。阿里 `HeadBucket` 通过 `GetBucketInfo` 实现，需要对应读桶信息权限。

使用编译和静态检查验证包及所有厂商子包：

```sh
cd server
go build ./app/util/oss/...
go vet ./app/util/oss/...
```

官方 SDK：[阿里 OSS](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2)、[腾讯 COS](https://github.com/tencentyun/cos-go-sdk-v5)、[华为 OBS](https://github.com/huaweicloud/huaweicloud-sdk-go-obs)、[火山 TOS](https://github.com/volcengine/ve-tos-golang-sdk)、[百度 BOS](https://github.com/baidubce/bce-sdk-go)。

各适配器说明：[阿里](aliyun/README.md)、[腾讯](tencent/README.md)、[华为](huawei/README.md)、[火山](volcengine/README.md)、[百度](baidu/README.md)。完整公共字段和错误类型见 [types](types/README.md)，其他工具见 [util 索引](../README.md)。
