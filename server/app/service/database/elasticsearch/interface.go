package elasticsearch

import (
	"context"
	"server/app/util/page"
)

// Interface 定义数据库管理操作，调用方通过接口使用连接。
type Interface interface {
	// Analyzers 返回当前索引可选的自定义、插件及内置分析器。
	Analyzers(ctx context.Context) ([]map[string]interface{}, error)
	// Close 释放本次操作创建的 HTTP 连接。
	Close() error
	// CloseSearch 仅释放指定分页会话，不允许关闭其他查询或使用 _all。
	CloseSearch(ctx context.Context, pagination, searchID string) error
	// CreateIndex 创建单个普通索引，可选 Settings 与 Mapping 保持原生 JSON 表达。
	CreateIndex(ctx context.Context, name, settings, mapping string) (map[string]interface{}, error)
	// Delete 删除一个具体文档，不提供批量匹配删除入口。
	Delete(ctx context.Context, documentID, documentType, routing string) (map[string]interface{}, error)
	// Documents 普通分页保留 from/size；深分页使用 PIT 快照，旧版自动使用 scroll。
	Documents(ctx context.Context, query, pagination, searchID, cursor string, page, pageSize int) (*DocumentPage, error)
	// DropIndex 删除一个普通索引，不接受通配符、别名列表或系统索引。
	DropIndex(ctx context.Context, name string) (map[string]interface{}, error)
	// DropIndices 先校验整批索引名称，再逐个删除并保留部分成功信息。
	DropIndices(ctx context.Context, names []string) error
	// Indices 返回普通索引及统计，系统索引保持只读并不加入工作区目录。
	Indices(ctx context.Context, keyword string, pageNumber, pageSize int, orderByType string) (*page.Result[Index], error)
	// Info 返回选中具体索引的 Mapping、Settings 和可读统计。
	Info(ctx context.Context) (map[string]interface{}, error)
	// Insert 添加文档，指定 ID 时使用 op_type=create，防止覆盖已有文档。
	Insert(ctx context.Context, documentID, document, documentType, routing string) (map[string]interface{}, error)
	// Mapping 读取实际索引映射，保留旧版文档类型层级。
	Mapping(ctx context.Context) (map[string]interface{}, error)
	// RefreshFields 使用现有文档按当前映射重新索引，使新增的多字段生效。
	RefreshFields(ctx context.Context, progress func(float64, string)) (map[string]interface{}, error)
	// Request 执行显式 REST 请求；写请求只作用于当前普通索引，不开放跨索引批量操作。
	Request(ctx context.Context, method, path, body string) (map[string]interface{}, error)
	// Settings 返回索引设置；不读取集群认证与密钥配置。
	Settings(ctx context.Context) (map[string]interface{}, error)
	// Status 并行读取集群健康和资源统计；旧版或无统计权限时仍返回健康信息。
	Status(ctx context.Context) (map[string]interface{}, error)
	// Test 返回连接状态和服务器版本，认证信息不会出现在结果中。
	Test(ctx context.Context) (map[string]interface{}, error)
	// Update 完整替换已存在的 _source；使用读取到的版本防止并发覆盖或无意创建文档。
	Update(ctx context.Context, documentID, document, documentType, routing, seqNo, primaryTerm, version string) (map[string]interface{}, error)
	// UpdateMapping 添加映射字段，字段类型等不可修改项由服务器按版本校验。
	UpdateMapping(ctx context.Context, mapping string) (map[string]interface{}, error)
	// UpdateSettings 修改当前普通索引的动态设置，不自动关闭索引或变更集群设置。
	UpdateSettings(ctx context.Context, settings string) (map[string]interface{}, error)
	// Version 通过 REST 功能适配有类型和无类型文档版本。
	Version(ctx context.Context) (map[string]interface{}, error)
}
