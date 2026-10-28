# container：镜像与容器运行管理

导入路径：`server/app/util/container`。本包基于 libcontainer 管理本地 Docker save 镜像、容器实例、交互终端、日志和资源状态。业务层负责用户授权、软件安装记录和任务调度；本包负责运行时及本地持久化。

## 运行条件与边界

- 实际启动、停止、执行命令和终端功能需要 Linux、root 身份、cgo、libseccomp，以及可用的 User Namespace、cgroup、挂载和 seccomp 能力。宿主内核及运行环境必须支持并允许 `openat2`，否则初始化失败；使用宿主目录挂载还需要 `open_tree`、`move_mount`。Linux 未启用 cgo 时，初始化安全运行环境会直接返回错误；macOS/Windows 不能运行 Linux 容器。
- 使用项目的 Linux 发布构建。自行构建完整服务需参考[项目构建说明](../../../../README.md)及仓库 Dockerfile；仅能编译包不代表宿主机具备容器运行条件。
- 容器运行时会重新执行当前二进制的 `init` 命令。项目已在 [container_init_linux.go](../../command/container_init_linux.go) 调用 `container.Init()`；独立集成时必须保留同等入口，且它仅在 `linux && cgo` 下存在。不要在普通服务初始化阶段直接调用 `Init()`。
- **共享宿主机网络**：实现没有创建 `NEWNET`，没有 Docker daemon、端口映射或虚拟网络 API。镜像 `ExposedPorts` 仅是元数据，容器程序监听的端口直接占用宿主机端口。
- 使用固定可执行文件启动，避免 `go run` 的临时二进制参与容器重入。数据根目录必须长期稳定，同一目录应由同一个管理器拥有。

## 初始化与目录

```go
package example

import "server/app/util/container"

func OpenContainers(dataRoot string) (*container.Manager, error) {
    return container.NewManager(dataRoot, container.ManagerOptions{
        DefaultPidsLimit: 512,
    })
}
```

`NewManager(root, options...) (*Manager, error)` 会校验数据目录、创建目录、加载元数据、核对运行状态并恢复实例，可能产生磁盘和进程操作。需要仅接管或清理而不自动拉起实例时，传 `DisableAutoRestart: true`。不要按 HTTP 请求重复初始化。

```text
dataRoot/
├── images/       # 镜像元数据与展开的根文件系统
├── instances/    # 实例配置、日志、overlay 写层
├── runtime/      # libcontainer 运行状态
├── volumes/      # 默认允许的持久化挂载根目录
└── ...           # 安全映射元数据及内部挂载辅助目录
```

管理器拒绝文件系统根目录及不安全的路径、属主和符号链接。Linux 下还会检查和收紧所需目录权限。不要绕过管理器直接修改 `instance.json`、镜像 rootfs 或运行状态文件。

### ManagerOptions

零值通常应用默认值；`DefaultPidsLimit: -1` 明确表示不限制进程数。

| 配置 | 默认 / 含义 |
| --- | --- |
| `DisableAutoRestart` | `false`；只影响本次管理器接管，不改已保存策略 |
| `ContainerLogMaxSize` | 单份日志 10 MiB |
| `RuntimeLogInterval` | 5 秒，日志大小检查间隔 |
| `RuntimePollInterval` / `RuntimeFailureLimit` | 1 秒 / 连续 3 次状态查询失败 |
| `RuntimeCleanupDelay` / `AutoRestartDelay` | 2 秒 / 1 秒 |
| `StopGracePeriod` / `ForceStopPeriod` | 优雅退出 10 秒 / 强制清理 5 秒 |
| `TerminalConsoleTimeout` / `TerminalCleanupTimeout` | 各 5 秒 |
| `DefaultPidsLimit` | 512；实例 `PidsLimit=0` 继承它 |
| `UIDMapStart` / `GIDMapStart` | `0` 时读取或分配 `/etc/subuid`、`/etc/subgid` 映射 |
| `IDMapSize` | 默认 65536，不能小于 65536 |
| `MountRoots` | 默认只允许本管理器 `volumes` 下的挂载源 |

映射配置与数据目录绑定，已经导入的镜像及实例记录了映射信息。不能随意更换映射参数后继续使用旧 rootfs。

## 镜像管理

| API | 返回与行为 |
| --- | --- |
| `Import(source) (*Image, error)` | 导入本地 Docker save tar 或 gzip 压缩归档，按归档内容摘要识别镜像 |
| `Images() []*Image` | 返回全部镜像的副本 |
| `Image(id) (*Image, error)` | 按 ID、名称或标签查询镜像 |
| `ExportImage(id, destination) error` | 导出 Docker save 兼容 tar，目标必须尚不存在 |
| `Commit(id, name, tags) (*Image, error)` | 从实例文件系统创建镜像；运行中也可提交，但不是业务一致性快照 |
| `RemoveImage(id) error` | 删除没有实例引用的镜像；停止的实例仍算引用 |

