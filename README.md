# Sinking Panel

Sinking Panel 是一个面向个人服务器运维场景的轻量级 Web 管理面板。项目采用 Go + Umi/React 开发，前端静态资源、SQLite 数据库和后端服务可以一起打包进单个 Linux 可执行文件，适合直接部署到服务器使用。

当前面板聚焦单机管理，已经接入数据概览、文件管理、计划任务、终端管理、操作日志和系统设置。仓库中的容器运行时与 HTTP 服务管理代码仍属于底层能力，尚未接入面板业务和前端页面。

## 功能概览

### 数据概览

- 展示主机名、操作系统、内核、架构和运行时间
- 查看 CPU、内存、系统负载及磁盘空间使用情况
- 查看磁盘分区、挂载点和容量信息
- 实时展示网络吞吐、磁盘读写速度和 IOPS 曲线

### 监控管理

- `/monitor` 展示 CPU 均值与峰值、内存、磁盘容量和读写速度、网络收发速度，以及 1 / 5 / 15 分钟负载曲线
- 支持快捷时间范围、自定义起止时间、开启或关闭历史监控，自定义保留 1–365 天（默认 30 天）
- 复用每秒实时采集，在内存中汇总后每分钟写入一条记录；关闭历史监控不影响首页实时数据
- 开关和保留天数存入 `cloud_configs` 表的 `monitor.enabled`、`monitor.retention_days`；保存成功后即时生效
- 历史记录独立保存在 `data/server/monitor/`，不写入业务 SQLite；按 UTC 日期分文件，按时间偏移读取，长时间范围查询自动聚合
- 每条记录 136 字节，一天最多约 191 KiB，30 天约 5.6 MiB，365 天约 68.2 MiB（不含少量文件系统开销）；后台定期删除过期日文件，边界日最多额外保留不足一天的文件空间
- 时间显示使用浏览器当地时间，接口使用毫秒时间戳；分钟均值会平滑瞬时变化，CPU 额外保留采样峰值。正常退出会保存当前分钟，异常退出可能丢失尚未写入的当前分钟及队列中的采样

网络汇总优先统计物理接口，排除环回和容器虚拟接口以减少重复计算；磁盘 I/O 汇总底层设备，容量按文件系统去重。Linux/macOS 使用系统原生负载；Windows 的 1 / 5 / 15 分钟曲线为 CPU 忙碌核心数的指数移动平均估计值。

`/system/monitor` 沿用面板登录验证，支持 `action=settings` 读取设置、`action=set` 配合 `enabled` 和 `retention_days` 修改设置，以及 `action=history` 配合 `start` / `end` 毫秒时间戳查询。监控查询以 1440 个时间点为目标，按常用间隔向上选择粒度，含区间边界最多返回 1441 点；例如 24 小时按 1 分钟、7 天按 10 分钟、30 天按 30 分钟、365 天按 12 小时聚合。均值按真实样本数加权，CPU 峰值取桶内最大值，原始分钟记录仍保留。接口从所选范围内的首条真实采样开始展示，不在采样前补零；范围内完全没有记录时返回空列表。首条采样之后的中间及末尾空桶按同一间隔补齐为 0，并标记 `samples=0`。补零只用于接口展示，不写入历史文件，也不参与真实数据的均值和峰值聚合。

采集和存储通过 [`util/recordstore`](server/app/util/recordstore/README.md) 分离。该工具的通用入口 `Open` / `Collection` 使用独立 bbolt 数据库保存任意 JSON 业务值和原始字节，提供集合隔离、键读写、原子批量写入、时间游标查询及可选 TTL，默认永久保存。系统监控使用其数值序列扩展 `Recorder`：service 负责提交指标和持久化配置，由记录器异步汇总、启停和清理。日文件使用 `.series`，字段校验信息保存在日文件头；配置仅从业务配置表读取。

### 文件管理

