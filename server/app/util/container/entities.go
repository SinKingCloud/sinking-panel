package container

import (
	"errors"
	"io"
	"sync"
	"time"
)

const (
	instanceLockCount       = 64    // 实例操作锁槽位数量
	securityMetadataVersion = 1     // 容器安全元数据版本
	defaultIDMapSize        = 65536 // 默认映射完整的常规容器 UID/GID 范围
)

var errRuntimeCleanupPending = errors.New("容器运行状态等待清理")
var errPlatformRuntimeUnavailable = errors.New("当前平台无法管理容器运行状态")

// ManagerOptions 控制容器运行时的轮询、日志和停止策略。
// 字段为零值时使用内置默认值，便于按需只覆盖单项配置。
type ManagerOptions struct {
	ContainerLogMaxSize    int64         // 单份实例日志保留大小，默认 10 MiB
	RuntimeLogInterval     time.Duration // 实例日志大小检查间隔，默认 5 秒
	RuntimePollInterval    time.Duration // 恢复实例状态检查间隔，默认 1 秒
	RuntimeFailureLimit    int           // 连续状态检查失败次数，默认 3 次
	RuntimeCleanupDelay    time.Duration // 运行状态清理失败后的首次重试间隔，默认 2 秒
	AutoRestartDelay       time.Duration // 自动重启实例的等待间隔，默认 1 秒
	StopGracePeriod        time.Duration // 优雅停止等待时间，默认 10 秒
	ForceStopPeriod        time.Duration // 强制停止和命令清理等待时间，默认 5 秒
	TerminalConsoleTimeout time.Duration // 等待容器返回 PTY 主端的最长时间，默认 5 秒
	TerminalCleanupTimeout time.Duration // 终端启动失败时等待 shell 退出的最长时间，默认 5 秒
	DefaultPidsLimit       int64         // 实例未设置进程数限制时使用的上限，默认 512；-1 表示不限制
	UIDMapStart            int64         // 容器 UID 0 对应的宿主 UID；0 表示从 /etc/subuid 自动读取
	GIDMapStart            int64         // 容器 GID 0 对应的宿主 GID；0 表示从 /etc/subgid 自动读取
	IDMapSize              int64         // User Namespace 映射数量，默认 65536，不能小于 65536
	MountRoots             []string      // 允许 bind mount 的宿主根目录；默认仅允许数据目录下的 volumes
}

// Mount 定义一个宿主机 bind mount。
type Mount struct {
	Source        string `json:"source"`
	Destination   string `json:"destination"`
	ReadOnly      bool   `json:"read_only"`
	DirectoryMode string `json:"directory_mode,omitempty"` // 已废弃；为兼容源码保留，非空值会被拒绝
	FileMode      string `json:"file_mode,omitempty"`      // 已废弃；为兼容源码保留，非空值会被拒绝
}

// ResourceLimits 定义实例的 cgroup 限制；PidsLimit 为 0 时继承 Manager 默认值，其他字段为 0 表示不限制。
type ResourceLimits struct {
	Memory    int64  `json:"memory,omitempty"`     // 字节
	CPUQuota  int64  `json:"cpu_quota,omitempty"`  // 微秒
	CPUPeriod uint64 `json:"cpu_period,omitempty"` // 微秒
	PidsLimit int64  `json:"pids_limit,omitempty"`
}

// RunOptions 定义实例启动参数。
type RunOptions struct {
	ID          string
	Name        string
	ImageID     string
	Mounts      []Mount
	Env         []string
	Command     []string
	WorkingDir  string
	Resources   ResourceLimits
	AutoRestart bool
	// ReadOnly 禁用实例可写层；默认使用独立 overlay 写层保护镜像。
	ReadOnly bool
}

// ExecOptions 定义在运行实例内执行的一次性命令。
type ExecOptions struct {
	Command    []string
	Env        []string
	WorkingDir string
	Timeout    time.Duration
	MaxOutput  int64
}

