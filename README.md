# Sinking Panel

Sinking Panel 是一个面向个人服务器运维场景的轻量级 Web 管理面板。项目采用 Go + Umi/React 开发，前端静态资源、SQLite 数据库和后端服务可以一起打包进单个 Linux 可执行文件，适合直接部署到服务器使用。

当前面板聚焦单机管理，已经接入数据概览、文件管理、计划任务、终端管理、操作日志和系统设置。仓库中的容器运行时与 HTTP 服务管理代码仍属于底层能力，尚未接入面板业务和前端页面。

## 功能概览

### 数据概览

- 展示主机名、操作系统、内核、架构和运行时间
- 查看 CPU、内存、系统负载及磁盘空间使用情况
- 查看磁盘分区、挂载点和容量信息
- 实时展示网络吞吐、磁盘读写速度和 IOPS 曲线

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
| 后端 | Go 1.25、SQLite、WebSocket、Cron、gopsutil |
| 前端 | Umi 4、React、Ant Design、sinking-antd |
| 终端 | xterm.js、SSH |
| 编辑器 | Ace Editor |
| 数据存储 | SQLite + 本地配置文件 |
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
| `data/site/` | 自动创建的网站文件，卸载时保留 |
| `data/containers/` | 轻量容器镜像、实例和运行数据，卸载时按选择保留或删除 |
| `runtime/recycle/` | 文件回收站 |
| `runtime/cron/` | 计划任务日志 |
| `runtime/task/` | 系统异步任务日志 |
| `server.pid` | 后台进程 PID |
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

## 数据库软件管理

Linux 的「数据管理」页面按 Redis、MySQL、PostgreSQL、MongoDB 分页签管理。各软件在 `server/app/service/container/` 下独立封装 `Manager`，通过 `service.Container.Redis()`、`Mysql()`、`Pgsql()`、`Mongo()` 调用。接口分别使用 `/redis/*`、`/mysql/*`、`/pgsql/*`、`/mongo/*`，仅在 Linux 注册并要求登录。

MySQL、PostgreSQL、MongoDB 支持本机官方镜像解析安装、启停重启、终端、日志、负载、可视化调优、配置原文、数据库和业务账户管理。根据镜像版本适配初始化、认证、管理命令和配置，不再限制为 MySQL 8.0+、PostgreSQL 14+、MongoDB 6+。PostgreSQL 自动处理旧版 MD5 与新版 SCRAM 认证，以及 18 起的数据卷位置；MongoDB 自动选择 `mongo` / `mongosh`，按版本和存储引擎提供调优项；Redis 按版本处理配置指令、在线修改和 AOF。

已实测 MySQL 5.5、5.6、5.7、9.7，PostgreSQL 9.3、9.4、9.6、10、18，MongoDB 3.6、4.4、8.3，Redis 3.2、5、8 的代表版本，并使用官方源码构建的 Redis 2.6 验证旧版配置与管理命令。适配针对官方 Linux 镜像，镜像架构必须与服务器一致；未逐一验证全部历史补丁版本，旧镜像自身的 CPU、内核与入口程序限制仍需满足。

安装默认仅监听 `127.0.0.1`，管理员密码通过只读文件传给官方镜像。数据保存在容器管理器的 `volumes/<实例 ID>/` 下。卸载默认保留数据、配置和管理凭据，再次安装使用原镜像恢复；显式选择删除数据才会一并清除。配置保存不会自动重启，页面会显示待重启状态；PostgreSQL 支持重载可在线应用的配置。

`server/data/images/` 已准备 Linux amd64 归档：MySQL 9.7.2 LTS、PostgreSQL 18.6、MongoDB 8.3.11、Redis 8.10.1。镜像版本、来源、平台和摘要记录在 `manifest.json` 与 `SHA256SUMS`。将所需归档复制到目标服务器，在安装表单填写服务器上的绝对路径即可；面板运行数据库不依赖 Docker。

## 本地开发

### 后端

```bash
cd server
go mod download
CGO_ENABLED=1 go build -tags seccomp -o server .
./server run
```

后端默认运行在 `http://127.0.0.1:5678`。项目要求 Go 1.25.1；在 Linux 上直接编译完整后端时还需要可用的 C 编译环境和 libseccomp 开发文件，或直接使用下方 Docker 构建方式。使用固定位置的可执行文件运行，避免 `go run` 将容器数据放入临时编译目录。

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
