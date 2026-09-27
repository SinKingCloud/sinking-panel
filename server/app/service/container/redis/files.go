package redis

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"server/app/util/container"
)

// readFile 限定在 Redis 管理目录内读取普通小文件。
func (m *Manager) readFile(name string) ([]byte, error) {
	limit := int64(redisConfigMaxSize)
	if name == "running.snapshot" {
		limit += 32 // 首行保存运行代次，其余为原始配置。
	}
	root, err := os.OpenRoot(m.root)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	info, err := root.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, errors.New("Redis 配置必须是不超过 256 KB 的普通文件")
	}
	file, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if int64(len(data)) > limit {
		return nil, errors.New("Redis 配置不能超过 256 KB")
	}
	return data, err
}

// writeFile 原子替换配置；启动配置置于只读挂载内，宿主父目录仅面板可访问。
func (m *Manager) writeFile(name string, data []byte) error {
	root, err := os.OpenRoot(m.root)
	if err != nil {
		return err
	}
	defer root.Close()
	directory := filepath.Dir(name)
	temporary := filepath.Join(directory, ".redis-"+strconv.FormatInt(time.Now().UnixNano(), 36))
	mode := os.FileMode(0600)
	if name == "config/redis.conf" {
		mode = 0644
	}
	file, err := root.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	defer root.Remove(temporary)
	err = file.Chmod(mode)
	if err == nil {
		_, err = file.Write(data)
	}
	if err == nil {
		err = file.Sync()
	}
	err = errors.Join(err, file.Close())
	if err != nil {
		return err
	}
	return root.Rename(temporary, name)
}

func (m *Manager) readConfig(name string) (*redisConfig, error) {
	data, err := m.readFile(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("Redis 尚未安装或配置文件不存在: %w", err)
		}
		return nil, err
	}
	return m.parseConfig(string(data))
}

// readRunningConfig 在配置锁内读取本次运行的快照，容器重启后自动使用新的启动配置。
func (m *Manager) readRunningConfig() (*redisConfig, error) {
	instance, err := m.container.Info(m.InstanceID)
	if err != nil {
		return nil, err
	}
	if instance.Status != container.StatusRunning {
		m.runningConfig, m.configGeneration = nil, 0
		return m.readConfig("config/redis.conf")
	}
	if m.runningConfig != nil && m.configGeneration == instance.Generation {
		return m.parseConfig(string(m.runningConfig))
	}
	snapshot, err := m.readFile("running.snapshot")
	if err == nil {
		generation, content, ok := bytes.Cut(snapshot, []byte("\n"))
		value, parseErr := strconv.ParseUint(string(generation), 10, 64)
		if !ok || parseErr != nil || len(content) > redisConfigMaxSize {
			return nil, errors.New("Redis 运行配置快照无效")
		}
		if value == instance.Generation {
			m.runningConfig, m.configGeneration = content, value
			return m.parseConfig(string(content))
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	config, err := m.readConfig("config/redis.conf")
	if err != nil {
		return nil, err
	}
	if err = m.writeRunningConfig([]byte(config.content), instance.Generation); err != nil {
		return nil, err
	}
	return config, nil
}

// writeRunningConfig 保留运行代次，面板恢复仍在运行的容器时也能取得原连接配置。
func (m *Manager) writeRunningConfig(data []byte, generation uint64) error {
	snapshot := append([]byte(strconv.FormatUint(generation, 10)+"\n"), data...)
	if err := m.writeFile("running.snapshot", snapshot); err != nil {
		return err
	}
	m.runningConfig, m.configGeneration = data, generation
	return nil
}