`Import` 接受本地普通文件，不会去远程仓库拉取镜像。归档需要 `manifest.json` 和有效配置、层信息；多个标签可以指向同一个镜像，不支持一个归档含多个不同镜像。不能把 `docker export` 的纯 rootfs 当作 `docker save` 导入。

```go
package example

import "server/app/util/container"

func ImportAndRun(manager *container.Manager, archivePath string) (*container.Info, error) {
    image, err := manager.Import(archivePath)
    if err != nil {
        return nil, err
    }
    return manager.Run(container.RunOptions{
        ID:          "example-service",
        Name:        "示例服务",
        ImageID:     image.ID,
        CreateOnly:  true,
        AutoRestart: true,
        Resources: container.ResourceLimits{
            Memory:    256 << 20,
            CPUQuota:  100000,
            CPUPeriod: 100000,
            PidsLimit: 128,
        },
    })
}
```

该例继承镜像入口与命令，因此传入的镜像必须具有可持续运行的服务入口。CPUQuota/CPUPeriod 单位均为微秒，`100000/100000` 表示每周期最多一个 CPU 核心的执行时间。它们不是百分比字段。

## 实例参数与生命周期

`Run(options) (*Info, error)` 创建或复用实例并启动。`ImageID` 可为已导入镜像 ID、名称、标签，或本机镜像归档的**绝对路径**，解析后持久化为固定镜像 ID。

| RunOptions 字段 | 含义 |
| --- | --- |
| `ID` / `Name` | 稳定实例标识与展示名称；空名称从镜像推导 |
| `Entrypoint` / `Command` | 分别覆盖入口和 Cmd；Command 仍追加在 Entrypoint 后，空值继承镜像 |
| `Env` | `KEY=VALUE` 列表，覆盖镜像同名变量 |
| `WorkingDir` / `User` | 容器内目录和用户，默认继承镜像；目录最终默认为 `/` |
| `Mounts` | 宿主 bind mount，必须处于允许的挂载根目录 |
| `Resources` | 内存字节、CPU 配额/周期、进程数限制 |
| `ReadOnly` | 禁用实例可写层；默认使用独立 overlay 写层保护镜像 |
| `AutoRestart` | 配置退出或恢复时自动启动；手动停止会暂停它 |
| `CreateOnly` | 同 ID 已存在时返回错误，避免误复用 |

命令以参数数组传递，不自动经过 shell。`Command: []string{"echo", "hello"}` 在镜像没有 Entrypoint 时才直接执行 echo；有 Entrypoint 时作为其参数。需要管道、重定向等语义时，同时设置 `Entrypoint: []string{"/bin/sh", "-c"}` 和 `Command: []string{"具体 shell 命令"}`，并确认镜像包含 shell。

| API | 行为 |
| --- | --- |
| `Start(id) (*Info, error)` | 启动保存的实例配置 |
| `Stop(id) error` | 手动停止并暂停自动恢复，超时后强制结束 |
| `Restart(id) (*Info, error)` | 停止后重新启动 |
| `Remove(id) error` | 停止实例并删除实例写层、日志和运行状态；业务挂载数据由调用方管理 |
| `List() []*Instance` | 刷新状态后返回摘要，不包含完整环境变量和挂载 |
| `Info(id) (*Info, error)` | 返回完整配置及最新生命周期状态 |
| `Status(id) (*Status, error)` | 查询生命周期和资源采样 |
| `Close()` | 关闭终端、轮询和管理器，**默认保留容器运行**，供下次接管 |
| `Close(true)` / `StopAll()` | 停止全部容器并关闭管理器，保留实例的自动启动配置 |

`Close` 幂等，但第一次调用决定是否停止容器；先 `Close()` 再 `Close(true)` 不会补做停止。临时演示或测试环境需退出即停时，应从一开始使用 `defer manager.Close(true)`。长期服务正常升级接管可选择 `Close()`。业务仍应处理关闭返回的错误。

## 持久化挂载与 UID/GID

`VolumePath(id)` 仅计算路径，不创建目录，也不要求实例存在。可写目录需要映射后的宿主属主，使用 `HostID(containerUID, containerGID)` 获取，不能假定容器 root 等同宿主 root。

