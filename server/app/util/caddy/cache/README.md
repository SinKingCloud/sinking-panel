# cache：Caddy 响应缓存的本地存储模块

导入路径：`server/app/util/caddy/cache`。此包为上层 cache-handler 提供本地 bbolt 存储，模块 ID 为 `storages.cache.simplefs`。它是 [caddy Manager](../README.md) 的内部配套模块；业务站点通常设置 `Site.Cache`，无需直接创建 SimpleFS。

不要与 [util/cache](../../cache/README.md) 的业务缓存混淆：本包保存 HTTP 响应和缓存元数据，不是任意业务对象的通用缓存接口。

## 注册与生命周期

导入 `server/app/util/caddy` 时，[modules.go](../modules.go) 自动导入本包和上游 cache-handler。启用 Site.Cache 后，Manager 生成 `handler: cache` 配置，指定 `storers: ["simplefs"]`，Caddy 再调用本模块的 Provision/Cleanup。

本包沿用上游 simplefs 的相同模块 ID，因此不要在同一个二进制中再导入另一份 `storages.cache.simplefs` 实现，否则会产生重复注册。

`SimpleFS.Provision` 只校验目录和容量并取得进程内共享存储引用，不创建目录、数据库或后台 goroutine。首次写入才打开/创建 `cache.db`；已有数据库在读取时按需打开。`Cleanup` 释放引用，最后一个使用者退出后关闭数据库。Manager.Stop 会执行总体资源清理。

## 示例：为已有站点开启响应缓存

`manager` 是已初始化的 Manager，`siteID` 是已有站点。以下完整函数通过统一站点 API 配置存储；示例仅适合允许公共缓存的资源，个性化响应应由上游返回合适的缓存控制头。

```go
package example

import (
    "time"

    panelcaddy "server/app/util/caddy"
)

func EnableResponseCache(manager *panelcaddy.Manager, siteID string) error {
    site, err := manager.Site(siteID)
    if err != nil {
        return err
    }
    site.Cache = panelcaddy.CacheOptions{
        Enabled: true,
        TTL: 2 * time.Minute,
        MaxSize: 256 << 20,
        Methods: []string{"GET", "HEAD"},
        KeyHeaders: []string{"Accept-Language"},
        DefaultCacheControl: "public, max-age=120",
    }
    return manager.UpdateSite(*site)
}
```

省略 DefaultCacheControl 时默认 `no-store`，不会默认把缺少明确缓存策略的响应视为公共内容。HTTP/HTTPS 缓存分开保存于 `CachePath/<siteID>/<scope>/cache.db`。KeyHeaders 用于区分响应变体；启用缓存不应忽略业务自己的 Cookie、授权信息和缓存响应头。

## 配置协议

下表针对 `simplefs.configuration` 中的存储设置；常规使用应交给 Manager 生成，避免手工维护上游配置结构。

| 字段 | 行为 |
| --- | --- |
| `path` | 也可来自 Provider.Path；最终规范化为绝对路径。未配置时退回当前工作目录。 |
| `allowed_root` | 可选受控根目录；配置后会校验目标路径位于其中，Manager 总会设置它。 |
| `directory_size` | 缓存逻辑数据容量，支持整数或可解析的容量字符串；0 表示不限制，Manager 默认传 256 MiB。 |
| `size` | 用于兼容上游存储标识，不等同于 directory_size 容量约束。 |

容量限制按存储的逻辑数据计算，不是数据库文件物理大小硬上限。bbolt 文件包含页面、索引和空闲空间，删除数据后不能假定磁盘文件立即按等量缩小。Manager 同时将 CacheOptions.MaxSize 传给上游单响应大小限制和本模块容量限制。

## 内部 API 与清理

| API | 职责 |
| --- | --- |
| `SimpleFS.CaddyModule` / `Provision` / `Cleanup` | Caddy 模块发现、初始化和资源释放。 |
| `Name` / `Uuid` / `Init` | 上游存储接口识别与初始化协议；Init 本身不负责打开数据库。 |
| `Get` / `Set` / `Delete` / `DeleteMany` / `Reset` | cache-handler 使用的键值读写、批量删除及清空。 |
| `MapKeys` / `ListKeys` | 枚举缓存键。 |
| `GetMultiLevel` / `SetMultiLevel` | 适配上游多级缓存读写。 |
| `Registry.PurgeTree(allowedRoot, path) error` | 清空受控树内缓存，允许清理仍被站点使用的缓存。 |
| `Registry.RemoveTree(allowedRoot, path) error` | 删除目录；存在活动引用时返回 ErrStoreInUse。 |
| `Registry.CloseAll() error` | 关闭进程内登记的全部缓存资源，供整体服务退出使用。 |

正常管理操作直接使用 `manager.ClearSiteCache(siteID)`；站点卸载后再 `manager.DeleteSiteCache(siteID)`。不要在请求处理中调用全局 CloseAll，也不要用 os.RemoveAll 直接删除仍打开的数据库文件。

同一规范路径的模块共享后端，租约区分容量和 stale 配置，能够跨热加载复用数据库。数据库事务和注册表锁处理并发访问；数据库文件权限为 `0600`、目录为 `0700`，打开锁等待上限为 1 秒。相同缓存目录不适合让多个独立进程同时写入。

一些上游读取接口没有 error 返回值，底层错误可能记录日志后表现为缓存未命中；应结合 Caddy 日志定位权限、磁盘或数据格式问题。缓存属于可重建数据，业务不能将它作为备份或唯一持久化存储。
