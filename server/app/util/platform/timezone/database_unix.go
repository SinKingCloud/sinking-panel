//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package timezone

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const maxZoneFileSize = 1 << 20

// zoneDatabase 保留稳定的系统目录与解析后的真实目录，分别用于链接和文件校验。
type zoneDatabase struct {
	directory string
	root      string
}

// openDatabase 定位本机时区库，不缓存目录，允许系统更新 tzdata 后立即使用新数据。
func openDatabase() (*zoneDatabase, error) {
	for _, path := range []string{"/usr/share/zoneinfo", "/etc/zoneinfo", "/var/db/timezone/zoneinfo", "/usr/share/lib/zoneinfo", "/usr/lib/zoneinfo"} {
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			root, err := filepath.EvalSymlinks(path)
			if err != nil {
				return nil, err
			}
			return &zoneDatabase{directory: path, root: root}, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("读取系统时区目录失败: %w", err)
		}
	}
	return nil, errors.New("系统未安装时区数据库，请先安装 tzdata")
}

// read 只读取库内的标准 TZif 文件，返回稳定路径，避免链接固定到旧版本时区库。
func (d *zoneDatabase) read(name string) (string, []byte, error) {
	if name == "" || strings.ContainsAny(name, "\\\x00") || filepath.IsAbs(name) || filepath.ToSlash(filepath.Clean(name)) != name || name == "." || name == ".." || strings.HasPrefix(name, "../") {
		return "", nil, errors.New("时区名称不合法")
	}
	if name == "localtime" || name == "posixrules" || strings.HasPrefix(name, "posix/") || strings.HasPrefix(name, "right/") {
		return "", nil, errors.New("请选择标准时区名称")
	}
	path := filepath.Join(d.directory, filepath.FromSlash(name))
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", nil, fmt.Errorf("读取时区 %q 失败: %w", name, err)
	}
	relative, err := filepath.Rel(d.root, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", nil, errors.New("时区文件不在系统时区目录内")
	}
	file, err := os.Open(resolved)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxZoneFileSize {
		return "", nil, errors.New("无效的时区文件")
	}
	data, err := io.ReadAll(io.LimitReader(file, maxZoneFileSize+1))
	if err != nil {
		return "", nil, err
	}
	if len(data) > maxZoneFileSize {
		return "", nil, errors.New("时区文件过大")
	}
	if _, err := time.LoadLocationFromTZData(name, data); err != nil {
		return "", nil, fmt.Errorf("时区 %q 的数据无效: %w", name, err)
	}
	return path, data, nil
}

func (d *zoneDatabase) list() ([]string, error) {
	names := make([]string, 0, 600)
	// WalkDir 不跟随根目录链接，macOS 的 zoneinfo 本身就是链接。
	err := filepath.WalkDir(d.root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(d.root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if relative == "posix" || relative == "right" {
				return filepath.SkipDir
			}
			return nil
		}
		name := filepath.ToSlash(relative)
		if _, _, err := d.read(name); err == nil {
			names = append(names, name)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("读取时区列表失败: %w", err)
	}
	if len(names) == 0 {
		return nil, errors.New("系统时区数据库中没有有效时区")
	}
	slices.Sort(names)
	return slices.Compact(names), nil
}
