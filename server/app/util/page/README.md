# page：统一分页参数与结果

导入路径：`server/app/util/page`。本包只负责描述分页请求、区分页码/游标模式、构造一致的 JSON 响应；它不执行 SQL、排序或切片分页。

## Query 字段

| 字段 | 用途 |
| --- | --- |
| `Page` | 大于 0 使用页码分页；小于等于 0 使用游标分页 |
| `PageSize` | 每页数量，本包不校验或自动截断 |
| `OrderByField`、`OrderByType` | 排序字段和方向，由查询层解释 |
| `CursorId` | 当前页的排序值游标，使用字符串保留大整数或时间精度 |
| `CursorLastId` | 非唯一排序字段下的定位字段值，通常为主键 |
| `HasMore` | 查询层写入是否存在下一页 |
| `NextCursorId`、`NextCursorLastId` | 查询层写入下一页游标 |

`(*Query).IsCursor()` 以 `Page <= 0` 判断游标模式；nil 指针调用返回 false。每次查询使用独立的 Query，避免多个并发请求修改同一对象。

## Result JSON 约定

`New[T](query *Query, total int64, list []T) *Result[T]` 根据模式构造结果。`list == nil` 会被转换为空数组，而不是 JSON null。

| 模式 | 返回字段 |
| --- | --- |
| 页码分页 | `total`、`page`、`page_size`、`list` |
| 游标分页 | `page_size`、`list`、`has_more`；有下一页时才有 `next_cursor_id`，定位游标非空时才有 `next_cursor_last_id` |
| query 为 nil | 按页码 1、每页 20 构造结果，total/list 使用传入值 |

游标模式不会输出 `total` 或 `page`，传入的 total 被忽略。该包不会根据 list 长度猜测下一页，应由查询层多取一条记录或按实际协议设置 `HasMore`。

## 示例：构造两种返回值

```go
package example

import (
    "encoding/json"
    "fmt"
    "server/app/util/page"
)

type Row struct {
    ID int64 `json:"id"`
    Name string `json:"name"`
}

func ShowPages() error {
    rows := []Row{{ID: 101, Name: "backup"}}
    numbered := page.New(&page.Query{Page: 1, PageSize: 20}, 42, rows)
    cursor := page.New(&page.Query{
        Page: 0, PageSize: 20, HasMore: true,
        NextCursorId: "101",
    }, 0, rows)
    for _, result := range []*page.Result[Row]{numbered, cursor} {
        data, err := json.Marshal(result)
        if err != nil { return err }
        fmt.Println(string(data))
    }
    return nil
}
```

第一种结构包含 `total:42`；第二种结构包含 `has_more:true` 和 `next_cursor_id:"101"`。不要把某页的 `list` 再交给本包期待它自动截取，因为 `New` 原样保留数据。

## 与其他工具配合

HTTP 请求通过 [context.ValidatePage](../context/README.md) 规范化页大小和排序白名单；GORM 查询通过 [repository.SelectPage](../repository/README.md) 获取结果。游标前后页必须保持过滤条件、排序字段和方向一致；API 客户端应把下一页游标原样传回，不转换为 JavaScript number。

源码：[entities.go](entities.go)、[page.go](page.go)、[query.go](query.go)。
