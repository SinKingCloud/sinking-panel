package container

import (
	"context"
	"io"
	"sync"
	"time"
)

const (
	instanceLockCount       = 64                  // 实例操作锁槽位数量
	securityMetadataVersion = 1                   // 容器安全元数据版本
	defaultIDMapSize        = 65536               // 默认映射完整的常规容器 UID/GID 范围
	timezoneMountLabel      = "sinking.timezone=" // 运行时系统时区挂载路径，创建镜像时排除
	containerTimezonePath   = "/etc/localtime"    //运行时系统时区时间信息
)

// ManagerOptions 控制容器运行时的轮询、日志和停止策略。
// 字段为零值时使用内置默认值，便于按需只覆盖单项配置。
type ManagerOptions struct {
	DisableAutoRestart     bool          // 仅接管或清理时关闭自动启动，不修改实例保存的保活策略
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
	UIDMapStart            int64         // 容器 UID 0 对应的宿主 UID；0 表示自动读取或分配 /etc/subuid 映射
	GIDMapStart            int64         // 容器 GID 0 对应的宿主 GID；0 表示自动读取或分配 /etc/subgid 映射
	IDMapSize              int64         // User Namespace 映射数量，默认 65536，不能小于 65536
	MountRoots             []string      // 允许 bind mount 的宿主根目录；默认仅允许数据目录下的 volumes
}

// Mount 定义一个宿主机 bind mount。
type Mount struct {
	Source        string `json:"source"`                   // 宿主机挂载源绝对路径，须位于允许的挂载根目录内
	Destination   string `json:"destination"`              // 容器内挂载目标绝对路径
	ReadOnly      bool   `json:"read_only"`                // 是否以只读方式挂载
	DirectoryMode string `json:"directory_mode,omitempty"` // 已废弃；为兼容源码保留，非空值会被拒绝
	FileMode      string `json:"file_mode,omitempty"`      // 已废弃；为兼容源码保留，非空值会被拒绝
}

// ResourceLimits 定义实例的 cgroup 限制；PidsLimit 为 0 时继承 Manager 默认值，其他字段为 0 表示不限制。
type ResourceLimits struct {
	Memory    int64  `json:"memory,omitempty"`     // 内存上限，单位字节；0 表示不限制
	CPUQuota  int64  `json:"cpu_quota,omitempty"`  // 每个调度周期可使用的 CPU 时间，单位微秒；0 表示不限制
	CPUPeriod uint64 `json:"cpu_period,omitempty"` // CPU 调度周期，单位微秒；设置配额时默认为 100000
	PidsLimit int64  `json:"pids_limit,omitempty"` // 进程数上限；0 继承 Manager 默认值
}

// RunOptions 定义实例启动参数。
type RunOptions struct {
	ID          string         // 实例唯一标识，同时用于实例目录名称
	Name        string         // 实例显示名称；空值使用镜像名称或首个标签
	ImageID     string         // 镜像 ID、名称、标签或本机归档绝对路径；Run 解析后保存镜像 ID
	Mounts      []Mount        // 宿主机文件或目录挂载配置
	Env         []string       // KEY=VALUE 格式的环境变量，覆盖镜像中的同名变量
	Command     []string       // 覆盖镜像 Cmd，仍追加在 Entrypoint 后；空值继承镜像 Cmd
	WorkingDir  string         // 容器内工作目录；空值继承镜像，仍为空时使用 /
	User        string         // 空值继承镜像运行用户
	Resources   ResourceLimits // 实例资源限制
	AutoRestart bool           // 是否在管理器恢复或进程退出后自动启动；手动停止时暂停
	ReadOnly    bool           // 是否禁用实例可写层；默认使用独立 overlay 写层保护镜像
	CreateOnly  bool           // 仅创建新实例；同标识已存在时拒绝覆盖，默认保留 Run 的复用行为
}

// ExecOptions 定义在运行实例内执行的一次性命令。
type ExecOptions struct {
	Command    []string      // 待执行程序及参数，首项为程序路径或名称
	Env        []string      // KEY=VALUE 格式的环境变量，覆盖实例和镜像中的同名变量
	Stdin      io.Reader     // 可通过标准输入传入敏感参数，避免出现在进程命令行
	Stderr     io.Writer     // 非 nil 时单独接收错误输出；默认与标准输出合并
	WorkingDir string        // 容器内工作目录；空值依次继承实例、镜像，最终使用 /
	Timeout    time.Duration // 命令执行超时，默认 30 秒
	MaxOutput  int64         // 返回输出的最大字节数，默认 512 KiB
}

