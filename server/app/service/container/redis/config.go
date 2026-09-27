package redis

import (
	"errors"
	"fmt"
	"math"
	"net/netip"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const redisConfigMaxSize = 256 * 1024

// defaultConfig 按镜像版本生成配置，旧版不写入其尚未支持的指令。
func (m *Manager) defaultConfig(major, minor int) string {
	content := `# Redis 默认配置。
bind 127.0.0.1
protected-mode yes
port 6379
timeout 0
tcp-backlog 511
tcp-keepalive 300
maxclients 10000
databases 16
requirepass ""
maxmemory 0
maxmemory-policy noeviction

# 持久化数据写入独立挂载的数据目录。
dir /data
dbfilename dump.rdb
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
appendonly yes
appendfilename appendonly.aof
appenddirname appendonlydir
appendfsync everysec
no-appendfsync-on-rewrite no
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb

daemonize no
supervised no
loglevel notice
logfile ""
pidfile /tmp/redis.pid
`
	// Redis 7 开始使用多文件 AOF，旧版本不识别 appenddirname。
	if major < 7 {
		content = strings.ReplaceAll(content, "appenddirname appendonlydir\n", "")
	}
	if major < 3 || major == 3 && minor < 2 {
		content = strings.ReplaceAll(content, "protected-mode yes\n", "")
	}
	if major < 3 {
		content = strings.ReplaceAll(content, "supervised no\n", "")
	}
	if major < 2 || major == 2 && minor < 8 {
		content = strings.ReplaceAll(content, "tcp-backlog 511\n", "")
		content = strings.ReplaceAll(content, "logfile \"\"\n", "logfile stdout\n")
	}
	return content
}

// parseConfig 解析面板表单和热更新所需的参数，其余指令保留原文，由 Redis 加载时校验。
func (m *Manager) parseConfig(content string) (*redisConfig, error) {
	if len(content) > redisConfigMaxSize {
		return nil, errors.New("Redis 配置不能超过 256 KiB")
	}
	if !utf8.ValidString(content) || strings.ContainsRune(content, '\x00') {
		return nil, errors.New("Redis 配置必须是有效 UTF-8 文本且不能包含空字节")
	}
	config := &redisConfig{
		Persistence: Persistence{Save: []SaveRule{}},
		values:      map[string][]string{"requirepass": {""}},
		content:     content,
	}
	externalBind := false
	parseInteger := func(value string, min, max int64) (int64, error) {
		if value == "" || strings.Trim(value, "0123456789-") != "" || strings.HasPrefix(value, "-0") {
			return 0, errors.New("必须为十进制整数")
		}
		number, err := strconv.ParseInt(value, 10, 64)
		if err != nil || number < min || number > max {
			return 0, fmt.Errorf("必须为 %d 至 %d 的整数", min, max)
		}
		return number, nil
	}
	lines := strings.SplitAfter(content, "\n")
	var passwordLines []int
	normalized := false
	for index, line := range lines {
		lineNumber := index + 1
		args, err := m.configArguments(line)
		if err != nil {
			return nil, fmt.Errorf("Redis 配置第 %d 行：%w", lineNumber, err)
		}
		if len(args) == 0 {
			continue
		}
		key, values := strings.ToLower(args[0]), args[1:]
		switch key {
		case "bind", "save":
		case "port", "databases", "timeout", "maxclients", "requirepass", "maxmemory", "appendonly", "appendfsync",
			"auto-aof-rewrite-percentage", "auto-aof-rewrite-min-size", "maxmemory-policy", "maxmemory-samples",
			"loglevel", "tcp-keepalive", "hz", "dynamic-hz", "lazyfree-lazy-eviction", "lazyfree-lazy-expire",
			"lazyfree-lazy-server-del", "lazyfree-lazy-user-del", "lazyfree-lazy-user-flush", "replica-lazy-flush",
			"stop-writes-on-bgsave-error", "rdbcompression", "no-appendfsync-on-rewrite",
			"aof-rewrite-incremental-fsync", "rdb-save-incremental-fsync":
			if len(values) != 1 {
				return nil, fmt.Errorf("Redis 配置第 %d 行的 %s 必须恰好有一个参数", lineNumber, key)
			}
		default:
			// include、loadmodule 等可多次出现，保留每一行及其在文件中的位置。
			config.includes = config.includes || key == "include"
			config.directives = append(config.directives, strings.TrimSpace(line))
			continue
		}
		previous := config.values[key]
		if key != "requirepass" && (key != "save" || len(config.directives) == 0 || config.directives[len(config.directives)-1] != "save") {
			// 已解析字段只记录名称；连续 save 合并比较，跨 include 等指令则保留位置。
			config.directives = append(config.directives, key)
		}
		switch key {
		case "port", "timeout", "maxclients", "databases", "auto-aof-rewrite-percentage":
			min, max := int64(0), int64(math.MaxInt32)
			switch key {
			case "port":
				min, max = 1, 65535
			case "maxclients":
				min = 1
			case "databases":
				min, max = 1, 65536
			}
			number, err := parseInteger(values[0], min, max)
			if err != nil {
				return nil, fmt.Errorf("Redis 配置 %s %w", key, err)
			}
			switch key {
			case "port":
				config.Port = int(number)
			case "timeout":
				config.Timeout = int(number)
			case "maxclients":
				config.Maxclients = int(number)
			case "databases":
				config.Databases = int(number)
			case "auto-aof-rewrite-percentage":
				config.AutoAofRewritePercentage = int(number)
			}
		case "appendonly":
			if values[0] != "yes" && values[0] != "no" {
				return nil, errors.New("Redis 配置 appendonly 只允许 yes 或 no")
			}
			config.AppendOnly = values[0] == "yes"
		case "appendfsync":
			if values[0] != "always" && values[0] != "everysec" && values[0] != "no" {
				return nil, errors.New("Redis 配置 appendfsync 只允许 always、everysec 或 no")
			}
			config.AppendFsync = values[0]
		case "bind":
			if len(values) == 0 || len(values) > 16 {
				return nil, errors.New("Redis bind 需要 1 至 16 个以空格分隔的 IP 地址")
			}
			required := false
			externalBind = false
			for _, value := range values {
				required = required || !strings.HasPrefix(value, "-")
				address, err := netip.ParseAddr(strings.TrimPrefix(value, "-"))
				if err != nil || address.Zone() != "" {
					return nil, errors.New("Redis bind 只允许实际 IPv4 或 IPv6 地址，可使用 - 前缀")
				}
				externalBind = externalBind || !address.Unmap().IsLoopback()
			}
			if !required {
				return nil, errors.New("Redis bind 至少需要一个不带 - 前缀的监听地址")
			}
			config.Bind = strings.Join(values, " ")
		case "requirepass":
			passwordLines = append(passwordLines, index)
			if len(values[0]) > 512 || !utf8.ValidString(values[0]) || strings.ContainsRune(values[0], '\x00') {
				return nil, errors.New("Redis 密码必须是有效 UTF-8 文本、不能包含空字节且不能超过 512 字节")
			}
			config.Requirepass = values[0]
		case "maxmemory", "auto-aof-rewrite-min-size":
			megabytes, err := m.configMemoryMB(values[0])
			if err != nil {
				return nil, fmt.Errorf("Redis 配置 %s：%w", key, err)
			}
			if key == "maxmemory" {
				config.Maxmemory = megabytes
			} else {
				config.AutoAofRewriteMinSize = megabytes
			}
		case "save":
			// 多行 save 依次累加；空参数清空此前规则，与 Redis 配置加载语义一致。
			if len(values) == 1 && values[0] == "" {
				config.Save = config.Save[:0]
				config.values[key] = values
				continue
			}
			if len(values) == 0 || len(values)%2 != 0 || len(config.Save)+len(values)/2 > 128 {
				return nil, errors.New("Redis save 必须为秒数和修改次数的成对参数，最多 128 条规则，或使用 save \"\" 关闭")
			}
			for i := 0; i < len(values); i += 2 {
				seconds, secondsErr := parseInteger(values[i], 1, math.MaxInt32)
				changes, changesErr := parseInteger(values[i+1], 0, math.MaxInt32)
				if secondsErr != nil || changesErr != nil {
					return nil, errors.New("Redis save 秒数必须为 1 至 2147483647，修改次数必须为 0 至 2147483647")
				}
				config.Save = append(config.Save, SaveRule{Seconds: int(seconds), Changes: int(changes)})
			}
			// 旧版只读取每行一对 save 参数，多对同一行会被静默忽略。
			// 统一为所有版本都支持的多行形式，保留原位置、缩进和换行风格。
			if len(values) > 2 {
				newline := "\n"
				if strings.HasSuffix(line, "\r\n") || strings.Contains(content, "\r\n") {
					newline = "\r\n"
				}
				indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
				var replacement strings.Builder
				for i := 0; i < len(values); i += 2 {
					if i > 0 {
						replacement.WriteString(newline)
					}
					replacement.WriteString(indent + "save " + values[i] + " " + values[i+1])
				}
				if strings.HasSuffix(line, "\n") {
					replacement.WriteString(newline)
				}
				lines[index], normalized = replacement.String(), true
			}
			if len(previous) == 1 && previous[0] == "" {
				previous = nil
			}
			values = append(previous, values...)
		}
		config.values[key] = values
	}
	// 表单读写和 Redis 连接需要这些字段；其他配置项可自行增删。
	for _, key := range []string{
		"bind", "port", "timeout", "maxclients", "databases", "requirepass", "maxmemory",
		"save", "appendonly", "appendfsync", "auto-aof-rewrite-percentage", "auto-aof-rewrite-min-size",
	} {
		if _, ok := config.values[key]; !ok {
			return nil, fmt.Errorf("Redis 配置必须保留 %s", key)
		}
	}
	if externalBind && strings.TrimSpace(config.Requirepass) == "" {
		return nil, errors.New("Redis 监听非回环地址时必须设置非空密码")
	}
	// 旧版将 requirepass "" 解释为需要空密码认证；不配置此指令才是真正关闭密码。
	if config.Requirepass == "" {
		for _, index := range passwordLines {
			lines[index], normalized = "", true
		}
	}
	if normalized {
		config.content = strings.Join(lines, "")
	}
	return config, nil
}

// patchConfig 在最后一次指令的位置更新，保留 include 覆盖顺序、注释、未改行和换行方式。
func (m *Manager) patchConfig(content string, values map[string][]string) (*redisConfig, error) {
	config, err := m.parseConfig(content)
	if err != nil || len(values) == 0 {
		return config, err
	}
	// Redis 不识别 Go 的 \u 转义；控制字节用 \xHH，UTF-8 字节原样保留。
	quote := func(value string) string {
		if value != "" && strings.IndexAny(value, " \t\r\n\"'\\#") < 0 && strings.IndexFunc(value, func(r rune) bool { return r < 32 || r == 127 }) < 0 {
			return value
		}
		var result strings.Builder
		result.WriteByte('"')
		const hex = "0123456789abcdef"
		for i := 0; i < len(value); i++ {
			c := value[i]
			switch {
			case c == '\\' || c == '"':
				result.WriteByte('\\')
				result.WriteByte(c)
			case c < 32 || c == 127:
				result.WriteString("\\x")
				result.WriteByte(hex[c>>4])
				result.WriteByte(hex[c&15])
			default:
				result.WriteByte(c)
			}
		}
		result.WriteByte('"')
		return result.String()
	}
	replacements := make(map[string]string, len(values))
	for key, args := range values {
		if key == "" || strings.Trim(key, "abcdefghijklmnopqrstuvwxyz0123456789-") != "" || len(args) == 0 {
			return nil, errors.New("Redis 配置更新指令或参数无效")
		}
		encoded := make([]string, len(args)+1)
		encoded[0] = key
		for index, value := range args {
			encoded[index+1] = quote(value)
		}
		replacements[key] = strings.Join(encoded, " ")
	}
	newline := "\n"
	if strings.Contains(content, "\r\n") {
		newline = "\r\n"
	}
	lines := strings.SplitAfter(content, "\n")
	for index := len(lines) - 1; index >= 0; index-- {
		original := lines[index]
		line := strings.TrimSuffix(strings.TrimSuffix(original, "\n"), "\r")
		args, _ := m.configArguments(line)
		if len(args) == 0 {
			continue
		}
		key := strings.ToLower(args[0])
		replacement, replace := replacements[key]
		if !replace {
			continue
		}
		// 倒序写入最后一次指令，删除此前同名行，避免把覆盖值移到 include 之前。
		if replacement == "" {
			lines[index] = ""
			continue
		}
		replacements[key] = ""
		lines[index] = line[:len(line)-len(strings.TrimLeft(line, " \t"))] + replacement
		if strings.HasSuffix(original, "\r\n") {
			lines[index] += "\r\n"
		} else if strings.HasSuffix(original, "\n") {
			lines[index] += "\n"
		}
	}
	var result strings.Builder
	result.Grow(len(content))
	result.WriteString(strings.Join(lines, ""))
	keys := make([]string, 0, len(replacements))
	for key, replacement := range replacements {
		if replacement != "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if len(keys) > 0 && result.Len() > 0 && !strings.HasSuffix(result.String(), "\n") {
		result.WriteString(newline)
	}
	for _, key := range keys {
		result.WriteString(replacements[key])
		result.WriteString(newline)
	}
	return m.parseConfig(result.String())
}

// configArguments 与 Redis sdssplitargs 保持相同引号和转义语义；# 仅可作整行注释。
func (m *Manager) configArguments(line string) ([]string, error) {
	line = strings.Trim(line, " \t\r\n")
	if line == "" || line[0] == '#' {
		return nil, nil
	}
	// 常见的无引号指令直接切片；只认 Redis 的 ASCII 空白，不把 Unicode 空格拆开。
	if !strings.ContainsAny(line, "\"'") {
		return strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '\t' || r == '\r' || r == '\n' }), nil
	}
	var args []string
	space := func(c byte) bool { return c == ' ' || c == '\t' || c == '\r' || c == '\n' }
	for i := 0; i < len(line); {
		for i < len(line) && space(line[i]) {
			i++
		}
		if i == len(line) {
			break
		}
		var token strings.Builder
		var quote byte
		for i < len(line) {
			c := line[i]
			if quote == 0 {
				if space(c) {
					break
				}
				if c == '"' || c == '\'' {
					quote = c
					i++
					continue
				}
				token.WriteByte(c)
				i++
				continue
			}
			if c == quote {
				quote = 0
				i++
				if i < len(line) && !space(line[i]) {
					return nil, errors.New("引号结束后必须为空白或行尾")
				}
				break
			}
			// 单引号只转义 \'; 双引号支持 \xHH 和常用控制字符，其余转义去掉反斜杠。
			if c == '\\' && i+1 < len(line) && (quote == '"' || line[i+1] == '\'') {
				i++
				c = line[i]
				if quote == '"' {
					if c == 'x' && i+2 < len(line) && strings.ContainsRune("0123456789abcdefABCDEF", rune(line[i+1])) && strings.ContainsRune("0123456789abcdefABCDEF", rune(line[i+2])) {
						number, _ := strconv.ParseUint(line[i+1:i+3], 16, 8)
						token.WriteByte(byte(number))
						i += 3
						continue
					}
					switch c {
					case 'n':
						c = '\n'
					case 'r':
						c = '\r'
					case 't':
						c = '\t'
					case 'b':
						c = '\b'
					case 'a':
						c = '\a'
					}
				}
			}
			token.WriteByte(c)
			i++
		}
		if quote != 0 {
			return nil, errors.New("字符串引号未闭合")
		}
		args = append(args, token.String())
	}
	return args, nil
}

