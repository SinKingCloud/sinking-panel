package server_auth_type

const (
	Password = iota //密码验证
	Cert            //证书验证
)

// Map 服务器验证类型数据
func Map() map[int]string {
	return map[int]string{
		Password: "密码验证",
		Cert:     "证书验证",
	}
}
