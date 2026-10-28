# cmd：带日志和取消控制的脚本执行器

导入路径：`server/app/util/cmd`。本包把脚本文本写入临时目录，通过系统解释器执行，收集 stdout/stderr，并在取消时回收进程树。适合安装任务、计划脚本及运维命令；长期驻留的受管进程使用 [process](../process/README.md)。

## 创建和执行

| API | 说明 |
| --- | --- |
| `NewScriptExec(tempPath string, timeout int, writeLog func(string))` | 设置临时文件父目录、超时秒数和实时行日志回调；返回可调用 Execute 的执行器 |
| `Execute(script string) (stdout, stderr string, err error)` | 使用后台 context 执行，仍受构造函数的 timeout 限制 |
| `ExecuteContext(ctx, script)` | 同时受调用方 context 和执行器 timeout 限制，较早结束者生效 |

超时必须设置为正秒数；0 或负值会立即过期，不能表示无限制。脚本不能为空，最大 2 MiB。每次调用在 `tempPath` 内创建独立的随机子目录，结束后删除，传入的父目录本身不会被删除。空 context 会替换成 `context.Background()`。

## 示例：运行并记录输出

```go
package example

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "sync"
    "server/app/util/cmd"
)

func RunScript(ctx context.Context) error {
    var mu sync.Mutex
    executor := cmd.NewScriptExec(filepath.Join(os.TempDir(), "panel-script-demo"), 30,
        func(line string) {
            // stdout/stderr 分别由协程读取，回调可能同时执行。
            mu.Lock()
            defer mu.Unlock()
            fmt.Println("log:", line)
        })
    script := "printf 'hello\\n'\n"
    if runtime.GOOS == "windows" {
        script = "@echo off\r\necho hello\r\n"
    }
    stdout, stderr, err := executor.ExecuteContext(ctx, script)
    if err != nil {
        return fmt.Errorf("脚本执行失败: %w; stderr=%s", err, stderr)
    }
    fmt.Print(stdout)
    return nil
}
```

不要把来自用户输入的参数直接拼入脚本。执行器按当前进程身份运行，不提供容器隔离、文件权限沙箱或命令白名单。脚本文件的临时目录不是工作目录设置：子进程继承宿主工作目录和环境，需要切换目录时在可信脚本内显式 `cd`。

## 输出、回调和资源

stdout 和 stderr 分别累积为字符串，同时逐行调用同一个 `writeLog`；回调没有标注输出来源，跨流顺序不保证。读取器去掉原换行后统一补 `\n`，所以结果不是逐字节原始输出。Windows 会尝试 UTF-8、UTF-16 BOM、系统 OEM/ANSI 代码页解码。

日志回调同步运行在两个读取协程中，必须线程安全且尽量短。回调阻塞会拖慢管道消费和进程退出。脚本输出会完整保存在内存中，没有输出大小上限；大量数据应由脚本写文件，stdout 只记录摘要。

## 平台和取消行为

- 非 Windows 使用 `bash`，运行环境必须存在该命令；Unix 使用独立进程组管理脚本及子进程。
- Windows 使用 `cmd.exe` 批处理，并为脚本切换 UTF-8 代码页。它不是 PowerShell 执行器；需要 PowerShell 时由批处理显式启动。
- context 结束后先尝试停止进程树，等待约 5 秒，再强制终止并最多等待约 2 秒。取消返回可能晚于超时点。
- Windows 优先使用 Job Object 回收子进程，部分场景回退为 `taskkill /T`。
- 非零退出会返回错误，通常仍带有已收集的 stdout/stderr。强制停止仍未完成等极端路径可能返回空输出。
- 正常取消可通过 `errors.Is(err, context.Canceled)` 判断；普通超时路径返回中文超时错误，没有统一包装 `context.DeadlineExceeded`，若需区分任务期限可同时检查调用方 `ctx.Err()`。

源码：[执行流程](cmd.go)、[Unix 控制](process_unix.go)、[Windows 控制](process_windows.go)。相关工具：[job](../job/README.md)、[daemon](../daemon/README.md)。
