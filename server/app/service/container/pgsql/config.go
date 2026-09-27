package pgsql

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	"server/app/util/container"
)

func (m *Manager) fields() []map[string]interface{} {
	fields := []map[string]interface{}{
		{"name": "bind", "label": "绑定地址", "type": "text", "tooltip": "默认只允许本机连接；对外监听前请配置防火墙和数据库访问权限。"},
		{"name": "port", "label": "监听端口", "type": "number", "min": 1, "max": 65535},
	}

	return append(fields, m.performanceFields()...)
}

// configKey 将表单名映射到 PostgreSQL 的原生配置项。
func (m *Manager) configKey(name string) string {
	if name == "bind" {
		return "listen_addresses"
	}

	return name
}

// parseConfig 只提取可视化管理项；其他原文始终保留并交由数据库验证。
func (m *Manager) parseConfig(content string) (map[string]interface{}, error) {
	if len(content) == 0 || len(content) > configMaxSize || strings.ContainsRune(content, 0) {
		return nil, errors.New("配置不能为空、包含空字符或超过 256 KB")
	}
	raw := make(map[string]string)

	for number, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "!") || strings.HasPrefix(line, "include") {
			return nil, errors.New("请将配置直接写入当前文件，不支持引用其他配置文件")
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				return nil, fmt.Errorf("配置第 %d 行缺少参数值", number+1)
			}
			key, value = parts[0], strings.TrimSpace(strings.TrimPrefix(line, parts[0]))
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		// 仅处理引号外的行尾注释，路径或密码内的 # 不受影响。
		quoted := byte(0)
		for i := 0; i < len(value); i++ {
			if value[i] == '\\' && quoted != 0 {
				i++
				continue
			}
			if value[i] == '\'' || value[i] == '"' {
				if quoted == 0 {
					quoted = value[i]
				} else if quoted == value[i] {
					quoted = 0
				}
			} else if value[i] == '#' && quoted == 0 {
				value = strings.TrimSpace(value[:i])
				break
			}
		}
		value = strings.Trim(value, "'\"")

		raw[key] = value
	}

	if raw["data_directory"] != "" || raw["hba_file"] != "" || raw["ident_file"] != "" {
		return nil, errors.New("数据目录和认证文件由官方镜像管理，请勿覆盖 data_directory、hba_file 或 ident_file")
	}

	values := make(map[string]interface{})
	for _, field := range m.fields() {
		name := field["name"].(string)
		key := m.configKey(name)

		value, exists := raw[key]
		if !exists {
			continue
		}
		switch field["type"] {
		case "number":
			number, err := m.configNumber(name, value, field["unit"] == "MB")
			if err != nil {
				return nil, fmt.Errorf("配置 %s 的数值无效", name)
			}
			values[name] = number

		default:
			values[name] = value
		}
	}
	if values["bind"] == nil || values["port"] == nil {
		return nil, errors.New("配置必须包含绑定地址和监听端口")
	}
	if err := m.validateValues(values); err != nil {
		return nil, err
	}
	return values, nil
}

func (m *Manager) configNumber(name, value string, memory bool) (float64, error) {
	value = strings.TrimSpace(value)
	if !memory {
		return strconv.ParseFloat(value, 64)
	}
	index := 0
	for index < len(value) && (value[index] >= '0' && value[index] <= '9' || value[index] == '.') {
		index++
	}
	number, err := strconv.ParseFloat(value[:index], 64)
	if err != nil {
		return 0, err
	}
	unit := strings.ToUpper(strings.TrimSpace(value[index:]))
	if unit == "" {
		if name == "shared_buffers" || name == "effective_cache_size" {
			return number / 128, nil
		}
		return number / 1024, nil
	}
	switch unit {
	case "B":
		return number / (1024 * 1024), nil
	case "K", "KB":
		return number / 1024, nil
	case "M", "MB":
		return number, nil
	case "G", "GB":
		return number * 1024, nil
	case "T", "TB":
		return number * 1024 * 1024, nil
	}
	return 0, errors.New("无效内存单位")
}

func (m *Manager) validateValues(values map[string]interface{}) error {
	fields := make(map[string]map[string]interface{})
	for _, field := range m.fields() {
		fields[field["name"].(string)] = field
	}
	for name, value := range values {
		field, exists := fields[name]
		if !exists {
			return fmt.Errorf("不支持的调优参数: %s", name)
		}
		switch field["type"] {
		case "number":
			number, err := strconv.ParseFloat(fmt.Sprint(value), 64)
			if err != nil || math.IsInf(number, 0) || math.IsNaN(number) {
				return fmt.Errorf("%s 必须是有效数字", field["label"])
			}
			for _, limit := range []string{"min", "max"} {
				bound, _ := strconv.ParseFloat(fmt.Sprint(field[limit]), 64)
				if field[limit] != nil && (limit == "min" && number < bound || limit == "max" && number > bound) {
					return fmt.Errorf("%s 超出允许范围", field["label"])
				}
			}
			if field["unit"] != "MB" && math.Trunc(number) != number {
				return fmt.Errorf("%s 必须是整数", field["label"])
			}
			values[name] = number

		case "text":
			bind, ok := value.(string)
			if !ok || strings.TrimSpace(bind) == "" {
				return errors.New("绑定地址不能为空")
			}
			for _, host := range strings.Split(bind, ",") {
				host = strings.TrimSpace(host)
				if net.ParseIP(host) == nil && host != "localhost" && host != "*" {
					return errors.New("绑定地址必须是 IP 地址，多个地址使用逗号分隔")
				}
			}
		}
	}
	return nil
}

