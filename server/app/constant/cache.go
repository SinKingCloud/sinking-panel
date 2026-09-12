package constant

import "time"

const (
	CacheNameWithSysConfig = "Sys"             //系统配置
	CacheTimeWithSysConfig = 600 * time.Second //系统配置配置储存时间

	CacheNameWithCaptcha = "Captcha_"        //验证码
	CacheTimeWithCaptcha = 600 * time.Second //验证码缓存时间

	CacheNameWithTypeEnum = "TypeEnum_"       //类型枚举
	CacheTimeWithTypeEnum = 600 * time.Second //类型枚举储存时间

	CacheNameWithSiteNameEnum = "SiteNameEnum"    //可选默认站点枚举
	CacheTimeWithSiteNameEnum = 600 * time.Second //可选默认站点枚举储存时间

	CacheNameWithSecretNameEnum = "SecretNameEnum"  //密钥名称枚举
	CacheTimeWithSecretNameEnum = 600 * time.Second //密钥名称枚举储存时间

	CacheNameWithFilePreview            = "FilePreview_"     //文件预览签名
	CacheTimeWithFilePreview            = 3600 * time.Second //文件预览签名储存时间
	CacheTimeWithFilePreviewRenewBefore = 600 * time.Second  //文件预览签名续期阈值
	CacheTimeWithFilePreviewRenew       = 1800 * time.Second //文件预览签名续期时间
)
