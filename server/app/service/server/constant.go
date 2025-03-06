package server

type AuthType int

const (
	Password AuthType = iota
	Cert
)

// AuthTypes 类型数据
func (s *Service) AuthTypes() map[AuthType]string {
	return map[AuthType]string{
		Password: "密码验证",
		Cert:     "证书验证",
	}
}