// patchConfig 逐项替换已知参数，未知参数及注释不会被表单保存覆盖。
func (m *Manager) patchConfig(content string, values map[string]interface{}) (string, error) {
	if _, err := m.parseConfig(content); err != nil {
		return "", err
	}
	if err := m.validateValues(values); err != nil {
		return "", err
	}

	patches := make(map[string]string)
	for name, value := range values {
		key, text := m.configKey(name), fmt.Sprint(value)
		if number, ok := value.(float64); ok {
			text = strconv.FormatFloat(number, 'f', -1, 64)
		}
		if flag, ok := value.(bool); ok {
			text = "0"
			if flag {
				text = "1"
			}
		}
		for _, field := range m.fields() {
			if field["name"] == name && field["unit"] == "MB" {
				text += "M"

				text += "B"
			}
		}
		if name == "bind" || strings.HasSuffix(text, "MB") {
			text = "'" + strings.ReplaceAll(text, "'", "''") + "'"
		}
		patches[strings.ReplaceAll(key, "-", "_")] = key + " = " + text
	}
	replaced := make(map[string]bool)
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}
		key, _, _ := strings.Cut(trimmed, "=")

		parts := strings.Fields(key)
		if len(parts) > 0 {
			key = parts[0]
		}

		key = strings.ReplaceAll(strings.TrimSpace(key), "-", "_")
		if replacement, exists := patches[key]; exists {
			lines[i] = replacement
			replaced[key] = true
		}
	}
	// 替换所有重复项，只追加缺失参数，反复保存不会膨胀文件。

	for _, field := range m.fields() {
		key := strings.ReplaceAll(m.configKey(field["name"].(string)), "-", "_")
		if value, exists := patches[key]; exists && !replaced[key] {
			lines = append(lines, value)
		}
	}
	content = strings.Join(lines, "\n") + "\n"

	_, err := m.parseConfig(content)
	return content, err
}

func (m *Manager) GetConfig() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.available(); err != nil {
		return "", err
	}
	content, err := m.readFile("server.conf")
	return string(content), err
}

func (m *Manager) GetPerformance() (map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.available(); err != nil {
		return nil, err
	}
	content, err := m.readFile("server.conf")
	if err != nil {
		return nil, err
	}
	values, err := m.parseConfig(string(content))
	return map[string]interface{}{"fields": m.fields(), "values": values}, err
}

func (m *Manager) UpdatePerformance(ctx context.Context, values map[string]interface{}) (*ApplyResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.available(); err != nil {
		return nil, err
	}
	content, err := m.readFile("server.conf")
	if err != nil {
		return nil, err
	}
	next, err := m.patchConfig(string(content), values)
	if err != nil {
		return nil, err
	}
	return m.saveConfig(ctx, next)
}

func (m *Manager) UpdateConfig(ctx context.Context, content string) (*ApplyResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.available(); err != nil {
		return nil, err
	}
	if _, err := m.parseConfig(content); err != nil {
		return nil, err
	}
	return m.saveConfig(ctx, content)
}

func (m *Manager) saveConfig(ctx context.Context, content string) (*ApplyResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	active, err := m.readRunningConfig()
	if err != nil {
		return nil, err
	}
	previous, err := m.readFile("server.conf")
	if err != nil {
		return nil, err
	}
	if err = m.writeFile("server.conf", []byte(content)); err != nil {
		return nil, err
	}
	// 通用启动和重启直接读取挂载文件，因此保存时同步准备下次启动配置。
	if err = m.writeFile("config/server.conf", []byte(content)); err != nil {
		return nil, errors.Join(err, m.writeFile("server.conf", previous))
	}
	return &ApplyResult{Applied: string(active) == content, RestartRequired: string(active) != content}, nil
}