// ExecResult 返回命令输出与退出状态。
type ExecResult struct {
	Output    string `json:"output"`
	ExitCode  int    `json:"exit_code"`
	TimedOut  bool   `json:"timed_out"`
	Truncated bool   `json:"truncated"`
}

// TerminalSession 是一个运行在容器命名空间内的交互式 PTY 会话。
type TerminalSession interface {
	io.ReadWriteCloser
	Resize(height, width int) error
}

// Stats 返回实例当前资源用量。
type Stats struct {
	CPUUsage    uint64 `json:"cpu_usage"`
	MemoryUsage uint64 `json:"memory_usage"`
	MemoryLimit uint64 `json:"memory_limit"`
	Pids        uint64 `json:"pids"`
}

// Image 保存导入镜像的元数据。
type Image struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Tags            []string `json:"tags"`
	Rootfs          string   `json:"rootfs"`
	OS              string   `json:"os"`
	Architecture    string   `json:"architecture"`
	User            string   `json:"user,omitempty"`
	Entrypoint      []string `json:"entrypoint"`
	Command         []string `json:"command"`
	Env             []string `json:"env"`
	WorkingDir      string   `json:"working_dir"`
	CreatedAt       int64    `json:"created_at"`
	SecurityVersion int      `json:"security_version"`
	UIDMapStart     int64    `json:"uid_map_start"`
	GIDMapStart     int64    `json:"gid_map_start"`
	IDMapSize       int64    `json:"id_map_size"`
}

// Instance 保存实例配置和运行状态。
type Instance struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	ImageID         string         `json:"image_id"`
	Status          Status         `json:"status"`
	PID             int            `json:"pid"`
	Mounts          []Mount        `json:"mounts"`
	CreatedAt       int64          `json:"created_at"`
	StartedAt       int64          `json:"started_at"`
	EndedAt         int64          `json:"ended_at"`
	ExitCode        int            `json:"exit_code"`
	Error           string         `json:"error,omitempty"`
	LogPath         string         `json:"log_path"`
	Command         []string       `json:"command,omitempty"`
	Env             []string       `json:"env,omitempty"`
	WorkingDir      string         `json:"working_dir,omitempty"`
	Resources       ResourceLimits `json:"resources,omitempty"`
	AutoRestart     bool           `json:"auto_restart"`
	WritableLayer   bool           `json:"writable_layer"`
	Rootfs          string         `json:"rootfs,omitempty"`
	UpperDir        string         `json:"upper_dir,omitempty"`
	WorkDir         string         `json:"work_dir,omitempty"`
	Generation      uint64         `json:"generation,omitempty"`
	SecurityVersion int            `json:"security_version"`
	UIDMapStart     int64          `json:"uid_map_start"`
	GIDMapStart     int64          `json:"gid_map_start"`
	IDMapSize       int64          `json:"id_map_size"`
}

// securityMetadata 固定容器数据目录使用的 User Namespace 映射。
type securityMetadata struct {
	Version     int   `json:"version"`
	UIDMapStart int64 `json:"uid_map_start"`
	GIDMapStart int64 `json:"gid_map_start"`
	IDMapSize   int64 `json:"id_map_size"`
}

// Manager 管理镜像、实例元数据和平台运行时状态。
type Manager struct {
	dataRoot          string
	imagesRoot        string
	instancesRoot     string
	runtimeRoot       string
	volumesRoot       string
	mountsRoot        string
	mountRoots        []string
	installation      string
	options           ManagerOptions
	requestedSecurity securityMetadata
	security          securityMetadata
	platformSecurity  interface{}

	mu             sync.RWMutex
	importMu       sync.Mutex
	imageMu        sync.RWMutex
	mountMu        sync.Mutex
	operationLocks [instanceLockCount]sync.Mutex
	images         map[string]*Image
	instances      map[string]*Instance
	runtime        map[string]interface{}
	autoRestart    map[string]uint64
	generation     uint64
}
