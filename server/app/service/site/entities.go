package site

import (
	"time"

	"server/app/model"
	webServer "server/app/util/server"
	"server/app/util/str"
)

// HTTPConfig 是四类网站共享的 HTTP 配置。
type HTTPConfig struct {
	Redirects    []RedirectConfig              `json:"redirects"`     // 重定向规则
	Routes       []RouteConfig                 `json:"routes"`        // 自定义路由
	Headers      webServer.HeaderOptions       `json:"headers"`       // 请求和响应头
	Compression  webServer.CompressionOptions  `json:"compression"`   // 响应压缩
	TLS          TLSConfig                     `json:"tls"`           // HTTPS 策略
	WAF          webServer.WAFOptions          `json:"waf"`           // Web 应用防火墙
	Cache        webServer.CacheOptions        `json:"cache"`         // 响应缓存
	RateLimit    webServer.RateLimitOptions    `json:"rate_limit"`    // 请求频率限制
	TrafficLimit webServer.TrafficLimitOptions `json:"traffic_limit"` // 并发和响应速度限制
}

// RedirectConfig 定义前端可维护的站点重定向规则。
type RedirectConfig struct {
	Name        string   `json:"name"`         // 规则名称
	Enabled     bool     `json:"enabled"`      // 是否启用规则
	Domains     []string `json:"domains"`      // 来源域名，空表示当前站点全部域名
	Paths       []string `json:"paths"`        // 来源路径，空表示全部路径
	Target      string   `json:"target"`       // 绝对 HTTP、HTTPS 地址或站内路径
	Status      int      `json:"status"`       // 重定向响应状态码
	PreserveURI bool     `json:"preserve_uri"` // 是否保留原请求 URI
}

// RouteConfig 定义前端可维护的站点路由，不暴露底层原始处理器配置。
type RouteConfig struct {
	Name          string                     `json:"name"`          // 路由名称
	Paths         []string                   `json:"paths"`         // 匹配的请求路径
	Methods       []string                   `json:"methods"`       // 匹配的 HTTP 请求方法
	StripPrefix   string                     `json:"strip_prefix"`  // 转发前移除的路径前缀
	Rewrite       string                     `json:"rewrite"`       // 重写后的请求 URI
	Root          string                     `json:"root"`          // 路由静态文件根目录
	Index         []string                   `json:"index"`         // 默认首页文件
	Browse        bool                       `json:"browse"`        // 是否开启目录浏览
	TryFiles      []string                   `json:"try_files"`     // 静态文件匹配顺序
	Hide          []string                   `json:"hide"`          // 禁止访问的文件或路径
	Precompressed []string                   `json:"precompressed"` // 预压缩文件格式
	Proxy         *webServer.ProxyOptions    `json:"proxy"`         // 路由反向代理配置
	Response      *webServer.ResponseOptions `json:"response"`      // 固定响应或跳转配置
}

// TLSConfig 保存站点级 TLS 策略，证书由域名表的 cert_id 关联。
type TLSConfig struct {
	RedirectHTTP bool   `json:"redirect_http"` // 是否将已开启 SSL 的域名跳转到 HTTPS
	MinVersion   string `json:"min_version"`   // 最低 TLS 版本
	MaxVersion   string `json:"max_version"`   // 最高 TLS 版本
}

// SSLSettings 保存 TLS 策略和网站当前域名证书绑定。
type SSLSettings struct {
	Config  TLSConfig       `json:"config"`  // TLS 策略
	Domains []*model.Domain `json:"domains"` // 域名及证书绑定
}

// DomainCertificate 保存一次域名证书绑定更新。
type DomainCertificate struct {
	DomainId int64 `json:"domain_id"` // 域名 ID
	CertId   int64 `json:"cert_id"`   // 证书 ID，0 表示关闭 SSL
}

// StaticOptions 保存静态网站专属设置。
type StaticOptions struct {
	Index         []string `json:"index"`         // 默认首页文件
	Browse        bool     `json:"browse"`        // 是否开启目录浏览
	TryFiles      []string `json:"try_files"`     // 静态文件匹配顺序
	Hide          []string `json:"hide"`          // 禁止访问的文件或路径
	Precompressed []string `json:"precompressed"` // 预压缩文件格式
}

