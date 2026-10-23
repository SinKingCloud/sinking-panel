package sqlite

import "context"

// UpdateForeignKey 在同一事务替换外键，新约束或数据校验失败时保留原约束。
func (m *Manager) UpdateForeignKey(ctx context.Context, schemaName, tableName, oldName string, key ForeignKeyDefinition) (*Result, error) {
	return m.rebuildTable(ctx, schemaName, tableName, nil, oldName, &key, false)
}
