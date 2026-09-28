package pgsql

import (
	"fmt"
	"strings"

	"server/app/util/platform"
)

// timezoneConfig 补齐独立配置文件缺失的时区；显式配置优先，不使用固定偏移或进程缓存。
// 面板指定 config_file 后，initdb 写入 PGDATA/postgresql.conf 的时区不会被读取。
func (m *Manager) timezoneConfig(content string) (string, error) {
	configured := make(map[string]bool, 2)
	for _, line := range strings.Split(content, "\n") {
		parts := strings.FieldsFunc(strings.TrimSpace(line), func(r rune) bool {
			return r == '=' || r == ' ' || r == '\t'
		})
		if len(parts) > 1 {
			configured[strings.ToLower(parts[0])] = true
		}
	}
	if configured["timezone"] && configured["log_timezone"] {
		return content, nil
	}
	name, _, err := platform.TimezoneGet()
	if err != nil {
		return "", fmt.Errorf("读取服务器时区失败: %w", err)
	}
	name = strings.NewReplacer("\\", "\\\\", "'", "''").Replace(name)
	var defaults strings.Builder
	for _, key := range []string{"timezone", "log_timezone"} {
		if !configured[key] {
			fmt.Fprintf(&defaults, "%s = '%s'\n", key, name)
		}
	}
	return defaults.String() + content, nil
}

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