- 浏览磁盘和目录，支持路径导航、搜索、排序、分页与多选
- 新建文件或目录，上传、下载及远程下载文件
- 复制、移动、重命名、修改权限、查看属性和统计目录大小
- 压缩及解压 ZIP、TAR、GZIP、TAR.GZ 等常用格式
- 大文件分片上传，支持 MD5 校验、暂停、继续和断点续传
- 图片、音频和视频签名预览，支持前后切换及全屏查看
- 内置多标签 Ace 文件编辑器，提供目录树、主题、字号和全屏设置
- 在当前目录打开持久化终端会话
- 回收站支持恢复、彻底删除和清空
- 文件复制、移动、压缩、解压及远程下载统一进入异步任务队列

### 计划任务

- 创建系统脚本或 HTTP 请求任务
- 支持按秒、分、小时、天、周、月及自定义 Cron 表达式执行
- 管理任务分类，按条件搜索和筛选任务
- 支持立即执行、暂停、恢复、编辑和删除
- 查看实时执行日志和历史日志

### 终端管理

- 使用浏览器连接本机终端和多台远程 SSH 服务器
- 支持密码和 PEM 私钥认证
- 支持多终端会话、会话切换和连接信息管理
- 管理常用脚本及脚本分类，并快速填入当前终端
- 终端输入输出采用 WebSocket 实时传输

### 操作日志

- 记录操作 IP、IP 归属地、操作类型、标题、内容和时间
- 支持查询、筛选和排序
- 可按保留天数批量清理历史日志

### 系统设置

- 修改面板名称和浏览器标题
- 修改登录账号和密码
- 配置上下或左右布局、亮色或暗色主题、紧凑模式和界面水印
- 自定义主题色和组件圆角
- 全局异步任务中心展示队列状态、执行进度、消息及详细日志
- 支持响应式布局，适配桌面端和移动端

## 技术栈

| 模块 | 技术 |
| --- | --- |
| 后端 | Go 1.27、SQLite、WebSocket、Cron、gopsutil |
| 前端 | Umi 4、React、Ant Design、sinking-antd |
| 终端 | xterm.js、SSH |
| 编辑器 | Ace Editor |
| 数据存储 | SQLite、bbolt 通用记录存储、本地配置与数值历史文件 |
| 发布方式 | 前端资源嵌入 Go，可构建为静态 Linux 单二进制 |

## 快速开始

将构建好的 `server` 放入独立目录，并始终在该目录中运行。配置、数据库和临时数据均按当前工作目录保存。

Linux 请使用 root 用户或 `sudo` 执行服务命令；Windows 请在以管理员身份运行的终端中执行。程序会在加载配置前检查权限，权限不足时直接退出。文件的 `777` 权限不代表进程拥有 root 权限，查看帮助不需要管理员权限。

```bash
chmod +x server
./server install
./server run
```

默认监听 `0.0.0.0:5678`，启动后访问：

```text
http://服务器IP:5678
```

执行 `./server install` 时会依次询问登录账号、登录密码和确认密码，全部校验通过后再安装系统自启动。登录页不会初始化账号密码。

## 服务命令

```bash
./server start       # 后台启动面板及容器，无需先安装系统服务
./server stop        # 停止面板，保留容器运行
./server restart     # 重启面板，运行中的容器不重启
./server run         # 直接前台运行，无需安装系统服务
./server install     # 设置账号密码、安装系统服务并立即启动
./server user        # 修改登录账号
./server pwd         # 修改登录密码
./server uninstall   # 卸载软件，选择清理或保留容器、镜像及数据
```

`run`、`start`、`restart` 和 `install` 支持通过启动参数指定运行模式、监听地址和端口：

```bash
./server run --mode dev --host 127.0.0.1 --port 5678
./server start --mode release --host 0.0.0.0 --port 8080
./server restart --port 8080
./server restart --port 8080 --containers  # 停止容器后随面板重新启动
./server stop --containers                # 同时停止容器，面板已退出时也可使用
./server install --mode release --host 0.0.0.0 --port 8080
./server --help
```

