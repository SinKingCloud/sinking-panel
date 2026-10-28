# util 工具使用索引

本目录包含面板后端的通用工具。每个有 Go 源码的工具包和子包都有自己的 README，说明真实 API、配置、示例、错误及适用边界。本文用于选型和导航，具体签名以相应源码为准。

## 使用方式

这些包属于仓库的 `server` Go module，通常在 `server/` 目录构建，导入路径从 `server/app/util/...` 开始。它们不是可直接 `go get server/...` 的独立公共模块。使用仓库 `go.mod` 指定的 Go 版本和依赖，不要把单个工具文件复制到新项目后忽略其配套类型、平台文件和初始化要求。

文档中的完整函数可放入业务包；没有 `package` / `import` 的代码段是接续示例，需沿用文中说明的上下文。示例地址、桶名、路径和凭据参数都需要替换成实际配置。涉及进程、容器、系统时区和网络监听的示例会产生真实系统操作，阅读配置及关闭语义后再运行。

## 文件、归档和对象存储

| 工具 | 适用场景 | 使用入口 |
| --- | --- | --- |
| `archive` | 支持取消和进度的归档/解压，新业务的文件归档入口 | [详细说明](archive/README.md) |
| `compress` | 旧压缩接口，维护已有调用时使用；不要假定与 archive 行为完全一致 | [详细说明](compress/README.md) |
| `file` | 文件、目录、权限及文件锁等本地操作；root 不是访问隔离边界 | [详细说明](file/README.md) |
| `oss` | 五家云厂商的统一对象存储管理、流式备份上传及分片恢复 | [详细说明](oss/README.md) |
| `log` | 按字节游标读取及清空日志文件，支持轮转文件；不是应用日志记录器 | [详细说明](log/README.md) |

OSS 厂商和公共类型：[阿里](oss/aliyun/README.md)、[腾讯](oss/tencent/README.md)、[华为](oss/huawei/README.md)、[火山](oss/volcengine/README.md)、[百度](oss/baidu/README.md)、[types](oss/types/README.md)。业务统一使用根包 `oss`，厂商子包面向适配器实现和维护。

## 数据访问、参数与基础类型

| 工具 | 适用场景 | 使用入口 |
| --- | --- | --- |
| `database` | GORM 数据库封装及 SQLite 初始化 | [详细说明](database/README.md) |
| `repository` | 基于数据库封装构建仓储和分页查询 | [详细说明](repository/README.md) |
| `page` | 分页参数、查询选项和返回数据结构 | [详细说明](page/README.md) |
| `cache` | 进程内带过期时间的缓存 | [详细说明](cache/README.md) |
| `str` | 字符串、日期、字节、JSON、编号及历史编解码辅助功能 | [详细说明](str/README.md) |
| `validator` | struct 校验、中文错误和标签规则 | [详细说明](validator/README.md) |

数据库连接、缓存和分页上下文的所有权不同，不能因为名称都叫“工具”就按请求反复初始化；关闭、过期及事务语义请看各包说明。

## HTTP、认证与网络

| 工具 | 适用场景 | 使用入口 |
| --- | --- | --- |
| `context` | 面板 HTTP 控制器上下文、参数绑定、响应和用户信息 | [详细说明](context/README.md) |
| `captcha` | 图片及滑块验证码生成和验证 | [详细说明](captcha/README.md) |
| `jwt` | 面板现有令牌生成、解析和校验约定 | [详细说明](jwt/README.md) |
| `ip` | IP 地址归属查询及随包 CZDB 数据库 | [详细说明](ip/README.md) |
| `webssh` | SSH 连接、命令与交互会话 | [详细说明](webssh/README.md) |
| `caddy` | 进程内 Web 服务、站点配置、证书、缓存、流量统计及 WAF 集成 | [详细说明](caddy/README.md) |

`util/context` 与标准库 `context` 是不同包，在同一文件里同时使用时建议给前者起别名，例如 `webContext`。

Caddy 子模块：[缓存](caddy/cache/README.md)、[百度 DNS](caddy/dns/baidu/README.md)、[流量统计](caddy/traffic/README.md)、[WAF writer](caddy/waf/README.md)。它们服务于父包的配置/注册链，不能默认当成独立 HTTP 服务使用。

