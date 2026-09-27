package mysql

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"server/app/util/container"
)

const configMaxSize = 256 * 1024

// readFile 将配置和凭据读取限定在实例私有目录内。
func (m *Manager) readFile(name string) ([]byte, error) {
	root, err := os.OpenRoot(m.root)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	info, err := root.Lstat(name)
	if err != nil {
		return nil, err
	}
	limit := configMaxSize
	if name == "running.snapshot" {
		limit += 32 // 运行代次和换行；配置正文仍限制为 256 KB。
	}
	if !info.Mode().IsRegular() || info.Size() > int64(limit) {
		return nil, errors.New("数据库配置必须是不超过 256 KB 的普通文件")
	}
	file, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, int64(limit)+1))
	if len(data) > limit {
		return nil, errors.New("数据库配置不能超过 256 KB")
	}
	return data, err
}

// readRunningConfig 须在 m.mu 下调用；保存已写入下次启动配置时，仍保留当前运行代次的连接配置。
// 运行快照持久化后，面板恢复已有容器时也不会提前使用尚未生效的新端口。
func (m *Manager) readRunningConfig() ([]byte, error) {
	instance, err := m.container.Info(m.InstanceID)
	if err != nil {
		return nil, err
	}
	if instance.Status != container.StatusRunning {
		m.runningConfig, m.configGeneration = nil, 0
		return m.readFile("config/server.conf")
	}
	if m.runningConfig != nil && m.configGeneration == instance.Generation {
		return m.runningConfig, nil
	}
	snapshot, err := m.readFile("running.snapshot")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if err == nil {
		header, content, found := bytes.Cut(snapshot, []byte("\n"))
		generation, parseErr := strconv.ParseUint(string(header), 10, 64)
		if !found || parseErr != nil || len(content) > configMaxSize {
			return nil, errors.New("数据库运行配置快照无效")
		}
		if generation == instance.Generation {
			m.runningConfig, m.configGeneration = content, generation
			return content, nil
		}
	}
	content, err := m.readFile("config/server.conf")
	if err != nil {
		return nil, err
	}
	if err = m.writeRunningConfig(content, instance.Generation); err != nil {
		return nil, err
	}
	return content, nil
}

// writeRunningConfig 须在 m.mu 下调用；先持久化原运行配置，再更新内存缓存。
func (m *Manager) writeRunningConfig(content []byte, generation uint64) error {
	snapshot := append([]byte(strconv.FormatUint(generation, 10)+"\n"), content...)
	if err := m.writeFile("running.snapshot", snapshot); err != nil {
		return err
	}
	m.runningConfig, m.configGeneration = content, generation
	return nil
}

// writeFile 原子替换文件；只读挂载中的文件允许数据库用户读取，宿主外层目录保持私有。
func (m *Manager) writeFile(name string, data []byte) error {
	root, err := os.OpenRoot(m.root)
	if err != nil {
		return err
	}
	defer root.Close()
	mode := os.FileMode(0600)
	if strings.HasPrefix(name, "config/") || strings.HasPrefix(name, "secrets/") {
		mode = 0644
	}
	temporary := filepath.Join(filepath.Dir(name), ".database-"+strconv.FormatInt(time.Now().UnixNano(), 36))
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

func (m *Manager) readState() (*installation, error) {
	data, err := m.readFile("installation.json")
	if err != nil {
		return nil, err
	}
	var state installation
	if err = json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	if state.Kind != "mysql" || state.ImageID == "" || state.Username == "" {
		return nil, errors.New("数据库安装信息无效")
	}
	return &state, nil
}
