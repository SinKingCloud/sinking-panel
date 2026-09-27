package container

import (
	"context"

	serviceMongo "server/app/service/container/mongo"
	serviceMysql "server/app/service/container/mysql"
	servicePgsql "server/app/service/container/pgsql"
	serviceRedis "server/app/service/container/redis"
	"server/app/util/container"
	"server/app/util/page"
)

// Service 统一容器操作与软件管理入口。
type Service interface {
	Redis() *serviceRedis.Manager                                                         // 获取 Redis 管理器
	Mysql() *serviceMysql.Manager                                                         // 获取 MySQL 管理器
	Pgsql() *servicePgsql.Manager                                                         // 获取 PostgreSQL 管理器
	Mongo() *serviceMongo.Manager                                                         // 获取 MongoDB 管理器
	List(where *SelectContainer, queryPage *page.Query) (*page.Result[*Container], error) // 筛选并分页查询容器摘要，支持 id、name、create_time 排序
	Info(id string) (*container.Info, error)                                              // 获取容器详情
	Run(options container.RunOptions) (*container.Info, error)                            // 创建并启动容器
	Start(ctx context.Context, id string) (*container.Info, error)                        // 启动已有容器
	Stop(ctx context.Context, id string) error                                            // 停止容器并暂停自动恢复
	Restart(ctx context.Context, id string) error                                         // 重启容器
	Remove(ids []string) error                                                            // 批量停止并删除容器，保留挂载数据
	Status(id string) (*container.Status, error)                                          // 获取容器运行状态及 CPU、内存、进程、磁盘 I/O 等动态信息
	Exec(id string, options container.ExecOptions) (*container.ExecResult, error)         // 在容器内执行命令
	Terminal(id string, height, width int) (container.TerminalSession, error)             // 打开容器交互终端
	ReadLog(id string, after, before int64, pageSize int) (map[string]interface{}, error) // 按游标读取容器日志
	ClearLog(id string) error                                                             // 清空容器日志
	Images(where *SelectImage, queryPage *page.Query) (*page.Result[*Image], error)       // 筛选并分页查询已导入镜像，支持 id、name、create_time 排序
	Image(id string) (*ImageInfo, error)                                                  // 按 ID、名称或标签查询镜像详情
	Import(source string) (*ImageInfo, error)                                             // 导入本机镜像归档
	RemoveImage(ids []string) error                                                       // 批量删除未被容器引用的镜像
}

// Manager 统一容器操作与软件服务，运行时的初始化和关闭仍由 bootstrap 管理。
type Manager struct {
	redis   *serviceRedis.Manager
	mysql   *serviceMysql.Manager
	pgsql   *servicePgsql.Manager
	mongo   *serviceMongo.Manager
	manager *container.Manager
}

// NewManager 复用底层容器 Manager，并初始化各软件管理器。
func NewManager(manager *container.Manager) *Manager {
	return &Manager{
		redis:   serviceRedis.NewManager(manager, "redis"),
		mysql:   serviceMysql.NewManager(manager, "mysql"),
		pgsql:   servicePgsql.NewManager(manager, "pgsql"),
		mongo:   serviceMongo.NewManager(manager, "mongo"),
		manager: manager,
	}
}