// ExecResult 返回命令输出与退出状态。
type ExecResult struct {
	Output    string `json:"output"`    // 标准输出；未单独指定 Stderr 时包含错误输出
	ExitCode  int    `json:"exit_code"` // 命令退出码，未取得正常退出码时为 -1
	TimedOut  bool   `json:"timed_out"` // 是否因超过 Timeout 而终止
	Truncated bool   `json:"truncated"` // 输出是否超过 MaxOutput 并被截断
}

// TerminalSession 是一个运行在容器命名空间内的交互式 PTY 会话。
type TerminalSession interface {
	io.ReadWriteCloser              // 读取终端输出、写入终端输入及关闭会话，不停止容器主进程
	Resize(height, width int) error // 调整终端尺寸，height 为行数，width 为列数
}

// Status 返回容器运行状态及资源采样；CPU 和磁盘 I/O 为累计值，可用相邻采样的差值计算使用率或速率。
// 容器重启后累计值会重置，计算差值时应跳过计数下降的采样。
// ResourcesAvailable 为 false 时不使用资源指标；内核未提供的可选指标为 0。
type Status struct {
	ReadTime time.Time `json:"read_time"` // 采样时间，JSON 使用带时区的 RFC3339 格式

	ID                 string `json:"id"`                       // 容器实例标识
	Status             State  `json:"status"`                   // 当前运行状态
	PID                int    `json:"pid"`                      // 容器主进程在宿主机中的 PID，未运行时为 0
	StartedAt          int64  `json:"started_at"`               // 本次启动时间，Unix 秒时间戳
	EndedAt            int64  `json:"ended_at"`                 // 本次结束时间，Unix 秒时间戳；尚未结束时为 0
	Uptime             int64  `json:"uptime"`                   // 本次运行时长，单位秒；结束后保留该次时长
	ExitCode           int    `json:"exit_code"`                // 主进程退出码，运行中不作为退出结果
	Error              string `json:"error,omitempty"`          // 最近一次启动、运行或状态刷新错误
	AutoRestart        bool   `json:"auto_restart"`             // 是否配置自动重启
	ManualStop         bool   `json:"manual_stop"`              // 是否因手动停止而暂停自动重启
	RestartPending     bool   `json:"restart_pending"`          // 当前运行代次是否有等待执行的自动重启任务
	Generation         uint64 `json:"generation"`               // 运行代次，变化时应重新计算资源差值
	ResourcesAvailable bool   `json:"resources_available"`      // 本次是否成功采集资源指标
	ResourceError      string `json:"resource_error,omitempty"` // 资源采集失败原因，不改变容器运行状态

	CPUUsage            uint64 `json:"cpu_usage"`             // 累计 CPU 使用时间，单位纳秒
	CPUUserUsage        uint64 `json:"cpu_user_usage"`        // 累计用户态 CPU 使用时间，单位纳秒
	CPUSystemUsage      uint64 `json:"cpu_system_usage"`      // 累计内核态 CPU 使用时间，单位纳秒
	CPUPeriods          uint64 `json:"cpu_periods"`           // 累计 CPU 带宽调度周期数
	CPUThrottledPeriods uint64 `json:"cpu_throttled_periods"` // 触发 CPU 配额限流的累计周期数
	CPUThrottledTime    uint64 `json:"cpu_throttled_time"`    // 累计 CPU 限流时间，单位纳秒

	MemoryUsage      uint64 `json:"memory_usage"`       // 当前内存总用量，包含缓存，单位字节
	MemoryLimit      uint64 `json:"memory_limit"`       // 内存上限，单位字节；0 表示未限制或未取得统计
	MemoryPeak       uint64 `json:"memory_peak"`        // 内存用量峰值，单位字节；旧内核未提供时为 0
	MemoryCache      uint64 `json:"memory_cache"`       // 文件缓存用量，单位字节
	MemoryRSS        uint64 `json:"memory_rss"`         // 匿名常驻内存，单位字节；cgroup v1 使用 rss，v2 使用 anon
	MemoryWorkingSet uint64 `json:"memory_working_set"` // 内存总用量扣除非活跃文件缓存，单位字节
	SwapUsage        uint64 `json:"swap_usage"`         // 仅 Swap 的用量，不包含物理内存，单位字节

	Pids      uint64 `json:"pids"`       // cgroup 当前任务数，包含进程和线程
	PidsLimit uint64 `json:"pids_limit"` // 任务数上限；0 表示未限制或未取得统计

	IOReadBytes       uint64 `json:"io_read_bytes"`       // 各块设备累计读取字节数
	IOWriteBytes      uint64 `json:"io_write_bytes"`      // 各块设备累计写入字节数
	IOReadOperations  uint64 `json:"io_read_operations"`  // 各块设备累计读取次数
	IOWriteOperations uint64 `json:"io_write_operations"` // 各块设备累计写入次数
}

