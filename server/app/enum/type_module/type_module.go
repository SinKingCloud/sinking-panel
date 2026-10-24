package type_module

const (
	Ssh      = "ssh"
	Site     = "site"
	Task     = "task"
	Database = "database"
)

// Map 类型所属模块数据。
func Map() map[string]string {
	return map[string]string{
		Ssh:      "终端管理",
		Site:     "网站管理",
		Task:     "计划任务",
		Database: "数据管理",
	}
}
