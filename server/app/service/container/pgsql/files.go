package pgsql

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"server/app/util/container"
)

const configMaxSize = 256 * 1024

// readRunningConfig 在 m.mu 锁内使用；进程变化后重新读取实际启动文件。
// 保存配置前保留本次进程的连接参数，避免尚未重启就连接到新端口。
func (m *Manager) readRunningConfig() ([]byte, error) {
	instance, err := m.container.Info(m.InstanceID)
	if err != nil {
		return nil, err
	}
	running := instance.Status == container.StatusRunning
	if running && m.runningConfig != nil && m.configGeneration == instance.Generation {
		return m.runningConfig, nil
	}
	if !running {
		m.runningConfig = nil
		m.configGeneration = 0
	} else {
		snapshot, err := m.readFile("running.snapshot")
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		if err == nil {
			header, content, ok := bytes.Cut(snapshot, []byte("\n"))
			if !ok || len(content) > configMaxSize {
				return nil, errors.New("数据库运行配置快照无效")
			}
			generation, err := strconv.ParseUint(string(header), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("数据库运行配置快照无效: %w", err)
			}
			if generation == instance.Generation {
				m.runningConfig, m.configGeneration = content, generation
				return content, nil
			}
		}
	}
	content, err := m.readFile("config/server.conf")
	if err != nil {
		return nil, err
	}
	if running {
		if err = m.writeRunningConfig(content, instance.Generation); err != nil {
			return nil, err
		}
	}
	return content, nil
}

// writeRunningConfig 保留本次进程快照，面板进程恢复后仍能使用尚未重启的连接参数。
func (m *Manager) writeRunningConfig(content []byte, generation uint64) error {
	snapshot := append([]byte(strconv.FormatUint(generation, 10)+"\n"), content...)
	if err := m.writeFile("running.snapshot", snapshot); err != nil {
		return err
	}
	m.runningConfig, m.configGeneration = content, generation
	return nil
}

// prepareDataDirectories 保留 PGDATA 的私有权限，仅允许 postgres 遍历其挂载父目录。
// PG18 的挂载根不再是 PGDATA；官方入口降权前只会 chown PGDATA，不会修复父目录。
func (m *Manager) prepareDataDirectories(image *ImageInfo) error {
	metadata, err := m.container.Image(image.ID)
	if err != nil {
		return err
	}
	pgdata := ""
	for _, env := range metadata.Env {
		if value, ok := strings.CutPrefix(env, "PGDATA="); ok {
			pgdata = value
		}
	}
	pgdata = path.Clean(pgdata)
	if !path.IsAbs(pgdata) || pgdata != image.DataPath && !strings.HasPrefix(pgdata, image.DataPath+"/") {
		return errors.New("镜像 PGDATA 不在官方持久化目录内")
	}
	parts := []string{"data"}
	if pgdata != image.DataPath {
		parts = append(parts, strings.Split(strings.TrimPrefix(pgdata, image.DataPath+"/"), "/")...)
	}
	info, err := os.Lstat(m.root)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("数据库路径必须是普通目录")
	}
	root, err := os.OpenRoot(m.root)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	opened, err := root.Stat(".")
	if err != nil {
		return err
	}
	if !os.SameFile(info, opened) {
		return errors.New("数据库路径在准备期间被替换")
	}
	for index, name := range parts {
		mode := os.FileMode(0755)
		if index == len(parts)-1 {
			mode = 0700
		}
		info, err := root.Lstat(name)
		created := errors.Is(err, os.ErrNotExist)
		if created {
			if err = root.Mkdir(name, mode); err != nil {
				return fmt.Errorf("创建数据库数据目录失败: %w", err)
			}
			info, err = root.Lstat(name)
		}
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("数据库数据路径必须是普通目录: %s", name)
		}
		next, err := root.OpenRoot(name)
		if err != nil {
			return err
		}
		opened, err := next.Stat(".")
		if err != nil || !os.SameFile(info, opened) {
			_ = next.Close()
			return errors.Join(errors.New("数据库数据路径在准备期间被替换"), err)
		}
		_ = root.Close()
		root = next
		if index < len(parts)-1 || created {
			directory, err := root.Open(".")
			if err != nil {
				return err
			}
			err = directory.Chmod(mode)
			err = errors.Join(err, directory.Close())
			if err != nil {
				return fmt.Errorf("设置数据库数据目录权限失败: %w", err)
			}
		}
	}
	return nil
}

// readFile 将配置和凭据读取限定在实例私有目录内。
func (m *Manager) readFile(name string) ([]byte, error) {
	limit := int64(configMaxSize)
	if name == "running.snapshot" {
		limit += 32 // 启动代次占用单独的首行，配置正文仍保留原大小上限。
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
		return nil, errors.New("数据库配置必须是不超过 256 KB 的普通文件")
	}
	file, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if int64(len(data)) > limit {
		return nil, errors.New("数据库配置不能超过 256 KB")
	}
	return data, err
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
	if state.Kind != "pgsql" || state.ImageID == "" || state.Username == "" {
		return nil, errors.New("数据库安装信息无效")
	}
	return &state, nil
}
