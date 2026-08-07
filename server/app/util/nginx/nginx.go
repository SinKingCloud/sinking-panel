package nginx

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultTimeout        = 30 * time.Second
	installTimeout        = 30 * time.Minute
	maxInstallArchiveSize = 512 << 20
	maxConfigFileSize     = 32 << 20
	maxLogReadSize        = 8 << 20
	managedServiceMarker  = "Managed-By: sinking-panel"
)

// Nginx 封装 Nginx 的检测、安装、配置和生命周期管理。
type Nginx struct {
	mu             sync.RWMutex
	operationMu    sync.Mutex
	installMu      sync.Mutex
	binaryPath     string
	configPath     string
	prefixPath     string
	pidPath        string
	configExplicit bool
	prefixExplicit bool
	serviceName    string
	instanceName   string
	runUser        string
	backupDir      string
	globalArgs     []string
	timeout        time.Duration
	readyTimeout   time.Duration
}

// Info Nginx 安装信息。
type Info struct {
	BinaryPath   string `json:"binary_path"`
	ConfigPath   string `json:"config_path"`
	PrefixPath   string `json:"prefix_path"`
	PidPath      string `json:"pid_path"`
	ServiceName  string `json:"service_name,omitempty"`
	InstanceName string `json:"instance_name,omitempty"`
	Version      string `json:"version"`
	Build        string `json:"build"`
}

// Status Nginx 运行状态。
type Status struct {
	Installed bool   `json:"installed"`
	Running   bool   `json:"running"`
	Pid       int    `json:"pid"`
	Version   string `json:"version"`
	Binary    string `json:"binary"`
	Config    string `json:"config"`
}

// Environment 当前服务器环境能力。
type Environment struct {
	OS                  string `json:"os"`
	Arch                string `json:"arch"`
	User                string `json:"user"`
	Root                bool   `json:"root"`
	Sudo                bool   `json:"sudo"`
	Doas                bool   `json:"doas"`
	CanElevate          bool   `json:"can_elevate"`
	Distribution        string `json:"distribution,omitempty"`
	DistributionVersion string `json:"distribution_version,omitempty"`
	DistributionLike    string `json:"distribution_like,omitempty"`
	Kernel              string `json:"kernel,omitempty"`
	Container           string `json:"container,omitempty"`
	Libc                string `json:"libc,omitempty"`
	PackageManager      string `json:"package_manager"`
	InitSystem          string `json:"init_system"`
	ArchiveInstall      bool   `json:"archive_install"`
	SourceBuild         bool   `json:"source_build"`
}

// OSRelease 是 Linux /etc/os-release 的结构化信息。
// Values 保留发行版提供的其他字段，便于后续新增发行版适配时无需改结构体。
type OSRelease struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	VersionID string            `json:"version_id"`
	IDLike    string            `json:"id_like"`
	Pretty    string            `json:"pretty_name"`
	Values    map[string]string `json:"values,omitempty"`
}

// ConfigFile 配置文件信息。
type ConfigFile struct {
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
}

// ConfigNode 是 Nginx 配置语法树中的一个指令或配置块。
// Children 非空时表示该节点后面跟随了大括号；空 Children 也可能表示空配置块。
type ConfigNode struct {
	Name     string       `json:"name"`
	Args     []string     `json:"args,omitempty"`
	Children []ConfigNode `json:"children,omitempty"`
	Block    bool         `json:"block,omitempty"`
	Source   string       `json:"source,omitempty"`
	Line     int          `json:"line,omitempty"`
}

// ConfigDocument 是解析后的 Nginx 配置文档。
type ConfigDocument struct {
	Path  string       `json:"path,omitempty"`
	Files []string     `json:"files,omitempty"`
	Nodes []ConfigNode `json:"nodes"`
}

type nginxConfigToken struct {
	Value string
	Line  int
}

// NginxModule 是 Nginx 编译模块或动态加载模块的信息。
type NginxModule struct {
	Name      string `json:"name"`
	Directive string `json:"directive,omitempty"`
	Kind      string `json:"kind"` // builtin、dynamic、disabled
	Enabled   bool   `json:"enabled"`
	Source    string `json:"source,omitempty"`
	Path      string `json:"path,omitempty"`
}

// NginxModuleReport 是当前 Nginx 的模块检测结果。
type NginxModuleReport struct {
	BinaryPath  string        `json:"binary_path"`
	ModulesPath string        `json:"modules_path,omitempty"`
	Modules     []NginxModule `json:"modules"`
}

// LogFormatDefinition 是 log_format 指令的结构化表示。
type LogFormatDefinition struct {
	Name      string   `json:"name"`
	Format    string   `json:"format"`
	Escape    string   `json:"escape,omitempty"`
	Variables []string `json:"variables,omitempty"`
	Source    string   `json:"source,omitempty"`
	Line      int      `json:"line,omitempty"`
}

// LogDirectiveDefinition 是 access_log 或 error_log 指令的结构化表示。
type LogDirectiveDefinition struct {
	Kind    string   `json:"kind"`
	Path    string   `json:"path"`
	Format  string   `json:"format,omitempty"`
	Options []string `json:"options,omitempty"`
	Level   string   `json:"level,omitempty"`
	Enabled bool     `json:"enabled"`
	Syslog  bool     `json:"syslog"`
	Scope   []string `json:"scope,omitempty"`
	Source  string   `json:"source,omitempty"`
	Line    int      `json:"line,omitempty"`
}

// LogConfigDefinition 是完整配置中的日志格式和日志输出定义。
type LogConfigDefinition struct {
	Formats    []LogFormatDefinition    `json:"formats"`
	Directives []LogDirectiveDefinition `json:"directives"`
}

// SiteInfo 站点配置摘要。
type SiteInfo struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Enabled    bool      `json:"enabled"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
}

// LogInfo Nginx 日志文件信息。
type LogInfo struct {
	Kind       string    `json:"kind"`
	Path       string    `json:"path"`
	Enabled    bool      `json:"enabled"`
	Format     string    `json:"format,omitempty"`
	Level      string    `json:"level,omitempty"`
	Options    []string  `json:"options,omitempty"`
	Syslog     bool      `json:"syslog,omitempty"`
	Exists     bool      `json:"exists"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
}

// CertificateInfo TLS 证书信息。
type CertificateInfo struct {
	Path      string    `json:"path"`
	Subject   string    `json:"subject"`
	Issuer    string    `json:"issuer"`
	DNSNames  []string  `json:"dns_names"`
	NotBefore time.Time `json:"not_before"`
	NotAfter  time.Time `json:"not_after"`
	Expired   bool      `json:"expired"`
}

// UpstreamServer upstream 后端节点。
type UpstreamServer struct {
	Address     string `json:"address"`
	Weight      int    `json:"weight,omitempty"`
	MaxFails    int    `json:"max_fails,omitempty"`
	FailTimeout string `json:"fail_timeout,omitempty"`
	Backup      bool   `json:"backup,omitempty"`
	Down        bool   `json:"down,omitempty"`
}

// ServiceOptions 定义自定义系统服务。
type ServiceOptions struct {
	Name             string
	Description      string
	User             string
	Group            string
	WorkingDirectory string
	Restart          string
	Enabled          bool
}

// ServiceInfo 系统服务状态。
type ServiceInfo struct {
	Name     string `json:"name"`
	Manager  string `json:"manager"`
	UnitPath string `json:"unit_path"`
	Active   bool   `json:"active"`
	Enabled  bool   `json:"enabled"`
}

// ProcessInfo 当前实例的进程身份信息。
type ProcessInfo struct {
	PID         int    `json:"pid"`
	Running     bool   `json:"running"`
	Executable  string `json:"executable"`
	CommandLine string `json:"command_line"`
	User        string `json:"user"`
}

// InstanceInfo 是发现到的 Nginx 实例信息。
type InstanceInfo struct {
	BinaryPath  string `json:"binary_path"`
	ConfigPath  string `json:"config_path"`
	PrefixPath  string `json:"prefix_path"`
	PIDPath     string `json:"pid_path"`
	PID         int    `json:"pid"`
	Running     bool   `json:"running"`
	CommandLine string `json:"command_line"`
}

// ConfigChange 是批量配置事务中的一个文件变更。
type ConfigChange struct {
	Path    string `json:"path"`
	Content string `json:"content,omitempty"`
	Remove  bool   `json:"remove,omitempty"`
}

// ConfigBackup 是配置版本备份摘要。
type ConfigBackup struct {
	ID        string             `json:"id"`
	CreatedAt time.Time          `json:"created_at"`
	Files     []ConfigBackupFile `json:"files"`
}

// ConfigBackupFile 是备份中的单个文件。
type ConfigBackupFile struct {
	Path       string `json:"path"`
	BackupPath string `json:"backup_path"`
	Size       int64  `json:"size"`
	SHA256     string `json:"sha256"`
	Mode       uint32 `json:"mode"`
}

// SiteLayout 描述站点目录布局。
type SiteLayout struct {
	Mode         string `json:"mode"`
	AvailableDir string `json:"available_dir"`
	EnabledDir   string `json:"enabled_dir"`
	ActiveDir    string `json:"active_dir"`
}

// ProxySubFilter 描述响应正文的内容替换规则。需要 Nginx 编译 ngx_http_sub_module。
type ProxySubFilter struct {
	Search  string   `json:"search"`
	Replace string   `json:"replace"`
	Types   []string `json:"types,omitempty"`
}

// SiteDefinition 是可由面板编辑的结构化站点配置。
type SiteDefinition struct {
	Name                string            `json:"name"`
	Type                string            `json:"type"`
	ServerName          string            `json:"server_name"`
	Listen              int               `json:"listen"`
	Root                string            `json:"root"`
	ProxyPass           string            `json:"proxy_pass"`
	FastCGIPass         string            `json:"fastcgi_pass"`
	TLS                 bool              `json:"tls"`
	CertificatePath     string            `json:"certificate_path"`
	KeyPath             string            `json:"key_path"`
	RedirectHTTPS       bool              `json:"redirect_https"`
	RedirectPort        int               `json:"redirect_port"`
	AccessLog           string            `json:"access_log"`
	ErrorLog            string            `json:"error_log"`
	AccessLogEnabled    bool              `json:"access_log_enabled"`
	ErrorLogEnabled     bool              `json:"error_log_enabled"`
	ClientMaxBodySize   string            `json:"client_max_body_size"`
	ProxyReadTimeout    string            `json:"proxy_read_timeout"`
	ProxyConnectTimeout string            `json:"proxy_connect_timeout"`
	ProxySendTimeout    string            `json:"proxy_send_timeout"`
	Gzip                bool              `json:"gzip"`
	HTTP2               bool              `json:"http2"`
	Headers             map[string]string `json:"headers"`
	ProxyHost           string            `json:"proxy_host"`
	ProxyHeaders        map[string]string `json:"proxy_headers"`
	ProxySetBody        string            `json:"proxy_set_body"`
	ProxyRedirect       string            `json:"proxy_redirect"`
	ProxyCookieDomain   string            `json:"proxy_cookie_domain"`
	ProxyCookiePath     string            `json:"proxy_cookie_path"`
	SubFilters          []ProxySubFilter  `json:"sub_filters"`
	SubFilterOnce       *bool             `json:"sub_filter_once,omitempty"`
	ExtraDirectives     []string          `json:"extra_directives"`
}

// LogPage 是按字节游标读取日志的结果。
type LogPage struct {
	Kind       string `json:"kind"`
	Path       string `json:"path"`
	Offset     int64  `json:"offset"`
	NextOffset int64  `json:"next_offset"`
	Size       int64  `json:"size"`
	Content    string `json:"content"`
	EOF        bool   `json:"eof"`
}

// LogCursorPage 是站点日志按行读取的游标结果。
// cursor 用于读取新增内容，before 用于向更早的历史日志翻页。
type LogCursorPage struct {
	Kind        string   `json:"kind"`
	Path        string   `json:"path"`
	Lines       []string `json:"lines"`
	Cursor      int64    `json:"cursor"`
	StartCursor int64    `json:"start_cursor"`
	HasPrevious bool     `json:"has_previous"`
	End         bool     `json:"end"`
	Size        int64    `json:"size"`
}

// CertificateFiles 是安装后的证书和私钥路径。
type CertificateFiles struct {
	Name     string `json:"name"`
	CertPath string `json:"cert_path"`
	KeyPath  string `json:"key_path"`
}

// PortCheck 是端口预检结果。
type PortCheck struct {
	Port      int    `json:"port"`
	Network   string `json:"network"`
	Address   string `json:"address"`
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}

// DependencyStatus 是命令或运行依赖的检测结果。
type DependencyStatus struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Available bool   `json:"available"`
}

// PreflightReport 是启动或部署前的综合预检结果。
type PreflightReport struct {
	Environment  *Environment       `json:"environment,omitempty"`
	Installed    bool               `json:"installed"`
	ConfigValid  bool               `json:"config_valid"`
	Running      bool               `json:"running"`
	Ports        []PortCheck        `json:"ports"`
	Dependencies []DependencyStatus `json:"dependencies"`
	Errors       []string           `json:"errors"`
}

// StreamProxyOptions 描述 TCP/UDP stream 代理。
type StreamProxyOptions struct {
	Name    string `json:"name"`
	Listen  int    `json:"listen"`
	Target  string `json:"target"`
	UDP     bool   `json:"udp"`
	Timeout string `json:"timeout"`
}

// AdvancedProxyOptions 是反向代理常用缓存、限流和访问控制模板。
type AdvancedProxyOptions struct {
	Site           SiteDefinition `json:"site"`
	CacheZone      string         `json:"cache_zone"`
	CacheValid     string         `json:"cache_valid"`
	RateLimitZone  string         `json:"rate_limit_zone"`
	RateLimitBurst int            `json:"rate_limit_burst"`
	Allow          []string       `json:"allow"`
	Deny           []string       `json:"deny"`
}

// RuntimeStats 是 Nginx 进程级运行指标。
type RuntimeStats struct {
	SampledAt   time.Time `json:"sampled_at"`
	MasterPID   int       `json:"master_pid"`
	WorkerPIDs  []int     `json:"worker_pids"`
	Running     bool      `json:"running"`
	MemoryBytes uint64    `json:"memory_bytes"`
	CPUSeconds  float64   `json:"cpu_seconds"`
	Threads     int       `json:"threads"`
	OpenFiles   int       `json:"open_files"`
}

// StubStatus 是 ngx_http_stub_status_module 的指标。
type StubStatus struct {
	Active   int64 `json:"active"`
	Accepted int64 `json:"accepted"`
	Handled  int64 `json:"handled"`
	Requests int64 `json:"requests"`
	Reading  int64 `json:"reading"`
	Writing  int64 `json:"writing"`
	Waiting  int64 `json:"waiting"`
}

// New 创建 Nginx 管理器。
// binaryPath 和 configPath 可以为空，为空时自动从系统路径和 Nginx 编译参数中查找。
func New(binaryPath, configPath string) *Nginx {
	return &Nginx{
		binaryPath:     strings.TrimSpace(binaryPath),
		configPath:     strings.TrimSpace(configPath),
		configExplicit: strings.TrimSpace(configPath) != "",
		serviceName:    "nginx",
		timeout:        defaultTimeout,
		readyTimeout:   10 * time.Second,
	}
}

// SetTimeout 设置单次 Nginx 命令超时时间。
func (n *Nginx) SetTimeout(timeout time.Duration) *Nginx {
	if timeout > 0 {
		n.mu.Lock()
		n.timeout = timeout
		n.mu.Unlock()
	}
	return n
}

// SetReadyTimeout 设置启动后等待实例进入运行状态的最长时间。
func (n *Nginx) SetReadyTimeout(timeout time.Duration) *Nginx {
	if timeout > 0 {
		n.mu.Lock()
		n.readyTimeout = timeout
		n.mu.Unlock()
	}
	return n
}

// SetServiceName 设置系统服务名称，默认是 nginx。
func (n *Nginx) SetServiceName(name string) *Nginx {
	name = strings.TrimSpace(name)
	if name != "" {
		if _, err := n.normalizeIdentifier(name); err != nil {
			return n
		}
		n.mu.Lock()
		n.serviceName = name
		n.mu.Unlock()
	}
	return n
}

// SetInstanceName 设置面板中的实例名称。
func (n *Nginx) SetInstanceName(name string) *Nginx {
	n.mu.Lock()
	n.instanceName = strings.TrimSpace(name)
	n.mu.Unlock()
	return n
}

// SetPrefixPath 显式设置 Nginx prefix，用于多实例和自定义目录安装。
func (n *Nginx) SetPrefixPath(path string) *Nginx {
	n.mu.Lock()
	n.prefixPath = strings.TrimSpace(path)
	n.prefixExplicit = n.prefixPath != ""
	n.mu.Unlock()
	return n
}

// SetPIDPath 显式设置 PID 文件路径，用于多实例精确识别。
func (n *Nginx) SetPIDPath(path string) *Nginx {
	n.mu.Lock()
	n.pidPath = strings.TrimSpace(path)
	n.mu.Unlock()
	return n
}

// SetGlobalArgs 设置实例启动时附加的 nginx -g 全局参数，例如 pid=/run/nginx-a.pid;。
func (n *Nginx) SetGlobalArgs(args ...string) *Nginx {
	n.mu.Lock()
	n.globalArgs = append([]string(nil), args...)
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if strings.HasPrefix(arg, "pid=") {
			n.pidPath = strings.TrimSuffix(strings.TrimPrefix(arg, "pid="), ";")
		}
	}
	n.mu.Unlock()
	return n
}

// SetBackupDir 设置配置版本备份目录。为空时使用主配置目录下的 .nginx-backups。
func (n *Nginx) SetBackupDir(path string) *Nginx {
	n.mu.Lock()
	n.backupDir = strings.TrimSpace(path)
	n.mu.Unlock()
	return n
}

// SetRunUser 设置服务默认运行用户，InstallService 未传 User 时使用。
func (n *Nginx) SetRunUser(name string) *Nginx {
	n.mu.Lock()
	n.runUser = strings.TrimSpace(name)
	n.mu.Unlock()
	return n
}

// Environment 获取当前服务器的系统、发行版、权限、包管理器和初始化系统信息。
func (n *Nginx) Environment() *Environment {
	currentUser, userErr := user.Current()
	userName := ""
	isRoot := n.isRootUser()
	if userErr == nil {
		userName = currentUser.Username
	}
	release := n.ReadOSRelease()
	canElevate := isRoot || n.commandExists("sudo") || n.commandExists("doas")
	return &Environment{
		OS:                  runtime.GOOS,
		Arch:                runtime.GOARCH,
		User:                userName,
		Root:                isRoot,
		Sudo:                n.commandExists("sudo"),
		Doas:                n.commandExists("doas"),
		CanElevate:          canElevate,
		Distribution:        release.ID,
		DistributionVersion: release.VersionID,
		DistributionLike:    release.IDLike,
		Kernel:              n.kernelRelease(),
		Container:           n.detectContainerRuntime(),
		Libc:                n.detectLibc(),
		PackageManager:      n.detectPackageManager(),
		InitSystem:          n.detectInitSystem(),
		ArchiveInstall:      runtime.GOOS == "linux" || runtime.GOOS == "windows",
		SourceBuild:         runtime.GOOS == "linux" && (n.commandExists("make") || n.commandExists("gmake")) && (n.commandExists("cc") || n.commandExists("gcc") || n.commandExists("clang")),
	}
}

// ReadOSRelease 读取 Linux 发行版信息。无法读取时返回空结构，不阻断归档安装。
func (n *Nginx) ReadOSRelease() *OSRelease {
	release := &OSRelease{Values: make(map[string]string)}
	if runtime.GOOS != "linux" {
		return release
	}
	data := []byte(nil)
	for _, path := range []string{"/etc/os-release", "/usr/lib/os-release"} {
		if value, err := os.ReadFile(path); err == nil {
			data = value
			break
		}
	}
	if len(data) == 0 {
		return release
	}
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.ToUpper(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		if len(value) >= 2 && ((value[0] == '\'' && value[len(value)-1] == '\'') || (value[0] == '"' && value[len(value)-1] == '"')) {
			value = value[1 : len(value)-1]
		}
		release.Values[key] = value
	}
	release.ID = strings.ToLower(release.Values["ID"])
	release.Name = release.Values["NAME"]
	release.VersionID = release.Values["VERSION_ID"]
	release.IDLike = strings.ToLower(strings.Join(strings.Fields(release.Values["ID_LIKE"]), " "))
	release.Pretty = release.Values["PRETTY_NAME"]
	return release
}

func (n *Nginx) kernelRelease() string {
	if runtime.GOOS != "linux" {
		return ""
	}
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func (n *Nginx) detectContainerRuntime() string {
	if runtime.GOOS != "linux" {
		return ""
	}
	if value := strings.TrimSpace(os.Getenv("container")); value != "" {
		return value
	}
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return "docker"
	}
	if _, err := os.Stat("/run/.containerenv"); err == nil {
		return "podman"
	}
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		value := strings.ToLower(string(data))
		for _, marker := range []string{"kubepods", "docker", "containerd", "libpod", "lxc"} {
			if strings.Contains(value, marker) {
				return marker
			}
		}
	}
	return ""
}

func (n *Nginx) detectLibc() string {
	if runtime.GOOS != "linux" {
		return ""
	}
	for _, pattern := range []string{"/lib*/ld-musl-*.so.1", "/usr/lib*/ld-musl-*.so.1"} {
		if matches, err := filepath.Glob(pattern); err == nil && len(matches) > 0 {
			return "musl"
		}
	}
	for _, pattern := range []string{"/lib*/ld-linux-*.so.*", "/usr/lib*/ld-linux-*.so.*"} {
		if matches, err := filepath.Glob(pattern); err == nil && len(matches) > 0 {
			return "glibc"
		}
	}
	return "unknown"
}

// Info 获取 Nginx 安装和编译信息。
func (n *Nginx) Info() (*Info, error) {
	binaryPath, err := n.binary()
	if err != nil {
		return nil, err
	}

	versionOutput, versionErr := n.runCommand(context.Background(), n.commandTimeout(), binaryPath, "-v")
	buildOutput, buildErr := n.runCommand(context.Background(), n.commandTimeout(), binaryPath, "-V")
	if versionErr != nil && buildErr != nil && versionOutput == "" && buildOutput == "" {
		return nil, fmt.Errorf("读取 nginx 信息失败: %w", versionErr)
	}

	prefix := n.parseBuildOption(buildOutput, "prefix")
	config := n.parseBuildOption(buildOutput, "conf-path")
	pid := n.parseBuildOption(buildOutput, "pid-path")
	if config != "" {
		config = n.resolveBuildPath(config, prefix)
	}
	if pid != "" {
		pid = n.resolveBuildPath(pid, prefix)
	}

	n.mu.Lock()
	if n.prefixPath == "" {
		n.prefixPath = prefix
	}
	if n.configPath == "" {
		n.configPath = config
	}
	if n.pidPath == "" {
		n.pidPath = pid
	}
	info := &Info{
		BinaryPath:   n.binaryPath,
		ConfigPath:   n.configPath,
		PrefixPath:   n.prefixPath,
		PidPath:      n.pidPath,
		ServiceName:  n.serviceName,
		InstanceName: n.instanceName,
		Version:      n.parseVersion(versionOutput + "\n" + buildOutput),
		Build:        strings.TrimSpace(buildOutput),
	}
	n.mu.Unlock()
	return info, nil
}

// Status 获取 Nginx 当前运行状态。
func (n *Nginx) Status() (*Status, error) {
	info, err := n.Info()
	if err != nil {
		return &Status{Installed: false}, err
	}
	return n.statusForInfo(info)
}

func (n *Nginx) statusForInfo(info *Info) (*Status, error) {
	if info == nil {
		return &Status{Installed: false}, errors.New("nginx 信息为空")
	}
	status := &Status{
		Installed: true,
		Version:   info.Version,
		Binary:    info.BinaryPath,
		Config:    info.ConfigPath,
	}
	pidPath := info.PidPath
	if pidPath == "" {
		pidPath = n.defaultPidPath(info)
	}
	pid, err := n.readPid(pidPath)
	if err != nil {
		if os.IsNotExist(err) {
			return status, nil
		}
		return status, fmt.Errorf("读取 nginx PID 失败: %w", err)
	}
	status.Pid = pid
	process, processErr := n.inspectProcess(pid)
	if processErr == nil {
		status.Running = process.Running && n.processMatches(info, process)
		return status, nil
	}
	status.Running, err = n.processRunning(pid)
	return status, err
}

// Process 返回当前实例对应的进程身份。PID 存在但不是当前 Nginx 实例时会返回 Running=false。
func (n *Nginx) Process() (*ProcessInfo, error) {
	info, err := n.Info()
	if err != nil {
		return nil, err
	}
	pidPath := info.PidPath
	if pidPath == "" {
		pidPath = n.defaultPidPath(info)
	}
	pid, err := n.readPid(pidPath)
	if err != nil {
		return &ProcessInfo{}, err
	}
	process, err := n.inspectProcess(pid)
	if err != nil {
		return nil, err
	}
	process.Running = process.Running && n.processMatches(info, process)
	return process, nil
}

// DiscoverInstances 发现当前机器上的 Nginx master 实例，用于多实例面板管理。
func (n *Nginx) DiscoverInstances() ([]InstanceInfo, error) {
	if runtime.GOOS != "linux" {
		info, err := n.Info()
		if err != nil {
			return nil, err
		}
		status, statusErr := n.Status()
		if statusErr != nil {
			return nil, statusErr
		}
		return []InstanceInfo{{BinaryPath: info.BinaryPath, ConfigPath: info.ConfigPath, PrefixPath: info.PrefixPath, PIDPath: info.PidPath, PID: status.Pid, Running: status.Running}}, nil
	}
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	instances := make([]InstanceInfo, 0)
	seen := make(map[int]struct{})
	for _, entry := range entries {
		pid, parseErr := strconv.Atoi(entry.Name())
		if parseErr != nil {
			continue
		}
		process, processErr := n.inspectProcess(pid)
		if processErr != nil || !process.Running || strings.Contains(process.CommandLine, "nginx: worker process") {
			continue
		}
		if filepath.Base(process.Executable) != "nginx" && filepath.Base(process.Executable) != "nginx.exe" && !strings.Contains(process.CommandLine, "nginx: master process") {
			continue
		}
		if _, exists := seen[pid]; exists {
			continue
		}
		seen[pid] = struct{}{}
		prefix := n.commandLineValue(process.CommandLine, "-p")
		config := n.commandLineValue(process.CommandLine, "-c")
		manager := New(process.Executable, config)
		if prefix != "" {
			manager.SetPrefixPath(prefix)
		}
		info, infoErr := manager.Info()
		if infoErr != nil {
			continue
		}
		pidPath := info.PidPath
		if global := n.commandLineValue(process.CommandLine, "-g"); strings.HasPrefix(global, "pid=") {
			pidPath = strings.TrimSuffix(strings.TrimPrefix(global, "pid="), ";")
		}
		instances = append(instances, InstanceInfo{BinaryPath: info.BinaryPath, ConfigPath: info.ConfigPath, PrefixPath: info.PrefixPath, PIDPath: pidPath, PID: pid, Running: true, CommandLine: process.CommandLine})
	}
	sort.Slice(instances, func(i, j int) bool { return instances[i].PID < instances[j].PID })
	return instances, nil
}

// Test 检查 Nginx 配置。
func (n *Nginx) Test() error {
	info, err := n.Info()
	if err != nil {
		return err
	}
	args := n.runtimeArgs(info, "-t")
	output, err := n.runCommand(context.Background(), n.commandTimeout(), info.BinaryPath, args...)
	if err != nil {
		return fmt.Errorf("nginx 配置检查失败: %w", n.formatCommandOutput(output))
	}
	return nil
}

// ListModules 检测 Nginx 编译模块、动态模块文件和 load_module 配置。
// 对 Linux 发行版自带的静态模块，结果以 nginx -V 的 configure 参数为准；
// 动态模块会同时标记是否已经被当前配置加载。
func (n *Nginx) ListModules() (*NginxModuleReport, error) {
	info, err := n.Info()
	if err != nil {
		return nil, err
	}
	report := &NginxModuleReport{BinaryPath: info.BinaryPath, Modules: make([]NginxModule, 0)}
	flags := n.nginxConfigureFlags(info.Build)
	if modulesPath := flags["--modules-path"]; modulesPath != "" {
		candidate := n.resolveBuildPath(strings.Trim(modulesPath, `"'`), info.PrefixPath)
		if stat, statErr := os.Stat(candidate); statErr == nil && stat.IsDir() {
			report.ModulesPath = candidate
		}
	}
	if report.ModulesPath == "" {
		for _, candidate := range []string{
			filepath.Join(info.PrefixPath, "modules"),
			"/usr/lib/nginx/modules",
			"/usr/lib64/nginx/modules",
			"/usr/local/nginx/modules",
		} {
			if candidate != "" {
				if stat, statErr := os.Stat(candidate); statErr == nil && stat.IsDir() {
					report.ModulesPath = candidate
					break
				}
			}
		}
	}
	modules := make(map[string]NginxModule)
	for _, spec := range n.nginxModuleSpecs() {
		disabled := false
		disableSource := ""
		if strings.HasPrefix(spec.Name, "ngx_http_") && flags["--without-http"] == "true" {
			disabled = true
			disableSource = "--without-http"
		}
		if spec.Flag != "" {
			if value, exists := flags["--without-"+spec.Flag]; exists && value == "true" {
				disabled = true
				disableSource = "--without-" + spec.Flag
			}
		}
		if disabled {
			modules[spec.Name] = NginxModule{Name: spec.Name, Directive: spec.Directive, Kind: "disabled", Enabled: false, Source: disableSource}
			continue
		}
		if spec.Flag != "" {
			if value, exists := flags["--with-"+spec.Flag]; exists && value == "true" {
				modules[spec.Name] = NginxModule{Name: spec.Name, Directive: spec.Directive, Kind: "builtin", Enabled: true, Source: "--with-" + spec.Flag}
				continue
			}
		}
		if spec.Default {
			modules[spec.Name] = NginxModule{Name: spec.Name, Directive: spec.Directive, Kind: "builtin", Enabled: true, Source: "default"}
		}
	}
	for flag, value := range flags {
		switch {
		case strings.HasPrefix(flag, "--with-") && strings.HasSuffix(flag, "_module") && value == "true":
			name := n.canonicalModuleName(strings.TrimPrefix(flag, "--with-"))
			if name != "" {
				if _, exists := modules[name]; !exists {
					modules[name] = NginxModule{Name: name, Kind: "builtin", Enabled: true, Source: flag}
				}
			}
		case strings.HasPrefix(flag, "--without-") && strings.HasSuffix(flag, "_module") && value == "true":
			name := n.canonicalModuleName(strings.TrimPrefix(flag, "--without-"))
			if name != "" {
				modules[name] = NginxModule{Name: name, Kind: "disabled", Enabled: false, Source: flag}
			}
		}
	}
	for _, value := range n.nginxConfigureValues(info.Build, "--add-module") {
		name := n.canonicalModuleName(filepath.Base(value))
		if name != "" {
			modules[name] = NginxModule{Name: name, Kind: "builtin", Enabled: true, Source: value}
		}
	}
	for _, value := range n.nginxConfigureValues(info.Build, "--add-dynamic-module") {
		name := n.canonicalModuleName(filepath.Base(value))
		if name != "" {
			modules[name] = NginxModule{Name: name, Kind: "dynamic", Enabled: false, Source: value}
		}
	}
	if report.ModulesPath != "" {
		entries, readErr := os.ReadDir(report.ModulesPath)
		if readErr == nil {
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".so") {
					continue
				}
				name := n.canonicalModuleName(entry.Name())
				if name == "" {
					continue
				}
				item := modules[name]
				item.Name = name
				item.Kind = "dynamic"
				item.Path = filepath.Join(report.ModulesPath, entry.Name())
				modules[name] = item
			}
		}
	}
	if document, treeErr := n.ReadConfigTree(); treeErr == nil {
		n.walkConfigNodes(document.Nodes, func(node ConfigNode) {
			if node.Name != "load_module" || len(node.Args) == 0 {
				return
			}
			modulePath := n.resolveModulePath(node.Args[0], info)
			name := n.canonicalModuleName(filepath.Base(modulePath))
			if name == "" {
				return
			}
			item := modules[name]
			item.Name = name
			item.Kind = "dynamic"
			item.Enabled = true
			item.Path = modulePath
			item.Source = node.Source
			modules[name] = item
		})
	}
	for _, module := range modules {
		report.Modules = append(report.Modules, module)
	}
	sort.Slice(report.Modules, func(i, j int) bool { return report.Modules[i].Name < report.Modules[j].Name })
	return report, nil
}

