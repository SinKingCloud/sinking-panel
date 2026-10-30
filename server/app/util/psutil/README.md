# psutil 系统信息工具

`server/app/util/psutil` 封装主机资源、进程和网络信息。调用方只使用本包的结构体，不依赖 gopsutil 的返回类型。现有底层依赖为 `github.com/shirou/gopsutil/v4`，无需增加新的依赖。

## 职责与调用方式

- 工具层负责同步读取、固定格式转换、差分速率、设备筛选和进程操作。
- 工具层没有定时器、后台协程、数据库、配置服务或日志文件；不会自行开始监控。
- `system service` 保留每秒采样、每 5 秒刷新磁盘容量、实时缓存、120 条 I/O 窗口，以及 `recordstore` 历史记录与保留策略。
- 文件管理的磁盘列表也复用 `Disks`，原接口的路径、容量和 inode 字段保持不变。
- 本次提供 Go API；没有增加进程/网络管理 HTTP 路由。后续控制器负责鉴权、参数校验、分页和操作日志。

所有读取和控制方法都接收 `context.Context`。使用请求上下文或业务生命周期上下文，不要传 `nil`。支持超时的底层调用会跟随上下文取消；部分系统调用本身不能中断，不承诺超时瞬间中断内核调用。

## API 总览

| API | 返回 | 用途 |
| --- | --- | --- |
| `Host(ctx)` | `HostInfo, error` | 主机名、系统版本、内核、启动时间、运行时间 |
| `CPU(ctx)` | `CPUInfo, error` | 型号、逻辑核心数、累计 CPU 时间 |
| `Memory(ctx)` | `MemoryInfo, error` | 物理内存总量、使用量、空闲/可用量 |
| `Load(ctx)` | `LoadInfo, error` | 原生 1/5/15 分钟负载 |
| `Disks(ctx)` | `[]DiskInfo, error` | 可浏览磁盘卷、容量、inode |
| `DiskUsage(ctx, path)` | `DiskUsageInfo, error` | 指定路径所在文件系统的容量 |
| `DiskCounters(ctx)` | `[]DiskIO, error` | 各磁盘累计 I/O 计数 |
| `NetworkCounters(ctx)` | `[]NetworkIO, error` | 各网卡累计流量和状态 |
| `NewSampler().Sample(ctx)` | `Snapshot` | CPU 使用率、内存、负载、网卡/磁盘速率 |
| `AggregateDevices(disks, networks)` | 两个名称切片 | 筛选适合整机汇总的设备，降低重复计数 |
| `Processes(ctx)` | `[]ProcessSummary, error` | 按 PID 升序列出进程摘要 |
| `NewProcessSampler().Sample(ctx)` | `ProcessTopSnapshot` | 独立差分基线的 CPU、RSS、磁盘 I/O 前 5 名 |
| `NewProcess(ctx, pid)` | `*Process, error` | 获取绑定进程身份的句柄 |
| `p.Info(ctx)` | `ProcessInfo, error` | 指定进程的路径、启动参数、用户、状态 |
| `p.Sample(ctx)` | `ProcessUsage, error` | 指定进程的 CPU、内存、线程和 I/O |
| `p.Control(ctx, action)` | `error` | 结束、强制结束、暂停、恢复进程 |
| `Interfaces(ctx)` | `[]InterfaceInfo, error` | 网卡地址、MAC、MTU、标志 |
| `Connections(ctx, filter)` | `[]ConnectionInfo, error` | TCP/UDP 连接及所属进程 |
| `Ports(ctx, filter)` | `[]ConnectionInfo, error` | TCP 监听和 UDP 已绑定端口 |

集合成功且无数据时返回空切片。失败要检查 `error`，不能把失败当成“没有进程/没有连接”。结构体字段不随平台增删，未支持的值通过错误或可用标志解释。

## 读取基础信息

```go
import (
    "context"
    "server/app/util/psutil"
    "time"
)

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

host, err := psutil.Host(ctx)
if err != nil {
    return err
}
// host.Hostname、host.Platform、host.KernelVersion、host.Uptime

memory, err := psutil.Memory(ctx)
if err != nil {
    return err
}
// memory.Total、memory.Used、memory.AvailableBytes、memory.UsedPercent
```

`CPU(ctx)` 返回累计时间，不会等待一个采样周期，也不会读取全局共享的 CPU 使用率基线。需要当前使用率时使用下面的 `Sampler`。

