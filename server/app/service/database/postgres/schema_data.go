package postgres

import (
	"context"
	"errors"
	"strings"
	"time"
)

// foreignKeySQL 生成字段一一对应的外键定义，不自动创建引用表或删除依赖。
func (m *Manager) foreignKeySQL(schemaName string, key ForeignKeyDefinition) (string, error) {
	name, err := m.quote(key.Name)
	if err != nil || len(key.Name) > 63 {
		return "", errors.New("外键名称不合法")
	}
	if len(key.Columns) == 0 || len(key.Columns) > 32 || len(key.Columns) != len(key.ReferenceColumns) {
		return "", errors.New("外键需要 1 至 32 组对应字段")
	}
	if key.ReferenceSchema == "" {
		key.ReferenceSchema = schemaName
	}
	schema, err := m.quote(key.ReferenceSchema)
	if err != nil {
		return "", err
	}
	table, err := m.quote(key.ReferenceTable)
	if err != nil {
		return "", err
	}
	columns := make([]string, 0, len(key.Columns))
	references := make([]string, 0, len(key.ReferenceColumns))
	known := make(map[string]bool, len(key.Columns))
	referenceKnown := make(map[string]bool, len(key.Columns))
	for index, column := range key.Columns {
		if known[column] || referenceKnown[key.ReferenceColumns[index]] {
			return "", errors.New("外键字段不能重复")
		}
		known[column] = true
		referenceKnown[key.ReferenceColumns[index]] = true
		field, quoteErr := m.quote(column)
		if quoteErr != nil {
			return "", quoteErr
		}
		reference, quoteErr := m.quote(key.ReferenceColumns[index])
		if quoteErr != nil {
			return "", quoteErr
		}
		columns = append(columns, field)
		references = append(references, reference)
	}
	clause := "CONSTRAINT " + name + " FOREIGN KEY (" + strings.Join(columns, ",") + ") REFERENCES " + schema + "." + table + " (" + strings.Join(references, ",") + ")"
	for _, rule := range []struct{ name, value string }{{"DELETE", key.OnDelete}, {"UPDATE", key.OnUpdate}} {
		value := strings.ToUpper(strings.Join(strings.Fields(rule.value), " "))
		if value == "" {
			continue
		}
		switch value {
		case "NO ACTION", "RESTRICT", "CASCADE", "SET NULL", "SET DEFAULT":
		default:
			return "", errors.New("外键规则不合法")
		}
		clause += " ON " + rule.name + " " + value
	}
	return clause, nil
}

// CreateForeignKey 在真实表上添加外键，字段类型和引用权限由数据库核验。
func (m *Manager) CreateForeignKey(ctx context.Context, schemaName, tableName string, key ForeignKeyDefinition) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	table, kind, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind != "r" && kind != "p" {
		return nil, errors.New("只能修改数据表的外键")
	}
	schema, _, _ := m.target(schemaName, tableName)
	clause, err := m.foreignKeySQL(schema, key)
	if err != nil {
		return nil, err
	}
	return m.executeMaintenance(ctx, "ALTER TABLE "+table+" ADD "+clause)
}

// DropForeignKey 删除指定表的外键，其他类型约束不可从此入口删除。
func (m *Manager) DropForeignKey(ctx context.Context, schemaName, tableName, name string) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	table, kind, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind != "r" && kind != "p" {
		return nil, errors.New("只能修改数据表的外键")
	}
	schema, _, _ := m.target(schemaName, tableName)
	quoted, err := m.quote(name)
	if err != nil {
		return nil, err
	}
	var count int
	err = m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pg_constraint p JOIN pg_class c ON c.oid = p.conrelid JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = $1 AND c.relname = $2 AND p.conname = $3 AND p.contype = 'f'", schema, tableName, name).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count != 1 {
		return nil, errors.New("当前表不存在此外键")
	}
	return m.executeMaintenance(ctx, "ALTER TABLE "+table+" DROP CONSTRAINT "+quoted)
}

// UpdateForeignKey 使用一条 ALTER TABLE 替换指定外键，失败保留原表约束。
func (m *Manager) UpdateForeignKey(ctx context.Context, schemaName, tableName, oldName string, key ForeignKeyDefinition) (out *Result, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	table, kind, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if kind != "r" && kind != "p" {
		return nil, errors.New("只能修改数据表的外键")
	}
	schema, _, _ := m.target(schemaName, tableName)
	old, err := m.quote(oldName)
	if err != nil {
		return nil, err
	}
	var count int
	err = m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pg_constraint p JOIN pg_class c ON c.oid = p.conrelid JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = $1 AND c.relname = $2 AND p.conname = $3 AND p.contype = 'f'", schema, tableName, oldName).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count != 1 {
		return nil, errors.New("当前表不存在此外键")
	}
	clause, err := m.foreignKeySQL(schema, key)
	if err != nil {
		return nil, err
	}
	return m.executeMaintenance(ctx, "ALTER TABLE "+table+" DROP CONSTRAINT "+old+", ADD "+clause)
}