// HasModule 判断当前 Nginx 是否具备指定模块。name 可传 ngx_http_ssl_module、http_ssl_module 或 http_ssl。
func (n *Nginx) HasModule(name string) (bool, error) {
	query := n.canonicalModuleName(name)
	if query == "" {
		return false, errors.New("模块名称不能为空")
	}
	report, err := n.ListModules()
	if err != nil {
		return false, err
	}
	for _, module := range report.Modules {
		if module.Name == query {
			return module.Enabled && module.Kind != "disabled", nil
		}
	}
	return false, nil
}

type nginxModuleSpec struct {
	Name      string
	Flag      string
	Directive string
	Default   bool
}

func (n *Nginx) nginxModuleSpecs() []nginxModuleSpec {
	return []nginxModuleSpec{
		{Name: "ngx_core_module", Flag: "", Directive: "core", Default: true},
		{Name: "ngx_http_core_module", Flag: "http", Directive: "http", Default: true},
		{Name: "ngx_http_charset_filter_module", Flag: "http_charset_module", Directive: "charset", Default: true},
		{Name: "ngx_http_log_module", Flag: "http_log_module", Directive: "access_log", Default: true},
		{Name: "ngx_http_proxy_module", Flag: "http_proxy_module", Directive: "proxy_pass", Default: true},
		{Name: "ngx_http_fastcgi_module", Flag: "http_fastcgi_module", Directive: "fastcgi_pass", Default: true},
		{Name: "ngx_http_memcached_module", Flag: "http_memcached_module", Directive: "memcached_pass", Default: true},
		{Name: "ngx_http_scgi_module", Flag: "http_scgi_module", Directive: "scgi_pass", Default: true},
		{Name: "ngx_http_uwsgi_module", Flag: "http_uwsgi_module", Directive: "uwsgi_pass", Default: true},
		{Name: "ngx_http_rewrite_module", Flag: "http_rewrite_module", Directive: "rewrite", Default: true},
		{Name: "ngx_http_headers_filter_module", Flag: "http_headers_module", Directive: "add_header", Default: true},
		{Name: "ngx_http_gzip_module", Flag: "http_gzip_module", Directive: "gzip", Default: true},
		{Name: "ngx_http_gunzip_filter_module", Flag: "http_gunzip_module", Directive: "gunzip", Default: false},
		{Name: "ngx_http_ssi_filter_module", Flag: "http_ssi_module", Directive: "ssi", Default: true},
		{Name: "ngx_http_access_module", Flag: "http_access_module", Directive: "allow", Default: true},
		{Name: "ngx_http_auth_basic_module", Flag: "http_auth_basic_module", Directive: "auth_basic", Default: true},
		{Name: "ngx_http_autoindex_module", Flag: "http_autoindex_module", Directive: "autoindex", Default: true},
		{Name: "ngx_http_browser_module", Flag: "http_browser_module", Directive: "modern_browser", Default: true},
		{Name: "ngx_http_empty_gif_module", Flag: "http_empty_gif_module", Directive: "empty_gif", Default: true},
		{Name: "ngx_http_index_module", Flag: "http_index_module", Directive: "index", Default: true},
		{Name: "ngx_http_referer_module", Flag: "http_referer_module", Directive: "valid_referers", Default: true},
		{Name: "ngx_http_static_module", Flag: "http_static_module", Directive: "root", Default: true},
		{Name: "ngx_http_upstream_module", Flag: "http_upstream_module", Directive: "upstream", Default: true},
		{Name: "ngx_http_upstream_hash_module", Flag: "http_upstream_hash_module", Directive: "hash", Default: true},
		{Name: "ngx_http_upstream_ip_hash_module", Flag: "http_upstream_ip_hash_module", Directive: "ip_hash", Default: true},
		{Name: "ngx_http_upstream_least_conn_module", Flag: "http_upstream_least_conn_module", Directive: "least_conn", Default: true},
		{Name: "ngx_http_upstream_keepalive_module", Flag: "http_upstream_keepalive_module", Directive: "keepalive", Default: true},
		{Name: "ngx_http_upstream_zone_module", Flag: "http_upstream_zone_module", Directive: "zone", Default: true},
		{Name: "ngx_http_limit_conn_module", Flag: "http_limit_conn_module", Directive: "limit_conn", Default: true},
		{Name: "ngx_http_limit_req_module", Flag: "http_limit_req_module", Directive: "limit_req", Default: true},
		{Name: "ngx_http_map_module", Flag: "http_map_module", Directive: "map", Default: true},
		{Name: "ngx_http_split_clients_module", Flag: "http_split_clients_module", Directive: "split_clients", Default: true},
		{Name: "ngx_http_geo_module", Flag: "http_geo_module", Directive: "geo", Default: true},
		{Name: "ngx_http_userid_filter_module", Flag: "http_userid_module", Directive: "userid", Default: true},
		{Name: "ngx_http_realip_module", Flag: "http_realip_module", Directive: "set_real_ip_from", Default: false},
		{Name: "ngx_http_ssl_module", Flag: "http_ssl_module", Directive: "ssl", Default: false},
		{Name: "ngx_http_v2_module", Flag: "http_v2_module", Directive: "http2", Default: false},
		{Name: "ngx_http_v3_module", Flag: "http_v3_module", Directive: "quic", Default: false},
		{Name: "ngx_http_stub_status_module", Flag: "http_stub_status_module", Directive: "stub_status", Default: false},
		{Name: "ngx_http_sub_filter_module", Flag: "http_sub_module", Directive: "sub_filter", Default: false},
		{Name: "ngx_http_gzip_static_module", Flag: "http_gzip_static_module", Directive: "gzip_static", Default: false},
		{Name: "ngx_http_auth_request_module", Flag: "http_auth_request_module", Directive: "auth_request", Default: false},
		{Name: "ngx_http_image_filter_module", Flag: "http_image_filter_module", Directive: "image_filter", Default: false},
		{Name: "ngx_http_xslt_filter_module", Flag: "http_xslt_module", Directive: "xslt_stylesheet", Default: false},
		{Name: "ngx_http_geoip_module", Flag: "http_geoip_module", Directive: "geoip_country", Default: false},
		{Name: "ngx_stream_module", Flag: "stream", Directive: "stream", Default: false},
		{Name: "ngx_stream_ssl_module", Flag: "stream_ssl_module", Directive: "ssl_preread", Default: false},
		{Name: "ngx_stream_ssl_preread_module", Flag: "stream_ssl_preread_module", Directive: "ssl_preread", Default: false},
		{Name: "ngx_mail_module", Flag: "mail", Directive: "mail", Default: false},
		{Name: "ngx_mail_ssl_module", Flag: "mail_ssl_module", Directive: "mail_ssl", Default: false},
	}
}

func (n *Nginx) nginxConfigureFlags(build string) map[string]string {
	flags := make(map[string]string)
	marker := "configure arguments:"
	if index := strings.Index(build, marker); index >= 0 {
		build = build[index+len(marker):]
	}
	for _, value := range strings.Fields(build) {
		value = strings.Trim(value, `"'`)
		if strings.HasPrefix(value, "--with-") || strings.HasPrefix(value, "--without-") || strings.HasPrefix(value, "--add-module=") || strings.HasPrefix(value, "--add-dynamic-module=") || strings.HasPrefix(value, "--modules-path=") {
			name := value
			parameter := "true"
			if index := strings.IndexByte(value, '='); index >= 0 {
				name = value[:index]
				parameter = value[index+1:]
			}
			flags[name] = parameter
		}
	}
	return flags
}

func (n *Nginx) nginxConfigureValues(build, name string) []string {
	values := make([]string, 0)
	prefix := name + "="
	for _, value := range strings.Fields(build) {
		value = strings.Trim(value, `"'`)
		if strings.HasPrefix(value, prefix) {
			values = append(values, strings.TrimPrefix(value, prefix))
		}
	}
	return values
}

func (n *Nginx) canonicalModuleName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = filepath.Base(value)
	value = strings.TrimSuffix(value, ".so")
	value = strings.TrimPrefix(value, "ngx_")
	if value == "" {
		return ""
	}
	if value == "core" {
		return "ngx_core_module"
	}
	if value == "http" {
		return "ngx_http_core_module"
	}
	if value != "http" && !strings.HasPrefix(value, "http_") && !strings.HasPrefix(value, "stream") && !strings.HasPrefix(value, "mail") {
		value = "http_" + value
	}
	if !strings.HasSuffix(value, "_module") {
		value += "_module"
	}
	return "ngx_" + value
}

func (n *Nginx) resolveModulePath(path string, info *Info) string {
	path = strings.Trim(strings.TrimSpace(path), `"'`)
	path = strings.ReplaceAll(path, "$prefix", info.PrefixPath)
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	if info.PrefixPath != "" {
		return filepath.Clean(filepath.Join(info.PrefixPath, path))
	}
	return filepath.Clean(path)
}

func (n *Nginx) walkConfigNodes(nodes []ConfigNode, callback func(ConfigNode)) {
	for _, node := range nodes {
		callback(node)
		if len(node.Children) > 0 {
			n.walkConfigNodes(node.Children, callback)
		}
	}
}

// ReadMainConfig 读取主配置文件的完整内容。
func (n *Nginx) ReadMainConfig() (string, error) {
	path, err := n.resolveConfigFile("")
	if err != nil {
		return "", err
	}
	return n.readTextFile(path)
}

// WriteMainConfig 原子更新主配置，并在提交前执行 nginx -t；检查失败会自动回滚。
func (n *Nginx) WriteMainConfig(content string) error {
	return n.WriteConfigFile("", content)
}

// GetConfig ReadMainConfig 的简短别名，便于控制器直接读取主配置。
func (n *Nginx) GetConfig() (string, error) {
	return n.ReadMainConfig()
}

// UpdateConfig WriteMainConfig 的简短别名，写入前会自动校验并失败回滚。
func (n *Nginx) UpdateConfig(content string) error {
	return n.WriteMainConfig(content)
}

// ReadConfigFile 读取主配置目录下的配置文件。path 可以是绝对路径或相对主配置目录的路径。
func (n *Nginx) ReadConfigFile(path string) (string, error) {
	path, err := n.resolveConfigFile(path)
	if err != nil {
		return "", err
	}
	return n.readTextFile(path)
}

// ParseConfig 解析 Nginx 配置文本，支持注释、引号、转义、分号和任意嵌套配置块。
func (n *Nginx) ParseConfig(content string) (*ConfigDocument, error) {
	return n.parseConfigDocument(content, "")
}

// ReadConfigDocument 解析单个配置文件，不展开 include。
func (n *Nginx) ReadConfigDocument(path string) (*ConfigDocument, error) {
	resolved, err := n.resolveConfigFile(path)
	if err != nil {
		return nil, err
	}
	content, err := n.readTextFile(resolved)
	if err != nil {
		return nil, err
	}
	return n.parseConfigDocument(content, resolved)
}

// ReadConfigTree 解析主配置并递归展开 include 文件。
// include 节点本身会保留，展开后的节点放在该节点的 Children 中，并带有 Source 文件路径。
func (n *Nginx) ReadConfigTree() (*ConfigDocument, error) {
	mainPath, err := n.resolveConfigFile("")
	if err != nil {
		return nil, err
	}
	includeBase := filepath.Dir(mainPath)
	if info, infoErr := n.Info(); infoErr == nil && info.PrefixPath != "" {
		includeBase = info.PrefixPath
	}
	visited := make(map[string]struct{})
	files := make([]string, 0)
	document, err := n.readConfigTreeFile(mainPath, visited, &files, includeBase)
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	document.Files = files
	return document, nil
}

// ReadLogConfig 解析主配置及所有 include 文件中的 log_format、access_log 和 error_log。
func (n *Nginx) ReadLogConfig() (*LogConfigDefinition, error) {
	document, err := n.ReadConfigTree()
	if err != nil {
		return nil, err
	}
	return n.extractLogConfig(document.Nodes), nil
}

// ParseLogConfig 解析一段配置文本中的复杂日志定义，不展开 include。
func (n *Nginx) ParseLogConfig(content string) (*LogConfigDefinition, error) {
	document, err := n.ParseConfig(content)
	if err != nil {
		return nil, err
	}
	return n.extractLogConfig(document.Nodes), nil
}

// ReadSiteLogConfig 解析单个站点文件中的全部访问日志和错误日志定义。
func (n *Nginx) ReadSiteLogConfig(name string) (*LogConfigDefinition, error) {
	path, err := n.sitePath(name)
	if err != nil {
		return nil, err
	}
	if _, statErr := os.Stat(path); statErr != nil {
		layout, layoutErr := n.SiteLayout()
		if layoutErr != nil || layout.Mode != "debian" {
			return nil, statErr
		}
		name, nameErr := n.normalizeSiteName(name)
		if nameErr != nil {
			return nil, nameErr
		}
		path = filepath.Join(layout.AvailableDir, name)
	}
	includeBase := filepath.Dir(path)
	if info, infoErr := n.Info(); infoErr == nil && info.PrefixPath != "" {
		includeBase = info.PrefixPath
	}
	visited := make(map[string]struct{})
	files := make([]string, 0)
	document, err := n.readConfigTreeFile(path, visited, &files, includeBase)
	if err != nil {
		return nil, err
	}
	return n.extractLogConfig(document.Nodes), nil
}

func (n *Nginx) extractLogConfig(nodes []ConfigNode) *LogConfigDefinition {
	result := &LogConfigDefinition{
		Formats:    make([]LogFormatDefinition, 0),
		Directives: make([]LogDirectiveDefinition, 0),
	}
	var walk func([]ConfigNode, []string)
	walk = func(current []ConfigNode, scope []string) {
		for _, node := range current {
			switch node.Name {
			case "log_format":
				if format, ok := n.parseLogFormat(node); ok {
					result.Formats = append(result.Formats, format)
				}
			case "access_log":
				if directive, ok := n.parseAccessLog(node); ok {
					directive.Scope = append([]string(nil), scope...)
					result.Directives = append(result.Directives, directive)
				}
			case "error_log":
				if directive, ok := n.parseErrorLog(node); ok {
					directive.Scope = append([]string(nil), scope...)
					result.Directives = append(result.Directives, directive)
				}
			}
			if len(node.Children) == 0 {
				continue
			}
			nextScope := scope
			if node.Name != "include" {
				nextScope = append(append([]string(nil), scope...), node.Name)
			}
			walk(node.Children, nextScope)
		}
	}
	walk(nodes, nil)
	return result
}

func (n *Nginx) parseLogFormat(node ConfigNode) (LogFormatDefinition, bool) {
	if len(node.Args) < 2 {
		return LogFormatDefinition{}, false
	}
	definition := LogFormatDefinition{
		Name:   node.Args[0],
		Source: node.Source,
		Line:   node.Line,
	}
	formatArgs := append([]string(nil), node.Args[1:]...)
	if len(formatArgs) > 0 && strings.HasPrefix(formatArgs[0], "escape=") {
		definition.Escape = strings.TrimPrefix(formatArgs[0], "escape=")
		formatArgs = formatArgs[1:]
	}
	if len(formatArgs) == 0 {
		return LogFormatDefinition{}, false
	}
	definition.Format = strings.Join(formatArgs, " ")
	definition.Variables = n.extractLogVariables(definition.Format)
	return definition, true
}

func (n *Nginx) parseAccessLog(node ConfigNode) (LogDirectiveDefinition, bool) {
	if len(node.Args) == 0 {
		return LogDirectiveDefinition{}, false
	}
	definition := LogDirectiveDefinition{
		Kind:    "access",
		Path:    node.Args[0],
		Source:  node.Source,
		Line:    node.Line,
		Enabled: !strings.EqualFold(node.Args[0], "off"),
		Syslog:  strings.HasPrefix(strings.ToLower(node.Args[0]), "syslog:"),
	}
	for _, argument := range node.Args[1:] {
		if n.isAccessLogOption(argument) {
			definition.Options = append(definition.Options, argument)
			continue
		}
		if definition.Format == "" {
			definition.Format = argument
		} else {
			definition.Options = append(definition.Options, argument)
		}
	}
	return definition, true
}

func (n *Nginx) parseErrorLog(node ConfigNode) (LogDirectiveDefinition, bool) {
	if len(node.Args) == 0 {
		return LogDirectiveDefinition{}, false
	}
	definition := LogDirectiveDefinition{
		Kind:    "error",
		Path:    node.Args[0],
		Source:  node.Source,
		Line:    node.Line,
		Enabled: !strings.EqualFold(node.Args[0], "off"),
		Syslog:  strings.HasPrefix(strings.ToLower(node.Args[0]), "syslog:"),
	}
	if len(node.Args) > 1 {
		if n.isErrorLogLevel(node.Args[1]) {
			definition.Level = node.Args[1]
			definition.Options = append(definition.Options, node.Args[2:]...)
		} else {
			definition.Options = append(definition.Options, node.Args[1:]...)
		}
	}
	return definition, true
}

func (n *Nginx) isAccessLogOption(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "gzip" || strings.HasPrefix(value, "gzip=") || strings.HasPrefix(value, "buffer=") || strings.HasPrefix(value, "flush=") || strings.HasPrefix(value, "if=")
}

func (n *Nginx) isErrorLogLevel(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug", "info", "notice", "warn", "error", "crit", "alert", "emerg":
		return true
	default:
		return false
	}
}

func (n *Nginx) extractLogVariables(format string) []string {
	variables := make([]string, 0)
	seen := make(map[string]struct{})
	for index := 0; index < len(format); index++ {
		if format[index] != '$' {
			continue
		}
		start := index
		index++
		if index < len(format) && format[index] == '{' {
			index++
			for index < len(format) && format[index] != '}' {
				index++
			}
			if index >= len(format) {
				break
			}
		} else {
			for index < len(format) {
				char := format[index]
				if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '_' {
					break
				}
				index++
			}
			index--
		}
		value := format[start : index+1]
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		variables = append(variables, value)
	}
	return variables
}

func (n *Nginx) parseConfigDocument(content, source string) (*ConfigDocument, error) {
	tokens, err := n.tokenizeConfig(content)
	if err != nil {
		return nil, err
	}
	nodes, index, err := n.parseConfigNodes(tokens, 0, false, source)
	if err != nil {
		return nil, err
	}
	if index != len(tokens) {
		return nil, errors.New("Nginx 配置解析失败: 存在未处理的配置内容")
	}
	document := &ConfigDocument{Path: source, Nodes: nodes}
	if source != "" {
		document.Files = []string{source}
	}
	return document, nil
}

func (n *Nginx) tokenizeConfig(content string) ([]nginxConfigToken, error) {
	tokens := make([]nginxConfigToken, 0)
	var token strings.Builder
	line := 1
	tokenLine := 1
	var quote byte
	tokenStarted := false
	flush := func() {
		if !tokenStarted {
			return
		}
		tokens = append(tokens, nginxConfigToken{Value: token.String(), Line: tokenLine})
		token.Reset()
		tokenStarted = false
	}
	for index := 0; index < len(content); index++ {
		char := content[index]
		if quote != 0 {
			if char == quote {
				quote = 0
				tokenStarted = true
				continue
			}
			if char == '\\' {
				token.WriteByte(char)
				tokenStarted = true
				if index+1 < len(content) {
					index++
					next := content[index]
					token.WriteByte(next)
					if next == '\n' {
						line++
					}
				}
				continue
			}
			if char == '\n' {
				line++
			}
			token.WriteByte(char)
			tokenStarted = true
			continue
		}
		switch char {
		case '\n':
			line++
		case ' ', '\t', '\r':
			flush()
		case '#':
			for index+1 < len(content) && content[index+1] != '\n' {
				index++
			}
		case '\'', '"':
			if !tokenStarted {
				tokenLine = line
			}
			quote = char
			tokenStarted = true
		case ';', '{', '}':
			flush()
			tokens = append(tokens, nginxConfigToken{Value: string(char), Line: line})
		default:
			if !tokenStarted {
				tokenLine = line
			}
			if char == '\\' {
				token.WriteByte(char)
				tokenStarted = true
				if index+1 < len(content) {
					index++
					next := content[index]
					token.WriteByte(next)
					if next == '\n' {
						line++
					}
				}
				continue
			}
			token.WriteByte(char)
			tokenStarted = true
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("Nginx 配置解析失败: 第 %d 行引号未闭合", line)
	}
	flush()
	return tokens, nil
}

func (n *Nginx) parseConfigNodes(tokens []nginxConfigToken, index int, nested bool, source string) ([]ConfigNode, int, error) {
	nodes := make([]ConfigNode, 0)
	for index < len(tokens) {
		if tokens[index].Value == "}" {
			if !nested {
				return nil, index, fmt.Errorf("Nginx 配置解析失败: 第 %d 行出现多余的 }", tokens[index].Line)
			}
			return nodes, index + 1, nil
		}
		if tokens[index].Value == ";" || tokens[index].Value == "{" {
			return nil, index, fmt.Errorf("Nginx 配置解析失败: 第 %d 行指令名称为空", tokens[index].Line)
		}
		line := tokens[index].Line
		arguments := make([]string, 0, 2)
		for index < len(tokens) {
			value := tokens[index].Value
			if value == ";" || value == "{" || value == "}" {
				break
			}
			arguments = append(arguments, value)
			index++
		}
		if len(arguments) == 0 {
			return nil, index, fmt.Errorf("Nginx 配置解析失败: 第 %d 行指令名称为空", line)
		}
		node := ConfigNode{Name: arguments[0], Source: source, Line: line}
		if len(arguments) > 1 {
			node.Args = append([]string(nil), arguments[1:]...)
		}
		if index >= len(tokens) {
			return nil, index, fmt.Errorf("Nginx 配置解析失败: 第 %d 行缺少分号或配置块", line)
		}
		switch tokens[index].Value {
		case ";":
			index++
		case "{":
			children, next, err := n.parseConfigNodes(tokens, index+1, true, source)
			if err != nil {
				return nil, index, err
			}
			node.Block = true
			node.Children = children
			index = next
		case "}":
			return nil, index, fmt.Errorf("Nginx 配置解析失败: 第 %d 行指令缺少分号", line)
		default:
			return nil, index, fmt.Errorf("Nginx 配置解析失败: 第 %d 行语法不完整", line)
		}
		nodes = append(nodes, node)
	}
	if nested {
		return nil, index, errors.New("Nginx 配置解析失败: 配置块缺少 }")
	}
	return nodes, index, nil
}

func (n *Nginx) readConfigTreeFile(path string, visited map[string]struct{}, files *[]string, includeBase string) (*ConfigDocument, error) {
	path, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	if realPath, evalErr := filepath.EvalSymlinks(path); evalErr == nil {
		path = realPath
	}
	if _, exists := visited[path]; exists {
		return &ConfigDocument{Path: path}, nil
	}
	if len(visited) >= 2048 {
		return nil, errors.New("Nginx include 文件数量超过限制")
	}
	visited[path] = struct{}{}
	*files = append(*files, path)
	content, err := n.readTextFile(path)
	if err != nil {
		return nil, err
	}
	document, err := n.parseConfigDocument(content, path)
	if err != nil {
		return nil, err
	}
	document.Nodes, err = n.expandConfigIncludes(document.Nodes, includeBase, visited, files)
	if err != nil {
		return nil, err
	}
	return document, nil
}

func (n *Nginx) expandConfigIncludes(nodes []ConfigNode, includeBase string, visited map[string]struct{}, files *[]string) ([]ConfigNode, error) {
	for index := range nodes {
		node := &nodes[index]
		if len(node.Children) > 0 {
			children, err := n.expandConfigIncludes(node.Children, includeBase, visited, files)
			if err != nil {
				return nil, err
			}
			node.Children = children
		}
		if node.Name != "include" || len(node.Args) == 0 {
			continue
		}
		for _, pattern := range node.Args {
			pattern = strings.TrimSpace(pattern)
			pattern = strings.ReplaceAll(pattern, "$prefix", includeBase)
			if pattern == "" || strings.ContainsAny(pattern, "$\x00\r\n") {
				continue
			}
			if !filepath.IsAbs(pattern) {
				pattern = filepath.Join(includeBase, pattern)
			}
			matches, globErr := filepath.Glob(pattern)
			if globErr != nil {
				return nil, fmt.Errorf("解析 include 路径失败: %w", globErr)
			}
			sort.Strings(matches)
			for _, match := range matches {
				included, readErr := n.readConfigTreeFile(match, visited, files, includeBase)
				if readErr != nil {
					return nil, readErr
				}
				node.Children = append(node.Children, included.Nodes...)
			}
		}
	}
	return nodes, nil
}

// WriteConfigFile 原子更新配置目录下的文件，并校验整个 Nginx 配置。
func (n *Nginx) WriteConfigFile(path, content string) error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	path, err := n.resolveConfigFile(path)
	if err != nil {
		return err
	}
	if err = n.ensureConfigWritable(path); err != nil {
		return err
	}
	if strings.TrimSpace(content) == "" {
		return errors.New("Nginx 配置不能为空")
	}
	if _, err = n.backupConfig(path); err != nil {
		return fmt.Errorf("备份 Nginx 配置失败: %w", err)
	}
	return n.writeValidatedConfig(path, []byte(content))
}

// RemoveConfigFile 删除配置目录下的配置文件，并在删除后校验配置；失败会恢复原文件。
func (n *Nginx) RemoveConfigFile(path string) error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	path, err := n.resolveConfigFile(path)
	if err != nil {
		return err
	}
	if err = n.ensureConfigWritable(path); err != nil {
		return err
	}
	mainPath, err := n.resolveConfigFile("")
	if err != nil {
		return err
	}
	if filepath.Clean(path) == filepath.Clean(mainPath) {
		return errors.New("不能删除 Nginx 主配置")
	}
	if _, err = n.backupConfig(path); err != nil {
		return fmt.Errorf("备份 Nginx 配置失败: %w", err)
	}
	old, oldMode, err := n.readSite(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	linkTarget, linkErr := os.Readlink(path)
	if err = os.Remove(path); err != nil {
		return err
	}
	if err = n.Test(); err != nil {
		if linkErr == nil {
			_ = os.Symlink(linkTarget, path)
		} else {
			_ = n.atomicWrite(path, old, oldMode)
		}
		return err
	}
	return nil
}

// ListConfigFiles 列出主配置目录下的 .conf 文件。
func (n *Nginx) ListConfigFiles() ([]ConfigFile, error) {
	root, err := n.configRoot()
	if err != nil {
		return nil, err
	}
	mainPath, err := n.resolveConfigFile("")
	if err != nil {
		return nil, err
	}
	files := make([]ConfigFile, 0)
	err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		lowerPath := strings.ToLower(path)
		isConfig := strings.HasSuffix(lowerPath, ".conf") || strings.HasSuffix(lowerPath, ".conf.disabled") || path == mainPath
		if info.IsDir() || !isConfig {
			return nil
		}
		files = append(files, ConfigFile{Path: path, Size: info.Size(), ModifiedAt: info.ModTime()})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("列出 Nginx 配置文件失败: %w", err)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

// BackupConfig 创建配置版本备份。未传路径时备份配置目录中的全部配置文件。
func (n *Nginx) BackupConfig(paths ...string) (*ConfigBackup, error) {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	return n.backupConfig(paths...)
}

// ListConfigBackups 列出配置版本。
func (n *Nginx) ListConfigBackups() ([]ConfigBackup, error) {
	root, err := n.backupRoot()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return []ConfigBackup{}, nil
		}
		return nil, err
	}
	backups := make([]ConfigBackup, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		backup, readErr := n.readBackupManifest(filepath.Join(root, entry.Name()))
		if readErr != nil {
			continue
		}
		backups = append(backups, *backup)
	}
	sort.Slice(backups, func(i, j int) bool { return backups[i].CreatedAt.After(backups[j].CreatedAt) })
	return backups, nil
}

// RestoreConfigBackup 恢复一个配置版本，并在恢复后执行 nginx -t。
func (n *Nginx) RestoreConfigBackup(id string) error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	backup, err := n.getBackup(id)
	if err != nil {
		return err
	}
	paths := make([]string, 0, len(backup.Files))
	for _, item := range backup.Files {
		path, pathErr := n.resolveConfigFile(item.Path)
		if pathErr != nil {
			return pathErr
		}
		paths = append(paths, path)
	}
	rollback, err := n.backupConfig(paths...)
	if err != nil {
		return err
	}
	if err = n.restoreBackupFiles(backup); err != nil {
		_ = n.restoreBackupFiles(rollback)
		return err
	}
	if err = n.Test(); err != nil {
		_ = n.restoreBackupFiles(rollback)
		return err
	}
	return nil
}

