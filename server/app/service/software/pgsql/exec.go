package pgsql

import (
	"context"
	"encoding/csv"
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
	content, err := m.readRunningConfig()
	if err != nil {
		return "", err
	}
	values, err := m.parseConfig(string(content))
	if err != nil {
		return "", err
	}
	host, port := m.address(values)
	process := "postgres"
	// 初始化期间入口脚本会临时启动数据库。只有 PID 1 已切换为正式进程才接受管理请求。
	script := `read -r process < /proc/1/comm; [ "$process" = "$1" ] || exit 75; shift; `
	args := []string{"sh", "-c", "", "database-query", process}

	script += `PGPASSWORD="$(cat /run/sinking-database/password)"; export PGPASSWORD; PGCONNECT_TIMEOUT=3; export PGCONNECT_TIMEOUT; exec psql "$@"`
	args = append(args, "--host="+host, "--port="+port, "--username="+state.Username, "--dbname=postgres", "--no-password", "--no-psqlrc", "--quiet", "--tuples-only", "--no-align", "--set=ON_ERROR_STOP=1")

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
	return strings.TrimSuffix(result.Output, "\n"), nil
}

// Status 补充数据库就绪和负载信息；容器资源和生命周期直接使用通用管理器。
func (m *Manager) Status(ctx context.Context) (map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.statusLocked(ctx)
}

func (m *Manager) statusLocked(ctx context.Context) (map[string]interface{}, error) {
	result := map[string]interface{}{"available": m.available() == nil, "installed": false, "running": false, "ready": false, "status": container.StatusStopped, "pid": 0, "kind": "pgsql", "id": m.InstanceID, "restart_required": false, "config_pending": false, "auto_start": false, "metrics": map[string]interface{}{}}
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
		result["error"] = "同名容器已存在，请先在容器管理中处理该实例"
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
	// COPY CSV 也能用于尚无 JSON 函数的旧版，字段中的逗号、换行由 CSV 正确转义。
	version := m.versionNumber(state.Version)
	pendingQuery, activeQuery, clients := "0", "state='active'", "datname IS NOT NULL"
	if version >= 90500 {
		pendingQuery = "(SELECT count(*) FROM pg_settings WHERE pending_restart)"
	}
	if version < 90200 {
		activeQuery = "current_query NOT LIKE '<IDLE>%'"
	}
	if version >= 100000 {
		clients = "backend_type='client backend'"
	}
	statement := `COPY (SELECT current_setting('server_version'),current_setting('max_connections'),` + pendingQuery + `,
(SELECT count(*) FROM pg_stat_activity WHERE ` + clients + `),
(SELECT count(*) FROM pg_stat_activity WHERE ` + clients + ` AND ` + activeQuery + `),
extract(epoch FROM now()-pg_postmaster_start_time())::bigint,(SELECT count(*) FROM pg_database WHERE NOT datistemplate),
coalesce(sum(xact_commit),0),coalesce(sum(xact_rollback),0),coalesce(sum(blks_hit),0),coalesce(sum(blks_read),0) FROM pg_stat_database) TO STDOUT WITH CSV;`

	output, queryErr := m.query(ctx, statement)
	if queryErr != nil {
		result["metrics_error"] = queryErr.Error()
		return result, nil
	}
	metrics := make(map[string]interface{})
	rows, err := csv.NewReader(strings.NewReader(output)).ReadAll()
	keys := []string{"version", "max_connections", "pending_restart", "connections", "active_connections", "uptime", "database_count", "transactions_committed", "transactions_rolled_back", "cache_hits", "disk_reads"}
	if err != nil || len(rows) != 1 || len(rows[0]) != len(keys) {
		result["metrics_error"] = "PostgreSQL 负载结果无法解析"
		return result, nil
	}
	for i, key := range keys {
		if key == "version" {
			metrics[key] = rows[0][i]
			continue
		}
		value, err := strconv.ParseFloat(rows[0][i], 64)
		if err != nil {
			result["metrics_error"] = "PostgreSQL 负载结果无法解析"
			return result, nil
		}
		metrics[key] = value
	}
	result["ready"], result["metrics"] = true, metrics
	if count, ok := metrics["pending_restart"].(float64); ok && count > 0 {
		result["restart_required"], result["config_pending"] = true, true
	}
	return result, nil
}
