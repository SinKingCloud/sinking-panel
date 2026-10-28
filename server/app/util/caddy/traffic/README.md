# traffic：Caddy 并发请求与响应速度限制

导入路径：`server/app/util/caddy/traffic`。此包注册 `http.handlers.traffic_limit`，供 [caddy Manager](../README.md) 在站点处理链中限制并发请求和响应速度。常规业务配置 `Site.TrafficLimit`，根包已通过 [modules.go](../modules.go) 完成注册。

这里的并发单位是“正在执行的 HTTP 请求”，不是 TCP 连接数量；一个 HTTP/2 连接的多个并发流会分别计数。时间窗内的请求次数限制由 Site.RateLimit 提供，不属于本模块。

## 配置字段

| Handler 字段 / JSON 名 | 含义 |
| --- | --- |
| `Scope` / `scope` | 必填作用域；相同值共享并发计数，Manager 使用站点 ID。 |
| `MaxConnections` / `max_connections` | 作用域内并发请求总上限。 |
| `MaxConnectionsPerIP` / `max_connections_per_ip` | 单客户端 IP 并发请求上限。 |
| `RatePerRequest` / `rate_per_request` | 每个请求的响应体传输速度，单位字节/秒。 |

三个数值均不可为负，0 表示该项不限制；至少有一项大于 0。Scope 会去除首尾空格，不能包含换行或 NUL。

## 示例：限制下载站点

`manager` 和 `siteID` 指向已有站点。此完整函数设置最多 100 个并发请求、每 IP 最多 4 个、单请求响应速度约 2 MiB/s。

```go
package example

import panelcaddy "server/app/util/caddy"

func LimitDownloads(manager *panelcaddy.Manager, siteID string) error {
    site, err := manager.Site(siteID)
    if err != nil {
        return err
    }
    site.TrafficLimit = panelcaddy.TrafficLimitOptions{
        Enabled: true,
        MaxConnections: 100,
        MaxConnectionsPerIP: 4,
        RatePerRequest: 2 << 20,
    }
    return manager.UpdateSite(*site)
}
```

如果维护原始 Caddy JSON，下面是 `handle` 数组中的单个处理器片段，不是完整服务器配置；它应放在需要限制的响应处理器之前。使用根包管理器的 SaveConfig 时模块已经注册。

```json
{
  "handler": "traffic_limit",
  "scope": "downloads",
  "max_connections": 100,
  "max_connections_per_ip": 4,
  "rate_per_request": 2097152
}
```

## Caddy 生命周期 API

- `CaddyModule()` 返回模块 ID 和工厂。
- `Provision(caddy.Context) error` 校验配置并取得共享状态。HTTP、HTTPS 以及热加载前后的相同 Scope 复用计数，切换配置不会把正在执行的请求计数清零。
- `Validate() error` 检查作用域和数值。
- `ServeHTTP(w, r, next) error` 执行并发检查、包装响应写入，然后调用下一处理器。
- `Cleanup() error` 释放模块引用；在途请求结束后仍可正常归还计数，最后一份引用和请求退出后释放状态。

业务不应只 `new(Handler)` 后直接 ServeHTTP；未 Provision 或已 Cleanup 的处理器会返回 503。生命周期应交给 Caddy 管理。

## 响应与限制边界

总并发达到上限返回 503，单 IP 上限返回 429；都附带 `Retry-After: 1` 与禁止缓存的响应头。模块没有等待队列，超额请求直接返回。

客户端 IP 优先取 Caddy 的 `client_ip` 变量，再退回 RemoteAddr。部署在代理之后时，正确设置 Manager.Options 中的 TrustedProxies 和 ClientIPHeaders，否则多个用户可能被计算成同一个代理 IP；不要直接信任任意 X-Forwarded-For。

响应速度使用每请求独立限速器，允许小幅突发，不是整个站点共享带宽上限。取消请求 context 会终止等待。限速只覆盖 HTTP 响应体，不限上传请求体；WebSocket 等通过 Hijack 升级后的连接不再经过 HTTP 响应限速。

总计数使用原子操作，IP 计数按分片加锁，完成请求时回收对应记录。配置加载后不要并发修改 Handler 的公开字段；通过 Manager.UpdateSite 生成新的配置实例即可。