// CacheConfig 保存客户端可修改的站点缓存设置，缓存实例名称由服务端生成。
type CacheConfig struct {
	Enabled             bool          `json:"enabled"`               // 是否启用响应缓存
	TTL                 time.Duration `json:"ttl"`                   // 缓存新鲜内容的有效时间
	Stale               time.Duration `json:"stale"`                 // 过期内容允许继续使用的时间
	Methods             []string      `json:"methods"`               // 允许缓存的 HTTP 请求方法
	KeyHeaders          []string      `json:"key_headers"`           // 参与生成缓存键的请求头
	DefaultCacheControl string        `json:"default_cache_control"` // 响应未声明时使用的缓存策略
	MaxSize             int64         `json:"max_size"`              // 缓存容量及单个响应体上限，单位字节
}

// TLSUpdate 部分更新站点 TLS 策略，nil 字段保持原值。
type TLSUpdate struct {
	RedirectHTTP *bool   `json:"redirect_http"` // 是否将已开启 SSL 的域名跳转到 HTTPS
	MinVersion   *string `json:"min_version"`   // 最低 TLS 版本
	MaxVersion   *string `json:"max_version"`   // 最高 TLS 版本
}

// OWASPUpdate 部分更新 OWASP CRS 设置，nil 字段保持原值。
type OWASPUpdate struct {
	Enabled                       *bool     `json:"enabled"`                          // 是否启用 OWASP CRS
	ParanoiaLevel                 *int      `json:"paranoia_level"`                   // 阻断规则敏感级别
	DetectionParanoiaLevel        *int      `json:"detection_paranoia_level"`         // 仅检测规则敏感级别
	InboundAnomalyScoreThreshold  *int      `json:"inbound_anomaly_score_threshold"`  // 入站请求异常分数阈值
	OutboundAnomalyScoreThreshold *int      `json:"outbound_anomaly_score_threshold"` // 出站响应异常分数阈值
	ReportingLevel                *int      `json:"reporting_level"`                  // CRS 日志报告级别
	EarlyBlocking                 *bool     `json:"early_blocking"`                   // 是否提前执行异常分数阻断
	SamplingPercentage            *int      `json:"sampling_percentage"`              // 接受 CRS 检查的请求百分比
	EnforceBodyProcessor          *bool     `json:"enforce_body_processor"`           // 是否强制处理 URL 编码请求体
	ValidateUTF8                  *bool     `json:"validate_utf8"`                    // 是否校验请求数据的 UTF-8 编码
	SkipResponseAnalysis          *bool     `json:"skip_response_analysis"`           // 是否跳过出站响应分析
	AllowedMethods                *[]string `json:"allowed_methods"`                  // 允许的 HTTP 请求方法
	AllowedContentTypes           *[]string `json:"allowed_content_types"`            // 允许的请求 Content-Type
	AllowedHTTPVersions           *[]string `json:"allowed_http_versions"`            // 允许的 HTTP 协议版本
	AllowedCharsets               *[]string `json:"allowed_charsets"`                 // 允许的请求字符集
	RestrictedExtensions          *[]string `json:"restricted_extensions"`            // 禁止上传或请求的文件扩展名
	RestrictedHeaders             *[]string `json:"restricted_headers"`               // 基础受限请求头
	RestrictedHeadersExtended     *[]string `json:"restricted_headers_extended"`      // 扩展受限请求头
	MaxArguments                  *int      `json:"max_arguments"`                    // 单个请求允许的最大参数数量
	MaxArgumentNameLength         *int      `json:"max_argument_name_length"`         // 单个参数名最大字节数
	MaxArgumentLength             *int      `json:"max_argument_length"`              // 单个参数值最大字节数
	TotalArgumentLength           *int      `json:"total_argument_length"`            // 所有参数名和值的最大总字节数
	MaxFileSize                   *int64    `json:"max_file_size"`                    // 单个上传文件最大字节数
	CombinedFileSize              *int64    `json:"combined_file_size"`               // 单次请求上传文件最大总字节数
	SetupDirectives               *[]string `json:"setup_directives"`                 // CRS 规则加载前执行的自定义指令
}

