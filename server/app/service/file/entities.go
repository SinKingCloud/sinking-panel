package file

import (
	"sync"
	"time"
)

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

type uploadState struct {
	sync.Mutex
	locks       map[string]*uploadSessionLock
	lastCleanup time.Time
}

type uploadSessionLock struct {
	mutex sync.Mutex
	refs  int
}