// Image 保存导入镜像的元数据。
type Image struct {
	ID              string            `json:"id"`                      // 镜像唯一标识，格式为 sha256:摘要
	Name            string            `json:"name"`                    // 镜像名称，导入时取首个仓库标签
	Tags            []string          `json:"tags"`                    // 镜像仓库标签
	Rootfs          string            `json:"rootfs"`                  // 解包后的镜像根文件系统宿主机路径
	OS              string            `json:"os"`                      // 镜像目标操作系统
	Architecture    string            `json:"architecture"`            // 镜像目标 CPU 架构
	User            string            `json:"user,omitempty"`          // 镜像默认运行用户，可包含用户组
	Entrypoint      []string          `json:"entrypoint"`              // 镜像入口程序及固定参数
	Command         []string          `json:"command"`                 // 镜像默认 Cmd，追加在 Entrypoint 后
	Env             []string          `json:"env"`                     // 镜像默认环境变量，格式为 KEY=VALUE
	WorkingDir      string            `json:"working_dir"`             // 镜像默认工作目录，位于容器内
	Labels          map[string]string `json:"labels,omitempty"`        // 镜像声明的标签元数据
	Volumes         []string          `json:"volumes,omitempty"`       // 镜像声明的数据卷路径，仅保存元数据，不自动挂载
	ExposedPorts    []string          `json:"exposed_ports,omitempty"` // 镜像声明的端口及协议，仅保存元数据，不自动映射
	CreatedAt       int64             `json:"created_at"`              // 镜像创建时间，Unix 秒时间戳；归档未提供时使用导入时间
	SecurityVersion int               `json:"security_version"`        // 镜像根文件系统使用的安全元数据版本
	UIDMapStart     int64             `json:"uid_map_start"`           // 镜像容器 UID 0 对应的宿主 UID
	GIDMapStart     int64             `json:"gid_map_start"`           // 镜像容器 GID 0 对应的宿主 GID
	IDMapSize       int64             `json:"id_map_size"`             // 镜像 UID/GID 映射范围大小
}

// Instance 用于容器列表，仅返回展示所需的摘要。
type Instance struct {
	ID          string `json:"id"`           // 容器实例标识
	Name        string `json:"name"`         // 容器显示名称
	ImageID     string `json:"image_id"`     // 关联的镜像 ID
	ImageName   string `json:"image_name"`   // 关联的镜像名称
	Status      State  `json:"status"`       // 当前运行状态
	CreatedAt   int64  `json:"created_at"`   // 创建时间，Unix 秒时间戳
	AutoRestart bool   `json:"auto_restart"` // 是否配置自动重启
	ManualStop  bool   `json:"manual_stop"`  // 是否因手动停止而暂停自动重启
}