// DeleteConfigBackup 删除一个配置版本目录。
func (n *Nginx) DeleteConfigBackup(id string) error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	if strings.TrimSpace(id) == "" || strings.ContainsAny(id, `/\\`) {
		return errors.New("配置备份 ID 不合法")
	}
	root, err := n.backupRoot()
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(root, id))
}

// ApplyConfigBatch 原子地应用多个配置文件变更，失败时恢复全部文件。
func (n *Nginx) ApplyConfigBatch(changes []ConfigChange) (*ConfigBackup, error) {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	if len(changes) == 0 {
		return nil, errors.New("配置变更不能为空")
	}
	paths := make([]string, 0, len(changes))
	seen := make(map[string]struct{}, len(changes))
	for _, change := range changes {
		path, err := n.resolveConfigFile(change.Path)
		if err != nil {
			return nil, err
		}
		mainPath, _ := n.resolveConfigFile("")
		if change.Remove && filepath.Clean(path) == filepath.Clean(mainPath) {
			return nil, errors.New("不能删除 Nginx 主配置")
		}
		if _, ok := seen[path]; ok {
			return nil, fmt.Errorf("配置文件重复变更: %s", path)
		}
		seen[path] = struct{}{}
		if err = n.ensureConfigWritable(path); err != nil {
			return nil, err
		}
		if !change.Remove && strings.TrimSpace(change.Content) == "" {
			return nil, errors.New("Nginx 配置不能为空")
		}
		paths = append(paths, path)
	}
	backup, err := n.backupConfig(paths...)
	if err != nil {
		return nil, err
	}
	old := make(map[string]fileState, len(paths))
	for _, path := range paths {
		old[path] = n.captureFileState(path)
	}
	for index, change := range changes {
		path := paths[index]
		if change.Remove {
			if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
				_ = n.restoreFileStates(old)
				return backup, err
			}
			continue
		}
		if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			_ = n.restoreFileStates(old)
			return backup, err
		}
		if err = n.atomicWrite(path, []byte(change.Content), old[path].Mode); err != nil {
			_ = n.restoreFileStates(old)
			return backup, err
		}
	}
	if err = n.Test(); err != nil {
		_ = n.restoreFileStates(old)
		return backup, err
	}
	return backup, nil
}

type fileState struct {
	Exists bool
	Data   []byte
	Mode   os.FileMode
}

func (n *Nginx) captureFileState(path string) fileState {
	data, mode, err := n.readSite(path)
	if err != nil {
		return fileState{Mode: 0644}
	}
	return fileState{Exists: true, Data: data, Mode: mode}
}

func (n *Nginx) restoreFileStates(states map[string]fileState) error {
	for path, state := range states {
		if !state.Exists {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
			continue
		}
		if err := n.atomicWrite(path, state.Data, state.Mode); err != nil {
			return err
		}
	}
	return nil
}

func (n *Nginx) backupRoot() (string, error) {
	root, err := n.configRoot()
	if err != nil {
		return "", err
	}
	n.mu.RLock()
	backupDir := n.backupDir
	n.mu.RUnlock()
	if backupDir == "" {
		backupDir = filepath.Join(root, ".nginx-backups")
	} else if !filepath.IsAbs(backupDir) {
		backupDir = filepath.Join(root, backupDir)
	}
	backupDir, err = filepath.Abs(filepath.Clean(backupDir))
	if err != nil {
		return "", err
	}
	if err = os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("创建 Nginx 备份目录失败: %w", err)
	}
	return backupDir, nil
}

func (n *Nginx) backupConfig(paths ...string) (*ConfigBackup, error) {
	root, err := n.configRoot()
	if err != nil {
		return nil, err
	}
	backupRoot, err := n.backupRoot()
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		files, listErr := n.ListConfigFiles()
		if listErr != nil {
			return nil, listErr
		}
		paths = make([]string, 0, len(files))
		for _, file := range files {
			paths = append(paths, file.Path)
		}
	}
	id := strconv.FormatInt(time.Now().UnixNano(), 10)
	directory := filepath.Join(backupRoot, id)
	if err = os.MkdirAll(filepath.Join(directory, "files"), 0755); err != nil {
		return nil, err
	}
	backup := &ConfigBackup{ID: id, CreatedAt: time.Now().UTC(), Files: make([]ConfigBackupFile, 0, len(paths))}
	for _, rawPath := range paths {
		path, pathErr := n.resolveConfigFile(rawPath)
		if pathErr != nil {
			_ = os.RemoveAll(directory)
			return nil, pathErr
		}
		info, statErr := os.Stat(path)
		if os.IsNotExist(statErr) {
			continue
		}
		if statErr != nil {
			_ = os.RemoveAll(directory)
			return nil, statErr
		}
		if info.IsDir() || info.Size() > maxConfigFileSize {
			_ = os.RemoveAll(directory)
			return nil, fmt.Errorf("配置文件无法备份: %s", path)
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			_ = os.RemoveAll(directory)
			return nil, readErr
		}
		relative, relErr := filepath.Rel(root, path)
		if relErr != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
			_ = os.RemoveAll(directory)
			return nil, errors.New("配置文件必须位于 Nginx 配置目录内")
		}
		backupPath := filepath.Join(directory, "files", relative)
		if err = os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
			_ = os.RemoveAll(directory)
			return nil, err
		}
		if err = os.WriteFile(backupPath, data, info.Mode().Perm()); err != nil {
			_ = os.RemoveAll(directory)
			return nil, err
		}
		digest := sha256.Sum256(data)
		backup.Files = append(backup.Files, ConfigBackupFile{
			Path:       relative,
			BackupPath: backupPath,
			Size:       int64(len(data)),
			SHA256:     hex.EncodeToString(digest[:]),
			Mode:       uint32(info.Mode().Perm()),
		})
	}
	manifest, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		_ = os.RemoveAll(directory)
		return nil, err
	}
	if err = os.WriteFile(filepath.Join(directory, "manifest.json"), manifest, 0644); err != nil {
		_ = os.RemoveAll(directory)
		return nil, err
	}
	return backup, nil
}

func (n *Nginx) backupSitePaths(paths ...string) error {
	root, err := n.configRoot()
	if err != nil {
		return err
	}
	inside := make([]string, 0, len(paths))
	for _, path := range paths {
		absolute, absErr := filepath.Abs(filepath.Clean(path))
		if absErr != nil {
			continue
		}
		relative, relErr := filepath.Rel(root, absolute)
		if relErr == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
			inside = append(inside, absolute)
		}
	}
	if len(inside) == 0 {
		return nil
	}
	_, err = n.backupConfig(inside...)
	return err
}

func (n *Nginx) readBackupManifest(directory string) (*ConfigBackup, error) {
	data, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var backup ConfigBackup
	if err = json.Unmarshal(data, &backup); err != nil {
		return nil, err
	}
	if backup.ID == "" {
		backup.ID = filepath.Base(directory)
	}
	return &backup, nil
}

func (n *Nginx) getBackup(id string) (*ConfigBackup, error) {
	id = strings.TrimSpace(id)
	if id == "" || strings.ContainsAny(id, `/\\`) {
		return nil, errors.New("配置备份 ID 不合法")
	}
	root, err := n.backupRoot()
	if err != nil {
		return nil, err
	}
	return n.readBackupManifest(filepath.Join(root, id))
}

func (n *Nginx) restoreBackupFiles(backup *ConfigBackup) error {
	if backup == nil {
		return errors.New("配置备份不存在")
	}
	for _, item := range backup.Files {
		path, err := n.resolveConfigFile(item.Path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(item.BackupPath)
		if err != nil {
			return err
		}
		digest := sha256.Sum256(data)
		if item.SHA256 != "" && item.SHA256 != hex.EncodeToString(digest[:]) {
			return fmt.Errorf("配置备份校验失败: %s", item.Path)
		}
		if err = n.atomicWrite(path, data, os.FileMode(item.Mode)); err != nil {
			return err
		}
	}
	return nil
}

// Start 启动 Nginx。启动前会先检查配置。
func (n *Nginx) Start() error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	info, err := n.Info()
	if err != nil {
		return err
	}
	status, statusErr := n.statusForInfo(info)
	if statusErr == nil && status.Running {
		return nil
	}
	if err = n.Test(); err != nil {
		return err
	}
	cmd := exec.Command(info.BinaryPath, n.runtimeArgs(info)...)
	if err = cmd.Start(); err != nil {
		if serviceErr := n.runService("start"); serviceErr == nil {
			return n.waitReady()
		}
		return fmt.Errorf("启动 nginx 失败: %w", err)
	}
	if err = cmd.Process.Release(); err != nil {
		return fmt.Errorf("释放 nginx 启动进程失败: %w", err)
	}
	if waitErr := n.waitReady(); waitErr == nil {
		return nil
	} else if serviceErr := n.runService("start"); serviceErr == nil {
		return n.waitReady()
	} else {
		return fmt.Errorf("%w；尝试系统服务启动失败: %v", waitErr, serviceErr)
	}
}

// Reload 平滑重载 Nginx 配置。
func (n *Nginx) Reload() error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	info, err := n.Info()
	if err != nil {
		return err
	}
	if status, statusErr := n.statusForInfo(info); statusErr == nil && !status.Running {
		return errors.New("Nginx 当前未运行，不能重载配置")
	}
	if err = n.Test(); err != nil {
		return err
	}
	args := n.runtimeArgs(info, "-s", "reload")
	output, err := n.runCommand(context.Background(), n.commandTimeout(), info.BinaryPath, args...)
	if err != nil {
		if serviceErr := n.runService("reload"); serviceErr == nil {
			return nil
		}
		return fmt.Errorf("重载 nginx 失败: %w", n.formatCommandOutput(output))
	}
	return nil
}

// Stop 停止 Nginx。graceful 为 true 时使用 quit 平滑退出，否则使用 stop 强制退出。
func (n *Nginx) Stop(graceful bool) error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	info, err := n.Info()
	if err != nil {
		return err
	}
	if status, statusErr := n.statusForInfo(info); statusErr == nil && !status.Running {
		return nil
	}
	operation := "stop"
	if graceful {
		operation = "quit"
	}
	args := n.runtimeArgs(info, "-s", operation)
	output, err := n.runCommand(context.Background(), n.commandTimeout(), info.BinaryPath, args...)
	if err != nil {
		if serviceErr := n.runService(operation); serviceErr == nil {
			return n.waitStopped()
		}
		if status, statusErr := n.Status(); statusErr == nil && !status.Running {
			return nil
		}
		return fmt.Errorf("停止 nginx 失败: %w", n.formatCommandOutput(output))
	}
	return n.waitStopped()
}

// Restart 重启 Nginx。
func (n *Nginx) Restart() error {
	if err := n.Stop(true); err != nil {
		return err
	}
	return n.Start()
}

// SetAutoStart 设置 Nginx 是否随系统启动。
func (n *Nginx) SetAutoStart(enabled bool) error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	return n.setAutoStart(enabled)
}

func (n *Nginx) setAutoStart(enabled bool) error {
	serviceName := n.getServiceName()
	if runtime.GOOS == "windows" {
		if !n.commandExists("sc.exe") {
			return errors.New("未找到 Windows 服务管理器")
		}
		startType := "disabled"
		if enabled {
			startType = "auto"
		}
		output, err := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "sc.exe", "config", serviceName, "start=", startType)
		if err != nil {
			return n.formatCommandOutput(output)
		}
		return nil
	}
	manager := n.detectInitSystem()
	if manager == "systemd" {
		action := "disable"
		if enabled {
			action = "enable"
		}
		output, err := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "systemctl", action, serviceName)
		if err == nil {
			return nil
		}
		return n.formatCommandOutput(output)
	}
	if manager == "openrc" && n.commandExists("rc-update") {
		args := []string{"del", serviceName, "default"}
		if enabled {
			args = []string{"add", serviceName, "default"}
		}
		output, err := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "rc-update", args...)
		if err == nil {
			return nil
		}
		return n.formatCommandOutput(output)
	}
	if manager == "procd" {
		servicePath := filepath.Join("/etc/init.d", serviceName)
		if !n.fileExists(servicePath) {
			return errors.New("未找到 procd 服务脚本")
		}
		action := "disable"
		if enabled {
			action = "enable"
		}
		output, err := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", servicePath, action)
		if err == nil {
			return nil
		}
		return n.formatCommandOutput(output)
	}
	if manager == "runit" {
		serviceRoot := n.runitServiceRoot()
		servicePath := filepath.Join("/etc/sv", serviceName)
		if enabled {
			if !n.fileExists(filepath.Join(servicePath, "run")) {
				return errors.New("未找到 runit 服务脚本")
			}
			if err := n.ensurePrivilegedDirectory(servicePath); err != nil {
				return err
			}
			return n.linkPrivileged(servicePath, filepath.Join(serviceRoot, serviceName))
		}
		return n.removePrivilegedFile(filepath.Join(serviceRoot, serviceName))
	}
	if manager == "sysvinit" {
		if n.commandExists("update-rc.d") {
			args := []string{serviceName, "disable"}
			if enabled {
				args = []string{serviceName, "defaults"}
			}
			output, err := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "update-rc.d", args...)
			if err == nil {
				return nil
			}
			return n.formatCommandOutput(output)
		}
		if n.commandExists("chkconfig") {
			action := "off"
			if enabled {
				action = "on"
			}
			output, err := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "chkconfig", serviceName, action)
			if err == nil {
				return nil
			}
			return n.formatCommandOutput(output)
		}
	}
	return errors.New("当前系统不支持设置 Nginx 开机启动")
}

// InstallService 为当前实例创建系统服务。Linux 按运行中的 systemd、OpenRC、procd、runit、SysVinit 选择；Windows 使用 sc.exe。
func (n *Nginx) InstallService(options ServiceOptions) error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	info, err := n.Info()
	if err != nil {
		return err
	}
	name := strings.TrimSpace(options.Name)
	if name == "" {
		name = n.getServiceName()
	}
	if _, err = n.normalizeIdentifier(name); err != nil {
		return fmt.Errorf("服务名称不合法: %w", err)
	}
	description := strings.TrimSpace(options.Description)
	if description == "" {
		description = "Nginx service (" + name + ")"
	}
	if options.User == "" {
		n.mu.RLock()
		options.User = n.runUser
		n.mu.RUnlock()
	}
	if options.User != "" {
		if _, err = n.normalizeIdentifier(options.User); err != nil {
			return fmt.Errorf("服务用户不合法: %w", err)
		}
	}
	if options.Group != "" {
		if _, err = n.normalizeIdentifier(options.Group); err != nil {
			return fmt.Errorf("服务用户组不合法: %w", err)
		}
	}
	if strings.TrimSpace(info.PidPath) == "" {
		info.PidPath = n.defaultPidPath(info)
	}
	n.SetServiceName(name)
	if runtime.GOOS == "windows" {
		command := n.serviceCommand(info)
		output, runErr := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "sc.exe", "create", name, "binPath=", command, "DisplayName=", description, "start=", "auto")
		if runErr != nil {
			return n.formatCommandOutput(output)
		}
		return nil
	}
	manager := n.detectInitSystem()
	if manager == "systemd" {
		unitPath := filepath.Join("/etc/systemd/system", name+".service")
		if err = n.ensureManagedServicePath(unitPath); err != nil {
			return err
		}
		unit := n.buildSystemdUnit(info, options, name, description)
		if err = n.writePrivilegedFile(unitPath, []byte(unit), 0644); err != nil {
			return err
		}
		if output, runErr := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "systemctl", "daemon-reload"); runErr != nil {
			_ = n.removePrivilegedFile(unitPath)
			return n.formatCommandOutput(output)
		}
		if options.Enabled {
			if err = n.setAutoStart(true); err != nil {
				_, _ = n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "systemctl", "disable", name)
				_ = n.removePrivilegedFile(unitPath)
				_, _ = n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "systemctl", "daemon-reload")
				return err
			}
		}
		return nil
	}
	if manager == "openrc" {
		unitPath := filepath.Join("/etc/init.d", name)
		if err = n.ensureManagedServicePath(unitPath); err != nil {
			return err
		}
		script := n.buildOpenRCService(info, options, name, description)
		if err = n.writePrivilegedFile(unitPath, []byte(script), 0755); err != nil {
			return err
		}
		if options.Enabled {
			if err = n.setAutoStart(true); err != nil {
				_ = n.setAutoStart(false)
				_ = n.removePrivilegedFile(unitPath)
				return err
			}
		}
		return nil
	}
	if manager == "procd" {
		unitPath := filepath.Join("/etc/init.d", name)
		if err = n.ensureManagedServicePath(unitPath); err != nil {
			return err
		}
		script := n.buildProcdService(info, options)
		if err = n.writePrivilegedFile(unitPath, []byte(script), 0755); err != nil {
			return err
		}
		if options.Enabled {
			if err = n.setAutoStart(true); err != nil {
				_ = n.setAutoStart(false)
				_ = n.removePrivilegedFile(unitPath)
				return err
			}
		}
		return nil
	}
	if manager == "runit" {
		servicePath := filepath.Join("/etc/sv", name)
		if err = n.ensurePrivilegedDirectory(servicePath); err != nil {
			return err
		}
		if err = n.ensureManagedServicePath(filepath.Join(servicePath, "run")); err != nil {
			return err
		}
		script := n.buildRunitService(info, options)
		if err = n.writePrivilegedFile(filepath.Join(servicePath, "run"), []byte(script), 0755); err != nil {
			return err
		}
		if options.Enabled {
			if err = n.setAutoStart(true); err != nil {
				_ = n.setAutoStart(false)
				_ = n.removePrivilegedDirectory(servicePath)
				return err
			}
		}
		return nil
	}
	if manager == "sysvinit" {
		unitPath := filepath.Join("/etc/init.d", name)
		if err = n.ensureManagedServicePath(unitPath); err != nil {
			return err
		}
		script := n.buildSysVService(info, options, name, description)
		if err = n.writePrivilegedFile(unitPath, []byte(script), 0755); err != nil {
			return err
		}
		if options.Enabled {
			if err = n.setAutoStart(true); err != nil {
				_ = n.setAutoStart(false)
				_ = n.removePrivilegedFile(unitPath)
				return err
			}
		}
		return nil
	}
	return errors.New("当前系统没有可用的 systemd、OpenRC、procd、runit、SysVinit 或 Windows 服务管理器")
}

// RemoveService 删除当前实例的系统服务定义，不会删除 Nginx 二进制和配置。
func (n *Nginx) RemoveService() error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	name := n.getServiceName()
	if runtime.GOOS == "windows" {
		output, err := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "sc.exe", "delete", name)
		if err != nil {
			return n.formatCommandOutput(output)
		}
		return nil
	}
	manager := n.detectInitSystem()
	if manager == "systemd" {
		path := filepath.Join("/etc/systemd/system", name+".service")
		if !n.managedServiceFile(path) {
			return errors.New("当前服务不是面板创建的定义，拒绝删除")
		}
		_, _ = n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "systemctl", "disable", "--now", name)
		if err := n.removePrivilegedFile(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		output, err := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "systemctl", "daemon-reload")
		if err != nil {
			return n.formatCommandOutput(output)
		}
		return nil
	}
	if manager == "openrc" {
		path := filepath.Join("/etc/init.d", name)
		if !n.managedServiceFile(path) {
			return errors.New("当前服务不是面板创建的定义，拒绝删除")
		}
		_, _ = n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "rc-service", name, "stop")
		if n.commandExists("rc-update") {
			_, _ = n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "rc-update", "del", name, "default")
		}
		if err := n.removePrivilegedFile(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if manager == "procd" {
		path := filepath.Join("/etc/init.d", name)
		if !n.managedServiceFile(path) {
			return errors.New("当前服务不是面板创建的定义，拒绝删除")
		}
		_, _ = n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", path, "stop")
		_, _ = n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", path, "disable")
		if err := n.removePrivilegedFile(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if manager == "runit" {
		if !n.managedServiceFile(filepath.Join("/etc/sv", name, "run")) {
			return errors.New("当前服务不是面板创建的定义，拒绝删除")
		}
		_, _ = n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "sv", "stop", name)
		_ = n.setAutoStart(false)
		return n.removePrivilegedDirectory(filepath.Join("/etc/sv", name))
	}
	if manager == "sysvinit" {
		path := filepath.Join("/etc/init.d", name)
		if !n.managedServiceFile(path) {
			return errors.New("当前服务不是面板创建的定义，拒绝删除")
		}
		if n.commandExists("service") {
			_, _ = n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "service", name, "stop")
		} else if _, statErr := os.Stat(path); statErr == nil {
			_, _ = n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", path, "stop")
		}
		_ = n.setAutoStart(false)
		if err := n.removePrivilegedFile(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	return errors.New("当前系统没有可用的服务管理器")
}

// ServiceStatus 获取当前实例的系统服务状态。
func (n *Nginx) ServiceStatus() (*ServiceInfo, error) {
	name := n.getServiceName()
	result := &ServiceInfo{Name: name}
	if runtime.GOOS == "windows" {
		result.Manager = "windows-service"
		output, err := n.runCommand(context.Background(), n.commandTimeout(), "sc.exe", "query", name)
		result.Active = err == nil && strings.Contains(strings.ToUpper(output), "RUNNING")
		return result, nil
	}
	manager := n.detectInitSystem()
	if manager == "systemd" {
		result.Manager = "systemd"
		result.UnitPath = filepath.Join("/etc/systemd/system", name+".service")
		active, activeErr := n.runCommand(context.Background(), n.commandTimeout(), "systemctl", "is-active", name)
		enabled, enabledErr := n.runCommand(context.Background(), n.commandTimeout(), "systemctl", "is-enabled", name)
		result.Active = activeErr == nil && strings.TrimSpace(active) == "active"
		result.Enabled = enabledErr == nil && strings.TrimSpace(enabled) == "enabled"
		return result, nil
	}
	if manager == "openrc" {
		result.Manager = "openrc"
		result.UnitPath = filepath.Join("/etc/init.d", name)
		output, err := n.runCommand(context.Background(), n.commandTimeout(), "rc-service", name, "status")
		result.Active = err == nil && !strings.Contains(strings.ToLower(output), "stopped")
		return result, nil
	}
	if manager == "procd" {
		result.Manager = "procd"
		result.UnitPath = filepath.Join("/etc/init.d", name)
		output, statusErr := n.runCommand(context.Background(), n.commandTimeout(), result.UnitPath, "status")
		result.Active = statusErr == nil && !strings.Contains(strings.ToLower(output), "not running") && !strings.Contains(strings.ToLower(output), "stopped")
		matches, _ := filepath.Glob(filepath.Join("/etc/rc.d", "S*"+name))
		result.Enabled = len(matches) > 0
		return result, nil
	}
	if manager == "runit" {
		result.Manager = "runit"
		result.UnitPath = filepath.Join("/etc/sv", name)
		output, statusErr := n.runCommand(context.Background(), n.commandTimeout(), "sv", "status", name)
		result.Active = statusErr == nil && strings.Contains(strings.ToLower(output), "run:")
		if _, linkErr := os.Lstat(filepath.Join(n.runitServiceRoot(), name)); linkErr == nil {
			result.Enabled = true
		}
		return result, nil
	}
	if manager == "sysvinit" {
		result.Manager = "sysvinit"
		result.UnitPath = filepath.Join("/etc/init.d", name)
		command := "service"
		args := []string{name, "status"}
		if !n.commandExists(command) {
			command = result.UnitPath
			args = []string{"status"}
		}
		output, statusErr := n.runCommand(context.Background(), n.commandTimeout(), command, args...)
		result.Active = statusErr == nil && !strings.Contains(strings.ToLower(output), "not running") && !strings.Contains(strings.ToLower(output), "stopped")
		if n.commandExists("chkconfig") {
			if enabled, enabledErr := n.runCommand(context.Background(), n.commandTimeout(), "chkconfig", "--list", name); enabledErr == nil {
				result.Enabled = !strings.Contains(strings.ToLower(enabled), "off")
			}
		}
		return result, nil
	}
	return nil, errors.New("当前系统没有可用的服务管理器")
}

func (n *Nginx) serviceCommand(info *Info) string {
	parts := []string{n.systemdQuote(info.BinaryPath)}
	for _, arg := range n.runtimeArgs(info) {
		parts = append(parts, n.systemdQuote(arg))
	}
	return strings.Join(parts, " ")
}

func (n *Nginx) buildSystemdUnit(info *Info, options ServiceOptions, name, description string) string {
	restart := strings.TrimSpace(options.Restart)
	switch restart {
	case "no", "on-success", "on-failure", "on-abnormal", "on-watchdog", "on-abort", "always":
	default:
		restart = "on-failure"
	}
	workingDirectory := strings.TrimSpace(options.WorkingDirectory)
	if workingDirectory == "" {
		workingDirectory = info.PrefixPath
	}
	lines := []string{
		"# " + managedServiceMarker,
		"[Unit]",
		"Description=" + n.systemdEscape(description),
		"After=network-online.target",
		"Wants=network-online.target",
		"",
		"[Service]",
		"Type=forking",
		"ExecStartPre=" + n.serviceCommand(info) + " -t",
		"ExecStart=" + n.serviceCommand(info),
		"ExecReload=" + n.serviceCommand(info) + " -s reload",
		"ExecStop=" + n.serviceCommand(info) + " -s quit",
		"PIDFile=" + n.systemdQuote(info.PidPath),
		"Restart=" + restart,
	}
	if workingDirectory != "" {
		lines = append(lines, "WorkingDirectory="+n.systemdQuote(workingDirectory))
	}
	if options.User != "" {
		lines = append(lines, "User="+n.systemdEscape(options.User))
	}
	if options.Group != "" {
		lines = append(lines, "Group="+n.systemdEscape(options.Group))
	}
	lines = append(lines, "", "[Install]", "WantedBy=multi-user.target", "")
	return strings.Join(lines, "\n")
}

func (n *Nginx) buildOpenRCService(info *Info, options ServiceOptions, name, description string) string {
	user := strings.TrimSpace(options.User)
	if user == "" {
		user = "root"
	}
	args := make([]string, 0, len(n.runtimeArgs(info)))
	for _, arg := range n.runtimeArgs(info) {
		args = append(args, n.shellQuote(arg))
	}
	return fmt.Sprintf(`#!/sbin/openrc-run
# %s
description=%q
command=%q
command_args=%q
command_background=true
pidfile=%q
command_user=%s
	`, managedServiceMarker, description, info.BinaryPath, strings.Join(args, " "), info.PidPath, user)
}

func (n *Nginx) buildRunitService(info *Info, options ServiceOptions) string {
	args := append([]string(nil), n.runtimeArgs(info)...)
	args = append(args, "-g", "daemon off;")
	command := n.shellCommand(info.BinaryPath, args)
	if user := strings.TrimSpace(options.User); user != "" && n.commandExists("chpst") {
		command = "chpst -u " + n.shellQuote(user) + " -- " + command
	}
	return "#!/bin/sh\n# " + managedServiceMarker + "\nexec " + command + "\n"
}

func (n *Nginx) buildProcdService(info *Info, options ServiceOptions) string {
	args := append([]string(nil), n.runtimeArgs(info)...)
	args = append(args, "-g", "daemon off;")
	command := n.shellCommand(info.BinaryPath, args)
	userLine := ""
	if user := strings.TrimSpace(options.User); user != "" {
		userLine = "\n\tprocd_set_param user " + n.shellQuote(user)
	}
	return "#!/bin/sh /etc/rc.common\n# " + managedServiceMarker + "\nSTART=99\nSTOP=10\nUSE_PROCD=1\n\nstart_service() {\n\tprocd_open_instance\n\tprocd_set_param command " + command + userLine + "\n\tprocd_set_param respawn\n\tprocd_close_instance\n}\n"
}

func (n *Nginx) buildSysVService(info *Info, options ServiceOptions, name, description string) string {
	start := n.shellCommand(info.BinaryPath, n.runtimeArgs(info))
	reload := n.shellCommand(info.BinaryPath, append(n.runtimeArgs(info), "-s", "reload"))
	stop := n.shellCommand(info.BinaryPath, append(n.runtimeArgs(info), "-s", "quit"))
	if strings.TrimSpace(info.PidPath) == "" {
		info.PidPath = n.defaultPidPath(info)
	}
	return fmt.Sprintf(`#!/bin/sh
# %s
### BEGIN INIT INFO
# Provides:          %s
# Required-Start:    $remote_fs $network
# Required-Stop:     $remote_fs $network
# Should-Start:      $named
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: %s
### END INIT INFO

PIDFILE=%s
case "$1" in
start)
    %s
    ;;
stop)
    %s
    ;;
reload)
    %s
    ;;
restart)
    %s
    %s
    ;;
status)
    if [ -s "$PIDFILE" ] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
        exit 0
    fi
    exit 3
    ;;
*)
    echo "Usage: $0 {start|stop|reload|restart|status}" >&2
    exit 2
    ;;
esac
exit 0
	`, managedServiceMarker, n.shellQuote(name), strings.ReplaceAll(description, "\n", " "), n.shellQuote(info.PidPath), start, stop, reload, stop, start)
}

func (n *Nginx) shellCommand(name string, args []string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, n.shellQuote(name))
	for _, arg := range args {
		parts = append(parts, n.shellQuote(arg))
	}
	return strings.Join(parts, " ")
}

func (n *Nginx) shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func (n *Nginx) managedServiceFile(path string) bool {
	data, err := os.ReadFile(path)
	return err == nil && bytes.Contains(data, []byte(managedServiceMarker))
}

func (n *Nginx) ensureManagedServicePath(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("服务定义不能是符号链接: %s", path)
	}
	if !n.managedServiceFile(path) {
		return fmt.Errorf("服务定义已存在且不是面板创建的定义: %s", path)
	}
	return nil
}

func (n *Nginx) systemdEscape(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\n", "\\n")
	return strings.ReplaceAll(value, "%", "%%")
}

func (n *Nginx) systemdQuote(value string) string {
	value = n.systemdEscape(value)
	if strings.ContainsAny(value, " \t\"") {
		return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
	}
	return value
}

