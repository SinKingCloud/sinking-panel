package task_type

const (
	Script  = iota //脚本
	Request        //HTTP请求
)

// Map 任务类型数据
func Map() map[int]string {
	return map[int]string{
		Script:  "系统脚本",
		Request: "HTTP请求",
	}
}
