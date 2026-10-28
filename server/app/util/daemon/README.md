# daemon：宿主服务启动、停止与自启动

导入路径：`server/app/util/daemon`。此包管理面板或其他长期服务本身，包括前台运行、后台启动、本机控制端点、平滑退出、外部重载、服务内部重启，以及系统自启动安装。

它与 [process](../process/README.md) 的分工不同：`daemon` 管理当前服务程序的生命周期，`process.Manager` 管理服务内部托管的多个业务子进程。本机控制通过 [ipc](../ipc/README.md) 完成，不依赖 PID 文件。

## 实例化和生命周期

```go
func NewDaemon(controlPath, logFileName string, service func(<-chan struct{})) (*Daemon, error)
```

三个参数均必填。`controlPath` 是通信目录；`logFileName` 是后台启动时 stdout/stderr 的追加日志路径。相对日志路径相对于配置的工作目录，日志父目录应提前创建。`service` 接收只读停止通道：收到关闭通知后，业务必须停止监听、取消任务、等待资源回收，然后返回。

| 方法 | 使用场景与结果 |
| --- | --- |
| `Run() error` | 前台运行；先占用独占控制端点，再调用 Service；Service 返回后释放端点。 |
| `Start() error` | 从外部启动已安装的系统服务或独立后台子进程；已运行时报错。 |
| `Stop() error` | 从外部发出停止请求并等待清理；未运行时返回 nil。 |
| `Running() (bool, error)` | 查询控制端点；初始化和退出清理期间仍视为运行中。 |
| `Reload() error` | 外部执行停止再启动，不是应用配置热更新。 |
| `Restart() error` | 仅在当前正在运行的服务进程内部调用，安排外部控制进程或系统服务重启。 |
| `SetChildArgs(args...) *Daemon` | 设置普通后台子进程的完整启动参数，第一个参数是可执行文件路径。 |
| `SetAutoStartOptions(options) *Daemon` | 设置系统自启动和启动路径/参数，返回自身便于链式调用。 |
| `InstallAutoStart() error` | 安装系统自启动配置；是否立即加载受平台实现影响，当前服务启动另用 Start。 |
| `UninstallAutoStart() error` | 删除系统自启动配置。 |
| `SyncAutoStart() error` | 同步已安装 Linux systemd unit，保留现有自启开关；其他情况不处理。 |
| `Uninstall() error` | 先停止服务，再移除自启动。 |
| `IsChildProcess() bool` | 识别普通后台启动时的重生标记，避免子进程再次启动子进程。 |

`StopTimeout` 默认 2 分钟。普通后台 `Start` 会等待控制端点出现，最长约 10 秒；它证明控制端点已建立，不等同于业务所有依赖已完成初始化。业务就绪状态应单独维护。

## 示例：一个支持 run/start/stop 的程序

以下示例展示完整入口。它会在当前工作目录下创建 `data/demo-daemon`；Service 中的注释位置应接入真实业务初始化及清理。系统自启动使用明确的 `run` 参数，避免把安装命令自身作为启动参数保存。

```go
package main

import (
    "log"
    "os"
    "path/filepath"

    "server/app/util/daemon"
)

func main() {
    workDir, err := os.Getwd()
    if err != nil {
        log.Fatal(err)
    }
    dataDir := filepath.Join(workDir, "data", "demo-daemon")
    if err = os.MkdirAll(dataDir, 0700); err != nil {
        log.Fatal(err)
    }
    service, err := daemon.NewDaemon(
        filepath.Join(dataDir, "control"),
        filepath.Join(dataDir, "service.log"),
        func(stop <-chan struct{}) {
            // 初始化 HTTP、任务调度和数据库等资源。
            <-stop
            // 取消任务并等待完成，然后关闭 HTTP、数据库等资源。
        },
    )
    if err != nil {
        log.Fatal(err)
    }
    service.SetAutoStartOptions(daemon.AutoStartOptions{
        Name:             "panel-demo",
        WorkingDirectory: workDir,
        Arguments:        []string{"run"},
    })

    action := "run"
    if len(os.Args) > 1 {
        action = os.Args[1]
    }
    if service.IsChildProcess() {
        action = "run"
    }
    switch action {
    case "run":
        err = service.Run()
    case "start":
        err = service.Start()
    case "stop":
        err = service.Stop()
    case "reload":
        err = service.Reload()
    case "install":
        err = service.InstallAutoStart()
    case "uninstall":
        err = service.Uninstall()
    default:
        log.Fatalf("unknown action: %s", action)
    }
    if err != nil {
        log.Fatal(err)
    }
}
```

