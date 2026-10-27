package sqlserver

import (
	"context"
	"errors"
	"time"
)

// UpdateForeignKey 同一事务内替换外键约束，新约束失败时保留原约束。
func (m *Manager) UpdateForeignKey(ctx context.Context, schemaName, tableName, oldName string, key ForeignKeyDefinition) (_ *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	table, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	old, err := m.quote(oldName)
	if err != nil {
		return nil, err
	}
	definition, err := m.foreignKeyDefinition(key)
	if err != nil {
		return nil, err
	}
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var found int
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys.foreign_keys WHERE parent_object_id = OBJECT_ID(@p1) AND name = @p2", table, oldName).Scan(&found); err != nil {
		return nil, err
	}
	if found != 1 {
		return nil, errors.New("当前表中不存在该外键")
	}
	if _, err = tx.ExecContext(ctx, "ALTER TABLE "+table+" DROP CONSTRAINT "+old); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, "ALTER TABLE "+table+" WITH CHECK ADD "+definition); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}
