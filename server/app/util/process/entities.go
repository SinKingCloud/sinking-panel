package process

import (
	"os/exec"
	"sync"
	"time"
)

const (
	DefaultMaxRetries   = 3                // 默认连续失败后的最大重试次数
	retryResetAfter     = time.Minute      // 稳定运行后重置连续失败计数
	defaultRestartDelay = 5 * time.Second  // 默认异常重启等待时间
	defaultStopTimeout  = 10 * time.Second // 默认优雅停止等待时间
	forceWaitTimeout    = 5 * time.Second  // 强制停止后的最长等待时间
)

// State 表示受管进程的当前状态。
type State string

const (
	StateStarting   State = "starting"   // 正在启动
	StateRunning    State = "running"    // 正在运行
	StateStopping   State = "stopping"   // 正在停止
	StateRestarting State = "restarting" // 等待重新启动
	StateStopped    State = "stopped"    // 已停止
	StateFailed     State = "failed"     // 运行失败
)

// Config 定义一个需要保活的系统进程。
// Command 是交给系统 shell 解析的完整命令行。
type Config struct {
	ID           string        `json:"id"`                      // 进程唯一标识
	Command      string        `json:"command"`                 // 交给系统 shell 执行的完整命令
	WorkingDir   string        `json:"working_dir,omitempty"`   // 进程工作目录
	Env          []string      `json:"env,omitempty"`           // KEY=VALUE 格式的环境变量
	AutoRestart  bool          `json:"auto_restart"`            // 退出后是否自动重新启动
	MaxRetries   int           `json:"max_retries"`             // 连续失败后的最大重试次数，0 使用默认值
	RestartDelay time.Duration `json:"restart_delay,omitempty"` // 自动重启等待时间
	StopTimeout  time.Duration `json:"stop_timeout,omitempty"`  // 优雅停止等待时间
	LogPath      string        `json:"log_path,omitempty"`      // 标准输出和错误日志路径
}

// Status 是进程状态的只读快照。
type Status struct {
	ID             string    `json:"id"`                    // 进程唯一标识
	State          State     `json:"state"`                 // 当前运行状态
	PID            int       `json:"pid"`                   // 当前进程 ID，未运行时为 0
	Command        string    `json:"command"`               // 当前启动命令
	WorkingDir     string    `json:"working_dir,omitempty"` // 当前工作目录
	AutoRestart    bool      `json:"auto_restart"`          // 是否启用异常自动重启
	MaxRetries     int       `json:"max_retries"`           // 连续失败后的最大重试次数
	LogPath        string    `json:"log_path,omitempty"`    // 标准输出和错误日志路径
	StartedAt      time.Time `json:"started_at,omitempty"`  // 最近启动时间
	ExitedAt       time.Time `json:"exited_at,omitempty"`   // 最近退出时间
	ExitCode       int       `json:"exit_code"`             // 最近退出码，未知时为 -1
	RestartCount   uint64    `json:"restart_count"`         // 连续失败后的自动重试次数，稳定运行后清零
	RetryExhausted bool      `json:"retry_exhausted"`       // 是否因重试耗尽而停止
	Error          string    `json:"error,omitempty"`       // 最近一次运行错误
}

// Manager 按 ID 管理继承当前用户身份运行的系统进程。
type Manager struct {
	operationMu sync.Mutex
	mu          sync.RWMutex
	processes   map[string]*managedProcess
}

type managedProcess struct {
	config         Config
	state          State
	pid            int
	startedAt      time.Time
	exitedAt       time.Time
	exitCode       int
	restartCount   uint64
	retryExhausted bool
	lastError      string
	desired        bool
	suspended      bool
	generation     uint64
	command        *exec.Cmd
	done           chan struct{}
	grouped        bool
	stopAt         time.Time
	restartTimer   *time.Timer
}
