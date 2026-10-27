package sqlserver

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
	// ConnectionInfo 获取当前连接信息。
	ConnectionInfo(ctx context.Context) (result map[string]interface{}, err error)
	// Databases 分页查询数据库列表。
	Databases(ctx context.Context, keyword string, pageNumber, pageSize int, showSystem bool, orderByType string) (_ *page.Result[string], err error)
	// DatabaseInfo 获取数据库详情。
	DatabaseInfo(ctx context.Context, name string) (_ map[string]interface{}, err error)
	// CreateDatabase 创建数据库。
	CreateDatabase(ctx context.Context, name string) (err error)
	// RenameDatabase 重命名数据库。
	RenameDatabase(ctx context.Context, name, newName string) (err error)
	// DropDatabase 删除数据库。
	DropDatabase(ctx context.Context, name string) (err error)
	// DropDatabases 批量删除数据库。
	DropDatabases(ctx context.Context, names []string) (err error)
	// Schemas 查询架构列表。
	Schemas(ctx context.Context) (_ interface{}, err error)
	// Tables 查询数据表和视图。
	Tables(ctx context.Context, schemaName string) (_ interface{}, err error)
	// Columns 查询字段信息。
	Columns(ctx context.Context, schemaName, tableName string) (_ interface{}, err error)
	// Indexes 查询索引及组成字段。
	Indexes(ctx context.Context, schemaName, tableName string) (_ interface{}, err error)
	// DDL 获取数据表定义。
	DDL(ctx context.Context, schemaName, tableName string, includeForeignKeys ...bool) (_ string, err error)
	// Query 执行查询并返回结果。
	Query(ctx context.Context, query, argsJSON string, rowLimit int) (_ interface{}, err error)
	// Execute 执行 SQL 并返回结果。
	Execute(ctx context.Context, query, argsJSON string, limit int) (_ interface{}, err error)
	// Rows 分页查询数据行。
	Rows(ctx context.Context, schemaName, tableName, filtersJSON, filterLogic, orderColumn, orderDirection string, pageNumber, pageSize int) (_ interface{}, err error)
	// Insert 新增数据行。
	Insert(ctx context.Context, schemaName, tableName, valuesJSON string) (_ interface{}, err error)
	// Update 修改数据行。
	Update(ctx context.Context, schemaName, tableName, valuesJSON, keysJSON string) (_ interface{}, err error)
	// Delete 删除数据行。
	Delete(ctx context.Context, schemaName, tableName, keysJSON string) (_ interface{}, err error)
	// CreateTable 创建数据表及索引、外键。
	CreateTable(ctx context.Context, schemaName, tableName string, columns []ColumnDefinition, primaryKeys []string, indexes []IndexDefinition, foreignKeys ...ForeignKeyDefinition) (_ *Result, err error)
	// RenameTable 重命名数据表。
	RenameTable(ctx context.Context, schemaName, tableName, newName string) (_ *Result, err error)
	// AddColumn 新增字段。
	AddColumn(ctx context.Context, schemaName, tableName string, column ColumnDefinition) (_ *Result, err error)
	// UpdateColumn 修改字段。
	UpdateColumn(ctx context.Context, schemaName, tableName, columnName string, column ColumnDefinition) (_ *Result, err error)
	// DropColumn 删除字段。
	DropColumn(ctx context.Context, schemaName, tableName, columnName string) (_ *Result, err error)
	// CreateIndex 创建索引。
	CreateIndex(ctx context.Context, schemaName, tableName, indexName string, columns []IndexColumn, unique, primary bool) (_ *Result, err error)
	// DropIndex 删除索引。
	DropIndex(ctx context.Context, schemaName, tableName, indexName string) (_ *Result, err error)
	// CreateForeignKey 创建外键约束。
	CreateForeignKey(ctx context.Context, schemaName, tableName string, key ForeignKeyDefinition) (_ *Result, err error)
	// UpdateForeignKey 修改外键约束。
	UpdateForeignKey(ctx context.Context, schemaName, tableName, oldName string, key ForeignKeyDefinition) (_ *Result, err error)
	// DropForeignKey 删除外键约束。
	DropForeignKey(ctx context.Context, schemaName, tableName, name string) (_ *Result, err error)
	// CreateView 创建视图。
	CreateView(ctx context.Context, schemaName, viewName, statement string) (_ *Result, err error)
	// Objects 查询数据库对象。
	Objects(ctx context.Context, schemaName, kind string) (_ interface{}, err error)
	// SelectObjects 分页查询数据库对象。
	SelectObjects(ctx context.Context, schema, kind, keyword string, pageNumber, pageSize int) (interface{}, error)
	// ObjectDefinition 获取对象定义。
	ObjectDefinition(ctx context.Context, schemaName, kind, name string) (_ string, err error)
	// SaveObject 创建或修改数据库对象。
	SaveObject(ctx context.Context, schemaName, kind, name, definition string, replace bool) (_ interface{}, err error)
	// DropObject 删除数据库对象。
	DropObject(ctx context.Context, schemaName, kind, name string) (_ interface{}, err error)
	// Users 分页查询账户。
	Users(ctx context.Context, keyword string, pageNumber, pageSize int) (_ interface{}, err error)
	// CreateUser 创建账户并设置权限。
	CreateUser(ctx context.Context, accountName, accountPassword string, permissions []string) (_ interface{}, err error)
	// ChangePassword 修改密码并返回是否为当前连接账户。
	ChangePassword(ctx context.Context, accountName, accountPassword string) (out *Result, current bool, err error)
	// Grant 授予账户权限。
	Grant(ctx context.Context, accountName string, permissions []string) (_ interface{}, err error)
	// Revoke 撤销账户权限。
	Revoke(ctx context.Context, accountName string, permissions []string) (_ interface{}, err error)
	// SetPrivileges 设置账户权限。
	SetPrivileges(ctx context.Context, accountName string, permissions []string) (_ interface{}, err error)
	// DeleteUser 删除账户及数据库关联。
	DeleteUser(ctx context.Context, accountName string) (_ interface{}, err error)
	// Import 从服务器文件导入 SQL。
	Import(ctx context.Context, path string, progress ...func(float64, string)) (map[string]interface{}, error)
	// Export 导出 SQL 到服务器文件。
	Export(ctx context.Context, path, schemaName, tableName string, progress ...func(float64, string)) (result map[string]interface{}, err error)
	// Check 检查数据表。
	Check(ctx context.Context, schemaName, tableName string) (_ interface{}, err error)
	// Analyze 更新查询统计。
	Analyze(ctx context.Context, schemaName, tableName string) (_ interface{}, err error)
	// Optimize 优化数据表。
	Optimize(ctx context.Context, schemaName, tableName string) (_ interface{}, err error)
	// Truncate 清空数据表。
	Truncate(ctx context.Context, schemaName, tableName string) (_ interface{}, err error)
	// Drop 删除数据表。
	Drop(ctx context.Context, schemaName, tableName string) (_ interface{}, err error)
}
