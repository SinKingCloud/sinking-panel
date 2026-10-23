package sqlite

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
	KeyColumns     []string        `json:"key_columns"` // 用于定位单行的主键或非空唯一键
	Affected       int64           `json:"affected"`
	ElapsedMs      int64           `json:"elapsed_ms"`
	Page           int             `json:"page,omitempty"`
	PageSize       int             `json:"page_size,omitempty"`
}

// Config 定义 Sqlite 文件连接参数；连接测试不会创建文件。
type Config struct {
	Path     string `json:"path" label:"数据库文件" required:"true"`
	ReadOnly bool   `json:"read_only" label:"只读连接" default:"false"`
	Database string `json:"database" label:"默认数据库" default:"" options:":默认,main:main"`
}

// IndexDefinition 配置建表时同时创建的普通或唯一索引。
type IndexDefinition struct {
	Name    string        `json:"name"`
	Columns []IndexColumn `json:"columns"`
	Unique  bool          `json:"unique"`
}
