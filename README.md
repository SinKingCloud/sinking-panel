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
./server start       # 后台启动
./server stop        # 停止服务
./server restart     # 重启服务
./server run         # 前台运行
./server install     # 设置登录信息并安装系统自启动
./server user        # 修改登录账号
./server pwd         # 修改登录密码
./server uninstall   # 卸载服务，保留网站数据并选择是否保留容器
```

`run`、`start`、`restart` 和 `install` 支持通过启动参数指定运行模式、监听地址和端口：

```bash
./server run --mode dev --host 127.0.0.1 --port 5678
./server start --mode release --host 0.0.0.0 --port 8080
./server restart --port 8080
./server install --mode release --host 0.0.0.0 --port 8080
./server --help
```

`install` 会将传入参数保存到系统自启动项，后台启动也会将参数传递给服务进程。Windows 的 `start` 和 `restart` 通过已安装的计划任务运行；传入参数时会更新该任务的启动脚本，不传参数时沿用已保存的参数。Linux 和 macOS 手动再次启动或重启时，需要重新传入希望覆盖的参数。

`uninstall` 会先要求输入 `yes` 确认，再询问是否保留轻量容器。第二次输入 `yes` 保留容器数据，其他输入（包括直接回车）会停止并删除安装目录下 `data/containers` 中的容器及其数据。中途结束输入会取消卸载。程序文件、运行日志、`runtime` 和 `data/server` 会被删除，网站文件和外部挂载源数据保留。

容器停止与挂载清理由服务退出流程负责。卸载时若仍有运行状态或挂载残留，会报错并保留数据，待清理完成后重试卸载。

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

## 本地开发

### 后端

```bash
cd server
go mod download
CGO_ENABLED=1 go build -tags seccomp -o server .
./server run
```

后端默认运行在 `http://127.0.0.1:5678`。项目要求 Go 1.25.1；在 Linux 上直接编译完整后端时还需要可用的 C 编译环境和 libseccomp 开发文件，或直接使用下方 Docker 构建方式。使用固定位置的可执行文件运行，避免 `go run` 将容器数据放入临时编译目录。

服务启动时会初始化 `global.App.Container`，正常退出时停止容器。Linux 运行时要求 root、支持 `openat2` 的内核，以及为 root 配置至少 65536 个连续 ID 的 `/etc/subuid` 和 `/etc/subgid`；容器目录及祖先目录需要符合运行库的 root 所有权与权限要求。安装和修改账号密码命令不会初始化容器管理器。

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

Sinking Panel 当前是单机服务器管理面板，不包含集群、多节点集中管理、告警通知、软件商店或网站管理页面。`server/app/util/container` 和 `server/app/util/server` 中的代码是后续业务能力储备，不代表相关功能已经可以从面板使用。

## 反馈

- 项目地址：[github.com/SinKingCloud/sinking-panel](https://github.com/SinKingCloud/sinking-panel)
- 问题反馈：[GitHub Issues](https://github.com/SinKingCloud/sinking-panel/issues)
