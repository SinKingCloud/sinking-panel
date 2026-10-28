# aesECB：CZDB 兼容加解密

本包实现 CZDB 格式需要的 AES-ECB 和填充操作，由 [entity](../entity/README.md) 与数据库解密流程使用。

```go
import "server/app/util/ip/czdb/aesECB"
```

## API

| API | 说明 |
| --- | --- |
| `AesEncrypt(origData, key []byte) ([]byte, error)` | 为明文添加填充，再使用 AES-ECB 加密 |
| `AesDecrypt(crypted, key []byte) ([]byte, error)` | 解密完整密文块并移除填充 |
| `NewECBEncrypter(b cipher.Block) cipher.BlockMode` | 创建不自动填充的分组加密器 |
| `NewECBDecrypter(b cipher.Block) cipher.BlockMode` | 创建不自动移除填充的分组解密器 |
| `PKCS5Padding(ciphertext []byte, blockSize int) []byte` | 按传入块大小补齐数据，已对齐时也补一个块 |
| `PKCS5UnPadding(origData []byte) []byte` | 根据末字节截去填充，不做完整校验 |

AES 密钥必须是 16、24 或 32 字节；这里接受原始密钥字节，数据库头中 Base64 密钥的解码由上层完成。虽然填充函数命名为 PKCS5，实际按给定块大小工作，用于 AES 时块大小为 16。

## 本地往返示例

下面只展示同一段代码产生并读取合法密文的过程。示例密钥仅用于演示。

```go
package example

import (
	"fmt"

	"server/app/util/ip/czdb/aesECB"
)

func EncryptionRoundTrip() error {
	key := []byte("0123456789abcdef")
	ciphertext, err := aesECB.AesEncrypt([]byte("CZDB example"), key)
	if err != nil {
		return err
	}
	plaintext, err := aesECB.AesDecrypt(ciphertext, key)
	if err != nil {
		return err
	}
	fmt.Println(string(plaintext))
	return nil
}
```

## 调用限制

本包用于既有数据库格式兼容，ECB 不提供消息认证，也不适合作为新备份文件或新业务凭据的加密方案。

`AesDecrypt` 的 `error` 主要覆盖密钥创建失败；空密文、长度不是块大小整数倍、损坏填充等情况可能 panic。`PKCS5UnPadding` 本身没有长度和填充值检查，不能仅以“没有 error”判断不可信数据有效。调用底层 `CryptBlocks` 时，源数据必须按块对齐，目标缓冲区必须足够长。

加解密函数没有全局状态或需要关闭的资源。输入缓冲区和底层分组模式的并发复用仍需由调用方管理。

参见 [头部解码](../hyperHeaderDecoder/README.md) 和 [CZDB 总说明](../README.md)。
