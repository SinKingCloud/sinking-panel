package oss

import (
	utilOss "server/app/util/oss"
	"server/app/util/str"
)

// Oss 对象存储详情，Config 返回对应厂商的连接配置。
type Oss struct {
	Id         int64             `json:"id"`          // 存储ID
	Name       string            `json:"name"`        // 存储名称
	SecretId   int64             `json:"secret_id"`   // 密钥ID
	Config     map[string]string `json:"config"`      // 当前厂商的连接配置
	CreateTime str.DateTime      `json:"create_time"` // 创建时间
	UpdateTime str.DateTime      `json:"update_time"` // 更新时间
}

// File 云端文件列表项，目录与文件使用相同的分页结构。
type File struct {
	utilOss.File
	Directory bool `json:"directory"` // 是否为目录前缀
}
