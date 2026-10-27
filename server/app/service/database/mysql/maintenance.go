package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Check 检查 MySQL 业务表。
func (m *Manager) Check(ctx context.Context, schemaName, tableName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	table, kind, _, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind == "VIEW" {
		return nil, errors.New("视图不支持此表维护操作")
	}
	return m.readMaintenance(ctx, "CHECK TABLE "+table)
}

// Analyze 更新 MySQL 表统计信息。
func (m *Manager) Analyze(ctx context.Context, schemaName, tableName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	table, kind, _, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind == "VIEW" {
		return nil, errors.New("视图不支持此表维护操作")
	}
	return m.readMaintenance(ctx, "ANALYZE TABLE "+table)
}

// Optimize 整理 MySQL 业务表。
func (m *Manager) Optimize(ctx context.Context, schemaName, tableName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	table, kind, _, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind == "VIEW" {
		return nil, errors.New("视图不支持此表维护操作")
	}
	return m.readMaintenance(ctx, "OPTIMIZE TABLE "+table)
}

// Repair 修复支持该功能的 MySQL 存储引擎表。
func (m *Manager) Repair(ctx context.Context, schemaName, tableName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	table, kind, engine, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind == "VIEW" {
		return nil, errors.New("视图不支持此表维护操作")
	}
	switch strings.ToUpper(engine) {
	case "MYISAM", "ARIA", "ARCHIVE", "CSV":
	default:
		return nil, errors.New("当前存储引擎不支持修复表；仅支持 MyISAM、Aria、ARCHIVE 和 CSV")
	}
	return m.readMaintenance(ctx, "REPAIR TABLE "+table)
}

// Truncate 清空 MySQL 业务表，保留表结构。
func (m *Manager) Truncate(ctx context.Context, schemaName, tableName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	table, kind, _, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind == "VIEW" {
		return nil, errors.New("视图不支持此表维护操作")
	}
	return m.executeMaintenance(ctx, "TRUNCATE TABLE "+table)
}

// Drop 删除 MySQL 业务表或视图。
func (m *Manager) Drop(ctx context.Context, schemaName, tableName string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	table, kind, _, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	statement := "DROP TABLE "
	if kind == "VIEW" {
		statement = "DROP VIEW "
	}
	return m.executeMaintenance(ctx, statement+table)
}

// maintenanceTable 校验对象范围并读取类型；不决定具体维护操作。
func (m *Manager) maintenanceTable(ctx context.Context, schemaName, tableName string) (string, string, string, error) {
	schema, table, err := m.target(schemaName, tableName)
	if err != nil {
		return "", "", "", err
	}
	switch strings.ToLower(schema) {
	case "", "mysql", "information_schema", "performance_schema", "sys":
		return "", "", "", errors.New("不能维护系统数据库中的对象")
	}
	var kind string
	var engine sql.NullString
	if err = m.db.QueryRowContext(ctx, "SELECT TABLE_TYPE, ENGINE FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?", schema, tableName).Scan(&kind, &engine); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", "", errors.New("表不存在或无权访问")
		}
		return "", "", "", err
	}
	if kind != "VIEW" && kind != "BASE TABLE" {
		return "", "", "", errors.New("此对象类型不支持表维护")
	}
	return table, kind, engine.String, nil
}

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

// readMaintenance 检查 MySQL 维护结果中的错误行，避免仅凭查询成功误报成功。
func (m *Manager) readMaintenance(ctx context.Context, statement string) (*Result, error) {
	started := time.Now()
	rows, err := m.db.QueryContext(ctx, statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result, err := m.read(rows, 1000)
	if err != nil {
		return nil, err
	}
	typeIndex, textIndex := -1, -1
	for index, column := range result.Columns {
		switch strings.ToLower(column) {
		case "msg_type":
			typeIndex = index
		case "msg_text":
			textIndex = index
		}
	}
	if typeIndex >= 0 && textIndex >= 0 {
		for _, row := range result.Rows {
			if strings.EqualFold(fmt.Sprint(row[typeIndex]), "error") {
				return nil, fmt.Errorf("表维护失败：%s", fmt.Sprint(row[textIndex]))
			}
		}
	}
	result.ElapsedMs = time.Since(started).Milliseconds()
	return result, nil
}
