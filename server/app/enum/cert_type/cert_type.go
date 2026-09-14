package cert_type

const (
	Import = iota // 导入证书
	Manual        // 手动申请
	Auto          // 自动申请
)

// Map 证书类型数据
func Map() map[int]string {
	return map[int]string{
		Import: "导入证书",
		Manual: "手动申请",
		Auto:   "自动申请",
	}
}
