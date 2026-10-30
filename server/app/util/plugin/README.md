# plugin 使用说明

`plugin.Manager` 使用 Yaegi 加载 Go 插件，统一管理插件包、解释器缓存、方法调用和前端资源。数据库记录、安装任务、容器等业务由调用方处理，再通过 `Options` 注入插件需要的能力。

插件直接交付源码：Go 后端由 Yaegi 解释，面板中的前端加载器使用 Babel Standalone 在浏览器中转换 React TSX。修改文件后压缩为 ZIP 即可，无需手动构建插件；转换结果由浏览器缓存。

宿主程序导入 `server/app/util/plugin`；插件代码导入解释器提供的 `sdk`。两者的 `Plugin` 是同一个类型。

本文面向宿主侧集成，说明通用 `Manager` 的行为。面板软件插件开发请先执行 `./server plugin init demo`，生成的 `README.md` 是包含完整调用示例的开发手册。通用管理器默认清单名为 `plugin.json`；面板软件服务通过 `Options.ManifestName` 使用 `software.json`，不要混用。

## 1. 编写一个插件

最小目录结构：

```text
demo/
├── plugin.json
└── server/
    └── plugin.go
```

`plugin.json`：

```json
{
  "id": "demo",
  "name": "示例插件",
  "version": "1.0.0"
}
```

`server/plugin.go`：

```go
package demo

import "sdk"

func New() *sdk.Plugin {
    return &sdk.Plugin{
        Actions: map[string]func(*sdk.Context){
            "hello": func(c *sdk.Context) {
                var form struct {
                    Name string `json:"name" default:"访客" validate:"required,max=100" label:"姓名"`
                }
                if ok, message := c.ValidatorAll(&form); !ok {
                    c.Error(message)
                    return
                }
                c.SuccessWithData("操作成功", map[string]string{"message": "你好，" + form.Name})
            },
        },
    }
}
```

后端代码固定放在插件根目录的 `server/` 中，无需在清单配置。同目录可以有多个 `.go` 文件，使用同一个包名，入口包不能是 `main`。`New` 只创建对象并注册方法；安装、卸载、升级分别注册在 `Install`、`Uninstall`、`Upgrade` 字段中；异步任务的启动和回收使用 `Start`、`Stop`，普通接口注册到 `Actions`。

业务接口使用面板控制器的 `func(c *sdk.Context)` 写法。`c.ValidatorAll` 绑定并校验 JSON、查询或表单参数，`c.SuccessWithData`、`c.Success`、`c.Error` 直接返回面板标准响应。耗时操作使用 `c.Request.Context()`；无需自行序列化或解析参数字符串。

生命周期继续接收标准 `context.Context` 和 JSON 参数字符串，返回结果与错误：

```go
func New() *sdk.Plugin {
    return &sdk.Plugin{
        Install: func(ctx context.Context, params string) (interface{}, error) {
            // 按 params 安装软件，并将 ctx 传给耗时操作。
            return nil, nil
        },
        Upgrade: func(ctx context.Context, params string) (interface{}, error) {
            // params 包含 previous_version 和 version，在这里升级配置或插件数据。
            return nil, nil
        },
        Uninstall: func(ctx context.Context, params string) (interface{}, error) {
            return nil, nil
        },
        Actions: map[string]func(*sdk.Context){
            "info": func(c *sdk.Context) {
                c.SuccessWithData("操作成功", map[string]string{"status": "ready"})
            },
        },
    }
}
```

生命周期字段都可省略。宿主通过 `Lifecycle` 主动执行它们，未定义的钩子返回 `nil, nil`；`Call` 只查找 `Actions`，不会回退到生命周期。即使 `Actions` 注册了 `"install"`，它与 `Install` 也是两个独立入口。

生命周期匿名函数保留参数名，例如 `func(ctx context.Context, params string)`；当前解释器对完全省略参数名的写法存在类型转换问题。

需要子包或第三方依赖时，在 `server/` 放置 `go.mod`。管理器按模块路径解析 `server/` 下的子包，从 `server/vendor/` 读取插件额外依赖。只调用宿主提供的 SDK 和业务服务时不需要 `vendor`。使用 `./server plugin init demo` 可生成包含 SDK、后端和页面的完整示例。

使用 `./server plugin init` 创建的软件插件工程自带独立 `sdk/`，由当前面板二进制自动输出公开 API 声明，不引用私有面板源码。开发时可执行 `go mod tidy` 准备公开依赖，增加第三方库时执行 `go mod vendor` 并随包携带。服务器不运行 Go 命令。

将目录内的文件打成 ZIP，ZIP 根目录直接包含 `plugin.json`，不要多套一层 `demo/`：

