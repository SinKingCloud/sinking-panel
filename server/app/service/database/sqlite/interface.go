package sqlite

import (
	"context"
	"server/app/util/page"
)

// Interface 定义数据库管理器对外提供的操作。
type Interface interface {
	// Close 关闭本次操作的数据库连接。
	Close() error
	// Test 测试连接并返回数据库版本。
	Test(ctx context.Context) (map[string]interface{}, error)
	// Version 获取版本及支持的功能。
	Version(ctx context.Context) (map[string]interface{}, error)
	// Databases 分页查询数据库列表。
	Databases(ctx context.Context, keyword string, pageNumber, pageSize int) (_ *page.Result[string], err error)
	// DatabaseInfo 获取数据库详情。
	DatabaseInfo(ctx context.Context) (map[string]interface{}, error)
	// Schemas 查询架构列表。
	Schemas(ctx context.Context) (interface{}, error)
	// Tables 查询数据表和视图。
	Tables(ctx context.Context, schemaName string) (interface{}, error)
	// Columns 查询字段信息。
	Columns(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Indexes 查询索引及组成字段。
	Indexes(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// DDL 获取数据表定义。
	DDL(ctx context.Context, schemaName, tableName string) (string, error)
	// Query 执行查询并返回结果。
	Query(ctx context.Context, query, argsJSON string, limit int) (interface{}, error)
	// Execute 执行 SQL 并返回结果。
	Execute(ctx context.Context, query, argsJSON string, limit int) (interface{}, error)
	// Rows 分页查询数据行。
	Rows(ctx context.Context, schemaName, tableName, filtersJSON, filterLogic, orderColumn, direction string, page, pageSize int) (interface{}, error)
	// Insert 新增数据行。
	Insert(ctx context.Context, schemaName, tableName, valuesJSON string) (interface{}, error)
	// Update 修改数据行。
	Update(ctx context.Context, schemaName, tableName, valuesJSON, keysJSON string) (interface{}, error)
	// Delete 删除数据行。
	Delete(ctx context.Context, schemaName, tableName, keysJSON string) (interface{}, error)
	// CreateTable 创建数据表及索引、外键。
	CreateTable(ctx context.Context, schemaName, tableName string, columns []ColumnDefinition, primaryKeys []string, indexes []IndexDefinition, foreignKeys ...ForeignKeyDefinition) (*Result, error)
	// RenameTable 重命名数据表。
	RenameTable(ctx context.Context, schemaName, tableName, newName string) (*Result, error)
	// AddColumn 新增字段。
	AddColumn(ctx context.Context, schemaName, tableName string, column ColumnDefinition) (*Result, error)
	// UpdateColumn 修改字段。
	UpdateColumn(ctx context.Context, schemaName, tableName, columnName string, column ColumnDefinition) (*Result, error)
	// DropColumn 删除字段。
	DropColumn(ctx context.Context, schemaName, tableName, columnName string) (*Result, error)
	// CreateIndex 创建索引。
	CreateIndex(ctx context.Context, schemaName, tableName, indexName string, columns []IndexColumn, unique, primary bool) (*Result, error)
	// DropIndex 删除索引。
	DropIndex(ctx context.Context, schemaName, tableName, indexName string) (*Result, error)
	// CreateForeignKey 创建外键约束。
	CreateForeignKey(ctx context.Context, schemaName, tableName string, key ForeignKeyDefinition) (*Result, error)
	// UpdateForeignKey 修改外键约束。
	UpdateForeignKey(ctx context.Context, schemaName, tableName, oldName string, key ForeignKeyDefinition) (*Result, error)
	// DropForeignKey 删除外键约束。
	DropForeignKey(ctx context.Context, schemaName, tableName, name string) (*Result, error)
	// CreateView 创建视图。
	CreateView(ctx context.Context, schemaName, viewName, statement string) (*Result, error)
	// Objects 查询数据库对象。
	Objects(ctx context.Context, schemaName, kind string) (interface{}, error)
	// SelectObjects 分页查询数据库对象。
	SelectObjects(ctx context.Context, schema, kind, keyword string, pageNumber, pageSize int) (interface{}, error)
	// ObjectDefinition 获取对象定义。
	ObjectDefinition(ctx context.Context, schemaName, kind, name string) (string, error)
	// SaveObject 创建或修改数据库对象。
	SaveObject(ctx context.Context, schemaName, kind, name, definition string, replace bool) (interface{}, error)
	// DropObject 删除数据库对象。
	DropObject(ctx context.Context, schemaName, kind, name string) (interface{}, error)
	// Import 从服务器文件导入 SQL。
	Import(ctx context.Context, path string, progress ...func(float64, string)) (map[string]interface{}, error)
	// Export 导出 SQL 到服务器文件。
	Export(ctx context.Context, path, schemaName, tableName string, progress ...func(float64, string)) (result map[string]interface{}, err error)
	// Check 检查数据表。
	Check(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Analyze 更新查询统计。
	Analyze(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Optimize 优化数据表。
	Optimize(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Truncate 清空数据表。
	Truncate(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Drop 删除数据表。
	Drop(ctx context.Context, schemaName, tableName string) (interface{}, error)
}
