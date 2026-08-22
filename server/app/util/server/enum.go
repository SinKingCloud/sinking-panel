package server

// Module 表示 HTTP 处理器模块名称。
type Module string

// ProxyTransport 表示反向代理传输实现。
type ProxyTransport string

// ProxyScheme 表示 HTTP 上游协议。
type ProxyScheme string

// WAFMode 表示 WAF 运行模式。
type WAFMode string

// CertificateCA 表示证书签发环境。
type CertificateCA string

const (
	ProxyTransportHTTP    ProxyTransport = "http"    // HTTP 反向代理传输
	ProxyTransportFastCGI ProxyTransport = "fastcgi" // FastCGI 反向代理传输
	ProxySchemeHTTP       ProxyScheme    = "http"    // HTTP 上游协议
	ProxySchemeHTTPS      ProxyScheme    = "https"   // HTTPS 上游协议

	HandlerWAF            Module = "waf"             // WAF 处理器
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

	CertificateCAProd    CertificateCA = "production" // Let's Encrypt 正式环境
	CertificateCAStaging CertificateCA = "staging"    // Let's Encrypt 测试环境
)
