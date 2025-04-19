# Sinking Panel - 无依赖服务器管理面板

Sinking Panel是一个轻量级、无外部依赖的服务器管理面板，基于Go语言开发，提供简洁高效的服务器管理解决方案。通过单一二进制文件部署，无需复杂的环境配置，即可实现对多台服务器的集中管理和监控。

本项目采用前后端分离架构设计：
- **后端(server)**：基于Go语言开发，提供RESTful API和WebSocket接口
- **前端(web)**：计划使用现代前端框架开发，目前正在规划中

## 特性

- **无外部依赖**：单一二进制文件部署，内置Web服务器和数据库支持
- **跨平台兼容**：支持Linux、Windows等多种操作系统
- **轻量高效**：占用资源少，运行高效稳定
- **安全可靠**：支持多种认证方式，保障系统安全
- **前后端分离**：清晰的API设计，便于二次开发和集成

## 主要功能

### 服务器管理
- 集中管理多台远程服务器
- 支持密码认证和密钥认证两种连接方式
- 远程终端直接操作服务器

### 系统监控
- 实时监控CPU、内存、磁盘使用情况
- 系统负载和网络流量统计
- 自定义告警阈值和通知规则

### 文件管理
- 远程文件浏览、上传、下载
- 文件编辑、权限修改
- 回收站功能，防止误删文件

### 定时任务
- 可视化创建和管理Cron定时任务
- 任务执行状态实时监控
- 详细的执行日志记录

### 配置管理
- 图形化配置管理界面
- 配置版本控制和在线编辑
- 配置热加载，无需重启服务

### 系统服务
- 守护进程模式运行，支持开机自启
- 完善的服务生命周期管理
- 系统日志查看和分析

## 开发状态

- **后端(server)**：已基本实现核心功能
- **前端(web)**：尚未开发，正在规划中
- **API文档**：开发中

## 快速开始

### 安装方法

1. **下载最新版本**

```bash
# Linux 64位
wget https://github.com/SinKingCloud/sinking-panel/releases/latest/download/sinking-panel-linux-amd64 -O sinking-panel
chmod +x sinking-panel

# Windows 64位
# 下载 sinking-panel-windows-amd64.exe
```

2. **启动服务**

```bash
# 启动服务
./sinking-panel start

# 停止服务
./sinking-panel stop

# 重启服务
./sinking-panel restart

# 前台运行（调试模式）
./sinking-panel run
```

### 配置说明

首次运行会在`config`目录下生成默认配置文件，修改`config.yaml`可自定义以下配置：

- 管理面板监听地址和端口
- 数据库类型和连接信息（支持SQLite和MySQL）
- 日志级别和存储路径
- 安全设置和访问控制

## 系统要求

- Go 1.16+（仅开发环境需要）
- 最低512MB内存
- 50MB磁盘空间

## 浏览器兼容性

- Chrome 80+
- Firefox 78+
- Safari 14+
- Edge 88+

## 开发指南

### 环境准备

```bash
# 克隆项目
git clone https://github.com/SinKingCloud/sinking-panel.git
cd sinking-panel

# 安装后端依赖
cd server
go mod tidy

# 编译后端
go build -o sinking-panel
```

### 项目结构

```
sinking-panel/
├── server/             # 后端代码
│   ├── app/            # 应用核心代码
│   │   ├── command/    # 命令行模块
│   │   ├── constant/   # 常量定义
│   │   ├── http/       # HTTP服务
│   │   ├── model/      # 数据模型
│   │   ├── service/    # 业务逻辑
│   │   └── util/       # 工具函数
│   ├── bootstrap/      # 启动引导
│   ├── config/         # 配置文件
│   ├── public/         # 静态资源
│   └── main.go         # 入口文件
│
└── web/                # 前端代码 (开发中)
```

### API文档

API文档将在前端开发启动时提供，包括：
- RESTful API接口列表
- WebSocket协议说明
- 认证和授权流程

## 参与贡献

我们欢迎各种形式的贡献，特别是：
- 前端开发（React/Vue等现代前端框架）
- 功能增强和缺陷修复
- 文档完善和国际化
- 用户体验优化建议

## 许可证

[MIT License](LICENSE)

## 联系方式

- 项目地址：[GitHub](https://github.com/SinKingCloud/sinking-panel)
- 问题反馈：[Issues](https://github.com/SinKingCloud/sinking-panel/issues)
