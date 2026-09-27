package redis

import (
	"context"
	"errors"
	"io"
	"maps"
	"net"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"server/app/util/container"
)

func (m *Manager) address(config *redisConfig) string {
	host := "127.0.0.1"
	for _, value := range strings.Fields(config.Bind) {
		optional := strings.HasPrefix(value, "-")
		value = strings.TrimPrefix(value, "-")
		ip := net.ParseIP(value)
		if ip == nil {
			continue
		}
		host = value
		if ip.IsUnspecified() {
			if ip.To4() == nil {
				host = "::1"
			} else {
				host = "127.0.0.1"
			}
		}
		if !optional {
			break
		}
	}
	return net.JoinHostPort(host, strconv.Itoa(config.Port))
}

// command 在容器内调用 redis-cli；进程创建、超时和回收由 Manager 负责。
// 认证和命令通过非交互标准输入传递，兼容不支持 REDISCLI_AUTH、-e 的旧版客户端。
func (m *Manager) command(ctx context.Context, config *redisConfig, args ...string) (string, error) {
	if err := m.available(); err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if config == nil || len(args) == 0 {
		return "", errors.New("Redis 命令参数不能为空")
	}
	host, port, err := net.SplitHostPort(m.address(config))
	if err != nil {
		return "", errors.New("Redis 连接地址无效")
	}
	options := container.ExecOptions{
		Command: []string{"redis-cli", "--raw", "-h", host, "-p", port},
		Env:     []string{"REDISCLI_HISTFILE=/dev/null"},
		// 只解析 redis-cli 的标准输出，避免容器初始化警告混入 PONG、INFO 等结果。
		Stderr:    io.Discard,
		Timeout:   10 * time.Second,
		MaxOutput: 1024 * 1024,
	}
	if deadline, ok := ctx.Deadline(); ok {
		options.Timeout = min(options.Timeout, time.Until(deadline))
		if options.Timeout <= 0 {
			return "", context.DeadlineExceeded
		}
	}
	if len(args) >= 2 && strings.EqualFold(args[0], "CONFIG") && strings.EqualFold(args[1], "SET") {
		if len(args) < 4 || len(args)%2 != 0 {
			return "", errors.New("Redis CONFIG SET 参数必须成对提供")
		}
		passwordSet := false
		for index := 2; index < len(args); index += 2 {
			if strings.EqualFold(args[index], "requirepass") {
				if passwordSet {
					return "", errors.New("Redis CONFIG SET 不能重复设置密码")
				}
				passwordSet = true
			}
		}
	}
	commands := [][]string{args}
	if config.Requirepass != "" {
		commands = append([][]string{{"AUTH", config.Requirepass}}, commands...)
	}
	var input strings.Builder
	const hex = "0123456789abcdef"
	for _, command := range commands {
		for _, value := range command {
			input.WriteByte('"')
			for index := 0; index < len(value); index++ {
				char := value[index]
				switch {
				case char == '\\' || char == '"':
					input.WriteByte('\\')
					input.WriteByte(char)
				case char < 32 || char == 127:
					input.WriteString("\\x")
					input.WriteByte(hex[char>>4])
					input.WriteByte(hex[char&15])
				default:
					input.WriteByte(char)
				}
			}
			input.WriteString("\" ")
		}
		input.WriteByte('\n')
	}
	options.Stdin = strings.NewReader(input.String())
	result, err := m.container.Exec(m.InstanceID, options)
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err != nil || result == nil {
		return "", errors.New("Redis 容器命令执行失败，请检查容器状态")
	}
	if result.TimedOut {
		return "", errors.New("Redis 命令执行超时")
	}
	if result.Truncated {
		return "", errors.New("Redis 命令输出超过 1 MiB 限制")
	}
	if result.ExitCode != 0 {
		return "", errors.New("Redis 命令执行失败，请检查连接、认证及配置参数")
	}
	// 旧版 CLI 从 stdin 读取命令时可能附带耗时行，不能将其当成服务端回复。
	lines := strings.Split(result.Output, "\n")
	lines = slices.DeleteFunc(lines, func(line string) bool {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "(") || !strings.HasSuffix(line, "s)") {
			return false
		}
		_, err := strconv.ParseFloat(line[1:len(line)-2], 64)
		return err == nil
	})
	output := strings.TrimSpace(strings.Join(lines, "\n"))
	if config.Requirepass != "" {
		auth, remaining, _ := strings.Cut(output, "\n")
		if strings.TrimSpace(auth) != "OK" {
			return "", errors.New("Redis 认证失败，请检查密码")
		}
		output = strings.TrimSpace(remaining)
	}
	// 旧版 CLI 的错误回复也会以退出码 0 返回，需要检查回复内容。
	if output == "" || strings.HasPrefix(output, "(error)") {
		return "", errors.New("Redis 命令没有返回有效结果")
	}
	for _, prefix := range []string{"ERR ", "NOAUTH ", "WRONGPASS ", "NOPERM ", "LOADING ", "BUSY ", "MISCONF ", "OOM "} {
		if strings.HasPrefix(output, prefix) {
			return "", errors.New("Redis 命令执行失败，请检查连接、认证及配置参数")
		}
	}
	return output, nil
}

