# file：本地文件管理与文本分页编辑

导入路径：`server/app/util/file`。本包提供目录浏览、文件元数据、创建/删除/复制/移动，以及带版本校验的 UTF-8 文本读取和编辑。适合面板文件管理；归档使用 [archive](../archive/README.md)，对象存储使用 [oss](../oss/README.md)。

## 创建 Disk 与路径规则

`file.NewDisk(root string) *Disk` 设置相对路径基准；root 为空时使用 `./`。Disk 没有需要 Close 的资源。

**root 不是访问沙箱。** 绝对路径会直接使用，`../` 也可能越出 root；符号链接还可能指向其他位置。向 HTTP 用户开放操作时，调用方必须独立落实路径访问权限，不能用 NewDisk(root) 替代授权检查。

`Path(name string, hasFile bool, full bool) string` 返回计算后的路径：hasFile 为 true 返回该路径，为 false 返回其父目录；full 为 true 时尝试转为绝对路径。它只做路径计算，不检查目标是否存在。

## 创建、权限与删除 API

| 方法 | 行为与注意事项 |
| --- | --- |
| `CreateDir(name) error` | 递归创建目录，初始权限 0755，最终受 umask 影响 |
| `CreateFile(name) error` | 使用排他创建，已有文件报错；不会自动创建父目录 |
| `AutoCreate(name) error` | 名称以 `/` 或 `\` 结尾时创建目录，否则补齐父目录并创建缺失文件 |
| `OpenFile(name, flag, perm int) (*os.File, error)` | 直接打开；不会自动创建父目录，调用方负责 Close |
| `SetFileMode(name, mode uint32) error` | 调用 Chmod；权限位使用八进制，如 0600 |
| `GetFileMode(name) uint32` | 返回权限位；出错时返回 0，无法区分错误和真正的 0000 权限 |
| `IsFile(name) / IsDir(name) bool` | 使用 Stat 判断，跟随符号链接；失败时返回 false |
| `Exists(name) bool` | 使用 Lstat，只有明确不存在时为 false；权限等其他错误也可能返回 true |
| `Delete(name) error` | 删除文件或递归删除目录，请先确认调用方的操作范围 |

```go
package example

import (
    "os"
    "server/app/util/file"
)

