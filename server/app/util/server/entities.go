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
	HTTPListen           []string      // HTTP 监听地址，nil 默认 :80，空切片关闭监听
	HTTPSListen          []string      // HTTPS 监听地址，nil 默认 :443，空切片关闭监听
	Protocols            []string      // 启用的 HTTP 协议，nil 默认 h1、h2、h3
	DataPath             string        // 证书和运行数据目录，默认 root/data
	CachePath            string        // 站点响应缓存根目录，默认 root/cache
	LogPath              string        // HTTP 访问日志文件，- 表示使用默认日志输出
	WAFLogPath           string        // WAF 审计日志文件，默认 root/logs/waf.log
	ConfigPath           string        // 配置快照文件，为空时不保存
	LogLevel             string        // HTTP 日志级别，默认 INFO
	TrustedProxies       []string      // 可信代理 IP 或 CIDR
	ClientIPHeaders      []string      // 从可信代理读取客户端 IP 的请求头
	TrustedProxiesStrict bool          // 是否严格解析可信代理链
	ReadTimeout          time.Duration // 读取完整请求的超时时间，0 表示不限制
	ReadHeaderTimeout    time.Duration // 读取请求头的超时时间，默认 10 秒
	WriteTimeout         time.Duration // 写入响应的超时时间，0 表示不限制
	IdleTimeout          time.Duration // 空闲连接的超时时间，默认 5 分钟
	GracePeriod          time.Duration // HTTP 服务优雅关闭等待时间，默认 10 秒
	MaxHeaderBytes       int           // 单个请求头最大字节数，默认 1 MiB
	HTTPChallengeHost    string        // ACME HTTP-01 验证监听地址
	HTTPChallengePort    int           // ACME HTTP-01 验证端口，默认 80
	ACMEEmail            string        // ACME 账户默认邮箱
}

// Site 是与数据库无关的站点配置，可由业务层直接持久化。
type Site struct {
	ID            string              `json:"id"`                      // 站点唯一标识
	Name          string              `json:"name"`                    // 站点显示名称
	Enabled       bool                `json:"enabled"`                 // 是否启用站点
	Domains       []string            `json:"domains"`                 // 站点绑定域名
	Root          string              `json:"root,omitempty"`          // 静态文件根目录
	Index         []string            `json:"index,omitempty"`         // 默认首页文件
	Browse        bool                `json:"browse,omitempty"`        // 是否开启目录浏览
	TryFiles      []string            `json:"try_files,omitempty"`     // 静态文件匹配顺序
	Hide          []string            `json:"hide,omitempty"`          // 禁止访问的文件或路径
	Precompressed []string            `json:"precompressed,omitempty"` // 预压缩文件格式
	Proxy         *ProxyOptions       `json:"proxy,omitempty"`         // 默认反向代理配置
	Routes        []Route             `json:"routes,omitempty"`        // 优先执行的自定义路由
	Headers       HeaderOptions       `json:"headers,omitempty"`       // 请求和响应头操作
	Compression   CompressionOptions  `json:"compression,omitempty"`   // 响应压缩配置
	TLS           TLSOptions          `json:"tls,omitempty"`           // HTTPS 和证书配置
	WAF           WAFOptions          `json:"waf,omitempty"`           // Web 应用防火墙配置
	Cache         CacheOptions        `json:"cache,omitempty"`         // 响应缓存配置
	RateLimit     RateLimitOptions    `json:"rate_limit,omitempty"`    // 访问频率限制配置
	TrafficLimit  TrafficLimitOptions `json:"traffic_limit,omitempty"` // 并发和响应速度限制配置
	HandlerOrder  []Module            `json:"handler_order,omitempty"` // 处理器阶段执行顺序
	Handlers      []json.RawMessage   `json:"handlers,omitempty"`      // 自定义 HTTP 处理器配置
}

// Route 定义站点内优先于默认处理器执行的路径规则。
type Route struct {
	Name          string            `json:"name,omitempty"`          // 路由名称
	Paths         []string          `json:"paths,omitempty"`         // 匹配的请求路径
	Methods       []string          `json:"methods,omitempty"`       // 匹配的 HTTP 请求方法
	StripPrefix   string            `json:"strip_prefix,omitempty"`  // 转发前移除的路径前缀
	Rewrite       string            `json:"rewrite,omitempty"`       // 重写后的请求 URI
	Root          string            `json:"root,omitempty"`          // 路由静态文件根目录
	Index         []string          `json:"index,omitempty"`         // 默认首页文件
	Browse        bool              `json:"browse,omitempty"`        // 是否开启目录浏览
	TryFiles      []string          `json:"try_files,omitempty"`     // 静态文件匹配顺序
	Hide          []string          `json:"hide,omitempty"`          // 禁止访问的文件或路径
	Precompressed []string          `json:"precompressed,omitempty"` // 预压缩文件格式
	Proxy         *ProxyOptions     `json:"proxy,omitempty"`         // 路由反向代理配置
	Response      *ResponseOptions  `json:"response,omitempty"`      // 固定响应或跳转配置
	Handlers      []json.RawMessage `json:"handlers,omitempty"`      // 自定义 HTTP 处理器配置
}

