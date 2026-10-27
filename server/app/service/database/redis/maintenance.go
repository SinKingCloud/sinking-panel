package redis

import (
	"context"
	"strings"
	"time"
)

// Snapshot 提交后台 RDB 快照任务。
func (m *Manager) Snapshot(ctx context.Context) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	if err = m.client.BgSave(ctx).Err(); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"pending": true,
		"scope":   "server",
		"message": "已提交 RDB 快照任务",
	}, nil
}

// Rewrite 提交后台 AOF 重写任务。
func (m *Manager) Rewrite(ctx context.Context) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	if err = m.client.BgRewriteAOF(ctx).Err(); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"pending": true,
		"scope":   "server",
		"message": "已提交 AOF 重写任务",
	}, nil
}

// Flush 仅清空当前登记逻辑库。
func (m *Manager) Flush(ctx context.Context) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	err = m.client.FlushDBAsync(ctx).Err()
	pending := true
	if err != nil {
		text := strings.ToLower(err.Error())
		if strings.Contains(text, "syntax error") || strings.Contains(text, "wrong number of arguments") {
			pending = false
			err = m.client.FlushDB(ctx).Err()
		}
	}
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"pending":  pending,
		"scope":    "database",
		"database": m.database,
		"message":  "当前逻辑库已清空",
	}, nil
}
