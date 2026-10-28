# process：系统子进程管理与保活

导入路径：`server/app/util/process`。此包在当前程序内按 ID 管理普通系统进程，提供启动、停止、重启、配置同步、状态快照和退出后的自动重启。进程继承宿主账户权限；它不提供容器隔离，也不负责将宿主程序安装成系统服务。

相关工具：[cmd 一次性命令执行](../cmd/README.md)、[daemon 宿主守护服务](../daemon/README.md)、[container 隔离运行环境](../container/README.md)。

## 创建与配置

`manager := process.NewManager()` 创建空管理器，不启动进程，也不读写持久化配置。应由业务层持有一个长期使用的实例，在宿主退出时调用 `StopAll()`。不要为了每次查询重新创建 Manager。

| Config 字段 | 行为 |
| --- | --- |
| `ID` | 必填，去除首尾空格后作为进程唯一标识。 |
| `Command` | 必填，完整 shell 命令；Unix 使用 `/bin/sh -c`，Windows 使用 `cmd.exe /D /S /C`。 |
| `WorkingDir` | 非空时必须是已经存在的目录；空时继承宿主工作目录。 |
| `Env` | `KEY=VALUE` 列表，在当前环境上覆盖同名变量；Windows 变量名比较不区分大小写。 |
| `AutoRestart` | 为 true 时，进程自行退出后会重新启动，包括退出码为 0 的情况。 |
| `MaxRetries` | 连续重启上限，0 使用 `DefaultMaxRetries`，当前为 3；负值报错。 |
| `RestartDelay` | 重启前等待时间，0 使用 5 秒，负值报错。 |
| `StopTimeout` | 优雅停止等待时间，0 使用 10 秒；之后尝试强制终止，再等待最多 5 秒。 |
| `LogPath` | stdout、stderr 共用的日志文件；父目录自动创建。空值不保存输出。 |

`time.Duration` 在 Go 配置中使用 `time.Second`、`time.Minute` 等常量。若由标准 JSON 直接反序列化配置，数值单位为纳秒，应由业务接口明确转换。

## 生命周期 API

| 方法 | 说明 |
| --- | --- |
| `Start(config) (*Status, error)` | 注册并启动；同 ID 仍处于期望运行状态、有活动进程或等待重启时拒绝覆盖，已停止记录可显式重新启动。 |
| `Stop(id) error` | 停止并抑制后续自动重启；即使 ID 尚未同步，也会保存停止意图。 |
| `Restart(id) (*Status, error)` | 使用已保存配置显式重新启动，并重置连续重试状态；未知 ID 报错。 |
| `Sync(configs) error` | 将整个进程集合与传入的完整配置集合对齐，不是增量添加。 |
| `StopAll() error` | 停止全部进程并取消重启计时器；各进程并发停止，错误用 `errors.Join` 汇总。 |
| `Status(id) (*Status, error)` | 返回状态副本；未知 ID 返回错误。 |

管理操作有互斥保护，读取状态使用读锁。不要通过反射或其他方式修改内部状态。

## 示例：托管一个长期工作进程

下例是一段可放入业务包的完整函数。`command` 必须由受信任的配置提供，`workDir` 已经存在。函数在上下文取消后停止子进程并返回清理结果。

```go
package example

import (
    "context"
    "errors"
    "fmt"
    "path/filepath"
    "time"

    "server/app/util/process"
)

func RunWorker(ctx context.Context, workDir, command string) (err error) {
    manager := process.NewManager()
    defer func() { err = errors.Join(err, manager.StopAll()) }()

    status, err := manager.Start(process.Config{
        ID:           "backup-worker",
        Command:      command,
        WorkingDir:   workDir,
        Env:          []string{"APP_ENV=production"},
        AutoRestart:  true,
        MaxRetries:   5,
        RestartDelay: 3 * time.Second,
        StopTimeout: 15 * time.Second,
        LogPath:      filepath.Join(workDir, "logs", "backup-worker.log"),
    })
    if err != nil {
        return err
    }
    fmt.Printf("worker PID=%d, state=%s\n", status.PID, status.State)
    <-ctx.Done()
    return nil
}
```

## 示例：从数据库配置同步

下面的函数接收业务层已经查出的完整配置列表。`Sync` 会停止并移除列表之外的进程，对配置变化的活动进程先停止再启动；配置不变且健康的进程不会重复启动。

```go
package example

import "server/app/util/process"

func Reconcile(manager *process.Manager, rows []process.Config) error {
    if err := manager.Sync(rows); err != nil {
        return err
    }
    return nil
}
```

传入空列表表示清空管理集合。手动 `Stop` 或重试耗尽的记录处于暂停状态，即使后续 `Sync` 仍包含它，也不会自动复活；需要业务层明确调用 `Start` 或 `Restart`。`Sync` 可能部分成功后返回聚合错误，应逐个 `Status` 核对，而不是假设整个集合已回滚。

## 状态与故障处理

状态包括 `starting`、`running`、`stopping`、`stopped`、`restarting`、`failed`。`PID=0` 表示当前没有活动进程，`ExitCode=-1` 表示尚无退出码。使用 `Error`、`RestartCount`、`RetryExhausted` 展示失败原因，不要仅凭 `State` 判断是否需要人工处理。

进程连续运行至少一分钟后再次退出，会清零连续重试计数。重试达到上限后进入 `failed`，停止继续拉起。`Start` 在首轮启动失败时仍返回错误；若启用了自动重启，内部可能已经安排重试，调用方决定保留管理器观察状态还是显式停止。

Unix 尽量使用独立进程组管理子树，Linux 还设置父进程退出信号；Windows 使用 `taskkill /PID ... /T` 处理进程树。被托管的命令应以前台方式运行，不要自行 `nohup`、双重 fork 或另开服务，否则可能绕开进程组的清理范围。普通进程自然退出后也会检查残余进程组。

日志采用 50 MB 轮转、最多 10 个备份、保留 30 天、gzip 压缩，文件权限为 `0600`。这些参数目前由实现固定，不在 Config 中开放。命令通过 shell 解析，外部输入不能直接拼接进入 `Command`；需要严格参数边界的一次性操作优先使用合适的命令执行 API。
