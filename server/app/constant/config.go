package constant

const (
	BasePath      = "."                   //基础目录
	RuntimePath   = BasePath + "/runtime" //运行时数据目录
	DataPath      = BasePath + "/data"    //面板数据目录
	PanelDataPath = DataPath + "/server"  //面板服务配置和数据库目录

	SiteRootPath   = DataPath + "/site"         //网站数据根目录
	BackupPath     = DataPath + "/backup"       //备份数据目录
	SiteBackupPath = DataPath + "/site_backups" //网站独立备份目录
	RecyclePath    = DataPath + "/recycle"      //回收站数据目录
	AcmePath       = DataPath + "/acme"         //ACME 账户和签发数据目录
	ContainerPath  = DataPath + "/containers"   //轻量容器数据目录
	SoftwarePath   = DataPath + "/software"     //软件包目录

	MonitorPath = PanelDataPath + "/monitor"         //系统监控历史数据目录
	DBFile      = PanelDataPath + "/server.db"       //数据库文件
	ConfFile    = PanelDataPath + "/application.yml" //可选配置文件，不自动创建

	CmdPath               = RuntimePath + "/cmd"               //脚本执行临时目录
	CronPath              = RuntimePath + "/cron"              //计划任务日志目录
	TaskPath              = RuntimePath + "/task"              //系统任务日志目录
	UploadPath            = RuntimePath + "/upload"            //上传临时文件目录
	IPPath                = RuntimePath + "/ip"                //IP 数据库目录
	SitePath              = RuntimePath + "/site"              //站点服务运行目录
	DaemonPath            = RuntimePath + "/daemon"            //面板进程本机通信目录
	PluginCommandPath     = RuntimePath + "/plugin"            //插件命令本机通信目录
	CaddyServerDataPath   = RuntimePath + "/caddy/data"        //HTTP 服务数据目录
	CaddyServerCachePath  = RuntimePath + "/caddy/cache"       //HTTP 服务缓存目录
	CaddyServerLogFile    = RuntimePath + "/caddy/server.log"  //HTTP 服务运行日志
	CaddyServerWAFLogFile = RuntimePath + "/caddy/waf.log"     //HTTP 服务 WAF 日志
	CaddyServerConfigFile = RuntimePath + "/caddy/config.json" //HTTP 服务配置快照

	ServerMode = "server.mode"
	ServerHost = "server.host"
	ServerPort = "server.port"
)
