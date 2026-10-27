package postgres

import (
	"context"
	"database/sql"
	"server/app/util/page"
)

// Interface 定义 Postgres 管理操作，由 Manager 实现。
type Interface interface {
	// Close 释放连接池。
	Close() error
	// Test 返回 Postgres 服务端实际版本，不暴露认证信息。
	Test(ctx context.Context) (map[string]interface{}, error)
	// Databases 按名称搜索并分页返回当前账户可访问的数据库，默认隐藏模板库。
	Databases(ctx context.Context, keyword string, pageNumber, pageSize int, showSystem bool, orderByType string) (*page.Result[string], error)
	// Version 返回 Postgres 实际版本，单独标识存储过程与物化视图等版本差异。
	Version(ctx context.Context) (map[string]interface{}, error)
	// ConnectionInfo 读取监听和 TLS 设置；客户端不改写服务器的 pg_hba.conf。
	ConnectionInfo(ctx context.Context) (map[string]interface{}, error)
	// DatabaseInfo 读取真实数据库的所有者、编码、排序规则和占用空间。
	DatabaseInfo(ctx context.Context, name string) (map[string]interface{}, error)
	// CreateDatabase 创建真实数据库，继承服务器模板库的默认配置。
	CreateDatabase(ctx context.Context, name string) error
	// RenameDatabase 从其他数据库连接重命名业务数据库，不终止已有连接。
	RenameDatabase(ctx context.Context, name, newName string) error
	// DropDatabase 删除真实业务数据库，不终止已有连接；不存在时允许重试。
	DropDatabase(ctx context.Context, name string) error
	// DropDatabases 在发送任何删除前检查整批目标，并报告失败前已完成的数量。
	DropDatabases(ctx context.Context, names []string) error
	// Schemas 列出 Postgres 业务模式。
	Schemas(ctx context.Context) (interface{}, error)
	// Tables 列出表、视图和物化视图。
	Tables(ctx context.Context, schemaName string) (interface{}, error)
	// Columns 读取 Postgres 字段、主键与只读生成属性。
	Columns(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Indexes 读取 Postgres 索引及组成表达式。
	Indexes(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Query 执行返回结果集的 SQL，限制行数与响应大小。
	Query(ctx context.Context, query, argsJSON string, rowLimit int) (interface{}, error)
	// Execute 执行 SQL；根据数据库返回的结果自动识别查询，并保留影响行数。
	Execute(ctx context.Context, query, argsJSON string, limit int) (interface{}, error)
	// Rows 分页浏览表数据，使用完整主键保持排序稳定。
	Rows(ctx context.Context, schemaName, tableName, filtersJSON, filterLogic, orderColumn, orderDirection string, pageNumber, pageSize int) (interface{}, error)
	// Insert 插入一行数据，字段名与值分别校验和绑定。
	Insert(ctx context.Context, schemaName, tableName, valuesJSON string) (interface{}, error)
	// Update 按完整主键或唯一键更新一行数据，影响行数异常时回滚。
	Update(ctx context.Context, schemaName, tableName, valuesJSON, keysJSON string) (interface{}, error)
	// Delete 按完整主键或唯一键删除一行数据，无法唯一定位时拒绝操作。
	Delete(ctx context.Context, schemaName, tableName, keysJSON string) (interface{}, error)
	// CreateTable 创建真实 Postgres 表，自增使用兼容旧版本的 serial 类型。
	CreateTable(ctx context.Context, schemaName, tableName string, columns []ColumnDefinition, primaryKeys []string, indexes []IndexDefinition, foreignKeys ...ForeignKeyDefinition) (*Result, error)
	// RenameTable 重命名表、视图或物化视图。
	RenameTable(ctx context.Context, schemaName, tableName, newName string) (*Result, error)
	// AddColumn 新增字段，兼容 serial 自增字段与自定义类型。
	AddColumn(ctx context.Context, schemaName, tableName string, column ColumnDefinition) (*Result, error)
	// UpdateColumn 在同一事务中修改字段类型、默认值、非空属性及名称，失败整体回滚。
	UpdateColumn(ctx context.Context, schemaName, tableName, columnName string, column ColumnDefinition) (*Result, error)
	// DropColumn 删除字段，不自动 CASCADE 删除依赖对象。
	DropColumn(ctx context.Context, schemaName, tableName, columnName string) (*Result, error)
	// CreateIndex 创建普通或唯一索引；primary 以真实主键约束建立。
	CreateIndex(ctx context.Context, schemaName, tableName, indexName string, columns []IndexColumn, unique, primary bool) (*Result, error)
	// DropIndex 删除指定表上的索引；约束索引通过 DROP CONSTRAINT 删除，依赖限制由数据库处理。
	DropIndex(ctx context.Context, schemaName, tableName, indexName string) (*Result, error)
	// DDL 从 Postgres 系统目录生成对象结构、约束和索引定义，不读取表数据。
	DDL(ctx context.Context, schemaName, tableName string, exportConnection ...*sql.Conn) (string, error)
	// CreateView 创建真实 Postgres 视图，只接受一条 SELECT 或 WITH 查询。
	CreateView(ctx context.Context, schemaName, viewName, statement string) (*Result, error)
	// CreateForeignKey 在真实表上添加外键，字段类型和引用权限由数据库核验。
	CreateForeignKey(ctx context.Context, schemaName, tableName string, key ForeignKeyDefinition) (*Result, error)
	// DropForeignKey 删除指定表的外键，其他类型约束不可从此入口删除。
	DropForeignKey(ctx context.Context, schemaName, tableName, name string) (*Result, error)
	// UpdateForeignKey 使用一条 ALTER TABLE 替换指定外键，失败保留原表约束。
	UpdateForeignKey(ctx context.Context, schemaName, tableName, oldName string, key ForeignKeyDefinition) (*Result, error)
	// Objects 返回 Postgres 业务模式中的视图、程序、触发器和外键。
	Objects(ctx context.Context, schemaName, kind string) (interface{}, error)
	// ObjectDefinition 返回指定对象的真实 DDL；重载程序和同名触发器通过 identity 定位。
	ObjectDefinition(ctx context.Context, schemaName, kind, name string) (string, error)
	// SaveObject 保存真实对象，函数、普通视图与新版本触发器使用原生 CREATE OR REPLACE。
	SaveObject(ctx context.Context, schemaName, kind, name, definition string, replace bool) (interface{}, error)
	// DropObject 删除指定业务对象，重载程序及同名触发器通过目录信息精确定位。
	DropObject(ctx context.Context, schemaName, kind, name string) (interface{}, error)
	// SelectObjects 按名称搜索并分页查询对象摘要；定义在详情接口按需读取。
	SelectObjects(ctx context.Context, schema, kind, keyword string, pageNumber, pageSize int) (interface{}, error)
	// Users 分页列出角色及其在当前数据库、架构内已有对象上的有效权限。
	Users(ctx context.Context, schemaName, keyword string, pageNumber, pageSize int) (interface{}, error)
	// CreateUser 创建登录角色，不自动授予对象权限。
	CreateUser(ctx context.Context, accountName, accountPassword string) (interface{}, error)
	// ChangePassword 修改角色密码，由 Postgres 校验权限；返回是否修改了当前登录账户。
	ChangePassword(ctx context.Context, accountName, accountPassword string) (*Result, bool, error)
	// Grant 在事务中授予当前数据库、模式和现有表的权限。
	Grant(ctx context.Context, accountName, schemaName string, permissions []string) (interface{}, error)
	// Revoke 在事务中撤回当前数据库、模式和现有表的权限。
	Revoke(ctx context.Context, accountName, schemaName string, permissions []string) (interface{}, error)
	// DeleteUser 撤回当前范围授权后删除角色，不删除角色拥有的数据。
	DeleteUser(ctx context.Context, accountName, schemaName string) (interface{}, error)
	// SetPrivileges 原子更新当前数据库和架构的权限方案，不修改角色身份和其他数据库。
	SetPrivileges(ctx context.Context, accountName, schemaName string, permissions []string) (interface{}, error)
	// Permissions 读取角色在各数据库中的直接权限，不把 PUBLIC、继承及对象所有权展开为授权。
	Permissions(ctx context.Context, accountName string) (interface{}, error)
	// UpdatePermissions 按数据库分别提交权限差量；未提交的数据库、架构和独立对象授权保持不变。
	UpdatePermissions(ctx context.Context, accountName string, grants []DatabasePrivileges) (interface{}, error)
	// Analyze 更新 Postgres 表或物化视图的统计信息。
	Analyze(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Optimize 整理 Postgres 表或物化视图，不终止其他会话。
	Optimize(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Truncate 清空 Postgres 业务表，保留表结构。
	Truncate(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Drop 删除 Postgres 业务表、视图或物化视图。
	Drop(ctx context.Context, schemaName, tableName string) (interface{}, error)
	// Import 逐条恢复 pg_dump plain SQL，支持 dollar quote、嵌套注释及 COPY FROM stdin。
	Import(ctx context.Context, path string, progress ...func(float64, string)) (map[string]interface{}, error)
	// Export 输出 plain SQL：结构、原生 COPY 数据、序列、视图、程序、索引和触发器。
	Export(ctx context.Context, path, schema, table string, progress ...func(float64, string)) (map[string]interface{}, error)
}