func (n *Nginx) runService(action string) error {
	serviceName := n.getServiceName()
	if action == "quit" {
		action = "stop"
	}
	if runtime.GOOS == "windows" {
		if !n.commandExists("sc.exe") {
			return errors.New("未找到 Windows 服务管理器")
		}
		output, err := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "sc.exe", action, serviceName)
		if err != nil {
			return n.formatCommandOutput(output)
		}
		return nil
	}
	commands := make([]packageCommand, 0, 1)
	switch n.detectInitSystem() {
	case "systemd":
		commands = append(commands, packageCommand{"systemctl", []string{action, serviceName}})
	case "openrc":
		commands = append(commands, packageCommand{"rc-service", []string{serviceName, action}})
	case "procd":
		commands = append(commands, packageCommand{filepath.Join("/etc/init.d", serviceName), []string{action}})
	case "runit":
		commands = append(commands, packageCommand{"sv", []string{action, serviceName}})
	case "sysvinit":
		if n.commandExists("service") {
			commands = append(commands, packageCommand{"service", []string{serviceName, action}})
		} else {
			commands = append(commands, packageCommand{filepath.Join("/etc/init.d", serviceName), []string{action}})
		}
	}
	if len(commands) == 0 {
		return errors.New("未找到系统服务管理器")
	}
	var lastErr error
	for _, command := range commands {
		output, err := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", command.Name, command.Args...)
		if err == nil {
			return nil
		}
		lastErr = fmt.Errorf("%s: %w", command.Name, n.formatCommandOutput(output))
	}
	return lastErr
}

func (n *Nginx) getServiceName() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.serviceName == "" {
		return "nginx"
	}
	return n.serviceName
}

func (n *Nginx) readyWaitDuration() time.Duration {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.readyTimeout <= 0 {
		return 10 * time.Second
	}
	return n.readyTimeout
}

func (n *Nginx) waitReady() error {
	info, infoErr := n.Info()
	if infoErr != nil {
		return infoErr
	}
	deadline := time.Now().Add(n.readyWaitDuration())
	var lastErr error
	for time.Now().Before(deadline) {
		status, err := n.statusForInfo(info)
		if err != nil {
			lastErr = err
		} else if status.Running {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	if lastErr != nil {
		return fmt.Errorf("nginx 启动后未就绪: %w", lastErr)
	}
	return errors.New("nginx 启动后未就绪，请检查错误日志和端口占用")
}

func (n *Nginx) waitStopped() error {
	info, infoErr := n.Info()
	if infoErr != nil {
		return infoErr
	}
	deadline := time.Now().Add(n.readyWaitDuration())
	var lastErr error
	for time.Now().Before(deadline) {
		status, err := n.statusForInfo(info)
		if err != nil {
			lastErr = err
		} else if !status.Running {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	if lastErr != nil {
		return fmt.Errorf("nginx 停止后状态无法确认: %w", lastErr)
	}
	return errors.New("nginx 停止超时，请检查进程和错误日志")
}

// WaitReady 等待当前 Nginx 实例进入运行状态。
func (n *Nginx) WaitReady() error {
	return n.waitReady()
}

func (n *Nginx) inspectProcess(pid int) (*ProcessInfo, error) {
	if pid <= 0 {
		return nil, errors.New("PID 无效")
	}
	if runtime.GOOS == "windows" {
		output, err := n.runCommand(context.Background(), defaultTimeout, "tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/FO", "CSV", "/NH")
		if err != nil || !strings.Contains(output, `"`+strconv.Itoa(pid)+`"`) {
			return &ProcessInfo{PID: pid}, nil
		}
		return &ProcessInfo{PID: pid, Running: true}, nil
	}
	procRoot := filepath.Join("/proc", strconv.Itoa(pid))
	if data, err := os.ReadFile(filepath.Join(procRoot, "cmdline")); err == nil {
		commandLine := strings.ReplaceAll(string(data), "\x00", " ")
		executable, _ := os.Readlink(filepath.Join(procRoot, "exe"))
		return &ProcessInfo{PID: pid, Running: true, Executable: executable, CommandLine: strings.TrimSpace(commandLine), User: n.procUser(procRoot)}, nil
	}
	output, err := n.runCommand(context.Background(), defaultTimeout, "ps", "-p", strconv.Itoa(pid), "-o", "pid=,comm=,args=")
	if err != nil || strings.TrimSpace(output) == "" {
		return &ProcessInfo{PID: pid}, nil
	}
	fields := strings.Fields(output)
	if len(fields) < 2 {
		return &ProcessInfo{PID: pid, Running: true}, nil
	}
	commandLine := strings.Join(fields[1:], " ")
	return &ProcessInfo{PID: pid, Running: true, Executable: fields[1], CommandLine: commandLine}, nil
}

func (n *Nginx) procUser(procRoot string) string {
	data, err := os.ReadFile(filepath.Join(procRoot, "status"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "Uid:" {
			continue
		}
		account, lookupErr := user.LookupId(fields[1])
		if lookupErr == nil {
			return account.Username
		}
		return fields[1]
	}
	return ""
}

func (n *Nginx) processMatches(info *Info, process *ProcessInfo) bool {
	if process == nil || !process.Running {
		return false
	}
	if process.Executable != "" {
		expected, expectedErr := filepath.EvalSymlinks(info.BinaryPath)
		actual, actualErr := filepath.EvalSymlinks(process.Executable)
		if expectedErr == nil && actualErr == nil {
			if expected != actual {
				return false
			}
		} else if filepath.Base(info.BinaryPath) != filepath.Base(process.Executable) {
			return false
		}
	}
	if process.CommandLine == "" {
		return true
	}
	if configured := n.commandLineValue(process.CommandLine, "-c"); configured != "" && info.ConfigPath != "" {
		configured, _ = filepath.Abs(configured)
		expected, _ := filepath.Abs(info.ConfigPath)
		if filepath.Clean(configured) != filepath.Clean(expected) {
			return false
		}
	}
	n.mu.RLock()
	configExplicit, prefixExplicit := n.configExplicit, n.prefixExplicit
	n.mu.RUnlock()
	if configExplicit && n.commandLineValue(process.CommandLine, "-c") == "" {
		compiled := n.resolveBuildPath(n.parseBuildOption(info.Build, "conf-path"), info.PrefixPath)
		if compiled == "" || filepath.Clean(compiled) != filepath.Clean(info.ConfigPath) {
			return false
		}
	}
	if prefixExplicit && n.commandLineValue(process.CommandLine, "-p") == "" {
		compiled := n.resolveBuildPath(n.parseBuildOption(info.Build, "prefix"), "")
		if compiled == "" || filepath.Clean(compiled) != filepath.Clean(info.PrefixPath) {
			return false
		}
	}
	if configured := n.commandLineValue(process.CommandLine, "-p"); configured != "" && info.PrefixPath != "" {
		configured, _ = filepath.Abs(configured)
		expected, _ := filepath.Abs(info.PrefixPath)
		if filepath.Clean(configured) != filepath.Clean(expected) {
			return false
		}
	}
	return true
}

func (n *Nginx) commandLineValue(commandLine, flag string) string {
	fields := strings.Fields(commandLine)
	for index := range fields {
		if fields[index] == flag && index+1 < len(fields) {
			return strings.Trim(fields[index+1], `"'`)
		}
		if strings.HasPrefix(fields[index], flag+"=") {
			return strings.TrimPrefix(fields[index], flag+"=")
		}
	}
	return ""
}

// Install 安装 Nginx。
// 优先使用当前系统的软件包管理器；包管理器不可用或安装失败时，如果设置了
// NGINX_ARCHIVE_URL/NGINX_SOURCE_URL，则自动回退到压缩包或源码安装。
// 已存在 Nginx 时不会覆盖已有安装，也不会修改软件源。
func (n *Nginx) Install(ctx context.Context) error {
	n.installMu.Lock()
	defer n.installMu.Unlock()
	if _, err := n.binary(); err == nil {
		return nil
	}
	commands, managerErr := n.packageCommands("install")
	var packageErr error
	if managerErr == nil {
		for _, command := range commands {
			if !n.commandExists(command.Name) {
				continue
			}
			output, runErr := n.runPrivilegedCommand(ctx, installTimeout, "", command.Name, command.Args...)
			if runErr != nil {
				packageErr = fmt.Errorf("安装 nginx 失败: %w", n.formatCommandOutput(output))
				break
			}
		}
		if packageErr == nil {
			if _, err := n.binary(); err == nil {
				return nil
			}
		}
		if _, err := n.binary(); err == nil {
			return nil
		}
	}
	if sourceURL := strings.TrimSpace(os.Getenv("NGINX_ARCHIVE_URL")); sourceURL != "" {
		return n.installArchiveWithSHA256(ctx, sourceURL, n.installDirFromEnvironment(), "")
	}
	if sourceURL := strings.TrimSpace(os.Getenv("NGINX_SOURCE_URL")); sourceURL != "" {
		return n.installSourceWithSHA256(ctx, sourceURL, n.installDirFromEnvironment(), "")
	}
	if managerErr != nil {
		return fmt.Errorf("%w；请提供 NGINX_ARCHIVE_URL/NGINX_SOURCE_URL，或直接调用 InstallArchive/InstallSource", managerErr)
	}
	if packageErr != nil {
		return fmt.Errorf("%w；也可以提供 NGINX_ARCHIVE_URL/NGINX_SOURCE_URL 进行回退安装", packageErr)
	}
	return errors.New("安装命令执行完成，但未找到 nginx；请检查 PATH 或传入 nginx 可执行文件路径")
}

// InstallArchive 从 HTTP(S)、file:// 或本地文件安装 Nginx 压缩包。
// 支持 zip、tar、tar.gz 和 tgz。解压时会阻止路径穿越，并自动查找 nginx 可执行文件。
func (n *Nginx) InstallArchive(ctx context.Context, sourceURL, installDir string) error {
	return n.InstallArchiveWithSHA256(ctx, sourceURL, installDir, "")
}

// InstallArchiveWithSHA256 校验 SHA-256 后从二进制压缩包安装 Nginx。
func (n *Nginx) InstallArchiveWithSHA256(ctx context.Context, sourceURL, installDir, expectedSHA256 string) error {
	n.installMu.Lock()
	defer n.installMu.Unlock()
	return n.installArchiveWithSHA256(ctx, sourceURL, installDir, expectedSHA256)
}

func (n *Nginx) installArchiveWithSHA256(ctx context.Context, sourceURL, installDir, expectedSHA256 string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	sourceURL = strings.TrimSpace(sourceURL)
	if sourceURL == "" {
		return errors.New("Nginx 安装包地址不能为空")
	}
	installDir = strings.TrimSpace(installDir)
	if installDir == "" {
		installDir = n.defaultInstallDir()
	}
	installDir, err := filepath.Abs(installDir)
	if err != nil {
		return fmt.Errorf("解析 Nginx 安装目录失败: %w", err)
	}
	cleanupInstallDir := n.emptyOrMissingDirectory(installDir)
	completed := false
	defer func() {
		if cleanupInstallDir && !completed {
			_ = os.RemoveAll(installDir)
		}
	}()
	if err = n.ensureInstallDirectory(installDir); err != nil {
		return err
	}
	if err = n.ensureInstallDirectoryEmpty(installDir); err != nil {
		return err
	}
	if err = os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("创建 Nginx 安装目录失败: %w", err)
	}
	temp, err := os.CreateTemp("", ".nginx-download-*")
	if err != nil {
		return fmt.Errorf("创建 Nginx 安装临时文件失败: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err = n.downloadArchiveWithSHA256(ctx, sourceURL, temp, expectedSHA256); err != nil {
		_ = temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return fmt.Errorf("关闭 Nginx 安装包失败: %w", err)
	}
	format, err := n.detectArchiveFormat(sourceURL, tempPath)
	if err != nil {
		return err
	}
	if err = n.extractArchive(tempPath, installDir, format); err != nil {
		return fmt.Errorf("解压 Nginx 安装包失败: %w", err)
	}
	binaryPath, err := n.findNginxBinary(installDir)
	if err != nil {
		return err
	}
	if err = n.configureArchiveInstall(binaryPath, installDir); err != nil {
		return err
	}
	if _, err = n.Info(); err != nil {
		return fmt.Errorf("读取解压后的 nginx 信息失败: %w", err)
	}
	completed = true
	return nil
}

// InstallSource 下载并编译 Nginx 源码包。
// 会先检测编译器、make、PCRE、zlib 和 OpenSSL，缺少时通过当前系统的
// 包管理器自动补齐；没有可用包管理器时返回明确错误。Windows 应使用 InstallArchive。
func (n *Nginx) InstallSource(ctx context.Context, sourceURL, installDir string) error {
	return n.InstallSourceWithSHA256(ctx, sourceURL, installDir, "")
}

// EnsureBuildEnvironment 检查并自动补齐源码编译 Nginx 所需的工具链和开发库。
// 只会在调用该方法或 InstallSource 时执行包管理器，不会修改软件源配置。
func (n *Nginx) EnsureBuildEnvironment(ctx context.Context) error {
	n.installMu.Lock()
	defer n.installMu.Unlock()
	return n.ensureBuildEnvironment(ctx)
}

// InstallSourceWithSHA256 校验 SHA-256 后下载并编译 Nginx 源码。
func (n *Nginx) InstallSourceWithSHA256(ctx context.Context, sourceURL, installDir, expectedSHA256 string) error {
	n.installMu.Lock()
	defer n.installMu.Unlock()
	return n.installSourceWithSHA256(ctx, sourceURL, installDir, expectedSHA256)
}

func (n *Nginx) installSourceWithSHA256(ctx context.Context, sourceURL, installDir, expectedSHA256 string) error {
	if runtime.GOOS == "windows" {
		return errors.New("Windows 不支持源码编译安装，请使用 InstallArchive")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	sourceURL = strings.TrimSpace(sourceURL)
	if sourceURL == "" {
		return errors.New("Nginx 源码地址不能为空")
	}
	installDir = strings.TrimSpace(installDir)
	if installDir == "" {
		installDir = n.defaultInstallDir()
	}
	installDir, err := filepath.Abs(installDir)
	if err != nil {
		return fmt.Errorf("解析 Nginx 安装目录失败: %w", err)
	}
	cleanupInstallDir := n.emptyOrMissingDirectory(installDir)
	completed := false
	defer func() {
		if cleanupInstallDir && !completed {
			_ = os.RemoveAll(installDir)
		}
	}()
	if err = n.ensureInstallDirectory(installDir); err != nil {
		return err
	}
	if err = n.ensureInstallDirectoryEmpty(installDir); err != nil {
		return err
	}
	if err = n.ensureBuildEnvironment(ctx); err != nil {
		return err
	}
	workDir, err := os.MkdirTemp("", ".nginx-build-*")
	if err != nil {
		return fmt.Errorf("创建 Nginx 编译目录失败: %w", err)
	}
	defer os.RemoveAll(workDir)
	archivePath := filepath.Join(workDir, "nginx-source")
	archive, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("创建 Nginx 源码临时文件失败: %w", err)
	}
	if err = n.downloadArchiveWithSHA256(ctx, sourceURL, archive, expectedSHA256); err != nil {
		_ = archive.Close()
		return err
	}
	if err = archive.Close(); err != nil {
		return fmt.Errorf("关闭 Nginx 源码包失败: %w", err)
	}
	format, err := n.detectArchiveFormat(sourceURL, archivePath)
	if err != nil {
		return err
	}
	if format == "zip" {
		return errors.New("Nginx 源码安装只支持 tar、tar.gz 或 tgz")
	}
	sourceDir := filepath.Join(workDir, "source")
	if err = os.MkdirAll(sourceDir, 0755); err != nil {
		return err
	}
	if err = n.extractArchive(archivePath, sourceDir, format); err != nil {
		return fmt.Errorf("解压 Nginx 源码失败: %w", err)
	}
	configurePath, err := n.findBuildScript(sourceDir)
	if err != nil {
		return err
	}
	buildDir := filepath.Dir(configurePath)
	if !n.commandExists("make") && !n.commandExists("gmake") {
		return errors.New("未找到 make/gmake，请先安装编译工具链")
	}
	makeCommand := "make"
	if !n.commandExists("make") && n.commandExists("gmake") {
		makeCommand = "gmake"
	}
	if runtime.GOOS == "freebsd" || runtime.GOOS == "openbsd" || runtime.GOOS == "netbsd" || runtime.GOOS == "dragonfly" {
		if n.commandExists("gmake") {
			makeCommand = "gmake"
		}
	}
	if !n.commandExists("cc") && !n.commandExists("gcc") && !n.commandExists("clang") {
		return errors.New("未找到 C 编译器，请先安装 gcc/clang")
	}
	configureArgs := []string{
		"--prefix=" + installDir,
		"--sbin-path=" + filepath.Join(installDir, "sbin", "nginx"),
		"--conf-path=" + filepath.Join(installDir, "conf", "nginx.conf"),
		"--pid-path=" + filepath.Join(installDir, "logs", "nginx.pid"),
		"--error-log-path=" + filepath.Join(installDir, "logs", "error.log"),
		"--http-log-path=" + filepath.Join(installDir, "logs", "access.log"),
		"--with-http_ssl_module",
		"--with-http_v2_module",
		"--with-http_realip_module",
		"--with-stream",
		"--with-stream_ssl_module",
	}
	configureName := configurePath
	if info, statErr := os.Stat(configurePath); statErr != nil || info.Mode().Perm()&0111 == 0 {
		configureName = "/bin/sh"
		configureArgs = append([]string{configurePath}, configureArgs...)
	}
	if output, runErr := n.runCommandDir(ctx, installTimeout, buildDir, configureName, configureArgs...); runErr != nil {
		return fmt.Errorf("配置 Nginx 源码失败: %w", n.formatCommandOutput(output))
	}
	jobs := n.buildJobs()
	if output, runErr := n.runCommandDir(ctx, installTimeout, buildDir, makeCommand, "-j", strconv.Itoa(jobs)); runErr != nil {
		return fmt.Errorf("编译 Nginx 失败: %w", n.formatCommandOutput(output))
	}
	if output, runErr := n.runInstallStep(ctx, installTimeout, buildDir, installDir, makeCommand, "install"); runErr != nil {
		return fmt.Errorf("安装编译后的 Nginx 失败: %w", n.formatCommandOutput(output))
	}
	binaryPath, err := n.findNginxBinary(installDir)
	if err != nil {
		return err
	}
	if err = n.configureArchiveInstall(binaryPath, installDir); err != nil {
		return err
	}
	_, err = n.Info()
	if err == nil {
		completed = true
	}
	return err
}

func (n *Nginx) findBuildScript(root string) (string, error) {
	rootConfigure := filepath.Join(root, "configure")
	if info, err := os.Stat(rootConfigure); err == nil && info.Mode().IsRegular() {
		return rootConfigure, nil
	}
	var found string
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.Mode().IsRegular() && info.Name() == "configure" {
			found = path
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("查找 Nginx configure 脚本失败: %w", err)
	}
	if found == "" {
		return "", errors.New("Nginx 源码包中未找到 configure 脚本")
	}
	return found, nil
}

func (n *Nginx) buildJobs() int {
	jobs := runtime.NumCPU()
	if value := strings.TrimSpace(os.Getenv("NGINX_BUILD_JOBS")); value != "" {
		if configured, err := strconv.Atoi(value); err == nil && configured > 0 {
			jobs = configured
		}
	}
	if jobs < 1 {
		jobs = 1
	}
	if jobs > 8 {
		jobs = 8
	}
	return jobs
}

func (n *Nginx) ensureBuildEnvironment(ctx context.Context) error {
	missing := n.missingBuildDependencies()
	if len(missing) == 0 {
		return nil
	}
	commands, err := n.buildDependencyCommands()
	if err != nil {
		return fmt.Errorf("源码编译缺少环境(%s): %w", strings.Join(missing, ", "), err)
	}
	for _, command := range commands {
		if !n.commandExists(command.Name) {
			continue
		}
		output, runErr := n.runPrivilegedCommand(ctx, installTimeout, "", command.Name, command.Args...)
		if runErr != nil {
			return fmt.Errorf("安装源码编译依赖失败: %w", n.formatCommandOutput(output))
		}
	}
	n.refreshUserProfilePath()
	n.refreshBuildLibraryPath()
	if missing = n.missingBuildDependencies(); len(missing) > 0 {
		return fmt.Errorf("自动补齐依赖后仍缺少: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (n *Nginx) missingBuildDependencies() []string {
	missing := make([]string, 0, 5)
	if !n.commandExists("cc") && !n.commandExists("gcc") && !n.commandExists("clang") {
		missing = append(missing, "C编译器")
	}
	if !n.commandExists("make") && !n.commandExists("gmake") {
		missing = append(missing, "make")
	}
	if !n.hasBuildLibrary("pcre2", "pcre2-config", "pcre-config", "pcre2.h", "pcre.h") {
		missing = append(missing, "PCRE")
	}
	if !n.hasBuildLibrary("zlib", "", "", "zlib.h") {
		missing = append(missing, "zlib")
	}
	if !n.hasBuildLibrary("openssl", "", "", "openssl/ssl.h") {
		missing = append(missing, "OpenSSL")
	}
	return missing
}

func (n *Nginx) hasBuildLibrary(pkgConfigName, configCommand, alternateConfigCommand string, headers ...string) bool {
	if configCommand != "" && n.commandExists(configCommand) {
		return true
	}
	if alternateConfigCommand != "" && n.commandExists(alternateConfigCommand) {
		return true
	}
	if n.commandExists("pkg-config") {
		if _, err := n.runCommand(context.Background(), defaultTimeout, "pkg-config", "--exists", pkgConfigName); err == nil {
			return true
		}
	}
	roots := []string{
		"/usr/include",
		"/usr/local/include",
		"/usr/local/opt/" + pkgConfigName + "/include",
		"/opt/homebrew/include",
		"/opt/homebrew/opt/" + pkgConfigName + "/include",
		"/opt/homebrew/opt/openssl@3/include",
		"/opt/local/include",
	}
	for _, key := range []string{"CPATH", "C_INCLUDE_PATH"} {
		for _, root := range filepath.SplitList(os.Getenv(key)) {
			if root != "" {
				roots = append(roots, root)
			}
		}
	}
	if n.commandExists("brew") {
		for _, formula := range []string{pkgConfigName, "openssl@3"} {
			if output, err := n.runCommand(context.Background(), defaultTimeout, "brew", "--prefix", formula); err == nil && strings.TrimSpace(output) != "" {
				roots = append(roots, filepath.Join(strings.TrimSpace(output), "include"))
			}
		}
	}
	for _, header := range headers {
		for _, root := range roots {
			if _, err := os.Stat(filepath.Join(root, header)); err == nil {
				return true
			}
		}
	}
	return false
}

func (n *Nginx) buildDependencyCommands() ([]packageCommand, error) {
	manager := n.detectPackageManager()
	release := n.ReadOSRelease()
	debian := []string{"build-essential", "libpcre3-dev", "zlib1g-dev", "libssl-dev", "pkg-config"}
	rpmPCRE := "pcre2-devel"
	if release.ID == "centos" || release.ID == "rhel" || release.ID == "ol" || release.ID == "oracle" || release.ID == "amzn" || release.ID == "amazon" {
		if major, parseErr := strconv.Atoi(strings.SplitN(release.VersionID, ".", 2)[0]); parseErr == nil && major > 0 && major < 8 {
			rpmPCRE = "pcre-devel"
		}
	}
	rpm := []string{"gcc", "make", rpmPCRE, "zlib-devel", "openssl-devel", "pkgconfig"}
	alpinePCRE := "pcre2-dev"
	if release.ID == "alpine" {
		if major, parseErr := strconv.Atoi(strings.SplitN(release.VersionID, ".", 2)[0]); parseErr == nil && major > 0 && major < 3 {
			alpinePCRE = "pcre-dev"
		}
	}
	alpine := []string{"build-base", alpinePCRE, "zlib-dev", "openssl-dev", "pkgconf", "linux-headers"}
	pacman := []string{"gcc", "make", "pcre2", "zlib", "openssl", "pkgconf"}
	bsd := []string{"gcc", "gmake", "pcre2", "zlib", "openssl", "pkgconf"}
	switch manager {
	case "apt-get":
		return []packageCommand{{"apt-get", []string{"-o", "DPkg::Lock::Timeout=60", "update"}}, {"apt-get", append([]string{"-o", "DPkg::Lock::Timeout=60", "install", "-y", "--no-install-recommends"}, debian...)}}, nil
	case "apt":
		return []packageCommand{{"apt", []string{"-o", "DPkg::Lock::Timeout=60", "update"}}, {"apt", append([]string{"-o", "DPkg::Lock::Timeout=60", "install", "-y", "--no-install-recommends"}, debian...)}}, nil
	case "nala":
		return []packageCommand{{"nala", append([]string{"install", "-y"}, debian...)}}, nil
	case "aptitude":
		return []packageCommand{{"aptitude", append([]string{"-y", "install"}, debian...)}}, nil
	case "dnf", "dnf5", "yum", "microdnf", "tdnf":
		return []packageCommand{{manager, append([]string{"install", "-y"}, rpm...)}}, nil
	case "apk":
		return []packageCommand{{"apk", append([]string{"add", "--no-cache"}, alpine...)}}, nil
	case "zypper":
		return []packageCommand{{"zypper", []string{"--non-interactive", "install", "gcc", "make", "pcre2-devel", "zlib-devel", "libopenssl-devel", "pkg-config"}}}, nil
	case "pacman":
		return []packageCommand{{"pacman", append([]string{"-S", "--needed", "--noconfirm"}, pacman...)}}, nil
	case "emerge":
		return []packageCommand{{"emerge", []string{"--quiet-build", "sys-devel/gcc", "sys-devel/make", "dev-libs/pcre2", "sys-libs/zlib", "dev-libs/openssl", "dev-util/pkgconf"}}}, nil
	case "nix-env":
		return []packageCommand{{"nix-env", []string{"-iA", "nixpkgs.gcc", "nixpkgs.gnumake", "nixpkgs.pcre2", "nixpkgs.zlib", "nixpkgs.openssl", "nixpkgs.pkg-config"}}}, nil
	case "guix":
		return []packageCommand{{"guix", []string{"install", "gcc-toolchain", "make", "pcre2", "zlib", "openssl", "pkg-config"}}}, nil
	case "eopkg":
		return []packageCommand{{"eopkg", []string{"install", "-y", "gcc", "make", "pcre2", "zlib", "openssl", "pkg-config"}}}, nil
	case "urpmi":
		return []packageCommand{{"urpmi", []string{"--auto", "gcc", "make", "pcre-devel", "zlib-devel", "openssl-devel", "pkgconfig"}}}, nil
	case "slackpkg":
		return []packageCommand{{"slackpkg", []string{"install", "gcc", "make", "pcre2", "zlib", "openssl", "pkg-config"}}}, nil
	case "xbps-install":
		return []packageCommand{{"xbps-install", []string{"-Sy", "gcc", "make", "pcre2-devel", "zlib-devel", "openssl-devel", "pkg-config"}}}, nil
	case "opkg":
		return []packageCommand{{"opkg", []string{"update"}}, {"opkg", []string{"install", "gcc", "make", "pcre2-dev", "zlib-dev", "openssl-dev", "pkgconfig"}}}, nil
	case "brew":
		return []packageCommand{{"brew", []string{"install", "pcre2", "zlib", "openssl@3", "pkg-config"}}}, nil
	case "pkg":
		return []packageCommand{{"pkg", append([]string{"install", "-y"}, bsd...)}}, nil
	case "pkg_add":
		return []packageCommand{{"pkg_add", append([]string{"-I"}, bsd...)}}, nil
	case "pkgin":
		return []packageCommand{{"pkgin", append([]string{"-y", "install"}, bsd...)}}, nil
	default:
		return nil, errors.New("未识别可用的编译依赖包管理器")
	}
}

func (n *Nginx) refreshUserProfilePath() {
	paths := make([]string, 0, 2)
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".nix-profile", "bin"), filepath.Join(home, ".guix-profile", "bin"))
	}
	current := os.Getenv("PATH")
	for i := len(paths) - 1; i >= 0; i-- {
		if _, err := os.Stat(paths[i]); err == nil && !strings.Contains(string(os.PathListSeparator)+current+string(os.PathListSeparator), string(os.PathListSeparator)+paths[i]+string(os.PathListSeparator)) {
			current = paths[i] + string(os.PathListSeparator) + current
		}
	}
	if current != os.Getenv("PATH") {
		_ = os.Setenv("PATH", current)
	}
}

func (n *Nginx) refreshBuildLibraryPath() {
	if !n.commandExists("brew") {
		return
	}
	for _, formula := range []string{"pcre2", "zlib", "openssl@3"} {
		output, err := n.runCommand(context.Background(), defaultTimeout, "brew", "--prefix", formula)
		if err != nil || strings.TrimSpace(output) == "" {
			continue
		}
		prefix := strings.TrimSpace(output)
		n.prependEnvironmentPath("CPATH", filepath.Join(prefix, "include"))
		n.prependEnvironmentPath("LIBRARY_PATH", filepath.Join(prefix, "lib"))
		n.prependEnvironmentPath("PKG_CONFIG_PATH", filepath.Join(prefix, "lib", "pkgconfig"))
	}
}

func (n *Nginx) prependEnvironmentPath(key, value string) {
	if value == "" {
		return
	}
	current := os.Getenv(key)
	for _, item := range filepath.SplitList(current) {
		if item == value {
			return
		}
	}
	if current == "" {
		_ = os.Setenv(key, value)
		return
	}
	_ = os.Setenv(key, value+string(os.PathListSeparator)+current)
}

func (n *Nginx) defaultInstallDir() string {
	if runtime.GOOS == "windows" {
		for _, key := range []string{"ProgramData", "ProgramFiles"} {
			if value := strings.TrimSpace(os.Getenv(key)); value != "" {
				return filepath.Join(value, "nginx")
			}
		}
		return filepath.Join(os.TempDir(), "nginx")
	}
	if !n.isRootUser() {
		if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
			return filepath.Join(home, ".local", "opt", "nginx")
		}
		return filepath.Join(os.TempDir(), "nginx")
	}
	return "/opt/nginx"
}

func (n *Nginx) ensureInstallDirectory(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("Nginx 安装目录不能为空")
	}
	path, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("解析 Nginx 安装目录失败: %w", err)
	}
	if filepath.Dir(path) == path {
		return fmt.Errorf("不能把文件系统根目录作为 Nginx 安装目录: %s", path)
	}
	if runtime.GOOS == "linux" {
		switch path {
		case "/usr", "/usr/local", "/etc", "/var", "/opt", "/home", "/root", "/tmp":
			return fmt.Errorf("不能把系统公共目录作为 Nginx 安装目录: %s", path)
		}
	}
	if info, statErr := os.Lstat(path); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("Nginx 安装目录不能是符号链接: %s", path)
	}
	if n.isRootUser() || n.writableDirectory(path) {
		return nil
	}
	return fmt.Errorf("Nginx 安装目录不可写: %s；请使用 root/sudo 或指定当前用户可写目录", path)
}

