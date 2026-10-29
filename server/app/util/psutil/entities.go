package psutil

import (
	"sync"
	"time"
)

// HostInfo 主机与操作系统信息。时间戳、运行时间的单位均为秒。
type HostInfo struct {
	Hostname        string `json:"hostname"`         // 主机名
	OS              string `json:"os"`               // 操作系统标识
	Platform        string `json:"platform"`         // 发行版或系统名称
	PlatformVersion string `json:"platform_version"` // 发行版或系统版本
	KernelVersion   string `json:"kernel_version"`   // 内核版本
	KernelArch      string `json:"kernel_arch"`      // 内核架构
	Uptime          uint64 `json:"uptime"`           // 已运行秒数
	BootTime        uint64 `json:"boot_time"`        // 启动 Unix 秒时间戳
}

// CPUInfo CPU 型号、逻辑核心数和累计时间。Usage 为两次采样之间的整体百分比（0–100）。
// 单独调用 CPU 时只有累计时间，Available 在 Sampler 建立有效差分后才为 true。
type CPUInfo struct {
	Model     string  `json:"model"`      // CPU 型号
	Cores     int     `json:"cores"`      // 逻辑 CPU 核心数
	Processor int     `json:"processor"`  // Go 运行时识别的处理器数量
	Available bool    `json:"available"`  // 当前指标是否有效
	Usage     float64 `json:"usage"`      // 整体 CPU 使用率（百分比）
	TotalTime float64 `json:"total_time"` // 累计 CPU 秒，所有核心合计
	BusyTime  float64 `json:"busy_time"`  // 不包含 idle、iowait 的累计 CPU 秒
}

// MemoryInfo 物理内存，容量单位为字节。Available 表示采集是否成功。
type MemoryInfo struct {
	Available      bool    `json:"available"`       // 当前指标是否有效
	Total          uint64  `json:"total"`           // 物理内存总字节数
	Used           uint64  `json:"used"`            // 已使用字节数
	Free           uint64  `json:"free"`            // 空闲字节数
	AvailableBytes uint64  `json:"available_bytes"` // 系统估算的可用容量，不等于 Free
	UsedPercent    float64 `json:"used_percent"`    // 内存使用率（百分比）
}

// LoadInfo 1/5/15 分钟平均负载；Windows 由 Sampler 估算并标记 Estimated。
type LoadInfo struct {
	Available bool    `json:"available"` // 当前指标是否有效
	Load1     float64 `json:"load1"`     // 1 分钟平均负载
	Load5     float64 `json:"load5"`     // 5 分钟平均负载
	Load15    float64 `json:"load15"`    // 15 分钟平均负载
	Estimated bool    `json:"estimated"` // 是否为估算负载
}

// DiskIO 磁盘累计计数和差分速率。容量单位为字节，时间为毫秒，速率为每秒。
// IOLatency 是自启动以来的累计平均延迟，不是当前采样窗口延迟。
type DiskIO struct {
	Name           string  `json:"name"`             // 设备名称
	SerialNumber   string  `json:"serial_number"`    // 设备序列号（系统未提供时为空）
	ReadCount      uint64  `json:"read_count"`       // 累计读取次数
	WriteCount     uint64  `json:"write_count"`      // 累计写入次数
	ReadBytes      uint64  `json:"read_bytes"`       // 累计读取字节数
	WriteBytes     uint64  `json:"write_bytes"`      // 累计写入字节数
	ReadTime       uint64  `json:"read_time"`        // 累计读取毫秒
	WriteTime      uint64  `json:"write_time"`       // 累计写入毫秒
	IOTime         uint64  `json:"io_time"`          // 累计 I/O 忙碌毫秒
	WeightedIO     uint64  `json:"weighted_io"`      // 累计加权 I/O 毫秒
	ReadBytesRate  float64 `json:"read_bytes_rate"`  // 读取字节/秒
	WriteBytesRate float64 `json:"write_bytes_rate"` // 写入字节/秒
	ReadCountRate  float64 `json:"read_count_rate"`  // 读取次数/秒
	WriteCountRate float64 `json:"write_count_rate"` // 写入次数/秒
	IOLatency      float64 `json:"io_latency"`       // 累计平均 I/O 延迟（毫秒）
	RateAvailable  bool    `json:"rate_available"`   // 差分基线是否有效
}

