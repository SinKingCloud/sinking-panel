package task_exec_type

const (
	Script  = iota // 系统脚本
	Request        // HTTP请求
)

// Map 任务执行类型数据
func Map() map[int]string {
	return map[int]string{
		Script:  "系统脚本",
		Request: "HTTP请求",
	}
}
