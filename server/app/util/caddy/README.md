# caddy：嵌入式网站服务与证书管理

导入路径：`server/app/util/caddy`。此包把 Caddy 嵌入当前进程，对外提供结构化站点管理、配置热加载、静态文件、HTTP/FastCGI 反向代理、TLS、WAF、响应缓存、请求频率限制和流量限制。业务层保存 `Site`，Manager 负责校验并生成运行配置；不需要另起 caddy 命令行进程。

相关子包：[cache 响应缓存存储](cache/README.md)、[traffic 并发和响应限速](traffic/README.md)、[waf 审计日志](waf/README.md)、[dns/baidu 百度 DNS 验证](dns/baidu/README.md)。这些模块已由本包的 [modules.go](modules.go) 注册，通常只导入本包即可。

## 初始化与两种配置模式

```go
func NewManager(root string, options ...Options) (*Manager, error)
```

`root` 必填，会转为绝对路径、解析符号链接并创建目录；最多传一个 Options。初始化只准备配置，不开始监听。一个进程同一时间只能有一个运行中的 Manager，因为底层 Caddy 运行时为进程全局实例。

- **结构化站点模式**：使用 `Site`、`SyncSites`、`AddSite` 等管理；适合面板数据库与网站配置联动。
- **原始 JSON 模式**：`SaveConfig` 接收完整的 Caddy JSON。`ConfigPath` 已有配置文件时，NewManager 也会进入此模式。原始 JSON 不会反向解析成 Site，不能继续直接增删改站点；用 `SyncSites` 明确切回结构化模式。

`ConfigPath` 保存的是 Caddy 配置快照，不是站点数据库。应用重启后，如果希望继续通过 Site 管理，应从业务数据库读取完整站点集合并调用 `SyncSites`，再 `Start`。不要仅凭磁盘快照重建业务配置。

## 全局配置

| Options 字段 | 默认与行为 |
| --- | --- |
| `HTTPListen` / `HTTPSListen` | nil 分别使用 `:80` / `:443`；显式 `[]string{}` 关闭对应监听。 |
| `Protocols` | nil 使用 h1、h2、h3；h2/h2c 需同时启用 h1，h3 用于 TLS。 |
| `DefaultSite` | 未匹配域名时使用的站点 ID；空时显示网站不存在页。 |
| `DataPath` / `ACMEPath` / `CachePath` | 默认 `root/data`、`root/acme`、`root/cache`。 |
| `LogPath` / `WAFLogPath` | 默认 `root/logs/http.log`、`root/logs/waf.log`；前者为 `-` 时不保存 HTTP 运行日志。 |
| `ConfigPath` | 空值不保存快照；设置后写入完整 Caddy JSON，文件权限为 `0600`。 |
| `LogLevel` | 默认 INFO；支持 DEBUG、INFO、WARN、ERROR、PANIC、FATAL。 |
| `ReadTimeout` / `WriteTimeout` | 0 表示不限制完整请求读取/响应写入时间。 |
| `ReadHeaderTimeout` / `IdleTimeout` / `GracePeriod` | 默认 10 秒、5 分钟、10 秒。 |
| `MaxHeaderBytes` | 默认 1 MiB。 |
| `TrustedProxies` / `ClientIPHeaders` / `TrustedProxiesStrict` | 配置代理 IP/CIDR、真实客户端地址来源与严格解析；影响限流、日志等客户端 IP 语义。 |
| `HTTPChallengeHost` / `HTTPChallengePort` / `ACMEEmail` | HTTP-01 验证地址、端口及默认账户邮箱；验证端口默认 80。 |
| `InitialPage` / `NotFoundPage` / `SiteNotFoundPage` / `SiteDisabledPage` | 可替换默认响应正文、状态码和响应头。 |

相对路径按 Manager 的 root 解析。Go 中的超时值使用 `time.Duration`；直接用标准 JSON 读写本包配置时，duration 数值为纳秒，不能把秒数直接赋给它。

## 示例：启动静态站点并平滑退出

