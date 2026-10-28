# platform：运行平台、权限和系统时区

导入路径：`server/app/util/platform`。该包集中提供操作系统分支判断、面板所需权限检查以及系统时区入口。详细时区实现见 [timezone 子包](timezone/README.md)。

## API

| API | 行为 |
| --- | --- |
| `IsLinux() bool` | 按编译运行平台 `runtime.GOOS` 判断 |
| `CheckPrivilege() error` | Linux 检查有效 UID 为 0；Windows 检查当前令牌是否已提升为管理员；其他平台返回 nil |
| `CanRunContainer() bool` | 仅检查 Linux 且有效 UID 为 0 |
| `TimezonesList() ([]string, error)` | 返回本机支持的已排序时区名称 |
| `TimezoneGet() (string, time.Time, error)` | 返回系统当前配置的时区及该时区的当前时间 |
| `TimezoneSet(name string) error` | 持久化修改系统时区，不调整时间戳 |

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

## 权限、超时和限制

- Windows 的管理员检查通过系统目录内的 PowerShell 查询当前进程令牌，最多等待 10 秒。属于管理员组但未提升的进程仍不通过；缺少 SystemRoot 或系统 PowerShell 也会返回错误。
- 时区子包使用原生系统能力，不调用 `timedatectl` 等外部命令。系统时区读取不使用进程 `TZ` 环境变量或 `time.Local` 缓存。
- 设置时区不会覆盖应用的 `time.Local`，已运行的其他程序或容器可能仍需重新加载配置。
- 部分平台没有实现时区管理，会返回明确的不支持错误。调用方应处理错误，不能只依据 IsLinux 分支推断可用性。

相关：[timezone](timezone/README.md)、[daemon](../daemon/README.md)、[container](../container/README.md)。