```sh
(cd demo && zip -r ../demo.zip plugin.json server)
```

## 2. 在宿主中调用

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "time"

    webcontext "server/app/util/context"
    "server/app/util/plugin"

    "github.com/SinKingCloud/sinking-go/sinking-web"
)

func run() error {
    manager := plugin.NewManager("./data/plugins", plugin.Options{})
    defer manager.Close()

    pkg, err := manager.Import(context.Background(), "./demo.zip")
    if err != nil {
        return err
    }
    staging := pkg.Directory
    defer manager.Remove(staging)
    if !pkg.Available {
        return fmt.Errorf("插件无法运行：%s", pkg.UnavailableReason)
    }
    if err := manager.Rename(staging, pkg.Manifest.Id); err != nil {
        return err
    }
    pkg.Directory = pkg.Manifest.Id

    engine := sinking_web.Default()
    engine.ANY("/hello", webcontext.HandleFunc(func(c *webcontext.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
        defer cancel()
        if err := manager.Call(ctx, pkg.Directory, "hello", c); err != nil && c.StatusCode == 0 {
            c.Error(err.Error())
        }
    }))
    return http.ListenAndServe(":8080", engine)
}

func main() {
    if err := run(); err != nil {
        log.Fatal(err)
    }
}
```

实际服务中应长期复用一个 `Manager`，退出时再 `Close`，避免每次请求都创建解释器。

`Import` 返回已解包的临时 `Directory`，相对于管理器根目录；宿主用 `Rename` 将它移到正式目录，并负责清理未采用的临时包。软件服务使用插件标识作为固定目录，例如 `data/software/demo/`，不保留 `release-*` 版本层级。进程重启后可直接 `Load("demo")` 或 `Call` 自动加载，无需再次导入 ZIP。目录参数均使用管理器根目录内的相对路径。

## 3. Manager API

| 方法 | 用途 |
| --- | --- |
| `NewManager(root, options) *Manager` | 创建管理器，不读取或执行插件代码。 |
| `Exports(app) (interp.Exports, error)` | 从注入对象的公开字段和方法签名收集原生类型；配置 `Options.App` 后由管理器自动调用。 |
| `WriteSDK(directory) error` | 根据 `Options.App` 注入的类型输出独立开发 SDK，不读取宿主源码；软件服务的 `Init` 自动调用。 |
| `Import(path) (*Package, error)` | 将本机 ZIP 解包到临时目录并验证清单与运行条件，不执行代码；失败自动清理，成功后的移动或清理由宿主负责。 |
| `Rename(from, to) error` | 停止并释放相关解释器，在管理器根目录下重命名插件目录；两侧均为直接子目录，不覆盖已存在的目标。 |
| `Info(directory) (*Package, error)` | 读取清单和运行条件，不执行代码。 |
| `Manifest(directory) (*Manifest, error)` | 只读取并校验清单；用于卸载依赖检查及停止顺序，不要求前后端入口文件完整。 |
| `Load(directory) error` | 加载并缓存解释器，仅初始化 `New`；已加载时复用，不重复初始化，不执行生命周期。 |
| `Call(ctx, directory, method, c) error` | 自动加载后将真实 HTTP 上下文交给 `Actions[method]`；插件自行返回响应，加载或执行异常通过错误返回。 |
| `Lifecycle(ctx, directory, event, params) (interface{}, error)` | `install`、`upgrade`、`start` 会自动加载；`stop` 只使用已加载实例。`uninstall` 未加载时尝试只加载、不启动，加载失败记录日志并跳过钩子。未定义的钩子直接成功，其他事件名返回错误。 |
| `Unload(directory)` | 等待当前调用结束，停止后台任务后释放解释器，保留包文件；下次调用重新加载。 |
| `Remove(directory) error` | 停止后台任务、释放解释器并删除代码目录。不会调用 `uninstall`，也不负责删除业务数据。 |
| `Page(directory, path) ([]byte, string, string, error)` | 返回前端资源内容、MIME 类型和实际文件路径；`path` 为包内路径，如 `web/index.tsx`，省略后缀时在服务端解析文件和目录入口。HTTP 路由和身份验证由宿主处理。 |
| `Close()` | 停止接收新调用，取消异步任务，调用已启动实例的 `Stop` 并释放解释器；保留包文件和业务数据。 |

`Package` 包含 `Directory`、`Manifest`、`Available` 和 `UnavailableReason`。运行条件不满足时，导入和查看仍可成功，但不会执行初始化代码，`Load`、`Call` 及安装、升级、启动生命周期会返回错误；`stop` 不加载源码，`uninstall` 的加载失败只记录日志，不阻止删除代码目录。

`Call` 直接传递 HTTP 上下文，不包装请求参数或插件响应；业务参数由插件使用 `sdk.Context` 读取。只有 `Lifecycle` 的 `params` 使用 JSON 对象字符串，空字符串按 `{}` 处理。同一插件目录的业务接口和生命周期共用串行队列，等待可通过 `ctx` 取消；不同插件目录互不阻塞。`Info` 和 `Page` 不等待后台长任务。

`Call` 和 `Lifecycle` 没有默认业务超时，由调用方设置 `context`。`Call` 会把调用上下文传入 `c.Request.Context()`，插件应将它传给下游耗时操作；取消不会强制杀死忽略上下文的代码。解释代码的阶段有 30 秒超时，`New` 应快速返回。加载和调用中的同步 `panic` 会转成错误返回。

## 4. 清单与运行条件

| 字段 | 说明 |
| --- | --- |
| `id` | 必填。小写字母开头，可包含小写字母、数字、`_`、`-`，最长 64 个字符。 |
| `name` / `version` | 必填。显示名称和版本，分别最多 100、50 字节。 |
| `frontend` | 可选。前端入口必须位于 `web/` 内，例如 `web/index.tsx`。 |
| `icon` | 可选。包内图标路径，例如 `web/icon.svg`，必须位于 `web/`；支持 SVG、PNG、JPG/JPEG、WebP、ICO、GIF，文件不超过 8 MiB。不配置或设为空字符串时显示面板默认图标。 |
| `platforms` | 可选。Go 系统名，如 `linux`、`windows`、`darwin`；空数组表示不限制系统。 |
| `requires` | 可选。依赖的软件标识数组，填写其他软件清单的 `id`，例如 `["redis", "mysql"]`；空数组表示无软件依赖。 |

例如仅在 Linux 运行，并依赖 Redis 软件插件：

```json
{
  "platforms": ["linux"],
  "requires": ["redis"]
}
```

上面是附加到完整清单的字段。`requires` 不是系统能力、容器镜像或 Go module 名称，只表示其他软件插件的标识，不包含版本要求。需要直接使用容器服务的插件可以只声明 `platforms: ["linux"]`，无需声明 `container` 依赖。

通用 `Manager` 只校验依赖标识是否合法、是否包含自身，以及当前系统是否符合 `platforms`；不查询软件安装记录，也不替调用方启动依赖。独立使用此工具类时，依赖图和生命周期顺序由宿主负责，`Options` 配置清单名称、归档限额、宿主对象与原生符号。

当前面板的软件服务根据已安装软件检查依赖。安装、升级与启动需要依赖已安装且可用；依赖缺失会返回具体标识，不自动下载安装。自身依赖、非法标识及循环依赖均不允许。启动按依赖先后顺序执行，停止时反向处理；卸载依赖软件前，必须先卸载仍依赖它的软件。离线安装只检查已安装依赖，不启动插件；在线安装先启动依赖，面板启动或后续接口调用也会确保依赖先启动。


## 5. 注入宿主能力

`Options` 都是可选配置：

| 配置 | 说明 |
| --- | --- |
| `ManifestName` | 清单文件名，默认 `plugin.json`；软件服务使用 `software.json`。 |
| `MaxArchiveSize` / `MaxExpandedSize` | ZIP 与展开总大小上限（字节）；零值分别为 32 MiB、64 MiB。扩大上限时，额外空间只供 `images/` 镜像归档使用，源码始终最多 64 MiB。 |
| `Symbols` | 可选的额外原生符号，例如另行提供的包级函数或常量；`App` 及其业务类型自动导出，无需逐项填写。 |
| `App` | `func(id string) interface{}`，按清单 ID 创建并传给插件 `New` 的宿主对象。 |

管理器自动识别 `App` 的公开字段、方法及相关参数和返回类型，无需为每个业务类型编写 `Symbols`。例如宿主中定义：

```go
package main