```go
package example

import (
    "os"
    "server/app/util/container"
)

func PrepareVolume(manager *container.Manager, id string) (container.Mount, error) {
    path, err := manager.VolumePath(id)
    if err != nil {
        return container.Mount{}, err
    }
    uid, gid, err := manager.HostID(0, 0)
    if err != nil {
        return container.Mount{}, err
    }
    if err = os.MkdirAll(path, 0750); err != nil {
        return container.Mount{}, err
    }
    if err = os.Chown(path, uid, gid); err != nil {
        return container.Mount{}, err
    }
    return container.Mount{Source: path, Destination: "/data"}, nil
}
```

示例只适用于新建、空的数据目录。目录已有内容时需要核对整个目录树的属主与文件类型；仅修改顶层属主不能修复已有数据。业务程序使用非 root 用户时，要传对应容器 UID/GID。挂载只声明来源、目标和只读状态，目录与文件权限由业务层在运行前设置。

`Commit` 不打包 bind mount 的宿主内容，镜像声明的 `Volumes` 也不会自动创建或挂载。数据库备份应单独处理持久化卷和数据库一致性；镜像归档可结合 [oss](../oss/README.md) 上传。

## 命令、终端和日志

```go
package example

import (
    "fmt"
    "time"
    "server/app/util/container"
)

func CheckService(manager *container.Manager, id string) error {
    result, err := manager.Exec(id, container.ExecOptions{
        Command:   []string{"/bin/sh", "-c", "printf 'ready\\n'"},
        Timeout:   5 * time.Second,
        MaxOutput: 16 << 10,
    })
    if err != nil {
        return err
    }
    if result.TimedOut || result.ExitCode != 0 || result.Truncated {
        return fmt.Errorf("命令未完整成功: %+v", result)
    }
    return nil
}
```

`Exec` 默认超时 30 秒、输出上限 512 KiB。`Stdin` 可传入流，`Stderr` 非 nil 时单独接收错误输出，否则与 stdout 合并；非零退出、超时和截断需结合结果判断，不能只查看错误字符串。该接口没有 context 参数，通过 `ExecOptions.Timeout` 控制期限。

`Terminal(id, height, width)` 返回 `TerminalSession`，可读写、`Resize` 和 `Close`，行列范围为 1..1000。关闭终端不会停止容器主进程；接入 WebSocket 时要在断开连接后关闭 session，不要遗留复制 goroutine。

`ReadLog(id, after, before, pageSize)` 使用字节游标，而非行号或页码；两个游标均为 0 时读最新内容。结果字段与 [log.Read](../log/README.md) 相同，并考虑当前日志和轮转日志。`ClearLog(id)` 清空已有实例日志。日志轮转后旧游标应根据返回信息重新同步。

## 资源采样与并发

`Status.CPUUsage`、`CPUUserUsage`、`CPUSystemUsage` 为累计纳秒，I/O 为累计字节和次数；`MemoryUsage`、`MemoryWorkingSet` 为当前字节数。计算速率应使用两次 `ReadTime` 的间隔，且仅在 `Generation` 相同、累计值未下降时求差。重启后累计值会重置。

`ResourcesAvailable=false` 时不要将零值当作真实使用量，查看 `ResourceError`；资源读取错误不等于容器停止。状态包括 `starting`、`running`、`stopping`、`stopped`、`failed`、`unknown`。

管理器内部按实例串行化生命周期操作，镜像读写和删除通过镜像锁协调，运行操作与 `Close` 协调。`Close` 不等待镜像导入、导出或删除，调用方关闭前应停止提交并等待这些操作完成。内部锁不是跨进程分布式锁，业务层应避免两个进程同时接管同一数据目录。`Info` 等对象包含环境变量和宿主路径，返回给前端前应按业务权限裁剪。

## 排查入口

| 现象 | 首先检查 |
| --- | --- |
| 创建管理器失败 | Linux root/cgo/seccomp、openat2 是否可用、数据目录属主权限、UID/GID 映射冲突 |
| 镜像导入失败 | 是否为 Docker save、是否有多个不同镜像、归档路径与文件类型 |
| 实例启动失败 | 镜像 OS/架构、入口程序、用户、挂载权限、资源限制、宿主端口占用 |
| 手动 Stop 后未自动拉起 | 这是暂停自动恢复的预期行为，显式 Start/Restart |
| 删除镜像失败 | 是否仍被运行中或停止的实例引用 |
| 关闭管理器后进程仍在 | `Close()` 默认保留容器，需停止时调用 `Close(true)` |

更多配置以 [entities.go](entities.go) 为准；项目实际初始化见 [bootstrap/container.go](../../../bootstrap/container.go)。需要管理普通宿主进程时使用 [process](../process/README.md)，无需容器隔离的一次性命令可使用 [cmd](../cmd/README.md)。
