package container

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"server/app/enum/container_status"
	serviceMongo "server/app/service/container/mongo"
	serviceMysql "server/app/service/container/mysql"
	servicePgsql "server/app/service/container/pgsql"
	serviceRedis "server/app/service/container/redis"
	"server/app/util/container"
	"server/app/util/page"
	"server/app/util/str"
)

// Redis 返回当前服务管理的 Redis 实例。
func (m *Manager) Redis() *serviceRedis.Manager {
	return m.redis
}

// Mysql 返回当前服务管理的 MySQL 实例。
func (m *Manager) Mysql() *serviceMysql.Manager {
	return m.mysql
}

// Pgsql 返回当前服务管理的 PostgreSQL 实例。
func (m *Manager) Pgsql() *servicePgsql.Manager {
	return m.pgsql
}

// Mongo 返回当前服务管理的 MongoDB 实例。
func (m *Manager) Mongo() *serviceMongo.Manager {
	return m.mongo
}

// available 判断是否初始化
func (m *Manager) available() error {
	if m == nil || m.manager == nil {
		return errors.New("容器管理器尚未初始化")
	}
	return nil
}

// List 同步容器运行状态，按条件筛选并分页返回摘要，默认按创建时间倒序。
func (m *Manager) List(where *SelectContainer, queryPage *page.Query) (*page.Result[*Container], error) {
	if where != nil && where.Status != nil {
		if _, ok := container_status.Map()[*where.Status]; !ok {
			return nil, errors.New("运行状态无效")
		}
	}
	if err := m.available(); err != nil {
		return nil, err
	}
	list := m.manager.List()
	var keyword, name, imageName string
	if where != nil {
		if where.Keyword != nil {
			keyword = strings.ToLower(*where.Keyword)
		}
		if where.Name != nil {
			name = strings.ToLower(*where.Name)
		}
		if where.ImageName != nil {
			imageName = strings.ToLower(*where.ImageName)
		}
	}
	total, err := m.selectPage(queryPage, &list, func(index int) bool {
		if where == nil {
			return true
		}
		item := list[index]
		if where.ID != nil && item.ID != *where.ID ||
			where.ImageID != nil && item.ImageID != *where.ImageID ||
			where.Status != nil && item.Status != *where.Status ||
			where.AutoRestart != nil && item.AutoRestart != *where.AutoRestart ||
			where.ManualStop != nil && item.ManualStop != *where.ManualStop ||
			where.CreateTimeStart != nil && item.CreatedAt < *where.CreateTimeStart ||
			where.CreateTimeEnd != nil && item.CreatedAt > *where.CreateTimeEnd {
			return false
		}
		if !strings.Contains(strings.ToLower(item.Name), name) ||
			!strings.Contains(strings.ToLower(item.ImageName), imageName) {
			return false
		}
		return keyword == "" ||
			strings.Contains(strings.ToLower(item.ID), keyword) ||
			strings.Contains(strings.ToLower(item.Name), keyword) ||
			strings.Contains(strings.ToLower(item.ImageID), keyword) ||
			strings.Contains(strings.ToLower(item.ImageName), keyword)
	}, func(index int) (string, string, int64) {
		item := list[index]
		return item.ID, item.Name, item.CreatedAt
	})
	if err != nil {
		return nil, err
	}
	data := make([]*Container, len(list))
	for index, item := range list {
		data[index] = &Container{
			Id: item.ID, Name: item.Name, ImageId: item.ImageID, ImageName: item.ImageName,
			Status: item.Status, CreateTime: str.DateTime(time.Unix(item.CreatedAt, 0)),
			AutoRestart: item.AutoRestart, ManualStop: item.ManualStop,
		}
	}
	return page.New(queryPage, total, data), nil
}

// Info 返回容器的最新状态、启动参数和挂载配置。
func (m *Manager) Info(id string) (*container.Info, error) {
	if err := m.available(); err != nil {
		return nil, err
	}
	return m.manager.Info(id)
}

// Run 解析镜像并创建、启动新容器，不覆盖已有实例。
func (m *Manager) Run(options container.RunOptions) (*container.Info, error) {
	if err := m.available(); err != nil {
		return nil, err
	}
	options.CreateOnly = true
	return m.manager.Run(options)
}

