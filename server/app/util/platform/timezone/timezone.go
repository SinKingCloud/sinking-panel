// Package timezone 通过系统原生接口读取和持久化系统时区，不执行外部命令。
package timezone

import (
	"errors"
	"strings"
)

// Set 设置系统时区，需要系统授予修改权限，不调整时间戳或覆盖应用的 time.Local。
// 已运行的程序可能缓存旧时区，容器需要重新加载其时区挂载。
func Set(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || strings.ContainsRune(name, 0) {
		return errors.New("时区名称不能为空或包含空字符")
	}
	return setTimezone(name)
}
