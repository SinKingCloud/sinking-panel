package constant

const (
	BasePath = "." //基础目录

	RuntimePath       = BasePath + "/runtime"   //运行数据目录
	CmdPath           = RuntimePath + "/cmd"    //脚本执行临时目录
	CronPath          = RuntimePath + "/cron"   //计划任务日志目录
	TaskPath          = RuntimePath + "/task"   //系统任务日志目录
	UploadPath        = RuntimePath + "/upload" //上传临时文件目录
	IPPath            = RuntimePath + "/ip"     //IP 数据库目录
	SitePath          = RuntimePath + "/site"   //站点服务运行目录
	ServerPath        = RuntimePath + "/server" //HTTP 服务运行目录
	DaemonPath        = RuntimePath + "/daemon" //面板进程本机通信目录
	PluginCommandPath = RuntimePath + "/plugin" //插件命令本机通信目录

	SiteRootPath   = BasePath + "/data/site"         //网站数据根目录
	BackupPath     = BasePath + "/data/backup"       //备份数据目录
	SiteBackupPath = BasePath + "/data/site_backups" //网站独立备份目录
	RecyclePath    = BasePath + "/data/recycle"      //回收站数据目录
	AcmePath       = BasePath + "/data/acme"         //ACME 账户和签发数据目录
	ContainerPath  = BasePath + "/data/containers"   //轻量容器数据目录
	SoftwarePath   = BasePath + "/data/software"     //软件包目录
	PanelDataPath  = BasePath + "/data/server"       //面板配置和数据库目录

	ServerDataPath   = ServerPath + "/data"        //HTTP 服务数据目录
	ServerCachePath  = ServerPath + "/cache"       //HTTP 服务缓存目录
	ServerLogFile    = ServerPath + "/server.log"  //HTTP 服务运行日志
	ServerWAFLogFile = ServerPath + "/waf.log"     //HTTP 服务 WAF 日志
	ServerConfigFile = ServerPath + "/config.json" //HTTP 服务配置快照

	MonitorPath = PanelDataPath + "/monitor"         //系统监控历史数据目录
	DBFile      = PanelDataPath + "/server.db"       //数据库文件
	ConfFile    = PanelDataPath + "/application.yml" //可选配置文件，不自动创建

	ServerMode = "server.mode"
	ServerHost = "server.host"
	ServerPort = "server.port"
)
