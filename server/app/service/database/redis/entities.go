package redis

// Config 定义 Redis 连接参数；可选用户名用于 Redis ACL 认证。
type Config struct {
	Host     string `json:"host" label:"主机地址" default:"127.0.0.1"`
	Port     int    `json:"port" label:"端口" default:"6379" min:"1" max:"65535"`
	User     string `json:"user" label:"用户名" aliases:"username"`
	Password string `json:"password" label:"密码" encrypt:"true"`
	DB       int    `json:"db" label:"默认数据库" default:"0" min:"0" max:"2147483647"`
	Socket   string `json:"socket" label:"Socket 路径"`
	TLS      string `json:"tls" label:"TLS 连接" default:"" options:":默认,false:关闭,true:验证证书,skip-verify:跳过证书验证,disable:禁用"`
	SSLCA    string `json:"ssl_ca" label:"CA 证书路径"`
	SSLCert  string `json:"ssl_cert" label:"客户端证书路径"`
	SSLKey   string `json:"ssl_key" label:"客户端私钥路径"`
}
