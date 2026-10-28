# str：字符串、日期、JSON 与标识工具

本包收集项目已有的字符串操作、密码摘要、日期序列化、编号生成和电话区号工具。多数简单转换用空字符串或 `nil` 表示失败，使用前需要了解各方法的返回约定。

```go
import "server/app/util/str"
```

## 字符串入口

`NewStringTool(value string) *StringTool` 保存待处理文本；`NewStrTool() *StringTool` 创建空文本工具，用于验证码、JSON 等不依赖接收者文本的操作。大部分方法为值接收者，调用后不会把原工具中的字符串替换为结果。

| 方法 | 参数和返回值 | 注意事项 |
| --- | --- | --- |
| `GetBetween(start, end string) string` | 截取开始与结束标记之间的文本 | 适用于已确认格式的文本；缺少标记、顺序错误或过短输入可能 panic |
| `Md5() string` | 当前文本的 MD5 小写十六进制摘要 | 不是密码哈希，也不提供防篡改能力 |
| `Show(left, right int, replacement string) string` | 按 Unicode 字符保留首尾，中间每个字符替换为指定文本 | 首尾数量须非负，且最好满足两者之和不超过字符数 |
| `ConvertToString(srcCode, tagCode string) string` | 使用 mahonia 进行编码转换 | 没有错误返回，需使用受支持编码并自行验证转换结果 |
| `GenValidateCode(width int) string` | 产生指定长度的十进制字符文本 | `width <= 0` 时为空，使用 `math/rand`，不是密码学随机数 |
| `RandInt64(min, max int64) int64` | 当 `min < max` 且两者非零时生成 `[min, max)` 随机数 | 无效范围或任一边界为零时直接返回 `max` |
| `ToJson(data interface{}) string` | `json.Marshal` 后转字符串 | 编码失败时为空字符串 |
| `ToMap(data string) map[string]string` | 解码字符串值的 JSON 对象 | 输入不合法、值不是字符串时返回 `nil` |

`Show` 按 rune 计算长度；首尾数量超出边界会返回空字符串，负数可能 panic，重叠区域可能被重复拼接。需要严格校验或明确错误原因的转换，可以直接使用相应 Go 标准库。

```go
package example

import (
	"fmt"

	"server/app/util/str"
)

func StringsAndJSON() {
	fmt.Println(str.NewStringTool("<id>42</id>").GetBetween("<id>", "</id>")) // 42
	fmt.Println(str.NewStringTool("13812345678").Show(3, 4, "*"))             // 138****5678
	fmt.Println(str.NewStringTool("hello").Md5())

	data := map[string]string{"name": "backup", "status": "ready"}
	encoded := str.NewStrTool().ToJson(data)
	decoded := str.NewStrTool().ToMap(encoded)
	fmt.Println(decoded["name"])
}
```

## 密码存储与校验

`GetPassword() string` 使用 bcrypt 默认成本生成密码哈希，失败时返回空字符串；bcrypt 的明文长度上限为 72 字节，中文等多字节字符按 UTF-8 字节数计算。`CheckPassword(password string) bool` 的接收者必须是**已保存的哈希**，参数为用户提交的明文；不匹配或哈希格式错误时返回 `false`。

```go
package example

import (
	"errors"

	"server/app/util/str"
)

func HashPassword(plain string) (string, error) {
	hashed := str.NewStringTool(plain).GetPassword()
	if hashed == "" {
		return "", errors.New("密码哈希生成失败，请检查密码字节长度")
	}
	return hashed, nil
}

func VerifyPassword(savedHash, submitted string) bool {
	return str.NewStringTool(savedHash).CheckPassword(submitted)
}
```

保存哈希并在登录时校验，不需要解密；不要用 `Md5` 替代密码哈希。验证码生成不包含存储、过期、发送频率或验证次数控制，调用方仍需实现这些流程。

## 历史 AES 接口

| 方法 | 密文格式和处理方式 |
| --- | --- |
| `AesCbcEncrypt(key string) string` | AES-CBC，不自动填充；明文长度必须是 16 字节整数倍；生成随机 IV，输出“IV + 密文”的十六进制文本并带末尾换行 |
| `AesCbcDecrypt(key string) string` | 解码十六进制，前 16 字节作为 IV，解密其余数据，不移除填充 |
| `AesEncrypt(key string) string` | AES-CBC + PKCS7 填充，使用密钥前 16 字节作为固定 IV，返回 Base64 |
| `AesDecrypt(key string) string` | 解码 Base64，使用固定 IV 解密并移除填充 |

密钥必须是 16、24 或 32 **字节**。两组接口的格式不同，必须成对使用。`AesCbc...` 部分失败返回空字符串，且解码错误没有完整保留；保存或传递十六进制密文时可先用 `strings.TrimSpace` 去掉末尾换行。`AesEncrypt/AesDecrypt` 忽略了部分底层错误，错误密钥、非完整密文、无效填充等可能 panic。

这些接口没有消息认证，用于兼容当前项目已有数据；新备份文件加密应设计独立的带认证格式，不能把当前方法当成可直接处理任意不可信密文的解密服务。它们也不是密码存储接口。

## 字节与 JSON 工具

