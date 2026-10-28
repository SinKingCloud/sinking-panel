package type_module

const (
	Site     = "site"
	Task     = "task"
	Database = "database"
)

// Map 业务分类所属模块；脚本分类使用 script_module 定义的模块标识。
func Map() map[string]string {
	return map[string]string{
		Site:     "网站管理",
		Task:     "计划任务",
		Database: "数据管理",
	}
}
