//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package timezone

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// List 返回本机支持的 IANA 时区名称，按名称排序，无需管理员权限。
func List() ([]string, error) {
	database, err := openDatabase()
	if err != nil {
		return nil, err
	}
	return database.list()
}

// Get 读取系统时区和当前当地时间，不受进程 TZ 环境变量或 time.Local 缓存影响。
func Get() (string, time.Time, error) {
	current := func(name string, data []byte) (string, time.Time, error) {
		location, err := time.LoadLocationFromTZData(name, data)
		if err != nil {
			return "", time.Time{}, fmt.Errorf("读取系统时区数据失败: %w", err)
		}
		return name, time.Now().In(location), nil
	}
	database, err := openDatabase()
	if err != nil {
		return "", time.Time{}, err
	}
	if target, err := os.Readlink(localtimePath); err == nil {
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(localtimePath), target)
		}
		// 保留 UTC、US/Eastern 等用户实际选择的别名。
		for _, base := range []string{database.directory, database.root} {
			if name, err := filepath.Rel(base, target); err == nil {
				name = filepath.ToSlash(name)
				if _, data, err := database.read(name); err == nil {
					return current(name, data)
				}
			}
		}
	}
	if target, err := filepath.EvalSymlinks(localtimePath); err == nil {
		if name, err := filepath.Rel(database.root, target); err == nil {
			name = filepath.ToSlash(name)
			if _, data, err := database.read(name); err == nil {
				return current(name, data)
			}
		}
	}
	data, err := os.ReadFile(localtimePath)
	if errors.Is(err, os.ErrNotExist) {
		return "UTC", time.Now().UTC(), nil
	}
	if err != nil {
		return "", time.Time{}, fmt.Errorf("读取系统时区失败: %w", err)
	}
	// 名称文件可能陈旧，只有与实际 TZif 内容相符才采用。
	for _, path := range []string{timezoneNamePath, "/var/db/zoneinfo"} {
		if value, err := os.ReadFile(path); err == nil {
			name := strings.TrimSpace(string(value))
			if _, expected, err := database.read(name); err == nil && bytes.Equal(data, expected) {
				return current(name, data)
			}
		}
	}
	// 只有无法从链接或有效名称记录识别的副本才需要遍历数据库。
	names, err := database.list()
	if err != nil {
		return "", time.Time{}, err
	}
	for _, name := range names {
		_, expected, err := database.read(name)
		if err == nil && bytes.Equal(data, expected) {
			return current(name, data)
		}
	}
	return "", time.Time{}, errors.New("系统使用自定义时区文件，无法识别对应的时区名称")
}
