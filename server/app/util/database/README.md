# database：SQLite 连接、事务与结构同步

导入路径：`server/app/util/database`。本包以 GORM 为基础，提供 SQLite 连接、统一事务、切片分批执行和声明式表结构同步。`Database.Db` 为 `*gorm.DB`，查询、创建、更新等仍使用 GORM；当前包只提供 `NewSqlite` 构造函数，没有 MySQL/PostgreSQL 连接入口。

## 创建连接和关闭资源

`NewSqlite(file string) *Database` 返回包含 `Db` 和 `DbError` 的对象，**没有第二个 error 返回值**。必须先检查 `DbError` 再使用 Db。连接失败目前统一返回 `sql connect error`，不会保留底层连接错误细节。默认 GORM 日志级别为 Silent。

```go
package example

import (
    "os"
    "path/filepath"
    "server/app/util/database"
)

func OpenStore(filename string) (*database.Database, error) {
    if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil { return nil, err }
    db := database.NewSqlite(filename)
    if db.DbError != nil { return nil, db.DbError }
    sqlDB, err := db.Db.DB()
    if err != nil { _ = db.Close(); return nil, err }
    sqlDB.SetMaxOpenConns(1)
    sqlDB.SetMaxIdleConns(1)
    return db, nil
}
```

业务生命周期结束时调用 `db.Close()` 并处理返回的错误。Close 关闭底层 SQL 连接池；对 nil Database 或 nil Db 返回 nil。不要在每次请求结束时关闭全局共享连接。连接池可以共享，已经附加条件的 GORM 查询链和事务应按操作单独构造。

支持使用 `":memory:"` 创建内存 SQLite。内存库通常与单个物理连接绑定，因此示例把连接池限制为一个连接；不要在同一事务中绕过 tx 去使用该单连接池，否则可能等待自己释放连接。

## 事务和分批写入

`Transaction(func(tx *gorm.DB) error) error` 委托 GORM 执行事务：回调 nil 提交，返回错误回滚；无有效 Db 返回 `gorm.ErrInvalidDB`。事务内部的所有数据库调用必须使用回调的 tx。

`BatchExecute(slice interface{}, batchSize int, fn func(batch interface{}) error) error` 按原顺序同步切分切片，回调类型保持不变，如传入 `[]Record`，每批也是 `[]Record`。batchSize ≤ 0 时使用 1000；空切片或非切片直接返回 nil，不会报告输入类型错误。回调首次返回错误就停止，**本身不建立事务**。切片与原始数据共享底层数组；非空切片需要非 nil 回调。

```go
package example

import (
    "server/app/util/database"
    "gorm.io/gorm"
)

type Note struct {
    ID int64 `gorm:"primaryKey"`
    Title string
}

func SaveNotes(db *database.Database, notes []Note) error {
    return db.Transaction(func(tx *gorm.DB) error {
        return db.BatchExecute(notes, 200, func(batch interface{}) error {
            items := batch.([]Note) // 输入固定为 []Note，因此类型一致。
            return tx.Create(&items).Error
        })
    })
}
```

示例要求 notes 表已经初始化。分批可以限制单次 SQL 的参数量，但上面的事务仍包含全部批次，不会每批提交；如果需要每批独立提交，应在业务层明确失败后的恢复策略。

## 使用 SQL 声明 SQLite 表结构

`SyncSqlite(schemaSql string) error` 读取 schemaSql 声明的表及显式索引，并同步到当前库。适合应用启动阶段管理自身表结构；不是版本迁移框架，也不维护迁移历史。

```go
package example

import "server/app/util/database"

const notesSchema = `
CREATE TABLE notes (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL DEFAULT ''
);
CREATE INDEX idx_notes_title ON notes(title);
`

func InitializeStore(db *database.Database) error {
    return db.SyncSqlite(notesSchema)
}
```

传入完整的期望声明，不要传只有 `ALTER TABLE` 的增量 SQL。内部先在临时内存库执行 SQL 获取期望结构，再在实际库中进行事务同步。空 SQL、没有声明表、非法 SQL、创建/复制失败等均返回错误。

同步规则：

1. 期望中缺少实际表：创建；结构一致：保留。
2. 结构发生变化：创建临时新表，将两边同名的列复制过去，再替换原表。删除的列及其数据不再保留；新增非空列需要兼容旧数据的默认值，否则复制可能失败。
3. 期望表的显式索引会对齐 SQL 声明，未声明或定义变化的旧索引会删除，再创建需要的索引。
4. schemaSql 未声明的数据表保留；它们的索引不作为期望表索引清理。
5. 在实际连接上按需要暂时关闭外键检查，结束时恢复；事务末执行外键完整性检查，违规会失败。

结构比较会规范化空白、标识符大小写等文本差异。发生表重建时实现要求建表 SQL 能匹配 `CREATE TABLE 表名` 前缀，因此采用示例中的普通未加引号表名和常规建表格式；特殊引用或建表语法即使首次可建表，也可能在重建时报告“无法识别…建表语句”。

只同步表和显式索引，不同步视图、触发器等对象。表重建可能影响附属触发器；不要把该 API 当作复杂数据库的无损迁移器。结构同步可能删除列、重建表、扫描复制整表，应在启动维护阶段执行，提前备份，并避免与业务写入或另一次结构同步竞争。

源码：[database.go](database.go)、[sqlite.go](sqlite.go)。相关工具：[repository](../repository/README.md)、[page](../page/README.md)、[file](../file/README.md)。
