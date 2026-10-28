# jwt：面板登录令牌

导入路径：`server/app/util/jwt`。本包用于面板现有的 JWT 登录协议，封装用户登录信息、令牌生成和有效性校验。它不负责用户数据库查询、退出登录或刷新令牌。

## API 与数据

| API/类型 | 说明 |
| --- | --- |
| `GetToken(user *User, expireTime int) string` | 使用 HS256 签名；时间单位为秒，签名失败返回空字符串 |
| `CheckToken(token string) *MyClaims` | 校验失败返回 nil；成功返回载荷 |
| `User` | 包含 `LoginToken`、`LoginIp`、`LoginTime` |
| `MyClaims` | 包含 `User *User`，并嵌入第三方 JWT 的 `RegisteredClaims` |

`expireTime == 0` 使用 `constant.JwtExpireTime`，当前为 86400 秒；正数表示从现在起经过多少秒过期；负数不设置 `exp`。业务通常应传入正数或使用默认值，避免无过期令牌。

JWT 载荷使用签名保护完整性，不加密；不要把密码、云密钥或其他不能公开的值放入载荷。当前载荷没有自动填充 issuer、audience、subject 等跨服务约束，不应直接当作多个系统之间的通用鉴权协议。

## 示例：生成并校验

```go
package example

import (
	"errors"
	paneljwt "server/app/util/jwt"
	"server/app/util/str"
	"time"
)

func IssueLogin(loginToken, ip string) (string, error) {
	token := paneljwt.GetToken(&paneljwt.User{
		LoginToken: loginToken,
		LoginIp:    ip,
		LoginTime:  str.DateTime(time.Now()),
	}, 3600)
	if token == "" {
		return "", errors.New("生成登录令牌失败")
	}
	return token, nil
}

func ReadLogin(token string) (*paneljwt.User, error) {
	claims := paneljwt.CheckToken(token)
	if claims == nil || claims.User == nil {
		return nil, errors.New("登录令牌无效或已过期")
	}
	// 此处还需查询业务会话，确认 LoginToken 尚未被撤销。
	return claims.User, nil
}
```

这里接收的是原始 token，HTTP `Authorization` 的 `Bearer ` 前缀应由调用方去掉。`CheckToken` 不查询会话、不比较请求 IP，也不检查账户是否被禁用；这些规则需在认证中间件中执行。

## 密钥与部署行为

密钥第一次使用时由 `constant.JwtKey` 加机器标识派生，通过 SHA-256 得到字节值并缓存在进程中：

- Linux 读取 DMI UUID 或 machine-id 文件。
- macOS 调用系统 `ioreg` 读取平台 UUID。
- Windows 调用 `reg.exe` 读取 MachineGuid。
- 读取失败时退回主机名。

因此机器迁移、机器标识变化、常量修改都可能让旧令牌失效；多机部署不能假设一台机器签发的令牌能被其他机器接受。进程内使用 `sync.Once` 初始化，不会在每次验证时重新探测机器。

## 错误和适用范围

当前接口把详细解析错误折叠为 nil，把签名错误折叠为空字符串。需要区分过期、签名错误、算法约束等情况时，应在新需求中设计明确错误接口，而不是解析提示文字。生成函数固定 HS256；校验函数没有额外配置 `WithValidMethods` 等选项，本文不承诺严格固定算法的验证策略。

`User.LoginTime` 使用项目的 [str.DateTime](../str/README.md)，序列化为没有时区偏移的日期字符串。相关：[context](../context/README.md)。