func (n *Nginx) emptyOrMissingDirectory(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return true
	}
	if err != nil || !info.IsDir() {
		return false
	}
	entries, readErr := os.ReadDir(path)
	return readErr == nil && len(entries) == 0
}

func (n *Nginx) ensureInstallDirectoryEmpty(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("检查 Nginx 安装目录失败: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("Nginx 安装路径不是目录: %s", path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("读取 Nginx 安装目录失败: %w", err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("Nginx 安装目录必须为空: %s；请指定新的目录或先显式卸载", path)
	}
	return nil
}

func (n *Nginx) installDirFromEnvironment() string {
	if value := strings.TrimSpace(os.Getenv("NGINX_INSTALL_DIR")); value != "" {
		return value
	}
	return n.defaultInstallDir()
}

func (n *Nginx) downloadArchive(ctx context.Context, sourceURL string, destination io.Writer) error {
	return n.downloadArchiveWithSHA256(ctx, sourceURL, destination, "")
}

func (n *Nginx) downloadArchiveWithSHA256(ctx context.Context, sourceURL string, destination io.Writer, expectedSHA256 string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	expectedSHA256 = strings.ToLower(strings.TrimSpace(expectedSHA256))
	if expectedSHA256 != "" && (len(expectedSHA256) != sha256.Size*2 || strings.Trim(expectedSHA256, "0123456789abcdef") != "") {
		return errors.New("期望的 SHA-256 格式不合法")
	}
	digest := sha256.New()
	output := io.MultiWriter(destination, digest)
	parsed, parseErr := url.Parse(sourceURL)
	if parseErr == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
		if err != nil {
			return fmt.Errorf("创建 Nginx 下载请求失败: %w", err)
		}
		response, err := (&http.Client{Timeout: installTimeout}).Do(request)
		if err != nil {
			return fmt.Errorf("下载 Nginx 安装包失败: %w", err)
		}
		defer response.Body.Close()
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return fmt.Errorf("下载 Nginx 安装包失败: HTTP %s", response.Status)
		}
		written, err := io.Copy(output, io.LimitReader(response.Body, maxInstallArchiveSize+1))
		if err != nil {
			return fmt.Errorf("保存 Nginx 安装包失败: %w", err)
		}
		if written > maxInstallArchiveSize {
			return fmt.Errorf("Nginx 安装包超过 %d MB 限制", maxInstallArchiveSize/(1<<20))
		}
		return n.verifyDigest(expectedSHA256, digest.Sum(nil))
	}

	path := sourceURL
	if parseErr == nil && parsed.Scheme == "file" {
		path, parseErr = url.PathUnescape(parsed.Path)
		if parseErr != nil {
			return fmt.Errorf("解析 Nginx 本地安装包路径失败: %w", parseErr)
		}
		if runtime.GOOS == "windows" {
			if parsed.Host != "" {
				path = `\\` + parsed.Host + path
			} else {
				path = strings.TrimPrefix(path, "/")
			}
		}
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("打开 Nginx 本地安装包失败: %w", err)
	}
	defer file.Close()
	if info, statErr := file.Stat(); statErr == nil && info.Size() > maxInstallArchiveSize {
		return fmt.Errorf("Nginx 安装包超过 %d MB 限制", maxInstallArchiveSize/(1<<20))
	}
	written, err := io.Copy(output, io.LimitReader(file, maxInstallArchiveSize+1))
	if err != nil {
		return fmt.Errorf("读取 Nginx 本地安装包失败: %w", err)
	}
	if written > maxInstallArchiveSize {
		return fmt.Errorf("Nginx 安装包超过 %d MB 限制", maxInstallArchiveSize/(1<<20))
	}
	return n.verifyDigest(expectedSHA256, digest.Sum(nil))
}

// DownloadVerified 下载文件并校验 SHA-256，适合业务层保存安装包或证书。
func (n *Nginx) DownloadVerified(ctx context.Context, sourceURL, destination, expectedSHA256 string) error {
	if strings.TrimSpace(destination) == "" {
		return errors.New("下载目标不能为空")
	}
	destination, err := filepath.Abs(destination)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(destination), ".nginx-download-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err = n.downloadArchiveWithSHA256(ctx, sourceURL, temp, expectedSHA256); err != nil {
		_ = temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		if err = os.Remove(destination); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return os.Rename(tempPath, destination)
}

func (n *Nginx) verifyDigest(expected string, digest []byte) error {
	if expected == "" {
		return nil
	}
	actual := hex.EncodeToString(digest)
	if actual != expected {
		return fmt.Errorf("Nginx 安装包 SHA-256 校验失败，实际值: %s", actual)
	}
	return nil
}

func (n *Nginx) detectArchiveFormat(sourceURL, archivePath string) (string, error) {
	name := sourceURL
	if parsed, err := url.Parse(sourceURL); err == nil && parsed.Path != "" {
		name = parsed.Path
	}
	name = strings.ToLower(name)
	switch {
	case strings.HasSuffix(name, ".zip"):
		return "zip", nil
	case strings.HasSuffix(name, ".tar.gz"), strings.HasSuffix(name, ".tgz"):
		return "tgz", nil
	case strings.HasSuffix(name, ".tar"):
		return "tar", nil
	}
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	header := make([]byte, 4)
	if _, err = io.ReadFull(file, header); err != nil {
		return "", errors.New("无法识别 Nginx 安装包格式")
	}
	if header[0] == 'P' && header[1] == 'K' {
		return "zip", nil
	}
	if header[0] == 0x1f && header[1] == 0x8b {
		return "tgz", nil
	}
	return "tar", nil
}

func (n *Nginx) extractArchive(archivePath, installDir, format string) error {
	switch format {
	case "zip":
		return n.extractZipArchive(archivePath, installDir)
	case "tar":
		return n.extractTarArchive(archivePath, installDir, false)
	case "tgz":
		return n.extractTarArchive(archivePath, installDir, true)
	default:
		return errors.New("不支持的 Nginx 安装包格式")
	}
}

func (n *Nginx) extractZipArchive(archivePath, installDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()
	var extracted int64
	for _, entry := range reader.File {
		if entry.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("Nginx 安装包包含不支持的符号链接: %s", entry.Name)
		}
		target, err := n.safeArchivePath(installDir, entry.Name)
		if err != nil {
			return err
		}
		if entry.FileInfo().IsDir() {
			if err = n.ensureArchiveParents(installDir, target); err != nil {
				return err
			}
			if err = os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}
		if entry.UncompressedSize64 > uint64(maxInstallArchiveSize-extracted) {
			return errors.New("Nginx 安装包解压后超过大小限制")
		}
		if err = n.ensureArchiveParents(installDir, target); err != nil {
			return err
		}
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		if existing, statErr := os.Lstat(target); statErr == nil && existing.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("Nginx 安装目录包含不支持的符号链接: %s", target)
		}
		input, err := entry.Open()
		if err != nil {
			return err
		}
		mode := entry.Mode().Perm()
		if mode == 0 {
			mode = 0644
		}
		if strings.EqualFold(filepath.Base(entry.Name), "nginx") {
			mode = 0755
		}
		output, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
		if err == nil {
			written, copyErr := io.Copy(output, io.LimitReader(input, maxInstallArchiveSize-extracted+1))
			extracted += written
			if closeErr := output.Close(); copyErr == nil {
				copyErr = closeErr
			}
			err = copyErr
		}
		_ = input.Close()
		if err != nil {
			return err
		}
		if extracted > maxInstallArchiveSize {
			return errors.New("Nginx 安装包解压后超过大小限制")
		}
	}
	return nil
}

func (n *Nginx) extractTarArchive(archivePath, installDir string, compressed bool) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()
	var input io.Reader = file
	var gzipReader *gzip.Reader
	if compressed {
		gzipReader, err = gzip.NewReader(file)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		input = gzipReader
	}
	reader := tar.NewReader(input)
	var extracted int64
	for {
		header, readErr := reader.Next()
		if errors.Is(readErr, io.EOF) {
			return nil
		}
		if readErr != nil {
			return readErr
		}
		if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeLink {
			return fmt.Errorf("Nginx 安装包包含不支持的链接: %s", header.Name)
		}
		target, pathErr := n.safeArchivePath(installDir, header.Name)
		if pathErr != nil {
			return pathErr
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err = n.ensureArchiveParents(installDir, target); err != nil {
				return err
			}
			if err = os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if header.Size < 0 || header.Size > maxInstallArchiveSize-extracted {
				return errors.New("Nginx 安装包解压后超过大小限制")
			}
			if err = n.ensureArchiveParents(installDir, target); err != nil {
				return err
			}
			if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if existing, statErr := os.Lstat(target); statErr == nil && existing.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("Nginx 安装目录包含不支持的符号链接: %s", target)
			}
			mode := os.FileMode(header.Mode).Perm()
			if mode == 0 {
				mode = 0644
			}
			output, openErr := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if openErr != nil {
				return openErr
			}
			written, copyErr := io.Copy(output, io.LimitReader(reader, maxInstallArchiveSize-extracted+1))
			extracted += written
			if closeErr := output.Close(); copyErr == nil {
				copyErr = closeErr
			}
			if copyErr != nil {
				return copyErr
			}
		default:
			return fmt.Errorf("Nginx 安装包包含不支持的文件类型: %s", header.Name)
		}
	}
}

func (n *Nginx) safeArchivePath(root, name string) (string, error) {
	name = strings.TrimSpace(strings.ReplaceAll(name, "\\", "/"))
	if name == "" || name == "." {
		return root, nil
	}
	if strings.ContainsRune(name, 0) || filepath.IsAbs(name) || (len(name) >= 2 && name[1] == ':') {
		return "", fmt.Errorf("Nginx 安装包包含非法绝对路径: %s", name)
	}
	relative := filepath.Clean(filepath.FromSlash(name))
	if relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("Nginx 安装包包含非法路径: %s", name)
	}
	target := filepath.Join(root, relative)
	actual, err := filepath.Rel(root, target)
	if err != nil || actual == ".." || strings.HasPrefix(actual, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("Nginx 安装包包含非法路径: %s", name)
	}
	return target, nil
}

func (n *Nginx) ensureArchiveParents(root, target string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(root, filepath.Dir(target))
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return errors.New("Nginx 安装包路径超出安装目录")
	}
	current := root
	if info, statErr := os.Lstat(current); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("Nginx 安装目录不能是符号链接")
	}
	if relative == "." {
		return nil
	}
	for _, part := range strings.Split(relative, string(os.PathSeparator)) {
		current = filepath.Join(current, part)
		info, statErr := os.Lstat(current)
		if os.IsNotExist(statErr) {
			continue
		}
		if statErr != nil {
			return statErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("Nginx 安装目录包含符号链接目录: %s", current)
		}
		if !info.IsDir() {
			return fmt.Errorf("Nginx 安装目录路径不是目录: %s", current)
		}
	}
	return nil
}

func (n *Nginx) findNginxBinary(root string) (string, error) {
	var found string
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			return nil
		}
		name := strings.ToLower(info.Name())
		if name != "nginx" && name != "nginx.exe" {
			return nil
		}
		if found == "" || strings.Contains(strings.ToLower(path), string(filepath.Separator)+"sbin"+string(filepath.Separator)) {
			found = path
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("查找 nginx 可执行文件失败: %w", err)
	}
	if found == "" {
		return "", errors.New("Nginx 安装包中未找到 nginx 可执行文件")
	}
	return found, nil
}

func (n *Nginx) configureArchiveInstall(binaryPath, installDir string) error {
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return fmt.Errorf("设置 nginx 可执行权限失败: %w", err)
	}
	root := installDir
	binaryDir := filepath.Dir(binaryPath)
	if strings.EqualFold(filepath.Base(binaryDir), "sbin") {
		root = filepath.Dir(binaryDir)
	}
	if err := os.MkdirAll(filepath.Join(root, "logs"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root, "conf.d"), 0755); err != nil {
		return err
	}
	n.mu.RLock()
	configPath := n.configPath
	n.mu.RUnlock()
	if configPath == "" {
		for _, candidate := range []string{
			filepath.Join(root, "conf", "nginx.conf"),
			filepath.Join(root, "nginx.conf"),
			filepath.Join(binaryDir, "conf", "nginx.conf"),
		} {
			if _, err := os.Stat(candidate); err == nil {
				configPath = candidate
				break
			}
		}
	}
	if configPath == "" {
		configPath = filepath.Join(root, "conf", "nginx.conf")
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return err
		}
		config := "worker_processes auto;\nerror_log logs/error.log;\npid logs/nginx.pid;\n\nevents {}\n\nhttp {\n    default_type application/octet-stream;\n    sendfile on;\n    include conf.d/*.conf;\n}\n"
		if err := n.atomicWrite(configPath, []byte(config), 0644); err != nil {
			return err
		}
	}
	n.mu.Lock()
	n.binaryPath = binaryPath
	if n.prefixPath == "" {
		n.prefixPath = root
	}
	if n.configPath == "" {
		n.configPath = configPath
	}
	if n.pidPath == "" {
		n.pidPath = filepath.Join(root, "logs", "nginx.pid")
	}
	n.mu.Unlock()
	return nil
}

// Uninstall 使用当前系统的软件包管理器卸载 Nginx，不删除站点配置和证书文件。
func (n *Nginx) Uninstall(ctx context.Context) error {
	n.installMu.Lock()
	defer n.installMu.Unlock()
	info, infoErr := n.Info()
	if infoErr != nil {
		return infoErr
	}
	if !n.packageOwnsBinary(info.BinaryPath) {
		return errors.New("当前 Nginx 不是可确认的软件包安装，已拒绝自动卸载；归档或源码安装请手动移除对应目录")
	}
	if status, err := n.Status(); err == nil && status.Running {
		if err = n.Stop(true); err != nil {
			return err
		}
	}
	commands, err := n.packageCommands("uninstall")
	if err != nil {
		return err
	}
	for _, command := range commands {
		if !n.commandExists(command.Name) {
			continue
		}
		output, runErr := n.runPrivilegedCommand(ctx, installTimeout, "", command.Name, command.Args...)
		if runErr != nil {
			return fmt.Errorf("卸载 nginx 失败: %w", n.formatCommandOutput(output))
		}
		return nil
	}
	return errors.New("未找到可用的软件包管理器")
}

func (n *Nginx) packageOwnsBinary(path string) bool {
	path, err := filepath.Abs(path)
	if err != nil || path == "" {
		return false
	}
	manager := n.detectPackageManager()
	commands := make([]packageCommand, 0, 1)
	switch manager {
	case "apt-get", "apt", "nala", "aptitude":
		commands = append(commands, packageCommand{"dpkg-query", []string{"-S", path}})
	case "dnf", "dnf5", "yum", "microdnf", "tdnf", "zypper", "urpmi":
		commands = append(commands, packageCommand{"rpm", []string{"-qf", path}})
	case "apk":
		commands = append(commands, packageCommand{"apk", []string{"info", "--who-owns", path}})
	case "pacman":
		commands = append(commands, packageCommand{"pacman", []string{"-Qo", path}})
	case "xbps-install":
		commands = append(commands, packageCommand{"xbps-query", []string{"-o", path}})
	case "opkg":
		commands = append(commands, packageCommand{"opkg", []string{"search", path}})
	case "brew":
		prefix, prefixErr := n.runCommand(context.Background(), n.commandTimeout(), "brew", "--prefix", "nginx")
		if prefixErr != nil {
			return false
		}
		prefix, _ = filepath.Abs(strings.TrimSpace(prefix))
		return prefix != "" && (path == filepath.Join(prefix, "sbin", "nginx") || strings.HasPrefix(path, prefix+string(os.PathSeparator)))
	default:
		return false
	}
	for _, command := range commands {
		if !n.commandExists(command.Name) {
			continue
		}
		if output, runErr := n.runCommand(context.Background(), n.commandTimeout(), command.Name, command.Args...); runErr == nil && strings.TrimSpace(output) != "" {
			return true
		}
	}
	return false
}

func (n *Nginx) configRoot() (string, error) {
	info, err := n.Info()
	if err != nil {
		return "", err
	}
	if info.ConfigPath == "" {
		return "", errors.New("无法确定 nginx 配置目录")
	}
	root, err := filepath.Abs(filepath.Dir(info.ConfigPath))
	if err != nil {
		return "", fmt.Errorf("解析 Nginx 配置目录失败: %w", err)
	}
	return filepath.Clean(root), nil
}

func (n *Nginx) resolveConfigFile(path string) (string, error) {
	root, err := n.configRoot()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(path) == "" {
		info, infoErr := n.Info()
		if infoErr != nil {
			return "", infoErr
		}
		path = info.ConfigPath
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	path, err = filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("解析 Nginx 配置文件路径失败: %w", err)
	}
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", errors.New("配置文件必须位于 Nginx 配置目录内")
	}
	return path, nil
}

func (n *Nginx) ensureConfigWritable(path string) error {
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("不允许通过符号链接修改 Nginx 配置")
	}
	return nil
}

func (n *Nginx) readTextFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", errors.New("配置路径不能是目录")
	}
	if info.Size() > maxConfigFileSize {
		return "", fmt.Errorf("配置文件超过 %d MB 限制", maxConfigFileSize/(1<<20))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (n *Nginx) writeValidatedConfig(path string, data []byte) error {
	if int64(len(data)) > maxConfigFileSize {
		return fmt.Errorf("配置文件超过 %d MB 限制", maxConfigFileSize/(1<<20))
	}
	if err := n.ensureConfigWritable(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("创建 Nginx 配置目录失败: %w", err)
	}
	old, oldMode, oldErr := n.readSite(path)
	if err := n.atomicWrite(path, data, oldMode); err != nil {
		return err
	}
	if err := n.Test(); err != nil {
		if oldErr == nil {
			_ = n.atomicWrite(path, old, oldMode)
		} else {
			_ = os.Remove(path)
		}
		return err
	}
	return nil
}

// LogPath 返回 error 或 access 日志路径。
func (n *Nginx) LogPath(kind string) (string, error) {
	info, err := n.Info()
	if err != nil {
		return "", err
	}
	kind = strings.ToLower(strings.TrimSpace(kind))
	option := ""
	defaultName := ""
	switch kind {
	case "error", "errors", "error_log":
		kind = "error"
		option, defaultName = "error-log-path", "error.log"
	case "access", "http", "access_log":
		kind = "access"
		option, defaultName = "http-log-path", "access.log"
	default:
		return "", errors.New("日志类型必须是 error 或 access")
	}
	if config, configErr := n.ReadLogConfig(); configErr == nil {
		configured := false
		for _, directive := range config.Directives {
			if directive.Kind != kind || len(directive.Scope) > 1 || (len(directive.Scope) == 1 && directive.Scope[0] != "http") {
				continue
			}
			configured = true
			if !directive.Enabled || directive.Syslog || directive.Path == "" {
				continue
			}
			return n.resolveLogPath(directive.Path)
		}
		if configured {
			return "", errors.New("当前日志没有可读取的文件路径")
		}
	}
	path := n.parseBuildOption(info.Build, option)
	if path == "" {
		root := info.PrefixPath
		if root == "" {
			root = filepath.Dir(info.ConfigPath)
		}
		path = filepath.Join(root, "logs", defaultName)
	}
	if strings.EqualFold(path, "stderr") || strings.EqualFold(path, "stdout") {
		return "", errors.New("当前日志输出到标准流，没有可读取的日志文件")
	}
	if strings.HasPrefix(strings.ToLower(path), "syslog:") {
		return "", errors.New("当前日志输出到 syslog，没有可读取的日志文件")
	}
	return n.resolveBuildPath(path, info.PrefixPath), nil
}

// ListLogs 列出 Nginx 常用日志文件。
func (n *Nginx) ListLogs() ([]LogInfo, error) {
	result := make([]LogInfo, 0, 2)
	definition, definitionErr := n.ReadLogConfig()
	for _, kind := range []string{"error", "access"} {
		item := LogInfo{Kind: kind, Enabled: true}
		configured := false
		if definitionErr == nil {
			for _, directive := range definition.Directives {
				if directive.Kind != kind || len(directive.Scope) > 1 || (len(directive.Scope) == 1 && directive.Scope[0] != "http") {
					continue
				}
				configured = true
				item.Path = directive.Path
				item.Enabled = directive.Enabled
				item.Format = directive.Format
				item.Level = directive.Level
				item.Options = append([]string(nil), directive.Options...)
				item.Syslog = directive.Syslog
				if directive.Enabled && !directive.Syslog && !n.isNonFileLogPath(directive.Path) && !n.isDynamicLogPath(directive.Path) {
					path, pathErr := n.resolveLogPath(directive.Path)
					if pathErr != nil {
						return nil, pathErr
					}
					item.Path = path
				}
				break
			}
		}
		if !configured {
			path, err := n.LogPath(kind)
			if err != nil {
				return nil, err
			}
			item.Path = path
		}
		if item.Path != "" && !item.Syslog && !n.isNonFileLogPath(item.Path) && !n.isDynamicLogPath(item.Path) {
			if info, statErr := os.Stat(item.Path); statErr == nil {
				item.Exists = !info.IsDir()
				if item.Exists {
					item.Size = info.Size()
					item.ModifiedAt = info.ModTime()
				}
			} else if !os.IsNotExist(statErr) {
				return nil, statErr
			}
		}
		result = append(result, item)
	}
	return result, nil
}

// ReadLogTail 读取日志末尾内容，避免一次性加载整个日志文件。
func (n *Nginx) ReadLogTail(kind string, lines int) (string, error) {
	path, err := n.LogPath(kind)
	if err != nil {
		return "", err
	}
	if lines <= 0 {
		lines = 200
	}
	if lines > 10000 {
		lines = 10000
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	if info.Size() == 0 {
		return "", nil
	}
	start := int64(0)
	if info.Size() > maxLogReadSize {
		start = info.Size() - maxLogReadSize
	}
	if _, err = file.Seek(start, io.SeekStart); err != nil {
		return "", err
	}
	data, err := io.ReadAll(io.LimitReader(file, maxLogReadSize))
	if err != nil {
		return "", err
	}
	text := string(data)
	parts := strings.Split(text, "\n")
	end := len(parts)
	trailingNewline := end > 0 && parts[end-1] == ""
	if trailingNewline {
		end--
	}
	startLine := end - lines
	if startLine < 0 {
		startLine = 0
	}
	text = strings.Join(parts[startLine:end], "\n")
	if trailingNewline {
		text += "\n"
	}
	return text, nil
}

// ReadLog 使用默认 200 行读取日志末尾。
func (n *Nginx) ReadLog(kind string) (string, error) {
	return n.ReadLogTail(kind, 200)
}

// ReadLogPage 按字节游标读取日志，适合前端历史日志懒加载。
func (n *Nginx) ReadLogPage(kind string, offset int64, maxBytes int64) (*LogPage, error) {
	path, err := n.LogPath(kind)
	if err != nil {
		return nil, err
	}
	if maxBytes <= 0 || maxBytes > maxLogReadSize {
		maxBytes = maxLogReadSize
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if offset < 0 {
		offset = 0
	}
	if offset > info.Size() {
		offset = info.Size()
	}
	if _, err = file.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(file, maxBytes))
	if err != nil {
		return nil, err
	}
	start := offset
	if offset > 0 && len(data) > 0 {
		if index := strings.IndexByte(string(data), '\n'); index >= 0 {
			start += int64(index + 1)
			data = data[index+1:]
		}
	}
	nextOffset := start + int64(len(data))
	if nextOffset < info.Size() {
		if index := strings.LastIndexByte(string(data), '\n'); index >= 0 {
			data = data[:index+1]
			nextOffset = start + int64(index+1)
		}
	}
	return &LogPage{Kind: strings.ToLower(strings.TrimSpace(kind)), Path: path, Offset: offset, NextOffset: nextOffset, Size: info.Size(), Content: string(data), EOF: nextOffset >= info.Size()}, nil
}

// FollowLog 持续读取日志新增内容，直到 ctx 取消或回调返回错误。
func (n *Nginx) FollowLog(ctx context.Context, kind string, offset int64, interval time.Duration, callback func(LogPage) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if callback == nil {
		return errors.New("日志回调不能为空")
	}
	if interval <= 0 {
		interval = time.Second
	}
	for {
		page, err := n.ReadLogPage(kind, offset, maxLogReadSize)
		if err != nil {
			if os.IsNotExist(err) {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(interval):
					continue
				}
			}
			return err
		}
		if page.Content != "" {
			if err = callback(*page); err != nil {
				return err
			}
		}
		offset = page.NextOffset
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}

// ClearLog 清空全局 error 或 access 日志。
func (n *Nginx) ClearLog(kind string) error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	path, err := n.LogPath(kind)
	if err != nil {
		return err
	}
	return n.clearLogFile(path)
}

// RotateLogs 轮转全局 error/access 日志并通知 Nginx 重新打开日志文件。
func (n *Nginx) RotateLogs() error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	for _, kind := range []string{"error", "access"} {
		path, err := n.LogPath(kind)
		if err != nil {
			return err
		}
		if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
			continue
		} else if statErr != nil {
			return statErr
		}
		rotated := path + "." + time.Now().UTC().Format("20060102-150405.000000000")
		if err = os.Rename(path, rotated); err != nil {
			return err
		}
		file, createErr := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0640)
		if createErr != nil {
			_ = os.Rename(rotated, path)
			return createErr
		}
		if closeErr := file.Close(); closeErr != nil {
			return closeErr
		}
	}
	info, err := n.Info()
	if err != nil {
		return err
	}
	output, err := n.runCommand(context.Background(), n.commandTimeout(), info.BinaryPath, append(n.runtimeArgs(info), "-s", "reopen")...)
	if err != nil {
		return fmt.Errorf("重新打开 Nginx 日志失败: %w", n.formatCommandOutput(output))
	}
	return nil
}

// SiteLogInfo 获取站点配置中的访问日志和错误日志。
func (n *Nginx) SiteLogInfo(name string) ([]LogInfo, error) {
	definition, err := n.ReadSiteLogConfig(name)
	if err != nil {
		return nil, err
	}
	result := make([]LogInfo, 0, len(definition.Directives))
	for _, item := range definition.Directives {
		if !item.Enabled || item.Path == "" || strings.EqualFold(item.Path, "off") {
			continue
		}
		logInfo := LogInfo{
			Kind:    item.Kind,
			Path:    item.Path,
			Enabled: item.Enabled,
			Format:  item.Format,
			Level:   item.Level,
			Options: append([]string(nil), item.Options...),
			Syslog:  item.Syslog,
		}
		if item.Syslog || n.isNonFileLogPath(item.Path) || n.isDynamicLogPath(item.Path) {
			result = append(result, logInfo)
			continue
		}
		path, pathErr := n.resolveLogPath(item.Path)
		if pathErr != nil {
			return nil, pathErr
		}
		logInfo.Path = path
		if stat, statErr := os.Stat(path); statErr == nil {
			logInfo.Exists = !stat.IsDir()
			if logInfo.Exists {
				logInfo.Size = stat.Size()
				logInfo.ModifiedAt = stat.ModTime()
			}
		} else if !os.IsNotExist(statErr) {
			return nil, statErr
		}
		result = append(result, logInfo)
	}
	return result, nil
}

// ReadSiteLogTail 读取站点日志末尾内容。
func (n *Nginx) ReadSiteLogTail(name, kind string, lines int) (string, error) {
	path, err := n.siteLogPath(name, kind)
	if err != nil {
		return "", err
	}
	return n.readLogTailPath(path, lines)
}

// ReadSiteLog 按计划任务日志的游标语义读取站点日志。
// cursor 大于 0 时读取新增内容；cursor 为 0 且 before 大于 0 时读取 before 之前的历史内容；
// 两者都为 0 时读取日志末尾。首次读取返回最新内容和可继续使用的 cursor。
func (n *Nginx) ReadSiteLog(name, kind string, cursor, before int64, pageSize int) (*LogCursorPage, error) {
	path, err := n.siteLogPath(name, kind)
	if err != nil {
		return nil, err
	}
	return n.readCursorLogPath(path, strings.ToLower(strings.TrimSpace(kind)), cursor, before, pageSize)
}

// ReadSiteLogPage 是 ReadSiteLog 的分页命名别名，便于按现有 ReadLogPage API 使用。
func (n *Nginx) ReadSiteLogPage(name, kind string, cursor, before int64, pageSize int) (*LogCursorPage, error) {
	return n.ReadSiteLog(name, kind, cursor, before, pageSize)
}

// ClearSiteLog 清空站点指定日志。
func (n *Nginx) ClearSiteLog(name, kind string) error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	path, err := n.siteLogPath(name, kind)
	if err != nil {
		return err
	}
	return n.clearLogFile(path)
}

// SetSiteLog 开关站点访问日志或错误日志，并保留站点其余原始配置。
func (n *Nginx) SetSiteLog(name, kind string, enabled bool, path string) error {
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind != "access" && kind != "error" {
		return errors.New("日志类型必须是 access 或 error")
	}
	directive := kind + "_log"
	if enabled {
		if strings.TrimSpace(path) == "" {
			info, err := n.Info()
			if err != nil {
				return err
			}
			base := strings.TrimSuffix(strings.TrimSuffix(filepath.Base(name), ".conf"), ".CONF")
			path = filepath.Join(info.PrefixPath, "logs", base+"-"+kind+".log")
		}
		path, err := n.nginxConfigPath(path)
		if err != nil {
			return err
		}
		return n.updateSiteDirective(name, directive, path+";")
	}
	return n.updateSiteDirective(name, directive, "off;")
}

