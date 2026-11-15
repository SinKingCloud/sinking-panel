# timezone：系统时区与时间同步

导入路径：`server/app/util/platform/timezone`。本包提供主机时区管理和单次系统时间同步，不依赖面板 service、数据库、CGO 或外部命令。业务代码也可通过父包的 `TimezonesList`、`TimezoneGet`、`TimezoneSet`、`TimeSync` 使用相同功能。

## API 和示例

```go
package example

import (
	"context"
	"fmt"
	"server/app/util/platform/timezone"
)

func CurrentTimezone() error {
	names, err := timezone.List()
	if err != nil {
		return err
	}
	name, now, err := timezone.Get()
	if err != nil {
		return err
	}
	fmt.Printf("当前 %s，时间 %s，可选时区数 %d\n", name, now, len(names))
	return nil
}

// 调用前应完成管理权限验证，selected 来自本机 List 返回的选项。
func ChangeTimezone(selected string) error {
	return timezone.Set(selected)
}

// 执行一次校时，调用前同样应完成管理权限验证。
func SyncClock(ctx context.Context) error {
	return timezone.Sync(ctx)
}
```

`List() ([]string, error)` 返回已排序、去重的名称；`Get() (string, time.Time, error)` 返回名称及当前当地时间；`Set(string) error` 会去掉首尾空白并拒绝空值、空字符及不合法系统时区；`Sync(context.Context) error` 发现本机时间源、测量偏差并校准系统时钟。

## 时区管理的平台行为

| 平台 | 时区名称 | 修改方式 |
| --- | --- | --- |
| Linux | IANA 名称 | 优先通过 system bus 调用 timedated；服务不存在时退回系统文件 |
| macOS、FreeBSD、OpenBSD、NetBSD、DragonFly | IANA 名称 | 持锁更新 `/etc/localtime` 链接，并同步适用的名称文件 |
| Windows | Windows 原生时区 ID | 读取注册表列表，通过原生动态时区 API 设置；临时启用并恢复相关进程特权 |
| 其他 | — | List、Get、Set 返回不支持错误 |

Linux 的系统服务调用最多等待 5 秒。服务明确拒绝或超时不会退回写文件，以免绕过系统权限或把失败误报成功。文件回退要求 root；硬件时钟使用本地时间且缺少可处理偏移的服务时会返回错误。

## Unix 时区库

按顺序定位 `/usr/share/zoneinfo`、`/etc/zoneinfo`、`/var/db/timezone/zoneinfo` 等系统目录。每次调用重新定位，支持系统更新 tzdata 后读取新数据；没有时区库时返回安装 tzdata 的提示。

只读取库内的普通有效 TZif 文件，单文件最大 1 MiB；拒绝绝对路径、目录穿越、反斜杠、`localtime`、`posixrules`、`posix/` 和 `right/` 等非标准选项。名称别名例如 UTC 可以保留。

Get 优先识别 `/etc/localtime` 的链接；如果是普通文件，会验证名称文件与内容是否匹配，必要时遍历时区库按内容比对。自定义 TZif 文件无法对应标准名称时返回错误，不会猜测名称。已找到时区库但 `/etc/localtime` 不存在时，返回 UTC。

## 时区写入和缓存边界

文件方式在系统配置目录上加排他锁，通过临时目录和重命名替换链接，避免直接修改共享 TZif 内容。现有 `/etc/timezone` 才会同步；FreeBSD/DragonFly 还会同步 `/var/db/zoneinfo`。

写操作并非多文件事务：如果链接替换成功但后续名称文件同步失败，错误会明确说明系统时区已经修改。此时先 Get 核对实际状态，再决定是否重试。

Set 不改系统时间戳，不刷新应用的 `time.Local`；返回的 time.Time 才代表本次读取的系统时区。Windows Get 使用当前标准/夏令时偏移构造 FixedZone，不能把这个返回位置当作包含全部历史与未来夏令时规则的时区库。

## 系统时间同步

`Sync(ctx context.Context) error` 返回 nil 表示本次校时成功；错误包含时间源读取、网络查询或系统写入失败的原因。方法不修改时区，不安装或启动定时同步服务，也不写时间服务配置。

## 时间源选择

| 平台 | 查找顺序 |
| --- | --- |
| Linux | 运行中的 systemd-timesyncd 有效时间源 → 标准 chrony/ntp 配置 → timesyncd 配置及 drop-in |
| macOS、FreeBSD、OpenBSD、NetBSD、DragonFly | 标准 chrony/ntp/ntpd 配置 |
| Windows | 组策略中的 W32Time 配置优先，本机配置补齐未设置项；域时间模式由系统定位域控制器 |
| 其他 | 返回不支持错误 |

