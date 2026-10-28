package plugin

import (
	"fmt"
	"strings"
)

const (
	MaxPackageFiles       = 1000
	MaxSourceSize   int64 = 64 << 20
	MaxImageSize    int64 = 16 << 30
)

// PackageImageEntry 分类随包镜像；JSON 清单仍占源码预算，不能绕过 64 MiB 限制。
// 调用方须先验证包内相对路径并拒绝符号链接和特殊文件。
func PackageImageEntry(name string, directory bool) (bool, error) {
	if name != "images" && !strings.HasPrefix(name, "images/") {
		return false, nil
	}
	if directory {
		return false, nil
	}
	if strings.HasPrefix(name, "images/") {
		for _, suffix := range []string{".tar", ".tar.gz", ".tgz"} {
			if strings.HasSuffix(name, suffix) {
				return true, nil
			}
		}
		if strings.HasSuffix(name, ".json") {
			return false, nil
		}
	}
	return false, fmt.Errorf("images 目录仅支持 .tar、.tar.gz、.tgz 镜像归档和 .json 清单: %s", name)
}
