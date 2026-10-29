# 通用记录存储

`recordstore.Open` 提供按业务集合隔离的本地持久化存储。集合可以保存 Go 结构体、map、数组、切片、字符串、数字、布尔、null，以及原始字节；数据内容由业务定义。默认永久保存，可为单条记录设置 TTL。

| 入口 | 用途 | 保存方式 |
| --- | --- | --- |
| `Open` → `Store.Collection` | 任意业务记录、配置、文档、业务快照、自定义编码数据 | 独立 `records.db`，bbolt 事务、键和时间索引 |
| `OpenRecorder` | 需要按分钟汇总的数值序列 | 有界异步队列、定长二进制日文件 |
| `OpenSeries` | 业务已有调度的数值序列 | 数值存储，由调用方刷新和清理 |
| `OpenEvents` | 结构化事件记录 | 按天追加 JSONL，提供查询与统计 |

后三项是数值汇总与事件存储扩展；通用 `Store` 没有预设业务字段，也不要求每条记录都是数值指标或事件。

## 打开存储与业务集合

同一目录只打开一个 `Store`，业务之间共享此实例，通过不同集合名隔离记录。不同集合可以使用相同 key，互不覆盖。`Store` 和 `Collection` 可供多个 goroutine 使用；集合不单独关闭。

```go
import "server/app/util/recordstore"

store, err := recordstore.Open("data/records/business")
if err != nil {
    return err
}
defer store.Close()

preferences, err := store.Collection("preferences")
if err != nil {
    return err
}

type Preferences struct {
    Theme     string   `json:"theme"`
    Languages []string `json:"languages"`
}
if err := preferences.Put("user:42", Preferences{
    Theme: "dark", Languages: []string{"zh-CN", "en"},
}); err != nil {
    return err
}
var saved Preferences
if err := preferences.GetInto("user:42", &saved); err != nil {
    return err
}
```

`Put` 是按 key 覆盖写入。`Get` 返回带 `Key`、业务时间 `Time`、过期时间 `ExpiresAt`、`Encoding` 和编码内容 `Data` 的 `Record`；JSON 记录可通过 `record.Decode(&target)` 解码。不存在或已过期返回 `ErrNotFound`，与已保存的 JSON `null`、空字节数组区分。解码到 `interface{}` 时，数字使用 `json.Number`。

集合中的值不必采用同一种结构。以下片段演示文档段落和库存快照，省略打开存储的代码：

```go
drafts, err := store.Collection("drafts")
if err != nil {
    return err
}
// 空间内自动生成 key；返回值可用于 Get、GetInto 或 Delete。
key, err := drafts.Append([]string{"第一段", "第二段"})
if err != nil {
    return err
}
if err := drafts.Delete(key); err != nil {
    return err
}

inventory, err := store.Collection("inventory")
if err != nil {
    return err
}
type Stock struct {
    Available int    `json:"available"`
    Warehouse string `json:"warehouse"`
}
keys, err := inventory.PutMany([]recordstore.Entry{
    {Key: "sku:A", Value: Stock{Available: 12, Warehouse: "east"}},
    {Key: "sku:B", Value: Stock{Available: 8, Warehouse: "west"}},
    {Value: []string{"sku:A", "sku:B"}}, // 空 Key 自动生成
})
if err != nil {
    return err
}
_ = keys
```

`PutMany` 在一个事务中提交所有 JSON 值，全部成功或全部回滚，返回与输入顺序对应的 key。标量和空值也直接支持，例如 `Put("enabled", true)`、`Put("title", "草稿")`、`Put("selection", nil)`。

## 原始字节与大小限制

```go
if err := drafts.PutBytes("encoded:example", []byte{0x00, 0xff, 0x01}); err != nil {
    return err
}
record, err := drafts.Get("encoded:example")
if err != nil {
    return err
}
// record.Encoding == recordstore.EncodingBytes，record.Data 是独立的字节副本。
payload := record.Data
_ = payload
```

