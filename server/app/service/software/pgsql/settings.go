package pgsql

func (m *Manager) performanceFields() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "max_connections", "label": "最大连接数", "type": "number", "min": 1, "max": 262143},
		{"name": "shared_buffers", "label": "共享缓冲区", "type": "number", "unit": "MB", "min": 1, "max": 1048576},
		{"name": "work_mem", "label": "查询工作内存", "type": "number", "unit": "MB", "min": 1, "max": 1048576, "tooltip": "每次排序或哈希操作可用的内存，并发连接会累加使用。"},
		{"name": "maintenance_work_mem", "label": "维护工作内存", "type": "number", "unit": "MB", "min": 1, "max": 1048576},
		{"name": "effective_cache_size", "label": "预计可用缓存", "type": "number", "unit": "MB", "min": 1, "max": 1048576, "tooltip": "供查询规划器估算可用缓存，不直接分配内存。"},
		{"name": "log_min_duration_statement", "label": "慢查询阈值", "type": "number", "unit": "毫秒", "min": -1, "max": 2147483647, "tooltip": "-1 关闭，0 记录全部语句。"},
	}
}

func (m *Manager) defaultConfig(version int) string {
	// 10 之前该参数是布尔值，认证只支持 MD5；新版继续使用 SCRAM。
	encryption := "'scram-sha-256'"
	if version < 100000 {
		encryption = "on"
	}
	return "listen_addresses = '127.0.0.1'\nport = 5432\npassword_encryption = " + encryption + "\nmax_connections = 100\nshared_buffers = '128MB'\nwork_mem = '4MB'\nmaintenance_work_mem = '64MB'\neffective_cache_size = '4096MB'\nlog_min_duration_statement = 1000\nlogging_collector = off\n"
}
