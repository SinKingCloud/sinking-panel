package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"path/filepath"
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
	process := "mongod"
	// 初始化期间入口脚本会临时启动数据库。只有 PID 1 已切换为正式进程才接受管理请求。
	script := `read -r process < /proc/1/comm; [ "$process" = "$1" ] || exit 75; shift; `
	args := []string{"sh", "-c", "", "database-query", process}

	// stdin 管道由宿主创建，映射后的用户不能通过 /dev/stdin 重新打开它。
	// 两种客户端统一读取容器内创建的脚本；打开后立即删除，超时退出也不遗留查询内容。
	script += `umask 077; query=$(mktemp /tmp/sinking-mongo-query.XXXXXX) || exit 1; trap 'rm -f "$query"' EXIT HUP INT TERM; cat > "$query" || exit 1; exec 3< "$query"; rm -f "$query"; if command -v mongo >/dev/null 2>&1; then exec mongo --norc --nodb --quiet /proc/self/fd/3; else exec mongosh --norc --nodb --quiet --file /proc/self/fd/3; fi`
	address, _ := json.Marshal("mongodb://" + net.JoinHostPort(host, port) + "/?directConnection=true&serverSelectionTimeoutMS=3000&connectTimeoutMS=3000")
	legacyAddress, _ := json.Marshal(net.JoinHostPort(host, port))
	username, _ := json.Marshal(state.Username)
	statement = "var panelLegacy = typeof cat === 'function'; var panelConnection = new Mongo(panelLegacy ? " + string(legacyAddress) + " : " + string(address) + "); db = panelConnection.getDB('admin'); var panelPassword = panelLegacy ? cat('/run/sinking-database/password') : require('fs').readFileSync('/run/sinking-database/password', 'utf8'); var panelAuth = db.auth(" + string(username) + ", panelPassword); if (!(panelAuth === 1 || panelAuth && panelAuth.ok === 1)) { throw new Error('认证失败'); }\n" + statement

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
	return strings.TrimSpace(result.Output), nil
}

// Status 补充数据库就绪和负载信息；容器资源和生命周期直接使用通用管理器。
func (m *Manager) Status(ctx context.Context) (map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.statusLocked(ctx)
}

func (m *Manager) statusLocked(ctx context.Context) (map[string]interface{}, error) {
	result := map[string]interface{}{"available": m.available() == nil, "installed": false, "running": false, "ready": false, "status": container.StatusStopped, "pid": 0, "kind": "mongo", "id": m.InstanceID, "restart_required": false, "config_pending": false, "auto_start": false, "metrics": map[string]interface{}{}}
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
	statement := `var s = db.runCommand({serverStatus: 1}); if (s.ok !== 1) { throw new Error('无法读取负载'); } var c=s.connections||{}, mem=s.mem||{}, op=s.opcounters||{}, network=s.network||{}; print(JSON.stringify({version:s.version, uptime:s.uptime, storage_engine:s.storageEngine?s.storageEngine.name:"mmapv1", connections:c.current, max_connections:c.current+c.available, resident_memory_mb:mem.resident, virtual_memory_mb:mem.virtual, operations_insert:op.insert, operations_query:op.query, operations_update:op.update, operations_delete:op.delete, bytes_in:Number(network.bytesIn||0), bytes_out:Number(network.bytesOut||0)}));`

	output, queryErr := m.query(ctx, statement)
	if queryErr != nil {
		result["metrics_error"] = queryErr.Error()
		return result, nil
	}
	metrics := make(map[string]interface{})
	if err = json.Unmarshal([]byte(output), &metrics); err != nil {
		result["metrics_error"] = "MongoDB 负载结果无法解析"
		return result, nil
	}
	result["ready"], result["metrics"] = true, metrics
	return result, nil
}
