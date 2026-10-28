# constant：CZDB 格式常量

```go
import "server/app/util/ip/czdb/constant"
```

本包只声明常量，不读取配置、不持有资源。所有字节偏移均针对 [CZDB](../README.md) 逻辑数据区；物理文件前面还包含需要跳过的加密超级头。

| 常量 | 值 | 用途 |
| --- | --- | --- |
| `FILE_SIZE_PTR` | `1` | 文件长度字段起点 |
| `FIRST_INDEX_PTR` | `5` | 首个索引指针字段起点 |
| `HEADER_BLOCK_PTR` | `9` | 头部索引区信息字段起点 |
| `END_INDEX_PTR` | `13` | 最后索引指针字段起点 |
| `SUPER_PART_LENGTH` | `17` | 逻辑超级块的长度，单位字节 |
| `IPV4` | `"IPV4"` | IPv4 数据类型 |
| `IPV6` | `"IPV6"` | IPv6 数据类型 |
| `BTREE` | `"BTREE"` | 文件索引查询模式 |
| `MEMORY` | `"MEMORY"` | 全库内存查询模式 |

查询模式用于 `czdb.NewDbSearcher`；数据类型用于 `entity.NewIndexBlock`、`entity.GetIndexBlockLength`，两类常量不能混用。当前没有可用的 `BINARY` 模式常量或查询实现。

## 小端字段示例

```go
package example

import (
	"fmt"

	"server/app/util/ip/czdb/byteUtil"
	"server/app/util/ip/czdb/constant"
)

func ShowFileSizeField() {
	header := make([]byte, constant.SUPER_PART_LENGTH)
	byteUtil.WriteIntLong(header, constant.FILE_SIZE_PTR, 1024)
	fmt.Println(byteUtil.GetIntLong(header, constant.FILE_SIZE_PTR)) // 1024
}
```

这里仅演示一个字段的编码，生成的切片不是可查询的数据库。真实文件必须同时包含合法的加密头、索引和数据块。

返回 [CZDB 接口](../README.md)，查看 [byteUtil](../byteUtil/README.md) 与 [entity](../entity/README.md)。