// ProxyOptions 定义 HTTP 或 FastCGI 上游。
type ProxyOptions struct {
	Upstreams             []Upstream        `json:"upstreams"`                          // 上游服务地址
	Transport             ProxyTransport    `json:"transport,omitempty"`                // 代理传输实现，默认 http
	Scheme                ProxyScheme       `json:"scheme,omitempty"`                   // HTTP 上游协议，默认 http
	Policy                string            `json:"policy,omitempty"`                   // 上游负载均衡策略，默认 random
	Retries               int               `json:"retries,omitempty"`                  // 选择可用上游的重试次数
	TryDuration           time.Duration     `json:"try_duration,omitempty"`             // 选择可用上游的最长时间
	TryInterval           time.Duration     `json:"try_interval,omitempty"`             // 上游重试间隔
	DialTimeout           time.Duration     `json:"dial_timeout,omitempty"`             // 建立上游连接的超时时间
	ReadTimeout           time.Duration     `json:"read_timeout,omitempty"`             // 读取上游响应的超时时间
	WriteTimeout          time.Duration     `json:"write_timeout,omitempty"`            // 写入上游请求的超时时间
	ResponseHeaderTimeout time.Duration     `json:"response_header_timeout,omitempty"`  // 等待上游响应头的超时时间
	FlushInterval         time.Duration     `json:"flush_interval,omitempty"`           // 向客户端刷新响应的间隔
	StreamTimeout         time.Duration     `json:"stream_timeout,omitempty"`           // 流式响应的最长持续时间
	StreamCloseDelay      time.Duration     `json:"stream_close_delay,omitempty"`       // 配置重载时关闭流连接的延迟
	Versions              []string          `json:"versions,omitempty"`                 // 允许使用的上游 HTTP 版本
	TLSServerName         string            `json:"tls_server_name,omitempty"`          // 上游 TLS SNI 服务器名称
	TLSInsecureSkipVerify *bool             `json:"tls_insecure_skip_verify,omitempty"` // 是否跳过上游证书校验，默认 true
	HealthURI             string            `json:"health_uri,omitempty"`               // 主动健康检查请求路径
	HealthInterval        time.Duration     `json:"health_interval,omitempty"`          // 主动健康检查间隔
	HealthTimeout         time.Duration     `json:"health_timeout,omitempty"`           // 单次健康检查超时时间
	HealthStatus          int               `json:"health_status,omitempty"`            // 健康检查期望状态码
	Root                  string            `json:"root,omitempty"`                     // FastCGI 站点根目录，默认继承站点根目录
	SplitPath             []string          `json:"split_path,omitempty"`               // FastCGI 脚本路径分割扩展名，默认 .php
	Index                 string            `json:"index,omitempty"`                    // FastCGI 默认入口文件，默认 index.php
	TryFiles              []string          `json:"try_files,omitempty"`                // FastCGI 文件匹配顺序
	TryPolicy             string            `json:"try_policy,omitempty"`               // FastCGI 文件匹配策略，默认 first_exist
	ResolveRootSymlink    bool              `json:"resolve_root_symlink,omitempty"`     // 是否解析 FastCGI 根目录符号链接
	Hide                  []string          `json:"hide,omitempty"`                     // FastCGI 静态文件隐藏规则
	Env                   map[string]string `json:"env,omitempty"`                      // FastCGI 环境变量
	CaptureStderr         bool              `json:"capture_stderr,omitempty"`           // 是否记录 FastCGI 标准错误输出
	Headers               HeaderOptions     `json:"headers,omitempty"`                  // 代理请求和响应头操作
}

// Upstream 定义一个静态上游地址。
type Upstream struct {
	Dial        string `json:"dial"`                   // 上游网络地址
	MaxRequests int    `json:"max_requests,omitempty"` // 允许转发到当前上游的最大并发请求数
}

// ResponseOptions 定义固定响应或跳转。
type ResponseOptions struct {
	Status  int                 `json:"status,omitempty"`  // HTTP 响应状态码
	Body    string              `json:"body,omitempty"`    // 固定响应内容
	Headers map[string][]string `json:"headers,omitempty"` // 固定响应头
}

