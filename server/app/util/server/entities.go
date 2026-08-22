package server

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/caddyserver/certmagic"
)

// Options 控制 HTTP 服务的监听地址、数据文件和运行参数。
// 路径为空时使用 root 下的默认目录，ConfigPath 为空表示不保存配置快照。
type Options struct {
	HTTPListen           []string
	HTTPSListen          []string
	Protocols            []string
	DataPath             string
	CachePath            string
	LogPath              string
	WAFLogPath           string
	ConfigPath           string
	LogLevel             string
	TrustedProxies       []string
	ClientIPHeaders      []string
	TrustedProxiesStrict bool
	ReadTimeout          time.Duration
	ReadHeaderTimeout    time.Duration
	WriteTimeout         time.Duration
	IdleTimeout          time.Duration
	GracePeriod          time.Duration
	MaxHeaderBytes       int
	HTTPChallengeHost    string
	HTTPChallengePort    int
	ACMEEmail            string
}

// Site 是与数据库无关的站点配置，可由业务层直接持久化。
type Site struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Enabled       bool                `json:"enabled"`
	Domains       []string            `json:"domains"`
	Root          string              `json:"root,omitempty"`
	Index         []string            `json:"index,omitempty"`
	Browse        bool                `json:"browse,omitempty"`
	TryFiles      []string            `json:"try_files,omitempty"`
	Hide          []string            `json:"hide,omitempty"`
	Precompressed []string            `json:"precompressed,omitempty"`
	Proxy         *ProxyOptions       `json:"proxy,omitempty"`
	Routes        []Route             `json:"routes,omitempty"`
	Headers       HeaderOptions       `json:"headers,omitempty"`
	Compression   CompressionOptions  `json:"compression,omitempty"`
	TLS           TLSOptions          `json:"tls,omitempty"`
	WAF           WAFOptions          `json:"waf,omitempty"`
	Cache         CacheOptions        `json:"cache,omitempty"`
	RateLimit     RateLimitOptions    `json:"rate_limit,omitempty"`
	TrafficLimit  TrafficLimitOptions `json:"traffic_limit,omitempty"`
	HandlerOrder  []Module            `json:"handler_order,omitempty"`
	Handlers      []json.RawMessage   `json:"handlers,omitempty"`
}

// Route 定义站点内优先于默认处理器执行的路径规则。
type Route struct {
	Name          string            `json:"name,omitempty"`
	Paths         []string          `json:"paths,omitempty"`
	Methods       []string          `json:"methods,omitempty"`
	StripPrefix   string            `json:"strip_prefix,omitempty"`
	Rewrite       string            `json:"rewrite,omitempty"`
	Root          string            `json:"root,omitempty"`
	Index         []string          `json:"index,omitempty"`
	Browse        bool              `json:"browse,omitempty"`
	TryFiles      []string          `json:"try_files,omitempty"`
	Hide          []string          `json:"hide,omitempty"`
	Precompressed []string          `json:"precompressed,omitempty"`
	Proxy         *ProxyOptions     `json:"proxy,omitempty"`
	Response      *ResponseOptions  `json:"response,omitempty"`
	Handlers      []json.RawMessage `json:"handlers,omitempty"`
}

// ProxyOptions 定义 HTTP 或 FastCGI 上游。
type ProxyOptions struct {
	Upstreams             []Upstream        `json:"upstreams"`
	Transport             ProxyTransport    `json:"transport,omitempty"`
	Scheme                ProxyScheme       `json:"scheme,omitempty"`
	Policy                string            `json:"policy,omitempty"`
	Retries               int               `json:"retries,omitempty"`
	TryDuration           time.Duration     `json:"try_duration,omitempty"`
	TryInterval           time.Duration     `json:"try_interval,omitempty"`
	DialTimeout           time.Duration     `json:"dial_timeout,omitempty"`
	ReadTimeout           time.Duration     `json:"read_timeout,omitempty"`
	WriteTimeout          time.Duration     `json:"write_timeout,omitempty"`
	ResponseHeaderTimeout time.Duration     `json:"response_header_timeout,omitempty"`
	FlushInterval         time.Duration     `json:"flush_interval,omitempty"`
	StreamTimeout         time.Duration     `json:"stream_timeout,omitempty"`
	StreamCloseDelay      time.Duration     `json:"stream_close_delay,omitempty"`
	Versions              []string          `json:"versions,omitempty"`
	TLSServerName         string            `json:"tls_server_name,omitempty"`
	TLSInsecureSkipVerify *bool             `json:"tls_insecure_skip_verify,omitempty"`
	HealthURI             string            `json:"health_uri,omitempty"`
	HealthInterval        time.Duration     `json:"health_interval,omitempty"`
	HealthTimeout         time.Duration     `json:"health_timeout,omitempty"`
	HealthStatus          int               `json:"health_status,omitempty"`
	Root                  string            `json:"root,omitempty"`
	SplitPath             []string          `json:"split_path,omitempty"`
	Index                 string            `json:"index,omitempty"`
	TryFiles              []string          `json:"try_files,omitempty"`
	TryPolicy             string            `json:"try_policy,omitempty"`
	ResolveRootSymlink    bool              `json:"resolve_root_symlink,omitempty"`
	Hide                  []string          `json:"hide,omitempty"`
	Env                   map[string]string `json:"env,omitempty"`
	CaptureStderr         bool              `json:"capture_stderr,omitempty"`
	Headers               HeaderOptions     `json:"headers,omitempty"`
}

