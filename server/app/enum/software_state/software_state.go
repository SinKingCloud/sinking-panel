package software_state

const (
	Installing = 1 // 安装中
	Installed  = 2 // 已安装
	Upgrading  = 3 // 升级中
	Removing   = 4 // 卸载中
)

// Map 软件安装状态，与启用状态独立。
func Map() map[int]string {
	return map[int]string{
		Installing: "安装中",
		Installed:  "已安装",
		Removing:   "卸载中",
		Upgrading:  "升级中",
	}
}