// configMemoryMB 遵循 Redis 的十进制 k/m/g 和二进制 kb/mb/gb 单位。
func (m *Manager) configMemoryMB(value string) (int64, error) {
	value = strings.ToLower(value)
	index := 0
	for index < len(value) && value[index] >= '0' && value[index] <= '9' {
		index++
	}
	if index == 0 {
		return 0, errors.New("内存大小必须为非负整数，可使用 kb、mb 或 gb 单位")
	}
	number, err := strconv.ParseInt(value[:index], 10, 64)
	if err != nil {
		return 0, errors.New("内存大小超出有效范围")
	}
	multiplier := int64(1)
	switch value[index:] {
	case "", "b":
	case "k":
		multiplier = 1000
	case "kb":
		multiplier = 1024
	case "m":
		multiplier = 1000 * 1000
	case "mb":
		multiplier = 1024 * 1024
	case "g":
		multiplier = 1000 * 1000 * 1000
	case "gb":
		multiplier = 1024 * 1024 * 1024
	default:
		return 0, errors.New("不支持此内存大小单位")
	}
	if number > math.MaxInt64/multiplier {
		return 0, errors.New("内存大小超出有效范围")
	}
	bytes := number * multiplier
	if bytes%(1024*1024) != 0 {
		return 0, errors.New("内存大小必须是整数 MiB（1 MiB = 1048576 字节），建议使用 mb 单位")
	}
	return bytes / (1024 * 1024), nil
}
