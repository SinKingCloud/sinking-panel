//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package timezone

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

// fileSources 读取标准 chrony、ntpd 及 Linux timesyncd 文件配置。
// 不执行命令，不猜测自定义启动参数或 macOS timed 的私有配置格式。
func fileSources(ctx context.Context) ([]string, error) {
	for _, path := range []string{"/etc/chrony.conf", "/etc/chrony/chrony.conf", "/usr/local/etc/chrony.conf", "/etc/ntp.conf", "/etc/ntpsec/ntp.conf", "/etc/ntpd.conf", "/usr/local/etc/ntp.conf"} {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("读取时间源配置 %s 失败: %w", path, err)
		}
		if values, configured, err := readSourceConfig(ctx, []string{path}, false); configured || err != nil {
			return values, err
		}
	}
	if runtime.GOOS != "linux" {
		return nil, ctx.Err()
	}
	// 主文件取最高优先级；同名 drop-in 由高优先级目录覆盖，然后按文件名处理。
	directories := []string{"/etc/systemd", "/run/systemd", "/usr/local/lib/systemd", "/usr/lib/systemd", "/lib/systemd"}
	var paths []string
	for _, directory := range directories {
		path := filepath.Join(directory, "timesyncd.conf")
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("读取时间源配置 %s 失败: %w", path, err)
		}
		paths = append(paths, path)
		break
	}
	fragments := make(map[string]string)
	for _, directory := range directories {
		entries, err := os.ReadDir(filepath.Join(directory, "timesyncd.conf.d"))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("读取时间源配置目录失败: %w", err)
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".conf") {
				if _, exists := fragments[entry.Name()]; !exists {
					fragments[entry.Name()] = filepath.Join(directory, "timesyncd.conf.d", entry.Name())
				}
			}
		}
	}
	names := make([]string, 0, len(fragments))
	for name := range fragments {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		paths = append(paths, fragments[name])
	}
	values, _, err := readSourceConfig(ctx, paths, true)
	return values, err
}

