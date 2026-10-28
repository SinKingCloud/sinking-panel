package plugin

import (
	"context"
	"io/fs"
	webContext "server/app/util/context"

	"github.com/traefik/yaegi/interp"
)

// Plugin 由包内 New 返回，生命周期由宿主主动调用，Actions 只注册业务接口。
type Plugin struct {
	Install   func(context.Context, string) (interface{}, error)
	Uninstall func(context.Context, string) (interface{}, error)
	Upgrade   func(context.Context, string) (interface{}, error) // 已有软件导入新包时执行，params 包含 previous_version 和 version
	Start     func(context.Context, string) (interface{}, error) // 启动后台任务；ctx 持续到 Stop 或宿主退出
	Stop      func(context.Context, string) (interface{}, error) // 后台 ctx 已取消，等待任务退出并释放资源
	Actions   map[string]func(*webContext.Context)               // 使用面板上下文绑定参数、校验并返回响应
}

// Manifest 描述插件入口和运行条件，随插件代码一起发布。
type Manifest struct {
	Id        string   `json:"id"`
	Name      string   `json:"name"`
	Version   string   `json:"version"`
	Frontend  string   `json:"frontend"`
	Icon      string   `json:"icon"` // 可选图标，使用 web/ 下的包内相对路径
	Platforms []string `json:"platforms"`
	Requires  []string `json:"requires"` // 依赖的软件标识，安装状态与启动顺序由宿主管理
}

// Package 插件包及宿主适用性；Directory 为相对于管理器根目录的路径。
type Package struct {
	Directory         string
	Manifest          *Manifest
	Available         bool
	UnavailableReason string
}

// Options 由宿主注入业务能力，通用管理器不依赖具体服务或容器。
type Options struct {
	ManifestName string
	// 零值沿用通用插件的 32 MiB ZIP / 64 MiB 展开上限。
	// 调大后仅 images/ 中的镜像归档享有额外空间，源码仍限制为 64 MiB。
	MaxArchiveSize  int64
	MaxExpandedSize int64
	Symbols         interp.Exports
	App             func(string) interface{}
}

// moduleFS 将模块源码映射到 Yaegi 的 GOPATH 视图，依赖仍从 server/vendor 读取。
type moduleFS struct {
	fs.FS
	module string
}
