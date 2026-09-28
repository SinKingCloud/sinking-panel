package mysql

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
)

func (m *Manager) fields() []map[string]interface{} {
	fields := []map[string]interface{}{
		{"name": "bind", "label": "绑定地址", "type": "text", "tooltip": "默认只允许本机连接；对外监听前请配置防火墙和数据库访问权限。"},
		{"name": "port", "label": "监听端口", "type": "number", "min": 1, "max": 65535},
	}

	return append(fields, m.performanceFields()...)
}

// configKey 将表单名映射到各数据库的原生配置项。
func (m *Manager) configKey(name string) string {
	if name == "bind" {
		return "bind-address"
	}

	return name
}

// timezoneConfig 仅补齐缺失的服务端时区选项，保留用户在有效分组中的显式配置。
func (m *Manager) timezoneConfig(content string) string {
	versionGroup := ""
	if state, err := m.readState(); err == nil {
		var major, minor int
		if _, err = fmt.Sscanf(state.Version, "%d.%d", &major, &minor); err == nil {
			versionGroup = fmt.Sprintf("mysqld-%d.%d", major, minor)
		}
	}
	section := ""
	hasTimezone, hasLogTimestamps := false, false
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if end := strings.IndexByte(line, ']'); strings.HasPrefix(line, "[") && end > 0 {
			section = strings.ToLower(strings.TrimSpace(line[1:end]))
			continue
		}
		if section != "mysqld" && section != "server" && (versionGroup == "" || section != versionGroup) {
			continue
		}
		key, _, _ := strings.Cut(line, "=")
		key = strings.TrimPrefix(strings.ReplaceAll(strings.TrimSpace(key), "-", "_"), "loose_")
		switch key {
		case "default_time_zone":
			hasTimezone = true
		case "log_timestamps":
			hasLogTimestamps = true
		}
	}
	if hasTimezone && hasLogTimestamps {
		return content
	}
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if section != "mysqld" && section != "server" {
		content += "[mysqld]\n"
	}
	if !hasTimezone {
		content += "default-time-zone=SYSTEM\n"
	}
	if !hasLogTimestamps {
		// 旧版不支持日志时区选项时忽略它，仍保留 SYSTEM 的数据库时区。
		content += "loose-log-timestamps=SYSTEM\n"
	}
	return content
}

