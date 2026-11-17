package backup

import (
	"context"
	"encoding/json"
	"io"
)

// SiteOptions 保存网站选择参数，全部模式会纳入后来新增的网站。
type SiteOptions struct {
	AllSites bool    `json:"all_sites"`
	SiteIds  []int64 `json:"site_ids"`
}

// SiteData 保存生成时的网站信息及副本位置，以 JSON 字符串存入备份表。
type SiteData struct {
	SiteName  string `json:"site_name"`  // 网站名称快照
	TaskId    int64  `json:"task_id"`    // 来源计划任务ID，0 表示网站手动或自动备份
	Root      string `json:"root"`       // 归档来源，恢复时核对目录名称
	LocalPath string `json:"local_path"` // 相对于网站备份根目录的文件路径
	OssId     int64  `json:"oss_id"`     // 云端存储实例ID
	OssPath   string `json:"oss_path"`   // 云端对象键
	Checksum  string `json:"checksum"`   // SHA-256 校验值
}

// PathOptions 保存自定义备份的文件和目录路径。
type PathOptions struct {
	Paths []string `json:"paths"`
}

// Config 保存备份类型、业务参数、文件存储和副本策略。
type Config struct {
	Type      string          `json:"type"`
	Options   json.RawMessage `json:"options"`
	LocalKeep int             `json:"local_keep"` // 本地副本数；云端备份允许为 0。
	OssKeep   int             `json:"oss_keep"`   // 云端副本数，仅 OSS 存储使用。
	Storage   string          `json:"storage"`    // 存储方式：local 或 oss。
	OssId     int64           `json:"oss_id"`     // 已保存的 OSS 存储实例 ID。
	Prefix    string          `json:"prefix"`     // OSS 目录；空值表示桶根目录。
}

// Target 定义文件的归属，同一对象可由多个计划分别保留副本。
type Target struct {
	Key         string // 业务标识，作为第一层目录，如 site、file。
	ID          string // 可选对象标识；非空时作为第二层目录。
	PlanId      int64  // 副本归属的计划 ID，Save 时必须为正数。
	Extension   string // 文件扩展名，默认 .bak，如 .tar.gz。
	ContentType string // OSS 文件内容类型，默认 application/octet-stream。
}

// File 返回已发布的备份信息，不包含文件内容。
type File struct {
	Name    string `json:"name"`              // 备份文件名。
	Size    int64  `json:"size"`              // 文件字节数。
	Time    int64  `json:"time"`              // 创建时间，Unix 毫秒。
	Storage string `json:"storage"`           // 存储方式。
	PlanId  int64  `json:"plan_id"`           // 创建该副本的计划 ID。
	Warning string `json:"warning,omitempty"` // 仅保留本地副本或清理副本时的提示。
}

type pathBackupReader struct {
	ctx    context.Context
	reader io.Reader
}
