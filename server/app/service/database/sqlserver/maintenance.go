package sqlserver

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

// maintenanceTable 验证所选用户表，拒绝系统对象和视图。
func (m *Manager) maintenanceTable(ctx context.Context, schemaName, tableName string) (string, error) {
	schema, table, err := m.target(schemaName, tableName)
	if err != nil {
		return "", err
	}
	var objectType string
	var system bool
	err = m.db.QueryRowContext(ctx, "SELECT o.type, o.is_ms_shipped FROM sys.objects o JOIN sys.schemas s ON s.schema_id = o.schema_id WHERE s.name = @p1 AND o.name = @p2", schema, tableName).Scan(&objectType, &system)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("数据表不存在或当前账户无权访问")
	}
	if err != nil {
		return "", err
	}
	if system || strings.EqualFold(schema, "sys") || strings.EqualFold(schema, "INFORMATION_SCHEMA") {
		return "", errors.New("不能维护 SqlServer 系统对象")
	}
	if strings.TrimSpace(objectType) != "U" {
		return "", errors.New("只有用户表支持维护操作，不能操作视图")
	}

	return table, nil
}

// Check 检查所选表的一致性。
func (m *Manager) Check(ctx context.Context, schemaName, tableName string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
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
	if _, err = m.db.ExecContext(ctx, "DBCC CHECKTABLE (N'"+strings.ReplaceAll(table, "'", "''")+"') WITH NO_INFOMSGS"); err != nil {
		return nil, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, ElapsedMs: time.Since(started).Milliseconds()}, nil
}

// Analyze 更新所选表的统计信息。
func (m *Manager) Analyze(ctx context.Context, schemaName, tableName string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
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
	if _, err = m.db.ExecContext(ctx, "UPDATE STATISTICS "+table); err != nil {
		return nil, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, ElapsedMs: time.Since(started).Milliseconds()}, nil
}

// Optimize 重建所选表的索引。
func (m *Manager) Optimize(ctx context.Context, schemaName, tableName string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
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
	if _, err = m.db.ExecContext(ctx, "ALTER INDEX ALL ON "+table+" REBUILD"); err != nil {
		return nil, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, ElapsedMs: time.Since(started).Milliseconds()}, nil
}

// Truncate 清空所选表并保留结构。
func (m *Manager) Truncate(ctx context.Context, schemaName, tableName string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
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
	if _, err = m.db.ExecContext(ctx, "TRUNCATE TABLE "+table); err != nil {
		return nil, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, ElapsedMs: time.Since(started).Milliseconds()}, nil
}

// Drop 删除所选用户表或视图。
func (m *Manager) Drop(ctx context.Context, schemaName, tableName string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	started := time.Now()
	schema, table, err := m.target(schemaName, tableName)
	if err != nil {
		return nil, err
	}
	var kind string
	var system bool
	if err = m.db.QueryRowContext(ctx, "SELECT o.type, o.is_ms_shipped FROM sys.objects o JOIN sys.schemas s ON s.schema_id = o.schema_id WHERE s.name = @p1 AND o.name = @p2", schema, tableName).Scan(&kind, &system); errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("数据表或视图不存在或当前账户无权访问")
	} else if err != nil {
		return nil, err
	}
	if system || strings.EqualFold(schema, "sys") || strings.EqualFold(schema, "INFORMATION_SCHEMA") {
		return nil, errors.New("不能删除 SqlServer 系统对象")
	}
	command := "DROP TABLE "
	if strings.TrimSpace(kind) == "V" {
		command = "DROP VIEW "
	} else if strings.TrimSpace(kind) != "U" {
		return nil, errors.New("只支持删除数据表和视图")
	}
	if _, err = m.db.ExecContext(ctx, command+table); err != nil {
		return nil, err
	}
	return &Result{Columns: []string{}, Rows: [][]interface{}{}, ElapsedMs: time.Since(started).Milliseconds()}, nil
}
