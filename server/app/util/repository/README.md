# repository：GORM 泛型分页仓储

导入路径：`server/app/util/repository`。本包提供 `Repository[T]`，在已有 GORM 条件查询上统一实现页码分页和游标分页。适合后台列表、批量处理数据的连续读取。结果结构来自 [page](../page/README.md)，HTTP 参数处理可结合 [context](../context/README.md)。

## 创建与查询入口

```go
NewRepository[T any](db *database.Database) *Repository[T]

SelectPage(query *gorm.DB, queryPage *page.Query,
    customCursorLastField ...string) (*page.Result[T], error)
```

`T` 是返回的单条记录类型，通常为模型结构体或模型指针。Repository 持有数据库引用但不拥有连接生命周期，不应在每次查询后关闭它。query 必须显式设置 `Model(&Model{})`；仅调用 `Table("...")` 无法提供本包需要的 GORM 字段信息。

queryPage 的 PageSize 必须大于 0；包内不提供默认值或最大值。对 HTTP 输入可使用 context.ValidatePage 的 1～1000 限制，后台调用应自己设置合理值。无效仓储/数据库/query/queryPage 返回 `gorm.ErrInvalidData`；没有 Model、未知排序/定位字段、无效游标等返回具体错误。

## 页码分页示例

```go
package example

import (
    "context"
    "server/app/util/database"
    "server/app/util/page"
    "server/app/util/repository"
)

type Record struct {
    ID int64 `gorm:"primaryKey"`
    Name string
    Enabled bool
}

func ListRecords(ctx context.Context, db *database.Database, number int) (*page.Result[Record], error) {
    if number < 1 { number = 1 }
    repo := repository.NewRepository[Record](db)
    query := db.Db.WithContext(ctx).Model(&Record{}).Where("enabled = ?", true)
    return repo.SelectPage(query, &page.Query{
        Page: number, PageSize: 20,
        OrderByField: "id", OrderByType: "desc",
    })
}
```

Page > 0 使用页码模式，会执行 COUNT 和 LIMIT/OFFSET 两次查询。结果包含 total、page、page_size、list；无结果时 list 为非 nil 的空切片。OFFSET 超过 500000 会报“分页位置过深，请使用游标分页”。COUNT 和列表查询并不自动处于同一快照事务，数据被并发增删时两者可能短暂不一致。

## 游标分页示例

Page ≤ 0 使用游标模式，首批不传游标；后续将响应的 next_cursor_id 和 next_cursor_last_id 原样回传，保持筛选条件、排序字段与方向一致。

```go
package example

import (
    "context"
    "time"
    "server/app/util/database"
    "server/app/util/page"
    "server/app/util/repository"
)

type Event struct {
    ID int64 `gorm:"primaryKey"`
    CreateTime time.Time
    Message string
}

func ReadEvents(ctx context.Context, db *database.Database, cursor, lastID string) (*page.Result[Event], error) {
    query := db.Db.WithContext(ctx).Model(&Event{})
    return repository.NewRepository[Event](db).SelectPage(query, &page.Query{
        Page: 0, PageSize: 100,
        OrderByField: "create_time", OrderByType: "desc",
        CursorId: cursor, CursorLastId: lastID,
    })
}

// ConsumeEvents 的 visit 在查询完成后同步处理每批结果；不自动建立数据库快照。
func ConsumeEvents(ctx context.Context, db *database.Database, visit func([]Event) error) error {
    var cursor, lastID string
    for {
        result, err := ReadEvents(ctx, db, cursor, lastID)
        if err != nil { return err }
        if err := visit(result.List); err != nil { return err }
        if result.HasMore == nil || !*result.HasMore { return nil }
        cursor = *result.NextCursorId
        lastID = *result.NextCursorLastId
    }
}
```

示例要求 events 表已存在，visit 非 nil。查询会多读取一条，判断 has_more，然后截断到 PageSize；不执行总数查询，因此游标结果没有 total/page。只有存在下一批时才输出下一游标，遍历应以 has_more 判断结束。

## 游标字段与稳定排序

排序字段通过 GORM schema 查找，可以使用能被 `LookUpField` 识别的模型字段名或数据库列名，不接受任意 SQL 表达式。OrderByType 忽略大小写，只有 `asc` 表示升序，其他值都按降序处理。

默认使用模型的优先主键作为第二排序字段，解决多个记录排序值相同的问题。也可以传 `SelectPage(query, q, "another_unique_field")` 替换定位字段；只读取第一个可选参数。调用方应确保定位字段唯一且非空，并为主要筛选/排序建立合适索引。

| 排序字段与定位字段 | 首批 | 后续请求 |
| --- | --- | --- |
| 相同，例如都为 id | 两个游标都为空 | 只传 CursorId，CursorLastId 必须为空 |
| 不同，例如 create_time + id | 两个游标都为空 | CursorId 和 CursorLastId 都必须提供 |

游标会按模型字段类型转换，无法转换时返回错误。时间输出为 RFC3339Nano，字节切片转字符串，其他值格式化为字符串；调用方不要自行截断精度或转换时区后再组装游标。

返回类型 T 必须保留排序和定位字段，并有可识别的 GORM 字段映射。如果使用 Select 缩减列，也应实际查出这些列，避免零值产生错误的下一游标。游标模式不适合 nullable 排序键：空值无法稳定生成/比较游标。已经附加复杂 Order、GROUP BY、JOIN 或聚合的查询应检查生成 SQL 和计数语义，不要假定本包会消除原有查询结构的影响。

## 并发和使用边界

SelectPage 会重置并修改 queryPage 的 HasMore、NextCursorId、NextCursorLastId；每次请求使用独立的 `*page.Query`，不要并发共享。基础数据库连接可以共享，query 则每次重新从 Db 构造。

游标减少深页 OFFSET 开销，但不提供跨页快照：排序值被修改或新数据插入时仍可能影响遍历范围。导出、备份等需要一致性时，由业务层选择事务或固定时间边界。不要把用户提供的筛选字符串直接拼进 SQL，示例使用 GORM 占位参数。

源码：[repository.go](repository.go)、[page.go](page.go)。相关工具：[database](../database/README.md)、[page](../page/README.md)、[context](../context/README.md)。
