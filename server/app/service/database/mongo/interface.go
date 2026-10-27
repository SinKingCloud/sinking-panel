package mongo

import (
	"context"
	"server/app/util/page"
)

// Interface 定义数据库管理操作，调用方通过接口使用连接。
type Interface interface {
	// Aggregate 执行读取型聚合。
	Aggregate(ctx context.Context, collectionName, pipelineJSON string, limit int) (result interface{}, err error)
	// ChangePassword 修改指定认证库账户的密码，允许管理员和当前连接账户修改密码。
	ChangePassword(ctx context.Context, name, password string) (result map[string]interface{}, current bool, err error)
	// Close 释放连接和驱动监控协程。
	Close() error
	// CloseContext 让连接测试的会话清理也使用原请求的剩余时间。
	CloseContext(ctx context.Context) error
	// Collections 列举当前数据库的集合。
	Collections(ctx context.Context) (result interface{}, err error)
	// Command 执行有界命令。
	Command(ctx context.Context, commandJSON string) (result interface{}, err error)
	// Compact 压缩集合，禁止操作系统集合。
	Compact(ctx context.Context, collectionName string) (result interface{}, err error)
	// ConnectionInfo 仅提取监听和认证字段，不返回启动参数中的密钥或其他敏感信息。
	ConnectionInfo(ctx context.Context) (result map[string]interface{}, err error)
	// CreateCollection 创建集合。
	CreateCollection(ctx context.Context, collectionName string) (result interface{}, err error)
	// CreateDatabase 通过创建初始集合生成真实数据库，不接受已有数据库。
	CreateDatabase(ctx context.Context, collectionName string) (err error)
	// CreateIndex 创建索引。
	CreateIndex(ctx context.Context, collectionName, name, keysJSON string, unique, sparse bool, expireAfterSeconds *int32) (result interface{}, err error)
	// CreateUser 在当前认证数据库创建账户，一次授予多个数据库的常用角色。
	CreateUser(ctx context.Context, name, password string, grants []DatabasePrivileges) (result interface{}, err error)
	// CreateView 创建真实只读视图，源集合或视图必须位于当前数据库。
	CreateView(ctx context.Context, viewName, sourceCollection, pipelineJSON string) (result interface{}, err error)
	// DatabaseInfo 确认当前数据库存在，并读取真实集合、文档和存储统计。
	DatabaseInfo(ctx context.Context) (result map[string]interface{}, err error)
	// Databases 按名称搜索并分页返回当前账户可访问的数据库，默认隐藏系统库。
	Databases(ctx context.Context, keyword string, pageNumber, pageSize int, showSystem bool, orderByType string) (_ *page.Result[string], err error)
	// Delete 删除指定数据。
	Delete(ctx context.Context, collectionName, filterJSON string, many bool) (result interface{}, err error)
	// DeleteUser 删除当前数据库的非保护账户。
	DeleteUser(ctx context.Context, name string) (result interface{}, err error)
	// Documents 分页查询文档。
	Documents(ctx context.Context, collectionName, filterJSON, sortJSON, projectionJSON string, page, pageSize int) (result interface{}, err error)
	// DropCollection 删除集合，保留重复删除的幂等行为并保护系统集合。
	DropCollection(ctx context.Context, collectionName string) (result interface{}, err error)
	// DropDatabase 删除当前真实数据库，不存在时允许重复删除。
	DropDatabase(ctx context.Context) (err error)
	// DropDatabases 先校验整批目标，再复用当前客户端逐个删除业务数据库。
	DropDatabases(ctx context.Context, names []string) (err error)
	// DropIndex 删除指定索引。
	DropIndex(ctx context.Context, collectionName, name string) (result interface{}, err error)
	// Export 将一个集合逐文档写入标准 BSON 文件，可交给 mongorestore 恢复。
	Export(ctx context.Context, path, collectionName string, progress ...func(float64, string)) (result map[string]interface{}, err error)
	// Grant 修改账户在当前数据库内的授权。
	Grant(ctx context.Context, name string, privileges []string) (result interface{}, err error)
	// Import 逐文档读取 mongodump 的单集合 BSON 文件，不加载整个备份或自动重试写入。
	Import(ctx context.Context, path, collectionName string, progress ...func(float64, string)) (result map[string]interface{}, err error)
	// Indexes 读取集合索引。
	Indexes(ctx context.Context, collectionName string) (result interface{}, err error)
	// Insert 插入文档。
	Insert(ctx context.Context, collectionName, documentJSON string) (result interface{}, err error)
	// Permissions 按授权数据库返回可编辑角色，自定义和全局角色单独保留。
	Permissions(ctx context.Context, name string) (result interface{}, err error)
	// Revoke 修改账户在当前数据库内的授权。
	Revoke(ctx context.Context, name string, privileges []string) (result interface{}, err error)
	// Test 测试登录和读取服务器版本；无 buildInfo 权限仍可确认连接成功。
	Test(ctx context.Context) (map[string]interface{}, error)
	// Update 按非空条件修改文档。
	Update(ctx context.Context, collectionName, filterJSON, updateJSON string, many, upsert bool) (result interface{}, err error)
	// UpdatePermissions 替换各数据库的常用角色；自定义、全局角色不受表单修改影响。
	UpdatePermissions(ctx context.Context, name string, grants []DatabasePrivileges) (result interface{}, err error)
	// UpdateView 使用 collMod 修改原视图管道，保留源对象及原排序规则。
	UpdateView(ctx context.Context, viewName, pipelineJSON string) (result interface{}, err error)
	// Users 按认证数据库和账户名列举可见账户，授权数据库与认证数据库分别返回。
	Users(ctx context.Context, keyword string, page, pageSize int) (result interface{}, err error)
	// Validate 校验集合，禁止操作系统集合。
	Validate(ctx context.Context, collectionName string) (result interface{}, err error)
	// Version 返回服务器版本和功能支持情况；读取失败时保留真实错误。
	Version(ctx context.Context) (map[string]interface{}, error)
	// ViewDefinition 返回现有视图的源名称和管道，普通集合不会被误认为视图。
	ViewDefinition(ctx context.Context, viewName string) (result interface{}, err error)
}
