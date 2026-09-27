package redis

// ImageInfo 保存根据官方 Redis 镜像生成的安装参数，供预览和安装共用。
type ImageInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Tags         []string `json:"tags"`
	Version      string   `json:"version"`
	OS           string   `json:"os"`
	Architecture string   `json:"architecture"`
	User         string   `json:"user"`
	Entrypoint   []string `json:"entrypoint"`
	Command      []string `json:"command"`
	WorkingDir   string   `json:"working_dir"`
	DataPath     string   `json:"data_path"`
	ConfigPath   string   `json:"config_path"`
	Config       string   `json:"config"`
}

// Performance 保存 Redis 的常用性能配置，Maxmemory 的单位为 MiB。
type Performance struct {
	Bind        string `json:"bind"`
	Port        int    `json:"port"`
	Timeout     int    `json:"timeout"`
	Maxclients  int    `json:"maxclients"`
	Databases   int    `json:"databases"`
	Requirepass string `json:"requirepass"`
	Maxmemory   int64  `json:"maxmemory"`
}

// PerformanceUpdate 只更新非 nil 字段，允许显式清空密码和取消内存限制。
type PerformanceUpdate struct {
	Bind        *string `json:"bind"`
	Port        *int    `json:"port"`
	Timeout     *int    `json:"timeout"`
	Maxclients  *int    `json:"maxclients"`
	Databases   *int    `json:"databases"`
	Requirepass *string `json:"requirepass"`
	Maxmemory   *int64  `json:"maxmemory"`
}

// SaveRule 定义在 Seconds 秒内至少 Changes 次修改时生成 RDB 快照。
type SaveRule struct {
	Seconds int `json:"seconds"`
	Changes int `json:"changes"`
}

// Persistence 保存 RDB 和 AOF 设置，AutoAofRewriteMinSize 的单位为 MiB。
type Persistence struct {
	Save                     []SaveRule `json:"save"`
	AppendOnly               bool       `json:"appendonly"`
	AppendFsync              string     `json:"appendfsync"`
	AutoAofRewritePercentage int        `json:"auto_aof_rewrite_percentage"`
	AutoAofRewriteMinSize    int64      `json:"auto_aof_rewrite_min_size"`
}

// PersistenceUpdate 的空 Save 数组关闭 RDB 自动快照，nil 保持原值。
type PersistenceUpdate struct {
	Save                     *[]SaveRule `json:"save"`
	AppendOnly               *bool       `json:"appendonly"`
	AppendFsync              *string     `json:"appendfsync"`
	AutoAofRewritePercentage *int        `json:"auto_aof_rewrite_percentage"`
	AutoAofRewriteMinSize    *int64      `json:"auto_aof_rewrite_min_size"`
}

// ApplyResult 区分配置已应用与需要重启后生效。
type ApplyResult struct {
	Applied         bool `json:"applied"`
	RestartRequired bool `json:"restart_required"`
}

type redisConfig struct {
	Performance
	Persistence
	values     map[string][]string
	directives []string // 记录指令顺序及其他配置原文，识别重复指令和 include 位置变化。
	includes   bool     // include 交由 Redis 展开，变更或重载时通过重启应用。
	content    string
}
