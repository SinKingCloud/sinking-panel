package file

// File 描述文件系统对象的元数据信息
type File struct {
	Name       string  `json:"name"`        // 文件/目录名称（不含路径）
	Size       int64   `json:"size"`        // 文件大小（字节），目录为0
	Mode       uint32  `json:"mode"`        // 权限模式（八进制表示，例如 0644）
	IsDir      bool    `json:"is_dir"`      // 是否为目录类型
	UpdateTime int64   `json:"update_time"` // 最后修改时间（Unix时间戳）
	Child      []*File `json:"child"`       // 子文件列表（仅当IsDir为true时有效）
}
