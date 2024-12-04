package constant

const (
	LockConfigSet     = "ConfigSet_"  //修改配置并发锁
	LockTimeConfigSet = 1 * 60 * 1000 //修改配置并发锁最大时间

	LockWebUserReg     = "WebUserReg_" //注册用户并发锁
	LockTimeWebUserReg = 1 * 60 * 1000 //注册用户并发锁最大时间

	LockWebUserBuy     = "WebUserBuy_" //用户购买并发锁
	LockTimeWebUserBuy = 1 * 60 * 1000 //用户购买并发锁最大时间

	LockUinInfo     = "LockUinInfo_" //QQ更新详情并发锁
	LockTimeUinInfo = 1 * 60 * 1000  //QQ更新详情并发锁最大时间

	LockUserUinJob     = "LockUserUinJob_" //用户挂机任务锁
	LockTimeUserUinJob = 1 * 60 * 1000     //用户挂机任务锁等待时间

	LockUinJob     = "LockUinJob_"  //挂机任务锁
	LockTimeUinJob = 10 * 60 * 1000 //挂机任务锁最大时间

	LockUinCheckJob     = "LockUinCheckJob_" //挂机检测任务锁
	LockTimeUinCheckJob = 10 * 60 * 1000     //挂机检测任务锁最大时间

	LockUinCookie     = "LockUinCookie_" //cookie更新并发锁
	LockTimeUinCookie = 1 * 60 * 1000    //cookie更新并发锁最大时间

	LockUinRefund     = "LockUinRefund_" //挂机应用退单并发锁
	LockTimeUinRefund = 1 * 60 * 1000    //挂机应用退单并发锁最大时间

	LockChangeOrderStatus     = "Order_"      //订单状态更改并发锁
	LockTimeChangeOrderStatus = 1 * 60 * 1000 //订单状态更改并发锁最大时间
)
