# hyperHeaderDecoder：CZDB 加密头解析

该包是 [CZDB 搜索器](../README.md) 初始化时使用的格式解析环节，负责从流中读取超级头并解密其客户信息。

```go
import "server/app/util/ip/czdb/hyperHeaderDecoder"
```

## 接口与处理过程

```go
func Decrypt(reader io.Reader, key string) (*entity.HyperHeaderBlock, error)
```

`reader` 必须位于完整 CZDB 文件的开头；`key` 是 Base64 编码的 AES 密钥。处理顺序为：

1. 读取 12 字节固定头，取得版本、客户标识、加密块长度。
2. 按头部长度读取加密块，解码密钥并解密。
3. 比较明文与加密块中的客户标识。
4. 按本地当前日期检查 `YYMMDD` 格式的有效期，早于今天则返回错误。

读取不足、密钥解码/解密失败、客户标识不符、已过期都会使初始化失败。当前代码读取版本字段，但不根据版本号做兼容性拒绝，不能仅凭调用成功认定任意版本均可查询。

返回结构的 `DecryptedBlock` 已设置。`EncryptedData` 字段未在此函数中填充。`GetHeaderSize()` 返回固定头、加密块与随机填充区的总长度，搜索器用这个值计算逻辑数据区起点。

## 读取头部示例

```go
package example

import (
	"fmt"
	"os"

	"server/app/util/ip/czdb/hyperHeaderDecoder"
)

func DescribeHeader(path, key string) error {
	reader, err := os.Open(path)
	if err != nil {
		return err
	}
	defer reader.Close()

	header, err := hyperHeaderDecoder.Decrypt(reader, key)
	if err != nil {
		return fmt.Errorf("解析 CZDB 头: %w", err)
	}
	fmt.Printf("版本=%d 客户标识=%d 逻辑数据区偏移=%d\n",
		header.GetVersion(), header.GetClientId(), header.GetHeaderSize())
	return nil
}
```

本函数消费固定头和加密块后不会继续读取随机填充区，也不会关闭传入流。后续读取数据区时，应根据总头长定位；不要直接把当前游标当作数据区起点。复用同一 `reader` 进行并发解析需要外部同步。

头部长度、解密后的数组边界没有覆盖所有畸形输入场景，底层 AES 填充处理也可能 panic；用于项目兼容的可信数据库文件，不作为任意文件上传的校验器。密钥不能记录到日志中。

结构说明见 [entity](../entity/README.md)，密码格式兼容实现见 [aesECB](../aesECB/README.md)。