// WAFUpdate 部分更新网站 WAF 设置，nil 字段保持原值。
type WAFUpdate struct {
	Enabled          *bool                `json:"enabled"`            // 是否启用 WAF
	Mode             *webServer.WAFMode   `json:"mode"`               // WAF 检测或拦截模式
	OWASP            *OWASPUpdate         `json:"owasp"`              // OWASP CRS 设置
	AuditLog         *bool                `json:"audit_log"`          // 是否记录 WAF 审计日志
	RequestBodyLimit *int64               `json:"request_body_limit"` // 可检查的请求体最大字节数
	Directives       *[]string            `json:"directives"`         // 自定义 Coraza 指令
	Rules            *[]webServer.WAFRule `json:"rules"`              // 自定义 WAF 规则
}

// CacheUpdate 部分更新网站缓存设置，nil 字段保持原值。
type CacheUpdate struct {
	Enabled             *bool          `json:"enabled"`               // 是否启用响应缓存
	TTL                 *time.Duration `json:"ttl"`                   // 缓存新鲜内容的有效时间
	Stale               *time.Duration `json:"stale"`                 // 过期内容允许继续使用的时间
	Methods             *[]string      `json:"methods"`               // 允许缓存的 HTTP 请求方法
	KeyHeaders          *[]string      `json:"key_headers"`           // 参与生成缓存键的请求头
	DefaultCacheControl *string        `json:"default_cache_control"` // 响应未声明时使用的缓存策略
	MaxSize             *int64         `json:"max_size"`              // 缓存容量及单个响应体上限
}

// RateLimitUpdate 部分更新网站访问频率限制，nil 字段保持原值。
type RateLimitUpdate struct {
	Enabled    *bool          `json:"enabled"`     // 是否启用访问频率限制
	Key        *string        `json:"key"`         // 限流对象键或占位符表达式
	Window     *time.Duration `json:"window"`      // 滑动统计时间窗口
	MaxEvents  *int           `json:"max_events"`  // 窗口内允许的最大请求数
	Paths      *[]string      `json:"paths"`       // 应用限流的请求路径
	Methods    *[]string      `json:"methods"`     // 应用限流的 HTTP 请求方法
	IPv4Prefix *int           `json:"ipv4_prefix"` // IPv4 地址聚合前缀长度
	IPv6Prefix *int           `json:"ipv6_prefix"` // IPv6 地址聚合前缀长度
}

// TrafficLimitUpdate 部分更新网站并发和响应速度限制，nil 字段保持原值。
type TrafficLimitUpdate struct {
	Enabled             *bool  `json:"enabled"`                // 是否启用流量限制
	MaxConnections      *int64 `json:"max_connections"`        // 站点最大并发请求数
	MaxConnectionsPerIP *int64 `json:"max_connections_per_ip"` // 单个客户端 IP 最大并发请求数
	RatePerRequest      *int64 `json:"rate_per_request"`       // 单个请求响应速度
}

// HeaderUpdate 部分更新请求头和响应头设置，nil 字段保持原值。
type HeaderUpdate struct {
	RequestAdd     *map[string][]string `json:"request_add"`     // 追加的请求头
	RequestSet     *map[string][]string `json:"request_set"`     // 覆盖设置的请求头
	RequestDelete  *[]string            `json:"request_delete"`  // 删除的请求头
	ResponseAdd    *map[string][]string `json:"response_add"`    // 追加的响应头
	ResponseSet    *map[string][]string `json:"response_set"`    // 覆盖设置的响应头
	ResponseDelete *[]string            `json:"response_delete"` // 删除的响应头
}

// CompressionUpdate 部分更新网站响应压缩设置，nil 字段保持原值。
type CompressionUpdate struct {
	Enabled    *bool     `json:"enabled"`    // 是否启用响应压缩
	Algorithms *[]string `json:"algorithms"` // 压缩算法及优先顺序
	MinLength  *int      `json:"min_length"` // 启用压缩的最小响应字节数
}

