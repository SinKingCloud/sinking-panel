package constant

const (
	BasePath = "." //基础目录

	RuntimePath = BasePath + "/runtime"    //运行数据目录
	CmdPath     = RuntimePath + "/cmd"     //脚本执行临时目录
	CronPath    = RuntimePath + "/cron"    //计划任务日志目录
	TaskPath    = RuntimePath + "/task"    //系统任务日志目录
	RecyclePath = RuntimePath + "/recycle" //回收站默认目录
	UploadPath  = RuntimePath + "/upload"  //上传临时文件目录
	IPPath      = RuntimePath + "/ip"      //IP 数据库目录
	SitePath    = RuntimePath + "/site"    //站点服务运行目录
	ServerPath  = RuntimePath + "/server"  //HTTP 服务运行目录

	SiteRootPath  = BasePath + "/data/site"       //网站数据根目录
	AcmePath      = BasePath + "/data/acme"       //ACME 账户和签发数据目录
	ContainerPath = BasePath + "/data/containers" //轻量容器数据目录

	ServerDataPath   = ServerPath + "/data"        //HTTP 服务数据目录
	ServerCachePath  = ServerPath + "/cache"       //HTTP 服务缓存目录
	ServerLogPath    = ServerPath + "/server.log"  //HTTP 服务运行日志
	ServerWAFLogPath = ServerPath + "/waf.log"     //HTTP 服务 WAF 日志
	ServerConfigPath = ServerPath + "/config.json" //HTTP 服务配置快照

	DBPath = BasePath + "/config" //数据库文件目录
	DBFile = "server.db"          //数据库文件

	ConfPath = BasePath + "/config" //配置文件目录
	ConfFile = "application.yml"    //配置文件

	ServerMode = "server.mode"
	ServerHost = "server.host"
	ServerPort = "server.port"
)
