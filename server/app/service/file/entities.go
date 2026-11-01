package file

import (
	"os"
	"server/app/util/file"
	"sync"
	"time"
)

// File 文件浏览条目，磁盘入口通过 Path 提供实际访问路径。
type File struct {
	file.File
	Path string `json:"path,omitempty"` // 磁盘入口的绝对路径
}

// UpdateFile 文件重命名、权限和文本内容更新参数。
type UpdateFile struct {
	Name        string  // 新文件名
	Permissions string  // 八进制权限
	Content     *string // 文本内容，nil 表示不修改
	Cursor      *int64  // 起始游标
	NextCursor  *int64  // 下一游标
	Version     string  // 文件版本
}

// UpdatedFile 文件更新结果，仅编辑内容时返回游标与版本。
type UpdatedFile struct {
	Operations []string `json:"operations"`
	Path       string   `json:"path"`
	Cursor     *int64   `json:"cursor,omitempty"`
	NextCursor *int64   `json:"next_cursor,omitempty"`
	Eof        *bool    `json:"eof,omitempty"`
	Version    *string  `json:"version,omitempty"`
}

type Disk struct {
	Filesystem string   `json:"filesystem"` //分区
	Type       string   `json:"type"`       //文件系统类型
	Path       string   `json:"path"`       //路径
	Size       struct { //存储信息
		Total      int64 `json:"total"`      //总空间大小
		Used       int64 `json:"used"`       //已用大小
		UnUsed     int64 `json:"un_used"`    //可用大小
		Percentage int   `json:"percentage"` //已用百分比(0-100)
	} `json:"size"`
	Inodes struct { //inode信息
		Total      int64 `json:"total"`      //Inode总数
		Used       int64 `json:"used"`       //已用Inode
		UnUsed     int64 `json:"un_used"`    //可用Inode
		Percentage int   `json:"percentage"` //已用百分比
	} `json:"inodes"`
}

type UploadMeta struct {
	UploadID    string `json:"upload_id"`
	Path        string `json:"path"`
	FileName    string `json:"file_name"`
	TotalSize   int64  `json:"total_size"`
	ChunkSize   int64  `json:"chunk_size"`
	TotalChunks int    `json:"total_chunks"`
	FileHash    string `json:"file_hash"`
}

type UploadedFile struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	FileHash string `json:"file_hash"`
}

type PreviewSign struct {
	Path      string
	FileName  string
	Download  bool
	ExpiresAt time.Time
}

// PreviewFile 已校验的预览文件，调用方负责关闭 File。
type PreviewFile struct {
	File     *os.File
	Name     string
	ModTime  time.Time
	Download bool
}

type uploadState struct {
	sync.Mutex
	locks       map[string]*uploadSessionLock
	lastCleanup time.Time
}

type uploadSessionLock struct {
	mutex sync.Mutex
	refs  int
}
