# compress：历史压缩工具

导入路径：`server/app/util/compress`。本包保留 ZIP、TAR、GZIP、TGZ 的早期文件压缩实现。当前文件管理业务使用 [archive](../archive/README.md)，新增归档、备份功能建议参照 archive 的接口和示例。本包没有在代码中声明 deprecated，但其目录层级、取消和符号链接行为与 archive 不同，维护已有调用时需按实际行为处理。

## 对外 API

使用 `compress.New()` 创建 `*Service`，或使用全局 `compress.Compress`。Service 无连接或可关闭资源。

| API | 说明 |
| --- | --- |
| `GetFormatByExt(filename string) string` | 从扩展名识别格式；未知返回 `""` |
| `IsSupportedFormat(format string) bool` | 检查精确的格式常量值 |
| `CompressFiles(ctx, srcPaths, destPath, format, callback) error` | 源文件/目录列表写入完整目标文件名 |
| `Extract(ctx, srcPath, destPath, callback) error` | 按归档扩展名解压到目标目录 |

格式常量为 `FormatZip = "zip"`、`FormatTar = "tar"`、`FormatGzip = "gz"`、`FormatTgz = "tgz"`。扩展名匹配忽略大小写，接受 `.zip`、`.tar`、`.gz`、`.gzip`、`.tar.gz`、`.tgz`；format 参数则只接受常量值，不接受 `"tar.gz"` 或大写 `"ZIP"`。

ctx 类型是 `context.Context`，不能为 nil。srcPaths 类型为 `[]string`；其余路径和 format 为 string。回调完整签名为：

```go
func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool
```

回调可为 nil，返回 false 表示中止；错误文本为“操作被取消”，不保证是 `context.Canceled`。current/total 为处理字节，totalFiles/currentIndex 为遍历进度；目录也可能计入条目数。回调同步执行，不要在回调里进行长时间阻塞工作。

## 示例：单文件 ZIP 压缩与解压

此例将源文件和归档放在不同目录层级，使用新的解压目录。函数返回后临时目录会清理；实际业务可以改为在成功后发布文件。

```go
package example

import (
    "context"
    "os"
    "path/filepath"
    "server/app/util/compress"
)

func LegacyRoundTrip(ctx context.Context) error {
    root, err := os.MkdirTemp("", "panel-compress-*")
    if err != nil { return err }
    defer os.RemoveAll(root)

    source := filepath.Join(root, "note.txt")
    if err := os.WriteFile(source, []byte("hello\n"), 0600); err != nil { return err }
    target := filepath.Join(root, "note.zip")
    service := compress.New()
    if err := service.CompressFiles(ctx, []string{source}, target, compress.FormatZip, nil); err != nil {
        return err
    }
    return service.Extract(ctx, target, filepath.Join(root, "extracted"), nil)
}
```

GZIP 只支持单个文件；归档目录或多个文件使用 ZIP/TAR/TGZ。目标父目录会自动创建，路径使用当前进程的工作目录解析。格式由压缩参数决定，而解压根据扩展名决定，二者要一致。

## 维护已有调用时的具体限制

- 本包目录打包路径会重复拼接源目录名，例如目录 `data` 内的 `a.txt` 可能归档为 `data/data/a.txt`。替换成 archive 前，应检查旧归档和消费者是否依赖这一层级。
- 没有 archive 中的目标路径检查。目标文件必须由调用方放在源目录之外，也不能覆盖源文件，否则可能把输出再次读入或破坏输入。
- ZIP/TAR 回调主要发生在文件边界，复制大文件期间取消响应有限；普通 GZIP 分支不调用进度回调，读写期间也不持续检查 ctx。不能仅凭设置超时就假定大文件会立即中断。
- 本包没有完善的符号链接保存/恢复语义；不要用于包含链接或需要精确保留文件类型的备份。解压使用较简单的路径检查，不能等同于 archive 的符号链接解析保护；需要处理外部归档时使用 archive 并设置业务资源限制。
- 压缩使用创建/截断方式写目标，解压可覆盖已有文件；失败时可能遗留部分文件，不能假定会回滚。使用独立临时目录，并在调用方清理失败产物。
- 没有解压大小、数量或压缩比限额。整个操作是同步的，本包不会自动启动后台任务或限制并发；相同目标路径不得并发写入。

文件句柄由操作内部关闭。压缩 writer 的延迟关闭错误没有全部返回，因此 nil 错误并不替代归档校验。源码：[compress.go](compress.go)。相关工具：[archive](../archive/README.md)、[file](../file/README.md)、[job](../job/README.md)。
