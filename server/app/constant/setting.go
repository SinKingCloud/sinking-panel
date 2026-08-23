package constant

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

	SiteGroup       = "site"              //网站服务配置组
	SiteHTTPOptions = SiteGroup + ".http" //网站 HTTP 运行参数
)
