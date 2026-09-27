package pgsql

// InstallOptions 只在首次初始化时创建管理员；已有数据的重新安装沿用原账户。
type InstallOptions struct {
	Bind     string                 `json:"bind"`
	Port     int                    `json:"port"`
	Username string                 `json:"username"`
	Password string                 `json:"password"`
	Config   map[string]interface{} `json:"config"`
}

// ImageInfo 是官方镜像的安装预览，不包含宿主机内部路径或密码。
type ImageInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Tags         []string `json:"tags"`
	Version      string   `json:"version"`
	OS           string   `json:"os"`
	Architecture string   `json:"architecture"`
	Entrypoint   []string `json:"entrypoint"`
	Command      []string `json:"command"`
	DataPath     string   `json:"data_path"`
	ConfigPath   string   `json:"config_path"`
	Config       string   `json:"config"`
}

// ApplyResult 区分已生效和等待重启，保存配置不会偷偷重启数据库。
type ApplyResult struct {
	Applied         bool `json:"applied"`
	RestartRequired bool `json:"restart_required"`
}

// installation 仅保存重装和连接所需的非敏感信息，密码独立保存在 secrets/password。
type installation struct {
	Kind     string `json:"kind"`
	ImageID  string `json:"image_id"`
	Version  string `json:"version"`
	Username string `json:"username"`
}
