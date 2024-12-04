package constant

const (
	CacheNameWithUserConfig = "User_"    //用户应用配置
	CacheTimeWithUserConfig = 600 * 1000 //用户应用配置储存时间

	CacheNameWithSysConfig = "Sys"      //系统应用配置
	CacheTimeWithSysConfig = 600 * 1000 //用户应用配置储存时间

	CacheNameWithCaptcha = "Captcha_" //验证码
	CacheTimeWithCaptcha = 600 * 1000 //验证码缓存时间

	CacheNameWithUser = "User_"      //网站登录用户信息缓存
	CacheTimeWithUser = 86400 * 1000 //默认网站登录用户信息缓存时间

	CacheNameWithUinCookie = "UinCookie_"    //挂机服务器信息
	CacheTimeWithUinCookie = 3600 * 3 * 1000 //挂机服务器信息储存时间

	CacheNameWithUinServer = "UinServer_" //挂机服务器信息
	CacheTimeWithUinServer = 600 * 1000   //挂机服务器信息储存时间

	CacheNameWithUinServerJob          = "UinServerJobNum_" //挂机代理并发
	CacheTimeWithCacheNameUinServerJob = 3600 * 1000        //挂机代理并发缓存时间
)
