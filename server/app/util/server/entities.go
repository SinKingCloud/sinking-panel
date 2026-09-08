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
	HTTPListen           []string        `json:"http_listen"`            // HTTP 监听地址，nil 默认 :80，空切片关闭监听
	HTTPSListen          []string        `json:"https_listen"`           // HTTPS 监听地址，nil 默认 :443，空切片关闭监听
	Protocols            []string        `json:"protocols"`              // 启用的 HTTP 协议，nil 默认 h1、h2、h3
	DefaultSite          string          `json:"default_site"`           // 默认站点 ID，空表示未知域名显示网站不存在页
	NotFoundPage         ResponseOptions `json:"not_found_page"`         // 站点内资源不存在页面
	SiteNotFoundPage     ResponseOptions `json:"site_not_found_page"`    // 请求域名未绑定网站时显示的页面
	SiteDisabledPage     ResponseOptions `json:"site_disabled_page"`     // 请求域名所属网站已停用时显示的页面
	DataPath             string          `json:"data_path"`              // HTTP 运行数据目录，默认 root/data
	ACMEPath             string          `json:"-"`                      // ACME 账户和签发数据目录，默认 root/acme
	CachePath            string          `json:"cache_path"`             // 站点响应缓存根目录，默认 root/cache
	LogPath              string          `json:"log_path"`               // HTTP 运行日志文件，- 表示不保存运行日志
	WAFLogPath           string          `json:"waf_log_path"`           // 全局 WAF 模块日志文件
	ConfigPath           string          `json:"config_path"`            // 配置快照文件，为空时不保存
	LogLevel             string          `json:"log_level"`              // HTTP 日志级别，默认 INFO
	TrustedProxies       []string        `json:"trusted_proxies"`        // 可信代理 IP 或 CIDR
	ClientIPHeaders      []string        `json:"client_ip_headers"`      // 从可信代理读取客户端 IP 的请求头
	TrustedProxiesStrict bool            `json:"trusted_proxies_strict"` // 是否严格解析可信代理链
	ReadTimeout          time.Duration   `json:"read_timeout"`           // 读取完整请求的超时时间，0 表示不限制
	ReadHeaderTimeout    time.Duration   `json:"read_header_timeout"`    // 读取请求头的超时时间，默认 10 秒
	WriteTimeout         time.Duration   `json:"write_timeout"`          // 写入响应的超时时间，0 表示不限制
	IdleTimeout          time.Duration   `json:"idle_timeout"`           // 空闲连接的超时时间，默认 5 分钟
	GracePeriod          time.Duration   `json:"grace_period"`           // HTTP 服务优雅关闭等待时间，默认 10 秒
	MaxHeaderBytes       int             `json:"max_header_bytes"`       // 单个请求头最大字节数，默认 1 MiB
	HTTPChallengeHost    string          `json:"http_challenge_host"`    // ACME HTTP-01 验证监听地址
	HTTPChallengePort    int             `json:"http_challenge_port"`    // ACME HTTP-01 验证端口，默认 80
	ACMEEmail            string          `json:"acme_email"`             // ACME 账户默认邮箱
}

// Site 是与数据库无关的站点配置，可由业务层直接持久化。
type Site struct {
	ID            string              `json:"id"`            // 站点唯一标识
	Name          string              `json:"name"`          // 站点显示名称
	Enabled       bool                `json:"enabled"`       // 是否启用站点
	Domains       []string            `json:"domains"`       // 站点绑定域名
	Root          string              `json:"root"`          // 静态文件根目录
	Index         []string            `json:"index"`         // 默认首页文件
	Browse        bool                `json:"browse"`        // 是否开启目录浏览
	TryFiles      []string            `json:"try_files"`     // 静态文件匹配顺序
	Hide          []string            `json:"hide"`          // 禁止访问的文件或路径
	Precompressed []string            `json:"precompressed"` // 预压缩文件格式
	Proxy         *ProxyOptions       `json:"proxy"`         // 默认反向代理配置
	Redirects     []RedirectOptions   `json:"redirects"`     // 优先执行的重定向规则
	Routes        []Route             `json:"routes"`        // 优先执行的自定义路由
	Headers       HeaderOptions       `json:"headers"`       // 请求和响应头操作
	Compression   CompressionOptions  `json:"compression"`   // 响应压缩配置
	TLS           TLSOptions          `json:"tls"`           // HTTPS 和证书配置
	WAF           WAFOptions          `json:"waf"`           // Web 应用防火墙配置
	Cache         CacheOptions        `json:"cache"`         // 响应缓存配置
	RateLimit     RateLimitOptions    `json:"rate_limit"`    // 访问频率限制配置
	TrafficLimit  TrafficLimitOptions `json:"traffic_limit"` // 并发和响应速度限制配置
	HandlerOrder  []Module            `json:"handler_order"` // 处理器阶段执行顺序
	Handlers      []json.RawMessage   `json:"handlers"`      // 自定义 HTTP 处理器配置
}

