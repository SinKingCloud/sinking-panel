# timezone：系统原生时区管理

导入路径：`server/app/util/platform/timezone`。一般业务代码也可通过父包的 `TimezonesList`、`TimezoneGet`、`TimezoneSet` 使用相同功能。本包操作的是主机时区配置，不是某个 Go context 的时区。

## API 和示例

```go
package example

import (
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
```

`List() ([]string, error)` 返回已排序、去重的名称；`Get() (string, time.Time, error)` 返回名称及当前当地时间；`Set(string) error` 会去掉首尾空白并拒绝空值、空字符及不合法系统时区。

## 平台行为

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

## 写入和缓存边界

文件方式在系统配置目录上加排他锁，通过临时目录和重命名替换链接，避免直接修改共享 TZif 内容。现有 `/etc/timezone` 才会同步；FreeBSD/DragonFly 还会同步 `/var/db/zoneinfo`。

写操作并非多文件事务：如果链接替换成功但后续名称文件同步失败，错误会明确说明系统时区已经修改。此时先 Get 核对实际状态，再决定是否重试。

Set 不改系统时间戳，不刷新应用的 `time.Local`；返回的 time.Time 才代表本次读取的系统时区。Windows Get 使用当前标准/夏令时偏移构造 FixedZone，不能把这个返回位置当作包含全部历史与未来夏令时规则的时区库。

相关：[platform](../README.md)、[str 的本地时间类型](../../str/README.md)。