正式部署前先编译二进制再安装自启动，不能把 `go run` 的临时可执行文件保存为服务路径。

## 示例：停止宿主时一并回收业务子进程

下面的完整函数构造可传给 NewDaemon 的 Service 回调，说明如何把停止通道连接到 process.Manager。`workDir` 必须存在，`command` 是受信任的前台命令。初始化失败会记录错误并返回；生产入口可另外记录服务失败状态。

```go
package example

import (
    "log"
    "path/filepath"

    "server/app/util/process"
)

func WorkerService(workDir, command string) func(<-chan struct{}) {
    return func(stop <-chan struct{}) {
        manager := process.NewManager()
        defer func() {
            if err := manager.StopAll(); err != nil {
                log.Printf("停止工作进程失败: %v", err)
            }
        }()
        if _, err := manager.Start(process.Config{
            ID: "worker",
            Command: command,
            WorkingDir: workDir,
            AutoRestart: true,
            LogPath: filepath.Join(workDir, "logs", "worker.log"),
        }); err != nil {
            log.Printf("启动工作进程失败: %v", err)
            return
        }
        <-stop
    }
}
```

将 `WorkerService(workDir, command)` 作为 NewDaemon 的第三个参数即可；Daemon.StopTimeout 应覆盖所有子进程和其他资源的最长清理时间。Service 没有 error 返回值，单纯从回调返回不会自动让主程序以非零退出码结束。

## AutoStartOptions 参数

| 字段 | 规则 |
| --- | --- |
| `Name` | 管理系统自启动时必填；仅允许实现接受的字母、数字、点、下划线和短横线。不同安装实例应使用不同名称。 |
| `Executable` | 默认当前可执行文件；归一化为绝对路径并检查。 |
| `WorkingDirectory` | 默认当前目录，必须存在。 |
| `Arguments` | nil 沿用当前进程参数；显式空切片表示无参数。通常应指定 `[]string{"run"}` 等前台入口参数。 |
| `KeepChildProcesses` | Linux systemd 使用 `KillMode=process`，让应用自行负责子进程回收；默认 `KillMode=mixed`。 |

这些配置应在开始运行前设置。不要在并发请求中修改公开配置字段或替换 `Service`。

## 各系统的实际接入

- **Linux**：写入 `/etc/systemd/system/<Name>.service`，执行 `daemon-reload` 并启用；通常需要 root。只有发现 systemd 运行环境和已安装 unit 才由 `Start/Stop` 接管该服务，否则使用普通后台进程。unit 的 `Restart=on-failure`、`RestartSec=5` 与应用内部 process 管理器的重试策略是两套机制。
- **macOS**：写入当前账户 `~/Library/LaunchAgents/<Name>.plist`，配置 `RunAtLoad` 和 `KeepAlive`；日志写入 `~/Library/Logs/<Name>.log`。安装方法写配置，`Start` 再尝试对应 launchctl domain 的 bootstrap/kickstart。`Stop` 可卸载已加载任务但保留 plist，避免 KeepAlive 立即拉起。
- **Windows**：创建启动计划任务和工作目录内的 `.<Name>-startup.cmd`，依赖 `schtasks.exe`；控制端点采用命名管道。系统任务安装及控制需相应权限。

普通后台方式继承当前环境，以当前账户运行。后台日志是追加文件，本包本身不为该文件提供轮转；若业务需要轮转日志，使用 [log](../log/README.md) 或专门日志组件。

## 平滑退出与重启注意事项

`Run` 的 Service 是同步回调，不应初始化后立即返回，否则控制端点也会被释放。控制端点会一直持有到业务清理完成，因此服务处于停止中时不能启动第二个实例。

`Restart` 用于服务内部请求重启；外部 CLI 应调用 `Reload`。内部重启调用成功表示已安排控制操作，不表示新进程已经就绪。不要在原进程中同步等待自身退出后才能完成的操作。

`Running` 的权限/认证错误必须向上传递，不要忽略错误后当成 false。控制路径、运行账户和自启动名称应保持一致，避免新旧目录中的服务相互混淆。`KeepChildProcesses=true` 时必须由 Service 清理子进程，否则它们可能继续存活。
