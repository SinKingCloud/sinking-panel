package script_module

const (
	Ssh      = "script_ssh"
	Task     = "script_task"
	Database = "script_database"
)

// Map 常用脚本及其分类共用的模块标识，实例范围在模块后追加 _实例ID。
func Map() map[string]string {
	return map[string]string{
		Ssh:      "终端脚本",
		Task:     "计划任务脚本",
		Database: "数据库脚本",
	}
}