// StaticUpdate 部分更新静态网站设置，nil 字段保持原值。
type StaticUpdate struct {
	Index         *[]string `json:"index"`         // 默认首页文件
	Browse        *bool     `json:"browse"`        // 是否开启目录浏览
	TryFiles      *[]string `json:"try_files"`     // 静态文件匹配顺序
	Hide          *[]string `json:"hide"`          // 禁止访问的文件或路径
	Precompressed *[]string `json:"precompressed"` // 预压缩文件格式
}

// ProxyUpdate 部分更新 HTTP 反向代理设置，nil 字段保持原值。
type ProxyUpdate struct {
	Upstreams             *[]webServer.Upstream  `json:"upstreams"`                // 上游服务地址
	Scheme                *webServer.ProxyScheme `json:"scheme"`                   // HTTP 上游协议
	Policy                *string                `json:"policy"`                   // 上游负载均衡策略
	Retries               *int                   `json:"retries"`                  // 选择可用上游的重试次数
	TryDuration           *time.Duration         `json:"try_duration"`             // 选择可用上游的最长时间
	TryInterval           *time.Duration         `json:"try_interval"`             // 上游重试间隔
	DialTimeout           *time.Duration         `json:"dial_timeout"`             // 建立上游连接的超时时间
	ReadTimeout           *time.Duration         `json:"read_timeout"`             // 读取上游响应的超时时间
	WriteTimeout          *time.Duration         `json:"write_timeout"`            // 写入上游请求的超时时间
	ResponseHeaderTimeout *time.Duration         `json:"response_header_timeout"`  // 等待上游响应头的超时时间
	FlushInterval         *time.Duration         `json:"flush_interval"`           // 向客户端刷新响应的间隔
	StreamTimeout         *time.Duration         `json:"stream_timeout"`           // 流式响应的最长持续时间
	StreamCloseDelay      *time.Duration         `json:"stream_close_delay"`       // 配置重载时关闭流连接的延迟
	Versions              *[]string              `json:"versions"`                 // 允许使用的上游 HTTP 版本
	TLSServerName         *string                `json:"tls_server_name"`          // 上游 TLS SNI 服务器名称
	TLSInsecureSkipVerify *bool                  `json:"tls_insecure_skip_verify"` // 是否跳过上游证书校验
	HealthURI             *string                `json:"health_uri"`               // 主动健康检查请求路径
	HealthInterval        *time.Duration         `json:"health_interval"`          // 主动健康检查间隔
	HealthTimeout         *time.Duration         `json:"health_timeout"`           // 单次健康检查超时时间
	HealthStatus          *int                   `json:"health_status"`            // 健康检查期望状态码
	Headers               *HeaderUpdate          `json:"headers"`                  // 代理请求和响应头操作
}

// FastCGIUpdate 部分更新 PHP-FPM FastCGI 设置，nil 字段保持原值。
type FastCGIUpdate struct {
	Upstreams          *[]webServer.Upstream `json:"upstreams"`            // PHP-FPM 上游服务地址
	Policy             *string               `json:"policy"`               // 上游负载均衡策略
	Retries            *int                  `json:"retries"`              // 选择可用上游的重试次数
	TryDuration        *time.Duration        `json:"try_duration"`         // 选择可用上游的最长时间
	TryInterval        *time.Duration        `json:"try_interval"`         // 上游重试间隔
	DialTimeout        *time.Duration        `json:"dial_timeout"`         // 建立上游连接的超时时间
	ReadTimeout        *time.Duration        `json:"read_timeout"`         // 读取上游响应的超时时间
	WriteTimeout       *time.Duration        `json:"write_timeout"`        // 写入上游请求的超时时间
	FlushInterval      *time.Duration        `json:"flush_interval"`       // 向客户端刷新响应的间隔
	StreamTimeout      *time.Duration        `json:"stream_timeout"`       // 流式响应的最长持续时间
	StreamCloseDelay   *time.Duration        `json:"stream_close_delay"`   // 配置重载时关闭流连接的延迟
	HealthURI          *string               `json:"health_uri"`           // 主动健康检查请求路径
	HealthInterval     *time.Duration        `json:"health_interval"`      // 主动健康检查间隔
	HealthTimeout      *time.Duration        `json:"health_timeout"`       // 单次健康检查超时时间
	HealthStatus       *int                  `json:"health_status"`        // 健康检查期望状态码
	SplitPath          *[]string             `json:"split_path"`           // FastCGI 脚本路径分割扩展名
	Index              *string               `json:"index"`                // FastCGI 默认入口文件
	TryFiles           *[]string             `json:"try_files"`            // FastCGI 文件匹配顺序
	TryPolicy          *string               `json:"try_policy"`           // FastCGI 文件匹配策略
	ResolveRootSymlink *bool                 `json:"resolve_root_symlink"` // 是否解析 FastCGI 根目录符号链接
	Hide               *[]string             `json:"hide"`                 // FastCGI 静态文件隐藏规则
	Env                *map[string]string    `json:"env"`                  // FastCGI 环境变量
	CaptureStderr      *bool                 `json:"capture_stderr"`       // 是否记录 FastCGI 标准错误输出
	Headers            *HeaderUpdate         `json:"headers"`              // 代理请求和响应头操作
}

