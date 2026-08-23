package site_status

const (
	Enabled  = iota // 已启用
	Disabled        // 已停用
)

// Map 网站状态数据
func Map() map[int]string {
	return map[int]string{
		Enabled:  "已启用",
		Disabled: "已停用",
	}
}
