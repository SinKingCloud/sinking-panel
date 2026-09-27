package mysql

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
		source = image.ID
	}
	image, err := m.Image(source)
	if err != nil {
		return nil, err
	}
	username := strings.TrimSpace(options.Username)
	if username == "" {
		username = "root"
	}
	if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,62}$`).MatchString(username) {
		return nil, errors.New("管理员名称需以字母或下划线开头，长度不超过 63 位")
	}
	if username != "root" {
		return nil, errors.New("MySQL 初始管理员名称为 root")
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
		metadata, _ := json.Marshal(&installation{Kind: "mysql", ImageID: image.ID, Version: image.Version, Username: username})
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
	// 旧版官方入口直接拼接初始化密码，因此先用随机十六进制密码初始化，
	// 再由官方初始化钩子通过 SQL 输入设置真实密码，兼容引号和反斜杠。
	bootstrap := make([]byte, 32)
	if _, err = rand.Read(bootstrap); err != nil {
		return nil, err
	}
	if err = m.writeFile("secrets/bootstrap-password", []byte(hex.EncodeToString(bootstrap))); err != nil {
		return nil, err
	}
	password, err := m.readFile("secrets/password")
	if err != nil {
		return nil, err
	}
	quotedPassword := "'" + strings.ReplaceAll(string(password), "'", "''") + "'"
	var major, minor, patch int
	if _, err = fmt.Sscanf(image.Version, "%d.%d.%d", &major, &minor, &patch); err != nil {
		return nil, errors.New("无法识别 MySQL 镜像版本")
	}
	passwordSQL := "ALTER USER 'root'@'localhost' IDENTIFIED BY " + quotedPassword + ";"
	if major < 5 || major == 5 && (minor < 7 || minor == 7 && patch < 6) {
		passwordSQL = "SET PASSWORD FOR 'root'@'localhost' = PASSWORD(" + quotedPassword + ");"
	}
	// 更早的入口可能忽略 MYSQL_ROOT_HOST；确保管理员只保留本机访问。
	setupSQL := "SET SESSION sql_mode='NO_BACKSLASH_ESCAPES';\nDELETE FROM mysql.user WHERE User='root' AND Host<>'localhost';\n" + passwordSQL + "\nFLUSH PRIVILEGES;\n"
	if err = m.writeFile("secrets/setup.sql", []byte(setupSQL)); err != nil {
		return nil, err
	}
	if err = m.writeFile("config/init-user.sh", []byte(`MYSQL_PWD="$MYSQL_ROOT_PASSWORD" mysql --protocol=SOCKET --socket=/var/run/mysqld/mysqld.sock --default-character-set=utf8mb4 --user=root < /run/sinking-database/setup.sql
MYSQL_ROOT_PASSWORD="$(cat /run/sinking-database/password)"
export MYSQL_ROOT_PASSWORD
# 避免官方停止函数将含引号的密码直接写进客户端选项文件。
docker_temp_server_stop() {
    MYSQL_PWD="$MYSQL_ROOT_PASSWORD" mysqladmin --no-defaults --protocol=SOCKET --socket=/var/run/mysqld/mysqld.sock --user=root shutdown
}
`)); err != nil {
		return nil, err
	}
	env := []string{"MYSQL_ROOT_HOST=localhost"}

	mounts := []container.Mount{
		{Source: filepath.Join(m.root, "data"), Destination: image.DataPath},
		{Source: filepath.Join(m.root, "config"), Destination: "/etc/sinking-database", ReadOnly: true},
		{Source: filepath.Join(m.root, "secrets"), Destination: "/run/sinking-database", ReadOnly: true},
		{Source: filepath.Join(m.root, "config/init-user.sh"), Destination: "/docker-entrypoint-initdb.d/zz-sinking-user.sh", ReadOnly: true},
	}

	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if _, err = m.container.Run(container.RunOptions{ID: m.InstanceID, Name: "MySQL", ImageID: image.ID, Command: image.Command, Env: env, Mounts: mounts, AutoRestart: true, CreateOnly: true}); err != nil {
		return nil, err
	}
	if err = m.waitReady(ctx, 120*time.Second); err != nil {
		return nil, errors.Join(err, m.container.Stop(m.InstanceID))
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
			return errors.New("该实例不是受管数据库，请在软件管理中操作")
		}
		if err = m.container.Remove(m.InstanceID); err != nil {
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
