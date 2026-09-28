package platform

import (
	"os"
	"runtime"
)

// CanRunContainer 判断当前系统和运行身份是否允许初始化轻量容器。
// OverlayFS、namespace 和安全策略由实际初始化及启动操作校验，避免预检误拒绝可用环境。
func CanRunContainer() bool {
	return runtime.GOOS == "linux" && os.Geteuid() == 0
}