// ProcessUpdate 部分更新通用网站进程保活设置，nil 字段保持原值。
type ProcessUpdate struct {
	Command      *string            `json:"command"`       // 完整启动命令行
	Environment  *map[string]string `json:"environment"`   // 追加的环境变量
	RestartDelay *time.Duration     `json:"restart_delay"` // 异常退出后的重启等待时间
	StopTimeout  *time.Duration     `json:"stop_timeout"`  // 停止命令的最长等待时间
}

// StaticConfig 静态网站配置。
type StaticConfig struct {
	HTTPConfig
	Index         []string `json:"index"`         // 默认首页文件
	Browse        bool     `json:"browse"`        // 是否开启目录浏览
	TryFiles      []string `json:"try_files"`     // 静态文件匹配顺序
	Hide          []string `json:"hide"`          // 禁止访问的文件或路径
	Precompressed []string `json:"precompressed"` // 预压缩文件格式
}

// ProxyConfig 反向代理网站配置。
type ProxyConfig struct {
	HTTPConfig
	Proxy webServer.ProxyOptions `json:"proxy"` // HTTP 上游配置
}

// PHPConfig PHP 网站配置。
type PHPConfig struct {
	HTTPConfig
	FastCGI webServer.ProxyOptions `json:"fastcgi"` // PHP-FPM FastCGI 配置
}

// GeneralConfig 通用网站配置。
type GeneralConfig struct {
	HTTPConfig
	Proxy   webServer.ProxyOptions `json:"proxy"`   // 本地进程的 HTTP 上游配置
	Process ProcessConfig          `json:"process"` // 进程保活配置
}

// ProcessConfig 定义通用网站的启动命令。
type ProcessConfig struct {
	Command      string            `json:"command"`       // 完整启动命令行
	Environment  map[string]string `json:"environment"`   // 追加的环境变量
	RestartDelay time.Duration     `json:"restart_delay"` // 异常退出后的重启等待时间
	StopTimeout  time.Duration     `json:"stop_timeout"`  // 停止命令的最长等待时间
}

// Domain 保存网站域名和证书绑定。
type Domain struct {
	Domain string `json:"domain"`  // 域名或通配符域名
	CertId int64  `json:"cert_id"` // 证书 ID，0 表示仅使用 HTTP
}

// CreateSite 创建网站参数。
type CreateSite struct {
	Name    string                  `json:"name"`     // 网站名称
	Type    int                     `json:"type"`     // 网站类型
	Status  int                     `json:"status"`   // 网站状态
	Root    string                  `json:"root"`     // 网站根目录
	RunPath string                  `json:"run_path"` // 相对网站根目录的运行目录
	Domains []string                `json:"domains"`  // 网站绑定的域名
	Static  *StaticOptions          `json:"static"`   // 静态网站配置
	Proxy   *webServer.ProxyOptions `json:"proxy"`    // 反向代理网站配置
	FastCGI *webServer.ProxyOptions `json:"fastcgi"`  // PHP 网站 FastCGI 配置
	Process *ProcessConfig          `json:"process"`  // 通用网站进程配置
}

