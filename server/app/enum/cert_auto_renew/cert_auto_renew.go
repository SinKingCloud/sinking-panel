package cert_auto_renew

const (
	Disabled = iota // 关闭
	Enabled         // 开启
)

// Map 证书自动续签状态数据。
func Map() map[int]string {
	return map[int]string{
		Disabled: "关闭",
		Enabled:  "开启",
	}
}