JSON 入口使用 Go 的 `encoding/json`：支持可编码的结构体及容器、标量和 null，不支持函数、channel、循环引用或非有限浮点数。需要 protobuf、压缩数据或其他自定义格式时，由业务编码后调用 `PutBytes`，读取后自行解码。普通 `Put` 接收 `[]byte` 时遵循 JSON 的 base64 编码规则；要逐字节保存请使用 `PutBytes`。

默认单条编码数据上限 16 MiB，可通过 `Open(dir, Options{MaxValueBytes: ...})` 调整。通用数据类型不意味着大小无限；图片、大文件等仍需按业务容量决定是否适合放入数据库。

## 时间、TTL 与查询

```go
err = drafts.Put("temporary", []string{"临时内容"}, recordstore.WriteOptions{
    Time: time.Now().UnixMilli(), // 省略时使用本次写入时间
    TTL:  30 * time.Minute,       // 从本次写入时刻计算；零值永久保存
})
if err != nil {
    return err
}

page, err := inventory.Query(ctx, recordstore.Query{
    Start: start.UnixMilli(), End: end.UnixMilli(),
    Prefix: "sku:", Desc: true, Limit: 100,
    Match: func(record recordstore.Record) (bool, error) {
        var item Stock
        if err := record.Decode(&item); err != nil {
            return false, err
        }
        return item.Available > 0, nil
    },
})
if err != nil {
    return err
}
// 下一页传入 page.NextCursor，并保持时间范围、前缀、方向和 Match 逻辑一致。
```

`Time` 是用于排序和查询的业务时间，单位为 Unix 毫秒；TTL 独立地从实际写入时刻计算。覆盖写入会替换时间和过期设置，省略 TTL 会改为永久保存。`Get`、`Query` 和 `Scan` 即时排除已过期记录，后台清理是否执行不影响过期可见性。

查询按 `(Time, Key)` 排序，`Start` / `End` 包含边界，零值表示不限；默认每页 100 条，上限 1000 条，同时按 16 MiB 编码数据量分页。若配置允许单条超过 16 MiB，该条独占一页。游标用于继续读取，不提供跨页全局快照；并发更新或删除记录可能改变后续页内容。

键查找和时间范围使用索引。`Match` 在读事务外运行，可以解码、检查任何业务结构；内容字段没有自动索引，因此复杂条件仍需扫描时间范围内的候选记录。分页边界可能重复执行 `Match`，不要依赖调用次数；业务回调应保持筛选语义稳定，并自行处理可能阻塞的操作。

## 自定义遍历与统计

`Scan` 按分页顺序遍历并调用业务函数，无需一次加载全部记录。例如汇总各仓库库存：

```go
totals := make(map[string]int)
err = inventory.Scan(ctx, recordstore.Query{Prefix: "sku:"}, func(record recordstore.Record) error {
    var item Stock
    if err := record.Decode(&item); err != nil {
        return err
    }
    totals[item.Warehouse] += item.Available
    return nil
})
if err != nil {
    return err
}
```

`Query` / `Scan` 支持 context 取消，回调返回错误会中止遍历。统计规则由业务决定，存储层不会假定请求数、流量或固定字段含义；扫描耗时仍随候选记录量增长。

## 通用存储的生命周期与性能边界

`Store` 不启动后台 goroutine。将 `store.Cleanup(ctx)` 接入现有任务调度，或用 `collection.Cleanup(ctx)` 只清理一个集合；返回清理记录数。服务退出时调用 `store.Close()`，处理其错误。没有 TTL 的记录不会自动删除。

`records.db` 使用 bbolt 默认同步事务提交，同一数据库一次执行一个写事务，读事务支持并发。`PutMany` 可减少多次独立提交的开销。数据库文件持有进程锁，不应重复打开同一路径，也不能让多个进程同时写入。命名空间隔离数据，但集合仍共享同一个数据库的写入能力。

TTL 清理和删除会释放可复用的数据库页面，不保证物理文件立即缩小。查询页上限限制返回数据量，不等于固定内存开销或固定查询时延；JSON 解码、业务回调和并发访问都会影响实际性能。这里没有对生产高并发或任意数据规模作吞吐保证。