// NetworkIO 网卡累计流量和差分速率。包含环回接口；容量为字节，速率为字节/秒。
type NetworkIO struct {
	Name           string  `json:"name"`            // 设备名称
	BytesSent      uint64  `json:"bytes_sent"`      // 累计发送字节数
	BytesRecv      uint64  `json:"bytes_recv"`      // 累计接收字节数
	PacketsSent    uint64  `json:"packets_sent"`    // 累计发送包数
	PacketsRecv    uint64  `json:"packets_recv"`    // 累计接收包数
	Errin          uint64  `json:"errin"`           // 累计接收错误数
	Errout         uint64  `json:"errout"`          // 累计发送错误数
	Dropin         uint64  `json:"dropin"`          // 累计接收丢包数
	Dropout        uint64  `json:"dropout"`         // 累计发送丢包数
	StateAvailable bool    `json:"state_available"` // 网卡启用与环回标志是否可读取
	IsUp           bool    `json:"is_up"`           // 网卡是否启用
	IsLoopback     bool    `json:"is_loopback"`     // 是否为环回接口
	SendRate       float64 `json:"send_rate"`       // 发送字节/秒
	RecvRate       float64 `json:"recv_rate"`       // 接收字节/秒
	RateAvailable  bool    `json:"rate_available"`  // 差分基线是否有效
}

// Snapshot 一次同步采样的固定结果。Errors 按 cpu/memory/load/network/disk 返回错误，
// 部分失败不丢弃其他指标；首次采样或计数器重置时 RateAvailable 为 false。
type Snapshot struct {
	Time    int64             `json:"time"`    // Unix 毫秒
	CPU     CPUInfo           `json:"cpu"`     // CPU 采样
	Memory  MemoryInfo        `json:"memory"`  // 物理内存采样
	Load    LoadInfo          `json:"load"`    // 系统负载采样
	Network []NetworkIO       `json:"network"` // 网卡采样（名称顺序）
	Disk    []DiskIO          `json:"disk"`    // 磁盘 I/O 采样（名称顺序）
	Errors  map[string]string `json:"errors"`  // 各采集项的错误，成功项不出现在此映射中
}

// DiskSpaceInfo 容量或 inode 用量；容量单位为字节，inode 单位为个。
type DiskSpaceInfo struct {
	Total      int64 `json:"total"`      // 总量
	Used       int64 `json:"used"`       // 已用量
	UnUsed     int64 `json:"un_used"`    // 空闲量
	Percentage int   `json:"percentage"` // 已用百分比，取整数
}

// DiskInfo 可浏览挂载卷的信息，不等同于物理硬盘数量。
type DiskInfo struct {
	Filesystem string        `json:"filesystem"` // 设备或文件系统名称
	Type       string        `json:"type"`       // 文件系统类型
	Path       string        `json:"path"`       // 挂载路径，Windows 使用 C:// 格式
	Size       DiskSpaceInfo `json:"size"`       // 存储容量
	Inodes     DiskSpaceInfo `json:"inodes"`     // inode 用量，不支持时为零
}

// DiskUsageInfo 路径所在文件系统的容量，不递归统计目录大小。
type DiskUsageInfo struct {
	Path   string        `json:"path"`   // 查询路径
	Size   DiskSpaceInfo `json:"size"`   // 存储容量
	Inodes DiskSpaceInfo `json:"inodes"` // inode 用量，不支持时为零
}

// ProcessSummary 进程列表摘要，不读取命令行、环境变量或文件列表。
type ProcessSummary struct {
	PID        int32             `json:"pid"`         // 进程 ID
	ParentPID  int32             `json:"parent_pid"`  // 父进程 ID
	Name       string            `json:"name"`        // 进程名称
	CreateTime int64             `json:"create_time"` // 创建时间，Unix 毫秒
	Errors     map[string]string `json:"errors"`      // 无法读取的字段及原因
}

// ProcessInfo 进程详情；资源变化通过同一个 Process 的 Sample 获取。
type ProcessInfo struct {
	ProcessSummary
	Status     []string `json:"status"`     // 系统提供的状态列表
	Executable string   `json:"executable"` // 可执行文件路径
	Command    []string `json:"command"`    // 启动参数，保持参数边界
	Directory  string   `json:"directory"`  // 工作目录
	Username   string   `json:"username"`   // 所属用户
}

