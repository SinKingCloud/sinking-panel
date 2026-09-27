package redis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"server/app/util/container"
)

// available 只检查初始化结果，容器运行状态由 Manager 负责。
func (m *Manager) available() error {
	if m.initErr != nil {
		return m.initErr
	}
	if m.container == nil {
		return errors.New("容器管理器尚未初始化")
	}
	return nil
}

// Install 接受本机归档路径或已导入镜像的 ID、名称、标签，按镜像生成配置和启动参数。
func (m *Manager) Install(ctx context.Context, source string, performance *PerformanceUpdate) (map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.available(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	instance, err := m.container.Info(m.InstanceID)
	if err != nil && err.Error() != "实例不存在" {
		return nil, err
	}
	if instance != nil && instance.Status != container.StatusStopped && instance.Status != container.StatusFailed {
		return nil, errors.New("Redis 已安装，请使用启动或重启操作")
	}
	if instance != nil {
		if err = m.container.Stop(ctx, m.InstanceID); err != nil {
			return nil, err
		}
	}
	dataPath := filepath.Join(m.root, "data")
	configPath := filepath.Join(m.root, "config")
	if entries, err := os.ReadDir(dataPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	} else if len(entries) > 0 {
		return nil, errors.New("Redis 数据目录已有文件，请使用启动或重启操作")
	}
	source = strings.TrimSpace(source)
	if filepath.IsAbs(source) {
		image, err := m.container.Import(source)
		if err != nil {
			return nil, err
		}
		source = image.Id
	}
	image, err := m.Image(source)
	if err != nil {
		return nil, err
	}
	config, err := m.patchConfig(image.Config, m.performanceValues(performance))
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	for _, directory := range []string{m.root, dataPath, configPath} {
		if info, err := os.Lstat(directory); err == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return nil, errors.New("Redis 数据路径必须是普通目录")
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		if err = os.MkdirAll(directory, 0700); err != nil {
			return nil, err
		}
	}
	// 外层私有目录保护宿主文件；容器内 redis 用户可读取只读配置。
	if err = os.Chmod(configPath, 0755); err != nil {
		return nil, err
	}
	if err = m.writeFile("redis.conf", []byte(config.content)); err != nil {
		return nil, err
	}
	if err = m.writeFile("config/redis.conf", []byte(config.content)); err != nil {
		return nil, err
	}
	// 重装可能复用运行代次，不沿用此前进程的运行配置。
	if err = m.writeRunningConfig([]byte(config.content), 0); err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", m.address(config))
	if err != nil {
		return nil, fmt.Errorf("Redis 监听地址不可用: %w", err)
	}
	_ = listener.Close()
	// 使用生成配置时的镜像 ID，避免重复读取归档或在归档变化后启动另一份镜像。
	if _, err = m.container.Run(container.RunOptions{
		ID: m.InstanceID, Name: "Redis", ImageID: image.ID,
		User: image.User, WorkingDir: image.WorkingDir, AutoRestart: true,
		Command: image.Command,
		Env:     []string{"SKIP_FIX_PERMS=1"},
		Mounts: []container.Mount{
			{Source: dataPath, Destination: image.DataPath},
			{Source: configPath, Destination: filepath.ToSlash(filepath.Dir(image.ConfigPath)), ReadOnly: true},
		},
	}); err != nil {
		return nil, err
	}
	if err = m.waitReady(ctx, config); err != nil {
		return nil, errors.Join(err, m.container.Stop(context.WithoutCancel(ctx), m.InstanceID))
	}
	return m.statusLocked(ctx)
}

// Uninstall 停止并删除 Redis 实例及其配置、持久化数据，通用镜像由容器管理器保留。
func (m *Manager) Uninstall(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.available(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	instance, err := m.container.Info(m.InstanceID)
	if err != nil {
		return err
	}
	if instance != nil {
		if err = m.container.Remove([]string{m.InstanceID}); err != nil {
			return err
		}
	}
	if err = os.RemoveAll(m.root); err != nil {
		return fmt.Errorf("删除 Redis 数据目录失败: %w", err)
	}
	return nil
}

// waitReady 进程存在不等于数据库已就绪，必须等待 Redis 加载持久化数据并响应 PING。
func (m *Manager) waitReady(ctx context.Context, config *redisConfig) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return errors.New("等待 Redis 启动超时或操作已取消，请查看实例日志")
		case <-ticker.C:
		}
		instance, err := m.container.Info(m.InstanceID)
		if err != nil {
			return err
		}
		if instance.Status == container.StatusFailed || instance.Status == container.StatusStopped {
			return errors.New("Redis 启动失败，请查看实例日志")
		}
		if instance.Status == container.StatusRunning {
			if output, err := m.command(ctx, config, "PING"); err == nil && strings.TrimSpace(output) == "PONG" {
				return nil
			}
		}
	}
}
