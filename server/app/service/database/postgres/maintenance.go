package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

// Analyze 更新 Postgres 表或物化视图的统计信息。
func (m *Manager) Analyze(ctx context.Context, schemaName, tableName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	table, kind, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind == "v" {
		return nil, errors.New("视图不支持此表维护操作")
	}
	return m.executeMaintenance(ctx, "ANALYZE "+table)
}

// Optimize 整理 Postgres 表或物化视图，不终止其他会话。
func (m *Manager) Optimize(ctx context.Context, schemaName, tableName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	table, kind, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind == "v" {
		return nil, errors.New("视图不支持此表维护操作")
	}
	return m.executeMaintenance(ctx, "VACUUM ANALYZE "+table)
}

// Truncate 清空 Postgres 业务表，保留表结构。
func (m *Manager) Truncate(ctx context.Context, schemaName, tableName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	table, kind, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind == "v" {
		return nil, errors.New("视图不支持此表维护操作")
	}
	if kind == "m" {
		return nil, errors.New("物化视图不支持清空表操作")
	}
	return m.executeMaintenance(ctx, "TRUNCATE TABLE "+table)
}

// Drop 删除 Postgres 业务表、视图或物化视图。
func (m *Manager) Drop(ctx context.Context, schemaName, tableName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	table, kind, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	statement := "DROP TABLE "
	switch kind {
	case "v":
		statement = "DROP VIEW "
	case "m":
		statement = "DROP MATERIALIZED VIEW "
	}
	return m.executeMaintenance(ctx, statement+table)
}

// maintenanceTable 校验维护对象范围并返回 Postgres 对象类型。
func (m *Manager) maintenanceTable(ctx context.Context, schemaName, tableName string) (string, string, error) {
	schema, table, err := m.target(schemaName, tableName)
	if err != nil {
		return "", "", err
	}
	if schema == "information_schema" || strings.HasPrefix(schema, "pg_") || m.database == "template0" || m.database == "template1" {
		return "", "", errors.New("不能维护系统模式中的对象")
	}
	var kind string
	if err = m.db.QueryRowContext(ctx, "SELECT c.relkind FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = $1 AND c.relname = $2", schema, tableName).Scan(&kind); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", errors.New("表不存在或无权访问")
		}
		return "", "", err
	}
	switch kind {
	case "v", "m", "r", "p":
	default:
		return "", "", errors.New("此对象类型不支持表维护")
	}
	return table, kind, nil
}

// executeMaintenance 直接执行维护语句；VACUUM 必须在事务外执行。
func (m *Manager) executeMaintenance(ctx context.Context, statement string) (*Result, error) {
	started := time.Now()
	result, err := m.db.ExecContext(ctx, statement)
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
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
