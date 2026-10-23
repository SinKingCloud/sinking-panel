package mongo

// Config 使用主机、端口、用户名和密码连接，其他参数为可选项。
type Config struct {
	Host             string `json:"host" label:"主机地址" default:"127.0.0.1"`
	Port             int    `json:"port" label:"端口" default:"27017" min:"1" max:"65535"`
	User             string `json:"user" label:"用户名" aliases:"username"`
	Password         string `json:"password" label:"密码" encrypt:"true"`
	AuthSource       string `json:"auth_source" label:"认证数据库" aliases:"authSource"`
	Database         string `json:"database" label:"默认数据库"`
	ReplicaSet       string `json:"replica_set" label:"副本集名称"`
	DirectConnection bool   `json:"direct_connection" label:"直连节点" default:"false"`
	TLS              string `json:"tls" label:"TLS 连接" default:"" options:":默认,false:关闭,true:验证证书,skip-verify:跳过证书验证,disable:禁用"`
	SSLCA            string `json:"ssl_ca" label:"CA 证书路径"`
	SSLCert          string `json:"ssl_cert" label:"客户端证书路径"`
	SSLKey           string `json:"ssl_key" label:"客户端私钥路径"`
}

// DatabasePrivileges 指定账户在一个数据库上的常用角色。
type DatabasePrivileges struct {
	Database   string   `json:"database"`
	Privileges []string `json:"privileges"`
}

type accountRole struct {
	Role string `bson:"role"`
	DB   string `bson:"db"`
}
type account struct {
	User           string        `bson:"user"`
	DB             string        `bson:"db"`
	Roles          []accountRole `bson:"roles"`
	InheritedRoles []accountRole `bson:"inheritedRoles"`
}