下例为完整函数，`webRoot` 应为已有的绝对路径。监听本机 8080，访问 `http://localhost:8080`；生产环境替换监听地址、域名和站点目录。

```go
package example

import (
    "context"
    "errors"

    panelcaddy "server/app/util/caddy"
)

func ServeStatic(ctx context.Context, dataRoot, webRoot string) (err error) {
    manager, err := panelcaddy.NewManager(dataRoot, panelcaddy.Options{
        HTTPListen:  []string{"127.0.0.1:8080"},
        HTTPSListen: []string{},
        Protocols:   []string{"h1"},
    })
    if err != nil {
        return err
    }
    defer func() { err = errors.Join(err, manager.Stop()) }()
    if err = manager.Start(panelcaddy.Site{
        ID:      "demo",
        Name:    "演示网站",
        Enabled: true,
        Domains: []string{"localhost"},
        Root:    webRoot,
        Index:   []string{"index.html"},
        Compression: panelcaddy.CompressionOptions{Enabled: true},
    }); err != nil {
        return err
    }
    <-ctx.Done()
    return nil
}
```

`Site.Enabled` 的零值为 false，需要显式开启。站点 ID 只能使用字母、数字、点、下划线、短横线；至少绑定一个合法域名，域名不带协议和路径。业务必须避免多个站点争抢同一域名。

## 生命周期与站点 API

| 方法 | 行为 |
| --- | --- |
| `Start(sites ...Site) error` | 非空参数替换整组站点后启动；无参数启动当前保存的配置。重复启动报错。 |
| `Stop() error` | 优雅停止并释放缓存资源，保留内存配置；未启动时可重复调用。 |
| `Restart() error` | 停止再启动当前配置。 |
| `Reload() error` | 强制重载当前运行配置，例如重新读取同路径证书文件；要求已运行。配置内容有变化时应使用相应更新 API。 |
| `Running() bool` | 当前 Manager 是否拥有运行中的服务。 |
| `Options() Options` / `UpdateOptions(options) error` | 获取副本/提交完整全局配置；修改部分字段时先读取再修改，避免零值覆盖其他设置。 |
| `Sites() []*Site` / `Site(id) (*Site, error)` | 获取排序后的全部站点副本/单个站点副本；修改返回值不会自动生效。 |
| `ValidateSites(sites) error` | 校验拟提交的完整站点集合。 |
| `SyncSites(sites) error` | 用完整集合替换当前站点；传空列表表示清空。 |
| `AddSite(site)` / `UpdateSite(site)` / `DeleteSite(id)` | 返回 error；分别新增、完整更新和删除，运行时同步热加载。 |
| `EnableSite(id)` / `DisableSite(id)` | 返回 error；切换已有站点开关。 |
| `ClearSiteCache(id) error` | 清空当前站点 HTTP/HTTPS 缓存，不重启网站。 |
| `DeleteSiteCache(id) error` | 删除已卸载站点的缓存目录；仍被路由使用时拒绝。 |
| `Config() []byte` / `SaveConfig(config) error` | 获取完整配置副本/保存并应用原始 Caddy JSON。 |
| `LogPath(id, logType) (string, error)` | 计算日志位置；LogServer 使用全局日志，忽略 id。 |

更新有操作锁保护，并在校验、加载或保存失败时尝试恢复旧配置。出现回滚本身失败的错误后，应核对 `Running()` 和 `Config()`，不要假定运行时一定保持原状。业务层还要自行处理“写数据库”和“更新运行配置”的一致性。

站点访问日志、WAF 审计日志和业务进程日志路径分别为 `root/<ID>/http.log`、`waf.log`、`process.log`。`LogProcess` 只提供约定路径，不会自动启动进程或创建进程日志；可交给 [process](../process/README.md) 的 LogPath 使用。

## 示例：增加代理站点及局部更新

`manager` 已由应用持有，可在启动前或运行时调用。`UpdateSite` 接收完整 Site，因此先读取副本再改字段。