// ProcessUsage 一次进程资源采样；CPU 百分比以单核 100% 计，多线程可超过 100%。
type ProcessUsage struct {
	PID              int32             `json:"pid"`                // 进程 ID
	CreateTime       int64             `json:"create_time"`        // 创建时间，Unix 毫秒
	SampleTime       int64             `json:"sample_time"`        // 采样时间，Unix 毫秒
	CPUAvailable     bool              `json:"cpu_available"`      // CPU 累计时间是否可用
	CPURateAvailable bool              `json:"cpu_rate_available"` // 是否已有两次有效 CPU 采样
	CPUPercent       float64           `json:"cpu_percent"`        // 区间 CPU 使用率
	CPUUser          float64           `json:"cpu_user"`           // 累计用户态 CPU 秒数
	CPUSystem        float64           `json:"cpu_system"`         // 累计内核态 CPU 秒数
	MemoryAvailable  bool              `json:"memory_available"`   // 内存信息是否可用
	MemoryRSS        uint64            `json:"memory_rss"`         // 驻留内存，字节
	MemoryVMS        uint64            `json:"memory_vms"`         // 虚拟内存，字节
	MemorySwap       uint64            `json:"memory_swap"`        // 交换内存，字节；部分系统不提供
	Threads          int32             `json:"threads"`            // 线程数
	IOAvailable      bool              `json:"io_available"`       // I/O 计数器是否可用
	IORateAvailable  bool              `json:"io_rate_available"`  // 是否已有两次有效 I/O 采样
	ReadBytes        uint64            `json:"read_bytes"`         // 累计读取字节；macOS 使用磁盘计数
	WriteBytes       uint64            `json:"write_bytes"`        // 累计写入字节；macOS 使用磁盘计数
	ReadCount        uint64            `json:"read_count"`         // 累计读取次数，平台支持时提供
	WriteCount       uint64            `json:"write_count"`        // 累计写入次数，平台支持时提供
	ReadRate         float64           `json:"read_rate"`          // 区间读取字节/秒
	WriteRate        float64           `json:"write_rate"`         // 区间写入字节/秒
	Errors           map[string]string `json:"errors"`             // 不可用指标及原因
}

// Process 绑定 PID 和创建时间的进程句柄。没有后台协程，也不需要 Close。
// 通过 NewProcess 创建；Sample 内部串行维护差分基线，不要复制此对象。
type Process struct {
	pid        int32
	createTime int64
	mu         sync.Mutex
	cpuTime    float64
	cpuAt      time.Time
	readBytes  uint64
	writeBytes uint64
	ioAt       time.Time
}

// ConnectionFilter 网络连接筛选；零值表示查询所有可见的 TCP/UDP 连接。
type ConnectionFilter struct {
	Kind       string `json:"kind"`        // 协议：all、tcp、tcp4、tcp6、udp、udp4、udp6，空值等同 all
	PID        int32  `json:"pid"`         // 进程 ID，0 表示全部进程
	LocalPort  uint32 `json:"local_port"`  // 本地端口，0 表示不限
	RemotePort uint32 `json:"remote_port"` // 远端端口，0 表示不限
	Status     string `json:"status"`      // 连接状态，不区分大小写，空值表示不限
}

// AddressInfo 网络端点；未连接的 UDP 或监听套接字没有远端地址。
type AddressInfo struct {
	IP   string `json:"ip"`   // IP 地址，通配监听地址在部分系统上表示为 *
	Port uint32 `json:"port"` // 端口，0 表示没有端口
}

// ConnectionInfo TCP/UDP 套接字信息，不包含 Unix 域套接字。
type ConnectionInfo struct {
	PID      int32       `json:"pid"`      // 所属进程 ID；0 表示系统未提供或当前权限无法读取
	FD       uint32      `json:"fd"`       // 文件描述符；Windows 不提供此值
	Protocol string      `json:"protocol"` // tcp 或 udp
	Family   string      `json:"family"`   // ipv4 或 ipv6
	Local    AddressInfo `json:"local"`    // 本地端点
	Remote   AddressInfo `json:"remote"`   // 远端端点
	Status   string      `json:"status"`   // TCP 状态；UDP 统一为 BOUND 或 CONNECTED
}

// InterfaceInfo 网卡基本信息，地址使用 CIDR 表示，不包含流量采样。
type InterfaceInfo struct {
	Index      int      `json:"index"`       // 系统接口索引
	Name       string   `json:"name"`        // 接口名称
	MTU        int      `json:"mtu"`         // 最大传输单元，字节
	MAC        string   `json:"mac"`         // 硬件地址，虚拟接口可能为空
	Flags      []string `json:"flags"`       // up、broadcast、loopback、pointtopoint、multicast、running
	Addresses  []string `json:"addresses"`   // IPv4/IPv6 地址及前缀长度
	IsUp       bool     `json:"is_up"`       // 接口是否启用，不等于互联网连接可用
	IsLoopback bool     `json:"is_loopback"` // 是否为环回接口
	Error      string   `json:"error"`       // 此网卡的地址读取错误；成功时为空
}