// RedirectOptions 定义一条站点重定向规则。
type RedirectOptions struct {
	Name        string   `json:"name"`         // 规则名称
	Enabled     bool     `json:"enabled"`      // 是否启用规则
	Domains     []string `json:"domains"`      // 来源域名，空表示当前站点全部域名
	Paths       []string `json:"paths"`        // 来源路径，空表示全部路径
	Target      string   `json:"target"`       // 绝对 HTTP、HTTPS 地址或站内路径
	Status      int      `json:"status"`       // 重定向响应状态码
	PreserveURI bool     `json:"preserve_uri"` // 是否保留原请求 URI
}

// Route 定义站点内优先于默认处理器执行的路径规则。
type Route struct {
	Name          string            `json:"name"`          // 路由名称
	Paths         []string          `json:"paths"`         // 匹配的请求路径
	Methods       []string          `json:"methods"`       // 匹配的 HTTP 请求方法
	StripPrefix   string            `json:"strip_prefix"`  // 转发前移除的路径前缀
	Rewrite       string            `json:"rewrite"`       // 重写后的请求 URI
	Root          string            `json:"root"`          // 路由静态文件根目录
	Index         []string          `json:"index"`         // 默认首页文件
	Browse        bool              `json:"browse"`        // 是否开启目录浏览
	TryFiles      []string          `json:"try_files"`     // 静态文件匹配顺序
	Hide          []string          `json:"hide"`          // 禁止访问的文件或路径
	Precompressed []string          `json:"precompressed"` // 预压缩文件格式
	Proxy         *ProxyOptions     `json:"proxy"`         // 路由反向代理配置
	Response      *ResponseOptions  `json:"response"`      // 固定响应或跳转配置
	Handlers      []json.RawMessage `json:"handlers"`      // 自定义 HTTP 处理器配置
}

// ProxyOptions 定义 HTTP 或 FastCGI 上游。
type ProxyOptions struct {
	Upstreams             []Upstream        `json:"upstreams"`                // 上游服务地址
	Transport             ProxyTransport    `json:"transport"`                // 代理传输实现，默认 http
	Scheme                ProxyScheme       `json:"scheme"`                   // HTTP 上游协议，默认 http
	Policy                string            `json:"policy"`                   // 上游负载均衡策略，默认 random
	Retries               int               `json:"retries"`                  // 选择可用上游的重试次数
	TryDuration           time.Duration     `json:"try_duration"`             // 选择可用上游的最长时间
	TryInterval           time.Duration     `json:"try_interval"`             // 上游重试间隔
	DialTimeout           time.Duration     `json:"dial_timeout"`             // 建立上游连接的超时时间
	ReadTimeout           time.Duration     `json:"read_timeout"`             // 读取上游响应的超时时间
	WriteTimeout          time.Duration     `json:"write_timeout"`            // 写入上游请求的超时时间
	ResponseHeaderTimeout time.Duration     `json:"response_header_timeout"`  // 等待上游响应头的超时时间
	FlushInterval         time.Duration     `json:"flush_interval"`           // 向客户端刷新响应的间隔
	StreamTimeout         time.Duration     `json:"stream_timeout"`           // 流式响应的最长持续时间
	StreamCloseDelay      time.Duration     `json:"stream_close_delay"`       // 配置重载时关闭流连接的延迟
	Versions              []string          `json:"versions"`                 // 允许使用的上游 HTTP 版本
	TLSServerName         string            `json:"tls_server_name"`          // 上游 TLS SNI 服务器名称
	TLSInsecureSkipVerify *bool             `json:"tls_insecure_skip_verify"` // 是否跳过上游证书校验，默认 true
	HealthURI             string            `json:"health_uri"`               // 主动健康检查请求路径
	HealthInterval        time.Duration     `json:"health_interval"`          // 主动健康检查间隔
	HealthTimeout         time.Duration     `json:"health_timeout"`           // 单次健康检查超时时间
	HealthStatus          int               `json:"health_status"`            // 健康检查期望状态码
	Root                  string            `json:"root"`                     // FastCGI 站点根目录，默认继承站点根目录
	SplitPath             []string          `json:"split_path"`               // FastCGI 脚本路径分割扩展名，默认 .php
	Index                 string            `json:"index"`                    // FastCGI 默认入口文件，默认 index.php
	TryFiles              []string          `json:"try_files"`                // FastCGI 文件匹配顺序
	TryPolicy             string            `json:"try_policy"`               // FastCGI 文件匹配策略，默认 first_exist
	ResolveRootSymlink    bool              `json:"resolve_root_symlink"`     // 是否解析 FastCGI 根目录符号链接
	Hide                  []string          `json:"hide"`                     // FastCGI 静态文件隐藏规则
	Env                   map[string]string `json:"env"`                      // FastCGI 环境变量
	CaptureStderr         bool              `json:"capture_stderr"`           // 是否记录 FastCGI 标准错误输出
	Headers               HeaderOptions     `json:"headers"`                  // 代理请求和响应头操作
}

