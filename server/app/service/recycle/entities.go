package recycle

import "server/app/util/str"

// File 描述文件系统对象的元数据信息
type File struct {
	ID         string       `json:"id"`          // 加密路径
	Name       string       `json:"name"`        // 原始名称
	Path       string       `json:"path"`        // 原始路径
	Size       int64        `json:"size"`        // 文件大小（字节），目录为0
	IsDir      bool         `json:"is_dir"`      // 是否为目录类型
	DeleteTime str.DateTime `json:"update_time"` //删除时间
}
