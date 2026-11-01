package oss

import (
	"io"
	repositoryOss "server/app/repository/oss"
	utilOss "server/app/util/oss"
	"server/app/util/str"
)

// OssListItem 对象存储列表项，包含关联密钥的名称和厂商。
type OssListItem struct {
	*repositoryOss.Oss
	SecretName string `json:"secret_name"` // 密钥名称
	Provider   *int   `json:"provider"`    // 厂商，关联密钥不存在时为空
}

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

// UploadOptions 配置上传方式；默认同步上传 Body，大文件由 util 自动分片。
// 上传直接覆盖同名文件；server 使用 Path 创建后台任务，其余分片操作通过 Action 指定。
type UploadOptions struct {
	utilOss.UploadOptions
	Action         string         // file、server、init、part、complete、abort、parts、uploads
	Body           io.Reader      // file、part 的数据流，由调用方关闭
	Size           int64          // 数据流字节数；file 允许 -1 表示未知长度
	Path           string         // server 的服务器文件绝对路径
	UploadId       string         // 分片会话 ID
	Number         int            // 上传分片编号，从 1 开始
	Parts          []utilOss.Part // 合并时提供的分片列表
	Marker         int            // parts 的分片游标
	Limit          int            // parts、uploads 的每页数量
	Prefix         string         // uploads 的文件前缀
	KeyMarker      string         // uploads 的文件游标
	UploadIdMarker string         // uploads 的会话游标
}

// DownloadOptions 默认返回下载流；server 将完整文件下载到 Path 并返回后台任务 ID。
type DownloadOptions struct {
	utilOss.GetOptions
	Action    string // file、server
	Path      string // server 的服务器文件绝对路径
	Overwrite bool   // server 是否允许覆盖同名文件
}

// DownloadResult 流式下载返回 Download（调用方关闭 Body），后台下载返回 TaskId。
type DownloadResult struct {
	*utilOss.Download
	TaskId string
}
