package site_type

const (
	Static = iota // 静态网站
	Proxy         // 反向代理
	PHP           // PHP网站
)

// Map 网站类型数据
func Map() map[int]string {
	return map[int]string{
		Static: "静态网站",
		Proxy:  "反向代理",
		PHP:    "PHP网站",
	}
}