// Info 保存实例的完整配置和运行状态，供 Info 查询详情及持久化使用。
type Info struct {
	ID              string         `json:"id"`                    // 实例唯一标识
	Name            string         `json:"name"`                  // 实例显示名称
	ImageID         string         `json:"image_id"`              // 实例关联的镜像 ID
	Status          State          `json:"status"`                // 实例当前运行状态
	PID             int            `json:"pid"`                   // 容器主进程在宿主机中的 PID；未运行时为 0
	Mounts          []Mount        `json:"mounts"`                // 实例绑定的宿主机文件和目录
	CreatedAt       int64          `json:"created_at"`            // 实例首次创建时间，Unix 秒时间戳
	StartedAt       int64          `json:"started_at"`            // 本次启动时间，Unix 秒时间戳；尚未启动时为 0
	EndedAt         int64          `json:"ended_at"`              // 本次结束时间，Unix 秒时间戳；运行中为 0
	ExitCode        int            `json:"exit_code"`             // 容器主进程退出码
	Error           string         `json:"error,omitempty"`       // 最近一次启动、运行或清理失败的信息
	LogPath         string         `json:"log_path"`              // 实例输出日志的宿主机路径
	Command         []string       `json:"command,omitempty"`     // 实例覆盖的镜像 Cmd；空值继承镜像
	Env             []string       `json:"env,omitempty"`         // 实例环境变量，覆盖镜像中的同名变量
	WorkingDir      string         `json:"working_dir,omitempty"` // 容器内工作目录；空值继承镜像
	User            string         `json:"user,omitempty"`        // 容器运行用户；空值继承镜像
	Resources       ResourceLimits `json:"resources,omitempty"`   // 实例资源限制
	AutoRestart     bool           `json:"auto_restart"`          // 是否在管理器恢复或进程退出后自动启动
	ManualStop      bool           `json:"manual_stop,omitempty"` // 手动停止后不自动恢复，保留原保活策略
	WritableLayer   bool           `json:"writable_layer"`        // 是否启用实例独立的 overlay 可写层
	Rootfs          string         `json:"rootfs,omitempty"`      // 实例实际使用的根文件系统宿主机路径
	UpperDir        string         `json:"upper_dir,omitempty"`   // overlay 可写层目录，保存实例文件改动
	WorkDir         string         `json:"work_dir,omitempty"`    // overlay 挂载所需的工作目录
	Generation      uint64         `json:"generation,omitempty"`  // 运行代次，用于识别过期退出事件和重启请求
	SecurityVersion int            `json:"security_version"`      // 实例使用的安全元数据版本
	UIDMapStart     int64          `json:"uid_map_start"`         // 容器 UID 0 对应的宿主 UID
	GIDMapStart     int64          `json:"gid_map_start"`         // 容器 GID 0 对应的宿主 GID
	IDMapSize       int64          `json:"id_map_size"`           // 容器 UID/GID 映射范围大小
}

// securityMetadata 固定容器数据目录使用的 User Namespace 映射。
type securityMetadata struct {
	Version     int   `json:"version"`       // 安全元数据格式版本
	UIDMapStart int64 `json:"uid_map_start"` // 容器 UID 0 对应的宿主 UID
	GIDMapStart int64 `json:"gid_map_start"` // 容器 GID 0 对应的宿主 GID
	IDMapSize   int64 `json:"id_map_size"`   // UID/GID 映射范围大小
}

// Manager 管理镜像、实例元数据和平台运行时状态。
type Manager struct {
	dataRoot          string           // 容器数据根目录，使用解析后的绝对路径
	imagesRoot        string           // 镜像元数据和根文件系统目录
	instancesRoot     string           // 实例元数据、日志和可写层目录
	runtimeRoot       string           // 平台运行时状态目录
	volumesRoot       string           // 默认允许挂载的持久化数据目录
	mountsRoot        string           // 固定挂载源时使用的临时目录
	mountRoots        []string         // 校验后的宿主机挂载根目录列表
	installation      string           // 根据数据根目录生成的安装标识，隔离管理标记和 cgroup 路径
	options           ManagerOptions   // 补齐默认值后的运行时配置
	requestedSecurity securityMetadata // 调用方显式传入的映射配置，用于校验持久化映射
	security          securityMetadata // 当前数据目录实际使用的安全映射
	platformSecurity  interface{}      // 平台特有的安全配置和运行环境状态

	mu             sync.RWMutex                  // 保护内存中的镜像、实例、运行时、终端及关闭状态
	lifecycleMu    sync.RWMutex                  // 协调运行请求与 Close，关闭期间阻止新增运行操作
	importMu       sync.Mutex                    // 串行导入镜像
	imageMu        sync.RWMutex                  // 协调镜像文件读写，避免使用期间删除或替换镜像
	mountMu        sync.Mutex                    // 串行检查挂载源冲突及启动实例
	operationLocks [instanceLockCount]sync.Mutex // 按实例 ID 哈希分槽，串行同一实例的操作
	images         map[string]*Image             // 镜像 ID 到镜像元数据的索引
	instances      map[string]*Info              // 实例 ID 到实例元数据的索引
	runtime        map[string]interface{}        // 实例 ID 到平台运行时对象的索引
	terminals      map[TerminalSession]struct{}  // 当前交互终端集合，管理器退出时统一关闭
	autoRestart    map[string]uint64             // 待自动重启的实例及对应运行代次
	generation     uint64                        // 单调递增的运行代次计数器，0 不作为有效代次
	ctx            context.Context               // 管理器后台任务共用的生命周期上下文
	cancel         context.CancelFunc            // 关闭管理器时取消后台任务
	workers        sync.WaitGroup                // 等待受管理的后台任务退出
	closing        bool                          // 管理器是否已进入关闭流程
	closeOnce      sync.Once                     // 保证 Close 只执行一次关闭流程
	closeErr       error                         // Close 汇总的关闭错误，后续调用复用
}
