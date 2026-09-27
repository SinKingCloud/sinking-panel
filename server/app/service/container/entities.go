package container

import (
	"server/app/util/container"
	"server/app/util/str"
)

// Container 软件管理列表数据。
type Container struct {
	Id          string          `json:"id"`           // 容器 ID
	Name        string          `json:"name"`         // 容器名称
	ImageId     string          `json:"image_id"`     // 镜像 ID
	ImageName   string          `json:"image_name"`   // 镜像名称
	Status      container.State `json:"status"`       // 运行状态
	CreateTime  str.DateTime    `json:"create_time"`  // 创建时间
	AutoRestart bool            `json:"auto_restart"` // 是否配置自动重启
	ManualStop  bool            `json:"manual_stop"`  // 是否因手动停止而暂停自动重启
}

// Image 镜像列表摘要。
type Image struct {
	Id         string       `json:"id"`          // 镜像 ID
	Name       string       `json:"name"`        // 镜像名称
	Tags       []string     `json:"tags"`        // 仓库标签
	CreateTime str.DateTime `json:"create_time"` // 创建时间
}

// ImageInfo 镜像详情，包含创建容器所需的默认配置。
type ImageInfo struct {
	Image
	User         string            `json:"user,omitempty"`          // 默认运行用户
	Entrypoint   []string          `json:"entrypoint"`              // 入口程序
	Command      []string          `json:"command"`                 // 默认命令
	Env          []string          `json:"env"`                     // 环境变量
	WorkingDir   string            `json:"working_dir"`             // 工作目录
	Labels       map[string]string `json:"labels,omitempty"`        // 标签元数据
	Volumes      []string          `json:"volumes,omitempty"`       // 声明的数据卷
	ExposedPorts []string          `json:"exposed_ports,omitempty"` // 声明的端口
}

// SelectContainer 容器查询条件；nil 字段不参与筛选。
type SelectContainer struct {
	Keyword         *string          // 模糊匹配容器 ID、名称、镜像 ID 和镜像名称
	ID              *string          // 容器 ID，精确匹配
	Name            *string          // 容器名称，模糊匹配
	ImageID         *string          // 镜像 ID，精确匹配
	ImageName       *string          // 镜像名称，模糊匹配
	Status          *container.State // 容器运行状态，精确匹配
	AutoRestart     *bool            // 是否配置自动重启
	ManualStop      *bool            // 是否因手动停止而暂停自动重启
	CreateTimeStart *int64           // 创建时间下限，Unix 秒时间戳，包含该时间
	CreateTimeEnd   *int64           // 创建时间上限，Unix 秒时间戳，包含该时间
}

// SelectImage 镜像查询条件；nil 字段不参与筛选。
type SelectImage struct {
	Keyword         *string // 模糊匹配镜像 ID、名称和标签
	ID              *string // 镜像 ID，精确匹配
	Name            *string // 镜像名称，模糊匹配
	Tag             *string // 镜像标签，精确匹配
	CreateTimeStart *int64  // 创建时间下限，Unix 秒时间戳，包含该时间
	CreateTimeEnd   *int64  // 创建时间上限，Unix 秒时间戳，包含该时间
}
