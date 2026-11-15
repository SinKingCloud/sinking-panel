# platform：运行平台、权限、系统时区与时间同步

导入路径：`server/app/util/platform`。该包集中提供操作系统分支判断、面板所需权限检查、系统时区和时间同步入口。实现与平台限制见 [timezone](timezone/README.md)。

## API

| API | 行为 |
| --- | --- |
| `IsLinux() bool` | 按编译运行平台 `runtime.GOOS` 判断 |
| `CheckPrivilege() error` | Linux 检查有效 UID 为 0；Windows 检查当前令牌是否已提升为管理员；其他平台返回 nil |
| `CanRunContainer() bool` | 仅检查 Linux 且有效 UID 为 0 |
| `TimezonesList() ([]string, error)` | 返回本机支持的已排序时区名称 |
| `TimezoneGet() (string, time.Time, error)` | 返回系统当前配置的时区及该时区的当前时间 |
| `TimezoneSet(name string) error` | 持久化修改系统时区，不调整时间戳 |
| `TimeSync(ctx context.Context) error` | 优先使用系统配置的 NTP 时间源，执行一次系统时间校准，不修改时区 |

`CanRunContainer` 只是初步条件，不检查 OverlayFS、namespace、cgroup、seccomp、CGO 等运行条件。返回 true 不等于容器必然能启动，最终由 [container](../container/README.md) 初始化和启动返回值判断。

## 示例：只读平台与时区信息

```go
package example

import (
	"fmt"
	"server/app/util/platform"
)

func ShowPlatform() error {
	name, local, err := platform.TimezoneGet()
	if err != nil {
		return err
	}
	fmt.Printf("linux=%t container_candidate=%t timezone=%s time=%s\n",
		platform.IsLinux(), platform.CanRunContainer(), name, local)
	return nil
}
```

系统操作入口可以先调用 `CheckPrivilege`，但这不是替代每个操作的权限检查；例如 macOS 上该函数返回 nil，修改系统时区仍可能需要 root。

## 修改时区

先通过 `TimezonesList` 获取本机接受的值，让用户选择后调用 `TimezoneSet(selected)`。Linux/macOS/BSD 使用 IANA 名称，如 `Asia/Shanghai`；Windows 使用系统原生名称，如 `China Standard Time`。不要把一个平台的名称硬编码给另一个平台。

`TimezoneSet` 是对整台主机生效的写操作，应在设置页鉴权与参数校验后执行。仅想按某时区显示业务时间时，应使用标准库 `time.LoadLocation` 与 `Time.In`，不必修改主机时区。

## 同步时间

调用 `TimeSync(ctx)` 即可，无需传入 NTP 地址；该入口位于 `platform/timezone.go`，委托 `timezone.Sync(ctx)` 完成。只有未配置系统时间源时才使用内置公共源；已有源不可达或配置错误时返回错误。调用整体最多等待 15 秒，支持上下文取消，同一进程不并发校时。成功只表示本次系统实时时钟已校准，不会启用定时同步服务。

面板通过 `System.SyncTimezone(ctx)` 接入，`POST /system/timezone` 传 `{"action":"sync"}`。`get`、`set` 保持原用法，同步成功后可再用 `get` 获取当前时间。

## 权限、超时和限制

- Windows 的管理员检查通过系统目录内的 PowerShell 查询当前进程令牌，最多等待 10 秒。属于管理员组但未提升的进程仍不通过；缺少 SystemRoot 或系统 PowerShell 也会返回错误。
- 时区子包使用原生系统能力，不调用 `timedatectl` 等外部命令。系统时区读取不使用进程 `TZ` 环境变量或 `time.Local` 缓存。
- 设置时区不会覆盖应用的 `time.Local`，已运行的其他程序或容器可能仍需重新加载配置。
- 时间同步通过 Unix `settimeofday` 或 Windows `SetSystemTime` 写入主机时钟，无外部命令。Linux 需要 `CAP_SYS_TIME`，其他 Unix 通常需要 root；Windows 需要系统时间修改特权。容器内的 root 不一定具有此权限。
- 部分平台没有实现时区管理，会返回明确的不支持错误。调用方应处理错误，不能只依据 IsLinux 分支推断可用性。

相关：[timezone](timezone/README.md)、[daemon](../daemon/README.md)、[container](../container/README.md)。
