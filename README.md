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
./server run
```

默认监听 `0.0.0.0:5678`，启动后访问：

```text
http://服务器IP:5678
```

首次登录时填写的账号和密码会被初始化为面板登录凭据。首次启动后请及时完成登录初始化，不要在未设置访问控制的情况下长期暴露到公网。

## 服务命令

```bash
./server start       # 后台启动
./server stop        # 停止服务
./server restart     # 重启服务
./server run         # 前台运行
./server install     # 安装系统自启动
./server user        # 修改登录账号
./server pwd         # 修改登录密码
./server uninstall   # 卸载自启动并删除软件数据
```

`uninstall` 会要求输入 `yes` 确认，并删除当前安装相关的数据文件，执行前请先备份。

## 配置与数据

程序使用当前工作目录作为数据根目录：

| 路径 | 用途 |
| --- | --- |
| `config/application.yml` | 服务监听配置 |
| `config/server.db` | SQLite 业务数据库 |
| `temp/recycle/` | 文件回收站 |
| `temp/cron/` | 计划任务日志 |
| `temp/task/` | 系统异步任务日志 |
| `server.pid` | 后台进程 PID |
| `server.log` | 后台运行日志 |

最小配置示例：

```yaml
server:
  mode: prod
  host: 0.0.0.0
  port: 5678
```

备份面板时至少应保留整个 `config` 目录。需要保留回收站和计划任务日志时，同时备份 `temp` 目录；系统异步任务日志属于运行期日志，服务初始化时可能被清理。

## 本地开发

### 后端

```bash
cd server
go mod download
go run . run
```

后端默认运行在 `http://127.0.0.1:5678`。项目要求 Go 1.25.1；在 Linux 上直接编译完整后端时还需要可用的 C 编译环境和 libseccomp 开发文件，或直接使用下方 Docker 构建方式。

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
│   ├── config/            # 开发配置
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
