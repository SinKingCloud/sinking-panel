package cert_type

const (
	Manual = iota // 手动证书
	ACME          // ACME证书
)

// Map 证书类型数据
func Map() map[int]string {
	return map[int]string{
		Manual: "手动证书",
		ACME:   "ACME证书",
	}
}
