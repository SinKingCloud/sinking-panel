package mongo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func (m *Manager) fields(content string, versions ...string) []map[string]interface{} {
	version := ""
	if len(versions) > 0 {
		version = versions[0]
	} else if state, err := m.readState(); err == nil {
		version = state.Version
	}
	var document yaml.Node
	_ = yaml.Unmarshal([]byte(content), &document)
	wiredTiger := !strings.HasPrefix(version, "2.") && !strings.HasPrefix(version, "3.0.")
	if engine := m.yamlValue(&document, "storage.engine", nil); engine != nil {
		wiredTiger = engine.Value == "wiredTiger"
	}
	fields := []map[string]interface{}{
		{"name": "bind", "label": "绑定地址", "type": "text", "tooltip": "默认只允许本机连接；对外监听前请配置防火墙和数据库访问权限。"},
		{"name": "port", "label": "监听端口", "type": "number", "min": 1, "max": 65535},
	}

	return append(fields, m.performanceFields(version, wiredTiger)...)
}

// configKey 将表单名映射到 MongoDB 的原生配置项。
func (m *Manager) configKey(name string) string {
	return map[string]string{"bind": "net.bindIp", "port": "net.port", "maxIncomingConnections": "net.maxIncomingConnections", "cacheSizeGB": "storage.wiredTiger.engineConfig.cacheSizeGB", "slowms": "operationProfiling.slowOpThresholdMs", "mode": "operationProfiling.mode"}[name]
}

// yamlValue 只修改指定节点，保留 MongoDB 的未知配置、顺序和注释。
func (m *Manager) yamlValue(document *yaml.Node, key string, value *yaml.Node) *yaml.Node {
	node := document
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}
	parts := strings.Split(key, ".")
	for index, part := range parts {
		if node.Kind != yaml.MappingNode {
			return nil
		}
		var found *yaml.Node
		for i := 0; i+1 < len(node.Content); i += 2 {
			if node.Content[i].Value == part {
				found = node.Content[i+1]
				break
			}
		}
		if found == nil {
			if value == nil {
				return nil
			}
			found = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: part}, found)
		}
		if index == len(parts)-1 {
			if value != nil {
				value.HeadComment, value.LineComment, value.FootComment = found.HeadComment, found.LineComment, found.FootComment
				*found = *value
			}
			return found
		}
		node = found
	}
	return nil
}

// parseConfig 只提取可视化管理项；其他原文始终保留并交由数据库验证。
func (m *Manager) parseConfig(content string, versions ...string) (map[string]interface{}, error) {
	if len(content) == 0 || len(content) > configMaxSize || strings.ContainsRune(content, 0) {
		return nil, errors.New("配置不能为空、包含空字符或超过 256 KB")
	}
	raw := make(map[string]string)

	var document yaml.Node
	var decoded map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &decoded); err != nil {
		return nil, fmt.Errorf("MongoDB YAML 格式无效: %w", err)
	}
	if err := yaml.Unmarshal([]byte(content), &document); err != nil || len(document.Content) == 0 || document.Content[0].Kind != yaml.MappingNode {
		return nil, errors.New("MongoDB 配置必须是 YAML 对象")
	}
	for _, field := range m.fields(content, versions...) {
		key := m.configKey(field["name"].(string))
		if node := m.yamlValue(&document, key, nil); node != nil {
			if node.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("配置 %s 必须是单个值", key)
			}
			raw[key] = node.Value
		}
	}
	for key, expected := range map[string]string{"storage.dbPath": "/data/db", "security.authorization": "enabled"} {
		if node := m.yamlValue(&document, key, nil); node == nil || node.Value != expected {
			return nil, fmt.Errorf("受管数据库的 %s 必须为 %s", key, expected)
		}
	}
	if node := m.yamlValue(&document, "processManagement.fork", nil); node != nil && node.Value != "false" {
		return nil, errors.New("数据库必须以前台方式运行")
	}

	values := make(map[string]interface{})
	for _, field := range m.fields(content, versions...) {
		name := field["name"].(string)
		key := m.configKey(name)

		value, exists := raw[key]
		if !exists {
			continue
		}
		switch field["type"] {
		case "number":
			number, err := strconv.ParseFloat(value, 64)
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
	if err := m.validateValues(content, values, versions...); err != nil {
		return nil, err
	}
	return values, nil
}

func (m *Manager) validateValues(content string, values map[string]interface{}, versions ...string) error {
	fields := make(map[string]map[string]interface{})
	for _, field := range m.fields(content, versions...) {
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
			if (name != "cacheSizeGB" || field["step"] == float64(1)) && math.Trunc(number) != number {
				return fmt.Errorf("%s 必须是整数", field["label"])
			}
			values[name] = number

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
		}
	}
	return nil
}

// patchConfig 逐项替换已知参数，未知参数及注释不会被表单保存覆盖。
func (m *Manager) patchConfig(content string, values map[string]interface{}, versions ...string) (string, error) {
	if _, err := m.parseConfig(content, versions...); err != nil {
		return "", err
	}
	if err := m.validateValues(content, values, versions...); err != nil {
		return "", err
	}

	var document yaml.Node
	_ = yaml.Unmarshal([]byte(content), &document)
	for name, value := range values {
		node := &yaml.Node{}
		if err := node.Encode(value); err != nil {
			return "", err
		}
		if m.yamlValue(&document, m.configKey(name), node) == nil {
			return "", fmt.Errorf("配置 %s 的上级节点不是对象", name)
		}
	}
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	if err := encoder.Encode(&document); err != nil {
		return "", err
	}
	content = buffer.String()

	_, err := m.parseConfig(content, versions...)
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
	return map[string]interface{}{"fields": m.fields(string(content)), "values": values}, err
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
	previous, err := m.readRunningConfig()
	if err != nil {
		return nil, err
	}
	if bytes.Equal(next, previous) {
		return &ApplyResult{Applied: true}, nil
	}

	return &ApplyResult{RestartRequired: true}, nil
}