## 连续资源采样

```go
sampler := psutil.NewSampler()
_ = sampler.Sample(ctx) // 建立基线，首次速率不可用

ticker := time.NewTicker(time.Second)
defer ticker.Stop()
for {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-ticker.C:
        data := sampler.Sample(ctx)
        if data.CPU.Available {
            // data.CPU.Usage 是整机使用率，范围 0–100
        }
        for _, network := range data.Network {
            if network.RateAvailable {
                // network.SendRate / RecvRate：字节/秒
            }
        }
        // data.Errors 保存本轮失败项；业务决定展示、重试或跳过历史记录。
    }
}
```

每个独立监控任务持有自己的 `Sampler`，不要每次请求新建，否则无法得到差分速率。同一实例可并发调用，内部串行采样；一个实例对应一个采样节奏。对象不需要 `Close`，没有后台任务需要回收。

`Sample` 不主动等待间隔，但读取系统信息仍可能耗时。慢挂载点容量不包含在 `Sample` 中，容量应低频单独读取。

暂时采集失败会保留上一次成功基线，恢复后按两次成功采样间的真实间隔计算平均速率。失败本轮不会当成有效历史点记录。

### 快照与失败语义

`Snapshot` 固定包含 `time / cpu / memory / load / network / disk / errors`。`errors` 按指标名保存原因；一项失败仍返回其他成功项。

- `CPU.Available`：是否得到两次有效累计计数的差分。首次为 `false`，不能把初始 `Usage=0` 绘成真实采样点。
- `Memory.Available`、`Load.Available`：该项是否读取成功。
- `NetworkIO.RateAvailable`、`DiskIO.RateAvailable`：是否拥有该设备的有效差分。新设备或累计计数回退时重新建立基线。
- `NetworkIO.StateAvailable`：网卡启用/环回标志是否取到；状态读取失败不妨碍流量采样。
- 平台不支持的子计数可能返回零，例如部分磁盘时间字段。零不总代表真实没有活动，需要结合平台说明。

Windows 没有 Unix 含义的系统负载；`Load` 返回不可用错误，`Sampler` 根据忙碌核心数、实际采样间隔计算指数移动均值，并标记 `Load.Estimated=true`。该值不是系统运行队列长度。

## 单位与字段约定

| 字段 | 单位/含义 |
| --- | --- |
| `HostInfo.BootTime`、`HostInfo.Uptime` | Unix 秒、运行秒数 |
| `Snapshot.Time`、进程 `CreateTime` / `SampleTime` | Unix 毫秒，不携带时区 |
| 内存、磁盘容量、累计 I/O | 字节 |
| `SendRate`、`RecvRate`、读写字节速率 | 字节/秒 |
| `ReadCountRate`、`WriteCountRate` | 次/秒 |
| CPU 累计时间 | 秒；整机为所有核心合计 |
| 整机 `CPU.Usage` | 0–100%，所有核心平均 |
| `ProcessUsage.CPUPercent` | 单核 100%，多线程进程可以超过 100% |
| `ProcessTopItem.CPUPercent` | 整机归一化 0–100%，与所有核心累计 CPU 时间差分比较 |
| 磁盘 `ReadTime`、`WriteTime`、`IOTime`、`WeightedIO` | 毫秒，是否提供取决于平台 |
| `DiskIO.IOLatency` | 累计读写时间 / 累计读写次数；不是当前窗口的延迟 |
| 磁盘 `Size.Percentage`、`Inodes.Percentage` | 整数百分比，沿用文件管理格式 |

`Memory.Free` 是空闲内存；`AvailableBytes` 是系统估计还能分配的内存，两者不可互换。数值保留原始单位，格式化为 MB、GB 等由调用方负责。

## 磁盘与设备汇总

```go
volumes, err := psutil.Disks(ctx)
if err != nil {
    return err
}
usage, err := psutil.DiskUsage(ctx, "/") // Windows 可传 C:\\
if err != nil {
    return err
}
// usage.Size.Total / Used / UnUsed；usage.Inodes
```

`Disks` 用于文件管理可浏览卷，过滤系统虚拟文件系统；它不是所有物理硬盘的硬件清单。单个卷读取失败会跳过，所有候选卷均失败则返回错误。macOS 隐藏系统卷不重复显示；APFS 根目录按共享空间计算使用量。Windows 列表路径延续面板的 `C://` 格式。对被过滤挂载点仍可直接调用 `DiskUsage`。