// Start 按实例保存的运行参数启动容器。
func (m *Manager) Start(ctx context.Context, id string) (*container.Info, error) {
	if err := m.available(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return m.manager.Start(id)
}

// Stop 停止容器，同时暂停该容器的自动恢复。
func (m *Manager) Stop(ctx context.Context, id string) error {
	if err := m.available(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return m.manager.Stop(id)
}

// Restart 按实例保存的运行参数重启容器。
func (m *Manager) Restart(ctx context.Context, id string) error {
	if err := m.available(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err := m.manager.Restart(id)
	return err
}

// Remove 批量删除容器的写层、日志和运行状态，保留挂载数据并汇总失败原因。
func (m *Manager) Remove(ids []string) error {
	if err := m.available(); err != nil {
		return err
	}
	var errs []error
	for _, id := range ids {
		if err := m.manager.Remove(id); err != nil {
			errs = append(errs, fmt.Errorf("容器[%s]：%w", id, err))
		}
	}
	return errors.Join(errs...)
}

// Status 返回容器运行状态及 CPU、内存、进程、磁盘 I/O 等动态信息。
func (m *Manager) Status(id string) (*container.Status, error) {
	if err := m.available(); err != nil {
		return nil, err
	}
	return m.manager.Status(id)
}

// Exec 在容器内执行命令，沿用 Manager 的超时和输出大小限制。
func (m *Manager) Exec(id string, options container.ExecOptions) (*container.ExecResult, error) {
	if err := m.available(); err != nil {
		return nil, err
	}
	return m.manager.Exec(id, options)
}

// Terminal 打开交互终端，关闭终端不会停止容器主进程。
func (m *Manager) Terminal(id string, height, width int) (container.TerminalSession, error) {
	if err := m.available(); err != nil {
		return nil, err
	}
	return m.manager.Terminal(id, height, width)
}

// ReadLog 按字节游标读取容器日志。
func (m *Manager) ReadLog(id string, after, before int64, pageSize int) (map[string]interface{}, error) {
	if err := m.available(); err != nil {
		return nil, err
	}
	return m.manager.ReadLog(id, after, before, pageSize)
}

// ClearLog 清空容器日志。
func (m *Manager) ClearLog(id string) error {
	if err := m.available(); err != nil {
		return err
	}
	return m.manager.ClearLog(id)
}

// Images 按条件筛选并分页返回已导入镜像，默认按创建时间倒序。
func (m *Manager) Images(where *SelectImage, queryPage *page.Query) (*page.Result[*Image], error) {
	if err := m.available(); err != nil {
		return nil, err
	}
	list := m.manager.Images()
	var keyword, name string
	if where != nil {
		if where.Keyword != nil {
			keyword = strings.ToLower(*where.Keyword)
		}
		if where.Name != nil {
			name = strings.ToLower(*where.Name)
		}
	}
	total, err := m.selectPage(queryPage, &list, func(index int) bool {
		if where == nil {
			return true
		}
		item := list[index]
		if where.ID != nil && item.ID != *where.ID ||
			where.Tag != nil && !slices.Contains(item.Tags, *where.Tag) ||
			where.CreateTimeStart != nil && item.CreatedAt < *where.CreateTimeStart ||
			where.CreateTimeEnd != nil && item.CreatedAt > *where.CreateTimeEnd {
			return false
		}
		if !strings.Contains(strings.ToLower(item.Name), name) {
			return false
		}
		return keyword == "" || strings.Contains(strings.ToLower(item.ID), keyword) ||
			strings.Contains(strings.ToLower(item.Name), keyword) ||
			slices.ContainsFunc(item.Tags, func(tag string) bool {
				return strings.Contains(strings.ToLower(tag), keyword)
			})
	}, func(index int) (string, string, int64) {
		item := list[index]
		return item.ID, item.Name, item.CreatedAt
	})
	if err != nil {
		return nil, err
	}
	data := make([]*Image, len(list))
	for index, item := range list {
		data[index] = &Image{
			Id: item.ID, Name: item.Name, Tags: item.Tags,
			CreateTime: str.DateTime(time.Unix(item.CreatedAt, 0)),
		}
	}
	return page.New(queryPage, total, data), nil
}

// Image 按 ID、名称或标签查询已导入的镜像。
func (m *Manager) Image(id string) (*ImageInfo, error) {
	if err := m.available(); err != nil {
		return nil, err
	}
	item, err := m.manager.Image(id)
	if err != nil {
		return nil, err
	}
	return &ImageInfo{
		Image: Image{
			Id: item.ID, Name: item.Name, Tags: item.Tags,
			CreateTime: str.DateTime(time.Unix(item.CreatedAt, 0)),
		},
		User:       item.User,
		Entrypoint: item.Entrypoint, Command: item.Command, Env: item.Env,
		WorkingDir: item.WorkingDir, Labels: item.Labels,
		Volumes: item.Volumes, ExposedPorts: item.ExposedPorts,
	}, nil
}

// Import 导入本机镜像归档并返回镜像信息。
func (m *Manager) Import(source string) (*ImageInfo, error) {
	if err := m.available(); err != nil {
		return nil, err
	}
	item, err := m.manager.Import(source)
	if err != nil {
		return nil, err
	}
	return m.Image(item.ID)
}

// RemoveImage 逐个删除镜像，汇总失败原因，不影响其余镜像的删除。
func (m *Manager) RemoveImage(ids []string) error {
	if err := m.available(); err != nil {
		return err
	}
	var errs []error
	for _, id := range ids {
		if err := m.manager.RemoveImage(id); err != nil {
			errs = append(errs, fmt.Errorf("镜像[%s]：%w", id, err))
		}
	}
	return errors.Join(errs...)
}