// UpdateSite 更新网站参数，nil 字段保持原值。
type UpdateSite struct {
	Name    *string `json:"name"`     // 网站名称
	Root    *string `json:"root"`     // 网站根目录
	RunPath *string `json:"run_path"` // 相对网站根目录的运行目录
}

// Site 网站基础详情，独立设置通过对应 Get 方法读取。
type Site struct {
	Id         int64        `json:"id"`          // 网站 ID
	Name       string       `json:"name"`        // 网站名称
	Type       int          `json:"type"`        // 网站类型
	Status     int          `json:"status"`      // 网站状态
	Root       string       `json:"root"`        // 网站根目录
	RunPath    string       `json:"run_path"`    // 网站运行目录
	CreateTime str.DateTime `json:"create_time"` // 创建时间
	UpdateTime str.DateTime `json:"update_time"` // 更新时间
}

// Options 配置网站 HTTP 运行目录和全局监听参数。
type Options struct {
	Root string            // 网站服务数据目录
	HTTP webServer.Options // HTTP Manager 全局配置
}

// HTTPPageUpdate 部分更新 HTTP 固定页面，nil 字段保持原值。
type HTTPPageUpdate struct {
	Status  *int                 `json:"status"`  // HTTP 响应状态码
	Body    *string              `json:"body"`    // 页面响应内容
	Headers *map[string][]string `json:"headers"` // 页面响应头
}

// HTTPUpdate 部分更新 HTTP 服务全局参数，nil 字段保持原值。
type HTTPUpdate struct {
	HTTPListen           *[]string       `json:"http_listen"`            // HTTP 监听地址
	HTTPSListen          *[]string       `json:"https_listen"`           // HTTPS 监听地址
	Protocols            *[]string       `json:"protocols"`              // 启用的 HTTP 协议
	DefaultSite          *string         `json:"default_site"`           // 默认站点 ID
	NotFoundPage         *HTTPPageUpdate `json:"not_found_page"`         // 站点内资源不存在页面
	SiteNotFoundPage     *HTTPPageUpdate `json:"site_not_found_page"`    // 请求域名未绑定网站时显示的页面
	SiteDisabledPage     *HTTPPageUpdate `json:"site_disabled_page"`     // 请求域名所属网站已停用时显示的页面
	DataPath             *string         `json:"data_path"`              // HTTP 运行数据目录
	CachePath            *string         `json:"cache_path"`             // 站点响应缓存根目录
	LogPath              *string         `json:"log_path"`               // HTTP 运行日志文件
	WAFLogPath           *string         `json:"waf_log_path"`           // WAF 审计日志文件
	ConfigPath           *string         `json:"config_path"`            // 配置快照文件
	LogLevel             *string         `json:"log_level"`              // HTTP 日志级别
	TrustedProxies       *[]string       `json:"trusted_proxies"`        // 可信代理 IP 或 CIDR
	ClientIPHeaders      *[]string       `json:"client_ip_headers"`      // 客户端 IP 请求头
	TrustedProxiesStrict *bool           `json:"trusted_proxies_strict"` // 是否严格解析可信代理链
	ReadTimeout          *time.Duration  `json:"read_timeout"`           // 读取完整请求的超时时间
	ReadHeaderTimeout    *time.Duration  `json:"read_header_timeout"`    // 读取请求头的超时时间
	WriteTimeout         *time.Duration  `json:"write_timeout"`          // 写入响应的超时时间
	IdleTimeout          *time.Duration  `json:"idle_timeout"`           // 空闲连接的超时时间
	GracePeriod          *time.Duration  `json:"grace_period"`           // HTTP 服务优雅关闭等待时间
	MaxHeaderBytes       *int            `json:"max_header_bytes"`       // 单个请求头最大字节数
	HTTPChallengeHost    *string         `json:"http_challenge_host"`    // ACME HTTP-01 验证监听地址
	HTTPChallengePort    *int            `json:"http_challenge_port"`    // ACME HTTP-01 验证端口
	ACMEEmail            *string         `json:"acme_email"`             // ACME 账户默认邮箱
}