磁盘容量与磁盘 I/O 是两种信息，卷不一定与物理 I/O 设备一一对应。`AggregateDevices` 用于整机流量/吞吐汇总：过滤回环、常见虚拟网卡、分区、叠加块设备等，Linux 优先使用有硬件设备的网卡。这是设备拓扑的启发式判断，不适合替代详细设备列表。输入不修改，结果按名称排序。

## 进程列表、详情和监控

```go
items, err := psutil.Processes(ctx)
if err != nil {
    return err
}
// items 只含 PID、父 PID、名称、创建时间和字段读取错误。

p, err := psutil.NewProcess(ctx, targetPID)
if err != nil {
    return err
}
info, err := p.Info(ctx)
if err != nil {
    return err
}
// info.Executable、info.Command、info.Directory、info.Username、info.Status

first, err := p.Sample(ctx)
if err != nil {
    return err
}
// 首次 CPU/IO 累计值可用，但 CPURateAvailable / IORateAvailable 为 false。
// 在业务下一次轮询时继续调用同一个 p.Sample(ctx)，不必重新 NewProcess。
```

列表刻意不逐个读取命令行、用户和资源详情，避免进程较多时遍历大量昂贵指标。状态也按需在 `Info` 中读取，避免 macOS 为列表中的每个进程调用 `ps`。列表遇到已退出进程会跳过；权限导致字段缺失时在该条 `Errors` 中说明。

`Info` 的非关键字段失败会返回部分信息；`Sample` 的 CPU、内存、I/O 可分别失败，对应可用标志与 `Errors` 会说明。进程已退出、身份改变或上下文取消属于方法整体错误，调用方应忽略该次采样。

进程 `ReadBytes/WriteBytes` 在 macOS 使用系统磁盘字节计数，Linux/Windows 使用进程 I/O 字节计数；各操作系统对缓存 I/O 的统计口径不同，不适合跨系统直接对比。

### 低开销进程排行

`NewProcessSampler()` 返回独立同步采样器，统一调用 gopsutil 的进程枚举、创建时间、累计 CPU 时间、RSS 和 I/O 接口，得到三类资源各前 5 名。每组按对应指标降序、同分时 PID 升序排列，只维护容量为 5 的有序数组，未达到末位的候选直接跳过，不全量排序资源用量。首次只建立 CPU/I/O 基线，内存可以立即显示；不可用的速率不作为零值加入榜单。

名称仅为三榜并集最多 15 个进程调用 `NameWithContext`，按进程创建时间绑定并缓存一分钟，失败时保留同一进程已知名称或显示 PID。返回值不包含命令参数、完整路径、用户、环境和线程信息。底层库可能读取命令行或可执行路径以恢复名称，本包不额外调用这些接口，也不把相关内容返回或存储。Linux/macOS/Windows 的这些采样接口不逐进程执行外部命令。

工具类不安排周期任务或后台 worker、不读业务配置、不写数据库。调用方维护一个采样器并决定刷新周期。`Sample(ctx, ProcessSampleOptions{Snapshot: &snapshot})` 共用该轮主机采样的时间与总 CPU 计数；省略选项自行读取主机计数。系统监控 service 每秒同轮采集主机与进程的 CPU、RSS、磁盘 I/O，保留限时与限量预算。`ProcessTopSnapshot.Clone()` 复制榜单切片，可用于发布不共享可变切片的缓存。

扫描采用协作式预算：Linux 为 200ms，macOS/Windows 为 500ms，每轮最多读取 10000 个进程。超限后保留轮转位置，下一轮继续后续 PID，不能把未扫描进程算成零。基线按 PID 和库返回的毫秒创建时间绑定；采样前后使用不同的 gopsutil 进程对象复核身份，避免对象的创建时间缓存掩盖 PID 复用。进程退出、身份变化或计数器回退时重新建立对应基线。未扫描基线最多保留 2 分钟，最多跟踪 100000 个进程。CPU 和磁盘 I/O 只比较相邻两轮成功读取的进程计数；轮转或失败跳过的进程先重建基线，不把跨多轮平均值当成当前窗口占用。创建时间明确处于当前窗口的新进程可直接使用其累计 CPU/I/O，避免漏掉第一轮；已在两轮间退出的进程无法恢复，会标记速率排行不完整。磁盘 I/O 使用同轮主机快照之间的时间间隔。主机 CPU 读取错误时丢弃可能陈旧的总计数，恢复首轮仅建立基线；主机磁盘读取错误及恢复首轮暂停进程 I/O 速率，避免把单轮进程速率配给跨周期整机速率。其他可用指标继续返回。单个不可中断的系统调用仍可能超过预算，不承诺硬实时取消。包初始化时启用 gopsutil 的启动时间缓存，减少 Linux 进程查询中的重复文件读取，运行中不切换该全局开关。