`run` 和 `start` 无需安装系统服务。`install` 会交互设置账号密码，将启动参数保存到系统自启动项，随后直接执行启动流程。已安装 Linux 系统服务时，`start`、`restart` 不传新启动配置会沿用安装时的配置；显式传入时更新配置。普通后台运行和 Windows 命令行启动时，需要重新传入希望覆盖的参数；从面板页面重启则保留当前启动参数和工作目录。

面板通过本机 Socket（Linux/macOS）或命名管道（Windows）提供运行状态和停止通知，不再创建根目录 `server.pid`。通信端点覆盖启动初始化与退出清理全过程，防止重复运行；`start` 在未运行时创建进程，`stop` 等待清理完成，`restart` 先停止再启动。

面板启动时接管运行中的轻量容器，并启动已停止的容器，包括未开启自动保活的实例。`stop`、`restart` 默认保留运行中的容器；传入 `--containers`（或 `--containers=true`）才会同时停止容器，重启随后通过正常启动流程重新运行容器。该参数仅影响本次操作，不改变实例的自动保活配置。面板停止期间，容器继续运行，但保活和日志轮转暂停。

首页支持重启面板和服务器。面板重启兼容普通后台进程、Linux systemd 和 macOS launchd；服务器重启根据操作系统调用正常的重启命令，运行账户需要具备对应权限。

`/system/restart` 接受 `{"action":"panel","containers":true}`，`containers` 默认为 `false`。重启整个服务器仍会中断所有容器，无法通过此参数保留进程。

`util/daemon` 的 `Restart()` 可在运行中的服务回调内请求自重启，不要求程序提供 `run` 或 `restart` 子命令。默认保留当前可执行文件、工作目录和启动参数，也可通过 `AutoStartOptions` 指定；`Arguments: nil` 沿用当前参数，空切片表示不传参数，普通重启无需配置自启动名称。方法返回成功表示已提交重启，控制进程会等待旧进程退出后再启动新进程，停止超时则记录日志并取消启动。外部控制进程重启服务仍使用 `Reload()`。

`uninstall` 先要求输入 `yes` 确认，再询问是否保留容器、镜像及数据。第二次输入 `yes` 会完整保留安装目录下的 `data/containers`，重新将软件安装到原目录后，启动即可恢复容器；其他输入（包括直接回车）会清理整个容器目录，包含镜像、实例、日志、配置和持久卷。中途结束输入会取消卸载。程序文件、运行日志、`runtime` 和 `data/server` 会被删除，网站文件和外部挂载源数据保留。

卸载始终停止容器；即使面板此前已经停止，也会接管并清理遗留容器。若仍有运行状态或挂载残留，会报错并保留数据，待清理完成后重试卸载。

## 配置与数据

配置和临时数据使用当前工作目录；轻量容器使用面板可执行文件所在目录下的 `data/containers`，初始化与卸载共用 `constant.ContainerPath`：

| 路径 | 用途 |
| --- | --- |
| `data/server/application.yml` | 可选的服务监听配置，不自动创建 |
| `data/server/server.db` | SQLite 业务数据库 |
| `data/server/monitor/` | 按天保存的系统监控分钟历史文件，设置存入业务数据库配置表 |
| `data/site/` | 自动创建的网站文件，卸载时保留 |
| `data/containers/` | 轻量容器镜像、实例和运行数据，卸载时按选择保留或删除 |
| `runtime/recycle/` | 文件回收站 |
| `runtime/cron/` | 计划任务日志 |
| `runtime/task/` | 系统异步任务日志 |
| `runtime/daemon` | 面板本机通信与运行互斥目录，不保存 PID |
| `server.log` | 后台运行日志 |

程序不会自动创建 `config` 目录或 `application.yml`。数据库首次打开时会创建 `data/server` 目录和数据库文件。无需配置文件即可启动，默认值为 `mode=release`、`host=0.0.0.0`、`port=5678`。

