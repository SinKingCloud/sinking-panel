package mongo

import (
	"fmt"
	"strconv"
	"strings"
)

func (m *Manager) performanceFields(version string, wiredTiger bool) []map[string]interface{} {
	fields := []map[string]interface{}{
		{"name": "maxIncomingConnections", "label": "最大连接数", "type": "number", "min": 5, "max": 1000000},
	}
	if wiredTiger {
		minimum, step := 0.25, 0.25
		// MongoDB 3.0/3.2 的 WiredTiger 缓存只接受整数 GB，3.4 起允许小数。
		if strings.HasPrefix(version, "3.0.") || strings.HasPrefix(version, "3.2.") {
			minimum, step = 1, 1
		}
		fields = append(fields, map[string]interface{}{"name": "cacheSizeGB", "label": "WiredTiger 缓存", "type": "number", "unit": "GB", "min": minimum, "max": 10000, "step": step, "tooltip": "数据库缓存上限，需为连接和系统预留内存。"})
	}
	return append(fields, []map[string]interface{}{
		{"name": "slowms", "label": "慢查询阈值", "type": "number", "unit": "毫秒", "min": 0, "max": 2147483647},
		{"name": "mode", "label": "性能分析", "type": "select", "options": []map[string]interface{}{{"label": "关闭", "value": "off"}, {"label": "慢查询", "value": "slowOp"}, {"label": "全部查询", "value": "all"}}, "tooltip": "性能分析会写入数据库，全部查询模式可能增加负载。"},
	}...)
}

func (m *Manager) defaultConfig(version string) string {
	majorText, _, _ := strings.Cut(version, ".")
	major, _ := strconv.Atoi(majorText)
	storage := ""
	if major >= 3 {
		cache := "0.25"
		if strings.HasPrefix(version, "3.0.") || strings.HasPrefix(version, "3.2.") {
			cache = "1"
		}
		// 3.0 默认仍为 MMAPv1；明确选择其支持的 WiredTiger，避免两种引擎参数混用。
		storage = "  engine: wiredTiger\n  wiredTiger:\n    engineConfig:\n      cacheSizeGB: " + cache + "\n"
	}
	return fmt.Sprintf("net:\n  bindIp: 127.0.0.1\n  port: 27017\n  maxIncomingConnections: 65536\nstorage:\n  dbPath: /data/db\n%ssecurity:\n  authorization: enabled\noperationProfiling:\n  mode: off\n  slowOpThresholdMs: 100\n", storage)
}
