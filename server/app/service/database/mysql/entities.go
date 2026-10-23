package mysql

// CharsetOptions 返回服务器支持的字符集、排序规则及当前默认值。
type CharsetOptions struct {
	Charset   string    `json:"charset"`
	Collation string    `json:"collation"`
	Charsets  []Charset `json:"charsets"`
}

// Charset 将排序规则按所属字符集分组。
type Charset struct {
	Name             string   `json:"name"`
	DefaultCollation string   `json:"default_collation"`
	Collations       []string `json:"collations"`
}

// User 返回账户的直接授权；角色继承权限独立标识，避免把对象授权误认为数据库管理权限。
type User struct {
	Name             string            `json:"name"`
	Host             string            `json:"host"`
	Protected        bool              `json:"protected"`
	Current          bool              `json:"current"`           // 当前连接实际匹配的账户，包含来源 Host。
	PasswordEditable bool              `json:"password_editable"` // 管理员可改密，内部系统账户除外。
	Roles            string            `json:"roles"`
	GlobalPrivileges []string          `json:"global_privileges"`
	Grants           []UserGrant       `json:"grants"`
	GrantedRoles     []string          `json:"granted_roles"`
	Restrictions     []UserRestriction `json:"restrictions"`
}

// UserGrant 描述数据库或对象授权，Pattern 表示库名仍包含授权通配符。
type UserGrant struct {
	Database   string   `json:"database"`
	Scope      string   `json:"scope"`
	Object     string   `json:"object,omitempty"`
	Privileges []string `json:"privileges"`
	Pattern    bool     `json:"pattern"`
}

// UserRestriction 是 MySQL 部分撤权记录，仅限制该库的全局权限，不撤销单独授予的库或对象权限。
type UserRestriction struct {
	Database   string   `json:"database"`
	Privileges []string `json:"privileges"`
}

// DatabasePrivileges 表示某个数据库最终保留的直接整库权限，空数组清空可管理权限。
type DatabasePrivileges struct {
	Database   string   `json:"database"`
	Privileges []string `json:"privileges"`
}

// UserPermissions 返回当前版本可管理的权限；其他权限保持原样。
type UserPermissions struct {
	Privileges []string             `json:"privileges"`
	Grants     []DatabasePrivileges `json:"grants"`
	Retained   bool                 `json:"retained"`
}

// Result 查询结果只保留有限行数和字节，截断信息供界面禁用不完整数据的行编辑。
type Result struct {
	Columns        []string        `json:"columns"`
	Rows           [][]interface{} `json:"rows"`
	HasMore        bool            `json:"has_more"`
	Truncated      bool            `json:"truncated"`
	TruncatedCells [][2]int        `json:"truncated_cells,omitempty"`
	BinaryColumns  []string        `json:"binary_columns,omitempty"`
	PrimaryKeys    []string        `json:"primary_keys,omitempty"`
	KeyColumns     []string        `json:"key_columns"` // 主键或可安全定位单行的非空唯一键。
	Affected       int64           `json:"affected"`
	ElapsedMs      int64           `json:"elapsed_ms"`
	Page           int             `json:"page,omitempty"`
	PageSize       int             `json:"page_size,omitempty"`
	Total          *int64          `json:"total,omitempty"` // 表数据分页返回准确总数；普通 SQL 查询不统计。
}

// Config 定义 MySQL 连接参数及校验规则。
type Config struct {
	Host     string `json:"host" label:"主机地址" default:"127.0.0.1"`
	Port     int    `json:"port" label:"端口" default:"3306" min:"1" max:"65535"`
	User     string `json:"user" label:"用户名" aliases:"username"`
	Password string `json:"password" label:"密码" encrypt:"true"`
	Database string `json:"database" label:"默认数据库"`
	Socket   string `json:"socket" label:"Socket 路径"`
	TLS      string `json:"tls" label:"TLS 连接" default:"" options:":默认,false:关闭,true:验证证书,preferred:优先加密,skip-verify:跳过证书验证,require:仅加密,disable:禁用"`
	SSLCA    string `json:"ssl_ca" label:"CA 证书路径"`
	SSLCert  string `json:"ssl_cert" label:"客户端证书路径"`
	SSLKey   string `json:"ssl_key" label:"客户端私钥路径"`
}

// ColumnDefinition 描述建表和字段修改需要的属性；nil 默认值表示不设置默认值。
type ColumnDefinition struct {
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	Nullable          bool    `json:"nullable"`
	Default           *string `json:"default"`
	DefaultExpression bool    `json:"default_expression"`
	AutoIncrement     bool    `json:"auto_increment"`
	Comment           *string `json:"comment"` // nil 保留原备注，空字符串清空备注。
}

// IndexColumn 描述索引字段及排序方向。
type IndexColumn struct {
	Name string `json:"name"`
	Desc bool   `json:"desc"`
}

// IndexDefinition 描述建表时同时创建的普通或唯一索引。
type IndexDefinition struct {
	Name    string        `json:"name"`
	Columns []IndexColumn `json:"columns"`
	Unique  bool          `json:"unique"`
	Comment string        `json:"comment"`
}

// ForeignKeyDefinition 描述外键名称、字段映射和更新删除规则。
type ForeignKeyDefinition struct {
	Name             string   `json:"name"`
	Columns          []string `json:"columns"`
	ReferenceSchema  string   `json:"reference_schema"`
	ReferenceTable   string   `json:"reference_table"`
	ReferenceColumns []string `json:"reference_columns"`
	OnDelete         string   `json:"on_delete"`
	OnUpdate         string   `json:"on_update"`
}