```go
package example

import (
    "time"

    panelcaddy "server/app/util/caddy"
)

func AddAPI(manager *panelcaddy.Manager) error {
    return manager.AddSite(panelcaddy.Site{
        ID:      "api",
        Enabled: true,
        Domains: []string{"api.example.com"},
        Proxy: &panelcaddy.ProxyOptions{
            Upstreams: []panelcaddy.Upstream{{Dial: "127.0.0.1:9000"}},
            DialTimeout: 5 * time.Second,
            ResponseHeaderTimeout: 30 * time.Second,
        },
        Routes: []panelcaddy.Route{{
            Name: "health",
            Paths: []string{"/health"},
            Methods: []string{"GET"},
            Response: &panelcaddy.ResponseOptions{Status: 200, Body: "ok"},
        }},
    })
}

func LimitAPI(manager *panelcaddy.Manager) error {
    site, err := manager.Site("api")
    if err != nil {
        return err
    }
    site.RateLimit = panelcaddy.RateLimitOptions{
        Enabled: true, Window: time.Minute, MaxEvents: 120,
    }
    site.TrafficLimit = panelcaddy.TrafficLimitOptions{
        Enabled: true, MaxConnections: 200, MaxConnectionsPerIP: 20,
    }
    return manager.UpdateSite(*site)
}
```

常用站点功能如下，完整字段和枚举见 [entities.go](entities.go) 与 [enum.go](enum.go)：

- 静态文件：`Root`、`Index`、`TryFiles`、`Hide`、`Precompressed`；`Browse=true` 开启目录浏览。
- 路由：`Redirects`、`Routes` 优先于默认站点处理器。Route 可按路径和方法匹配，再选择固定响应、静态文件、代理、StripPrefix 或 Rewrite。
- HTTP 代理：`Proxy.Upstreams`、重试、健康检查、超时和流连接设置。HTTPS 上游还可设 `ProxySchemeHTTPS` 与 SNI；当前 `TLSInsecureSkipVerify` 默认 true，需要校验上游证书时显式传入指向 false 的指针。
- FastCGI：`Transport: ProxyTransportFastCGI`，配合 PHP-FPM 地址、Root、SplitPath、Index、TryFiles 和 Env。PHP-FPM 进程应由外部服务或 process 管理器负责。
- 响应优化：`Compression` 默认优先 zstd 再 gzip；`Cache` 默认 TTL 2 分钟、容量 256 MiB、GET/HEAD 方法，默认无显式缓存控制的响应按 `no-store` 处理。
- 访问控制：`WAF` 支持检测/拦截、内嵌 OWASP CRS、规则分组与自定义指令；`RateLimit` 控制时间窗内请求数，`TrafficLimit` 控制同时在处理的请求数及单请求响应速度，含义不同。
- 扩展：`Handlers` 接受已注册模块的 JSON，`HandlerOrder` 调整阶段顺序。输入需要符合实际模块协议，不能把未注册的模块名称直接写进配置。

## TLS：部署已有证书

开启 TLS 不会自动申请证书。先准备有效的证书链和匹配私钥，再更新站点。证书必须覆盖期望启用的域名；可通过 CertificatePair.Domains 指定该证书用于哪些站点域名。

```go
package example

import panelcaddy "server/app/util/caddy"

func DeployTLS(manager *panelcaddy.Manager, siteID, certificateFile, keyFile string) error {
    site, err := manager.Site(siteID)
    if err != nil {
        return err
    }
    site.TLS = panelcaddy.TLSOptions{
        Enabled: true,
        RedirectHTTP: true,
        Certificates: []panelcaddy.CertificatePair{{
            CertificateFile: certificateFile,
            KeyFile: keyFile,
        }},
    }
    return manager.UpdateSite(*site)
}
```

也可使用 `CertificatePEM` 与 `PrivateKeyPEM`，文件方式与内存方式不能混用。PrivateKeyPEM 不参与 CertificatePair 的标准 JSON 序列化，业务持久化需要另行保存；生成后的完整 Caddy JSON 仍可能包含私钥，`Config()` 不能作为普通公开接口响应。

## 自动申请与手动验证

