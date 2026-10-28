# validator：结构体参数校验

导入路径：`server/app/util/validator`。对 `go-playground/validator/v10` 做轻量封装，提供中文错误信息，并从字段 `label` 标签读取用户可读名称。适合 service 层或不依赖 HTTP 的结构体输入检查。

## API 和标签

唯一业务入口是 `Check(obj interface{}) (bool, string)`。第一次调用通过 `sync.Once` 初始化校验器和中文翻译器，后续共用实例。

| 标签 | 用途 | 示例 |
| --- | --- | --- |
| `validate` | 字段约束 | `required,min=1,max=100` |
| `label` | 中文错误中的字段名称 | `label:"任务名称"` |
| `json` | JSON 字段映射，由编码/绑定层使用 | `json:"name"` |
| `default` | 本包不处理默认值；由 HTTP 绑定等上层处理 | `default:"20"` |

验证失败时返回 `false` 和第一条翻译后的错误，不收集所有字段错误。没有失败则返回 `true, ""`。

## 完整示例

```go
package example

import (
    "errors"
    "strings"
    "server/app/util/validator"
)

type CreateTask struct {
    Name string `json:"name" validate:"required,max=100" label:"任务名称"`
    Retry int `json:"retry" validate:"min=0,max=10" label:"重试次数"`
    Email string `json:"email" validate:"omitempty,email" label:"通知邮箱"`
}

func ValidateTask(input *CreateTask) error {
    if input == nil {
        return errors.New("任务参数不能为空")
    }
    input.Name = strings.TrimSpace(input.Name)
    if ok, message := validator.Check(input); !ok {
        return errors.New(message)
    }
    return nil
}
```

`required` 不负责去掉字符串前后空格；需要禁止纯空格时，应像示例一样先规范化。数值字段的 `required` 会拒绝零值，允许零值的数值通常只使用 `min=0` 等约束。可选字段用 `omitempty` 跳过零值后续检查。

## 输入和错误边界

必须传入有效结构体或非 nil 结构体指针。当前实现对底层 `InvalidValidationError` 不返回失败，因此不能用 `Check(nil)` 或 `Check("abc")` 来判断输入类型是否正确；类型/空指针检查由调用方负责。未知的校验规则属于开发期错误，应按底层 validator 的标签规范使用，不应由用户动态拼接规则。

本包不会解析 JSON、读取 HTTP 请求、填充默认值或修改字段。HTTP 控制器中若还需要参数绑定，使用 [context](../context/README.md) 的 `ValidatorAll`、`ValidatorJson` 等方法。业务一致性检查（名称唯一、权限、数据库记录存在等）仍应放在 service 层。

源码：[validator.go](validator.go)。
