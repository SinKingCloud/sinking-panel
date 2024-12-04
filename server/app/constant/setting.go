package constant

// 系统设置
const (
	WebGroup    = "web"                  //网站配置
	WebUrl      = WebGroup + ".url"      //网站名称
	WebName     = WebGroup + ".name"     //网站名称
	WebTitle    = WebGroup + ".title"    //网站标题
	WebKeyWords = WebGroup + ".keywords" //网站关键词
	WebDescribe = WebGroup + ".describe" //网站描述
	WebContact  = WebGroup + ".contact"  //网站联系人

	AliGroup       = "ali"                      //阿里云配置
	AliKey         = AliGroup + ".key"          //阿里云key
	AliSecret      = AliGroup + ".secret"       //阿里云secret
	AliSmsSign     = AliGroup + ".sms.sign"     //阿里云sms sign
	AliSmsTemplate = AliGroup + ".sms.template" //阿里云sms template
	AliSmsVar      = AliGroup + ".sms.var"      //阿里云sms变量名称

	UinGroup        = "uin"                       //挂机设置
	UinMonthPrice   = UinGroup + ".price.month"   //月度挂机价格
	UinQuarterPrice = UinGroup + ".price.quarter" //季度挂机价格
	UinHalfPrice    = UinGroup + ".price.half"    //半年挂机价格
	UinYearPrice    = UinGroup + ".price.year"    //年度挂机价格
	UinForeverPrice = UinGroup + ".price.forever" //永久挂机价格

	ReFoundGroup = "refund"                //退单设置
	RefundOpen   = ReFoundGroup + ".open"  //退单功能是否开启
	RefundPoint  = ReFoundGroup + ".point" //退单功能手续费(百分比)
	RefundMin    = ReFoundGroup + ".min"   //退单功能手续费(单笔最低)
	RefundMax    = ReFoundGroup + ".max"   //退单功能手续费(单笔最高)
	RefundDay    = ReFoundGroup + ".day"   //退单功能最久退单日期
	RefundFree   = ReFoundGroup + ".free"  //退单功能免费退单日期

	PayGroup = "pay"             //支付设置
	PayMin   = PayGroup + ".min" //最小充值金额
	PayMax   = PayGroup + ".max" //最大充值金额
	PayAli   = PayGroup + ".ali" //是否开启支付宝支付
	PayWx    = PayGroup + ".wx"  //是否开启微信支付
	PayQQ    = PayGroup + ".qq"  //是否开启QQ支付
	PayPid   = PayGroup + ".pid" //易支付PID
	PayKey   = PayGroup + ".key" //易支付KEY
	PayUrl   = PayGroup + ".url" //易支付URL

	Proto            = "proto"
	ProtoQrCode      = Proto + ".qrcode"       //二维码登录使用协议
	ProtoPwdLogin    = Proto + ".pwd"          //密码登录使用协议
	ProtoSignAndroid = Proto + ".sign.android" //安卓sign
	ProtoSignPc      = Proto + ".sign.pc"      //安卓sign
	ProtoAuthLevel   = Proto + ".auth.level"   //实名等级
)

// 用户设置
const (
	UserUinGroup     = "uin"                    //用户挂机设置
	UserUinPriceOpen = UinGroup + ".price.open" //用户自定义价格(0:关闭/1:开启)
)
