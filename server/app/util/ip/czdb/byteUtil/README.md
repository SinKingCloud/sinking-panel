# byteUtil：CZDB 小端整数读写

此包供 [CZDB](../README.md) 的头部和索引编解码使用，直接在调用方字节切片上读写整数，不分配数据库对象。

```go
import "server/app/util/ip/czdb/byteUtil"
```

| API | 字节数与行为 |
| --- | --- |
| `Write(b []byte, offset int, v int64, bytes int)` | 从 `offset` 开始按小端顺序写入 `v` 的低 `bytes` 个字节 |
| `WriteIntLong(b []byte, offset int, v int64)` | 固定写入低 4 字节；名字中的 Long 不代表 8 字节 |
| `GetIntLong(b []byte, offset int) int64` | 读 4 字节无符号小端整数，以 `int64` 返回 |
| `GetInt3(b []byte, offset int) int` | 读 3 字节 |
| `GetInt2(b []byte, offset int) int` | 读 2 字节 |
| `GetInt1(b []byte, offset int) int` | 读 1 字节 |

```go
package example

import (
	"fmt"

	"server/app/util/ip/czdb/byteUtil"
)

func EncodeField() {
	data := make([]byte, 8)
	byteUtil.WriteIntLong(data, 0, 0x12345678)
	byteUtil.Write(data, 4, 0x010203, 3)
	fmt.Printf("%x\n", data[:4])                     // 78563412
	fmt.Printf("%x\n", byteUtil.GetIntLong(data, 0)) // 12345678
	fmt.Printf("%x\n", byteUtil.GetInt3(data, 4))    // 10203
}
```

所有方法都没有 `error` 返回值，也没有完整参数检查。调用前必须保证偏移非负、缓冲区足够、写入宽度符合协议；越界会 panic。超出字段位宽的高位会被舍弃，读操作不做有符号扩展。`Write` 修改原切片，共享缓冲区的并发读写需要调用方同步。

这些方法表达 CZDB 的磁盘布局，不等同于 Go 平台相关的 `int` 内存表示。格式偏移见 [constant](../constant/README.md)，使用这些字段的结构见 [entity](../entity/README.md)。