import "server/app/util/plugin"

type App struct {
    ID string
}

func newManager() *plugin.Manager {
    return plugin.NewManager("./data/plugins", plugin.Options{
        App: func(id string) interface{} {
            return &App{ID: id}
        },
    })
}
```

插件通过 `sdk.App` 接收注入对象：

```go
package demo

import "sdk"

func New(app *sdk.App) *sdk.Plugin {
    return &sdk.Plugin{
        Actions: map[string]func(*sdk.Context){
            "info": func(c *sdk.Context) {
                c.SuccessWithData("操作成功", map[string]string{"id": app.ID})
            },
        },
    }
}
```

未配置 `App` 时使用 `func New() *sdk.Plugin`；配置后接收一个参数，参数类型须能接收 `App` 返回的对象。返回值必须是非空的 `*sdk.Plugin`。

自动发现处理的是注入对象可达的原生类型和方法，不会自动提供任意 Go 包的包级函数或常量。业务参数在独立 SDK 中提供对应类型，例如 `sdk.ModelScript`、`sdk.UtilPageQuery`，插件无需导入私有 `server` 模块。额外源码依赖的交付方式见前面的依赖说明。

## 6. 包更新与资源管理

`Import` 只准备临时包，软件服务按以下顺序更新固定目录：

1. 解包并验证新包，正式目录继续使用插件标识。
2. 将旧目录移至本次操作的临时备份，再把新包移入正式目录。
3. 首次安装先建立软件记录，再执行安装、启动生命周期；升级沿用原记录和配置。成功后更新版本并删除临时备份。
4. 失败时删除新代码、恢复旧目录及运行状态；首次安装删除新记录，升级恢复旧版本 Stop 完成后的配置，清理本次未使用的临时包。

软件服务直接使用插件标识 `Code` 作为固定目录，不从插件配置读取目录，也不迁移历史版本子目录。清理或恢复遇到文件系统错误会明确返回；无法恢复时保留临时备份供恢复，不删除唯一的旧代码。

宿主决定生命周期调用时机，例如软件导入成功前调用 `manager.Lifecycle(ctx, directory, "install", params)`，升级任务调用 `"upgrade"`。通用 `Manager` 的加载、导入不会自动安装或启动插件；释放已启动实例时会取消其后台上下文并调用 `Stop`。调用 `Remove` 前，可通过 `Lifecycle(..., "uninstall", ...)` 执行插件的业务卸载；未加载时尝试只加载源码、不调用 Start。加载失败会记录错误并跳过钩子，不阻塞后续删除代码目录。仅替换旧版本文件时无需执行卸载钩子。

通用管理器默认限制：ZIP 不超过 32 MiB，展开不超过 64 MiB，最多 1000 个 ZIP 条目（含目录）；拒绝路径越界、重复路径、符号链接及特殊文件。软件服务单独设置大包上限：镜像归档合计 16 GiB、源码合计 64 MiB，ZIP 总量另留 16 MiB 元数据空间。`images/` 仅接受 `.tar`、`.tar.gz`、`.tgz` 归档和 `.json` 清单，JSON 仍计入源码预算；打包和解包均流式复制，不将镜像读入内存。清单不超过 256 KiB。`Page` 支持 `web/` 下的 `.ts`、`.tsx`、`.jsx`、`.js`、`.mjs` 源码以及 `.json`、`.css`、`.png`、`.svg`、`.jpg`、`.jpeg`、`.webp`、`.ico`、`.gif` 静态资源，单个资源不超过 8 MiB。静态资源可以读取，不代表前端模块加载器支持 CSS 导入。

Manager 读取 `server/go.mod` 的模块名，将本模块和随包交付的 vendor 源码映射给 Yaegi；宿主 SDK 和业务类型由运行时直接提供，本地 `replace` 路径不会在服务器使用。不会在服务器自动下载依赖，也不能直接执行 CGO、汇编或解释器尚未支持的 Go 语法。插件在面板进程权限下运行，当前管理器不提供不可信代码沙箱。插件自己的 `fmt.Print` 等标准输出会被丢弃；业务日志使用宿主注入的日志能力。

## 7. 当前 software 接入

实际用法见 [业务服务初始化](../../service/service.go) 和 [SDK 管理器](../../service/software/sdk/manager.go)：清单使用 `software.json`，代码存储在 `constant.SoftwarePath`，注入软件 SDK。软件服务负责安装记录、依赖检查、卸载任务和版本切换；首次导入时用 `{}` 执行安装生命周期，成功后即可使用。配置表单与容器初始化由插件自己的页面和业务接口负责。

软件服务按清单 `id` 判断安装或升级：首次导入调用 `Install`，已存在时调用新包的 `Upgrade`，相同版本再次导入也按升级处理。新清单的依赖也会重新检查，不能通过升级引入缺失或循环依赖。升级参数为 `{"previous_version":"1.0.0","version":"1.1.0"}`。在线升级依次停止旧版本、执行新包 `Upgrade`，启用的软件再执行 `Start`，成功后更新版本并删除旧包，软件 ID、名称和业务配置保留。失败时恢复旧版本 `Stop` 完成后的内部配置与代码，再尝试恢复旧版本运行；插件需要自行处理业务表、公共配置组及外部资源迁移的回滚。

`app.InstanceID` 直接使用清单的 `id`。宿主不注入默认业务接口，容器启动、停止、重启、日志等功能均由插件在 `Plugin.Actions` 中自行注册，再调用 `app.Service.Container`。软件表的 `Config` 只保存插件内部配置，通过 `app.Service.Software.GetConfig(app.InstanceID)` 和 `SetConfig(app.InstanceID, data)` 原样读写字符串，可以为空，不强制 JSON。目录直接使用软件标识，依赖从已安装清单读取，不写入内部配置。首次安装在生命周期前建立记录，使 `Install`、`Start` 可以保存设置；失败时回滚。软件详情、列表都不返回 `Config`，安装接口只返回任务 ID。

软件插件使用 `sdk`，入口为 `func New(app *sdk.App) *sdk.Plugin`。`sdk.Plugin` 复用这里的 `Plugin`，`sdk.Context` 复用面板控制器上下文。接口路径为 `/software/<软件标识>/<方法名>`，沿用登录验证；控制器通过 `service.Software.Invoke(c, code, method)` 分发，插件负责校验参数和返回响应。

`/software/<软件标识>/page` 直接返回清单 `frontend` 指定的入口，无需传入 ID 或路径参数；多文件资源使用 `/software/<软件标识>/page/<资源路径>`，路径相对插件的 `web/` 目录，例如 `/software/demo/page/components.js` 对应 `web/components.js`。软件标识和路径各段需分别 URL 编码。`page` 是保留资源路由，沿用登录验证，不交给 `Actions` 处理。

列表、详情的 `icon` 返回清单中的包内路径，例如 `web/icon.svg`；宿主通过已有资源接口 `/software/demo/page/icon.svg` 获取图片，URL 中不再添加 `web/`。图标也需要登录认证，由宿主携带认证信息读取后显示。图标路径缓存在服务内存中，启动、安装及升级时更新，卸载时移除；首次查询缺失项只读取清单并缓存，不解释插件代码，不新增数据库字段。

省略 `icon` 或设置 `"icon": ""` 时，软件列表与默认管理标题使用应用图标（`AppstoreOutlined`）；图标加载期间、读取失败及图片解码失败也会回退，尺寸保持一致。生成的 demo 包含 `web/icon.svg`，并沿用宿主默认标题。插件自行设置 `modal.title` 会替换整个标题，宿主不会强行插入图标。删除图标文件前应同时移除清单配置，避免安装时因缺少声明的资源失败。

从面板二进制即可开始开发，不需要面板源码：

```sh
./server plugin init my-plugin ./my-plugin --name "我的插件"
```

`init` 调用软件服务的 `Init(directory, id, name)`，从嵌入二进制的 `public/plugin` 资源渲染 `software.json`、`server/` 和 `web/`，并根据当前程序公开的 API 自动生成独立 `sdk/`。模板文件使用 `.tmpl` 后缀，普通资源直接复制；新增模板文件会自动加入生成工程，无需在 Go 代码中拼接内容。目标目录必须尚不存在；省略目录时默认使用当前目录下的插件 ID。该命令不要求面板正在运行，也不要求管理员权限。

生成的 `server/go.mod` 通过 `replace sdk => ../sdk` 使用工程内 SDK。SDK 只包含公开类型和方法声明，不包含面板实现，也不依赖私有 `server` 模块；开发者可直接使用补全和类型检查。只编辑插件源码和前端即可，运行时由宿主注入真实服务实例。

从面板工作目录管理插件：

```sh
./server plugin list
./server plugin install /path/to/my-plugin
./server plugin install /path/to/my-plugin.zip
./server plugin uninstall my-plugin
```

`install` 调用软件服务的 `Install(path)`，接受目录或 ZIP；服务自动临时打包目录中的清单、后端、前端和可选 `images/`，排除根目录的开发 SDK。`uninstall` 调用 `Uninstall(target)`，可传插件标识或软件记录 ID，直接清理插件数据。`Uninstall` 固定接收 `{"remove_data":true}`，负责删除自身拥有的业务资源，不再提供保留数据选项。卸载先停止已加载实例；未加载时尝试只加载源码、不启动，以执行 `Uninstall`，支持离线清理。源码加载失败会记录错误并继续删除软件记录和代码目录；此时无法执行插件自定义的清理钩子，也不能推断其任意创建的容器、数据库等外部资源，需按插件说明清理。Install(path) 和 Uninstall(target) 均返回任务 ID 与错误，返回成功表示任务已提交，实际结果从系统后台任务获取。HTTP 接口返回 task_id；在线 CLI 提交后退出，离线 CLI 等待完成后退出，避免关闭任务宿主。

`list` 直接调用本地 software service，只读查询已安装软件。`install`、`uninstall` 在面板运行时通过本机进程通信交给面板的 software service 执行，复用插件、容器和后台任务；面板停止时自动直接调用本地服务。通信使用 Unix socket 或 Windows 命名管道，不请求面板 HTTP 接口，也不需要端口和登录 token。离线命令只打开现有数据库，不运行表结构同步，也不会启动网站或定时任务、清空系统任务日志。

面板在网站、计划任务等业务运行时就绪后，根据 `requires` 先启动依赖软件，再启动依赖它的软件。单个插件启动失败会记录错误，依赖它的软件不会带着失败的依赖继续启动；其他无关插件仍可启动。面板退出时先停止依赖方，再停止它们依赖的软件，最后关闭计划任务、网站、系统任务、数据库和缓存，插件的 `Stop` 仍可使用宿主服务。

面板重启使用 `Start` / `Stop`。在线首次安装会启动插件，卸载和替换版本前会停止已加载的旧实例。在线命令行安装同样启动插件；离线安装不启动常驻任务，之后由面板启动恢复。

软件记录的 `status` 表示启用状态：0 禁用、1 启用；独立的 `state` 表示安装进度：1 安装中、2 已安装、3 升级中、4 卸载中。只有已安装且启用的软件可以访问 Actions 或页面资源。禁用调用 `Stop`，重新启用调用 `Start`；升级保留启用状态，禁用的软件不会被升级或面板启动重新拉起。进程异常退出导致升级中断时，会恢复代码并禁用软件，检查配置后再启用。软件信息缓存在进程内，由软件服务写入后刷新，接口和资源请求无需反复查库。

`Start` 的 `ctx` 供异步任务长期使用：方法应创建 goroutine 后及时返回，成功返回后不会受本次启动调用的超时影响。宿主在调用 `Stop` 前取消该上下文；`Stop` 使用新的停止上下文，负责等待 goroutine 退出并清理资源。同一实例重复启动、停止不会重复执行钩子。软件服务关闭时遵守反向依赖顺序；耗时操作必须响应上下文取消，宿主无法强行结束忽略取消的 Go 协程。

```go
func New(app *sdk.App) *sdk.Plugin {
    var done chan struct{}
    return &sdk.Plugin{
        Start: func(ctx context.Context, _ string) (interface{}, error) {
            finished := make(chan struct{})
            done = finished
            go func() {
                defer close(finished)
                // 循环处理任务；阻塞操作需要传入 ctx。
                <-ctx.Done()
            }()
            return nil, nil
        },
        Stop: func(ctx context.Context, _ string) (interface{}, error) {
            if done != nil {
                select {
                case <-done:
                    done = nil
                case <-ctx.Done():
                    return nil, ctx.Err()
                }
            }
            return nil, nil
        },
    }
}
```

`Start` / `Stop` 生命周期与 `Actions["start"]` / `Actions["stop"]` 互相独立，后者仍可用于软件页面里的容器启停。

软件宿主直接注入 `global.App` 已有的四项基础依赖：

| 字段 | 原始对象 | 用法 |
| --- | --- | --- |
| `app.Database` | `global.App.Database` | `Db` 的 GORM 操作，以及 `Transaction`、`BatchExecute` |
| `app.Cache` | `global.App.Cache` | 缓存读取、写入、有效期和锁 |
| `app.Config` | `global.App.Config` | Viper 运行配置，例如 `GetString("server.mode")` |
| `app.Log` | `global.App.Log` | 使用面板日志输出 `Println`、`Printf` |

这些对象共用面板实例，插件不主动关闭连接池或缓存。插件自行命名和维护数据表，不要求固定前缀。面板结构同步只处理安装 SQL 中声明的表及其索引，其他表及其索引会保留。

```go
err := app.Database.Transaction(func(tx *gorm.DB) error {
    return tx.WithContext(ctx).Table("plugin_demo_records").Create(&record).Error
})
app.Cache.Set("plugin:demo:last_greeting", record.Greeting)
app.Log.Printf("[plugin:demo] 保存问候: %d", record.ID)
```

`app.Service` 注入面板现有的全部业务服务，直接复用同一组实例：

| 字段 | 业务 |
| --- | --- |
| `Config` / `Auth` | 配置表读写、登录认证 |
| `Server` / `Script` / `Type` | 终端连接、常用脚本、分类 |
| `Secret` / `Log` | 密钥管理、操作日志 |
| `Task` / `System` | 计划任务、后台任务及系统管理 |
| `File` / `Recycle` | 文件操作、回收站 |
| `Site` | 网站配置、HTTP 服务与证书 |
| `Container` / `Software` | 容器管理、软件管理及插件内部配置读写 |
| `Database` | MySQL、Redis 等数据库客户端业务 |

调用方式与面板内部一致：

```go
size := app.Service.File.FormatSize(1024 * 1024)
categories, err := app.Service.Type.GetEnum("database")
mode := app.Service.Config.Get("demo", "mode")
```

容器同样直接使用业务服务 `app.Service.Container`，包括镜像、终端、日志和资源状态等完整方法。例如创建容器：

```go
info, err := app.Service.Container.Run(sdk.UtilContainerRunOptions{
    ID: app.InstanceID,
    ImageID: "/opt/images/redis.tar",
})
```

参数类型可直接在工程内的 `sdk/` 查看，不需要面板源码。容器状态可与 `"running"`、`"stopped"` 等字符串比较；原生包的枚举常量不随类型自动导出。

`app.Database` 负责面板自身数据库，`app.Service.Database` 负责外部数据库实例及其操作；`app.Config` 是 Viper 运行配置，`app.Service.Config` 读写公共配置表，`app.Service.Software.GetConfig/SetConfig` 读写软件记录中的内部配置；`app.Log` 是程序日志，`app.Service.Log` 是操作日志业务。插件无需通过 HTTP 请求调用这些服务。面板服务自身的启动与关闭由面板管理，插件不要调用其 `Boot`、`Close` 或系统、计划任务的 `Start` 方法；网站和容器的启动、停止等业务操作可正常调用。

通过 `App` 可达的 GORM、Viper、原生 `log.Logger` 等类型由宿主注册，插件使用这些注入对象时无需携带宿主依赖源码。包级函数、变量和常量是否可调用取决于宿主提供的符号，不能仅凭某个类型已注册判断。例如自动收集不会导出 `gorm.ErrRecordNotFound`；示例详情使用 `Limit(1).Find(&record)` 后检查主键是否存在。使用解释器定义的数据结构时显式指定 `Table(...)`。

软件管理页面直接维护 `web/index.tsx`，默认导出普通 React 组件：

```tsx
import React from "react";
import {Button} from "antd";

