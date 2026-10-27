package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

// maintenanceTable 验证所选用户表，拒绝系统对象和视图。
func (m *Manager) maintenanceTable(ctx context.Context, schemaName, tableName string) (string, error) {
	table, err := m.target(schemaName, tableName)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(strings.ToLower(tableName), "sqlite_") {
		return "", errors.New("不能维护 Sqlite 系统表")
	}
	var objectType string
	err = m.db.QueryRowContext(ctx, "SELECT type FROM main.sqlite_master WHERE name = ? COLLATE NOCASE AND type IN ('table', 'view')", tableName).Scan(&objectType)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("数据表不存在")
	}
	if err != nil {
		return "", err
	}
	if objectType != "table" {
		return "", errors.New("视图不支持表维护操作")
	}

	return table, nil
}

// Check 检查所选表的一致性。
func (m *Manager) Check(ctx context.Context, schemaName, tableName string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	started := time.Now()
	_, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	name, _ := m.quote(tableName)
	rows, err := m.db.QueryContext(ctx, "PRAGMA main.integrity_check("+name+")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result, err := m.read(rows, 1000)
	if err != nil {
		return nil, err
	}
	result.ElapsedMs = time.Since(started).Milliseconds()
	return result, nil
}

// Analyze 更新所选表的统计信息。
func (m *Manager) Analyze(ctx context.Context, schemaName, tableName string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	started := time.Now()
	table, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if _, err = m.db.ExecContext(ctx, "ANALYZE "+table); err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}

// Optimize 使用 VACUUM 整理整个 Sqlite 文件。
func (m *Manager) Optimize(ctx context.Context, schemaName, tableName string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	started := time.Now()
	_, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if _, err = m.db.ExecContext(ctx, "VACUUM main"); err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}

// Truncate 清空所选表并保留结构。
func (m *Manager) Truncate(ctx context.Context, schemaName, tableName string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	started := time.Now()
	table, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	executed, err := m.db.ExecContext(ctx, "DELETE FROM "+table)
	if err != nil {
		return nil, err
	}
	affected, err := executed.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		Affected:  affected,
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}

// Drop 删除所选用户表或视图。
func (m *Manager) Drop(ctx context.Context, schemaName, tableName string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	started := time.Now()
	table, err := m.target(schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(strings.ToLower(tableName), "sqlite_") {
		return nil, errors.New("不能删除 Sqlite 系统对象")
	}
	var kind string
	if err = m.db.QueryRowContext(ctx, "SELECT type FROM main.sqlite_master WHERE name = ? COLLATE NOCASE AND type IN ('table', 'view')", tableName).Scan(&kind); errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("数据表或视图不存在")
	} else if err != nil {
		return nil, err
	}
	if _, err = m.db.ExecContext(ctx, "DROP "+strings.ToUpper(kind)+" "+table); err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}
