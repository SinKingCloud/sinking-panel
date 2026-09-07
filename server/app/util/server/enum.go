package server

// Module 表示 HTTP 处理器模块名称。
type Module string

// ProxyTransport 表示反向代理传输实现。
type ProxyTransport string

// ProxyScheme 表示 HTTP 上游协议。
type ProxyScheme string

// WAFMode 表示 WAF 运行模式。
type WAFMode string

// WAFMatch 表示一组 WAF 条件的匹配逻辑。
type WAFMatch string

// WAFAction 表示 WAF 规则命中后的动作。
type WAFAction string

// WAFOperator 表示 WAF 条件的匹配操作符。
type WAFOperator string

// CertificateCA 表示证书签发环境。
type CertificateCA string

// CertificateChallenge 表示 ACME 域名所有权验证方式。
type CertificateChallenge string

// DNSProvider 表示 ACME DNS-01 服务商。
type DNSProvider string

// LogType 表示 HTTP 服务和站点日志类型。
type LogType string

const (
	ProxyTransportHTTP    ProxyTransport = "http"    // HTTP 反向代理传输
	ProxyTransportFastCGI ProxyTransport = "fastcgi" // FastCGI 反向代理传输
	ProxySchemeHTTP       ProxyScheme    = "http"    // HTTP 上游协议
	ProxySchemeHTTPS      ProxyScheme    = "https"   // HTTPS 上游协议

	HandlerWAF            Module = "waf"             // WAF 处理器
	HandlerVars           Module = "vars"            // 请求上下文变量处理器
	HandlerRateLimit      Module = "rate_limit"      // 访问频率限制处理器
	HandlerTrafficLimit   Module = "traffic_limit"   // 并发和响应速度限制处理器
	HandlerHeaders        Module = "headers"         // 请求和响应头处理器
	HandlerCompression    Module = "encode"          // 响应压缩处理器
	HandlerCache          Module = "cache"           // 响应缓存处理器
	HandlerSubroute       Module = "subroute"        // 子路由处理器
	HandlerRewrite        Module = "rewrite"         // 请求重写处理器
	HandlerStaticResponse Module = "static_response" // 固定响应处理器
	HandlerFileServer     Module = "file_server"     // 静态文件处理器
	HandlerReverseProxy   Module = "reverse_proxy"   // 反向代理处理器

	WAFModeDetection WAFMode = "detection" // WAF 仅记录不拦截
	WAFModeBlock     WAFMode = "block"     // WAF 检测并拦截

	WAFMatchAny WAFMatch = "any" // 任一条件满足即命中
	WAFMatchAll WAFMatch = "all" // 全部条件满足才命中

	WAFActionBlock WAFAction = "block" // 拦截命中的请求
	WAFActionLog   WAFAction = "log"   // 仅记录命中的请求

	WAFOperatorContains   WAFOperator = "contains"     // 包含指定内容
	WAFOperatorEquals     WAFOperator = "equals"       // 等于指定内容
	WAFOperatorStartsWith WAFOperator = "starts_with"  // 以指定内容开头
	WAFOperatorEndsWith   WAFOperator = "ends_with"    // 以指定内容结尾
	WAFOperatorRegex      WAFOperator = "regex"        // 匹配正则表达式
	WAFOperatorIP         WAFOperator = "ip"           // 匹配 IP 或 CIDR
	WAFOperatorExists     WAFOperator = "exists"       // 检查对象是否存在
	WAFOperatorNotEmpty   WAFOperator = "not_empty"    // 检查对象内容是否非空
	WAFOperatorGreater    WAFOperator = "greater_than" // 大于指定整数
	WAFOperatorLess       WAFOperator = "less_than"    // 小于指定整数

	CertificateCAProd    CertificateCA = "production" // Let's Encrypt 正式环境
	CertificateCAStaging CertificateCA = "staging"    // Let's Encrypt 测试环境

	CertificateChallengeHTTP CertificateChallenge = "http" // HTTP-01 验证
	CertificateChallengeDNS  CertificateChallenge = "dns"  // DNS-01 验证

	DNSProviderAliDNS       DNSProvider = "alidns"       // 阿里云 DNS
	DNSProviderDNSPod       DNSProvider = "dnspod"       // DNSPod
	DNSProviderTencentCloud DNSProvider = "tencentcloud" // 腾讯云 DNS
	DNSProviderHuaweiCloud  DNSProvider = "huaweicloud"  // 华为云 DNS

	LogAccess  LogType = "access"  // 访问日志
	LogWAF     LogType = "waf"     // WAF 审计日志
	LogProcess LogType = "process" // 通用网站进程日志
	LogServer  LogType = "server"  // HTTP 服务运行日志
)