- `NetByteTool(data []byte) *ByteTool`：注意构造函数实际拼写为 **NetByteTool**；`Md5() string` 返回字节内容的 MD5。构造时没有复制切片，调用前修改原切片会影响摘要，并发修改需自行同步。
- `NewJsonTool(data interface{}) *JsonTool`：保存数据，`ToString() string` 返回 JSON；序列化失败时为空字符串。适合已知可编码的数据；需要区分失败原因时直接使用 `encoding/json`。

## DateTime：无时区日期字段

`type DateTime time.Time` 用于 JSON 和数据库字段，统一文本格式为 `2006-01-02 15:04:05`。

| API | 行为 |
| --- | --- |
| `MarshalJSON() ([]byte, error)` | 输出带引号的日期字符串，不包含时区偏移和亚秒信息 |
| `(*DateTime).UnmarshalJSON(data []byte) error` | 按 `time.Local` 解析；JSON `null` 不改变已有值 |
| `Value() (driver.Value, error)` | 向 SQL 驱动提供日期字符串 |
| `(*DateTime).Scan(v interface{}) error` | 接受 `time.Time`、`string`、`[]byte`、`nil`；其他类型报错 |
| `(*DateTime).String() string` | 返回同一日期格式 |

`Scan(time.Time)` 保留输入年月日时分秒后按本地时区重新解释，并不是将同一瞬间转换到本地时区。`Scan(nil)` 设为零值，与 JSON `null` 的行为不同。解析失败时必须检查错误，不能继续使用结果；JSON 解析失败可能已将接收者改为零时间。

```go
package example

import (
	"encoding/json"
	"fmt"
	"time"

	"server/app/util/str"
)

func DateExample() error {
	record := struct {
		CreatedAt str.DateTime `json:"created_at"`
	}{CreatedAt: str.DateTime(time.Now())}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	fmt.Println(string(data))

	var saved str.DateTime
	if err := saved.Scan("2026-01-02 03:04:05"); err != nil {
		return err
	}
	fmt.Println(saved.String())
	return nil
}
```

这类文本适合系统约定的本地日期字段；跨时区接口若需要保留绝对时间，应明确采用 Unix 时间戳或含时区的格式。

## 时间边界

`NewTimeTool() *TimeTool` 提供：

- `GetZeroTime(d time.Time) time.Time`：当天零点。
- `GetFirstDateOfMonth(d time.Time) time.Time`：当月第一天零点。
- `GetLastDateOfMonth(d time.Time) time.Time`：当月最后一天**零点**，不是当天的最后一秒。

结果保留输入时间的 `Location`。按月查询推荐半开区间：`start := tool.GetFirstDateOfMonth(now)`，`end := start.AddDate(0, 1, 0)`，条件为 `start <= timestamp && timestamp < end`，这样不会漏掉最后一天或亚秒数据。

## Snowflake 风格编号

| API | 含义 |
| --- | --- |
| `NewSnowWorker(workerId int64) (*Worker, error)` | 创建节点生成器，节点编号必须在 `0..1023` |
| `GetSnowWorkIns() *Worker` | 获取进程内默认单例 |
| `(*Worker).GetId() int64` | 生成正整数编号，内部互斥锁保护并发调用 |
| `(*Worker).GetUuid() string` | 对生成编号做 MD5，再按项目自定义分组格式返回字符串 |

时间部分按秒计算，以 2024-01-01 UTC 为基准，每个节点每秒最多使用 16384 个序列值，序列耗尽会等到下一秒。时钟回拨时先沿用上次时间，序列耗尽后可能等待时钟追上。时间超出可表达范围时会 panic。

```go
package example

import (
	"fmt"

	"server/app/util/str"
)

func CreateIDs() error {
	// 节点编号必须由部署配置统一分配，不同进程不能重复使用。
	worker, err := str.NewSnowWorker(7)
	if err != nil {
		return err
	}
	fmt.Println(worker.GetId())
	fmt.Println(worker.GetUuid())
	return nil
}
```

一个进程应复用同一个节点实例。默认单例的节点号由主机名、进程信息、网络地址等哈希后取模生成，不提供集群节点冲突协调；多进程部署、重启及同节点重复实例的唯一性需由部署方案保证。

`GetUuid` 的分组长度为 `6-4-4-6-12`，不是标准 RFC UUID 的 `8-4-4-4-12`，不要传给要求标准 UUID 格式的库。编号具有时间和节点规律，不应直接用作秘密令牌。

## 电话区号与昵称

`str.Area.Get()` 返回国际电话区号列表，每项为 `map[string]string`，包含 `area_en`、`area_cn`、`code`。`str.Area.CheckIn(code string) bool` 按字符串精确匹配区号，例如 `str.Area.CheckIn("86")`；传入 `"+86"` 不会自动去掉加号。列表是项目内置数据，不是 ISO 地区编码库，也可能存在共享同一区号的地区。

`Area` 的类型未导出，当前使用其方法无须自行实例化；每次 `Get` 会构造列表，循环查询时可以在外层缓存需要的数据。不要把它当作需要赋值初始化的配置字段。

`NewNickTool().GetRand()` 返回随机昵称，不保证唯一；需要唯一昵称时由调用方结合数据库约束或重试处理。

返回 [util 索引](../README.md)，系统时区操作见 [platform/timezone](../platform/timezone/README.md)。
