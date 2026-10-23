package sqlserver

import (
	"context"
	"net"
	"sync"

	mssql "github.com/microsoft/go-mssqldb"
)

type connectionConnector struct {
	*mssql.Connector
	mutex       sync.Mutex
	connections []net.Conn
}

type connectionDialer func(context.Context, string, string) (net.Conn, error)

// ColumnDefinition 定义建表和字段修改参数；nil 默认值表示没有默认值。
type ColumnDefinition struct {
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	Nullable          bool    `json:"nullable"`
	Default           *string `json:"default"`
	DefaultExpression bool    `json:"default_expression"`
	AutoIncrement     bool    `json:"auto_increment"`
}

// IndexColumn 定义索引字段及排序方向。
type IndexColumn struct {
	Name string `json:"name"`
	Desc bool   `json:"desc"`
}

// ForeignKeyDefinition 定义本表字段与引用表字段的对应关系。
type ForeignKeyDefinition struct {
	Name             string   `json:"name"`
	Columns          []string `json:"columns"`
	ReferenceSchema  string   `json:"reference_schema"`
	ReferenceTable   string   `json:"reference_table"`
	ReferenceColumns []string `json:"reference_columns"`
	OnDelete         string   `json:"on_delete"`
	OnUpdate         string   `json:"on_update"`
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

// Config 定义 SqlServer 连接参数及校验规则。
type Config struct {
	Host                   string `json:"host" label:"主机地址" default:"127.0.0.1"`
	Port                   int    `json:"port" label:"端口" default:"1433" min:"1" max:"65535"`
	User                   string `json:"user" label:"用户名" aliases:"username"`
	Password               string `json:"password" label:"密码" encrypt:"true"`
	Database               string `json:"database" label:"默认数据库"`
	Encrypt                string `json:"encrypt" label:"连接加密" default:"" options:":默认,true:开启,false:按服务端协商,strict:严格加密,disable:关闭"`
	TrustServerCertificate string `json:"trust_server_certificate" label:"证书校验" default:"" options:":默认,false:验证服务器证书,true:信任服务器证书"`
}

// IndexDefinition 配置建表时同时创建的普通或唯一索引。
type IndexDefinition struct {
	Name    string        `json:"name"`
	Columns []IndexColumn `json:"columns"`
	Unique  bool          `json:"unique"`
}
