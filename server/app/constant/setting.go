package constant

// SensitiveGroups 按配置组定义读取和修改策略。
// Read 表示读取时脱敏，Write 表示禁止通过通用接口修改。
var SensitiveGroups = map[string]struct {
	Read  bool
	Write bool
}{
	LoginGroup: {Read: true, Write: true},
	SshGroup:   {Read: true, Write: true},
	SiteGroup:  {Read: false, Write: true},
}

const (
	WebGroup = "web"               //网站配置组
	WebTitle = WebGroup + ".title" //网站标题
	WebName  = WebGroup + ".name"  //网站名称

	UiGroup     = "ui"                   //界面配置组
	UiLayout    = UiGroup + ".layout"    //界面布局
	UiWaterMark = UiGroup + ".watermark" //水印内容
	UiTheme     = UiGroup + ".theme"     //主题
	UiCompact   = UiGroup + ".compact"   //紧凑模式
	UiColor     = UiGroup + ".color"     //主题色
	UiRadius    = UiGroup + ".radius"    //主题圆角

	LoginGroup    = "login"                  //登录配置组
	LoginAccount  = LoginGroup + ".account"  //登录账号
	LoginPassword = LoginGroup + ".password" //登录密码
	LoginToken    = LoginGroup + ".token"    //登录token
	LoginExpire   = LoginGroup + ".expire"   //登录token过期时间

	SshGroup    = "ssh"                   //本地ssh配置组
	SshIP       = SshGroup + ".ip"        //本地ssh IP
	SshUser     = SshGroup + ".user"      //本地ssh 账户
	SshPort     = SshGroup + ".port"      //本地ssh 端口
	SshAuthType = SshGroup + ".auth_type" //本地ssh 验证方式
	SshPassword = SshGroup + ".password"  //本地ssh 密码
	SshName     = SshGroup + ".name"      //本地ssh 名称

	SiteGroup                    = "site"                                    //网站服务配置组
	SiteHTTPGroup                = SiteGroup + ".http"                       //网站 HTTP 配置前缀
	SiteHTTPEnabled              = SiteHTTPGroup + ".enabled"                //网站 HTTP 服务启用状态
	SiteHTTPListen               = SiteHTTPGroup + ".http_listen"            //HTTP 监听地址
	SiteHTTPSListen              = SiteHTTPGroup + ".https_listen"           //HTTPS 监听地址
	SiteHTTPProtocols            = SiteHTTPGroup + ".protocols"              //HTTP 协议
	SiteHTTPDefaultSite          = SiteHTTPGroup + ".default_site"           //默认网站
	SiteHTTPNotFoundPage         = SiteHTTPGroup + ".not_found_page"         //资源不存在页面
	SiteHTTPSiteNotFoundPage     = SiteHTTPGroup + ".site_not_found_page"    //网站不存在页面
	SiteHTTPSiteDisabledPage     = SiteHTTPGroup + ".site_disabled_page"     //网站停用页面
	SiteHTTPDataPath             = SiteHTTPGroup + ".data_path"              //HTTP 运行数据目录
	SiteHTTPCachePath            = SiteHTTPGroup + ".cache_path"             //HTTP 缓存目录
	SiteHTTPLogPath              = SiteHTTPGroup + ".log_path"               //HTTP 运行日志
	SiteHTTPWAFLogPath           = SiteHTTPGroup + ".waf_log_path"           //WAF 审计日志
	SiteHTTPConfigPath           = SiteHTTPGroup + ".config_path"            //HTTP 配置快照
	SiteHTTPLogLevel             = SiteHTTPGroup + ".log_level"              //HTTP 日志级别
	SiteHTTPTrustedProxies       = SiteHTTPGroup + ".trusted_proxies"        //可信代理地址
	SiteHTTPClientIPHeaders      = SiteHTTPGroup + ".client_ip_headers"      //客户端 IP 请求头
	SiteHTTPTrustedProxiesStrict = SiteHTTPGroup + ".trusted_proxies_strict" //可信代理严格模式
	SiteHTTPReadTimeout          = SiteHTTPGroup + ".read_timeout"           //请求读取超时
	SiteHTTPReadHeaderTimeout    = SiteHTTPGroup + ".read_header_timeout"    //请求头读取超时
	SiteHTTPWriteTimeout         = SiteHTTPGroup + ".write_timeout"          //响应写入超时
	SiteHTTPIdleTimeout          = SiteHTTPGroup + ".idle_timeout"           //连接空闲超时
	SiteHTTPGracePeriod          = SiteHTTPGroup + ".grace_period"           //优雅关闭等待时间
	SiteHTTPMaxHeaderBytes       = SiteHTTPGroup + ".max_header_bytes"       //请求头大小限制
	SiteHTTPChallengeHost        = SiteHTTPGroup + ".http_challenge_host"    //ACME 验证监听地址
	SiteHTTPChallengePort        = SiteHTTPGroup + ".http_challenge_port"    //ACME 验证监听端口
	SiteHTTPACMEEmail            = SiteHTTPGroup + ".acme_email"             //ACME 账户邮箱
)
