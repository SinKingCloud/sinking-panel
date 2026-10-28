# entity：CZDB 头部、索引与数据块

本包定义 [CZDB](../README.md) 二进制布局对应的结构，供搜索器和头部解码器使用。业务获取结构化地区信息时应使用 [ip.Query](../../README.md)。

```go
import "server/app/util/ip/czdb/entity"
```

## 头部结构

`HyperHeaderBlock` 的公开字段为：

```go
type HyperHeaderBlock struct {
    Version            int64
    ClientId           int64
    EncryptedBlockSize int
    EncryptedData      []byte
    DecryptedBlock     *DecryptedBlock
}
```

各字段都有同名 `Get...` / `Set...` 方法，setter 不返回链式对象。

| 方法/常量 | 行为 |
| --- | --- |
| `HeaderSize` | 固定头长度，12 字节 |
| `(*HyperHeaderBlock).ToBytes() []byte` | 仅序列化版本、客户标识和加密块长度三个 4 字节字段 |
| `FromBytes(bytes []byte) *HyperHeaderBlock` | 从前 12 字节读取上述字段，不进行解密和有效性校验 |
| `(*HyperHeaderBlock).GetHeaderSize() int` | `12 + EncryptedBlockSize + DecryptedBlock.GetRandomSize()` |

调用 `GetHeaderSize` 前必须初始化 `DecryptedBlock`；`ToBytes` 不会自动附加 `EncryptedData` 或随机填充区，因此返回值不是完整数据库头。

`DecryptedBlock` 保存客户标识、到期日期、随机填充长度，使用 `GetClientId/SetClientId`、`GetExpirationDate/SetExpirationDate`、`GetRandomSize/SetRandomSize` 访问，参数和 getter 返回值均为 `int`。其 `ToBytes() []byte` 输出 16 字节：客户标识使用 12 位，到期日期使用 20 位（十进制 `YYMMDD`），随机区长度使用 32 位，其余为保留字节。

- `(*DecryptedBlock).Encrypt(data []byte, key string) ([]byte, error)`：用 Base64 解码后的 AES 密钥加密给定字节。
- `(*DecryptedBlock).ToEncryptedBytes(key string) ([]byte, error)`：加密本结构的序列化结果。
- `Decrypt(key string, encryptedBytes []byte) (*DecryptedBlock, error)`：解密块并还原字段；客户标识匹配和到期校验在 [hyperHeaderDecoder](../hyperHeaderDecoder/README.md) 中进行。

## 索引结构

| 结构与构造函数 | 字段访问 | 序列化 |
| --- | --- | --- |
| `NewHeaderBlock(indexStartIp []byte, indexPtr int) *HeaderBlock` | `Get/SetIndexStartIp`、`Get/SetIndexPtr` | `GetBytes()` 返回 20 字节：16 字节 IP 区 + 4 字节指针 |
| `NewIndexBlock(startIp, endIp []byte, dataPtr, dataLen int, dbType string) *IndexBlock` | `Get/SetStartIp`、`Get/SetEndIp`、`Get/SetDataPtr`、`Get/SetDataLen` | `GetBytes()` 包含起止地址、4 字节数据指针和 1 字节长度 |

这两个结构的 setter 返回自身指针，可以链式设置。`HEADER_LINE_SIZE` 为 `20`。`GetIndexBlockLength(dbType string) int` 对 `constant.IPV4` 返回 13，对 IPv6 返回 37；只应传入有效的 `IPV4`/`IPV6` 常量，其他字符串在不同方法中的处理并不一致。

```go
package example

import (
	"fmt"
	"net"

	"server/app/util/ip/czdb/constant"
	"server/app/util/ip/czdb/entity"
)

func ShowIndexLayout() {
	start := net.ParseIP("192.0.2.0").To4()
	end := net.ParseIP("192.0.2.255").To4()
	header := entity.NewHeaderBlock(start, 1024)
	index := entity.NewIndexBlock(start, end, 2048, 32, constant.IPV4)
	fmt.Println(len(header.GetBytes()))                    // 20
	fmt.Println(len(index.GetBytes()))                     // 13
	fmt.Println(entity.GetIndexBlockLength(constant.IPV6)) // 37
}
```

这里是结构编码演示，不包含有效的地区数据，不能单独作为 CZDB 文件使用。IPv4 地址应传入 `To4()` 的 4 字节结果，IPv6 地址使用 16 字节；指针和长度需符合磁盘格式的位宽。尤其 `dataLen` 虽然使用 `int`，序列化字段只有 1 字节，超过 255 会截断。

## 地区数据块

```go
func NewDataBlock(region []byte, dataPtr int) *DataBlock
func (db *DataBlock) GetRegion(geoMapData []byte, columnSelection int64) string
```

`region` 是数据库中的 MsgPack 数据，不是普通 UTF-8 地区字符串。`GetRegion` 读取地区映射位置和附加数据，根据 `columnSelection` 的位选择映射列，使用制表符连接输出；映射位置为零时直接返回附加数据。它没有 `error` 返回值，部分解码错误会返回字符串 `"null"`。

`SetRegion(region []byte) *DataBlock`、`GetDataPtr() int`、`SetDataPtr(dataPtr int) *DataBlock` 用于修改数据块。正常查询时由 `DbSearcher` 提供映射数据和列掩码，不需要业务手工拼装。

## 边界与生命周期

结构不持有需要关闭的资源，但包含调用方传入的切片；不要假定构造或 setter 会复制数据。读写同一对象、复用或修改底层切片时需自行同步。二进制读写、位宽和部分 MsgPack 索引操作缺少完整检查，畸形长度可能 panic；解析前应确保文件来源和格式可靠。

小端读写见 [byteUtil](../byteUtil/README.md)，格式常量见 [constant](../constant/README.md)，完整查询流程见 [CZDB](../README.md)。