// Upstream 定义一个静态上游地址。
type Upstream struct {
	Dial        string `json:"dial"`         // 上游网络地址
	MaxRequests int    `json:"max_requests"` // 允许转发到当前上游的最大并发请求数
}

// ResponseOptions 定义固定响应或跳转。
type ResponseOptions struct {
	Status  int                 `json:"status"`  // HTTP 响应状态码
	Body    string              `json:"body"`    // 固定响应内容
	Headers map[string][]string `json:"headers"` // 固定响应头
}

// HeaderOptions 定义请求和响应头的增删改操作。
type HeaderOptions struct {
	RequestAdd     map[string][]string `json:"request_add"`     // 追加的请求头
	RequestSet     map[string][]string `json:"request_set"`     // 覆盖设置的请求头
	RequestDelete  []string            `json:"request_delete"`  // 删除的请求头
	ResponseAdd    map[string][]string `json:"response_add"`    // 追加的响应头
	ResponseSet    map[string][]string `json:"response_set"`    // 覆盖设置的响应头
	ResponseDelete []string            `json:"response_delete"` // 删除的响应头
}

// CompressionOptions 定义响应压缩，默认优先 zstd，其次 gzip。
type CompressionOptions struct {
	Enabled    bool     `json:"enabled"`    // 是否启用响应压缩
	Algorithms []string `json:"algorithms"` // 压缩算法及优先顺序
	MinLength  int      `json:"min_length"` // 启用压缩的最小响应字节数
}

// TLSOptions 定义手动部署的证书。启用 TLS 不会触发自动申请。
type TLSOptions struct {
	Enabled        bool              `json:"enabled"`       // 是否启用 HTTPS
	RedirectHTTP   bool              `json:"redirect_http"` // 是否将已覆盖域名的 HTTP 重定向到 HTTPS
	MinVersion     string            `json:"min_version"`   // 允许的最低 TLS 版本
	MaxVersion     string            `json:"max_version"`   // 允许的最高 TLS 版本
	Certificates   []CertificatePair `json:"certificates"`  // 站点使用的证书
	coveredDomains []string          // 当前证书实际覆盖的站点域名，仅用于生成运行配置
}

// CertificatePair 支持从文件或内存 PEM 加载证书，两个来源只能选择一个。
type CertificatePair struct {
	CertificateFile string   `json:"certificate_file"` // 证书链文件路径
	KeyFile         string   `json:"key_file"`         // 私钥文件路径
	CertificatePEM  string   `json:"certificate_pem"`  // PEM 格式证书链内容
	PrivateKeyPEM   string   `json:"-"`                // PEM 格式私钥内容，不参与 JSON 序列化
	Domains         []string `json:"domains"`          // 此证书在当前站点明确启用的域名，空表示自动匹配
}

// WAFOptions 定义 Coraza WAF。默认模式为 DetectionOnly。
type WAFOptions struct {
	Enabled          bool         `json:"enabled"`            // 是否启用 WAF
	Mode             WAFMode      `json:"mode"`               // WAF 检测或拦截模式
	BlockPage        string       `json:"block_page"`         // 自定义拦截 HTML，留空使用默认页面
	OWASP            OWASPOptions `json:"owasp"`              // OWASP CRS 配置
	AuditLog         bool         `json:"audit_log"`          // 是否记录 WAF 审计日志
	RequestBodyLimit int64        `json:"request_body_limit"` // 可检查的请求体最大字节数
	Directives       []string     `json:"directives"`         // 自定义 Coraza 指令
	Rules            []WAFRule    `json:"rules"`              // 自定义 WAF 规则
}