// Upstream 定义一个静态上游地址。
type Upstream struct {
	Dial        string `json:"dial"`
	MaxRequests int    `json:"max_requests,omitempty"`
}

// ResponseOptions 定义固定响应或跳转。
type ResponseOptions struct {
	Status  int                 `json:"status,omitempty"`
	Body    string              `json:"body,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
}

// HeaderOptions 定义请求和响应头的增删改操作。
type HeaderOptions struct {
	RequestAdd     map[string][]string `json:"request_add,omitempty"`
	RequestSet     map[string][]string `json:"request_set,omitempty"`
	RequestDelete  []string            `json:"request_delete,omitempty"`
	ResponseAdd    map[string][]string `json:"response_add,omitempty"`
	ResponseSet    map[string][]string `json:"response_set,omitempty"`
	ResponseDelete []string            `json:"response_delete,omitempty"`
}

// CompressionOptions 定义响应压缩，默认优先 zstd，其次 gzip。
type CompressionOptions struct {
	Enabled    bool     `json:"enabled"`
	Algorithms []string `json:"algorithms,omitempty"`
	MinLength  int      `json:"min_length,omitempty"`
}

// TLSOptions 定义手动部署的证书。启用 TLS 不会触发自动申请。
type TLSOptions struct {
	Enabled      bool              `json:"enabled"`
	RedirectHTTP bool              `json:"redirect_http,omitempty"`
	MinVersion   string            `json:"min_version,omitempty"`
	MaxVersion   string            `json:"max_version,omitempty"`
	Certificates []CertificatePair `json:"certificates,omitempty"`
}

// CertificatePair 支持从文件或内存 PEM 加载证书，两个来源只能选择一个。
type CertificatePair struct {
	CertificateFile string `json:"certificate_file,omitempty"`
	KeyFile         string `json:"key_file,omitempty"`
	CertificatePEM  string `json:"certificate_pem,omitempty"`
	PrivateKeyPEM   string `json:"-"`
}

// WAFOptions 定义 Coraza WAF。默认模式为 DetectionOnly。
type WAFOptions struct {
	Enabled          bool         `json:"enabled"`
	Mode             WAFMode      `json:"mode,omitempty"`
	OWASP            OWASPOptions `json:"owasp,omitempty"`
	AuditLog         bool         `json:"audit_log,omitempty"`
	RequestBodyLimit int64        `json:"request_body_limit,omitempty"`
	Directives       []string     `json:"directives,omitempty"`
	Rules            []WAFRule    `json:"rules,omitempty"`
}

// OWASPOptions 定义内嵌 OWASP Core Rule Set 的运行参数。
type OWASPOptions struct {
	Enabled                       bool     `json:"enabled"`
	ParanoiaLevel                 int      `json:"paranoia_level,omitempty"`
	DetectionParanoiaLevel        int      `json:"detection_paranoia_level,omitempty"`
	InboundAnomalyScoreThreshold  int      `json:"inbound_anomaly_score_threshold,omitempty"`
	OutboundAnomalyScoreThreshold int      `json:"outbound_anomaly_score_threshold,omitempty"`
	ReportingLevel                *int     `json:"reporting_level,omitempty"`
	EarlyBlocking                 bool     `json:"early_blocking,omitempty"`
	SamplingPercentage            int      `json:"sampling_percentage,omitempty"`
	EnforceBodyProcessor          bool     `json:"enforce_body_processor,omitempty"`
	ValidateUTF8                  bool     `json:"validate_utf8,omitempty"`
	SkipResponseAnalysis          bool     `json:"skip_response_analysis,omitempty"`
	AllowedMethods                []string `json:"allowed_methods,omitempty"`
	AllowedContentTypes           []string `json:"allowed_content_types,omitempty"`
	AllowedHTTPVersions           []string `json:"allowed_http_versions,omitempty"`
	AllowedCharsets               []string `json:"allowed_charsets,omitempty"`
	RestrictedExtensions          []string `json:"restricted_extensions,omitempty"`
	RestrictedHeaders             []string `json:"restricted_headers,omitempty"`
	RestrictedHeadersExtended     []string `json:"restricted_headers_extended,omitempty"`
	MaxArguments                  int      `json:"max_arguments,omitempty"`
	MaxArgumentNameLength         int      `json:"max_argument_name_length,omitempty"`
	MaxArgumentLength             int      `json:"max_argument_length,omitempty"`
	TotalArgumentLength           int      `json:"total_argument_length,omitempty"`
	MaxFileSize                   int64    `json:"max_file_size,omitempty"`
	CombinedFileSize              int64    `json:"combined_file_size,omitempty"`
	SetupDirectives               []string `json:"setup_directives,omitempty"`
}

// WAFRule 保存一条可由业务层独立开关的 Coraza 规则。
type WAFRule struct {
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	Enabled   bool   `json:"enabled"`
	Directive string `json:"directive"`
}

// CacheOptions 定义站点响应缓存。默认 no-store，必须显式允许缓存动态响应。
type CacheOptions struct {
	Enabled             bool          `json:"enabled"`
	TTL                 time.Duration `json:"ttl,omitempty"`
	Stale               time.Duration `json:"stale,omitempty"`
	Methods             []string      `json:"methods,omitempty"`
	KeyHeaders          []string      `json:"key_headers,omitempty"`
	DefaultCacheControl string        `json:"default_cache_control,omitempty"`
	Name                string        `json:"name,omitempty"`
	MaxSize             int64         `json:"max_size,omitempty"`
}

// RateLimitOptions 定义单机滑动窗口限流。
type RateLimitOptions struct {
	Enabled    bool          `json:"enabled"`
	Key        string        `json:"key,omitempty"`
	Window     time.Duration `json:"window,omitempty"`
	MaxEvents  int           `json:"max_events,omitempty"`
	Paths      []string      `json:"paths,omitempty"`
	Methods    []string      `json:"methods,omitempty"`
	IPv4Prefix int           `json:"ipv4_prefix,omitempty"`
	IPv6Prefix int           `json:"ipv6_prefix,omitempty"`
}

// TrafficLimitOptions 定义站点并发和单个请求的响应速度限制，值为 0 表示不限制。
type TrafficLimitOptions struct {
	Enabled             bool  `json:"enabled"`                          // 是否启用流量限制
	MaxConnections      int64 `json:"max_connections,omitempty"`        // 站点最大并发请求数
	MaxConnectionsPerIP int64 `json:"max_connections_per_ip,omitempty"` // 单个客户端 IP 最大并发请求数
	RatePerRequest      int64 `json:"rate_per_request,omitempty"`       // 单个请求响应速度，单位为字节/秒
}

// CertificateRequest 定义一次显式 ACME 申请。
type CertificateRequest struct {
	Domain string        `json:"domain"`
	Email  string        `json:"email,omitempty"`
	CA     CertificateCA `json:"ca,omitempty"`
}

// Certificate 返回证书信息和 PEM。私钥仅供 Go 调用方使用，不参与 JSON 序列化。
type Certificate struct {
	Domain          string    `json:"domain"`
	Issuer          string    `json:"issuer"`
	DNSNames        []string  `json:"dns_names"`
	SerialNumber    string    `json:"serial_number"`
	NotBefore       time.Time `json:"not_before"`
	NotAfter        time.Time `json:"not_after"`
	CertificateFile string    `json:"certificate_file"`
	KeyFile         string    `json:"-"`
	CertificatePEM  []byte    `json:"certificate_pem"`
	PrivateKeyPEM   []byte    `json:"-"`
}

// Manager 在内存中管理站点，并以事务方式热加载 HTTP 配置。
type Manager struct {
	root       string
	dataPath   string
	cachePath  string
	logPath    string
	wafLogPath string
	configPath string
	options    Options
	storage    *certmagic.FileStorage

	mu          sync.RWMutex
	operationMu sync.Mutex
	sites       map[string]*Site
	config      []byte
	configMode  bool
	running     bool
}

var httpRuntime struct {
	sync.Mutex
	owner *Manager
}