## 数值序列与事件存储扩展

`OpenRecorder`、`OpenSeries`、`OpenEvents` 分别使用独立目录。

**数值序列。** 初始化 `[]Field{{Name: "requests", Aggregate: Sum}, ...}` 后，用 `recorder.Record(timestamp, map[string]float64{...})` 提交完整字段集合。支持 `Mean`（按实际样本数加权）、`Max`、`Min`、`Sum`、`Last`；`Sum` 提交周期增量，`Last` 按提交顺序取值。底层 `Series.Add(Point{Time: ..., Samples: ..., Values: ...})` 可指定权重，`Sum` 不会再次乘以 `Samples`。

`Recorder` 有 120 条队列，队列满时返回错误并在 `Settings.LastError` 累计提示；成功返回表示入队。它自动按分钟汇总和刷新、按小时清理，退出时 `Close` 排空队列并落盘。`Series` 本身不启动协程，由业务调用 `Flush`、`Cleanup`。数值目录默认开启、保留 30 天，可设置 1–365 天；这套默认值不适用于通用 `Store`。

字段名称、顺序和聚合规则的校验值保存在 `.series` 日文件头。打开目录时验证已有日文件，拒绝不匹配的字段定义，不读取额外的字段定义文件。每条分钟记录占 `16 + 8 × 字段数` 字节。

`Series.Query(start, end)` 和 `Recorder.Query(start, end)` 默认将长区间聚合为最多约 1200 点。可传入 `SeriesQueryOptions{MaxPoints: 1440}` 设置目标点数（1–1440，0 使用默认行为），按常用分钟、小时、天间隔向上选择粒度，含两端边界最多返回目标点数加 1。聚合仍读取所选区间的分钟记录，并按字段规则保留加权均值、峰值等信息；查询期间持读锁，可能暂时阻塞该序列写入。通用查询只返回真实记录；系统监控的 `history` 请求可传入 `max_points`（120–720，省略时默认 480），页面对超过 6 小时至 24 小时的范围使用桌面端 300、移动端 120 点预算（含自定义日期），其他范围使用桌面端 720、移动端 240 点预算。监控 service 从首条真实采样开始，按同一时间桶补齐中间及末尾缺失值为 0（`samples=0`）；不补采样前的空白，无记录时返回空列表，含两端边界最多返回目标点数加 1。补零不会改变存储及聚合结果，返回指标保持原始聚合精度。

数值序列默认在 `config.json` 保存开关和保留天数。业务已有配置表时，同时传入 `SeriesOptions.Configuration`（`*SeriesConfiguration`）和 `SaveConfiguration` 保存回调，即可完全接管配置持久化，不再读写本地配置文件。更新先验证参数、刷新待保存的数据，再执行保存回调；回调失败时运行配置不变，成功后即时生效。回调在内部锁下执行，不应再次调用当前序列或记录器。系统监控通过此方式将配置保存到 `cloud_configs` 表，采样和查询使用内存配置，无需每秒读取配置表。

**事件存储。** `events.Append(timestamp, map[string]interface{}{...})` 保留顶层对象及嵌套 JSON，写入 `.events.jsonl`。`QueryContext` 通过 `EventQuery` 按时间、顶层标量字段精确匹配和偏移游标分页；`StatsContext` 通过 `StatsQuery` 提供计数及数值统计。它们扫描所选日期文件，没有内容字段索引。

`Events` 不启动后台协程；`Append` 写入操作系统缓存，业务按需调用 `Sync`、`Cleanup(time.Now())`，退出时 `Close` 同步并关闭。事件目录默认开启、保留 30 天，允许 1–365 天；单条上限 1 MiB、顶层字段上限 128、每页最多 1000 条且序列化数据不超过 8 MiB、统计最多 1000 个分组。完整过期日文件直接删除，边界日额外保留不足一天空间，查询严格过滤过期记录。