配置优先级：显式传入的命令行参数 > `data/server/application.yml` > 内置默认值，不读取环境变量覆盖配置。

需要使用配置文件时，手动创建 `data/server/application.yml`，内容示例：

```yaml
server:
  mode: release
  host: 0.0.0.0
  port: 5678
```

备份面板时至少应保留整个 `data/server` 目录。需要保留回收站和计划任务日志时，同时备份 `runtime` 目录；系统异步任务日志属于运行期日志，服务初始化时可能被清理。

## 软件插件

「软件管理」页面选择服务器上的源码目录或 ZIP 路径完成插件安装，业务配置由插件在使用时提供。后端由 `util/plugin` 使用 Yaegi 加载，通过 `Plugin` 生命周期处理安装、卸载、升级和面板启停；插件通过 `Actions: map[string]func(*sdk.Context)` 注册 `/software/<软件标识>/<方法名>` 接口，沿用面板登录验证。接口像面板控制器一样使用 `c.ValidatorAll` 校验参数、`c.SuccessWithData` 或 `c.Error` 返回响应，耗时操作使用 `c.Request.Context()`；前端直接传 JSON 或表单参数。`/software/<软件标识>/page` 直接读取清单中的前端入口，多文件资源使用 `/software/<软件标识>/page/<资源路径>`，路径相对插件的 `web/` 目录，例如 `/software/demo/page/components.js` 对应 `web/components.js`；标识和路径各段分别 URL 编码。`page` 是保留资源路由，不交给 `Actions` 处理。插件可声明适用系统和依赖软件，并通过 `Start` / `Stop` 在面板启停时启动、回收异步任务。软件安装后即可使用，不再单独设置启用状态；`app.InstanceID` 直接使用清单 ID，插件内部配置通过 `app.Service.Software.GetConfig(app.InstanceID)` 和 `SetConfig(app.InstanceID, data)` 读写软件记录的 `config` 字符串；详情接口不返回该字段，目录与依赖不存入配置。

插件模板保存在 `server/public/plugin` 并嵌入程序。使用面板二进制即可初始化插件工程和匹配的独立 SDK，不需要面板源码或额外生成命令：

```sh
./server plugin init my-plugin ./my-plugin --name "我的插件"
./server plugin install ./my-plugin
./server plugin list
./server plugin uninstall my-plugin
```

初始化命令可独立运行；其余命令从面板工作目录执行。列表直接读取本地软件服务；安装和卸载在面板运行时通过本机进程通信交给面板执行，停止时直接调用本地服务，无需手动切换。通信使用 Unix socket 或 Windows 命名管道，不依赖面板 HTTP 接口、端口或登录 token。在线安装会启动插件，卸载会先停止插件后台任务。安装、卸载分别统一调用 software service 的 `Install`、`Uninstall`，打包、生命周期和卸载任务等待都由 service 处理。安装支持目录或 ZIP，卸载会清理插件业务数据、代码目录和软件记录，不再提供保留数据选项。插件后端和前端均无需手动构建。

`plugin init` 生成的工程包含 Go 后端、React TSX 多文件页面和独立 SDK，不依赖容器。前端默认导出普通 React 组件，使用标准 `import`、Hooks 和 `antd-style`；浏览器通过 Babel Standalone 按需转换并缓存，React、Ant Design、sinking-antd 和主题能力复用面板已有实例。共享模块之外的 npm 包不会自动安装，当前不支持 CSS、Less、Sass 模块导入。模板位于 [server/public/plugin](server/public/plugin)，运行机制与完整命令说明见 [插件说明](server/app/util/plugin/README.md)。原数据库软件源码及分发包已归档到 `backups/`，待插件系统完善后再继续维护。

## 本地开发

### 后端

```bash
cd server
go mod download
CGO_ENABLED=1 go build -tags seccomp -o server .
./server run
```

