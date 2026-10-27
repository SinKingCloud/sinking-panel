package software_status

const (
	Installed = 2 // 已安装
	Removing  = 4 // 卸载中
)

// Map 软件安装状态。
func Map() map[int]string {
	return map[int]string{
		Installed: "已安装",
		Removing:  "卸载中",
	}
}