func WriteConfig(root string) error {
    disk := file.NewDisk(root)
    if err := disk.CreateDir("config"); err != nil { return err }
    output, err := disk.OpenFile("config/app.conf", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
    if err != nil { return err }
    _, writeErr := output.WriteString("enabled=true\n")
    closeErr := output.Close()
    if writeErr != nil { return writeErr }
    return closeErr
}
```

## 列表、树与元数据

`FileInfo(name) (*File, error)` 返回名称、字节大小、权限位 Mode、IsDir、Unix 秒级修改时间 UpdateTime。目录 Size 取决于文件系统的元数据，不应当作目录内容总大小，也不保证为 0。Child 在目录树中保存子项。

- `FileList(dir) ([]*File, error)`：只读取一层；尽管源码旧注释出现“递归”，实际不递归。
- `FileTree(dir) ([]*File, error)`：递归填充 Child；子目录读取失败时实现会忽略对应子错误，因此结果可能不是完整目录清单。
- `FileListWithPage(dir, page, pageSize, orderByField, orderByType, keywords...) ([]*File, int64, error)`：单层目录分页，页码从 1 开始，page/pageSize 必须为正数。支持 name、size、update_time 排序，方向为 desc 时降序，其他值升序。只使用第一个关键词，去空白后按不区分大小写的名称包含匹配。
- `Count(name) (totalSize, fileCount, dirCount int64, err error)`：递归统计；目录计数包含传入的根目录，符号链接计为非目录条目，不递归跟随到目标目录。

```go
package example

import (
    "fmt"
    "server/app/util/file"
)

func ListLogs(root string) error {
    disk := file.NewDisk(root)
    list, total, err := disk.FileListWithPage("logs", 1, 50, "update_time", "desc", ".log")
    if err != nil { return err }
    fmt.Printf("matching entries: %d\n", total)
    for _, item := range list {
        fmt.Printf("%s size=%d mode=%04o dir=%t\n", item.Name, item.Size, item.Mode, item.IsDir)
    }
    return nil
}
```

分页会读取整个目录并在内存中过滤、排序后截取；不会减少海量目录的扫描成本。获取单项元数据失败的条目会被跳过，total 是成功读取且匹配过滤条件的项数。未知排序字段不会报错，调用方应使用固定白名单。

## 复制、移动与重命名

| 方法 | 目标参数含义 |
| --- | --- |
| `Rename(oldName, newName) error` | newName 为完整新路径，可同时改变文件名与所在目录 |
| `Copy(src, destDir) error` | destDir 是目录，最终目标为 destDir + 源 basename |
| `Move(src, destDir) error` | 同上，优先 rename，无法 rename 时复制后删除源 |
| `CopyWithProcess(ctx, src, destDir, callback) error` | 带取消与复制进度，方法名称确实是 Process |
| `MoveWithProcess(ctx, src, destDir, callback) error` | 带取消与移动进度 |

目标同名路径会被清理后重建，不提供自动改名或“存在则失败”策略；覆盖前应由业务决定是否允许。Copy/Move 会检查目标与源相同或位于源目录内部的情况。Rename 不应传入相同源/目标。跨文件系统移动退化为复制后删除，不是原子操作；失败时不能假定源和目标同时恢复到原状态。

```go
package example

import (
    "context"
    "fmt"
    "server/app/util/file"
)

func CopyBackup(ctx context.Context, root string) error {
    disk := file.NewDisk(root)
    return disk.CopyWithProcess(ctx, "data", "backup",
        func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool {
            fmt.Printf("%s: %d/%d bytes; %d/%d files\n", currentFile,
                current, total, currentIndex, totalFiles)
            return ctx.Err() == nil
        })
}
```

此例最终生成 `root/backup/data`；目录可事先不存在。ctx 必须非 nil，callback 可 nil。回调同步执行，false 会取消复制；操作按块检查 ctx。带进度方法失败时会尝试清理部分目标，但不会恢复已覆盖的旧目标；移动中已经发生的源删除也不构成可回滚事务。普通文件内容流式复制，不把整个文件加载进内存；符号链接单独复制为链接。

## UTF-8 文本分页读取

```go
GetFileContent(name string, cursor int64, pageSize int, version string)
    (content string, nextCursor int64, eof bool, newVersion string, err error)
```

cursor 是**字节偏移**，不是行号；首批传 0 和空 version。pageSize 是希望读取的行数，必须大于 0；cursor 不能为负。后续批次必须传上次返回的 nextCursor 和 newVersion。

每次最多返回约 8 MiB，超长行可能分批截取；nextCursor 用于准确续读，不要用字符串字符数重算。内容必须是有效 UTF-8；二进制文件、无效 UTF-8 或读取过程中版本变化会报错。eof 表示已到达末尾，与“还有下一页”的语义相反。

```go
package example

import (
    "io"
    "server/app/util/file"
)

func ReadText(root, name string, output io.Writer) error {
    disk := file.NewDisk(root)
    var cursor int64
    var version string
    for {
        content, next, eof, nextVersion, err := disk.GetFileContent(name, cursor, 200, version)
        if err != nil { return err }
        if _, err := io.WriteString(output, content); err != nil { return err }
        if eof { return nil }
        cursor, version = next, nextVersion
    }
}
```

version 由文件大小和纳秒修改时间生成，并不是内容哈希。文件被其他进程修改后，应重新读取而非强行复用旧游标。读取时使用共享文件锁并再次检查文件状态；不遵守锁协议的外部程序仍可能并发写入。

## 文本更新和版本冲突

```go
UpdateFileContent(name, content string, cursor, nextCursor *int64, version string)
    (newCursor int64, eof bool, newVersion string, err error)
```

文件必须已存在。两个游标都 nil 表示整体替换；两个都非 nil 表示将字节区间 `[cursor, nextCursor)` 替换为 content，不能只传一个游标。局部替换要求非空 version；整体替换虽允许空版本，为避免覆盖他人修改也建议传读取获得的版本。

```go
package example

import "server/app/util/file"

func ReplaceFirstLines(root, name, replacement string) error {
    disk := file.NewDisk(root)
    _, next, _, version, err := disk.GetFileContent(name, 0, 10, "")
    if err != nil { return err }
    start := int64(0)
    _, _, _, err = disk.UpdateFileContent(name, replacement, &start, &next, version)
    return err
}
```

新游标是替换区间起点加 `len(content)` 字节，eof 表示原替换范围是否触及文件尾部。更新会解析符号链接、加独占锁、备份原文件后原地写入和截断，尽量保留原文件 inode/权限/硬链接关系；写入失败时尝试恢复备份。这需要额外磁盘空间和整文件备份 I/O，不是内存级字符串替换，也不是原子 rename 发布。写入 content 不会预先校验 UTF-8；请保证内容编码正确，否则之后的文本读取可能失败。

文件锁在 Linux/macOS/BSD 使用 flock，Windows 使用 LockFileEx；其他构建平台的锁实现为空操作。对同一文件的多个业务操作仍需按业务需求协调，不应把 Disk 当作全局事务或目录锁。源码：[file.go](file.go)、[返回类型](entities.go)。相关工具：[archive](../archive/README.md)、[oss](../oss/README.md)、[page](../page/README.md)。