// OWASPOptions 定义内嵌 OWASP Core Rule Set 的运行参数。
type OWASPOptions struct {
	Enabled                       bool     `json:"enabled"`                          // 是否启用 OWASP CRS
	ParanoiaLevel                 int      `json:"paranoia_level"`                   // 阻断规则敏感级别，范围 1-4，默认 1
	DetectionParanoiaLevel        int      `json:"detection_paranoia_level"`         // 仅检测规则敏感级别，默认等于阻断级别
	InboundAnomalyScoreThreshold  int      `json:"inbound_anomaly_score_threshold"`  // 入站请求异常分数阈值，默认 5
	OutboundAnomalyScoreThreshold int      `json:"outbound_anomaly_score_threshold"` // 出站响应异常分数阈值，默认 4
	ReportingLevel                *int     `json:"reporting_level"`                  // CRS 日志报告级别，范围 0-5
	EarlyBlocking                 bool     `json:"early_blocking"`                   // 是否提前执行异常分数阻断
	SamplingPercentage            int      `json:"sampling_percentage"`              // 接受 CRS 检查的请求百分比，默认 100
	EnforceBodyProcessor          bool     `json:"enforce_body_processor"`           // 是否强制处理 URL 编码请求体
	ValidateUTF8                  bool     `json:"validate_utf8"`                    // 是否校验请求数据的 UTF-8 编码
	SkipResponseAnalysis          bool     `json:"skip_response_analysis"`           // 是否跳过出站响应分析
	AllowedMethods                []string `json:"allowed_methods"`                  // 允许的 HTTP 请求方法
	AllowedContentTypes           []string `json:"allowed_content_types"`            // 允许的请求 Content-Type
	AllowedHTTPVersions           []string `json:"allowed_http_versions"`            // 允许的 HTTP 协议版本
	AllowedCharsets               []string `json:"allowed_charsets"`                 // 允许的请求字符集
	RestrictedExtensions          []string `json:"restricted_extensions"`            // 禁止上传或请求的文件扩展名
	RestrictedHeaders             []string `json:"restricted_headers"`               // 基础受限请求头
	RestrictedHeadersExtended     []string `json:"restricted_headers_extended"`      // 扩展受限请求头
	MaxArguments                  int      `json:"max_arguments"`                    // 单个请求允许的最大参数数量
	MaxArgumentNameLength         int      `json:"max_argument_name_length"`         // 单个参数名最大字节数
	MaxArgumentLength             int      `json:"max_argument_length"`              // 单个参数值最大字节数
	TotalArgumentLength           int      `json:"total_argument_length"`            // 所有参数名和值的最大总字节数
	MaxFileSize                   int64    `json:"max_file_size"`                    // 单个上传文件最大字节数
	CombinedFileSize              int64    `json:"combined_file_size"`               // 单次请求上传文件最大总字节数
	SetupDirectives               []string `json:"setup_directives"`                 // CRS 规则加载前执行的自定义指令
}

// WAFRule 保存一条可由业务层独立开关的 WAF 规则。
type WAFRule struct {
	ID         string         `json:"id"`         // 规则唯一标识
	Name       string         `json:"name"`       // 规则显示名称
	Enabled    bool           `json:"enabled"`    // 是否启用规则
	Match      WAFMatch       `json:"match"`      // 多个条件组之间的匹配逻辑
	Action     WAFAction      `json:"action"`     // 规则命中后的处理动作
	Conditions []WAFCondition `json:"conditions"` // 旧版扁平条件，读取已有配置时继续兼容
	Groups     []WAFGroup     `json:"groups"`     // 最多两层的条件分组
	Directive  string         `json:"directive"`  // 旧版 Coraza 规则指令，仅用于兼容已有配置
}

// WAFGroup 定义 WAF 规则中的一组匹配条件。
type WAFGroup struct {
	Match      WAFMatch       `json:"match"`      // 组内条件的匹配逻辑
	Conditions []WAFCondition `json:"conditions"` // 当前组包含的匹配条件
}

// WAFCondition 定义 WAF 规则中的一个匹配条件。
type WAFCondition struct {
	Target     string      `json:"target"`      // 检查的请求数据
	Operator   WAFOperator `json:"operator"`    // 匹配操作符
	Value      string      `json:"value"`       // 匹配内容
	Negated    bool        `json:"negated"`     // 是否对当前匹配结果取反
	IgnoreCase bool        `json:"ignore_case"` // 是否忽略英文字母大小写
}