export default function Page({software, request}: any) {
    return <Button variant="filled" onClick={async () => {
        try {
            const result = await request("hello", {name: software.name});
            console.log(result.message);
        } catch (error) {
            console.error(error);
        }
    }}>发送问候</Button>;
}
```

`software` 只包含 id、name、code、version、icon、status、state、create_time、update_time 和 frontend，不包含 config、重复清单或进程任务信息。`request(method, params)` 调用当前软件的业务接口，成功返回响应中的 `data`，失败抛出错误；`onClose()` 关闭当前管理弹窗。模板采用侧栏工作台，窄屏导航改为横向；页面样式在 `styles.ts` 中通过 `antd-style` 定义，文档组件自行管理其章节样式。

| 示例页面 | 代码位置 | 可直接参考的功能 |
| --- | --- | --- |
| 运行概览 | `web/components/overview.tsx` | 实际记录总数、缓存与配置状态、心跳、已注册接口与源码入口。 |
| 数据管理 | `web/components/notes.tsx`、`server/notes.go` | PageTable 服务端分页和排序、详情按需读取、filled 表单、事务批量删除。 |
| 配置缓存 | `web/components/settings.tsx`、`server/settings.go` | Software.GetConfig/SetConfig 保存设置，临时缓存单独管理有效期和清理。 |
| 后台任务 | `web/components/tasks.tsx`、`server/task.go` | 提交后立即返回、进度轮询、协作式取消、关闭时清理轮询。 |
| 接口调试 | `web/components/playground.tsx` | 预置参数、真实请求、业务错误和耗时展示，写入前确认。 |
| 开发文档 | `web/components/docs.tsx` | 按章节搜索、参数速查、可复制代码、源码位置与排错。 |

后端所有示例复用同一个插件对象；实体集中在 `server/entities.go`，安装、升级、启动、停止和卸载集中在 `server/lifecycle.go`。记录列表只返回 id、标题和时间，正文通过 `notes?action=info` 读取，避免随分页发送大文本。批量删除使用一次请求的 ids 数组，在一个事务内分批处理。

入口可额外命名导出 `modal`，直接配置 Ant Design Modal 的宽度、标题、底部、样式等属性。可以使用属性对象 `export const modal = {width: 800}`，也可以按当前环境返回配置：

```tsx
import type {PluginModal} from "sdk";

