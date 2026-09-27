package mysql

// performanceFields 的内存项以 MiB 展示，写入配置时转换成字节。
func (m *Manager) performanceFields() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "max_connections", "label": "最大连接数", "type": "number", "min": 1, "max": 100000},
		{"name": "innodb_buffer_pool_size", "label": "InnoDB 缓冲池", "type": "number", "unit": "MB", "min": 5, "max": 1048576, "tooltip": "数据库缓存使用的内存，需为系统和其他服务预留空间。"},
		{"name": "wait_timeout", "label": "空闲连接超时", "type": "number", "unit": "秒", "min": 1, "max": 31536000},
		{"name": "max_allowed_packet", "label": "最大数据包", "type": "number", "unit": "MB", "min": 1, "max": 1024},
		{"name": "slow_query_log", "label": "慢查询日志", "type": "switch"},
		{"name": "long_query_time", "label": "慢查询阈值", "type": "number", "unit": "秒", "min": 0, "max": 31536000},
	}
}

func (m *Manager) defaultConfig() string {
	return "[mysqld]\nuser=mysql\ndatadir=/var/lib/mysql\nsocket=/var/run/mysqld/mysqld.sock\nbind-address=127.0.0.1\nport=3306\nloose-mysqlx=0\nmax_connections=151\ninnodb_buffer_pool_size=128M\nwait_timeout=28800\nmax_allowed_packet=64M\nslow_query_log=0\nlong_query_time=2\n"
}