// Reload 热加载可动态修改的参数，其余改动保留到用户显式重启。
func (m *Manager) Reload(ctx context.Context) (*ApplyResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.available(); err != nil {
		return nil, err
	}
	state, err := m.readState()
	if err != nil {
		return nil, err
	}
	version := m.versionNumber(state.Version)
	next, err := m.readFile("server.conf")
	if err != nil {
		return nil, err
	}
	previous, err := m.readRunningConfig()
	if err != nil {
		return nil, err
	}
	instance, err := m.container.Info(m.InstanceID)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(next, previous) {
		if instance.Status == container.StatusRunning && version >= 90500 {
			output, err := m.query(ctx, "SELECT count(*) FROM pg_settings WHERE pending_restart;")
			if err != nil {
				return nil, err
			}
			pending := strings.TrimSpace(output) != "0"
			return &ApplyResult{Applied: !pending, RestartRequired: pending}, nil
		}
		return &ApplyResult{Applied: true}, nil
	}
	if instance.Status != container.StatusRunning {
		return &ApplyResult{RestartRequired: true}, nil
	}

	oldValues, err := m.parseConfig(string(previous))
	if err != nil {
		return nil, err
	}
	nextValues, err := m.parseConfig(string(next))
	if err != nil {
		return nil, err
	}
	for _, name := range []string{"bind", "port", "max_connections", "shared_buffers"} {
		if oldValues[name] != nextValues[name] {
			return &ApplyResult{RestartRequired: true}, nil
		}
	}
	// 只热加载明确支持的参数。原文中的其他变更保留为待重启，避免
	// SIGHUP 异步处理期间误报成功，尤其是 wal_level 等启动参数。
	remaining := [2]string{}
	for index, content := range [][]byte{previous, next} {
		var lines []string
		for _, line := range strings.Split(string(content), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, _, _ := strings.Cut(line, "=")
			parts := strings.Fields(key)
			if len(parts) > 0 {
				switch parts[0] {
				case "work_mem", "maintenance_work_mem", "effective_cache_size", "log_min_duration_statement":
					continue
				}
			}
			lines = append(lines, line)
		}
		remaining[index] = strings.Join(lines, "\n")
	}
	if remaining[0] != remaining[1] {
		return &ApplyResult{RestartRequired: true}, nil
	}
	// 旧版没有 pg_file_settings；用事务内的 SET 校验同一组热加载参数，随后回滚。
	validation := "BEGIN;"
	for _, line := range strings.Split(string(next), "\n") {
		key, _, _ := strings.Cut(strings.TrimSpace(line), "=")
		parts := strings.Fields(key)
		if len(parts) == 0 {
			continue
		}
		name := parts[0]
		switch name {
		case "work_mem", "maintenance_work_mem", "effective_cache_size", "log_min_duration_statement":
		default:
			continue
		}
		// 逐行校验，避免后面的重复配置遮住前一行的无效值。
		values, err := m.parseConfig("listen_addresses='127.0.0.1'\nport=5432\n" + line)
		if err != nil {
			return nil, err
		}
		value := strconv.FormatFloat(values[name].(float64), 'f', -1, 64)
		if name != "log_min_duration_statement" {
			value += "MB"
		}
		validation += "SELECT set_config('" + name + "','" + value + "',true);"
	}
	if _, err = m.query(ctx, validation+"ROLLBACK;"); err != nil {
		return nil, err
	}
	if err = m.writeFile("config/server.conf", next); err != nil {
		return nil, err
	}
	var output string
	if version >= 90500 {
		output, err = m.query(ctx, "SELECT count(*) FROM pg_file_settings WHERE error IS NOT NULL;")
		if err == nil && strings.TrimSpace(output) != "0" {
			err = errors.New("PostgreSQL 配置校验失败，请检查配置项名称和值")
		}
	}
	var loadedAt string
	if err == nil {
		loadedAt, err = m.query(ctx, "SELECT pg_conf_load_time();")
	}
	if err == nil {
		output, err = m.query(ctx, "SELECT pg_reload_conf();")
		if err == nil && strings.TrimSpace(output) != "t" {
			err = errors.New("PostgreSQL 未能重新加载配置")
		}
	}
	if err != nil {
		// 挂载文件保留已保存的目标，确保热加载失败后通用重启仍使用新配置。
		return nil, err
	}
	// pg_reload_conf 只确认信号已发出；新连接看到加载时间变化后再更新运行快照。
	reloadCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		output, err = m.query(reloadCtx, "SELECT pg_conf_load_time();")
		if err != nil {
			return nil, err
		}
		if output != loadedAt {
			break
		}
		select {
		case <-reloadCtx.Done():
			return nil, errors.New("等待 PostgreSQL 加载配置超时，配置已保存，可重试热加载或重启")
		case <-ticker.C:
		}
	}
	if version >= 90500 {
		// PostgreSQL 对额外的只读启动参数会设置 pending_restart，交由页面提示。
		output, err = m.query(ctx, "SELECT count(*) FROM pg_settings WHERE pending_restart;")
		if err != nil || strings.TrimSpace(output) != "0" {
			return &ApplyResult{RestartRequired: true}, nil
		}
	}
	if err = m.writeRunningConfig(next, instance.Generation); err != nil {
		return nil, err
	}
	return &ApplyResult{Applied: true}, nil
}