export const modal: PluginModal = ({compact, mobile, token}) => ({
    width: mobile ? "calc(100vw - 24px)" : compact ? 880 : 1000,
    styles: {body: {minHeight: mobile ? 360 : 440, background: token.colorBgContainer}},
});
```

同步配置函数接收 `{software, compact, mobile, token, onClose}`，主题或屏幕断点改变时重新计算；函数内不调用 Hooks 或发送请求。`style` 浅合并宿主默认值，插件优先；`styles.body` 默认的 `minWidth: 0` 也可覆盖。`styles`、`classNames` 支持原生对象或函数形式，其他展示属性直接使用 Ant Design Modal API。未提供 `modal` 时保持普通模式宽 1080、紧凑模式宽 960、无底部操作栏。

`open`、`onCancel`、`afterClose`、`afterOpenChange`、`destroyOnHidden`、`destroyOnClose`、`forceRender`、`getContainer`、`children`、`loading` 始终由宿主管理，`closable` 内的 `onClose`、`afterClose` 回调也不接收；自定义按钮通过 `onClose()` 关闭，页面资源在 effect 清理函数中释放。配置写在插件前端入口，不需要新增清单字段。

前端支持 `.ts`、`.tsx`、`.jsx`、`.js`、`.mjs` 多文件，直接写 `import Hello from "./components/hello"`，相对导入会自动尝试扩展名和目录中的 `index`。加载时由浏览器中的 Babel Standalone 按需转换并缓存，不需要开发者手动 build，也不要求服务器安装 Node。转换只处理语法，不执行 TypeScript 类型检查。

省略后缀的导入只发送一次资源请求，由服务端在本地解析文件，并通过 `Content-Location` 返回实际路径；浏览器不会逐个请求候选文件产生 404。目录入口中的相对依赖和模块缓存均以实际路径为准。

可直接导入的共享模块是 `react`、`react/jsx-runtime`、`react/jsx-dev-runtime`、`antd`、`sinking-antd`、`antd-style`、`dayjs`、`sdk`；它们使用面板已有的对象，React Hooks 和主题上下文保持一致。入口无需工厂包装，也不要自己创建 React 根节点。

`sdk` 提供 `ActionDropdown`、`Log`、`ContainerTerminal`、`FormField`、`formatSize`。本地 JSON 文件可直接导入，`import()` 支持按需加载组件；关闭页面会释放本次模块实例，转换缓存按源码内容更新并限制容量。

额外 npm 依赖不会自动下载或安装；可以使用宿主共享库，或将适合浏览器的依赖源码随包放在 `web/`，改为相对导入。当前不支持模块导入 CSS、Less 或 Sass，样式使用 `antd-style`。工程内 `tsconfig.json` 供编辑器识别 TSX，不会安装库类型，也不产生构建文件。

在示例目录中执行 `zip -r demo-1.0.0.zip software.json server web` 即可交付。面板自身发布仍需构建，插件变化不需要重新构建面板。

面板和 CLI 都支持服务器上的源码目录或 ZIP 路径，不提供软件包上传安装接口。先把文件放到服务器，再从「添加软件」填写路径，或执行 `./server plugin install /path/to/demo`、`./server plugin install /path/to/demo.zip`。同一软件标识再次安装会执行升级，安装目录是源码副本，修改开发目录不会自动影响运行中的软件。

使用 `./server plugin init demo` 生成示例工程。生成的 `README.md` 包含从第一次调用到增加完整业务的教程，以及清单每个字段、生命周期与取消、所有示例 Action 的参数和响应、数据库与事务、Config/Cache、全部业务服务、PageTable、可配置弹窗、多文件与 vendor、发布和排错。开发时优先以该工程的 SDK 和可运行代码核对签名。

## 8. 归档辅助 API 与相关工具

除 Manager 外，包还公开 `PackageImageEntry(name string, directory bool) (bool, error)`，供打包和导入逻辑统一识别镜像文件。它只做分类，调用方必须先验证包内相对路径、拒绝符号链接和特殊文件；不要单独把此函数当成路径安全校验。

| 输入 | 结果 |
| --- | --- |
| `images/base.tar`、`images/base.tar.gz`、`images/base.tgz` | `true, nil`，属于镜像预算 |
| `images/config.json` | `false, nil`，仍占源码预算 |
| `server/plugin.go` | `false, nil`，不属于镜像目录 |
| `images/readme.txt` 或普通文件 `images` | 错误，不是允许的镜像目录文件 |
| 目录条目，`directory=true` | `false, nil`；路径是否合法仍由调用方确认 |

公开常量 `MaxPackageFiles=1000`、`MaxSourceSize=64 MiB`、`MaxImageSize=16 GiB` 用于统一归档策略；通用 Manager 的默认总量仍受 `Options.MaxArchiveSize` / `MaxExpandedSize` 限制，不能仅看 MaxImageSize 判断能否导入。

```go
package example

import "server/app/util/plugin"

// IsBundledImage 的 name 应是上层已校验过的归档相对路径。
func IsBundledImage(name string, isDirectory bool) (bool, error) {
    return plugin.PackageImageEntry(name, isDirectory)
}
```

宿主 HTTP 接入见 [context](../context/README.md)，后台并发任务见 [job](../job/README.md)，容器实例与镜像管理见 [container](../container/README.md)，通用文件归档见 [archive](../archive/README.md)。返回 [util 工具索引](../README.md)。
