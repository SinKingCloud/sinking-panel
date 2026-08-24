package constant

const (
	BasePath = "." //基础目录

	TempPath     = BasePath + "/temp"      //缓存目录
	SitePath     = BasePath + "/data/site" //站点服务数据目录
	SiteRootPath = BasePath + "/wwwroot"   //网站根目录

	DBPath = BasePath + "/config" //数据库文件目录
	DBFile = "server.db"          //数据库文件

	ConfPath = BasePath + "/config" //配置文件目录
	ConfFile = "application.yml"    //配置文件

	ServerMode = "server.mode"
	ServerHost = "server.host"
	ServerPort = "server.port"
)