后端默认运行在 `http://127.0.0.1:5678`。项目要求 Go 1.27.0 或更高版本；在 Linux 上直接编译完整后端时还需要可用的 C 编译环境和 libseccomp 开发文件，或直接使用下方 Docker 构建方式。使用固定位置的可执行文件运行，避免 `go run` 将容器数据放入临时编译目录。

服务启动时会初始化 `global.App.Container`，正常退出时停止容器。Linux 运行时要求 root 和支持 `openat2` 的内核。首次启动会复用 root 已有的 `/etc/subuid`、`/etc/subgid` 映射，缺少时自动分配至少 65536 个连续的空闲 ID 并写入，不需要手动配置或安装额外命令；后续启动复用 `data/containers/security.json` 中保存的映射。容器目录及祖先目录需要符合运行库的 root 所有权与权限要求。安装和修改账号密码命令不会初始化容器管理器。

### 前端

```bash
cd web
yarn install
yarn dev
```

前端开发服务器默认运行在 `http://127.0.0.1:9002`。联调其他后端地址时，修改 `web/config/defaultSettings.ts` 中的 `gateway`。

## 构建发布版

### 1. 构建前端

```bash
cd web
yarn install
yarn build
cd ..

mkdir -p server/public/dist
cp -R web/dist/. server/public/dist/
```

后端通过 `go:embed` 嵌入 `server/public/dist`。因此每次发布前端后，需要先更新该目录，再构建后端。

### 2. 构建 Linux AMD64 后端

在项目根目录执行：

```bash
docker build --platform linux/amd64 -f server/Dockerfile \
  --output type=local,dest=server/build/linux-amd64 server
```

生成文件位于：

```text
server/build/linux-amd64/server
```

构建 ARM64 时修改平台和输出目录：

```bash
docker build --platform linux/arm64 -f server/Dockerfile \
  --output type=local,dest=server/build/linux-arm64 server
```

Dockerfile 会将 cgo 和 libseccomp 静态链接进 Linux 可执行文件，并在构建阶段检查动态依赖。目标服务器不需要安装 Docker、runc、libseccomp 或 glibc 动态运行库。

已有 `server/public/dist` 时，可运行 `./server/deploy.sh` 一次完成构建、打包和 SSH 部署，部署后不启动程序；仅构建使用 `./server/deploy.sh --build-only`。目标服务器、端口、目录和 SSH 私钥等配置见 [部署说明](server/DEPLOY.md)。

## 项目结构

```text
sinking-panel/
├── server/
│   ├── app/
│   │   ├── command/       # 服务命令
│   │   ├── http/          # HTTP 路由、控制器和中间件
│   │   ├── model/         # 数据模型
│   │   ├── service/       # 业务服务
│   │   └── util/          # 通用能力与底层工具
│   ├── bootstrap/         # 配置、数据库和缓存初始化
│   ├── data/server/       # SQLite 数据库与可选启动配置
│   ├── public/            # SQL 与嵌入式前端资源
│   ├── Dockerfile         # 静态 Linux 构建
│   └── main.go
└── web/
    ├── config/            # Umi 路由和构建配置
    ├── public/            # 前端公共资源
    └── src/
        ├── layouts/       # 全局布局
        ├── pages/         # 页面与业务组件
        ├── service/       # API 调用
        └── utils/         # 前端工具
```

## 当前范围

Sinking Panel 当前是单机服务器管理面板，不包含集群、多节点集中管理、告警通知、软件商店或网站管理页面。`server/app/util/container` 和 `server/app/util/caddy` 中的代码是后续业务能力储备，不代表相关功能已经可以从面板使用。

## 反馈

- 项目地址：[github.com/SinKingCloud/sinking-panel](https://github.com/SinKingCloud/sinking-panel)
- 问题反馈：[GitHub Issues](https://github.com/SinKingCloud/sinking-panel/issues)
