package type_module

const (
	Script = "script"
	Site   = "site"
	Task   = "task"
)

// Map 类型所属模块数据
func Map() map[string]string {
	return map[string]string{
		Script: "常用脚本",
		Site:   "网站管理",
		Task:   "计划任务",
	}
}