func (n *Nginx) readCursorLogPath(path, kind string, cursor, before int64, pageSize int) (*LogCursorPage, error) {
	if pageSize < 1 {
		pageSize = 300
	}
	if pageSize > 10000 {
		pageSize = 10000
	}
	if cursor < 0 {
		cursor = 0
	}
	if before < 0 {
		before = 0
	}
	result := &LogCursorPage{
		Kind:        kind,
		Path:        path,
		Lines:       []string{},
		Cursor:      cursor,
		StartCursor: cursor,
		HasPrevious: cursor > 0,
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			result.Cursor = 0
			result.StartCursor = 0
			result.HasPrevious = false
			result.End = cursor > 0
			return result, nil
		}
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	result.Size = info.Size()
	if info.Size() == 0 {
		result.Cursor = 0
		result.StartCursor = 0
		result.HasPrevious = false
		return result, nil
	}

	readBefore := func(end int64) ([]string, []int64, error) {
		if end < 0 {
			end = 0
		}
		if end > info.Size() {
			end = info.Size()
		}
		const chunkSize int64 = 64 * 1024
		position := end
		lineCount := 0
		chunks := make([][]byte, 0, 2)
		totalSize := 0
		for position > 0 && lineCount <= pageSize {
			readSize := chunkSize
			if position < readSize {
				readSize = position
			}
			position -= readSize
			chunk := make([]byte, int(readSize))
			readCount, readErr := file.ReadAt(chunk, position)
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				return nil, nil, readErr
			}
			chunk = chunk[:readCount]
			chunks = append(chunks, chunk)
			totalSize += len(chunk)
			lineCount += bytes.Count(chunk, []byte{'\n'})
		}
		data := make([]byte, 0, totalSize)
		for index := len(chunks) - 1; index >= 0; index-- {
			data = append(data, chunks[index]...)
		}
		start := 0
		if position > 0 {
			lineEnd := bytes.IndexByte(data, '\n')
			if lineEnd < 0 {
				return nil, nil, nil
			}
			start = lineEnd + 1
		}
		lines := make([]string, 0, pageSize)
		starts := make([]int64, 0, pageSize)
		for start < len(data) {
			relEnd := bytes.IndexByte(data[start:], '\n')
			if relEnd < 0 {
				if end == info.Size() {
					lines = append(lines, strings.TrimSuffix(string(data[start:]), "\r"))
					starts = append(starts, position+int64(start))
				}
				break
			}
			lineEnd := start + relEnd
			lines = append(lines, strings.TrimSuffix(string(data[start:lineEnd]), "\r"))
			starts = append(starts, position+int64(start))
			start = lineEnd + 1
		}
		if len(lines) > pageSize {
			first := len(lines) - pageSize
			lines = lines[first:]
			starts = starts[first:]
		}
		return lines, starts, nil
	}

	readAfter := func(start int64) ([]string, int64, error) {
		if start < 0 {
			start = 0
		}
		if start >= info.Size() {
			return []string{}, start, nil
		}
		if _, seekErr := file.Seek(start, io.SeekStart); seekErr != nil {
			return nil, start, seekErr
		}
		reader := bufio.NewReaderSize(file, 64*1024)
		lines := make([]string, 0, pageSize)
		next := start
		for len(lines) < pageSize {
			line, readErr := reader.ReadString('\n')
			if len(line) == 0 && readErr != nil {
				if errors.Is(readErr, io.EOF) {
					break
				}
				return nil, next, readErr
			}
			if len(line) == 0 || line[len(line)-1] != '\n' {
				break
			}
			next += int64(len(line))
			line = strings.TrimSuffix(line, "\n")
			line = strings.TrimSuffix(line, "\r")
			lines = append(lines, line)
			if readErr != nil {
				if errors.Is(readErr, io.EOF) {
					break
				}
				return nil, next, readErr
			}
		}
		return lines, next, nil
	}

	if cursor > 0 {
		if cursor > info.Size() {
			lines, starts, readErr := readBefore(info.Size())
			if readErr != nil {
				return nil, readErr
			}
			result.Lines = lines
			result.Cursor = info.Size()
			result.StartCursor = 0
			result.HasPrevious = false
			result.End = true
			if len(starts) > 0 {
				result.StartCursor = starts[0]
				result.HasPrevious = starts[0] > 0
			}
			return result, nil
		}
		lines, next, readErr := readAfter(cursor)
		if readErr != nil {
			return nil, readErr
		}
		result.Lines = lines
		result.Cursor = next
		result.StartCursor = cursor
		result.HasPrevious = cursor > 0
		return result, nil
	}

	end := info.Size()
	if before > 0 && before < end {
		end = before
	}
	lines, starts, readErr := readBefore(end)
	if readErr != nil {
		return nil, readErr
	}
	result.Lines = lines
	result.Cursor = info.Size()
	result.StartCursor = 0
	result.HasPrevious = false
	if len(starts) > 0 {
		result.StartCursor = starts[0]
		result.HasPrevious = starts[0] > 0
	}
	return result, nil
}

func (n *Nginx) readLogTailPath(path string, lines int) (string, error) {
	if lines <= 0 {
		lines = 200
	}
	if lines > 10000 {
		lines = 10000
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	start := int64(0)
	if info.Size() > maxLogReadSize {
		start = info.Size() - maxLogReadSize
	}
	if _, err = file.Seek(start, io.SeekStart); err != nil {
		return "", err
	}
	data, err := io.ReadAll(io.LimitReader(file, maxLogReadSize))
	if err != nil {
		return "", err
	}
	parts := strings.Split(string(data), "\n")
	end := len(parts)
	trailingNewline := end > 0 && parts[end-1] == ""
	if trailingNewline {
		end--
	}
	startLine := end - lines
	if startLine < 0 {
		startLine = 0
	}
	result := strings.Join(parts[startLine:end], "\n")
	if trailingNewline {
		result += "\n"
	}
	return result, nil
}

func (n *Nginx) clearLogFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("日志路径不能是目录")
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	return file.Close()
}

func (n *Nginx) siteLogPath(name, kind string) (string, error) {
	definition, err := n.ReadSiteLogConfig(name)
	if err != nil {
		return "", err
	}
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind != "access" && kind != "error" {
		return "", errors.New("日志类型必须是 access 或 error")
	}
	for _, item := range definition.Directives {
		if item.Kind != kind || !item.Enabled || item.Syslog || n.isNonFileLogPath(item.Path) || n.isDynamicLogPath(item.Path) || item.Path == "" || strings.EqualFold(item.Path, "off") {
			continue
		}
		return n.resolveLogPath(item.Path)
	}
	return "", errors.New("该站点未启用对应日志文件")
}

func (n *Nginx) resolveLogPath(path string) (string, error) {
	path = strings.Trim(strings.TrimSpace(path), `"'`)
	if n.isNonFileLogPath(path) || n.isDynamicLogPath(path) || strings.Contains(path, "://") || strings.HasPrefix(strings.ToLower(path), "syslog:") {
		return "", errors.New("日志没有文件路径")
	}
	info, err := n.Info()
	if err != nil {
		return "", err
	}
	return n.resolveBuildPath(path, info.PrefixPath), nil
}

func (n *Nginx) isDynamicLogPath(path string) bool {
	return strings.Contains(path, "$")
}

func (n *Nginx) isNonFileLogPath(path string) bool {
	path = strings.ToLower(strings.Trim(strings.TrimSpace(path), `"'`))
	return path == "" || path == "off" || path == "stdout" || path == "stderr" || strings.HasPrefix(path, "syslog:")
}

func (n *Nginx) updateSiteDirective(name, directive, value string) error {
	content, err := n.ReadSite(name)
	if err != nil {
		return err
	}
	lines := strings.Split(content, "\n")
	found := false
	depth := 0
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, directive+" ") {
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			lines[index] = indent + directive + " " + value
			found = true
		}
		if !found && depth == 1 && trimmed == "}" {
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			lines = append(lines[:index], append([]string{indent + directive + " " + value}, lines[index:]...)...)
			found = true
			break
		}
		depth += strings.Count(line, "{")
		depth -= strings.Count(line, "}")
	}
	if !found {
		return errors.New("未找到 server 配置块")
	}
	return n.WriteSite(name, strings.Join(lines, "\n"))
}

// InspectCertificate 读取 PEM 证书并返回有效期、主题和域名信息。
func (n *Nginx) InspectCertificate(path string) (*CertificateInfo, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("证书路径不能为空")
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() || info.Size() > maxConfigFileSize {
		return nil, errors.New("证书文件无效")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, errors.New("证书不是有效的 PEM X.509 证书")
	}
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析 TLS 证书失败: %w", err)
	}
	return &CertificateInfo{
		Path:      path,
		Subject:   certificate.Subject.String(),
		Issuer:    certificate.Issuer.String(),
		DNSNames:  append([]string(nil), certificate.DNSNames...),
		NotBefore: certificate.NotBefore,
		NotAfter:  certificate.NotAfter,
		Expired:   time.Now().After(certificate.NotAfter),
	}, nil
}

// ValidateTLSFiles 校验证书和私钥文件是否可用于 Nginx 配置。
func (n *Nginx) ValidateTLSFiles(certPath, keyPath string) error {
	if _, err := n.InspectCertificate(certPath); err != nil {
		return err
	}
	certificateData, err := os.ReadFile(certPath)
	if err != nil {
		return err
	}
	keyInfo, err := os.Stat(keyPath)
	if err != nil {
		return err
	}
	if keyInfo.IsDir() || keyInfo.Size() == 0 || keyInfo.Size() > maxConfigFileSize {
		return errors.New("TLS 私钥文件无效")
	}
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}
	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil || !strings.Contains(keyBlock.Type, "PRIVATE KEY") {
		return errors.New("TLS 私钥不是有效的 PEM 私钥")
	}
	if _, err = tls.X509KeyPair(certificateData, keyData); err != nil {
		return fmt.Errorf("TLS 证书和私钥不匹配: %w", err)
	}
	return nil
}

// InstallCertificate 校验证书和私钥后复制到 Nginx 证书目录，并将私钥权限设置为 0600。
func (n *Nginx) InstallCertificate(name, certPath, keyPath, destinationDir string) (*CertificateFiles, error) {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	if _, err := n.normalizeIdentifier(name); err != nil {
		return nil, err
	}
	if err := n.ValidateTLSFiles(certPath, keyPath); err != nil {
		return nil, err
	}
	if strings.TrimSpace(destinationDir) == "" {
		info, err := n.Info()
		if err != nil {
			return nil, err
		}
		destinationDir = filepath.Join(info.PrefixPath, "certs")
		if info.PrefixPath == "" {
			destinationDir = filepath.Join(filepath.Dir(info.ConfigPath), "certs")
		}
	}
	destinationDir, err := filepath.Abs(destinationDir)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(destinationDir, 0755); err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	certTarget := filepath.Join(destinationDir, name+".crt")
	keyTarget := filepath.Join(destinationDir, name+".key")
	if err = n.copySecureFile(certPath, certTarget, 0644); err != nil {
		return nil, err
	}
	if err = n.copySecureFile(keyPath, keyTarget, 0600); err != nil {
		return nil, err
	}
	return &CertificateFiles{Name: name, CertPath: certTarget, KeyPath: keyTarget}, nil
}

// ListCertificates 列出证书目录中能够找到对应私钥的证书。
func (n *Nginx) ListCertificates(directory string) ([]CertificateFiles, error) {
	if strings.TrimSpace(directory) == "" {
		info, err := n.Info()
		if err != nil {
			return nil, err
		}
		directory = filepath.Join(info.PrefixPath, "certs")
		if info.PrefixPath == "" {
			directory = filepath.Join(filepath.Dir(info.ConfigPath), "certs")
		}
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		if os.IsNotExist(err) {
			return []CertificateFiles{}, nil
		}
		return nil, err
	}
	result := make([]CertificateFiles, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".crt") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		keyPath := filepath.Join(directory, name+".key")
		if _, statErr := os.Stat(keyPath); statErr != nil {
			continue
		}
		result = append(result, CertificateFiles{Name: name, CertPath: filepath.Join(directory, entry.Name()), KeyPath: keyPath})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

// RemoveCertificate 删除证书和私钥文件。
func (n *Nginx) RemoveCertificate(name, directory string) error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	if _, err := n.normalizeIdentifier(name); err != nil {
		return err
	}
	if strings.TrimSpace(directory) == "" {
		info, err := n.Info()
		if err != nil {
			return err
		}
		directory = filepath.Join(info.PrefixPath, "certs")
		if info.PrefixPath == "" {
			directory = filepath.Join(filepath.Dir(info.ConfigPath), "certs")
		}
	}
	for _, suffix := range []string{".crt", ".key"} {
		path := filepath.Join(directory, strings.TrimSpace(name)+suffix)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// VerifyFileSHA256 校验文件 SHA-256，expected 可以带或不带空格和换行。
func (n *Nginx) VerifyFileSHA256(path, expected string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(data)
	expected = strings.ToLower(strings.TrimSpace(expected))
	if expected == "" || expected != hex.EncodeToString(digest[:]) {
		return fmt.Errorf("文件 SHA-256 校验失败: %s", path)
	}
	return nil
}

// CheckPort 检查本机 TCP 端口是否可以监听。
func (n *Nginx) CheckPort(port int) (*PortCheck, error) {
	return n.CheckPortAddress("tcp", "0.0.0.0", port)
}

// CheckPortAddress 按指定网络和地址检查 TCP/UDP 端口。
func (n *Nginx) CheckPortAddress(network, host string, port int) (*PortCheck, error) {
	if err := n.validateListenPort(port); err != nil {
		return nil, err
	}
	network = strings.ToLower(strings.TrimSpace(network))
	if network != "tcp" && network != "tcp4" && network != "tcp6" && network != "udp" && network != "udp4" && network != "udp6" {
		return nil, errors.New("网络类型必须是 tcp、tcp4、tcp6、udp、udp4 或 udp6")
	}
	if strings.TrimSpace(host) == "" {
		host = "0.0.0.0"
		if network == "tcp6" || network == "udp6" {
			host = "::"
		}
	}
	address := net.JoinHostPort(host, strconv.Itoa(port))
	var listener io.Closer
	var err error
	if strings.HasPrefix(network, "udp") {
		listener, err = net.ListenPacket(network, address)
	} else {
		listener, err = net.Listen(network, address)
	}
	result := &PortCheck{Port: port, Network: network, Address: address, Available: err == nil}
	if err != nil {
		result.Error = err.Error()
		return result, nil
	}
	return result, listener.Close()
}

// CheckPorts 批量检查端口。
func (n *Nginx) CheckPorts(ports ...int) ([]PortCheck, error) {
	result := make([]PortCheck, 0, len(ports))
	for _, port := range ports {
		check, err := n.CheckPort(port)
		if err != nil {
			return nil, err
		}
		result = append(result, *check)
	}
	return result, nil
}

// CheckDependencies 检查命令依赖是否存在。不传参数时检查 nginx 和 openssl。
func (n *Nginx) CheckDependencies(names ...string) []DependencyStatus {
	if len(names) == 0 {
		names = []string{"nginx", "openssl"}
	}
	result := make([]DependencyStatus, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		path := n.commandPath(name)
		if path == "" {
			result = append(result, DependencyStatus{Name: name})
			continue
		}
		result = append(result, DependencyStatus{Name: name, Path: path, Available: true})
	}
	return result
}

// Preflight 执行安装、配置、进程、端口和命令依赖预检。
func (n *Nginx) Preflight(ports ...int) (*PreflightReport, error) {
	report := &PreflightReport{Environment: n.Environment(), Ports: make([]PortCheck, 0), Dependencies: n.CheckDependencies(), Errors: make([]string, 0)}
	_, infoErr := n.Info()
	if infoErr != nil {
		report.Errors = append(report.Errors, infoErr.Error())
	} else {
		report.Installed = true
		if testErr := n.Test(); testErr != nil {
			report.Errors = append(report.Errors, testErr.Error())
		} else {
			report.ConfigValid = true
		}
		if status, statusErr := n.Status(); statusErr == nil {
			report.Running = status.Running
		} else {
			report.Errors = append(report.Errors, statusErr.Error())
		}
	}
	if checks, checkErr := n.CheckPorts(ports...); checkErr != nil {
		report.Errors = append(report.Errors, checkErr.Error())
	} else {
		report.Ports = checks
	}
	for _, dependency := range report.Dependencies {
		if !dependency.Available {
			report.Errors = append(report.Errors, "缺少依赖: "+dependency.Name)
		}
	}
	return report, nil
}

// RuntimeStats 读取 Nginx master/worker 进程的运行指标，不依赖 gopsutil。
func (n *Nginx) RuntimeStats() (*RuntimeStats, error) {
	status, err := n.Status()
	if err != nil {
		return nil, err
	}
	result := &RuntimeStats{SampledAt: time.Now(), MasterPID: status.Pid, Running: status.Running, WorkerPIDs: make([]int, 0)}
	if !status.Running || status.Pid <= 0 {
		return result, nil
	}
	if runtime.GOOS == "linux" {
		result.MemoryBytes, result.Threads, result.OpenFiles = n.procMetrics(status.Pid)
		result.CPUSeconds = n.procCPUSeconds(status.Pid)
		result.WorkerPIDs = n.findWorkerPIDs(status.Pid, status.Binary)
		for _, pid := range result.WorkerPIDs {
			memory, threads, files := n.procMetrics(pid)
			result.MemoryBytes += memory
			result.Threads += threads
			result.OpenFiles += files
			result.CPUSeconds += n.procCPUSeconds(pid)
		}
	}
	return result, nil
}

// ReadStubStatus 读取 ngx_http_stub_status_module 的状态接口。
func (n *Nginx) ReadStubStatus(endpoint string) (*StubStatus, error) {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, errors.New("stub_status 地址必须是 http 或 https URL")
	}
	response, err := (&http.Client{Timeout: n.commandTimeout()}).Get(parsed.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("读取 stub_status 失败: HTTP %s", response.Status)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 64<<10))
	if err != nil {
		return nil, err
	}
	lines := strings.Fields(string(body))
	if len(lines) < 10 {
		return nil, errors.New("stub_status 响应格式不正确")
	}
	result := &StubStatus{}
	parseAfter := func(label string) (int64, error) {
		for index, value := range lines {
			if strings.TrimSuffix(value, ":") == strings.TrimSuffix(label, ":") && index+1 < len(lines) {
				return strconv.ParseInt(strings.TrimSuffix(lines[index+1], ":"), 10, 64)
			}
		}
		return 0, fmt.Errorf("stub_status 缺少 %s", label)
	}
	if result.Active, err = parseAfter("connections"); err != nil {
		return nil, err
	}
	requestIndex := -1
	for index, value := range lines {
		if value == "requests" {
			requestIndex = index
			break
		}
	}
	if requestIndex < 0 || requestIndex+3 >= len(lines) {
		return nil, errors.New("stub_status 缺少请求统计")
	}
	if result.Accepted, err = strconv.ParseInt(lines[requestIndex+1], 10, 64); err != nil {
		return nil, err
	}
	if result.Handled, err = strconv.ParseInt(lines[requestIndex+2], 10, 64); err != nil {
		return nil, err
	}
	if result.Requests, err = strconv.ParseInt(lines[requestIndex+3], 10, 64); err != nil {
		return nil, err
	}
	if result.Reading, err = parseAfter("Reading"); err != nil {
		return nil, err
	}
	if result.Writing, err = parseAfter("Writing"); err != nil {
		return nil, err
	}
	if result.Waiting, err = parseAfter("Waiting"); err != nil {
		return nil, err
	}
	return result, nil
}

func (n *Nginx) procMetrics(pid int) (uint64, int, int) {
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "status"))
	if err != nil {
		return 0, 0, 0
	}
	var memory uint64
	var threads int
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "VmRSS:":
			value, _ := strconv.ParseUint(fields[1], 10, 64)
			memory = value * 1024
		case "Threads:":
			threads, _ = strconv.Atoi(fields[1])
		}
	}
	openFiles := 0
	if entries, readErr := os.ReadDir(filepath.Join("/proc", strconv.Itoa(pid), "fd")); readErr == nil {
		openFiles = len(entries)
	}
	return memory, threads, openFiles
}

func (n *Nginx) procCPUSeconds(pid int) float64 {
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		return 0
	}
	closeIndex := strings.LastIndex(string(data), ")")
	if closeIndex < 0 {
		return 0
	}
	fields := strings.Fields(string(data)[closeIndex+1:])
	if len(fields) < 13 {
		return 0
	}
	userTicks, userErr := strconv.ParseFloat(fields[11], 64)
	systemTicks, systemErr := strconv.ParseFloat(fields[12], 64)
	if userErr != nil || systemErr != nil {
		return 0
	}
	return (userTicks + systemTicks) / 100
}

func (n *Nginx) findWorkerPIDs(masterPID int, binary string) []int {
	if runtime.GOOS != "linux" {
		return []int{}
	}
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return []int{}
	}
	workers := make([]int, 0)
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid == masterPID {
			continue
		}
		process, processErr := n.inspectProcess(pid)
		if processErr != nil || !process.Running {
			continue
		}
		if strings.Contains(process.CommandLine, "nginx: worker process") || (filepath.Base(process.Executable) == filepath.Base(binary) && strings.Contains(process.CommandLine, "worker")) {
			workers = append(workers, pid)
		}
	}
	sort.Ints(workers)
	return workers
}

func (n *Nginx) copySecureFile(source, destination string, mode os.FileMode) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if int64(len(data)) > maxConfigFileSize {
		return errors.New("证书文件超过大小限制")
	}
	return n.atomicWrite(destination, data, mode)
}

// BuildHTTPSReverseProxy 生成带 TLS 的反向代理配置。
func (n *Nginx) BuildHTTPSReverseProxy(serverName, target, certPath, keyPath string, listen int) (string, error) {
	if err := n.ValidateTLSFiles(certPath, keyPath); err != nil {
		return "", err
	}
	if err := n.validateListenPort(listen); err != nil {
		return "", err
	}
	serverName, err := n.normalizeServerName(serverName)
	if err != nil {
		return "", err
	}
	targetURL, err := n.proxyURL(target)
	if err != nil {
		return "", err
	}
	cert, err := n.nginxConfigPath(certPath)
	if err != nil {
		return "", err
	}
	key, err := n.nginxConfigPath(keyPath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`server {
    listen %d ssl;
    server_name %s;
    ssl_certificate %s;
    ssl_certificate_key %s;
    ssl_protocols TLSv1.2 TLSv1.3;

    location / {
        proxy_pass %s;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
`, listen, serverName, cert, key, targetURL.String()), nil
}

// WriteHTTPSReverseProxy 写入带 TLS 的反向代理站点。
func (n *Nginx) WriteHTTPSReverseProxy(name, serverName, target, certPath, keyPath string, listen int) error {
	content, err := n.BuildHTTPSReverseProxy(serverName, target, certPath, keyPath, listen)
	if err != nil {
		return err
	}
	return n.WriteSite(name, content)
}

// BuildHTTPSStaticSite 生成带 TLS 的静态站点配置。
func (n *Nginx) BuildHTTPSStaticSite(serverName, root, certPath, keyPath string, listen int) (string, error) {
	if err := n.ValidateTLSFiles(certPath, keyPath); err != nil {
		return "", err
	}
	if err := n.validateListenPort(listen); err != nil {
		return "", err
	}
	serverName, err := n.normalizeServerName(serverName)
	if err != nil {
		return "", err
	}
	root, err = n.nginxConfigPath(root)
	if err != nil {
		return "", err
	}
	cert, err := n.nginxConfigPath(certPath)
	if err != nil {
		return "", err
	}
	key, err := n.nginxConfigPath(keyPath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`server {
    listen %d ssl;
    server_name %s;
    root %s;
    index index.html index.htm;
    ssl_certificate %s;
    ssl_certificate_key %s;
    ssl_protocols TLSv1.2 TLSv1.3;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
`, listen, serverName, root, cert, key), nil
}

// WriteHTTPSStaticSite 写入带 TLS 的静态站点。
func (n *Nginx) WriteHTTPSStaticSite(name, serverName, root, certPath, keyPath string, listen int) error {
	content, err := n.BuildHTTPSStaticSite(serverName, root, certPath, keyPath, listen)
	if err != nil {
		return err
	}
	return n.WriteSite(name, content)
}

// SiteDir 返回站点配置目录。优先使用主配置中常见的站点 include 目录，找不到时回退到 conf.d。
func (n *Nginx) SiteDir() (string, error) {
	info, err := n.Info()
	if err != nil {
		return "", err
	}
	if info.ConfigPath == "" {
		return "", errors.New("无法确定 nginx 配置目录")
	}
	configDir := filepath.Dir(info.ConfigPath)
	includeFallback := ""
	if data, readErr := os.ReadFile(info.ConfigPath); readErr == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if comment := strings.IndexByte(line, '#'); comment >= 0 {
				line = strings.TrimSpace(line[:comment])
			}
			fields := strings.Fields(line)
			if len(fields) < 2 || fields[0] != "include" {
				continue
			}
			include := strings.Trim(strings.TrimSuffix(fields[1], ";"), "\"'")
			if !strings.ContainsAny(include, "*?[") || strings.Contains(include, "$") {
				continue
			}
			include = n.resolveBuildPath(include, info.PrefixPath)
			if !filepath.IsAbs(include) {
				include = filepath.Join(configDir, include)
			}
			includeDir := filepath.Clean(filepath.Dir(include))
			if includeFallback == "" {
				includeFallback = includeDir
			}
			switch strings.ToLower(filepath.Base(includeDir)) {
			case "conf.d", "sites-enabled", "sites-available", "servers", "vhosts", "virtualhosts":
				return includeDir, nil
			}
		}
	}
	if includeFallback != "" {
		return includeFallback, nil
	}
	return filepath.Join(configDir, "conf.d"), nil
}

// SiteLayout 检测当前 Nginx 使用的是 conf.d 直写布局还是 Debian 站点布局。
func (n *Nginx) SiteLayout() (*SiteLayout, error) {
	dir, err := n.SiteDir()
	if err != nil {
		return nil, err
	}
	dir, err = filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return nil, err
	}
	layout := &SiteLayout{Mode: "direct", ActiveDir: dir, AvailableDir: dir, EnabledDir: dir}
	switch strings.ToLower(filepath.Base(dir)) {
	case "sites-enabled":
		layout.Mode = "debian"
		layout.EnabledDir = dir
		layout.AvailableDir = filepath.Join(filepath.Dir(dir), "sites-available")
		layout.ActiveDir = dir
	case "sites-available":
		layout.Mode = "debian"
		layout.AvailableDir = dir
		layout.EnabledDir = filepath.Join(filepath.Dir(dir), "sites-enabled")
		layout.ActiveDir = layout.EnabledDir
	}
	return layout, nil
}

// ListSites 列出站点配置目录下的 .conf 文件。
func (n *Nginx) ListSites() ([]string, error) {
	dir, err := n.SiteDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	list := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".conf") {
			list = append(list, entry.Name())
		}
	}
	return list, nil
}

// ReadSite 读取站点配置完整内容。
func (n *Nginx) ReadSite(name string) (string, error) {
	path, err := n.sitePath(name)
	if err != nil {
		return "", err
	}
	content, readErr := n.readTextFile(path)
	if readErr == nil {
		return content, nil
	}
	layout, layoutErr := n.SiteLayout()
	if layoutErr != nil || layout.Mode != "debian" {
		return "", readErr
	}
	name, nameErr := n.normalizeSiteName(name)
	if nameErr != nil {
		return "", nameErr
	}
	return n.readTextFile(filepath.Join(layout.AvailableDir, name))
}

// ListSiteInfo 列出站点配置及其启用状态。
func (n *Nginx) ListSiteInfo() ([]SiteInfo, error) {
	dir, err := n.SiteDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	list := make([]SiteInfo, 0, len(entries))
	known := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		enabled := strings.HasSuffix(strings.ToLower(name), ".conf")
		if strings.HasSuffix(strings.ToLower(name), ".conf.disabled") {
			name = strings.TrimSuffix(name, ".disabled")
		} else if !enabled {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		item, itemErr := n.siteInfo(name, path, enabled)
		if itemErr != nil {
			return nil, itemErr
		}
		list = append(list, item)
		known[name] = struct{}{}
	}
	if strings.EqualFold(filepath.Base(dir), "sites-enabled") {
		availableDir := filepath.Join(filepath.Dir(dir), "sites-available")
		if availableEntries, readErr := os.ReadDir(availableDir); readErr == nil {
			for _, entry := range availableEntries {
				name := entry.Name()
				if entry.IsDir() || !strings.HasSuffix(strings.ToLower(name), ".conf") {
					continue
				}
				if _, exists := known[name]; exists {
					continue
				}
				item, itemErr := n.siteInfo(name, filepath.Join(availableDir, name), false)
				if itemErr != nil {
					return nil, itemErr
				}
				list = append(list, item)
			}
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list, nil
}

// SetSiteEnabled 启用或停用站点。支持 sites-available/sites-enabled 和 .disabled 两种布局。
func (n *Nginx) SetSiteEnabled(name string, enabled bool) error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	name, err := n.normalizeSiteName(name)
	if err != nil {
		return err
	}
	dir, err := n.SiteDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, name)
	backupPaths := []string{path, path + ".disabled"}
	if strings.EqualFold(filepath.Base(dir), "sites-enabled") {
		backupPaths = append(backupPaths, filepath.Join(filepath.Dir(dir), "sites-available", name))
	}
	if enabled {
		if _, statErr := os.Lstat(path); statErr == nil {
			return nil
		} else if !os.IsNotExist(statErr) {
			return statErr
		}
		if disabledPath := path + ".disabled"; n.fileExists(disabledPath) {
			if _, backupErr := n.backupConfig(backupPaths...); backupErr != nil {
				return backupErr
			}
			if err = os.Rename(disabledPath, path); err != nil {
				return fmt.Errorf("启用站点失败: %w", err)
			}
			if testErr := n.Test(); testErr != nil {
				_ = os.Rename(path, disabledPath)
				return testErr
			}
			return nil
		}
		if strings.EqualFold(filepath.Base(dir), "sites-enabled") {
			available := filepath.Join(filepath.Dir(dir), "sites-available", name)
			if !n.fileExists(available) {
				return errors.New("未找到对应的 sites-available 配置")
			}
			if err = os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			if _, backupErr := n.backupConfig(backupPaths...); backupErr != nil {
				return backupErr
			}
			if runtime.GOOS == "windows" {
				err = n.copyFile(available, path)
			} else {
				err = os.Symlink(available, path)
			}
			if err != nil {
				return fmt.Errorf("启用站点失败: %w", err)
			}
			if testErr := n.Test(); testErr != nil {
				_ = os.Remove(path)
				if runtime.GOOS == "windows" {
					_ = n.copyFile(available, path)
				} else {
					_ = os.Symlink(available, path)
				}
				return testErr
			}
			return nil
		}
		return errors.New("未找到可启用的站点配置")
	}

	info, statErr := os.Lstat(path)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return nil
		}
		return statErr
	}
	if info.Mode()&os.ModeSymlink != 0 {
		if _, backupErr := n.backupConfig(backupPaths...); backupErr != nil {
			return backupErr
		}
		linkTarget, linkErr := os.Readlink(path)
		if err = os.Remove(path); err != nil {
			return fmt.Errorf("停用站点失败: %w", err)
		}
		if testErr := n.Test(); testErr != nil {
			if linkErr == nil {
				_ = os.Symlink(linkTarget, path)
			}
			return testErr
		}
		return nil
	}
	disabledPath := path + ".disabled"
	if _, backupErr := n.backupConfig(backupPaths...); backupErr != nil {
		return backupErr
	}
	if err = os.Rename(path, disabledPath); err != nil {
		return fmt.Errorf("停用站点失败: %w", err)
	}
	if testErr := n.Test(); testErr != nil {
		_ = os.Rename(disabledPath, path)
		return testErr
	}
	return nil
}

