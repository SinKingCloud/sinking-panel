# webssh：SSH 命令与交互终端

本包封装 SSH 登录、单次命令执行和供 WebSocket 转发的交互终端。它不直接创建 WebSocket，HTTP 鉴权、输入转发、连接关闭和前端消息协议由调用方处理。

```go
import "server/app/util/webssh"
```

## 创建与认证

```go
func NewSshClient(ip string, port int, timeout time.Duration) *SshClient
```

`ip` 可传服务器 IP 或主机名；`port` 要显式传入端口（通常为 `22`），没有自动补全默认值。`timeout` 传给 SSH 客户端连接配置，不是远程命令或交互会话的执行时限。

| 方法 | 说明 |
| --- | --- |
| `AuthWithPassword(user, password string) error` | 密码登录并建立连接 |
| `AuthWithPrivateKey(user, key string) error` | 解析私钥文本后登录；这里传密钥内容，不是文件路径 |
| `Auth(user string, method ssh.AuthMethod) error` | 使用调用方提供的单个认证方式建立连接 |
| `Close() error` | 关闭底层 SSH 连接；未连接或 nil 接收者时返回 nil |

客户端公开字段有 `Addr`、`Port`、`User`、`Password`、`Key`、`Timeout`。不要把包含密码/私钥的完整结构打印到日志。普通私钥入口使用 `ssh.ParsePrivateKey`，没有口令参数；带口令私钥可由调用方使用 `ssh.ParsePrivateKeyWithPassphrase` 解析，再通过 `Auth(user, ssh.PublicKeys(signer))` 认证。

当前 `Auth` 固定使用 `ssh.InsecureIgnoreHostKey()`，没有验证服务器主机公钥，也没有开放主机公钥校验配置。连接到需要校验服务器身份的环境时，需先扩展这部分能力。配置和认证过程没有并发保护，不要同时修改客户端字段或对同一实例并发重连；重新认证前应先释放旧连接。

## 执行单次命令

`Exec(cmd string) (string, error)` 为每条命令创建新 SSH session，返回标准输出，并在结束时关闭该 session；客户端连接可以继续复用。

```go
package example

import (
	"fmt"
	"time"

	"server/app/util/webssh"
)

func RemoteHostname(host, user, password string) (string, error) {
	client := webssh.NewSshClient(host, 22, 10*time.Second)
	if err := client.AuthWithPassword(user, password); err != nil {
		return "", fmt.Errorf("SSH 登录失败: %w", err)
	}
	defer client.Close()

	result, err := client.Exec("hostname")
	if err != nil {
		return "", fmt.Errorf("读取远程主机名: %w", err)
	}
	return result, nil
}
```

必须在认证成功后调用 `Exec` 或 `NewSession`，否则底层连接为空，可能 panic。`Exec` 不收集标准错误，也不提供完整合并输出；命令返回错误时当前实现返回空字符串，不保留已产生的部分标准输出。

`Exec` 没有 `context.Context` 参数。建立连接时的 `Timeout` 不能用来限制命令执行时间；要中止长时间执行的会话，需由调用方管理会话或关闭连接。命令字符串会交给远端 shell，拼入路径、文件名或用户参数前必须做正确转义。

## 交互终端 API

```go
func (s *SshClient) NewSession(height, width int) (*SshSession, error)
```

顺序是**行数 height、列数 width**。创建时申请 `xterm` PTY 并启动 shell，标准输出和标准错误写入同一个缓冲区。

| 方法 | 说明 |
| --- | --- |
| `Write(payload []byte) error` | 向远端 shell 标准输入写入终端字节 |
| `Read() []byte` | 立即取走一段已缓冲输出，最多 32 KiB；没有数据时返回 nil |
| `OutputReady() <-chan struct{}` | 有可读数据时发出通知，容量为 1，多次到达可合并 |
| `Wait() error` | 等待远端 shell 退出，不主动读取输出 |
| `Resize(height, width int) error` | 调整 PTY 行列数 |
| `Close() error` | 关闭输出缓冲写入和底层 session，唤醒等待缓冲空间的写入方 |

`Read() == nil` 只表示此刻缓冲区为空，**不表示会话已经退出**。退出应通过 `Wait` 判断。`OutputReady` 通道不会在会话关闭后自动关闭，不能只用 `range` 等待它结束。

单个会话的输出缓冲上限为 1 MiB；达到上限时后台写入会等待，必须持续消费输出，否则远端进程可能受阻。`Read` 返回独立的字节副本，读取后可交给 WebSocket 发送。当前内置锁保护输出缓冲，但应用层对输入写入、会话关闭和消息顺序仍应明确协调。

## 输出转发与取消示例

下面封装一个已经登录的客户端的交互会话，读取输出并在调用方取消时释放会话。WebSocket 场景可将 `output` 换成自己的发送适配器，并把收到的输入通过 `session.Write` 转发。

```go
package example

import (
	"context"
	"io"

	"server/app/util/webssh"
)

func StreamTerminal(ctx context.Context, client *webssh.SshClient, output io.Writer) error {
	session, err := client.NewSession(24, 80)
	if err != nil {
		return err
	}
	defer session.Close()

	done := make(chan error, 1)
	go func() { done <- session.Wait() }()

	// 演示发送输入。实际 WebSocket 转发时，由输入循环调用 Write。
	if err := session.Write([]byte("hostname\nexit\n")); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-session.OutputReady():
			if data := session.Read(); len(data) > 0 {
				if _, err := output.Write(data); err != nil {
					return err
				}
			}
		case err := <-done:
			// shell 已退出后，发送剩余缓冲数据。
			for {
				data := session.Read()
				if len(data) == 0 {
					return err
				}
				if _, writeErr := output.Write(data); writeErr != nil {
					return writeErr
				}
			}
		}
	}
}
```

调用者负责客户端连接的生命周期；退出终端不一定要关闭可复用的客户端。上述示例的取消通过 `defer session.Close()` 生效；`output.Write` 本身若阻塞，需由该输出实现提供超时或可关闭机制。实际 WebSocket 实现还需在浏览器断开、认证过期或写入失败时取消 context 并关闭会话。

## 错误分类

认证入口会将部分连接错误包装为可使用 `errors.Is` 判断的哨兵错误：

| 错误 | 含义 |
| --- | --- |
| `ErrAuthentication` | 用户名、密码或公钥认证未通过 |
| `ErrInvalidPrivateKey` | 私钥内容解析失败 |
| `ErrConnectionTimeout` | 建立连接超时 |
| `ErrConnectionRefused` | 目标端口拒绝连接 |
| `ErrHostUnreachable` | DNS 解析失败、网络或路由不可达 |
| `ErrHandshake` | SSH 协议握手失败 |

例如 `errors.Is(err, webssh.ErrAuthentication)` 可用于区分登录失败。分类依赖底层错误类型和文本匹配，未匹配的错误会原样返回；会话创建、执行、读写等错误也不一定属于这些分类。保留 `error` 的包装链，用户界面可展示友好原因，排查日志记录必要上下文即可。

调用本包前还需确保目标可达、SSH 服务已开启、账号有执行命令或启动 shell 的权限；交互终端要求服务器允许 PTY。返回 [util 索引](../README.md)。
