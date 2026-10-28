package postgres

// DatabasePrivileges 配置一个数据库或架构的直接权限；Schema 为空表示数据库本身。
type DatabasePrivileges struct {
	Database   string   `json:"database"`
	Schema     string   `json:"schema"`
	Privileges []string `json:"privileges"`
}

// accountPermissionObject 保存对象的直接授权，避免把继承权限当作可撤销的直接权限。
type accountPermissionObject struct {
	Name       string
	Kind       string
	Privileges map[string]bool // 值为 true 表示可转授或由其他授权人授予，保持原样。
	Owned      bool            // 角色拥有的对象不重写。
	Protected  bool            // 含独立列权限的表不撤销表级权限，避免连带删除列授权。
}

// accountPermissionChange 记录本次实际执行的变动，用于提交前确认数据库没有仅警告而忽略授权。
type accountPermissionChange struct {
	Name      string
	Kind      string
	Privilege string
	Granted   bool
}

// Result 查询结果只保留有限行数和字节，截断信息供界面禁用不完整数据的行编辑。
type Result struct {
	Results        []*Result       `json:"results,omitempty"`
	Total          int64           `json:"total"`
	Columns        []string        `json:"columns"`
	Rows           [][]interface{} `json:"rows"`
	HasMore        bool            `json:"has_more"`
	Truncated      bool            `json:"truncated"`
	TruncatedCells [][2]int        `json:"truncated_cells,omitempty"`
	BinaryColumns  []string        `json:"binary_columns,omitempty"`
	PrimaryKeys    []string        `json:"primary_keys,omitempty"`
	KeyColumns     []string        `json:"key_columns"` // 编辑时使用的主键或非空唯一键
	Affected       int64           `json:"affected"`
	ElapsedMs      int64           `json:"elapsed_ms"`
	Page           int             `json:"page,omitempty"`
	PageSize       int             `json:"page_size,omitempty"`
}

// Config 定义 Postgres 连接参数及校验规则。
type Config struct {
	Host        string `json:"host" label:"主机地址" default:"127.0.0.1"`
	Port        int    `json:"port" label:"端口" default:"5432" min:"1" max:"65535"`
	User        string `json:"user" label:"用户名"`
	Password    string `json:"password" label:"密码" encrypt:"true"`
	Database    string `json:"database" label:"默认数据库"`
	SSLMode     string `json:"sslmode" label:"TLS 连接" default:"" options:":默认,disable:关闭,allow:允许,prefer:优先加密,require:要求加密,verify-ca:验证 CA,verify-full:完整验证"`
	SSLRootCert string `json:"sslrootcert" label:"CA 证书路径"`
	SSLCert     string `json:"sslcert" label:"客户端证书路径"`
	SSLKey      string `json:"sslkey" label:"客户端私钥路径"`
}

// ColumnDefinition 描述建表和字段修改需要的属性；nil 默认值表示不设置默认值。
type ColumnDefinition struct {
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	Nullable          bool    `json:"nullable"`
	Default           *string `json:"default"`
	DefaultExpression bool    `json:"default_expression"`
	AutoIncrement     bool    `json:"auto_increment"`
}

// IndexColumn 描述索引字段及排序方向。
type IndexColumn struct {
	Name string `json:"name"`
	Desc bool   `json:"desc"`
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

// copyInput 用有界读取函数适配 pgx COPY，不需要缓存整个数据段或额外投递 goroutine。
type copyInput func([]byte) (int, error)

func (read copyInput) Read(data []byte) (int, error) { return read(data) }

// copyOutput 用流式写入函数适配 pgx COPY，可在发布备份前检查单行大小。
type copyOutput func([]byte) (int, error)

func (write copyOutput) Write(data []byte) (int, error) { return write(data) }

// IndexDefinition 配置建表时同时创建的普通或唯一索引。
type IndexDefinition struct {
	Name    string        `json:"name"`
	Columns []IndexColumn `json:"columns"`
	Unique  bool          `json:"unique"`
}
