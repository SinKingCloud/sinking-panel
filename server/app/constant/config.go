package constant

const (
	BasePath = "." //基础目录

	TempPath = BasePath + "/temp" //缓存目录

	SitePath     = TempPath + "/site"         //站点服务运行目录
	SiteRootPath = BasePath + "/data/wwwroot" //网站根目录
	AcmePath     = BasePath + "/data/acme"    //ACME 账户和签发数据目录

	ServerPath       = TempPath + "/server"        //HTTP 服务运行目录
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
