# cache：进程内缓存与短期互斥

导入路径：`server/app/util/cache`。本包封装 `patrickmn/go-cache`，提供普通缓存、按需加载和独立的带过期时间锁表。适合当前面板进程内的枚举缓存、短期计算结果及重复任务抑制；它不是跨进程或分布式缓存。

## 初始化与生命周期

`NewMem(defaultExpiration, cleanupInterval time.Duration) *Cache` 创建实例。业务依赖建议声明为 `cache.Interface`，由启动层注入一个共享实例。

- 默认过期时间作用于 `Set`，`SetWithExpire` 可以覆盖。
- 缓存条目过期时间沿用底层库约定：`0` 使用实例默认值，负值表示不自动过期。
- `cleanupInterval > 0` 时启动后台 goroutine，定期清理过期缓存和锁；非正值不启动后台清理。读取过期缓存仍会返回未命中。
- 由实例拥有者在应用退出时调用 `Close()`，不要在每个请求中关闭共享缓存。`Close` 可重复调用，停止清理并清空锁表，但不会清空全部缓存值，也不将实例置为禁止使用状态。

## API

| 方法 | 行为 |
| --- | --- |
| `Get(key)` | 返回缓存值，未命中返回 `nil`；需要类型断言 |
| `Set(key, value)` | 使用默认过期时间写入 |
| `SetWithExpire(key, value, duration)` | 指定条目过期时间 |
| `Delete(key)` | 删除普通缓存，不影响同名锁 |
| `Remember(key, loader, duration)` | 命中直接返回，否则执行 loader 并缓存其返回值 |
| `Lock(key, duration) bool` | 已有未过期锁时返回 false，否则设置新锁 |
| `IsLock(key) bool` | 查询锁是否仍有效 |
| `UnLock(key)` | 无条件删除该键的锁 |
| `CleanExpiredLock()` | 手动清理锁表中的过期记录 |
| `Close()` | 停止清理协程并等待退出 |

## 示例：缓存计算结果

```go
package example

import (
    "fmt"
    "time"
    "server/app/util/cache"
)

func CachedNames() {
    c := cache.NewMem(5*time.Minute, time.Minute)
    defer c.Close()

    value := c.Remember("demo:names", func() interface{} {
        return []string{"web", "database"}
    }, 30*time.Second)
    if names, ok := value.([]string); ok {
        fmt.Println(names)
    }
    c.Delete("demo:names")
}
```

生产场景中将 `c` 放到应用依赖容器，不要像演示函数这样每次重建。命名键时加业务前缀，避免不同功能缓存相互覆盖。

## 示例：阻止同进程重复启动任务

```go
package example

import (
    "errors"
    "time"
    "server/app/util/cache"
)

func RunOnce(c cache.Interface, run func() error) error {
    const key = "backup:running"
    if !c.Lock(key, 10*time.Minute) {
        return errors.New("已有备份任务运行")
    }
    defer c.UnLock(key)
    return run()
}
```

锁的期限必须覆盖任务实际运行时间，且使用正值。锁没有持有者 token、续租或比较删除机制：旧任务过期后执行 `UnLock`，可能删掉新任务取得的锁。因此本示例适合运行时间受控的同进程任务；需要严格任务互斥时使用业务锁或任务调度层。

## 并发与数据约定

缓存表及锁操作内部有同步保护，但缓存中的 map、slice、指针对象不会深拷贝。读取后修改这些对象需要调用方同步。`Remember` 不是 singleflight，多个并发未命中可以同时执行 loader；昂贵加载应在外层做合并。loader 没有 error 返回通道，需处理错误的查询应先执行查询，再决定是否写入缓存。

源码：[接口](cache.go)、[实现](memory.go)。相关工具：[job](../job/README.md)、[process](../process/README.md)。