// EnableSite 启用站点。
func (n *Nginx) EnableSite(name string) error {
	return n.SetSiteEnabled(name, true)
}

// DisableSite 停用站点。
func (n *Nginx) DisableSite(name string) error {
	return n.SetSiteEnabled(name, false)
}

func (n *Nginx) siteInfo(name, path string, enabled bool) (SiteInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return SiteInfo{}, err
	}
	return SiteInfo{Name: name, Path: path, Enabled: enabled, Size: info.Size(), ModifiedAt: info.ModTime()}, nil
}

func (n *Nginx) sitePath(name string) (string, error) {
	dir, err := n.SiteDir()
	if err != nil {
		return "", err
	}
	name, err = n.normalizeSiteName(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

func (n *Nginx) resolveSiteWritePath(path string) (string, error) {
	target, err := os.Readlink(path)
	if err != nil {
		if os.IsNotExist(err) {
			return path, nil
		}
		return path, nil
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("解析站点链接目标失败: %w", err)
	}
	return target, nil
}

func (n *Nginx) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (n *Nginx) copyFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("不能复制目录为站点配置")
	}
	if err = os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err = io.Copy(output, input); err != nil {
		_ = output.Close()
		return err
	}
	return output.Close()
}

type siteState struct {
	Exists bool
	Link   string
	Data   []byte
	Mode   os.FileMode
}

func (n *Nginx) captureSiteState(path string) siteState {
	info, err := os.Lstat(path)
	if err != nil {
		return siteState{Mode: 0644}
	}
	if info.Mode()&os.ModeSymlink != 0 {
		link, linkErr := os.Readlink(path)
		if linkErr == nil {
			return siteState{Exists: true, Link: link, Mode: info.Mode().Perm()}
		}
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		return siteState{Mode: info.Mode().Perm()}
	}
	return siteState{Exists: true, Data: data, Mode: info.Mode().Perm()}
}

func (n *Nginx) restoreSiteStates(states map[string]siteState) error {
	for path, state := range states {
		_ = os.Remove(path)
		if !state.Exists {
			continue
		}
		if state.Link != "" {
			if err := os.Symlink(state.Link, path); err != nil {
				return err
			}
			continue
		}
		if err := n.atomicWrite(path, state.Data, state.Mode); err != nil {
			return err
		}
	}
	return nil
}

func (n *Nginx) writeDebianSite(layout *SiteLayout, name, content string) error {
	if err := os.MkdirAll(layout.AvailableDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(layout.EnabledDir, 0755); err != nil {
		return err
	}
	available := filepath.Join(layout.AvailableDir, name)
	active := filepath.Join(layout.EnabledDir, name)
	paths := []string{available, active}
	if err := n.backupSitePaths(paths...); err != nil {
		return fmt.Errorf("备份站点配置失败: %w", err)
	}
	old := map[string]siteState{available: n.captureSiteState(available), active: n.captureSiteState(active)}
	if err := n.atomicWrite(available, []byte(content), old[available].Mode); err != nil {
		return err
	}
	if _, err := os.Lstat(active); os.IsNotExist(err) {
		if runtime.GOOS == "windows" {
			if err = n.copyFile(available, active); err != nil {
				_ = n.restoreSiteStates(old)
				return err
			}
		} else if err = os.Symlink(available, active); err != nil {
			_ = n.restoreSiteStates(old)
			return err
		}
	} else if err == nil {
		if info, statErr := os.Lstat(active); statErr == nil && info.Mode()&os.ModeSymlink == 0 {
			if err = n.atomicWrite(active, []byte(content), old[active].Mode); err != nil {
				_ = n.restoreSiteStates(old)
				return err
			}
		}
	}
	if err := n.Test(); err != nil {
		_ = n.restoreSiteStates(old)
		return err
	}
	return nil
}

// WriteSite 原子写入站点配置，并在写入后检查整个 Nginx 配置。
// 检查失败时会恢复原文件。
func (n *Nginx) WriteSite(name, content string) error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	if strings.TrimSpace(content) == "" {
		return errors.New("站点配置不能为空")
	}
	dir, err := n.SiteDir()
	if err != nil {
		return err
	}
	name, err = n.normalizeSiteName(name)
	if err != nil {
		return err
	}
	layout, err := n.SiteLayout()
	if err != nil {
		return err
	}
	if layout.Mode == "debian" {
		return n.writeDebianSite(layout, name, content)
	}
	if err = os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建站点配置目录失败: %w", err)
	}
	path := filepath.Join(dir, name)
	writePath, err := n.resolveSiteWritePath(path)
	if err != nil {
		return err
	}
	if err = n.backupSitePaths(writePath); err != nil {
		return fmt.Errorf("备份站点配置失败: %w", err)
	}
	old, oldMode, oldErr := n.readSite(writePath)
	if err = n.atomicWrite(writePath, []byte(content), oldMode); err != nil {
		return err
	}
	if err = n.Test(); err != nil {
		if oldErr == nil {
			_ = n.atomicWrite(writePath, old, oldMode)
		} else {
			_ = os.Remove(writePath)
		}
		return err
	}
	return nil
}

// RemoveSite 删除站点配置，并在删除后检查整个 Nginx 配置。
func (n *Nginx) RemoveSite(name string) error {
	n.operationMu.Lock()
	defer n.operationMu.Unlock()
	dir, err := n.SiteDir()
	if err != nil {
		return err
	}
	name, err = n.normalizeSiteName(name)
	if err != nil {
		return err
	}
	layout, err := n.SiteLayout()
	if err != nil {
		return err
	}
	if layout.Mode == "debian" {
		available := filepath.Join(layout.AvailableDir, name)
		active := filepath.Join(layout.EnabledDir, name)
		if err = n.backupSitePaths(available, active); err != nil {
			return fmt.Errorf("备份站点配置失败: %w", err)
		}
		old := map[string]siteState{available: n.captureSiteState(available), active: n.captureSiteState(active)}
		if err = os.Remove(active); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err = os.Remove(available); err != nil && !os.IsNotExist(err) {
			_ = n.restoreSiteStates(old)
			return err
		}
		if err = n.Test(); err != nil {
			_ = n.restoreSiteStates(old)
			return err
		}
		return nil
	}
	path := filepath.Join(dir, name)
	old, oldMode, err := n.readSite(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	linkTarget, linkErr := os.Readlink(path)
	if err = n.backupSitePaths(path); err != nil {
		return fmt.Errorf("备份站点配置失败: %w", err)
	}
	if err = os.Remove(path); err != nil {
		return err
	}
	if err = n.Test(); err != nil {
		if linkErr == nil {
			_ = os.Symlink(linkTarget, path)
		} else {
			_ = n.atomicWrite(path, old, oldMode)
		}
		return err
	}
	return nil
}

func (n *Nginx) validateListenPort(listen int) error {
	if listen < 1 || listen > 65535 {
		return errors.New("监听端口不合法")
	}
	return nil
}

func (n *Nginx) normalizeServerName(serverName string) (string, error) {
	serverName = strings.TrimSpace(serverName)
	if serverName == "" {
		return "_", nil
	}
	if strings.ContainsAny(serverName, " \t\r\n{};\"/\\") {
		return "", errors.New("站点名称不合法")
	}
	return serverName, nil
}

func (n *Nginx) proxyURL(target string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(target))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, errors.New("反向代理地址必须是 http 或 https URL")
	}
	return parsed, nil
}

func (n *Nginx) normalizeIdentifier(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, " \t\r\n{};\"/\\:") {
		return "", errors.New("Nginx 标识不合法")
	}
	return value, nil
}

func (n *Nginx) fastCGIEndpoint(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, " \t\r\n{};\"") {
		return "", errors.New("FastCGI 地址不合法")
	}
	if strings.HasPrefix(value, "unix:") || strings.Contains(value, ":") {
		return value, nil
	}
	return "127.0.0.1:" + value, nil
}

func (n *Nginx) nginxConfigPath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, "\r\n{};") {
		return "", errors.New("Nginx 路径不合法")
	}
	path, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("解析 Nginx 路径失败: %w", err)
	}
	path = filepath.ToSlash(filepath.Clean(path))
	path = strings.ReplaceAll(path, `"`, `\"`)
	if strings.ContainsAny(path, " \t") {
		return `"` + path + `"`, nil
	}
	return path, nil
}

// BuildReverseProxy 生成一个支持 WebSocket 的反向代理 server 配置。
func (n *Nginx) BuildReverseProxy(serverName, target string, listen int) (string, error) {
	return n.BuildReverseProxyConfig(SiteDefinition{
		ServerName: serverName,
		Listen:     listen,
		ProxyPass:  target,
	})
}

// BuildStaticSite 生成一个静态站点 server 配置。
func (n *Nginx) BuildStaticSite(serverName, root string, listen int) (string, error) {
	if err := n.validateListenPort(listen); err != nil {
		return "", err
	}
	serverName, err := n.normalizeServerName(serverName)
	if err != nil {
		return "", err
	}
	root = strings.TrimSpace(root)
	if root == "" {
		return "", errors.New("网站根目录不能为空")
	}
	root, err = n.nginxConfigPath(root)
	if err != nil {
		return "", fmt.Errorf("解析网站根目录失败: %w", err)
	}
	return fmt.Sprintf(`server {
    listen %d;
    server_name %s;
    root %s;
    index index.html index.htm;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
`, listen, serverName, root), nil
}

// BuildSiteConfig 根据结构化定义生成站点配置。
func (n *Nginx) BuildSiteConfig(definition SiteDefinition) (string, error) {
	listen := definition.Listen
	if listen == 0 {
		listen = 80
		if definition.TLS {
			listen = 443
		}
	}
	if err := n.validateListenPort(listen); err != nil {
		return "", err
	}
	serverName, err := n.normalizeServerName(definition.ServerName)
	if err != nil {
		return "", err
	}
	typeName := strings.ToLower(strings.TrimSpace(definition.Type))
	if definition.RedirectHTTPS {
		typeName = "redirect"
	}
	if typeName == "" {
		switch {
		case definition.ProxyPass != "":
			typeName = "proxy"
		case definition.FastCGIPass != "":
			typeName = "fastcgi"
		default:
			typeName = "static"
		}
	}
	if typeName != "static" && typeName != "proxy" && typeName != "fastcgi" && typeName != "redirect" {
		return "", errors.New("站点类型必须是 static、proxy、fastcgi 或 redirect")
	}
	if typeName == "redirect" {
		targetPort := definition.RedirectPort
		if targetPort == 0 {
			targetPort = 443
		}
		return n.BuildRedirectSite(serverName, listen, targetPort)
	}
	var cert, key string
	if definition.TLS {
		if err = n.ValidateTLSFiles(definition.CertificatePath, definition.KeyPath); err != nil {
			return "", err
		}
		cert, err = n.nginxConfigPath(definition.CertificatePath)
		if err != nil {
			return "", err
		}
		key, err = n.nginxConfigPath(definition.KeyPath)
		if err != nil {
			return "", err
		}
	}
	lines := []string{"server {"}
	listenLine := fmt.Sprintf("    listen %d", listen)
	if definition.TLS {
		listenLine += " ssl"
		if definition.HTTP2 {
			listenLine += " http2"
		}
	}
	lines = append(lines, listenLine+";", "    server_name "+serverName+";")
	if definition.TLS {
		lines = append(lines, "    ssl_certificate "+cert+";", "    ssl_certificate_key "+key+";", "    ssl_protocols TLSv1.2 TLSv1.3;")
	}
	if definition.AccessLogEnabled && definition.AccessLog != "" {
		accessLog, pathErr := n.nginxConfigPath(definition.AccessLog)
		if pathErr != nil {
			return "", pathErr
		}
		lines = append(lines, "    access_log "+accessLog+";")
	}
	if definition.ErrorLogEnabled && definition.ErrorLog != "" {
		errorLog, pathErr := n.nginxConfigPath(definition.ErrorLog)
		if pathErr != nil {
			return "", pathErr
		}
		lines = append(lines, "    error_log "+errorLog+";")
	}
	if definition.ClientMaxBodySize != "" {
		if err = n.validateDirectiveValue(definition.ClientMaxBodySize); err != nil {
			return "", err
		}
		lines = append(lines, "    client_max_body_size "+definition.ClientMaxBodySize+";")
	}
	if definition.Gzip {
		lines = append(lines, "    gzip on;")
	}
	for key, value := range definition.Headers {
		if strings.TrimSpace(key) == "" || strings.ContainsAny(key, " \t\r\n{};") || strings.ContainsAny(value, "\r\n{};") {
			return "", errors.New("响应头参数不合法")
		}
		lines = append(lines, fmt.Sprintf("    add_header %s %q always;", key, value))
	}
	for _, directive := range definition.ExtraDirectives {
		directive = strings.TrimSpace(directive)
		if directive == "" || strings.ContainsAny(directive, "\r\n") {
			return "", errors.New("扩展指令必须是单行非空内容")
		}
		lines = append(lines, "    "+directive)
	}
	switch typeName {
	case "static":
		root, rootErr := n.nginxConfigPath(definition.Root)
		if rootErr != nil {
			return "", rootErr
		}
		lines = append(lines, "    root "+root+";", "    index index.html index.htm;", "", "    location / {", "        try_files $uri $uri/ /index.html;", "    }")
	case "proxy":
		target, targetErr := n.proxyURL(definition.ProxyPass)
		if targetErr != nil {
			return "", targetErr
		}
		proxyHeaders := map[string]string{
			"Host":              "$host",
			"X-Real-IP":         "$remote_addr",
			"X-Forwarded-For":   "$proxy_add_x_forwarded_for",
			"X-Forwarded-Proto": "$scheme",
			"Upgrade":           "$http_upgrade",
			"Connection":        "upgrade",
		}
		if strings.TrimSpace(definition.ProxyHost) != "" {
			proxyHeaders["Host"] = definition.ProxyHost
		}
		for key, value := range definition.ProxyHeaders {
			if err = n.validateProxyHeaderName(key); err != nil {
				return "", err
			}
			if err = n.validateProxyValue(value); err != nil {
				return "", err
			}
			proxyHeaders[key] = value
		}
		if len(definition.SubFilters) > 0 {
			if _, exists := proxyHeaders["Accept-Encoding"]; !exists {
				// sub_filter 需要拿到未压缩的响应正文，否则上游 gzip 内容无法替换。
				proxyHeaders["Accept-Encoding"] = ""
			}
		}
		headerNames := make([]string, 0, len(proxyHeaders))
		for key := range proxyHeaders {
			headerNames = append(headerNames, key)
		}
		sort.Strings(headerNames)
		lines = append(lines, "", "    location / {", "        proxy_pass "+target.String()+";", "        proxy_http_version 1.1;")
		for _, key := range headerNames {
			value, valueErr := n.nginxArgument(proxyHeaders[key])
			if valueErr != nil {
				return "", valueErr
			}
			lines = append(lines, "        proxy_set_header "+key+" "+value+";")
		}
		if definition.ProxySetBody != "" {
			value, valueErr := n.nginxArgument(definition.ProxySetBody)
			if valueErr != nil {
				return "", valueErr
			}
			lines = append(lines, "        proxy_set_body "+value+";")
		}
		proxyDirectives := []struct {
			name  string
			value string
		}{
			{name: "proxy_redirect", value: definition.ProxyRedirect},
			{name: "proxy_cookie_domain", value: definition.ProxyCookieDomain},
			{name: "proxy_cookie_path", value: definition.ProxyCookiePath},
		}
		for _, directive := range proxyDirectives {
			value := directive.value
			if strings.TrimSpace(value) == "" {
				continue
			}
			if err = n.validateProxyDirective(value); err != nil {
				return "", fmt.Errorf("%s 参数不合法: %w", directive.name, err)
			}
			lines = append(lines, "        "+directive.name+" "+strings.TrimSpace(value)+";")
		}
		if len(definition.SubFilters) > 0 {
			for _, filter := range definition.SubFilters {
				search, searchErr := n.nginxQuote(filter.Search)
				if searchErr != nil {
					return "", fmt.Errorf("sub_filter 匹配内容不合法: %w", searchErr)
				}
				replace, replaceErr := n.nginxQuote(filter.Replace)
				if replaceErr != nil {
					return "", fmt.Errorf("sub_filter 替换内容不合法: %w", replaceErr)
				}
				lines = append(lines, "        sub_filter "+search+" "+replace+";")
				for _, filterType := range filter.Types {
					filterType = strings.TrimSpace(filterType)
					if filterType == "" || strings.ContainsAny(filterType, " \t\r\n{};\"") {
						return "", errors.New("sub_filter 类型不合法")
					}
				}
			}
			if definition.SubFilterOnce != nil {
				value := "off"
				if *definition.SubFilterOnce {
					value = "on"
				}
				lines = append(lines, "        sub_filter_once "+value+";")
			}
			types := make([]string, 0)
			seenTypes := make(map[string]struct{})
			for _, filter := range definition.SubFilters {
				for _, filterType := range filter.Types {
					filterType = strings.TrimSpace(filterType)
					if filterType == "" {
						continue
					}
					if _, exists := seenTypes[filterType]; exists {
						continue
					}
					seenTypes[filterType] = struct{}{}
					types = append(types, filterType)
				}
			}
			if len(types) > 0 {
				sort.Strings(types)
				lines = append(lines, "        sub_filter_types "+strings.Join(types, " ")+";")
			}
		}
		for directive, value := range map[string]string{"proxy_read_timeout": definition.ProxyReadTimeout, "proxy_connect_timeout": definition.ProxyConnectTimeout, "proxy_send_timeout": definition.ProxySendTimeout} {
			if value == "" {
				continue
			}
			if err = n.validateDirectiveValue(value); err != nil {
				return "", err
			}
			lines = append(lines, "        "+directive+" "+value+";")
		}
		lines = append(lines, "    }")
	case "fastcgi":
		root, rootErr := n.nginxConfigPath(definition.Root)
		if rootErr != nil {
			return "", rootErr
		}
		endpoint, endpointErr := n.fastCGIEndpoint(definition.FastCGIPass)
		if endpointErr != nil {
			return "", endpointErr
		}
		lines = append(lines, "    root "+root+";", "    index index.php index.html;", "", "    location / {", "        try_files $uri $uri/ /index.php?$query_string;", "    }", "", "    location ~ \\.php$ {", "        try_files $uri =404;", "        include fastcgi_params;", "        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;", "        fastcgi_pass "+endpoint+";", "    }")
	}
	lines = append(lines, "}", "")
	return strings.Join(lines, "\n"), nil
}

// WriteSiteConfig 写入结构化站点配置。
func (n *Nginx) WriteSiteConfig(definition SiteDefinition) error {
	name := strings.TrimSpace(definition.Name)
	if name == "" {
		return errors.New("站点名称不能为空")
	}
	content, err := n.BuildSiteConfig(definition)
	if err != nil {
		return err
	}
	return n.WriteSite(name, content)
}

// BuildReverseProxyConfig 根据结构化定义生成反向代理配置。
// 相比 BuildReverseProxy，该方法支持自定义发送域名、请求头、请求体和响应内容替换。
func (n *Nginx) BuildReverseProxyConfig(definition SiteDefinition) (string, error) {
	definition.Type = "proxy"
	return n.BuildSiteConfig(definition)
}

// WriteReverseProxyConfig 写入结构化反向代理配置。
func (n *Nginx) WriteReverseProxyConfig(definition SiteDefinition) error {
	definition.Type = "proxy"
	return n.WriteSiteConfig(definition)
}

// BuildAdvancedProxy 生成带缓存、限流和访问控制的反向代理配置。
func (n *Nginx) BuildAdvancedProxy(options AdvancedProxyOptions) (string, error) {
	definition := options.Site
	definition.Type = "proxy"
	for _, value := range options.Allow {
		if strings.TrimSpace(value) == "" || strings.ContainsAny(value, " \t\r\n{};\"") {
			return "", errors.New("allow 地址不合法")
		}
		definition.ExtraDirectives = append(definition.ExtraDirectives, "allow "+value+";")
	}
	for _, value := range options.Deny {
		if strings.TrimSpace(value) == "" || strings.ContainsAny(value, " \t\r\n{};\"") {
			return "", errors.New("deny 地址不合法")
		}
		definition.ExtraDirectives = append(definition.ExtraDirectives, "deny "+value+";")
	}
	if options.CacheZone != "" {
		zone, err := n.normalizeIdentifier(options.CacheZone)
		if err != nil {
			return "", err
		}
		definition.ExtraDirectives = append(definition.ExtraDirectives, "proxy_cache "+zone+";")
		if options.CacheValid != "" {
			if err = n.validateDirectiveValue(options.CacheValid); err != nil {
				return "", err
			}
			definition.ExtraDirectives = append(definition.ExtraDirectives, "proxy_cache_valid "+options.CacheValid+";")
		}
	}
	if options.RateLimitZone != "" {
		zone, err := n.normalizeIdentifier(options.RateLimitZone)
		if err != nil {
			return "", err
		}
		if options.RateLimitBurst < 0 {
			return "", errors.New("限流 burst 不能为负数")
		}
		line := "limit_req zone=" + zone
		if options.RateLimitBurst > 0 {
			line += " burst=" + strconv.Itoa(options.RateLimitBurst)
		}
		line += " nodelay;"
		definition.ExtraDirectives = append(definition.ExtraDirectives, line)
	}
	if definition.TLS {
		definition.Headers = n.cloneStringMap(definition.Headers)
		if _, exists := definition.Headers["Strict-Transport-Security"]; !exists {
			definition.Headers["Strict-Transport-Security"] = "max-age=31536000; includeSubDomains"
		}
	}
	return n.BuildSiteConfig(definition)
}

// WriteAdvancedProxy 写入高级反向代理站点配置。
func (n *Nginx) WriteAdvancedProxy(options AdvancedProxyOptions) error {
	definition := options.Site
	content, err := n.BuildAdvancedProxy(options)
	if err != nil {
		return err
	}
	if strings.TrimSpace(definition.Name) == "" {
		return errors.New("站点名称不能为空")
	}
	return n.WriteSite(definition.Name, content)
}

func (n *Nginx) cloneStringMap(source map[string]string) map[string]string {
	result := make(map[string]string, len(source)+1)
	for key, value := range source {
		result[key] = value
	}
	return result
}

// ReadSiteConfig 读取并解析常用站点指令，无法识别的内容仍保留在原始配置中供高级编辑使用。
func (n *Nginx) ReadSiteConfig(name string) (*SiteDefinition, error) {
	content, err := n.ReadSite(name)
	if err != nil {
		return nil, err
	}
	definition := &SiteDefinition{Name: name, Type: "static", Listen: 80, Headers: make(map[string]string)}
	var subFilterTypes []string
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if value := n.directiveValue(line, "listen"); value != "" {
			fields := strings.Fields(value)
			if len(fields) > 0 {
				if port, parseErr := strconv.Atoi(fields[0]); parseErr == nil {
					definition.Listen = port
				}
			}
			definition.TLS = strings.Contains(value, " ssl") || strings.HasSuffix(value, " ssl")
			definition.HTTP2 = strings.Contains(value, "http2")
		}
		if value := n.directiveValue(line, "server_name"); value != "" {
			definition.ServerName = strings.Fields(value)[0]
		}
		if value := n.directiveValue(line, "root"); value != "" {
			definition.Root = strings.Trim(value, `"'`)
		}
		if value := n.directiveValue(line, "proxy_pass"); value != "" {
			definition.ProxyPass = strings.Trim(value, `"'`)
			definition.Type = "proxy"
		}
		if value := n.directiveValue(line, "fastcgi_pass"); value != "" {
			definition.FastCGIPass = strings.Trim(value, `"'`)
			definition.Type = "fastcgi"
		}
		if value := n.directiveValue(line, "ssl_certificate"); value != "" {
			definition.CertificatePath = strings.Trim(value, `"'`)
		}
		if value := n.directiveValue(line, "ssl_certificate_key"); value != "" {
			definition.KeyPath = strings.Trim(value, `"'`)
		}
		if args := n.directiveArguments(line, "access_log"); len(args) > 0 {
			definition.AccessLog = strings.Trim(args[0], `"'`)
			definition.AccessLogEnabled = !strings.EqualFold(definition.AccessLog, "off")
		}
		if args := n.directiveArguments(line, "error_log"); len(args) > 0 {
			definition.ErrorLog = strings.Trim(args[0], `"'`)
			definition.ErrorLogEnabled = !strings.EqualFold(definition.ErrorLog, "off")
		}
		if value := n.directiveValue(line, "client_max_body_size"); value != "" {
			definition.ClientMaxBodySize = value
		}
		if value := n.directiveValue(line, "gzip"); strings.EqualFold(value, "on") {
			definition.Gzip = true
		}
		if args := n.directiveArguments(line, "add_header"); len(args) >= 2 {
			values := args[1:]
			if len(values) > 1 && strings.EqualFold(values[len(values)-1], "always") {
				values = values[:len(values)-1]
			}
			definition.Headers[args[0]] = strings.Join(values, " ")
		}
		if args := n.directiveArguments(line, "proxy_set_header"); len(args) >= 2 {
			if definition.ProxyHeaders == nil {
				definition.ProxyHeaders = make(map[string]string)
			}
			value := strings.Join(args[1:], " ")
			definition.ProxyHeaders[args[0]] = value
			if strings.EqualFold(args[0], "Host") && value != "$host" {
				definition.ProxyHost = value
			}
			definition.Type = "proxy"
		}
		if args := n.directiveArguments(line, "proxy_set_body"); len(args) > 0 {
			definition.ProxySetBody = strings.Join(args, " ")
			definition.Type = "proxy"
		}
		if value := n.directiveValue(line, "proxy_redirect"); value != "" {
			definition.ProxyRedirect = strings.TrimSpace(value)
			definition.Type = "proxy"
		}
		if value := n.directiveValue(line, "proxy_cookie_domain"); value != "" {
			definition.ProxyCookieDomain = strings.TrimSpace(value)
			definition.Type = "proxy"
		}
		if value := n.directiveValue(line, "proxy_cookie_path"); value != "" {
			definition.ProxyCookiePath = strings.TrimSpace(value)
			definition.Type = "proxy"
		}
		if args := n.directiveArguments(line, "sub_filter"); len(args) >= 2 {
			definition.SubFilters = append(definition.SubFilters, ProxySubFilter{Search: args[0], Replace: args[1]})
			definition.Type = "proxy"
		}
		if value := n.directiveValue(line, "sub_filter_once"); value != "" {
			once := strings.EqualFold(strings.TrimSpace(value), "on")
			definition.SubFilterOnce = &once
			definition.Type = "proxy"
		}
		if value := n.directiveValue(line, "sub_filter_types"); value != "" {
			subFilterTypes = append(subFilterTypes, strings.Fields(value)...)
			definition.Type = "proxy"
		}
		if strings.HasPrefix(line, "return 301 https://") {
			definition.RedirectHTTPS = true
			definition.Type = "redirect"
		}
	}
	if len(subFilterTypes) > 0 && len(definition.SubFilters) > 0 {
		seen := make(map[string]struct{}, len(subFilterTypes))
		for _, filterType := range subFilterTypes {
			if _, exists := seen[filterType]; exists {
				continue
			}
			seen[filterType] = struct{}{}
			for index := range definition.SubFilters {
				definition.SubFilters[index].Types = append(definition.SubFilters[index].Types, filterType)
			}
		}
	}
	return definition, nil
}

// UpdateSiteConfig 更新结构化站点配置。
func (n *Nginx) UpdateSiteConfig(definition SiteDefinition) error {
	return n.WriteSiteConfig(definition)
}

func (n *Nginx) directiveValue(line, directive string) string {
	line = strings.TrimSpace(strings.TrimSuffix(line, ";"))
	prefix := directive + " "
	if !strings.HasPrefix(line, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(line, prefix))
}

func (n *Nginx) validateDirectiveValue(value string) error {
	if strings.TrimSpace(value) == "" || strings.ContainsAny(value, "\r\n{};\"") {
		return errors.New("Nginx 指令参数不合法")
	}
	return nil
}

func (n *Nginx) validateProxyHeaderName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || strings.ContainsAny(name, " \t\r\n{};\"'\\:") {
		return errors.New("代理请求头名称不合法")
	}
	return nil
}

func (n *Nginx) validateProxyValue(value string) error {
	if strings.ContainsAny(value, "\x00\r\n") {
		return errors.New("代理参数不能包含控制字符")
	}
	return nil
}

func (n *Nginx) validateProxyDirective(value string) error {
	if strings.TrimSpace(value) == "" || strings.ContainsAny(value, "\x00\r\n{};") {
		return errors.New("代理指令参数不合法")
	}
	return nil
}

func (n *Nginx) nginxQuote(value string) (string, error) {
	if err := n.validateProxyValue(value); err != nil {
		return "", err
	}
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return `"` + value + `"`, nil
}

func (n *Nginx) nginxArgument(value string) (string, error) {
	if err := n.validateProxyValue(value); err != nil {
		return "", err
	}
	if value == "" {
		return `""`, nil
	}
	if !strings.ContainsAny(value, " \t;\"'") {
		return value, nil
	}
	return n.nginxQuote(value)
}

func (n *Nginx) directiveArguments(line, directive string) []string {
	value := n.directiveValue(line, directive)
	if value == "" {
		return nil
	}
	var result []string
	var token strings.Builder
	var quote rune
	escaped := false
	tokenStarted := false
	flush := func() {
		if !tokenStarted {
			return
		}
		result = append(result, token.String())
		token.Reset()
		tokenStarted = false
	}
	for _, char := range value {
		if escaped {
			token.WriteRune(char)
			tokenStarted = true
			escaped = false
			continue
		}
		if char == '\\' {
			tokenStarted = true
			escaped = true
			continue
		}
		if quote != 0 {
			if char == quote {
				quote = 0
			} else {
				token.WriteRune(char)
			}
			continue
		}
		if char == '\'' || char == '"' {
			tokenStarted = true
			quote = char
			continue
		}
		if char == ' ' || char == '\t' {
			flush()
			continue
		}
		token.WriteRune(char)
		tokenStarted = true
	}
	if escaped {
		token.WriteByte('\\')
		tokenStarted = true
	}
	flush()
	return result
}