// readSourceConfig 读取 NTP 网络源和标准 include/confdir/sourcedir。
// 认证源、引用时钟与无法展开的宏明确报错，避免悄悄换成公共服务器。
func readSourceConfig(ctx context.Context, paths []string, systemd bool) ([]string, bool, error) {
	var values, fallback []string
	configured := false
	stack := make(map[string]bool)
	var read func(string) error
	read = func(path string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		path = filepath.Clean(path)
		if stack[path] {
			return fmt.Errorf("时间源配置循环引用: %s", path)
		}
		stack[path] = true
		defer delete(stack, path)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("读取时间源配置 %s 失败: %w", path, err)
		}
		scanner := bufio.NewScanner(strings.NewReader(strings.ReplaceAll(string(content), "\\\n", " ")))
		section := ""
		for line := 1; scanner.Scan(); line++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			text := strings.TrimSpace(scanner.Text())
			if text == "" || strings.ContainsRune("#;!", rune(text[0])) {
				continue
			}
			invalid := func(reason string) error {
				return fmt.Errorf("时间源配置 %s:%d 无法使用: %s", path, line, reason)
			}
			if systemd {
				if strings.HasPrefix(text, "[") {
					section = strings.TrimSpace(strings.Trim(text, "[]"))
					continue
				}
				if section != "Time" {
					continue
				}
				key, value, ok := strings.Cut(text, "=")
				key = strings.TrimSpace(key)
				if !ok {
					key = strings.Fields(key)[0]
				}
				if key != "NTP" && key != "FallbackNTP" {
					continue
				}
				configured = true
				if !ok {
					return invalid("缺少等号")
				}
				fields, err := sourceFields(value)
				if err != nil {
					return invalid(err.Error())
				}
				list := &values
				if key == "FallbackNTP" {
					list = &fallback
				}
				if len(fields) == 0 {
					*list = nil
				} else {
					*list = append(*list, fields...)
				}
				continue
			}
			fields, err := sourceFields(text)
			if err != nil {
				return invalid(err.Error())
			}
			if len(fields) == 0 {
				continue
			}
			switch strings.ToLower(fields[0]) {
			case "include", "includefile", "confdir", "sourcedir":
				if len(fields) < 2 {
					return invalid("缺少引用路径")
				}
				files := make(map[string]string)
				for _, pattern := range fields[1:] {
					if !filepath.IsAbs(pattern) {
						pattern = filepath.Join(filepath.Dir(path), pattern)
					}
					directory := fields[0] == "confdir" || fields[0] == "sourcedir"
					if directory {
						if _, err := os.ReadDir(pattern); err != nil && !errors.Is(err, os.ErrNotExist) {
							return invalid("读取引用目录失败: " + err.Error())
						}
						suffix := "*.conf"
						if fields[0] == "sourcedir" {
							suffix = "*.sources"
						}
						pattern = filepath.Join(pattern, suffix)
					}
					matches, err := filepath.Glob(pattern)
					if err != nil || (!directory && len(matches) == 0) {
						return invalid("引用路径无效或文件不存在: " + pattern)
					}
					for _, file := range matches {
						key := file
						if directory {
							key = filepath.Base(file)
						}
						if _, exists := files[key]; !exists {
							files[key] = file
						}
					}
				}
				names := make([]string, 0, len(files))
				for name := range files {
					names = append(names, name)
				}
				slices.Sort(names)
				for _, name := range names {
					if err := read(files[name]); err != nil {
						return err
					}
				}
			case "server", "servers", "pool", "peer":
				configured = true
				fields = fields[1:]
				if len(fields) > 0 && (fields[0] == "-4" || fields[0] == "-6") {
					fields = fields[1:]
				}
				if len(fields) == 0 || fields[0] == "" || strings.ContainsAny(fields[0], "$={} \t") {
					return invalid("缺少服务器地址或使用了不支持的地址表达式")
				}
				if slices.Contains(fields[1:], "noselect") || slices.Contains(fields[1:], "offline") {
					continue
				}
				address := fields[0]
				for index := 1; index < len(fields); index++ {
					switch strings.ToLower(fields[index]) {
					case "nts", "key", "autokey":
						return invalid("当前单次校时不支持 NTS 或密钥认证，请使用系统时间服务")
					case "port":
						if index+1 >= len(fields) {
							return invalid("缺少 NTP 端口")
						}
						index++
						port, err := strconv.Atoi(fields[index])
						if err != nil || port <= 0 || port > 65535 {
							return invalid("NTP 端口无效")
						}
						address = net.JoinHostPort(strings.Trim(address, "[]"), fields[index])
					}
				}
				if strings.HasPrefix(address, "127.127.") {
					return invalid("引用时钟不是可查询的网络时间源")
				}
				values = append(values, address)
			case "refclock", "sensor", "broadcastclient", "manycastclient":
				return invalid("当前单次校时不支持引用时钟或自动发现时间源")
			}
		}
		return scanner.Err()
	}
	for _, path := range paths {
		if err := read(path); err != nil {
			return nil, configured, err
		}
	}
	if len(values) == 0 {
		values = fallback
	}
	if configured && len(values) == 0 {
		return nil, true, errors.New("系统已配置时间源，但没有可用于单次校时的 NTP 服务器")
	}
	return values, configured, ctx.Err()
}

// sourceFields 仅拆分配置词项，保留带引号的路径，拒绝未闭合的表达式。
func sourceFields(line string) ([]string, error) {
	var fields []string
	var value strings.Builder
	var quote rune
	escaped := false
	for _, character := range line {
		if escaped {
			value.WriteRune(character)
			escaped = false
		} else if character == '\\' {
			escaped = true
		} else if quote != 0 {
			if character == quote {
				quote = 0
			} else {
				value.WriteRune(character)
			}
		} else if character == '"' || character == '\'' {
			quote = character
		} else if character == '#' || character == ';' || character == '!' {
			break
		} else if unicode.IsSpace(character) {
			if value.Len() > 0 {
				fields = append(fields, value.String())
				value.Reset()
			}
		} else {
			value.WriteRune(character)
		}
	}
	if quote != 0 || escaped {
		return nil, errors.New("引号或转义未闭合")
	}
	if value.Len() > 0 {
		fields = append(fields, value.String())
	}
	return fields, nil
}