IP 内部实现：[CZDB](ip/czdb/README.md)、[数据模型](ip/czdb/entity/README.md)、[二进制转换](ip/czdb/byteUtil/README.md)、[常量](ip/czdb/constant/README.md)、[头部解码](ip/czdb/hyperHeaderDecoder/README.md)、[AES 解码](ip/czdb/aesECB/README.md)。普通业务优先使用 `ip` 顶层入口。

## 进程、运行时与系统集成

| 工具 | 适用场景 | 使用入口 |
| --- | --- | --- |
| `cmd` | 一次性宿主命令，控制执行时间、输出和取消 | [详细说明](cmd/README.md) |
| `process` | 长期宿主进程、保活、重启及状态管理 | [详细说明](process/README.md) |
| `container` | Linux 容器镜像、实例、终端、日志及资源采样 | [详细说明](container/README.md) |
| `daemon` | 应用守护运行、启停控制及开机启动集成 | [详细说明](daemon/README.md) |
| `ipc` | 本机进程间控制通信，平台间使用不同传输机制 | [详细说明](ipc/README.md) |
| `job` | 有界并发生产/消费任务处理 | [详细说明](job/README.md) |
| `recordstore` | 任意 JSON / 原始字节的集合存储、键读写、时间查询及 TTL；附数值序列和旧事件格式扩展 | [详细说明](recordstore/README.md) |
| `platform` | 操作系统、运行环境及系统时区管理 | [详细说明](platform/README.md) |
| `plugin` | Go 插件源码解释、清单、生命周期和宿主能力注入 | [详细说明](plugin/README.md) |

时区底层实现及平台差异另见 [platform/timezone](platform/timezone/README.md)。`job` 提供并发处理能力，不是定时任务调度器；`process` 管理进程，`container` 管理实例隔离，两者的停止和恢复语义不同。

## 常见组合

| 需求 | 推荐阅读顺序 |
| --- | --- |
| 数据库或文件备份上传云存储 | `archive` / `cmd` → `oss`，通过 context 和 io.Reader 连接备份与上传 |
| 新增 CRUD 列表接口 | `context` → `validator` → `page` → `repository` / `database` |
| 轻量本地业务记录、配置与自定义编码数据 | `recordstore` 的 `Open` / `Collection`；业务自行定义数据结构与筛选统计 |
| 宿主后台常驻任务 | `process`，结合 `log` 提供增量日志；一次性命令使用 `cmd` |
| 容器软件管理 | `container` → `log` / `recordstore`；插件侧接入再阅读 `plugin` |
| 网站与自动证书 | `caddy`，并按相应 DNS 提供商配置凭据 |
| 系统服务启停控制 | `daemon` → `ipc`，根据平台确认控制通道与身份要求 |
| IP 归属展示 | `ip`；只有维护数据格式时才阅读 czdb 子包 |

## 开发与排查约定

1. **先确认资源所有权。** 下载流、文件、SSH 会话、数据库、管理器、后台 worker 各有不同关闭方式。特别注意 `container.Close()` 默认保留容器运行。
2. **沿用当前包的错误约定。** 有些旧工具以布尔值、空串或固定返回值表达失败，不都遵循 `(value, error)`；文档明确列出这些行为，不要只根据函数名推测。
3. **区分 context 取消和强制停止。** 接口接收 context 不意味着能中断任意用户回调或阻塞 reader，业务需协作取消。
4. **核对路径和命令参数。** 文件工具的路径基准不等于沙箱，shell 字符串也不等于参数数组。HTTP 参数不能未经业务校验直接传给这些底层工具。
5. **确认平台构建条件。** Linux 容器真实运行需 cgo/libseccomp；Windows/Unix 的进程控制、IPC、时区实现也不同。某平台下编译成功不代表另一平台运行环境已验证。
6. **只构建需要修改的包。** 在 `server/` 下可执行 `go build ./app/util/工具名/...` 和 `go vet ./app/util/工具名/...`；完整服务构建条件见[项目 README](../../../README.md)。

维护工具 API 时请同步更新对应 README 的签名、默认值、示例和关联链接，避免调用方继续照过期示例集成。
