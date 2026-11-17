package site_backup_status

const (
	Disabled = iota // 关闭自动备份
	Enabled         // 开启自动备份
)

// Map 网站自动备份状态。
func Map() map[int]string {
	return map[int]string{
		Disabled: "关闭",
		Enabled:  "开启",
	}
}
