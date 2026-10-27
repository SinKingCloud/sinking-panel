package redis

import (
	"context"
	"server/app/util/page"
)

// Interface 定义数据库管理操作，调用方通过接口使用连接。
type Interface interface {
	// ChangePassword 修改非保护账户的密码并持久化。
	ChangePassword(ctx context.Context, name, password string) (result interface{}, err error)
	// Close 释放本次请求的连接。
	Close() error
	// Command 执行有界命令；arguments 第一项为命令名，其余项按原样传递。
	Command(ctx context.Context, arguments []string) (result interface{}, err error)
	// ConnectionInfo 只读取网络配置，不读取 requirepass 或 ACL 密码摘要。
	ConnectionInfo(ctx context.Context) (result map[string]interface{}, err error)
	// CreateUser 一次创建账户及其数据权限，避免创建成功但授权失败。
	CreateUser(ctx context.Context, name, password string, privileges []string) (result interface{}, err error)
	// DatabaseInfo 读取当前逻辑库的键数量及过期统计，不返回其他逻辑库的数据。
	DatabaseInfo(ctx context.Context) (result map[string]interface{}, err error)
	// Databases 分页查询逻辑库；受限账号仅返回可见库和当前库。
	Databases(ctx context.Context, keyword string, pageNumber, pageSize int) (*page.Result[string], error)
	// Delete 删除指定数据。
	Delete(ctx context.Context, keys []string) (result interface{}, err error)
	// DeleteUser 删除非保护账户并持久化。
	DeleteUser(ctx context.Context, name string) (result interface{}, err error)
	// Expire 修改键的有效期（秒）；-1 永不过期，0 立即过期。
	Expire(ctx context.Context, key string, ttl int64) (result interface{}, err error)
	// Export 按键写出标准 RESP 命令流，可用 redis-cli --pipe 恢复；单键内存有界。
	Export(ctx context.Context, path, match string, key *string, overwrite bool, progress ...func(float64, string)) (result map[string]interface{}, err error)
	// Flush 仅清空当前登记逻辑库。
	Flush(ctx context.Context) (result interface{}, err error)
	// Grant 修改指定账户的读写授权。
	Grant(ctx context.Context, name string, privileges []string) (result interface{}, err error)
	// Import 流式执行 Redis RESP/AOF 命令，兼容 redis-cli --pipe 使用的二进制命令格式。
	Import(ctx context.Context, path string, progress ...func(float64, string)) (result map[string]interface{}, err error)
	// Info 读取服务器运行信息。
	Info(ctx context.Context) (result interface{}, err error)
	// Keys 分页列举键。
	Keys(ctx context.Context, match, cursorText string, pageSize int) (result interface{}, err error)
	// Permissions 设置完整读写方案，以单条 ACL 命令替换权限。
	Permissions(ctx context.Context, name string, privileges []string) (result interface{}, err error)
	// Rename 重命名键，禁止覆盖已存在的目标。
	Rename(ctx context.Context, key, newKey string) (result interface{}, err error)
	// Revoke 修改指定账户的读写授权。
	Revoke(ctx context.Context, name string, privileges []string) (result interface{}, err error)
	// Rewrite 提交后台 AOF 重写任务。
	Rewrite(ctx context.Context) (result interface{}, err error)
	// Save 原子保存键；ttl 为 nil 保留原有效期，-1 永不过期，0 立即过期（单位秒）。
	Save(ctx context.Context, key, kind, value string, ttl *int64) (result interface{}, err error)
	// Snapshot 提交后台 RDB 快照任务。
	Snapshot(ctx context.Context) (result interface{}, err error)
	// Test 验证连接，INFO 权限不足不影响 PING 成功。
	Test(ctx context.Context) (map[string]interface{}, error)
	// Users 列举实例 ACL 账户，权限不能限定到单个逻辑库。
	Users(ctx context.Context) (result interface{}, err error)
	// Value 读取键值、类型和有效期。
	Value(ctx context.Context, key, match, cursorText string, start, pageSize int) (result interface{}, err error)
	// Version 返回服务器版本及版本支持的功能，不将账户权限视为版本能力。
	Version(ctx context.Context) (map[string]interface{}, error)
}
