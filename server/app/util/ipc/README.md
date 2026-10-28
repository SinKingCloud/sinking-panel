# ipc：本机进程通信

导入路径：`server/app/util/ipc`。此包创建经过操作系统身份校验的本机连接，供守护进程、控制命令和插件宿主通信使用。它不监听 TCP 端口，不提供消息格式、请求路由或 RPC 序列化。

相关工具：[daemon 服务控制](../daemon/README.md)、[plugin 插件管理](../plugin/README.md)。

## 公共 API

```go
func Listen(path string) (net.Listener, error)
func Dial(ctx context.Context, path string) (net.Conn, error)
```

`path` 是双方约定的通信目录，不是 `control.sock` 文件名，也不是 TCP 地址。服务端、客户端必须使用相同的目录标识。路径会转为绝对路径并解析已存在父目录的软链接。

- `Listen` 返回标准 `net.Listener`，调用方负责 `Accept` 循环与 `Close`。
- `Dial` 的 context 控制连接阶段；连接建立后的读写应另外设置 deadline，或在 context 取消时关闭连接。
- 连接是字节流。业务协议必须自行约定分帧、最大消息长度、超时和响应格式。
- `ErrUnavailable` 仅表示没有服务监听。通过 `errors.Is(err, ipc.ErrUnavailable)` 判断；权限不足、身份不匹配、服务启动/退出过渡中的锁占用均不是正常离线。

## 示例：一次本机请求与响应

下例创建临时通信目录，在同一程序中演示服务端与客户端。它只处理一个请求；生产服务应使用持续 Accept 循环，并限制每个请求的大小与执行时间。

```go
package example

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "time"

    "server/app/util/ipc"
)

func LocalRoundTrip(ctx context.Context) (string, error) {
    directory, err := os.MkdirTemp("", "panel-ipc-")
    if err != nil {
        return "", err
    }
    defer os.RemoveAll(directory)

    listener, err := ipc.Listen(directory)
    if err != nil {
        return "", err
    }
    defer listener.Close()

    go func() {
        connection, err := listener.Accept()
        if err != nil {
            return
        }
        defer connection.Close()
        _ = connection.SetDeadline(time.Now().Add(5 * time.Second))
        request, err := bufio.NewReader(connection).ReadString('\n')
        if err == nil {
            _, _ = fmt.Fprint(connection, "received: "+request)
        }
    }()

    connection, err := ipc.Dial(ctx, directory)
    if err != nil {
        return "", err
    }
    defer connection.Close()
    if err = connection.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
        return "", err
    }
    if _, err = fmt.Fprintln(connection, "status"); err != nil {
        return "", err
    }
    return bufio.NewReader(connection).ReadString('\n')
}
```

离线判断示例，`directory` 为约定的服务目录：

```go
connection, err := ipc.Dial(ctx, directory)
if errors.Is(err, ipc.ErrUnavailable) {
    // 只有此分支可以按照“服务未启动”处理。
    return nil
}
if err != nil {
    return err
}
defer connection.Close()
```

该片段需导入标准 `errors` 包并在一个返回 `error` 的函数中使用。

## Unix 系统行为

支持 Linux、macOS、FreeBSD、OpenBSD、NetBSD、DragonFly。目录内使用 `control.sock` 和 `control.lock`：

- 目录权限为 `0700`，套接字和锁文件为 `0600`。
- 文件锁阻止两个实例同时监听；锁文件可能留在磁盘上，文件存在不代表服务仍在运行。
- 双方均校验对端内核 UID，要求与当前有效 UID 一致。另一个账户或 root 以不同 UID 直接访问普通账户服务也不会被视为同一身份。
- 校验路径各级属主和可写权限；可信账户持有的 sticky 临时目录可作为祖先，最终目录仍必须归当前账户所有。
- 超长路径会映射至 `/tmp/sinking-ipc-<uid>-<digest>` 的短目录，调用方仍传原路径，不要自行拼接实际 socket 名。
- 关闭 listener 时只删除本实例创建且未被替换的 socket，之后释放锁；不要绕过 listener 手动删除运行中的 socket 或 lock。

## Windows 行为

Windows 使用由规范路径摘要生成的本机命名管道。路径大小写会归一化，不需要创建真实的 socket 文件。

非提权服务允许本账户及受信任的提权主体；提权服务只允许 SYSTEM 和已提权管理员。程序会校验已连接管道的对端 PID、真实令牌及服务端管道属主，不信任消息中自报的账户或 PID。UAC 过滤令牌不会被当作已提权管理员。

其他系统返回“不支持本机通信”的错误，不会退回 TCP。当前实现属于平台相关代码，变更后应分别检查 Unix 和 Windows 构建。

## 集成注意事项

`Accept` 会丢弃未经授权的连接并继续等待，不把单个非法连接当成整个服务退出。服务关闭时调用 `listener.Close()` 解除阻塞；客户端读写 deadline 到期后应关闭连接，不要无限重试。相同目录表示同一个服务端点，多个独立服务要分配不同目录。