// HeaderOptions 定义请求和响应头的增删改操作。
type HeaderOptions struct {
	RequestAdd     map[string][]string `json:"request_add,omitempty"`     // 追加的请求头
	RequestSet     map[string][]string `json:"request_set,omitempty"`     // 覆盖设置的请求头
	RequestDelete  []string            `json:"request_delete,omitempty"`  // 删除的请求头
	ResponseAdd    map[string][]string `json:"response_add,omitempty"`    // 追加的响应头
	ResponseSet    map[string][]string `json:"response_set,omitempty"`    // 覆盖设置的响应头
	ResponseDelete []string            `json:"response_delete,omitempty"` // 删除的响应头
}

// CompressionOptions 定义响应压缩，默认优先 zstd，其次 gzip。
type CompressionOptions struct {
	Enabled    bool     `json:"enabled"`              // 是否启用响应压缩
	Algorithms []string `json:"algorithms,omitempty"` // 压缩算法及优先顺序
	MinLength  int      `json:"min_length,omitempty"` // 启用压缩的最小响应字节数
}

// TLSOptions 定义手动部署的证书。启用 TLS 不会触发自动申请。
type TLSOptions struct {
	Enabled        bool              `json:"enabled"`                 // 是否启用 HTTPS
	RedirectHTTP   bool              `json:"redirect_http,omitempty"` // 是否将已覆盖域名的 HTTP 重定向到 HTTPS
	MinVersion     string            `json:"min_version,omitempty"`   // 允许的最低 TLS 版本
	MaxVersion     string            `json:"max_version,omitempty"`   // 允许的最高 TLS 版本
	Certificates   []CertificatePair `json:"certificates,omitempty"`  // 站点使用的证书
	coveredDomains []string          // 当前证书实际覆盖的站点域名，仅用于生成运行配置
}

// CertificatePair 支持从文件或内存 PEM 加载证书，两个来源只能选择一个。
type CertificatePair struct {
	CertificateFile string `json:"certificate_file,omitempty"` // 证书链文件路径
	KeyFile         string `json:"key_file,omitempty"`         // 私钥文件路径
	CertificatePEM  string `json:"certificate_pem,omitempty"`  // PEM 格式证书链内容
	PrivateKeyPEM   string `json:"-"`                          // PEM 格式私钥内容，不参与 JSON 序列化
}

// WAFOptions 定义 Coraza WAF。默认模式为 DetectionOnly。
type WAFOptions struct {
	Enabled          bool         `json:"enabled"`                      // 是否启用 WAF
	Mode             WAFMode      `json:"mode,omitempty"`               // WAF 检测或拦截模式
	OWASP            OWASPOptions `json:"owasp,omitempty"`              // OWASP CRS 配置
	AuditLog         bool         `json:"audit_log,omitempty"`          // 是否记录 WAF 审计日志
	RequestBodyLimit int64        `json:"request_body_limit,omitempty"` // 可检查的请求体最大字节数
	Directives       []string     `json:"directives,omitempty"`         // 自定义 Coraza 指令
	Rules            []WAFRule    `json:"rules,omitempty"`              // 自定义 WAF 规则
}

// OWASPOptions 定义内嵌 OWASP Core Rule Set 的运行参数。
type OWASPOptions struct {
	Enabled                       bool     `json:"enabled"`                                    // 是否启用 OWASP CRS
	ParanoiaLevel                 int      `json:"paranoia_level,omitempty"`                   // 阻断规则敏感级别，范围 1-4，默认 1
	DetectionParanoiaLevel        int      `json:"detection_paranoia_level,omitempty"`         // 仅检测规则敏感级别，默认等于阻断级别
	InboundAnomalyScoreThreshold  int      `json:"inbound_anomaly_score_threshold,omitempty"`  // 入站请求异常分数阈值，默认 5
	OutboundAnomalyScoreThreshold int      `json:"outbound_anomaly_score_threshold,omitempty"` // 出站响应异常分数阈值，默认 4
	ReportingLevel                *int     `json:"reporting_level,omitempty"`                  // CRS 日志报告级别，范围 0-5
	EarlyBlocking                 bool     `json:"early_blocking,omitempty"`                   // 是否提前执行异常分数阻断
	SamplingPercentage            int      `json:"sampling_percentage,omitempty"`              // 接受 CRS 检查的请求百分比，默认 100
	EnforceBodyProcessor          bool     `json:"enforce_body_processor,omitempty"`           // 是否强制处理 URL 编码请求体
	ValidateUTF8                  bool     `json:"validate_utf8,omitempty"`                    // 是否校验请求数据的 UTF-8 编码
	SkipResponseAnalysis          bool     `json:"skip_response_analysis,omitempty"`           // 是否跳过出站响应分析
	AllowedMethods                []string `json:"allowed_methods,omitempty"`                  // 允许的 HTTP 请求方法
	AllowedContentTypes           []string `json:"allowed_content_types,omitempty"`            // 允许的请求 Content-Type
	AllowedHTTPVersions           []string `json:"allowed_http_versions,omitempty"`            // 允许的 HTTP 协议版本
	AllowedCharsets               []string `json:"allowed_charsets,omitempty"`                 // 允许的请求字符集
	RestrictedExtensions          []string `json:"restricted_extensions,omitempty"`            // 禁止上传或请求的文件扩展名
	RestrictedHeaders             []string `json:"restricted_headers,omitempty"`               // 基础受限请求头
	RestrictedHeadersExtended     []string `json:"restricted_headers_extended,omitempty"`      // 扩展受限请求头
	MaxArguments                  int      `json:"max_arguments,omitempty"`                    // 单个请求允许的最大参数数量
	MaxArgumentNameLength         int      `json:"max_argument_name_length,omitempty"`         // 单个参数名最大字节数
	MaxArgumentLength             int      `json:"max_argument_length,omitempty"`              // 单个参数值最大字节数
	TotalArgumentLength           int      `json:"total_argument_length,omitempty"`            // 所有参数名和值的最大总字节数
	MaxFileSize                   int64    `json:"max_file_size,omitempty"`                    // 单个上传文件最大字节数
	CombinedFileSize              int64    `json:"combined_file_size,omitempty"`               // 单次请求上传文件最大总字节数
	SetupDirectives               []string `json:"setup_directives,omitempty"`                 // CRS 规则加载前执行的自定义指令
}