func (m *Manager) Status(ctx context.Context) (map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.statusLocked(ctx)
}

func (m *Manager) statusLocked(ctx context.Context) (map[string]interface{}, error) {
	available := m.available() == nil
	result := map[string]interface{}{
		"available": available, "installed": false, "status": container.StatusStopped,
		"ready": false, "pid": 0, "auto_start": false, "restart_required": false,
	}
	if !available {
		return result, nil
	}
	// 实例状态、PID 和自动恢复策略以 Manager 为准，此处只补充 Redis 就绪和配置状态。
	instance, err := m.container.Info(m.InstanceID)
	if err != nil && err.Error() != "实例不存在" {
		return nil, err
	}
	if instance == nil {
		return result, nil
	}
	result["installed"], result["image_id"], result["auto_start"] = true, instance.ImageID, instance.AutoRestart && !instance.ManualStop
	result["status"], result["pid"] = instance.Status, instance.PID
	if image, err := m.container.Image(instance.ImageID); err == nil {
		result["image_name"] = image.Name
		for _, value := range image.Env {
			if version, ok := strings.CutPrefix(value, "REDIS_VERSION="); ok {
				result["image_version"] = version
			}
		}
	}
	result["config_path"], result["data_path"] = filepath.Join(m.root, "redis.conf"), filepath.Join(m.root, "data")
	desired, configErr := m.readConfig("redis.conf")
	if configErr == nil {
		result["bind"], result["port"] = desired.Bind, desired.Port
	} else {
		result["error"] = "待应用配置无效，请检查配置文件"
	}
	if instance.Error != "" {
		result["error"] = "容器运行异常，请查看实例日志"
	}
	if instance.Status != container.StatusRunning {
		return result, nil
	}
	active, err := m.readRunningConfig()
	if err != nil {
		result["error"] = "运行配置不可读取，请检查配置文件"
		return result, nil
	}
	result["bind"], result["port"] = active.Bind, active.Port
	result["restart_required"] = configErr != nil ||
		!slices.Equal(active.directives, desired.directives) ||
		!maps.EqualFunc(active.values, desired.values, slices.Equal[[]string]) ||
		(active.includes || desired.includes) && active.content != desired.content
	if resources, err := m.container.Status(m.InstanceID); err == nil && resources != nil {
		result["resources"] = resources
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	// 不传多个分组参数，兼容各版本的 INFO 命令。
	info, err := m.command(ctx, active, "INFO")
	if err != nil {
		result["error"] = "Redis 暂不可连接，可能仍在加载持久化数据，请查看日志"
		return result, nil
	}
	result["ready"] = true
	sections := make(map[string]map[string]string)
	section := "server"
	for _, line := range strings.Split(info, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			section = strings.ToLower(strings.TrimPrefix(line, "# "))
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if sections[section] == nil {
			sections[section] = make(map[string]string)
		}
		sections[section][key] = value
	}
	result["info"] = sections
	return result, nil
}