`ProcessTopSnapshot` 的 `cpu / memory / disk_io` 均为 `ProcessTopItem` 切片。对应 `CPUAvailable / MemoryAvailable / IOAvailable` 表示本轮是否有有效排行；`CPUPartial / MemoryPartial / IOPartial` 分别表示对应排行存在基线、退出、权限或预算缺项，Windows 缺少磁盘排行不影响 CPU 完整性；`Partial` 表示任一采集指标不完整，`Error` 给出摘要，不能把部分扫描称为全机精确 Top 5。身份不可读取或改变的进程计入 `Skipped` 并清除基线；CPU 或 I/O 读取失败会重新建立对应基线，不影响同一进程的其他有效指标。每项的 CPU、RSS、读取速率、写入速率 JSON 字段分别为 `cpu / memory / disk_read / disk_write`；单位分别为百分比、字节、字节/秒、字节/秒，`CreateTime` 为 Unix 毫秒。单项 `CPUAvailable / IOAvailable` 可区分内存榜单里的速率尚未就绪与真实零值。

排行榜只包装现成 gopsutil API，不再维护独立的 procfs、sysctl、libproc 或 Windows 原生实现。当前依赖版本在 Linux/macOS 的 `DiskReadBytes / DiskWriteBytes` 提供实际磁盘计数，Linux 的 `ReadBytes / WriteBytes` 是逻辑 I/O，不能替代使用。macOS 库的 CPU/RSS 查询在部分权限失败时可能返回零结构而不报错，因此累计 CPU 为零、RSS 为零时保守标记相应指标不可用；真实刚创建进程可能延后一轮进入 CPU 榜单。Windows 通用进程 I/O 混合文件、网络和设备传输，因此进程磁盘 Top 5 暂不可用（`IOAvailable=false`、空列表，并返回说明）；CPU、内存排行和整机磁盘趋势不受影响。已有 `Process.Sample` 接口保持不变。其他平台的排行榜返回不可用信息，不影响已有进程详情 API。

### 进程控制

```go
p, err := psutil.NewProcess(ctx, targetPID)
if err != nil {
    return err
}
if err := p.Control(ctx, "terminate"); err != nil {
    return err
}
```

| action | Linux/macOS 等 POSIX | Windows |
| --- | --- | --- |
| `terminate` | SIGTERM，请求进程自行结束 | TerminateProcess，强制结束 |
| `kill` | SIGKILL，强制结束 | 强制结束 |
| `suspend` | SIGSTOP | 暂停进程 |
| `resume` | SIGCONT | 恢复进程 |

控制方法不等待退出、不递归处理子进程、不自动提升权限。操作需要系统允许当前用户控制目标进程。进程退出后旧句柄不再复用；需要操作后来创建的进程时重新 `NewProcess`。

句柄保存 PID 与创建时间，每次读取/控制前重新核对身份，采样结束也核对一次，以发现 PID 复用。检查与系统发送信号不是跨平台原子操作，不能承诺完全消除二者之间的竞争窗口。创建时间无法读取时不能建立句柄。启动新程序仍使用项目的命令执行工具；这里不猜测已退出程序的工作目录、环境或重启参数。

## 网络连接、端口与网卡

```go
connections, err := psutil.Connections(ctx, psutil.ConnectionFilter{
    Kind: "tcp",
    PID: targetPID,
})
if err != nil {
    return err
}

listeners, err := psutil.Ports(ctx, psutil.ConnectionFilter{
    LocalPort: 6379,
})
if err != nil {
    return err
}

interfaces, err := psutil.Interfaces(ctx)
if err != nil {
    // 可能已有部分网卡信息；单个网卡的 Error 提供地址读取失败原因。
    return err
}
```

筛选参数均可省略：