// WAFRule 保存一条可由业务层独立开关的 Coraza 规则。
type WAFRule struct {
	ID        string `json:"id"`             // 规则唯一标识
	Name      string `json:"name,omitempty"` // 规则显示名称
	Enabled   bool   `json:"enabled"`        // 是否启用规则
	Directive string `json:"directive"`      // Coraza 规则指令
}

// CacheOptions 定义站点响应缓存。默认 no-store，必须显式允许缓存动态响应。
type CacheOptions struct {
	Enabled             bool          `json:"enabled"`                         // 是否启用响应缓存
	TTL                 time.Duration `json:"ttl,omitempty"`                   // 缓存新鲜内容的有效时间
	Stale               time.Duration `json:"stale,omitempty"`                 // 过期内容允许继续使用的时间
	Methods             []string      `json:"methods,omitempty"`               // 允许缓存的 HTTP 请求方法
	KeyHeaders          []string      `json:"key_headers,omitempty"`           // 参与生成缓存键的请求头
	DefaultCacheControl string        `json:"default_cache_control,omitempty"` // 响应未声明时使用的缓存策略
	Name                string        `json:"name,omitempty"`                  // 缓存实例名称
	MaxSize             int64         `json:"max_size,omitempty"`              // 缓存容量及单个响应体上限，单位字节
}

// RateLimitOptions 定义单机滑动窗口限流。
type RateLimitOptions struct {
	Enabled    bool          `json:"enabled"`               // 是否启用访问频率限制
	Key        string        `json:"key,omitempty"`         // 限流对象键或占位符表达式
	Window     time.Duration `json:"window,omitempty"`      // 滑动统计时间窗口
	MaxEvents  int           `json:"max_events,omitempty"`  // 窗口内允许的最大请求数
	Paths      []string      `json:"paths,omitempty"`       // 应用限流的请求路径
	Methods    []string      `json:"methods,omitempty"`     // 应用限流的 HTTP 请求方法
	IPv4Prefix int           `json:"ipv4_prefix,omitempty"` // IPv4 地址聚合前缀长度
	IPv6Prefix int           `json:"ipv6_prefix,omitempty"` // IPv6 地址聚合前缀长度
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
	Domain string        `json:"domain"`          // 申请证书的域名
	Email  string        `json:"email,omitempty"` // ACME 账户邮箱，为空时使用全局配置
	CA     CertificateCA `json:"ca,omitempty"`    // 证书签发环境，默认 production
}

// Certificate 返回证书信息和 PEM。私钥仅供 Go 调用方使用，不参与 JSON 序列化。
type Certificate struct {
	Domain          string    `json:"domain"`           // 申请证书的主域名
	Issuer          string    `json:"issuer"`           // 证书签发环境
	DNSNames        []string  `json:"dns_names"`        // 证书覆盖的 DNS 名称
	SerialNumber    string    `json:"serial_number"`    // 证书序列号
	NotBefore       time.Time `json:"not_before"`       // 证书生效时间
	NotAfter        time.Time `json:"not_after"`        // 证书到期时间
	CertificateFile string    `json:"certificate_file"` // 本地证书文件路径
	KeyFile         string    `json:"key_file"`         // 本地私钥文件路径
	CertificatePEM  []byte    `json:"certificate_pem"`  // PEM 格式证书链内容
	PrivateKeyPEM   []byte    `json:"private_key_pem"`  // PEM 格式私钥内容
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
