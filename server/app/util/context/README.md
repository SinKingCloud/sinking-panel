# context：面板 HTTP 上下文

导入路径：`server/app/util/context`。`Context` 嵌入 `sinking_web.Context`，补充参数验证、分页、统一响应、请求标识和当前用户。它与标准库 `context.Context` 不是同一种类型，建议分别使用 `appContext` 和 `context` 导入别名。

## 控制器示例

```go
package example

import appContext "server/app/util/context"

func Create(c *appContext.Context) {
    var input struct {
        Name string `json:"name" default:"" validate:"required,max=100" label:"名称"`
        Count int `json:"count" default:"1" validate:"min=1,max=100" label:"数量"`
    }
    if ok, message := c.ValidatorAll(&input); !ok {
        c.Error(message)
        return
    }
    c.SuccessWithData("创建成功", map[string]interface{}{
        "name": input.Name,
        "count": input.Count,
    })
}
```

注册到 Web 框架时使用 `appContext.HandleFunc(Create)`；它返回 `func(*sinking_web.Context)`。项目注册方式可参考 [HTTP 路由](../../http/route/app.go)。耗时操作使用 `c.Request.Context()` 传递取消信号，不将整个 HTTP Context 保存到后台任务中。

## 参数绑定与校验

| 方法 | 输入来源 |
| --- | --- |
| `Validator(data)` | 只验证已有结构体，不绑定参数 |
| `ValidatorAll(data)` | 调用框架 BindAll；非表单且正文非空时再按 JSON 解码正文 |
| `ValidatorJson(data)` | JSON 正文 |
| `ValidatorPost(data)` | 表单 |
| `ValidatorGet(data)` | query string |
| `ValidatorPath(data)` | 路由参数 |

这些方法统一返回 `(ok bool, message string)`，绑定失败与标签校验失败均返回 false；调用方需要主动 `return`。输入必须是有效结构体指针。`validate` 和 `label` 标签由 [validator](../validator/README.md) 处理，`json`/`default` 标签的绑定行为由框架提供。

`ValidatorAll` 把非表单正文当成 JSON，因此纯文本、二进制上传不要走该入口，应直接使用框架正文/文件接口。需要严格指定来源时优先选对应的专用方法，减少查询参数与正文混用。

## 分页

`ValidatePage(defaultField, defaultType, cursorAllowField string, pageAllowFields ...string) *page.Query` 不返回 error，会规范化无效输入并回退默认值。

- 请求参数：`page`、`page_size`、`order_by_field`、`order_by_type`、`cursor_id`、`cursor_last_id`。
- `page > 0` 为普通分页，其他情况为游标分页。
- `page_size` 必须在 1..1000，否则回退 20。
- 排序方向只接受 `asc`/`desc`，大小写会归一化；默认方向也可以传空。
- 游标和普通分页可使用不同字段白名单，逗号或空白分隔。未传普通分页白名单时复用游标白名单。
- 白名单为空或 `*` 表示不限制；未知字段回退到 defaultField，因此默认字段也必须由开发者提供合法值。

控制器中的使用片段：

```go
query := c.ValidatePage("id", "desc", "id,create_time", "id,name,create_time")
// query 交给 repository.SelectPage，或业务层自己的查询函数。
```

参见 [page](../page/README.md)、[repository](../repository/README.md)。

## 统一响应格式

所有以下方法使用 HTTP 200 返回 JSON，业务状态通过 `code` 判断，不要只检查 HTTP 状态。

| 方法 | 业务码 |
| --- | --- |
| `Success(message)`、`SuccessWithData(message, data)` | 200 |
| `Error(message)`、`ErrorWithData(message, data)` | 500 |
| `TokenError(message, data)` | 503 |
| `NotLogin(message, data)` | 403 |
| `Return(code int32, message, data)` | 调用方指定 |

响应固定包含 `code`、`message`、`data`、`request_id`；data 为 nil 时输出空对象 `{}`。这些帮助方法不代替控制器的流程控制，发送错误响应后仍应显式 return，避免再次写入响应。

## 请求信息与用户

`GetRequestId()` 优先读取上下文中的请求 ID，再读对应请求头；没有则生成 ID 并写入上下文。该值用于链路关联，不是身份凭据。`GetRequestIp()` 先使用框架代理感知 IP，再回退直接客户端 IP，最后返回 `localhost`；可信代理范围应由部署/框架层配置。

`SetUserInfo(*jwt.User)` / `GetUserInfo()` 存取登录用户。未设置时返回零值用户指针，不表示已认证。存储位置固定用于 `*jwt.User`，不要往该键放其他类型。验证 token 和权限检查仍由认证中间件或业务层完成，见 [jwt](../jwt/README.md)。

## 默认值帮助方法

`GetStringWithDefault` 只在空字符串时回退。`GetIntWithDefault`、`GetInt64WithDefault`、`GetFloatWithDefault` 在解析失败或负数时回退；其中 Int64 当前调用 `strconv.ParseInt(value, 64, 10)`，base=64 超出允许范围，因此始终回退默认值；读取 ID 请直接使用 `strconv.ParseInt(value, 10, 64)`，不要依赖这个帮助方法。`GetBoolWithDefault` 仅把 `true`、`1` 识别为 true，其他非空值都为 false，不支持大小写归一化。

源码：[类型与路由适配](context.go)、[验证](validator.go)、[响应](response.go)、[默认值](default.go)。
