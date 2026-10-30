package software_status

const (
	Disabled = 0 // 禁用
	Enabled  = 1 // 启用
)

// Map 软件启用状态。
func Map() map[int]string {
	return map[int]string{
		Disabled: "禁用",
		Enabled:  "启用",
	}
}
