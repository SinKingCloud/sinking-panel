package pgsql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"server/app/util/container"
)

// Install 保留官方初始化过程，数据、配置和密码分别挂载，密码不写入容器 Env。
// 已保留的数据只允许使用原镜像和原账户重新安装，避免隐式跨版本迁移或修改认证信息。
func (m *Manager) Install(ctx context.Context, source string, options *InstallOptions) (map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.available(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if options == nil || len(options.Password) > 4096 || strings.ContainsAny(options.Password, "\x00\r\n") {
		return nil, errors.New("请填写管理员密码，密码不能包含换行或空字符")
	}
	instance, err := m.container.Info(m.InstanceID)
	if err != nil && err.Error() != "实例不存在" {
		return nil, err
	}
	if instance != nil {
		return nil, errors.New("数据库已安装，请使用启动、重启或卸载操作")
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
	username := strings.TrimSpace(options.Username)
	if username == "" {
		username = "postgres"
	}
	if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,62}$`).MatchString(username) {
		return nil, errors.New("管理员名称需以字母或下划线开头，长度不超过 63 位")
	}

	state, stateErr := m.readState()
	reinstall := stateErr == nil
	if stateErr != nil && !errors.Is(stateErr, os.ErrNotExist) {
		return nil, stateErr
	}
	content := image.Config
	if reinstall {
		if state.ImageID != image.ID {
			return nil, errors.New("已有数据仅支持使用原镜像重新安装，升级请先备份并迁移")
		}
		password, err := m.readFile("secrets/password")
		if err != nil {
			return nil, err
		}
		if options.Username == "" {
			username = state.Username
		}
		if state.Username != username || options.Password != "" && string(password) != options.Password {
			return nil, errors.New("已有数据需使用原管理员名称和密码，不会重新初始化账户")
		}
		previous, err := m.readFile("server.conf")
		if err != nil {
			return nil, err
		}
		content = string(previous)
	} else {
		if options.Password == "" {
			return nil, errors.New("首次安装必须设置管理员密码")
		}
		for _, name := range []string{"data"} {
			entries, err := os.ReadDir(filepath.Join(m.root, name))
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
			if len(entries) > 0 {
				return nil, errors.New("数据目录已有文件但缺少安装信息，请先备份并检查目录")
			}
		}
	}
	values := make(map[string]interface{}, len(options.Config)+2)
	for key, value := range options.Config {
		values[key] = value
	}
	if options.Bind != "" {
		values["bind"] = options.Bind
	}
	if options.Port != 0 {
		values["port"] = options.Port
	}
	content, err = m.patchConfig(content, values)
	if err != nil {
		return nil, err
	}
	content, err = m.timezoneConfig(content)
	if err != nil {
		return nil, err
	}
	config, err := m.parseConfig(content)
	if err != nil {
		return nil, err
	}
	if err = m.checkAddress(config); err != nil {
		return nil, err
	}
	for _, name := range []string{"", "data", "config", "secrets"} {
		directory := filepath.Join(m.root, name)
		if info, err := os.Lstat(directory); err == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return nil, errors.New("数据库路径必须是普通目录")
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		mode := os.FileMode(0700)
		if name == "config" || name == "secrets" {
			mode = 0755
		}
		if err = os.MkdirAll(directory, mode); err != nil {
			return nil, err
		}
		if name == "config" || name == "secrets" {
			if err = os.Chmod(directory, mode); err != nil {
				return nil, err
			}
		}
	}
	if !reinstall {
		if err = m.writeFile("secrets/password", []byte(options.Password)); err != nil {
			return nil, err
		}
		metadata, _ := json.Marshal(&installation{Kind: "pgsql", ImageID: image.ID, Version: image.Version, Username: username})
		if err = m.writeFile("installation.json", metadata); err != nil {
			return nil, err
		}
	}
	if err = m.writeFile("server.conf", []byte(content)); err != nil {
		return nil, err
	}
	if err = m.writeFile("config/server.conf", []byte(content)); err != nil {
		return nil, err
	}
	// 重装可能复用运行代次，安装配置写入后清除旧进程的快照标识。
	if err = m.writeRunningConfig([]byte(content), 0); err != nil {
		return nil, err
	}
	authentication := "scram-sha-256"
	if m.versionNumber(image.Version) < 100000 {
		authentication = "md5"
		// 早期官方入口不识别 PASSWORD_FILE，启动时从挂载文件读取，再交回官方初始化。
		// 密码不写进实例 Env 或命令参数，新旧镜像都保留原有 ENTRYPOINT。
		image.Command = append([]string{"sh", "-ec", `POSTGRES_PASSWORD="$(cat /run/sinking-database/password)"; PGPASSWORD="$POSTGRES_PASSWORD"; export POSTGRES_PASSWORD PGPASSWORD; unset POSTGRES_PASSWORD_FILE; exec "$@"`, "postgres-init", image.Entrypoint[0]}, image.Command...)
	}
	// --auth 同时设置本地和 TCP 认证，也兼容尚无 --auth-local/--auth-host 的 9.0、9.1。
	initdbArgs := "--auth=" + authentication
	if m.versionNumber(image.Version) < 90300 {
		initdbArgs += " --pwfile=/run/sinking-database/password"
	}
	env := []string{"POSTGRES_USER=" + username, "POSTGRES_PASSWORD_FILE=/run/sinking-database/password", "POSTGRES_INITDB_ARGS=" + initdbArgs, "POSTGRES_HOST_AUTH_METHOD=" + authentication}

	mounts := []container.Mount{
		{Source: filepath.Join(m.root, "data"), Destination: image.DataPath},
		{Source: filepath.Join(m.root, "config"), Destination: "/etc/sinking-database", ReadOnly: true},
		{Source: filepath.Join(m.root, "secrets"), Destination: "/run/sinking-database", ReadOnly: true},
	}

	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if err = m.prepareDataDirectories(image); err != nil {
		return nil, err
	}
	if _, err = m.container.Run(container.RunOptions{ID: m.InstanceID, Name: "PostgreSQL", ImageID: image.ID, Command: image.Command, Env: env, Mounts: mounts, AutoRestart: true, CreateOnly: true}); err != nil {
		return nil, err
	}
	if err = m.waitReady(ctx, 120*time.Second); err != nil {
		return nil, errors.Join(err, m.container.Stop(context.WithoutCancel(ctx), m.InstanceID))
	}
	return m.statusLocked(ctx)
}

// Uninstall 只删除实例；removeData 为 false 时保留原配置、凭据和数据供同镜像重装。
func (m *Manager) Uninstall(ctx context.Context, removeData bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.available(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	instance, err := m.container.Info(m.InstanceID)
	if err != nil && err.Error() != "实例不存在" {
		return err
	}
	if instance != nil {
		if !m.managed(instance) {
			return errors.New("该实例不是受管数据库，请在容器管理中操作")
		}
		if err = m.container.Remove([]string{m.InstanceID}); err != nil {
			return err
		}
	}
	if removeData {
		return os.RemoveAll(m.root)
	}
	return nil
}

func (m *Manager) managed(instance *container.Info) bool {
	for _, mount := range instance.Mounts {
		if filepath.Clean(mount.Source) == filepath.Join(m.root, "config") && mount.Destination == "/etc/sinking-database" {
			return true
		}
	}
	return false
}

func (m *Manager) address(values map[string]interface{}) (string, string) {
	host := strings.TrimSpace(strings.Split(fmt.Sprint(values["bind"]), ",")[0])
	if host == "*" || host == "0.0.0.0" || host == "localhost" {
		host = "127.0.0.1"
	}
	if host == "::" {
		host = "::1"
	}
	port := strconv.Itoa(int(values["port"].(float64)))
	return host, port
}

func (m *Manager) checkAddress(values map[string]interface{}) error {
	host := strings.TrimSpace(strings.Split(fmt.Sprint(values["bind"]), ",")[0])
	if host == "*" {
		host = "0.0.0.0"
	}
	_, port := m.address(values)
	listener, err := net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return fmt.Errorf("数据库监听地址不可用: %w", err)
	}
	return listener.Close()
}

func (m *Manager) waitReady(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()
	statement := "SELECT 1;"

	for {
		select {
		case <-ctx.Done():
			return errors.New("等待数据库启动超时或操作已取消，数据已保留，请查看容器日志")
		case <-ticker.C:
		}
		instance, err := m.container.Info(m.InstanceID)
		if err != nil {
			return err
		}
		if instance.Status == container.StatusStopped || instance.Status == container.StatusFailed {
			return errors.New("数据库进程已退出，请查看容器日志")
		}
		if _, err = m.query(ctx, statement); err == nil {
			return nil
		}
	}
}