`ObtainCertificate(ctx, request)`、`RenewCertificate(ctx, request)` 返回 `(*Certificate, error)`。它们只签发或续签，不会为站点自动开启 TLS，也不负责安排定时续期；业务应保存返回结果后调用站点更新。自动模式会保存证书文件，返回文件路径和 PEM。单次自动申请总等待上限为 `CertificateRequestTimeout`，当前 30 秒，也受调用方 context 的更早截止时间约束。

CertificateRequest 包含 Domain、Email、CA、Challenge、Action、DNSProvider、DNSCredentials 和可复用的 AccountKeyPEM。CA 支持 production/staging，默认生产；测试联调建议显式用 staging，其证书不能作为浏览器受信任生产证书。HTTP 验证需要公网能到达相应挑战服务；DNS 验证需要 DNS 修改权限及记录传播。通配符只能 DNS 验证，公网 IP 只支持 HTTP 验证。

```go
package example

import (
    "context"

    panelcaddy "server/app/util/caddy"
)

func ObtainAliyunCertificate(ctx context.Context, manager *panelcaddy.Manager,
    domain, email, accessKeyID, accessKeySecret string) (*panelcaddy.Certificate, error) {
    return manager.ObtainCertificate(ctx, panelcaddy.CertificateRequest{
        Domain: domain,
        Email: email,
        CA: panelcaddy.CertificateCAStaging,
        Challenge: panelcaddy.CertificateChallengeDNS,
        DNSProvider: panelcaddy.DNSProviderAliDNS,
        DNSCredentials: panelcaddy.DNSCredentials{
            AliyunAccessKeyID: accessKeyID,
            AliyunAccessKeySecret: accessKeySecret,
        },
    })
}
```

DNS 服务商映射：

| DNSProvider | 凭据字段 |
| --- | --- |
| `DNSProviderAliDNS` | AliyunAccessKeyID / AliyunAccessKeySecret |
| `DNSProviderTencentCloud` | TencentSecretID / TencentSecretKey |
| `DNSProviderHuaweiCloud` | HuaweiAccessKeyID / HuaweiSecretAccessKey |
| `DNSProviderVolcengine` | VolcengineAccessKeyID / VolcengineAccessKeySecret |
| `DNSProviderBaiduCloud` | BaiduAccessKeyID / BaiduSecretAccessKey |

手动模式使用独立的 `NewAcme()`：先 `Obtain(ctx, request)` 或 `Renew` 获得 `[]acme.Challenge`，业务引导用户配置挑战记录/文件，待准备完成后对**同一个订单实例**调用 `Submit(ctx)` 获得 Certificate。Manager 的自动方法不接受手动模式。以下片段展示调用顺序，`request` 是已经设置 Domain、Email、Challenge、`Action: CertificateActionManual` 的 CertificateRequest，`ctx` 是有效上下文：

```go
order := panelcaddy.NewAcme()
challenges, err := order.Obtain(ctx, request)
if err != nil {
    return err
}
// 将 challenges 中的挑战信息交给业务页面；等待用户实际完成配置。
_ = challenges
// 这里应在后续“提交验证”操作中执行，而不是未配置挑战就立即执行。
certificate, err := order.Submit(ctx)
if err != nil {
    return err
}
_ = certificate // 保存证书并部署到站点。
```

一个 Acme 实例同一时间处理一个订单，并发操作会返回“正在处理”；不要让多个用户共享一个订单实例。订单可序列化后恢复，但其中包含账户私钥和证书私钥，只能存入服务端受控存储，不能把整个订单 JSON 发给前端。即使 challenges 为空，仍需 Submit 获取签发结果。

## 资源与错误边界

应用退出时始终调用 Stop；不要绕过 Manager 直接操作底层全局 Caddy 运行时。站点删除只卸载配置并尝试清理缓存，不删除网站根目录和证书文件。缓存清理失败可能记录警告，业务不要因此假定站点删除已回滚。

证书请求、DNS 权限、端口占用、目录权限、证书域名覆盖和模块配置均可能返回错误，应保留错误上下文。依赖代理传来的客户端 IP 时，要先正确配置 TrustedProxies，再启用按 IP 限流；直接信任任意请求头会破坏限制语义。