// BuildUpstream 生成负载均衡 upstream 配置。
func (n *Nginx) BuildUpstream(name string, servers []UpstreamServer) (string, error) {
	name, err := n.normalizeIdentifier(name)
	if err != nil {
		return "", err
	}
	if len(servers) == 0 {
		return "", errors.New("upstream 至少需要一个后端节点")
	}
	lines := make([]string, 0, len(servers)+2)
	lines = append(lines, "upstream "+name+" {")
	for _, server := range servers {
		address := strings.TrimSpace(server.Address)
		if address == "" || strings.ContainsAny(address, " \t\r\n{};\"") {
			return "", errors.New("upstream 节点地址不合法")
		}
		line := "    server " + address
		if server.Weight > 0 {
			line += " weight=" + strconv.Itoa(server.Weight)
		}
		if server.MaxFails > 0 {
			line += " max_fails=" + strconv.Itoa(server.MaxFails)
		}
		if server.FailTimeout != "" {
			if strings.ContainsAny(server.FailTimeout, " \t\r\n{};\"") {
				return "", errors.New("upstream 失败超时参数不合法")
			}
			line += " fail_timeout=" + server.FailTimeout
		}
		if server.Backup {
			line += " backup"
		}
		if server.Down {
			line += " down"
		}
		lines = append(lines, line+";")
	}
	lines = append(lines, "}", "")
	return strings.Join(lines, "\n"), nil
}

// WriteUpstream 写入 upstream 配置并检查整个 Nginx 配置。
func (n *Nginx) WriteUpstream(name string, servers []UpstreamServer) error {
	content, err := n.BuildUpstream(name, servers)
	if err != nil {
		return err
	}
	identifier, err := n.normalizeIdentifier(name)
	if err != nil {
		return err
	}
	return n.WriteSite("upstream-"+identifier, content)
}

// BuildLoadBalancedProxy 生成指向 upstream 的反向代理配置。
func (n *Nginx) BuildLoadBalancedProxy(serverName, upstream string, listen int) (string, error) {
	upstream, err := n.normalizeIdentifier(upstream)
	if err != nil {
		return "", err
	}
	return n.BuildReverseProxy(serverName, "http://"+upstream, listen)
}

// WriteLoadBalancedProxy 写入指向 upstream 的反向代理站点。
func (n *Nginx) WriteLoadBalancedProxy(name, serverName, upstream string, listen int) error {
	content, err := n.BuildLoadBalancedProxy(serverName, upstream, listen)
	if err != nil {
		return err
	}
	return n.WriteSite(name, content)
}

// BuildFastCGI 生成 PHP 等 FastCGI 应用的站点配置。
func (n *Nginx) BuildFastCGI(serverName, root, upstream string, listen int) (string, error) {
	if err := n.validateListenPort(listen); err != nil {
		return "", err
	}
	serverName, err := n.normalizeServerName(serverName)
	if err != nil {
		return "", err
	}
	upstream, err = n.fastCGIEndpoint(upstream)
	if err != nil {
		return "", err
	}
	root, err = n.nginxConfigPath(root)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`server {
    listen %d;
    server_name %s;
    root %s;
    index index.php index.html;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        try_files $uri =404;
        include fastcgi_params;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        fastcgi_pass %s;
    }
}
`, listen, serverName, root, upstream), nil
}

// WriteFastCGI 写入 FastCGI 站点配置。
func (n *Nginx) WriteFastCGI(name, serverName, root, upstream string, listen int) error {
	content, err := n.BuildFastCGI(serverName, root, upstream, listen)
	if err != nil {
		return err
	}
	return n.WriteSite(name, content)
}

// BuildRedirectSite 生成 HTTP 跳转到 HTTPS 的站点配置。
func (n *Nginx) BuildRedirectSite(serverName string, listen, targetPort int) (string, error) {
	if err := n.validateListenPort(listen); err != nil {
		return "", err
	}
	if err := n.validateListenPort(targetPort); err != nil {
		return "", err
	}
	serverName, err := n.normalizeServerName(serverName)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`server {
    listen %d;
    server_name %s;
    return 301 https://$host:%d$request_uri;
}
`, listen, serverName, targetPort), nil
}

// BuildStreamProxy 生成 TCP/UDP stream 代理配置。该配置必须被主配置在 stream 上下文中 include。
func (n *Nginx) BuildStreamProxy(options StreamProxyOptions) (string, error) {
	if err := n.validateListenPort(options.Listen); err != nil {
		return "", err
	}
	target := strings.TrimSpace(options.Target)
	if target == "" || strings.ContainsAny(target, " \t\r\n{};\"") {
		return "", errors.New("stream 后端地址不合法")
	}
	if _, _, err := net.SplitHostPort(target); err != nil {
		return "", errors.New("stream 后端地址必须是 host:port")
	}
	protocol := ""
	if options.UDP {
		protocol = " udp"
	}
	lines := []string{"stream {", "    server {", fmt.Sprintf("        listen %d%s;", options.Listen, protocol), "        proxy_pass " + target + ";"}
	if options.Timeout != "" {
		if err := n.validateDirectiveValue(options.Timeout); err != nil {
			return "", err
		}
		lines = append(lines, "        proxy_timeout "+options.Timeout+";")
	}
	lines = append(lines, "    }", "}", "")
	return strings.Join(lines, "\n"), nil
}

// WriteStreamProxy 写入 stream 配置文件。path 必须被主配置在 stream 上下文中 include。
func (n *Nginx) WriteStreamProxy(path string, options StreamProxyOptions) error {
	content, err := n.BuildStreamProxy(options)
	if err != nil {
		return err
	}
	return n.WriteConfigFile(path, content)
}

// WriteRedirectSite 写入 HTTP 到 HTTPS 跳转站点。
func (n *Nginx) WriteRedirectSite(name, serverName string, listen, targetPort int) error {
	content, err := n.BuildRedirectSite(serverName, listen, targetPort)
	if err != nil {
		return err
	}
	return n.WriteSite(name, content)
}

// WriteReverseProxy 写入反向代理站点配置并检查配置。
func (n *Nginx) WriteReverseProxy(name, serverName, target string, listen int) error {
	content, err := n.BuildReverseProxy(serverName, target, listen)
	if err != nil {
		return err
	}
	return n.WriteSite(name, content)
}

// WriteStaticSite 写入静态站点配置并检查整个 Nginx 配置。
func (n *Nginx) WriteStaticSite(name, serverName, root string, listen int) error {
	content, err := n.BuildStaticSite(serverName, root, listen)
	if err != nil {
		return err
	}
	return n.WriteSite(name, content)
}

func (n *Nginx) binary() (string, error) {
	n.mu.RLock()
	binaryPath := n.binaryPath
	n.mu.RUnlock()
	if binaryPath != "" {
		if n.executableFile(binaryPath) {
			return binaryPath, nil
		}
		if resolved := n.commandPath(binaryPath); resolved != "" {
			n.mu.Lock()
			n.binaryPath = resolved
			n.mu.Unlock()
			return resolved, nil
		}
		return "", fmt.Errorf("nginx 可执行文件不存在: %s", binaryPath)
	}
	if binaryPath := n.commandPath("nginx"); binaryPath != "" {
		n.mu.Lock()
		n.binaryPath = binaryPath
		n.mu.Unlock()
		return binaryPath, nil
	}
	for _, candidate := range n.nginxBinaryPaths() {
		if n.executableFile(candidate) {
			n.mu.Lock()
			n.binaryPath = candidate
			n.mu.Unlock()
			return candidate, nil
		}
	}
	return "", errors.New("未找到 nginx，请先安装或传入 nginx 可执行文件路径")
}

func (n *Nginx) executableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode()&0111 != 0
}

func (n *Nginx) commandTimeout() time.Duration {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.timeout <= 0 {
		return defaultTimeout
	}
	return n.timeout
}

func (n *Nginx) runtimeArgs(info *Info, operation ...string) []string {
	n.mu.RLock()
	globalArgs := append([]string(nil), n.globalArgs...)
	n.mu.RUnlock()
	args := make([]string, 0, 5+len(globalArgs)+len(operation))
	if info.PrefixPath != "" {
		args = append(args, "-p", info.PrefixPath)
	}
	if info.ConfigPath != "" {
		args = append(args, "-c", info.ConfigPath)
	}
	if len(globalArgs) > 0 {
		args = append(args, "-g", strings.Join(globalArgs, " "))
	}
	return append(args, operation...)
}

func (n *Nginx) defaultPidPath(info *Info) string {
	if info.PrefixPath != "" {
		return filepath.Join(info.PrefixPath, "logs", "nginx.pid")
	}
	if info.ConfigPath != "" {
		return filepath.Join(filepath.Dir(info.ConfigPath), "nginx.pid")
	}
	return ""
}

func (n *Nginx) nginxBinaryPaths() []string {
	if runtime.GOOS == "windows" {
		paths := []string{"nginx.exe", `C:\nginx\nginx.exe`}
		for _, key := range []string{"ProgramFiles", "ProgramFiles(x86)"} {
			if value := os.Getenv(key); value != "" {
				paths = append(paths, filepath.Join(value, "nginx", "nginx.exe"))
			}
		}
		return paths
	}
	paths := []string{"/usr/sbin/nginx", "/usr/local/sbin/nginx", "/usr/local/nginx/sbin/nginx", "/opt/homebrew/bin/nginx", "/usr/local/bin/nginx"}
	return paths
}

func (n *Nginx) packageCommands(action string) ([]packageCommand, error) {
	install := action == "install"
	if action != "install" && action != "uninstall" {
		return nil, errors.New("不支持的软件包操作")
	}
	switch runtime.GOOS {
	case "linux":
		manager := n.detectPackageManager()
		switch manager {
		case "apt-get":
			if install {
				return []packageCommand{{"apt-get", []string{"-o", "DPkg::Lock::Timeout=60", "update"}}, {"apt-get", []string{"-o", "DPkg::Lock::Timeout=60", "install", "-y", "--no-install-recommends", "nginx"}}}, nil
			}
			return []packageCommand{{"apt-get", []string{"-o", "DPkg::Lock::Timeout=60", "remove", "-y", "nginx"}}}, nil
		case "apt":
			if install {
				return []packageCommand{{"apt", []string{"-o", "DPkg::Lock::Timeout=60", "update"}}, {"apt", []string{"-o", "DPkg::Lock::Timeout=60", "install", "-y", "--no-install-recommends", "nginx"}}}, nil
			}
			return []packageCommand{{"apt", []string{"-o", "DPkg::Lock::Timeout=60", "remove", "-y", "nginx"}}}, nil
		case "nala":
			verb := "remove"
			if install {
				verb = "install"
			}
			return []packageCommand{{"nala", []string{verb, "-y", "nginx"}}}, nil
		case "aptitude":
			verb := "remove"
			if install {
				verb = "install"
			}
			return []packageCommand{{"aptitude", []string{"-y", verb, "nginx"}}}, nil
		case "dnf", "dnf5", "yum", "microdnf", "tdnf":
			verb := "remove"
			if install {
				verb = "install"
			}
			return []packageCommand{{manager, []string{verb, "-y", "nginx"}}}, nil
		case "apk":
			verb := "del"
			args := []string{"del", "nginx"}
			if install {
				verb = "add"
				args = []string{verb, "--no-cache", "nginx"}
			}
			return []packageCommand{{"apk", args}}, nil
		case "zypper":
			verb := "remove"
			if install {
				verb = "install"
			}
			return []packageCommand{{"zypper", []string{"--non-interactive", verb, "nginx"}}}, nil
		case "pacman":
			if install {
				return []packageCommand{{"pacman", []string{"-S", "--needed", "--noconfirm", "nginx"}}}, nil
			}
			return []packageCommand{{"pacman", []string{"-Rns", "--noconfirm", "nginx"}}}, nil
		case "emerge":
			if install {
				return []packageCommand{{"emerge", []string{"--quiet-build", "www-servers/nginx"}}}, nil
			}
			return []packageCommand{{"emerge", []string{"--unmerge", "www-servers/nginx"}}}, nil
		case "nix-env":
			if install {
				return []packageCommand{{"nix-env", []string{"-iA", "nixpkgs.nginx"}}}, nil
			}
			return []packageCommand{{"nix-env", []string{"-e", "nginx"}}}, nil
		case "guix":
			verb := "remove"
			if install {
				verb = "install"
			}
			return []packageCommand{{"guix", []string{verb, "nginx"}}}, nil
		case "eopkg":
			verb := "remove"
			if install {
				verb = "install"
			}
			return []packageCommand{{"eopkg", []string{verb, "-y", "nginx"}}}, nil
		case "urpmi":
			if install {
				return []packageCommand{{"urpmi", []string{"--auto", "nginx"}}}, nil
			}
			if n.commandExists("urpme") {
				return []packageCommand{{"urpme", []string{"nginx"}}}, nil
			}
			return nil, errors.New("检测到 urpmi，但未找到 urpme 卸载命令")
		case "slackpkg":
			if install {
				return []packageCommand{{"slackpkg", []string{"install", "nginx"}}}, nil
			}
			if n.commandExists("removepkg") {
				return []packageCommand{{"removepkg", []string{"nginx"}}}, nil
			}
			return nil, errors.New("检测到 slackpkg，但未找到 removepkg 卸载命令")
		case "xbps-install":
			if install {
				return []packageCommand{{"xbps-install", []string{"-Sy", "-y", "nginx"}}}, nil
			}
			return []packageCommand{{"xbps-remove", []string{"-y", "nginx"}}}, nil
		case "opkg":
			if install {
				return []packageCommand{{"opkg", []string{"update"}}, {"opkg", []string{"install", "nginx"}}}, nil
			}
			return []packageCommand{{"opkg", []string{"remove", "nginx"}}}, nil
		case "brew":
			verb := "uninstall"
			if install {
				verb = "install"
			}
			return []packageCommand{{"brew", []string{verb, "nginx"}}}, nil
		default:
			return nil, errors.New("Linux 系统未找到可用的软件包管理器")
		}
	case "darwin":
		if n.commandExists("brew") {
			verb := "uninstall"
			if install {
				verb = "install"
			}
			return []packageCommand{{"brew", []string{verb, "nginx"}}}, nil
		}
		if n.commandExists("port") {
			verb := "uninstall"
			if install {
				verb = "install"
			}
			return []packageCommand{{"port", []string{verb, "nginx"}}}, nil
		}
	case "windows":
		if n.commandExists("choco") {
			verb := "uninstall"
			if install {
				verb = "install"
			}
			return []packageCommand{{"choco", []string{verb, "nginx", "-y"}}}, nil
		}
		if n.commandExists("scoop") {
			verb := "uninstall"
			if install {
				verb = "install"
			}
			return []packageCommand{{"scoop", []string{verb, "nginx"}}}, nil
		}
		if n.commandExists("winget") {
			if install {
				return []packageCommand{{"winget", []string{"install", "--id", "nginxinc.nginx", "--exact", "--silent", "--accept-package-agreements", "--accept-source-agreements"}}}, nil
			}
			return []packageCommand{{"winget", []string{"uninstall", "--id", "nginxinc.nginx", "--exact", "--silent"}}}, nil
		}
	case "freebsd", "dragonfly":
		if n.commandExists("pkg") {
			if install {
				return []packageCommand{{"pkg", []string{"install", "-y", "nginx"}}}, nil
			}
			return []packageCommand{{"pkg", []string{"delete", "-y", "nginx"}}}, nil
		}
	case "openbsd":
		if install && n.commandExists("pkg_add") {
			return []packageCommand{{"pkg_add", []string{"-I", "nginx"}}}, nil
		}
		if !install && n.commandExists("pkg_delete") {
			return []packageCommand{{"pkg_delete", []string{"nginx"}}}, nil
		}
	case "netbsd":
		if n.commandExists("pkgin") {
			verb := "remove"
			if install {
				verb = "install"
			}
			return []packageCommand{{"pkgin", []string{"-y", verb, "nginx"}}}, nil
		}
	}
	return nil, fmt.Errorf("%s 系统未找到可用的软件包管理器", runtime.GOOS)
}

type packageCommand struct {
	Name string
	Args []string
}

func (n *Nginx) detectPackageManager() string {
	if override := strings.ToLower(strings.TrimSpace(os.Getenv("NGINX_PACKAGE_MANAGER"))); override != "" && n.commandExists(override) {
		return filepath.Base(override)
	}
	preferred := make([]string, 0, 8)
	if runtime.GOOS == "linux" {
		release := n.ReadOSRelease()
		ids := append([]string{release.ID}, strings.Fields(release.IDLike)...)
		for _, id := range ids {
			switch id {
			case "debian", "ubuntu", "raspbian", "linuxmint", "kali", "devuan", "deepin", "elementary", "pop":
				preferred = append(preferred, "apt-get", "apt", "nala", "aptitude")
			case "rhel", "centos", "fedora", "rocky", "almalinux", "alma", "ol", "oracle", "scientific":
				preferred = append(preferred, "dnf", "dnf5", "yum", "microdnf", "tdnf")
			case "alpine":
				preferred = append(preferred, "apk")
			case "suse", "opensuse", "opensuse-leap", "opensuse-tumbleweed", "sles":
				preferred = append(preferred, "zypper")
			case "arch", "manjaro", "endeavouros":
				preferred = append(preferred, "pacman")
			case "gentoo":
				preferred = append(preferred, "emerge")
			case "nixos":
				preferred = append(preferred, "nix-env")
			case "guix":
				preferred = append(preferred, "guix")
			case "solus":
				preferred = append(preferred, "eopkg")
			case "mageia", "mandriva":
				preferred = append(preferred, "urpmi")
			case "slackware":
				preferred = append(preferred, "slackpkg")
			case "void":
				preferred = append(preferred, "xbps-install")
			case "openwrt":
				preferred = append(preferred, "opkg")
			}
		}
	}
	names := append(preferred,
		"apt-get", "apt", "dnf", "dnf5", "yum", "microdnf", "tdnf", "apk", "zypper",
		"pacman", "emerge", "nix-env", "guix", "eopkg", "urpmi", "slackpkg", "xbps-install", "opkg",
		"brew", "pkg", "pkg_add", "pkgin", "choco", "scoop", "winget",
	)
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		if n.commandExists(name) {
			return name
		}
	}
	return ""
}

func (n *Nginx) detectInitSystem() string {
	if runtime.GOOS == "windows" {
		return "windows-service"
	}
	if runtime.GOOS == "linux" {
		if override := strings.ToLower(strings.TrimSpace(os.Getenv("NGINX_INIT_SYSTEM"))); override != "" {
			switch override {
			case "systemd", "openrc", "procd", "runit", "sysvinit":
				return override
			}
		}
	}
	if n.systemdAvailable() {
		return "systemd"
	}
	if n.openRCAvailable() {
		return "openrc"
	}
	if n.procdAvailable() {
		return "procd"
	}
	if n.runitAvailable() {
		return "runit"
	}
	if n.commandExists("service") || n.fileExists("/etc/init.d") {
		return "sysvinit"
	}
	return "unknown"
}

func (n *Nginx) systemdAvailable() bool {
	if runtime.GOOS != "linux" || !n.commandExists("systemctl") {
		return false
	}
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return true
	}
	if data, err := os.ReadFile("/proc/1/comm"); err == nil && strings.TrimSpace(string(data)) == "systemd" {
		return true
	}
	return false
}

func (n *Nginx) openRCAvailable() bool {
	if runtime.GOOS != "linux" || !n.commandExists("rc-service") {
		return false
	}
	if _, err := os.Stat("/run/openrc"); err == nil {
		return true
	}
	return n.commandExists("rc-update")
}

func (n *Nginx) runitAvailable() bool {
	if runtime.GOOS != "linux" || !n.commandExists("sv") {
		return false
	}
	for _, path := range []string{"/var/service", "/etc/service", "/etc/sv"} {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

func (n *Nginx) procdAvailable() bool {
	if runtime.GOOS != "linux" || !n.fileExists("/etc/rc.common") {
		return false
	}
	return n.commandExists("procd") || n.fileExists("/sbin/procd")
}

func (n *Nginx) runitServiceRoot() string {
	for _, path := range []string{"/var/service", "/etc/service"} {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path
		}
	}
	return "/var/service"
}

func (n *Nginx) commandPath(name string) string {
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	if runtime.GOOS == "windows" || filepath.Base(name) != name {
		return ""
	}
	for _, dir := range n.commandSearchDirs() {
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
			return path
		}
	}
	return ""
}

func (n *Nginx) commandSearchDirs() []string {
	dirs := []string{
		"/usr/local/sbin",
		"/usr/local/bin",
		"/usr/sbin",
		"/usr/bin",
		"/sbin",
		"/bin",
		"/snap/bin",
		"/opt/homebrew/bin",
		"/opt/homebrew/sbin",
		"/opt/local/bin",
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		dirs = append(dirs,
			filepath.Join(home, ".local", "bin"),
			filepath.Join(home, ".nix-profile", "bin"),
			filepath.Join(home, ".guix-profile", "bin"),
		)
	}
	if runtime.GOOS == "linux" {
		dirs = append(dirs,
			"/nix/var/nix/profiles/default/bin",
			"/run/current-system/sw/bin",
			"/usr/libexec",
			"/usr/local/libexec",
		)
	}
	return dirs
}

func (n *Nginx) commandExists(name string) bool {
	return n.commandPath(name) != ""
}

func (n *Nginx) isRootUser() bool {
	if runtime.GOOS == "windows" {
		return true
	}
	currentUser, err := user.Current()
	if err == nil {
		return currentUser.Uid == "0"
	}
	if data, readErr := os.ReadFile("/proc/self/status"); readErr == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "Uid:") {
				fields := strings.Fields(strings.TrimPrefix(line, "Uid:"))
				return len(fields) > 0 && fields[0] == "0"
			}
		}
	}
	return false
}

func (n *Nginx) requiresRootCommand(name string) bool {
	if runtime.GOOS == "windows" {
		return false
	}
	switch filepath.Base(name) {
	case "brew", "nix-env", "guix":
		return false
	default:
		return true
	}
}

func (n *Nginx) runPrivilegedCommand(ctx context.Context, timeout time.Duration, dir, name string, args ...string) (string, error) {
	if !n.requiresRootCommand(name) || n.isRootUser() {
		return n.runCommandDir(ctx, timeout, dir, name, args...)
	}
	privilege := ""
	if n.commandExists("sudo") {
		privilege = "sudo"
	} else if n.commandExists("doas") {
		privilege = "doas"
	}
	if privilege == "" {
		return "", errors.New("当前用户不是 root，且系统没有 sudo/doas")
	}
	privilegedArgs := make([]string, 0, len(args)+2)
	privilegedArgs = append(privilegedArgs, "-n", name)
	privilegedArgs = append(privilegedArgs, args...)
	return n.runCommandDir(ctx, timeout, dir, privilege, privilegedArgs...)
}

func (n *Nginx) writePrivilegedFile(path string, data []byte, mode os.FileMode) error {
	if n.isRootUser() || n.writableDirectory(filepath.Dir(path)) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		return n.atomicWrite(path, data, mode)
	}
	temp, err := os.CreateTemp("", ".nginx-privileged-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err = temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err = temp.Chmod(mode); err != nil {
		_ = temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	output, err := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "install", "-m", fmt.Sprintf("%04o", mode.Perm()), tempPath, path)
	if err != nil {
		return n.formatCommandOutput(output)
	}
	return nil
}

func (n *Nginx) ensurePrivilegedDirectory(path string) error {
	path = filepath.Clean(path)
	if path == "." || path == "" {
		return errors.New("目录不能为空")
	}
	if n.isRootUser() || n.writableDirectory(path) {
		return os.MkdirAll(path, 0755)
	}
	output, err := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "mkdir", "-p", path)
	if err != nil {
		return n.formatCommandOutput(output)
	}
	return nil
}

func (n *Nginx) linkPrivileged(target, link string) error {
	if strings.TrimSpace(target) == "" || strings.TrimSpace(link) == "" {
		return errors.New("服务链接路径不能为空")
	}
	if info, err := os.Lstat(link); err == nil {
		if info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("服务启用路径已存在且不是符号链接: %s", link)
		}
		if err = os.Remove(link); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if n.isRootUser() || n.writableDirectory(filepath.Dir(link)) {
		return os.Symlink(target, link)
	}
	output, err := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "ln", "-s", target, link)
	if err != nil {
		return n.formatCommandOutput(output)
	}
	return nil
}

func (n *Nginx) removePrivilegedDirectory(path string) error {
	if strings.TrimSpace(path) == "" || filepath.Clean(path) == string(filepath.Separator) {
		return errors.New("拒绝删除空目录或文件系统根目录")
	}
	if n.isRootUser() || n.writableDirectory(filepath.Dir(path)) {
		return os.RemoveAll(path)
	}
	output, err := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "rm", "-rf", path)
	if err != nil {
		return n.formatCommandOutput(output)
	}
	return nil
}

func (n *Nginx) removePrivilegedFile(path string) error {
	if n.isRootUser() || n.writableDirectory(filepath.Dir(path)) {
		return os.Remove(path)
	}
	output, err := n.runPrivilegedCommand(context.Background(), n.commandTimeout(), "", "rm", "-f", path)
	if err != nil {
		return n.formatCommandOutput(output)
	}
	return nil
}

func (n *Nginx) runInstallStep(ctx context.Context, timeout time.Duration, dir, installDir, name string, args ...string) (string, error) {
	if n.isRootUser() || n.writableDirectory(installDir) {
		return n.runCommandDir(ctx, timeout, dir, name, args...)
	}
	return n.runPrivilegedCommand(ctx, timeout, dir, name, args...)
}

func (n *Nginx) writableDirectory(path string) bool {
	path = filepath.Clean(path)
	for {
		info, err := os.Stat(path)
		if err == nil {
			if !info.IsDir() {
				return false
			}
			break
		}
		parent := filepath.Dir(path)
		if parent == path {
			return false
		}
		path = parent
	}
	temp, err := os.CreateTemp(path, ".nginx-write-*")
	if err != nil {
		return false
	}
	tempPath := temp.Name()
	_ = temp.Close()
	_ = os.Remove(tempPath)
	return true
}

func (n *Nginx) runCommand(ctx context.Context, timeout time.Duration, name string, args ...string) (string, error) {
	return n.runCommandDir(ctx, timeout, "", name, args...)
}

func (n *Nginx) runCommandDir(ctx context.Context, timeout time.Duration, dir, name string, args ...string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout > 0 {
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
	}
	commandName := name
	if filepath.Base(name) == name {
		if path := n.commandPath(name); path != "" {
			commandName = path
		}
	}
	command := exec.CommandContext(ctx, commandName, args...)
	if dir != "" {
		command.Dir = dir
	}
	output, err := command.CombinedOutput()
	text := strings.TrimSpace(string(output))
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return text, fmt.Errorf("命令执行超时")
	}
	if err != nil {
		return text, err
	}
	return text, nil
}

func (n *Nginx) parseBuildOption(output, name string) string {
	marker := "--" + name + "="
	index := strings.Index(output, marker)
	if index < 0 {
		return ""
	}
	value := output[index+len(marker):]
	if len(value) > 0 && (value[0] == '\'' || value[0] == '"') {
		quote := value[0]
		if end := strings.IndexByte(value[1:], quote); end >= 0 {
			value = value[1 : end+1]
		} else {
			value = value[1:]
		}
	} else if end := strings.IndexAny(value, " \t\r\n"); end >= 0 {
		value = value[:end]
	}
	return strings.Trim(value, "'\"")
}

func (n *Nginx) resolveBuildPath(value, prefix string) string {
	if value == "" || filepath.IsAbs(value) || prefix == "" {
		return value
	}
	return filepath.Join(prefix, value)
}

func (n *Nginx) parseVersion(output string) string {
	for _, part := range strings.Fields(output) {
		if index := strings.Index(part, "nginx/"); index >= 0 {
			return strings.Trim(part[index+len("nginx/"):], ") ,")
		}
	}
	return ""
}

func (n *Nginx) readPid(path string) (int, error) {
	if path == "" {
		return 0, os.ErrNotExist
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return 0, errors.New("nginx PID 无效")
	}
	return pid, nil
}

func (n *Nginx) processRunning(pid int) (bool, error) {
	if runtime.GOOS == "windows" {
		output, err := n.runCommand(context.Background(), defaultTimeout, "tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/FO", "CSV", "/NH")
		if err != nil {
			return false, nil
		}
		return strings.Contains(output, `"`+strconv.Itoa(pid)+`"`), nil
	}
	_, err := n.runCommand(context.Background(), defaultTimeout, "kill", "-0", strconv.Itoa(pid))
	if err != nil {
		output, psErr := n.runCommand(context.Background(), defaultTimeout, "ps", "-p", strconv.Itoa(pid), "-o", "pid=")
		if psErr == nil && strings.TrimSpace(output) != "" {
			return true, nil
		}
		return false, nil
	}
	return true, nil
}

func (n *Nginx) normalizeSiteName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\\:`) {
		return "", errors.New("站点文件名不合法")
	}
	if !strings.HasSuffix(strings.ToLower(name), ".conf") {
		name += ".conf"
	}
	return name, nil
}

func (n *Nginx) readSite(path string) ([]byte, os.FileMode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, 0644, err
	}
	data, err := os.ReadFile(path)
	return data, info.Mode().Perm(), err
}

func (n *Nginx) atomicWrite(path string, data []byte, mode os.FileMode) error {
	temp, err := os.CreateTemp(filepath.Dir(path), ".nginx-site-*")
	if err != nil {
		return fmt.Errorf("创建临时配置失败: %w", err)
	}
	tempPath := temp.Name()
	defer func() {
		_ = temp.Close()
		_ = os.Remove(tempPath)
	}()
	if err = temp.Chmod(mode); err == nil {
		_, err = temp.Write(data)
	}
	if err == nil {
		err = temp.Sync()
	}
	if closeErr := temp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("写入配置失败: %w", err)
	}
	if runtime.GOOS == "windows" {
		if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("替换配置失败: %w", err)
		}
	}
	if err = os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("替换配置失败: %w", err)
	}
	return nil
}

func (n *Nginx) formatCommandOutput(output string) error {
	output = strings.TrimSpace(output)
	if output == "" {
		return errors.New("命令执行失败")
	}
	return errors.New(output)
}