// parseConfig 只提取可视化管理项；其他原文始终保留并交由数据库验证。
func (m *Manager) parseConfig(content string) (map[string]interface{}, error) {
	if len(content) == 0 || len(content) > configMaxSize || strings.ContainsRune(content, 0) {
		return nil, errors.New("配置不能为空、包含空字符或超过 256 KB")
	}
	raw := make(map[string]string)
	versionGroup := ""
	if state, err := m.readState(); err == nil {
		var major, minor int
		if _, err = fmt.Sscanf(state.Version, "%d.%d", &major, &minor); err == nil {
			versionGroup = fmt.Sprintf("mysqld-%d.%d", major, minor)
		}
	}

	section := ""
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "!") || strings.HasPrefix(line, "include") {
			return nil, errors.New("请将配置直接写入当前文件，不支持引用其他配置文件")
		}
		if end := strings.IndexByte(line, ']'); strings.HasPrefix(line, "[") && end > 0 {
			section = strings.ToLower(strings.TrimSpace(line[1:end]))
			if strings.HasPrefix(section, "mysqld-") && versionGroup == "" {
				return nil, errors.New("未取得 MySQL 版本，请先安装后再配置版本专属分组")
			}
			continue
		}
		if section != "mysqld" && section != "server" && (versionGroup == "" || section != versionGroup) {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			key, value = line, "1"
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

		// loose 只决定未知参数是否报错，对数据库认识的参数仍会生效。
		key = strings.TrimPrefix(strings.ReplaceAll(key, "-", "_"), "loose_")
		for _, name := range []string{"datadir", "user", "socket"} {
			if key == "enable_"+name || key == "disable_"+name || key == "skip_"+name {
				return nil, fmt.Errorf("受管配置 %s 不能使用布尔选项前缀", name)
			}
		}
		for _, name := range []string{"mysqlx", "skip_grant_tables", "daemonize", "skip_networking"} {
			inverse := key == "skip_"+name || key == "disable_"+name
			if key != "enable_"+name && !inverse {
				continue
			}
			// 旧版 enable-/disable- 对 off/false 的解释与普通布尔选项不同。
			if value != "0" && value != "1" {
				return nil, fmt.Errorf("配置 %s 的布尔别名请使用 0 或 1", key)
			}
			key = name
			if inverse {
				if value == "0" {
					value = "1"
				} else {
					value = "0"
				}
			}
		}

		raw[key] = value
	}

	if raw["datadir"] != "/var/lib/mysql" || raw["user"] != "mysql" {
		return nil, errors.New("请保留 datadir=/var/lib/mysql 和 user=mysql，保证数据持久化及运行用户正确")
	}
	if raw["socket"] != "/var/run/mysqld/mysqld.sock" {
		return nil, errors.New("请保留 socket=/var/run/mysqld/mysqld.sock，保证本地管理连接正常")
	}
	for _, key := range []string{"skip_grant_tables", "daemonize", "skip_networking"} {
		if value := raw[key]; value != "" && value != "0" && !strings.EqualFold(value, "off") && !strings.EqualFold(value, "false") {
			return nil, fmt.Errorf("受管数据库不支持开启 %s", key)
		}
	}
	// loose 前缀让未内置 X Plugin 的旧版本忽略该项，8.0 及之后仍关闭额外端口。
	if value := raw["mysqlx"]; value != "0" && !strings.EqualFold(value, "off") && !strings.EqualFold(value, "false") {
		return nil, errors.New("请保留 loose-mysqlx=0，避免额外监听未经配置的端口")
	}

	values := make(map[string]interface{})
	for _, field := range m.fields() {
		name := field["name"].(string)
		key := m.configKey(name)

		key = strings.ReplaceAll(key, "-", "_")

		value, exists := raw[key]
		if !exists {
			continue
		}
		switch field["type"] {
		case "number":
			number, err := m.configNumber(value, field["unit"] == "MB")
			if err != nil {
				return nil, fmt.Errorf("配置 %s 的数值无效", name)
			}
			values[name] = number
		case "switch":
			switch strings.ToLower(value) {
			case "1", "on", "true":
				values[name] = true
			case "0", "off", "false":
				values[name] = false
			default:
				return nil, fmt.Errorf("配置 %s 的开关值无效", name)
			}
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

func (m *Manager) configNumber(value string, memory bool) (float64, error) {
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
		return number / (1024 * 1024), nil
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
			if name != "long_query_time" && field["unit"] != "MB" && math.Trunc(number) != number {
				return fmt.Errorf("%s 必须是整数", field["label"])
			}
			values[name] = number
		case "switch":
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("%s 必须是布尔值", field["label"])
			}
		case "select":
			if value != "off" && value != "slowOp" && value != "all" {
				return errors.New("无效的性能分析模式")
			}
		case "text":
			bind, ok := value.(string)
			if !ok || strings.TrimSpace(bind) == "" {
				return errors.New("绑定地址不能为空")
			}
			for _, host := range strings.Split(bind, ",") {
				host = strings.TrimSpace(host)
				if net.ParseIP(host) == nil && host != "localhost" {
					return errors.New("绑定地址必须是 IP 地址，多个地址使用逗号分隔")
				}
			}
			if strings.Contains(bind, ",") {
				return errors.New("MySQL 绑定地址请填写单个 IP 地址")
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
				text = strconv.FormatInt(int64(value.(float64)*1024*1024), 10)
			}
		}

		patches[strings.ReplaceAll(key, "-", "_")] = key + " = " + text
	}
	versionGroup := ""
	if state, err := m.readState(); err == nil {
		var major, minor int
		if _, err = fmt.Sscanf(state.Version, "%d.%d", &major, &minor); err == nil {
			versionGroup = fmt.Sprintf("mysqld-%d.%d", major, minor)
		}
	}
	section := ""
	replaced := make(map[string]bool)
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if end := strings.IndexByte(trimmed, ']'); strings.HasPrefix(trimmed, "[") && end > 0 {
			section = strings.ToLower(strings.TrimSpace(trimmed[1:end]))
			continue
		}
		if section != "mysqld" && section != "server" && (versionGroup == "" || section != versionGroup) || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}
		key, _, _ := strings.Cut(trimmed, "=")

		key = strings.TrimPrefix(strings.ReplaceAll(strings.TrimSpace(key), "-", "_"), "loose_")
		if replacement, exists := patches[key]; exists {
			lines[i] = replacement
			replaced[key] = true
		}
	}
	// 替换所有重复项，只为缺失参数追加 section，反复保存不会膨胀文件。
	if len(patches) > len(replaced) {
		lines = append(lines, "[mysqld]")
	}
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

// saveConfig 同时保存编辑配置和下次启动文件；运行配置快照保留到实例运行代次变化。
func (m *Manager) saveConfig(ctx context.Context, content string) (*ApplyResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	content = m.timezoneConfig(content)
	if _, err := m.parseConfig(content); err != nil {
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
	if err = m.writeFile("config/server.conf", []byte(content)); err != nil {
		return nil, errors.Join(err, m.writeFile("server.conf", previous))
	}
	return &ApplyResult{Applied: string(active) == content, RestartRequired: string(active) != content}, nil
}

// Reload 检查配置是否待生效；数据库不支持整体热重载时由用户明确执行重启。
func (m *Manager) Reload(ctx context.Context) (*ApplyResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.available(); err != nil {
		return nil, err
	}
	next, err := m.readFile("server.conf")
	if err != nil {
		return nil, err
	}
	if content := m.timezoneConfig(string(next)); content != string(next) {
		return m.saveConfig(ctx, content)
	}
	previous, err := m.readRunningConfig()
	if err != nil {
		return nil, err
	}
	if bytes.Equal(next, previous) {
		return &ApplyResult{Applied: true}, nil
	}

	return &ApplyResult{RestartRequired: true}, nil
}
