package database_type

const (
	Redis         = iota + 1 // Redis
	MySQL                    // MySQL
	Postgres                 // Postgres
	MongoDB                  // MongoDB
	SqlServer                // SqlServer
	Sqlite                   // Sqlite
	Elasticsearch            // Elasticsearch
)

// Map 数据库类型数据。
func Map() map[int]string {
	return map[int]string{
		Redis:         "Redis",
		MySQL:         "MySQL",
		Postgres:      "Postgres",
		MongoDB:       "MongoDB",
		SqlServer:     "SqlServer",
		Sqlite:        "Sqlite",
		Elasticsearch: "Elasticsearch",
	}
}
