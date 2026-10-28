# archive：文件归档与解压

导入路径：`server/app/util/archive`。本包封装 ZIP、TAR、GZIP 和 TGZ 的本地文件操作，适合文件管理、软件包导入以及备份上传前生成归档。当前文件管理业务使用本包，参见 [文件压缩服务](../../service/file/compress.go)。上传对象存储请继续使用 [oss](../oss/README.md)。

## 创建实例与格式

`archive.New()` 返回无配置、无连接状态的 `*Service`；也可以直接使用 `archive.Compress`。不需要 Close。相互独立的操作可以复用实例，但不要并发读写同一个目标路径。

| 常量 | 值 | 用途 |
| --- | --- | --- |
| `FormatZip` | `zip` | 一个或多个文件/目录的 ZIP 归档 |
| `FormatTar` | `tar` | TAR 打包，不额外压缩 |
| `FormatGzip` | `gz` | 仅支持一个普通文件，不支持目录 |
| `FormatTgz` | `tgz` | TAR 打包后使用 GZIP 压缩 |

`GetFormatByExt(filename) string` 按扩展名识别格式，忽略大小写，支持 `.zip`、`.tar`、`.gz`、`.gzip`、`.tar.gz`、`.tgz`；未知格式返回空字符串。`IsSupportedFormat(format) bool` 只接受表中的格式值，不会转换大小写，也不接受 `tar.gz` 作为 format 参数。

## API 和路径含义

```go
CompressFiles(ctx context.Context, srcPaths []string, destPath, format string,
    callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error

Extract(ctx context.Context, srcPath, destPath string,
    callback func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool) error
```

- `CompressFiles` 的 srcPaths 为本地源路径，destPath 为完整归档文件名；自动创建目标父目录。format 决定写入格式，文件名后缀不会替代 format。
- `Extract` 的 srcPath 为归档文件，destPath 为解压目录；按源文件扩展名选择解压器，不检测文件魔数，因此压缩时应使用匹配的后缀。
- 必须传非 nil 的 ctx；无需取消时使用 `context.Background()`。callback 可为 nil，返回 false 会中止操作。
- 压缩目录时保留目录名。例如源 `/srv/data` 中的 `a.txt` 会保存为 `data/a.txt`。多个源使用各自名称，相同名称可能冲突；调用方应组织好目录结构。
- 压缩目标不能等于源文件，也不能位于源目录内部；检查包含符号链接路径解析，避免把输出文件再次打包进自己。

## 示例：打包目录并报告进度

下面函数不持有文件句柄，返回前内部会关闭文件。调用方传入真实 sourceDir 和归档目标；目标必须放在源目录之外。

```go
package example

import (
    "context"
    "fmt"
    "server/app/util/archive"
)

func PackBackup(ctx context.Context, sourceDir, archivePath string) error {
    service := archive.New()
    return service.CompressFiles(ctx, []string{sourceDir}, archivePath, archive.FormatTgz,
        func(current, total int64, currentFile string, totalFiles, currentIndex int64) bool {
            if total > 0 {
                fmt.Printf("%s: %.1f%% (%d/%d entries)\n", currentFile,
                    float64(current)*100/float64(total), currentIndex, totalFiles)
            }
            return ctx.Err() == nil
        })
}
```

回调表示处理进度，不代表数据已经持久化。ZIP/TAR 对大文件约每处理 1 MiB 回调，并在文件边界回调；目录也计入扫描项数，因此不要把 totalFiles 理解成纯普通文件数量，也不要依赖各格式的 currentIndex 起点统一。回调在执行操作的 goroutine 内同步运行，耗时回调会拖慢操作。若把进度转发给其他 goroutine，应自行同步共享状态。

ZIP/TAR 主要按未压缩文件字节统计，GZIP 解压按压缩输入字节统计；空文件或目录的 total 可能为 0。扫描和实际读写之间源文件可能改变，进度不是快照，也不应作为内容完整性校验。

## 示例：解压到新目录，失败时由调用方清理

```go
package example

import (
    "context"
    "os"
    "server/app/util/archive"
)

// ExtractToTemp 成功后由调用方决定移动目录或执行 os.RemoveAll。
func ExtractToTemp(ctx context.Context, archivePath string) (string, error) {
    dir, err := os.MkdirTemp("", "panel-extract-*")
    if err != nil { return "", err }
    err = archive.Compress.Extract(ctx, archivePath, dir,
        func(_, _ int64, _ string, _, _ int64) bool { return ctx.Err() == nil })
    if err != nil {
        _ = os.RemoveAll(dir)
        return "", err
    }
    return dir, nil
}
```

对于单文件 GZIP，解压文件名优先使用归档头的 Name，没有时根据源文件名推导。想归档整个目录时使用 TGZ，而不是 GZIP。

## 取消、失败与恢复约束

操作返回格式错误、文件系统/压缩流错误或取消错误。回调返回 false 产生“操作被取消”错误，不能统一依赖 `errors.Is(err, context.Canceled)` 判断；调用方可以同时检查自己的 ctx。ZIP/TAR 内部文件复制并不是每次 Read 都直接检查 ctx，大文件场景建议像上面示例一样在回调内检查 ctx；传 nil 回调时取消可能延迟到文件边界。GZIP 的分块读写会检查 ctx。

压缩会创建/截断目标文件，失败会尝试删除未完成的归档，不会恢复原有同名文件。解压会覆盖已有文件，失败或取消可能留下已解出的文件；包内对目标目录的清理不是回滚。需要“成功后发布”的流程时，应先使用独立临时路径，成功校验后再移动到最终路径。

ZIP/TAR 解压检查成员路径和符号链接，拒绝逃出目标目录的路径与链接；可以保存符号链接，但指向归档目录之外的链接不能原样解压恢复。普通 GZIP 解压只有较简单的路径前缀检查，没有同等的父路径符号链接解析；应解压到调用方新建、无链接的临时目录。TAR 并非完整的系统备份恢复器，不应假定它会恢复硬链接、设备节点、所有权、扩展属性等全部元数据。Windows 上创建符号链接仍受系统权限限制。

当前接口没有文件数、总解压大小、单文件大小或压缩比限额。来自外部的归档需要调用方限制资源和磁盘占用。TAR/TGZ 解压会先扫描统计，再重新读取解包，TGZ 会产生额外解压成本。内部延迟执行的压缩 writer.Close 错误没有全部向上传递；严肃备份流程应额外校验归档可读性，并在上传后核验结果。

源码：[archive.go](archive.go)、[progress.go](progress.go)。相关工具：[file](../file/README.md)、[compress 兼容实现](../compress/README.md)、[oss](../oss/README.md)。
