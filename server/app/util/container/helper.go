package container

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (m *Manager) ensureManagedDirectory(path string) error {
	parent := filepath.Dir(path)
	if parent != path {
		info, err := os.Lstat(parent)
		if errors.Is(err, os.ErrNotExist) {
			if err := m.ensureManagedDirectory(parent); err != nil {
				return err
			}
		} else if err != nil {
			return fmt.Errorf("读取容器受管目录失败: %w", err)
		} else if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("容器受管目录父路径不是普通目录: %s", parent)
		}
	}
	if err := os.Mkdir(path, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("创建容器受管目录失败: %w", err)
	}
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("读取容器受管目录失败: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("容器受管目录不能是符号链接: %s", path)
	}
	if !info.IsDir() {
		return fmt.Errorf("容器受管路径不是目录: %s", path)
	}
	return nil
}

func (m *Manager) validateManagedNode(path string, directory, allowMissing bool) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) && allowMissing {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("受管路径不能是符号链接: %s", path)
	}
	if directory && !info.IsDir() {
		return fmt.Errorf("受管路径不是目录: %s", path)
	}
	if !directory && !info.Mode().IsRegular() {
		return fmt.Errorf("受管路径不是普通文件: %s", path)
	}
	return nil
}

// defaultManagerOptions 返回容器运行时的默认配置。
func defaultManagerOptions() ManagerOptions {
	return ManagerOptions{
		ContainerLogMaxSize:    10 << 20, // 单份实例日志保留大小
		RuntimeLogInterval:     5 * time.Second,
		RuntimePollInterval:    time.Second,
		RuntimeFailureLimit:    3,
		RuntimeCleanupDelay:    2 * time.Second,
		StopGracePeriod:        10 * time.Second,
		ForceStopPeriod:        5 * time.Second,
		TerminalConsoleTimeout: 5 * time.Second,
		TerminalCleanupTimeout: 5 * time.Second,
	}
}

// normalizeManagerOptions 用默认值补齐未设置项，避免零值破坏定时器和超时控制。
func normalizeManagerOptions(options ManagerOptions) ManagerOptions {
	defaults := defaultManagerOptions()
	if options.ContainerLogMaxSize <= 0 {
		options.ContainerLogMaxSize = defaults.ContainerLogMaxSize
	}
	if options.RuntimeLogInterval <= 0 {
		options.RuntimeLogInterval = defaults.RuntimeLogInterval
	}
	if options.RuntimePollInterval <= 0 {
		options.RuntimePollInterval = defaults.RuntimePollInterval
	}
	if options.RuntimeFailureLimit <= 0 {
		options.RuntimeFailureLimit = defaults.RuntimeFailureLimit
	}
	if options.RuntimeCleanupDelay <= 0 {
		options.RuntimeCleanupDelay = defaults.RuntimeCleanupDelay
	}
	if options.StopGracePeriod <= 0 {
		options.StopGracePeriod = defaults.StopGracePeriod
	}
	if options.ForceStopPeriod <= 0 {
		options.ForceStopPeriod = defaults.ForceStopPeriod
	}
	if options.TerminalConsoleTimeout <= 0 {
		options.TerminalConsoleTimeout = defaults.TerminalConsoleTimeout
	}
	if options.TerminalCleanupTimeout <= 0 {
		options.TerminalCleanupTimeout = defaults.TerminalCleanupTimeout
	}
	return options
}

// managerOptions 返回当前 Manager 的有效配置。
// 允许测试和嵌入方直接构造 Manager 时仍然获得安全默认值。
func (m *Manager) managerOptions() ManagerOptions {
	return normalizeManagerOptions(m.options)
}
func (m *Manager) validateID(id string) error {
	if len(id) == 0 || len(id) > 64 || id == "." || id == ".." {
		return errors.New("实例ID不合法")
	}
	for index := 0; index < len(id); index++ {
		value := id[index]
		letter := value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z'
		digit := value >= '0' && value <= '9'
		if index == 0 && !letter && !digit || index > 0 && !letter && !digit && value != '-' && value != '_' && value != '.' {
			return errors.New("实例ID只能包含字母、数字、点、横线和下划线，并且必须以字母或数字开头")
		}
	}
	return nil
}

func (m *Manager) validateImageID(id string) error {
	digest := strings.TrimPrefix(id, "sha256:")
	if !strings.HasPrefix(id, "sha256:") || len(digest) != 64 {
		return errors.New("镜像ID不合法")
	}
	for index := 0; index < len(digest); index++ {
		value := digest[index]
		if !(value >= '0' && value <= '9') && !(value >= 'a' && value <= 'f') {
			return errors.New("镜像ID不合法")
		}
	}
	return nil
}

func (m *Manager) validateMount(mount Mount) error {
	if !filepath.IsAbs(mount.Source) || !filepath.IsAbs(mount.Destination) {
		return errors.New("挂载路径必须是绝对路径")
	}
	if filepath.Clean(mount.Destination) == string(filepath.Separator) {
		return errors.New("挂载目标不能是容器根目录")
	}
	info, err := os.Lstat(mount.Source)
	if err != nil {
		return fmt.Errorf("挂载源不存在: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errors.New("挂载源不能是符号链接")
	}
	return nil
}

func (m *Manager) writeJSON(path string, value interface{}) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".metadata-*.tmp")
	if err != nil {
		return err
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	if err := file.Chmod(0600); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}

