# baidu：百度智能云 ACME DNS 验证适配

导入路径：`server/app/util/caddy/dns/baidu`。此包实现 libdns 的记录新增和删除接口，并注册 Caddy 模块 `dns.providers.baiducloud`，用于 ACME DNS-01 挑战记录管理。

它操作的是百度云 DNS，不是 BOS 对象存储。对象文件上传请使用 [oss/baidu](../../../oss/baidu/README.md)；一般网站证书申请使用 [caddy Manager](../../README.md) 的统一证书 API。

## Provider 与公共 API

```go
provider := &baidu.Provider{
    AccessKeyID: "从服务端密钥配置读取",
    SecretAccessKey: "从服务端密钥配置读取",
}
```

上面为初始化片段，需导入 `server/app/util/caddy/dns/baidu`；不要把实际密钥写进源码或公开配置。

| 方法 | 返回和用途 |
| --- | --- |
| `CaddyModule()` | 返回 dns.providers.baiducloud 的模块信息。 |
| `AppendRecords(ctx, zone, records)` | 返回 `([]libdns.Record, error)`，逐条创建 DNS 记录；返回记录包含实际采用的 TTL。 |
| `DeleteRecords(ctx, zone, records)` | 返回 `([]libdns.Record, error)`，分页查询并删除名称、类型和值全部匹配的记录。 |

zone 为托管域名，例如 `example.com.` 或 `example.com`；结尾的点会移除。记录 Name 使用相对于 zone 的名称，例如 `_acme-challenge`。此适配只提供 ACME 需要的新增与删除，不提供通用 DNS 控制台的查询、修改等完整接口。

## 示例：添加验证记录并在业务结束后删除

下面为完整函数。`verify` 由业务实现，在记录添加后等待传播并执行验证；函数始终尝试删除已成功创建的记录，包括部分创建成功后遇到错误的情况。清理使用独立 context，避免主请求取消后无法清理。

```go
package example

import (
    "context"
    "errors"
    "time"

    "server/app/util/caddy/dns/baidu"
    "github.com/libdns/libdns"
)

func WithDNSChallenge(ctx context.Context, accessKeyID, secretAccessKey,
    zone, token string, verify func(context.Context) error) (err error) {
    provider := &baidu.Provider{
        AccessKeyID: accessKeyID,
        SecretAccessKey: secretAccessKey,
    }
    record, err := (libdns.RR{
        Type: "TXT",
        Name: "_acme-challenge",
        Data: token,
        TTL: 5 * time.Minute,
    }).Parse()
    if err != nil {
        return err
    }
    created, err := provider.AppendRecords(ctx, zone, []libdns.Record{record})
    if len(created) > 0 {
        defer func() {
            cleanupCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
            defer cancel()
            _, cleanupErr := provider.DeleteRecords(cleanupCtx, zone, created)
            err = errors.Join(err, cleanupErr)
        }()
    }
    if err != nil {
        return err
    }
    return verify(ctx)
}
```

`verify` 必须非 nil。这段示例会实际修改传入域名下的 DNS；通常不必在业务里手动调用，Manager.ObtainCertificate 已通过 libdns 管理自动挑战的创建与清理。

## 通过统一证书接口接入

根 caddy 包已注册本模块，业务只要在 CertificateRequest 中设置：

```go
request := panelcaddy.CertificateRequest{
    Domain: "*.example.com",
    Email: "ops@example.com",
    Challenge: panelcaddy.CertificateChallengeDNS,
    DNSProvider: panelcaddy.DNSProviderBaiduCloud,
    DNSCredentials: panelcaddy.DNSCredentials{
        BaiduAccessKeyID: accessKeyID,
        BaiduSecretAccessKey: secretAccessKey,
    },
}
certificate, err := manager.ObtainCertificate(ctx, request)
```

该片段需 `import panelcaddy "server/app/util/caddy"`，并已有 manager、ctx 及从服务端密钥管理取得的凭据；调用方继续处理 certificate 和 err，保存证书并部署到站点。正式/测试 CA 可通过 request.CA 明确指定，默认生产环境。

## TTL、分页与部分失败

- 新增时最低 TTL 为 300 秒，小于该值会提升；只保留完整秒数，超过 int32 秒范围返回错误。以返回记录的 TTL 为准。
- 多条记录逐条处理，不保证事务性。发生错误时，结果切片仍包含已完成的部分，调用方需要清理或重试。
- 删除先查完所有分页，再按记录 ID 删除，避免边删边翻页导致漏项；分页标记缺失或不前进会报错。
- 删除按名称、类型和值完整匹配，不会删除同名称下不同 token 的其他验证记录；相同三项的重复记录可能全部删除。

## 并发、超时与密钥更新

百度 SDK 调用会修改共享 HTTP 客户端超时配置，因此本包使用进程级互斥锁串行化所有 Provider 实例的请求。它并不适合高吞吐批量 DNS 管理。

客户端按实例惰性初始化，sync.Once 会同时缓存初始化错误。设置凭据后再首次调用；密钥轮换或首次因缺少凭据而失败时应创建新 Provider，不能只修改旧实例字段。运行期间也不要并发写这些字段。

SDK 不支持请求级 context：本包在开始和逐条操作前检查取消，并将单次 SDK 连接超时设为 30 秒、禁用 SDK 重试；已经发出的请求不能保证被 context 立即中断，等待全局互斥锁也不能即时响应取消。因此调用方 context 截止时间不等于硬性墙钟耗时上限。Provider 无需显式 Close，DNS 验证记录仍应按业务流程清理。
