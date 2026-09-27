package mysql

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"server/app/util/container"
)

// query 在数据库自身的容器中执行语句，密码仅从只读文件读取，不进入参数、日志或实例环境。
// 调用方持有 m.mu；只解析 stdout，容器初始化日志与客户端报错不会混入结果。
func (m *Manager) query(ctx context.Context, statement string) (string, error) {
	if err := m.available(); err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	state, err := m.readState()
	if err != nil {
		return "", err
	}
	process := "mysqld"
	// 初始化期间入口脚本会临时启动数据库。只有 PID 1 已切换为正式进程才接受管理请求。
	script := `read -r process < /proc/1/comm; [ "$process" = "$1" ] || exit 75; shift; `
	args := []string{"sh", "-c", "", "database-query", process}

	script += `MYSQL_PWD="$(cat /run/sinking-database/password)"; export MYSQL_PWD; exec mysql "$@"`
	// 使用本地 socket，调整 TCP 监听地址不会影响管理连接，root 仍只允许本机登录。
	args = append(args, "--protocol=SOCKET", "--socket=/var/run/mysqld/mysqld.sock", "--user="+state.Username, "--default-character-set=utf8mb4", "--batch", "--raw", "--skip-column-names", "--connect-timeout=3")

	args[2] = script
	options := container.ExecOptions{Command: args, Stdin: strings.NewReader(statement + "\n"), Stderr: io.Discard, Timeout: 10 * time.Second, MaxOutput: 2 * 1024 * 1024}
	if deadline, ok := ctx.Deadline(); ok {
		options.Timeout = min(options.Timeout, time.Until(deadline))
		if options.Timeout <= 0 {
			return "", context.DeadlineExceeded
		}
	}
	result, err := m.container.Exec(m.InstanceID, options)
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err != nil || result == nil {
		return "", errors.New("数据库命令执行失败，请检查容器是否已启动")
	}
	if result.TimedOut {
		return "", errors.New("数据库命令执行超时")
	}
	if result.Truncated {
		return "", errors.New("数据库查询结果超过 2 MiB 限制")
	}
	if result.ExitCode != 0 {
		return "", errors.New("数据库命令执行失败，请检查数据库状态、账户权限和参数")
	}
	return strings.TrimRight(result.Output, "\r\n"), nil
}

// Status 补充数据库就绪和负载信息；容器资源和生命周期直接使用通用管理器。
func (m *Manager) Status(ctx context.Context) (map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.statusLocked(ctx)
}

func (m *Manager) statusLocked(ctx context.Context) (map[string]interface{}, error) {
	result := map[string]interface{}{"available": m.available() == nil, "installed": false, "running": false, "ready": false, "status": container.StatusStopped, "pid": 0, "kind": "mysql", "id": m.InstanceID, "restart_required": false, "config_pending": false, "auto_start": false, "metrics": map[string]interface{}{}}
	if err := m.available(); err != nil {
		return result, nil
	}
	instance, err := m.container.Info(m.InstanceID)
	if err != nil && err.Error() != "实例不存在" {
		return nil, err
	}
	if instance == nil {
		if state, err := m.readState(); err == nil {
			result["data_preserved"], result["image_id"], result["version"], result["username"] = true, state.ImageID, state.Version, state.Username
			if content, err := m.readFile("server.conf"); err == nil {
				if values, err := m.parseConfig(string(content)); err == nil {
					result["bind"], result["port"] = values["bind"], values["port"]
				}
			}
		}
		return result, nil
	}
	result["installed"], result["running"], result["status"], result["pid"] = true, instance.Status == container.StatusRunning, instance.Status, instance.PID
	result["managed"] = m.managed(instance)
	result["auto_start"], result["image_id"] = instance.AutoRestart && !instance.ManualStop, instance.ImageID
	result["config_path"], result["data_path"] = filepath.Join(m.root, "server.conf"), filepath.Join(m.root, "data")
	if !m.managed(instance) {
		result["error"] = "同名容器已存在，请先在软件管理中处理该实例"
		return result, nil
	}
	state, err := m.readState()
	if err != nil {
		return nil, err
	}
	result["version"], result["image_version"], result["username"] = state.Version, state.Version, state.Username
	if image, err := m.container.Image(instance.ImageID); err == nil {
		result["image_name"], result["name"] = image.Name, image.Name
	}
	desired, configErr := m.readFile("server.conf")
	if configErr == nil {
		var values map[string]interface{}
		values, configErr = m.parseConfig(string(desired))
		if configErr == nil {
			result["bind"], result["port"] = values["bind"], values["port"]
		}
	}
	if configErr != nil {
		result["error"] = "待应用配置无效，请检查配置文件"
	}
	active, activeErr := m.readRunningConfig()
	pending := configErr != nil || activeErr != nil || string(active) != string(desired)
	result["restart_required"], result["config_pending"] = pending, pending
	if resources, err := m.container.Status(m.InstanceID); err == nil {
		result["resources"] = resources
	}
	if instance.Status != container.StatusRunning {
		return result, nil
	}
	// SHOW GLOBAL STATUS 在旧版及关闭 performance_schema 时仍可读取负载。
	output, queryErr := m.query(ctx, `SELECT @@version,@@max_connections,@@innodb_buffer_pool_size;
SHOW GLOBAL STATUS WHERE Variable_name IN ('Threads_connected','Threads_running','Uptime','Questions','Slow_queries','Innodb_buffer_pool_read_requests','Innodb_buffer_pool_reads');`)
	if queryErr != nil {
		result["metrics_error"] = queryErr.Error()
		return result, nil
	}
	rows := strings.Split(output, "\n")
	values := strings.Split(rows[0], "\t")
	if len(values) != 3 {
		result["metrics_error"] = "MySQL 负载结果无法解析"
		return result, nil
	}
	connections, connectionErr := strconv.ParseFloat(values[1], 64)
	buffer, bufferErr := strconv.ParseFloat(values[2], 64)
	if connectionErr != nil || bufferErr != nil {
		result["metrics_error"] = "MySQL 负载结果无法解析"
		return result, nil
	}
	metrics := map[string]interface{}{"version": values[0], "max_connections": connections, "buffer_pool_bytes": buffer}
	keys := map[string]string{"threads_connected": "connections", "threads_running": "threads_running", "uptime": "uptime", "questions": "queries", "slow_queries": "slow_queries", "innodb_buffer_pool_read_requests": "buffer_pool_read_requests", "innodb_buffer_pool_reads": "buffer_pool_reads"}
	for _, row := range rows[1:] {
		columns := strings.Split(row, "\t")
		if len(columns) != 2 {
			continue
		}
		if key, exists := keys[strings.ToLower(columns[0])]; exists {
			if value, err := strconv.ParseFloat(columns[1], 64); err == nil {
				metrics[key] = value
			}
		}
	}
	result["ready"], result["metrics"] = true, metrics
	return result, nil
}