func readCursorLog(data []byte, fileName string, after int64, before int64, pageSize int) map[string]interface{} {
	if pageSize < 1 {
		pageSize = 300
	}
	if pageSize > 10000 {
		pageSize = 10000
	}
	if after < 0 {
		after = 0
	}
	if before < 0 {
		before = 0
	}
	result := map[string]interface{}{
		"file_name":    fileName,
		"lines":        []string{},
		"cursor":       after,
		"start_cursor": after,
		"has_previous": after > 0,
		"end":          false,
	}
	if len(data) == 0 {
		result["cursor"] = int64(0)
		result["start_cursor"] = int64(0)
		result["has_previous"] = false
		return result
	}

	readBefore := func(end int64) ([]string, []int64) {
		if end < 0 {
			end = 0
		}
		if end > int64(len(data)) {
			end = int64(len(data))
		}
		lines := make([]string, 0, pageSize)
		starts := make([]int64, 0, pageSize)
		start := 0
		for start < int(end) {
			relEnd := bytes.IndexByte(data[start:end], '\n')
			if relEnd < 0 {
				if end == int64(len(data)) {
					lines = append(lines, strings.TrimSuffix(strings.ToValidUTF8(string(data[start:end]), "�"), "\r"))
					starts = append(starts, int64(start))
				}
				break
			}
			lineEnd := start + relEnd
			lines = append(lines, strings.TrimSuffix(strings.ToValidUTF8(string(data[start:lineEnd]), "�"), "\r"))
			starts = append(starts, int64(start))
			start = lineEnd + 1
		}
		if len(lines) > pageSize {
			first := len(lines) - pageSize
			lines = lines[first:]
			starts = starts[first:]
		}
		return lines, starts
	}

	readAfter := func(start int64) ([]string, int64) {
		if start < 0 {
			start = 0
		}
		if start >= int64(len(data)) {
			return []string{}, start
		}
		lines := make([]string, 0, pageSize)
		next := start
		for len(lines) < pageSize && next < int64(len(data)) {
			relEnd := bytes.IndexByte(data[next:], '\n')
			if relEnd < 0 {
				break
			}
			lineEnd := next + int64(relEnd)
			lines = append(lines, strings.TrimSuffix(strings.ToValidUTF8(string(data[next:lineEnd]), "�"), "\r"))
			next = lineEnd + 1
		}
		return lines, next
	}

	if after > 0 {
		if after > int64(len(data)) {
			lines, starts := readBefore(int64(len(data)))
			result["lines"] = lines
			result["cursor"] = int64(len(data))
			result["end"] = true
			if len(starts) > 0 {
				result["start_cursor"] = starts[0]
				result["has_previous"] = starts[0] > 0
			}
			return result
		}
		lines, next := readAfter(after)
		result["lines"] = lines
		result["cursor"] = next
		result["start_cursor"] = after
		result["has_previous"] = after > 0
		result["end"] = next >= int64(len(data))
		return result
	}

	end := int64(len(data))
	if before > 0 {
		end = before
	}
	lines, starts := readBefore(end)
	result["lines"] = lines
	result["cursor"] = int64(len(data))
	if len(starts) > 0 {
		result["start_cursor"] = starts[0]
		result["has_previous"] = starts[0] > 0
		result["end"] = starts[0] <= 0
	} else {
		result["start_cursor"] = int64(0)
		result["has_previous"] = false
		result["end"] = true
	}
	return result
}

func (m *Manager) readLogData(paths []string, maxBytes int64) ([]byte, error) {
	type segment struct {
		path string
		size int64
	}
	segments := make([]segment, 0, len(paths))
	var total int64
	for _, path := range paths {
		info, err := os.Stat(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("实例日志不是普通文件: %s", path)
		}
		segments = append(segments, segment{path: path, size: info.Size()})
		total += info.Size()
	}
	skip := int64(0)
	if maxBytes > 0 && total > maxBytes {
		skip = total - maxBytes
	}
	var output bytes.Buffer
	for _, item := range segments {
		if skip >= item.size {
			skip -= item.size
			continue
		}
		file, err := os.Open(item.path)
		if err != nil {
			return nil, err
		}
		if skip > 0 {
			if _, err := file.Seek(skip, io.SeekStart); err != nil {
				_ = file.Close()
				return nil, err
			}
		}
		length := item.size - skip
		skip = 0
		_, copyErr := io.CopyN(&output, file, length)
		closeErr := file.Close()
		if copyErr != nil && !errors.Is(copyErr, io.EOF) || closeErr != nil {
			return nil, errors.Join(copyErr, closeErr)
		}
	}
	return append([]byte(nil), output.Bytes()...), nil
}

func (m *Manager) readJSON(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func (m *Manager) cloneImage(image *Image) *Image {
	if image == nil {
		return nil
	}
	copyImage := *image
	copyImage.Tags = append([]string(nil), image.Tags...)
	copyImage.Entrypoint = append([]string(nil), image.Entrypoint...)
	copyImage.Command = append([]string(nil), image.Command...)
	copyImage.Env = append([]string(nil), image.Env...)
	return &copyImage
}

func (m *Manager) cloneInstance(instance *Instance) *Instance {
	if instance == nil {
		return nil
	}
	copyInstance := *instance
	copyInstance.Mounts = append([]Mount(nil), instance.Mounts...)
	copyInstance.Command = append([]string(nil), instance.Command...)
	copyInstance.Env = append([]string(nil), instance.Env...)
	return &copyInstance
}