// CacheOptions 定义站点响应缓存。默认 no-store，必须显式允许缓存动态响应。
type CacheOptions struct {
	Enabled             bool          `json:"enabled"`               // 是否启用响应缓存
	TTL                 time.Duration `json:"ttl"`                   // 缓存新鲜内容的有效时间
	Stale               time.Duration `json:"stale"`                 // 过期内容允许继续使用的时间
	Methods             []string      `json:"methods"`               // 允许缓存的 HTTP 请求方法
	KeyHeaders          []string      `json:"key_headers"`           // 参与生成缓存键的请求头
	DefaultCacheControl string        `json:"default_cache_control"` // 响应未声明时使用的缓存策略
	Name                string        `json:"name"`                  // 缓存实例名称
	MaxSize             int64         `json:"max_size"`              // 缓存容量及单个响应体上限，单位字节
}

// RateLimitOptions 定义单机滑动窗口限流。
type RateLimitOptions struct {
	Enabled    bool          `json:"enabled"`     // 是否启用访问频率限制
	Key        string        `json:"key"`         // 限流对象键或占位符表达式
	Window     time.Duration `json:"window"`      // 滑动统计时间窗口
	MaxEvents  int           `json:"max_events"`  // 窗口内允许的最大请求数
	Paths      []string      `json:"paths"`       // 应用限流的请求路径
	Methods    []string      `json:"methods"`     // 应用限流的 HTTP 请求方法
	IPv4Prefix int           `json:"ipv4_prefix"` // IPv4 地址聚合前缀长度
	IPv6Prefix int           `json:"ipv6_prefix"` // IPv6 地址聚合前缀长度
}

// TrafficLimitOptions 定义站点并发和单个请求的响应速度限制，值为 0 表示不限制。
type TrafficLimitOptions struct {
	Enabled             bool  `json:"enabled"`                // 是否启用流量限制
	MaxConnections      int64 `json:"max_connections"`        // 站点最大并发请求数
	MaxConnectionsPerIP int64 `json:"max_connections_per_ip"` // 单个客户端 IP 最大并发请求数
	RatePerRequest      int64 `json:"rate_per_request"`       // 单个请求响应速度，单位为字节/秒
}

// DNSCredentials 定义 ACME DNS-01 验证凭据，未使用的字段保持为空。
type DNSCredentials struct {
	AliyunAccessKeyID     string `json:"aliyun_access_key_id"`     // 阿里云 AccessKey ID
	AliyunAccessKeySecret string `json:"aliyun_access_key_secret"` // 阿里云 AccessKey Secret
	DNSPodAPIToken        string `json:"dnspod_api_token"`         // DNSPod API Token，格式为 ID,TOKEN
	TencentSecretID       string `json:"tencent_secret_id"`        // 腾讯云 Secret ID
	TencentSecretKey      string `json:"tencent_secret_key"`       // 腾讯云 Secret Key
	HuaweiAccessKeyID     string `json:"huawei_access_key_id"`     // 华为云 AccessKey ID
	HuaweiSecretAccessKey string `json:"huawei_secret_access_key"` // 华为云 Secret Access Key
}

// CertificateRequest 定义一次显式 ACME 申请。
type CertificateRequest struct {
	Domain         string               `json:"domain"`          // 申请证书的域名或公网 IP 地址
	Email          string               `json:"email"`           // ACME 账户邮箱，为空时使用全局配置
	CA             CertificateCA        `json:"ca"`              // 证书签发环境，默认 production
	Challenge      CertificateChallenge `json:"challenge"`       // 所有权验证方式，默认 http，IP 地址仅支持 http
	DNSProvider    DNSProvider          `json:"dns_provider"`    // DNS-01 服务商
	DNSCredentials DNSCredentials       `json:"dns_credentials"` // DNS-01 服务商凭据
}

// Certificate 返回证书信息和 PEM。私钥仅供 Go 调用方使用，不参与 JSON 序列化。
type Certificate struct {
	Domain          string    `json:"domain"`           // 申请证书的主域名或 IP 地址
	Issuer          string    `json:"issuer"`           // 证书签发环境
	DNSNames        []string  `json:"dns_names"`        // 证书覆盖的 DNS 名称
	IPAddresses     []string  `json:"ip_addresses"`     // 证书覆盖的 IP 地址
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
	root        string
	dataPath    string
	acmePath    string
	cachePath   string
	logPath     string
	wafLogPath  string
	configPath  string
	options     Options
	acmeStorage *certmagic.FileStorage

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
