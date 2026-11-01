package container

import (
	"context"

	"server/app/util/container"
	"server/app/util/page"
)

// Service 统一容器和镜像的管理操作。
type Service interface {
	VolumePath(id string) (string, error)                                                 // 获取实例数据目录
	List(where *SelectContainer, queryPage *page.Query) (*page.Result[*Container], error) // 筛选并分页查询容器摘要，支持 id、name、create_time 排序
	Info(id string) (*container.Info, error)                                              // 获取容器详情
	Run(options container.RunOptions) (*container.Info, error)                            // 创建或恢复容器
	Start(ctx context.Context, id string) (*container.Info, error)                        // 启动已有容器
	Stop(ctx context.Context, id string) error                                            // 停止容器并暂停自动恢复
	Restart(ctx context.Context, id string) error                                         // 重启容器
	Remove(ids []string) error                                                            // 批量删除容器，保留挂载数据
	Status(id string) (*container.Status, error)                                          // 获取容器资源状态
	Exec(id string, options container.ExecOptions) (*container.ExecResult, error)         // 执行容器命令
	Terminal(id string, height, width int) (container.TerminalSession, error)             // 打开容器交互终端
	ReadLog(id string, after, before int64, pageSize int) (map[string]interface{}, error) // 按游标读取容器日志
	ClearLog(id string) error                                                             // 清空容器日志
	Images(where *SelectImage, queryPage *page.Query) (*page.Result[*Image], error)       // 筛选并分页查询已导入镜像，支持 id、name、create_time 排序
	Image(id string) (*ImageInfo, error)                                                  // 获取镜像详情
	Import(source string) (*ImageInfo, error)                                             // 导入本机镜像归档
	RemoveImage(ids []string) error                                                       // 批量删除镜像
}

// Manager 管理容器和镜像，运行时的初始化和关闭由 bootstrap 管理。
type Manager struct {
	manager *container.Manager
}

// NewManager 复用底层容器 Manager。
func NewManager(manager *container.Manager) *Manager {
	return &Manager{
		manager: manager,
	}
}
