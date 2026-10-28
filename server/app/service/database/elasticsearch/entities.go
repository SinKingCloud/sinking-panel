package elasticsearch

// Config 定义 Elasticsearch HTTP REST 连接参数及校验规则。
type Config struct {
	Host     string `json:"host" label:"主机地址" default:"127.0.0.1"`
	Port     int    `json:"port" label:"端口" default:"9200" min:"1" max:"65535"`
	User     string `json:"user" label:"用户名"`
	Password string `json:"password" label:"密码" encrypt:"true"`
	APIKey   string `json:"api_key" label:"API Key" encrypt:"true"`
	TLS      string `json:"tls" label:"TLS 连接" default:"" options:":默认,false:关闭,true:验证证书,skip-verify:跳过证书验证"`
	SSLCA    string `json:"ssl_ca" label:"CA 证书路径"`
	SSLCert  string `json:"ssl_cert" label:"客户端证书路径"`
	SSLKey   string `json:"ssl_key" label:"客户端私钥路径"`
}

// Index 描述实际索引，不登记面板内部数据库记录。
type Index struct {
	Name          string `json:"name"`
	Health        string `json:"health"`
	Status        string `json:"status"`
	Documents     int64  `json:"documents"`
	SizeBytes     int64  `json:"size_bytes"`
	PrimaryShards int    `json:"primary_shards"`
	Replicas      int    `json:"replicas"`
}

// DocumentPage 保留 Elasticsearch 文档元信息和完整 _source。
type DocumentPage struct {
	Pagination    string                   `json:"pagination"`
	SearchID      string                   `json:"search_id"`
	NextCursor    string                   `json:"next_cursor"`
	Documents     []map[string]interface{} `json:"documents"`
	Total         int64                    `json:"total"`
	TotalRelation string                   `json:"total_relation"`
	Page          int                      `json:"page"`
	PageSize      int                      `json:"page_size"`
	HasMore       bool                     `json:"has_more"`
	ElapsedMS     int64                    `json:"elapsed_ms"`
	TimedOut      bool                     `json:"timed_out"`
}
