package log_type

const (
	User = iota //用户操作
)

// Map 日志类型数据
func Map() map[int]string {
	return map[int]string{
		User: "用户操作",
	}
}