Linux 查询 D-Bus 前检查 timesyncd 是否已运行，不会隐式启动服务。优先采用当前源、运行时源、系统源及网络提供的源；只有没有普通源时才采用系统的 `FallbackNTPServers`。配置文件方式支持 timesyncd 主配置与同名 drop-in 覆盖，chrony/ntpd 的 `server`、`pool`、`peer` 和标准 `include`、`includefile`、`confdir`、`sourcedir`。

Windows 支持 `NTP`、`NT5DS`、`AllSync` 及 `NtpServer` 的标记格式。`NT5DS` 只查询域时间源，域发现失败时返回错误，不使用遗留的手动源。`AllSync` 优先使用域时间源；域发现失败或超时时，仍可使用系统明确配置的 NTP 源，没有手动源则返回错误。禁用 NTP、选择 `NoSync`、配置无效或没有可用源时返回错误，不另行回退到公共源。域定位使用系统原生 API；该 API 不能中途取消，因此最多保留一个在途原生查询，避免反复超时积累线程。

只有没有找到系统时间源配置时，才依次尝试 `time.cloudflare.com`、`ntp.aliyun.com`。已配置源发生查询失败时不切换到上述公共源。所有候选源按顺序查询，去掉重复项，首次获得有效响应后写入一次时钟。

## 超时和测量

- 整体最多 15 秒；发现系统源最多 3 秒；每个源的 DNS/UDP 请求最多 3 秒。调用方的更短期限优先。
- 同一进程同时只允许一次校时，重复调用立即返回“时间正在同步”。
- 验证 NTP 版本、模式、来源回显、同步状态、时间戳和误差；拒绝未同步、拒绝服务或不匹配的响应。
- 使用单调时钟测量往返耗时，并计入网络传输估算；查询期间本机时钟发生明显变化会失败，避免应用过期偏差。
- 支持 IPv4、IPv6、域名和显式端口；默认 UDP 123。网络必须允许访问相应时间源。
- NTP 32 位秒字段按最接近本机日期的 era 展开，支持 2036 年回绕；本机日期相差超过约 68 年时无法据此可靠确定 era。

上下文取消会停止等待并关闭网络连接；写时钟前再次检查取消状态。已经进入的原生写入调用不可撤销。

## 校时写入与限制

Unix 使用 `settimeofday`：Linux 需要 `CAP_SYS_TIME`，macOS/BSD 通常需要 root，还可能受容器、内核安全级别等限制。Windows 使用 `SetSystemTime` 写入 UTC；系统 API 临时启用并恢复 `SeSystemtimePrivilege`，进程令牌必须具有该权限。

这里调整的是系统实时时钟，不是修改时间显示的时区。不会直接写 RTC 硬件时钟或承诺重启后的硬件时钟状态，也不会改变应用自设时区。

当前不解析守护进程的自定义启动参数或 macOS timed 私有配置格式。不支持 NTS、密钥认证、硬件参考时钟及配置宏；遇到这些源会返回明确错误，不能把认证源当普通 NTP 查询。公共 NTP 和普通 NTP 本身不提供加密认证。

## 文件布局与验证

- `timezone.go`、`timezone_*.go`：时区 API 与各平台实现。
- `database_unix.go`、`filesystem_unix.go`：Unix 时区库读取及配置写入。
- `native_windows.go`、`privilege_windows.go`：Windows 原生时区调用与权限管理。
- `sync.go`：单次同步流程、超时和并发控制。
- `ntp.go`：NTP 查询与偏差测量，不写时钟。
- `sources_linux.go`：Linux 有效时间源查询，D-Bus 依赖仅在 Linux 编译。
- `sources_bsd.go`：macOS/BSD 时间源入口。
- `sources_unix.go`：Unix 共用配置文件发现和解析，Linux 同时处理 timesyncd drop-in。
- `sync_unix.go`、`sync_windows.go`、`sync_other.go`：校时的平台调用及 Windows 时间源发现。

验证优先使用本机 UDP 模拟时间源测试 `query`，用临时配置测试解析，替换平台写入实现后测试同步流程。编译检查可以跨平台进行；不要在自动验证时直接调用真实的 `Sync`，避免修改开发机时钟。

参考：[NTPv4 协议](https://www.rfc-editor.org/rfc/rfc5905.html)、[timesyncd 配置](https://www.freedesktop.org/software/systemd/man/latest/timesyncd.conf.html)、[Windows SetSystemTime](https://learn.microsoft.com/en-us/windows/win32/api/sysinfoapi/nf-sysinfoapi-setsystemtime)、[Cloudflare NTP](https://developers.cloudflare.com/time-services/ntp/)、[阿里云 NTP](https://help.aliyun.com/en/ecs/user-guide/alibaba-cloud-ntp-server/)。

相关：[platform](../README.md)、[str 的本地时间类型](../../str/README.md)。
