# czdb：纯真数据库读取器

本包负责打开 CZDB 数据库、解密头部、加载索引并查询 IP 原始信息。普通业务优先调用上层 [ip.Query](../README.md)，它已处理内嵌数据库初始化和 IPv4/IPv6 路由。只有需要自定义数据库路径、密钥或生命周期时才直接使用本包。

```go
import "server/app/util/ip/czdb"
```

## 搜索器

| API | 说明 |
| --- | --- |
| `NewDbSearcher(dbFile, queryType, key string) (*DbSearcher, error)` | 打开文件并读取数据库头及索引；`key` 为数据库匹配的 Base64 AES 密钥 |
| `(*DbSearcher).Search(ip string) (string, error)` | 查询单个地址，返回数据库中的原始地区/运营商文本 |
| `(*DbSearcher).Close()` | 关闭文件并释放缓存；没有错误返回值 |

`queryType` 当前仅支持 `constant.MEMORY` 和 `constant.BTREE`。虽然源码部分说明提到其他算法，实际 `Search` 没有 `BINARY` 查询分支，不应传入该值。

| 模式 | 资源行为 | 使用建议 |
| --- | --- | --- |
| `MEMORY` | 将逻辑数据库读取到内存，初始化后关闭底层文件 | 重复查询、可接受数据库内存占用的场景 |
| `BTREE` | 缓存索引，查询时继续使用文件游标 | 希望减少完整文件内存占用时使用；共享实例的查询需在外层串行化 |

`BTREE` 的 `Seek` 和读取组合没有互斥保护，多个 goroutine 不能无保护地共用搜索器。任意模式都不能在查询期间调用 `Close`，关闭后不要继续使用。

## 自定义数据库查询示例

```go
package example

import (
	"fmt"

	"server/app/util/ip/czdb"
	czconstant "server/app/util/ip/czdb/constant"
)

// Lookup 要求 address 的地址族与 databasePath 指向的数据库一致。
func Lookup(databasePath, databaseKey, address string) (string, error) {
	searcher, err := czdb.NewDbSearcher(databasePath, czconstant.MEMORY, databaseKey)
	if err != nil {
		return "", fmt.Errorf("打开 CZDB: %w", err)
	}
	defer searcher.Close()

	region, err := searcher.Search(address)
	if err != nil {
		return "", fmt.Errorf("查询 CZDB: %w", err)
	}
	if region == "" || region == "null" {
		return "", fmt.Errorf("地址 %s 没有可用的地区信息", address)
	}
	return region, nil
}
```

对持续服务应在启动时创建搜索器并复用，在停止接收查询并等待在途请求完成后统一关闭，避免每个请求重新加载整库。不要把 IPv6 地址传给 IPv4 数据库，或反向混用；当前底层没有完整防御性检查，地址族不匹配可能导致切片越界。

## 返回值与数据要求

- 无效 IP 通常返回解析错误；合法地址未命中可以返回空字符串和 `nil`。
- 数据块解码失败时，`DataBlock.GetRegion` 可能以字符串 `"null"` 表示失败，上层需要结合业务处理。
- 文件必须是当前实现兼容的 CZDB 格式，密钥、客户标识、有效期必须匹配。头部解密会检查客户标识和有效期，系统本地日期影响有效期判断。
- 返回的是原始数据库文本，不是结构化 JSON；上层 `ip` 包提供地区字段拆分。
- 读取器假定数据库来源可信，部分长度和切片操作未做全面边界检查，不适合直接解析任意上传文件。

## 随包资源与辅助接口

`Ip` 是嵌入了 `data/*` 的 `embed.FS`，可通过 `czdb.Ip.Open("data/ipv4")`、`ReadFile("data/ipv6")` 读取打包资源。直接查询仍需向 `NewDbSearcher` 提供文件路径，上层 `ip` 负责释放这些资源到运行目录。

`NewDecrypted(key string) *Decrypted` 是地区映射内部解密器的构造入口，实际解密方法未导出，业务层无需单独创建它。

`RandomAccessFile` 将 CZDB 超级头之后的内容视为逻辑文件：

| API | 含义 |
| --- | --- |
| `NewRandomAccessFile(name string, offset int64) (*RandomAccessFile, error)` | 只读打开真实文件，记录逻辑起点；不会自动跳到该起点 |
| `Seek1(offset int64) (int64, error)` | 定位到“逻辑起点 + offset”，返回物理文件位置 |
| `Length() (int64, error)` | 物理长度减去逻辑起点 |
| `ReadFully(p []byte) error` | 从当前游标精确读取 `len(p)` 个字节 |
| `ReadFullyAt(p []byte, off, length int) error` | `off` 是目标切片下标，**不是文件偏移**；仍从当前文件游标读取 |
| `Close() error` | 来自嵌入的 `os.File`，调用方负责关闭 |

`ReadFullyAt` 使用 `io.ReadAtLeast`；当 `p[off:]` 比 `length` 长时可能读取更多字节，并因读取数量与 `length` 不同返回错误。需要精确长度时使用大小匹配的缓冲区和 `ReadFully`。

## 子包关系

- [constant](constant/README.md)：格式偏移和查询模式常量。
- [byteUtil](byteUtil/README.md)：小端整数读写。
- [entity](entity/README.md)：头部、索引、地区数据块结构。
- [aesECB](aesECB/README.md)：CZDB 兼容所需的 AES-ECB 与填充。
- [hyperHeaderDecoder](hyperHeaderDecoder/README.md)：头部解密和客户标识、有效期校验。

返回 [ip 使用说明](../README.md) 或 [util 索引](../../README.md)。
