package mysql

import (
	"context"
	"server/app/util/page"
)

// Interface 定义 MySQL 管理操作，由 Manager 实现。
type Interface interface {
	// Close 关闭连接。
	Close() error
	// Test 返回 MySQL 实际版本，不暴露认证信息。
	Test(ctx context.Context) (map[string]interface{}, error)
	// Databases 按名称分页查询当前账户有权访问的数据库，默认隐藏系统库。
	Databases(ctx context.Context, keyword string, pageNumber, pageSize int, showSystem bool, orderByType string) (*page.Result[string], error)
	// Version 返回服务端版本及当前管理功能的版本能力，权限仍由数据库检查。
	Version(ctx context.Context) (map[string]interface{}, error)
	// ConnectionInfo 读取实际网络配置，不把账号来源授权误认为服务器已开放公网访问。
	ConnectionInfo(ctx context.Context) (map[string]interface{}, error)
	// Charsets 读取远端实际支持的字符集和排序规则，供创建、修改数据库共用。
	Charsets(ctx context.Context) (*CharsetOptions, error)
	// DatabaseInfo 读取真实数据库的字符集、排序规则和表空间信息。
	DatabaseInfo(ctx context.Context, name string) (map[string]interface{}, error)
	// CreateDatabase 创建真实数据库；字符集和排序规则使用服务器支持的名称。
	CreateDatabase(ctx context.Context, name, charset, collation string) error
	// UpdateDatabase 修改真实数据库的默认字符集和排序规则，不移动或重命名数据库。
	UpdateDatabase(ctx context.Context, name, charset, collation string) error
	// DropDatabase 删除真实业务数据库；已不存在的数据库无需重复删除。
	DropDatabase(ctx context.Context, name string) error
	// DropDatabases 在发送任何删除前检查整批目标，并报告失败前已完成的数量。
	DropDatabases(ctx context.Context, names []string) error
	// Tables 列出当前数据库的表与视图。
	Tables(ctx context.Context, schemaName string) (interface{}, error)
	// Columns 读取字段类型、默认值、主键和生成列属性。
	Columns(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Indexes 列出 MySQL 索引与组成字段。
	Indexes(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Execute 执行一次 SQL，根据数据库返回的结果集自动识别查询，并保留影响行数。
	Execute(ctx context.Context, query string, limit int) (interface{}, error)
	// Rows 分页浏览表数据，使用完整主键保持排序稳定。
	Rows(ctx context.Context, schemaName, tableName, filtersJSON, filterLogic, orderColumn, direction string, page, pageSize int) (interface{}, error)
	// Insert 插入一行数据，字段名与值分别校验和绑定。
	Insert(ctx context.Context, schemaName, tableName, valuesJSON string) (interface{}, error)
	// Update 按完整主键或非空唯一键更新一行数据，影响行数异常时回滚。
	Update(ctx context.Context, schemaName, tableName, valuesJSON, keysJSON string) (interface{}, error)
	// Delete 按完整主键或非空唯一键删除一行数据，没有可靠定位条件时拒绝删除。
	Delete(ctx context.Context, schemaName, tableName, keysJSON string) (interface{}, error)
	// CreateTable 同时创建真实 MySQL 表及索引、外键，复合索引按提交顺序建立。
	CreateTable(ctx context.Context, schemaName, tableName string, columns []ColumnDefinition, primaryKeys []string, indexes []IndexDefinition, foreignKeys ...ForeignKeyDefinition) (*Result, error)
	// RenameTable 重命名真实表或视图，数据库原生维护其索引与约束。
	RenameTable(ctx context.Context, schemaName, tableName, newName string) (*Result, error)
	// AddColumn 新增真实字段，重复字段和版本限制交由 MySQL 原生验证。
	AddColumn(ctx context.Context, schemaName, tableName string, column ColumnDefinition) (*Result, error)
	// UpdateColumn 修改字段，保留未提交的备注、排序规则和 ON UPDATE 属性。
	UpdateColumn(ctx context.Context, schemaName, tableName, columnName string, column ColumnDefinition) (*Result, error)
	// DropColumn 删除真实字段，不自动级联删除其他表的约束。
	DropColumn(ctx context.Context, schemaName, tableName, columnName string) (*Result, error)
	// CreateIndex 创建普通、唯一索引或主键。
	CreateIndex(ctx context.Context, schemaName, tableName, indexName string, columns []IndexColumn, unique, primary bool, comment string) (*Result, error)
	// UpdateIndex 修改索引备注，沿用服务器生成的完整定义，保留前缀、表达式、可见性等属性。
	UpdateIndex(ctx context.Context, schemaName, tableName, indexName, comment string) (*Result, error)
	// DropIndex 删除指定索引；主键使用 MySQL 的 DROP PRIMARY KEY。
	DropIndex(ctx context.Context, schemaName, tableName, indexName string) (*Result, error)
	// DDL 返回服务器生成的建表或建视图语句。
	DDL(ctx context.Context, schemaName, tableName string) (string, error)
	// CreateForeignKey 在真实表上添加外键，字段类型和引用权限由数据库核验。
	CreateForeignKey(ctx context.Context, schemaName, tableName string, key ForeignKeyDefinition) (*Result, error)
	// DropForeignKey 删除指定表的外键，其他类型约束不可从此入口删除。
	DropForeignKey(ctx context.Context, schemaName, tableName, name string) (*Result, error)
	// UpdateForeignKey 使用一条 ALTER TABLE 替换指定外键，失败保留原表约束。
	UpdateForeignKey(ctx context.Context, schemaName, tableName, oldName string, key ForeignKeyDefinition) (*Result, error)
	// Objects 分页列举当前库的对象；页码和容量均为 0 时供内部业务读取完整元数据，外键仍返回聚合数组。
	Objects(ctx context.Context, schemaName, kind, keyword string, pageNumber, pageSize int) (interface{}, error)
	// ObjectDefinition 返回服务器的对象定义，保留存储程序及触发器正文。
	ObjectDefinition(ctx context.Context, schemaName, kind, name string) (string, error)
	// SaveObject 保存真实对象定义；存储程序及触发器替换失败时恢复原定义与授权。
	SaveObject(ctx context.Context, schemaName, kind, name, definition string, replace bool) (interface{}, error)
	// DropObject 删除真实视图、存储程序或触发器，外键仍定位到所属表后删除。
	DropObject(ctx context.Context, schemaName, kind, name string) (interface{}, error)
	// Users 分页列出账号及其在所有数据库的授权，角色授权单独返回，不将对象权限误认为整库权限。
	Users(ctx context.Context, keyword string, pageNumber, pageSize int) (interface{}, error)
	// CreateUser 创建独立 MySQL 账号，不自动授予数据库权限。
	CreateUser(ctx context.Context, name, host, password string) (interface{}, error)
	// ChangePassword 修改账号密码，返回是否为当前连接账户，供调用方同步连接凭据。
	ChangePassword(ctx context.Context, name, host, password string) (*Result, bool, error)
	// ChangeAccess 修改账号的连接来源，保留原账号的密码和已有授权。
	ChangeAccess(ctx context.Context, name, host, newHost string) (interface{}, error)
	// Permissions 读取账号可编辑的直接整库权限；未指定账号时仅返回当前版本支持的权限。
	Permissions(ctx context.Context, name, host string) (*UserPermissions, error)
	// UpdatePermissions 按提交的最终状态计算差量，只修改提交数据库的直接整库权限。
	UpdatePermissions(ctx context.Context, name, host string, grants []DatabasePrivileges) (*Result, error)
	// DeleteUser 删除业务账号，保护当前连接与系统管理员。
	DeleteUser(ctx context.Context, name, host string) (interface{}, error)
	// Check 检查 MySQL 业务表。
	Check(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Analyze 更新 MySQL 表统计信息。
	Analyze(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Optimize 整理 MySQL 业务表。
	Optimize(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Repair 修复支持该功能的 MySQL 存储引擎表。
	Repair(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Truncate 清空 MySQL 业务表，保留表结构。
	Truncate(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Drop 删除 MySQL 业务表或视图。
	Drop(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Import 在同一会话逐条恢复 SQL，支持 mysqldump 的 DELIMITER、版本注释及二进制字面量。
	Import(ctx context.Context, path string, progress ...func(float64, string)) (map[string]interface{}, error)
	// Export 输出可直接恢复的 SQL，包含表结构、完整数据、视图、存储程序和触发器。
	Export(ctx context.Context, path, schema, table string, progress ...func(float64, string)) (map[string]interface{}, error)
}
