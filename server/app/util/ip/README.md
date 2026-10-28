# ip：离线 IP 归属地查询

`ip` 是业务层查询入口，使用项目内嵌的纯真 CZDB 数据库，支持 IPv4 和 IPv6。查询本身不调用第三方 HTTP 服务。地址解析、数据库选择和结果拆分由本包统一完成。

```go
import "server/app/util/ip"
```

## 查询接口

```go
func Query(ip string) (*Info, error)
```

参数必须是 `net.ParseIP` 能解析的单个 IP 地址；域名、CIDR 网段、带端口的地址不是有效参数。返回值结构如下，所有字段均为字符串：

| 字段 | JSON 字段 | 含义 |
| --- | --- | --- |
| `IP` | `ip` | 调用时传入的地址文本 |
| `Country` | `country` | 国家或地区 |
| `Province` | `province` | 省级地区 |
| `City` | `city` | 城市 |
| `Districts` | `districts` | 区县 |
| `Isp` | `isp` | 运营商信息 |

地理信息依赖当前数据库，字段可能为空。实现使用原始结果中的 `–` 拆分地区，使用最后一个制表符分离运营商，不能把查询结果视为精确定位。

## 使用示例

```go
package example

import (
	"fmt"

	"server/app/util/ip"
)

func PrintAddress(address string) error {
	info, err := ip.Query(address)
	if err != nil {
		return fmt.Errorf("查询 IP %s: %w", address, err)
	}
	fmt.Printf("IP=%s 国家=%s 省份=%s 城市=%s 区县=%s 运营商=%s\n",
		info.IP, info.Country, info.Province, info.City, info.Districts, info.Isp)
	return nil
}
```

例如调用 `PrintAddress("8.8.8.8")` 或传入一个完整 IPv6 地址。结果随数据库内容变化，不应在业务中依赖示例地址必定属于某一地区。

## 初始化与运行要求

- 第一次查询合法 IP 时通过 `sync.Once` 初始化 IPv4、IPv6 两个搜索器；非法 IP 会在初始化前返回错误。
- 数据文件位于 `server/app/constant.IPPath`，名称是 `ipv4`、`ipv6`。文件不存在时从内嵌资源释放，目录需要有创建、写入权限；已存在的文件会直接复用。
- 两个数据库使用 `MEMORY` 模式，初始化后查询不需要逐条读取磁盘，但会持有数据库内容所需的内存。
- 初始化只执行一次。如果第一次初始化因权限、文件损坏等失败，修复后需要重启进程；本包没有重新初始化、切换数据库或关闭搜索器的公开接口。
- 已存在的数据文件不会因更新程序自动覆盖。更新库文件时要保证格式和配套密钥一致，并在进程重新初始化之前完成替换。

`Query` 可作为进程内共享入口使用，初始化由 `sync.Once` 协调；不要据此推断底层所有 CZDB 查询模式都适合共享并发访问。

## 错误处理

常见错误包括“ip格式不正确”“ipv4数据库初始化失败”“ipv6数据库初始化失败”“未查询到数据”，搜索器也可能返回数据库读取或解码错误。调用方应先检查 `error`，再访问 `Info`，不要只根据错误文本进行业务分类。当前入口不会保留初始化过程的具体错误，排查时需同时检查数据目录和文件。

需要使用自己的 CZDB 文件、显式控制内存或释放搜索器时，参考 [czdb 底层接口](czdb/README.md)。文件操作封装见 [file](../file/README.md)，全部工具见 [util 索引](../README.md)。