- `Kind`：空/`all`、`tcp`、`tcp4`、`tcp6`、`udp`、`udp4`、`udp6`。不含 Unix 域套接字。
- `PID`：0 表示所有可见进程。指定后优先使用底层按进程查询，避免逐连接查询进程详情。
- `LocalPort` / `RemotePort`：0 不限制；有效端口不超过 65535。
- `Status`：大小写不敏感。TCP 状态统一为 `LISTEN`、`ESTABLISHED`、`SYN_SENT`、`SYN_RECV`、`FIN_WAIT1`、`FIN_WAIT2`、`TIME_WAIT`、`CLOSE`、`CLOSE_WAIT`、`LAST_ACK`、`CLOSING` 等。Windows 别名也可作为筛选输入。

UDP 没有 TCP 的 LISTEN 状态，本包统一返回 `BOUND` 或 `CONNECTED`。`Ports` 同时返回 TCP 监听和 UDP 已绑定端口，保留底层返回的进程/套接字信息，本包不再次按端口号合并。结果完整性受系统权限及底层枚举、去重策略限制，例如 Linux 上复用端口的多个套接字可能被底层合并，不能保证列出全部持有进程。

`PID=0` 表示系统未提供或权限不足以确定所属进程。Windows 不提供文件描述符时 `FD=0`。没有远端的监听或 UDP 端点可能为空/零端口。接口 `IsUp=true` 只代表启用，不代表可以访问互联网。

这些查询不主动扫描网络、不连接目标服务，也不修改防火墙。要释放监听端口，需要通过业务确认目标进程后控制该进程；本工具不提供粗暴的“按端口杀进程”。

## 平台与权限

| 功能 | Linux | Windows | macOS |
| --- | --- | --- | --- |
| 主机/CPU/内存/磁盘 | 底层支持 | 底层支持 | 底层支持，保留 APFS 处理 |
| 原生平均负载 | 支持 | 不支持，采样器提供估算 | 支持 |
| 进程读取/控制 | 受用户、procfs、命名空间权限限制 | 受进程访问权限与受保护进程限制 | 受用户和系统隐私/沙箱限制 |
| TCP/UDP 查询 | procfs，权限影响 PID 完整性 | 原生网络表，部分底层错误可能被库忽略 | 底层依赖系统 `lsof` |
| 进程状态 | 底层支持 | 库不提供时在 Errors 中返回原因 | 底层可能调用系统 `ps` |

FreeBSD、OpenBSD 等平台沿用 gopsutil 的支持范围，部分功能可能返回未实现错误；不能仅凭交叉编译成功宣称所有指标可用。macOS/FreeBSD 连接查询依赖 `lsof`，不能假定所有平台都完全不执行系统命令。Windows/Linux 权限受限时即使查询无错误，也可能只返回当前权限可见连接；本包无法恢复底层未公开的数据。

参考：[gopsutil 官方文档](https://github.com/shirou/gopsutil)。升级依赖时应重新核对各平台实现与字段语义。

## 文件组织与维护

| 文件 | 内容 |
| --- | --- |
| `entities.go` | 统一定义所有固定返回结构、筛选参数及进程句柄 |
| `resources.go` | 主机、CPU、内存、负载、累计计数读取 |
| `sampler.go` | 每个调用方独立的 CPU/I/O 基线与 Windows 负载估算 |
| `disk.go` | 磁盘容量、文件管理挂载卷过滤 |
| `devices.go` | 整机统计的设备去重选择 |
| `process.go` | 进程列表、句柄、资源采样与控制 |
| `process_top.go` | 统一包装 gopsutil 的进程排行榜、轮转预算、差分基线与按需名称 |
| `network.go` | 网卡、连接、监听端口与筛选 |

新增指标时先定义固定结构和单位，再封装系统读取；不要把控制器、全局 service、数据库或后台任务放入此包。新增昂贵指标采用按需 API，避免塞进每秒资源采样。

维护验证建议：

1. 编译 `./app/util/psutil`、`./app/service/system`、`./app/service/file` 和全后端。
2. 用独立采样器验证首轮、第二轮、设备计数回退与取消，不能把缺测写成零。
3. 进程控制只用测试程序自己创建的子进程；查询与端口测试只开临时回环监听。
4. Linux/Windows 做交叉编译，并在目标系统验证权限、字段可用性；编译不替代运行测试。
5. 检查 `/system/info`、`/system/statis`、`/system/io` 及监控历史字段仍符合前端约定。
