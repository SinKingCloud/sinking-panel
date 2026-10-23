package sqlite

import (
	"context"
	"time"
)

// DatabaseInfo 读取已配置文件的 main 数据库结构和页统计，不创建或附加文件。
func (m *Manager) DatabaseInfo(ctx context.Context) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	conn, err := m.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	result := map[string]interface{}{"name": "main", "exists": true}
	for key, query := range map[string]string{
		"page_size":  "PRAGMA main.page_size",
		"page_count": "PRAGMA main.page_count",
		"free_pages": "PRAGMA main.freelist_count",
		"tables":     "SELECT COUNT(*) FROM main.sqlite_master WHERE type = 'table' AND name NOT GLOB 'sqlite_*'",
	} {
		var value int64
		if err = conn.QueryRowContext(ctx, query).Scan(&value); err != nil {
			return nil, err
		}
		result[key] = value
	}
	for key, query := range map[string]string{
		"encoding":     "PRAGMA main.encoding",
		"journal_mode": "PRAGMA main.journal_mode",
	} {
		var value string
		if err = conn.QueryRowContext(ctx, query).Scan(&value); err != nil {
			return nil, err
		}
		result[key] = value
	}
	result["size"] = result["page_size"].(int64) * result["page_count"].(int64)
	return result, nil
}
