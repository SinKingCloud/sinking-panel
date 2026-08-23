package container

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
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
	if err := os.Mkdir(path, 0700); err != nil && !errors.Is(err, os.ErrExist) {
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
		AutoRestartDelay:       time.Second,
		StopGracePeriod:        10 * time.Second,
		ForceStopPeriod:        5 * time.Second,
		TerminalConsoleTimeout: 5 * time.Second,
		TerminalCleanupTimeout: 5 * time.Second,
		DefaultPidsLimit:       512,
		IDMapSize:              defaultIDMapSize,
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
	if options.AutoRestartDelay <= 0 {
		options.AutoRestartDelay = defaults.AutoRestartDelay
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
	if options.DefaultPidsLimit == 0 || options.DefaultPidsLimit < -1 {
		options.DefaultPidsLimit = defaults.DefaultPidsLimit
	}
	if options.IDMapSize == 0 {
		options.IDMapSize = defaults.IDMapSize
	}
	options.MountRoots = append([]string(nil), options.MountRoots...)
	return options
}

// managerOptions 返回当前 Manager 的有效配置。
// 允许测试和嵌入方直接构造 Manager 时仍然获得安全默认值。
func (m *Manager) managerOptions() ManagerOptions {
	return normalizeManagerOptions(m.options)
}

// mergeEnv 按组依次覆盖同名变量，保持镜像、实例和命令的优先级顺序。
func (m *Manager) mergeEnv(groups ...[]string) []string {
	result := make([]string, 0)
	positions := make(map[string]int)
	for _, group := range groups {
		for _, value := range group {
			key, _, ok := strings.Cut(value, "=")
			if !ok || key == "" {
				continue
			}
			if position, exists := positions[key]; exists {
				result[position] = value
				continue
			}
			positions[key] = len(result)
			result = append(result, value)
		}
	}
	return result
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

func (m *Manager) resolveMountSource(source string) (string, error) {
	source = strings.TrimSpace(source)
	if !filepath.IsAbs(source) {
		return "", errors.New("挂载源必须是绝对路径")
	}
	source = filepath.Clean(source)
	info, err := os.Lstat(source)
	if err != nil {
		return "", fmt.Errorf("挂载源不存在: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("挂载源不能是符号链接")
	}
	resolved, err := filepath.EvalSymlinks(source)
	if err != nil {
		return "", fmt.Errorf("解析挂载源失败: %w", err)
	}
	return filepath.Clean(resolved), nil
}

func (m *Manager) validateMount(mount Mount) error {
	if strings.TrimSpace(mount.DirectoryMode) != "" || strings.TrimSpace(mount.FileMode) != "" {
		return errors.New("挂载权限自动修改已停用，请在 PrepareMount 前由业务层设置目录和文件权限")
	}
	if !filepath.IsAbs(mount.Source) || !path.IsAbs(mount.Destination) {
		return errors.New("挂载路径必须是绝对路径")
	}
	if strings.ContainsRune(mount.Source, 0) || strings.ContainsRune(mount.Destination, 0) {
		return errors.New("挂载路径不能包含空字符")
	}
	destination := path.Clean(mount.Destination)
	if destination == "/" {
		return errors.New("挂载目标不能是容器根目录")
	}
	for _, protected := range []string{"/proc", "/sys", "/dev"} {
		if destination == protected || strings.HasPrefix(destination, protected+"/") {
			return fmt.Errorf("挂载目标不能覆盖容器安全目录: %s", protected)
		}
	}
	info, err := os.Lstat(mount.Source)
	if err != nil {
		return fmt.Errorf("挂载源不存在: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errors.New("挂载源不能是符号链接")
	}
	if !info.IsDir() && !info.Mode().IsRegular() {
		return errors.New("挂载源只能是普通文件或目录")
	}
	resolved, err := filepath.EvalSymlinks(mount.Source)
	if err != nil {
		return fmt.Errorf("解析挂载源失败: %w", err)
	}
	if filepath.Clean(resolved) != filepath.Clean(mount.Source) {
		return errors.New("挂载源必须使用解析符号链接后的规范路径")
	}
	allowed := false
	for _, root := range m.mountRoots {
		if filepath.Clean(root) != filepath.Clean(mount.Source) && m.ensureInside(root, mount.Source) == nil {
			allowed = true
			break
		}
	}
	if !allowed {
		return errors.New("挂载源不在 ManagerOptions.MountRoots 允许的目录内")
	}
	for _, protected := range []string{
		m.imagesRoot,
		m.instancesRoot,
		m.runtimeRoot,
		m.mountsRoot,
		filepath.Join(m.dataRoot, "security.json"),
		filepath.Join(m.dataRoot, ".mount-staging.lock"),
	} {
		if m.ensureInside(protected, mount.Source) == nil || m.ensureInside(mount.Source, protected) == nil {
			return errors.New("挂载源不能与容器受管数据目录重叠")
		}
	}
	return nil
}

func (m *Manager) mapOwnership(uid, gid int) (int, int, error) {
	if m.security.Version == 0 {
		return uid, gid, nil
	}
	convert := func(value int, start int64, name string) (int, error) {
		if value < 0 || int64(value) >= m.security.IDMapSize {
			return 0, fmt.Errorf("容器 %s %d 超出 User Namespace 映射范围", name, value)
		}
		mapped := start + int64(value)
		if int64(int(mapped)) != mapped {
			return 0, fmt.Errorf("容器 %s %d 无法转换为宿主 ID", name, value)
		}
		return int(mapped), nil
	}
	mappedUID, err := convert(uid, m.security.UIDMapStart, "UID")
	if err != nil {
		return 0, 0, err
	}
	mappedGID, err := convert(gid, m.security.GIDMapStart, "GID")
	if err != nil {
		return 0, 0, err
	}
	return mappedUID, mappedGID, nil
}

func (m *Manager) unmapOwnership(uid, gid int) (int, int, error) {
	if m.security.Version == 0 {
		return uid, gid, nil
	}
	convert := func(value int, start int64, name string) (int, error) {
		mapped := int64(value) - start
		if mapped < 0 || mapped >= m.security.IDMapSize {
			return 0, fmt.Errorf("宿主 %s %d 不属于当前容器映射范围", name, value)
		}
		return int(mapped), nil
	}
	containerUID, err := convert(uid, m.security.UIDMapStart, "UID")
	if err != nil {
		return 0, 0, err
	}
	containerGID, err := convert(gid, m.security.GIDMapStart, "GID")
	if err != nil {
		return 0, 0, err
	}
	return containerUID, containerGID, nil
}

func (m *Manager) validateImageSecurity(image *Image) error {
	if m.security.Version == 0 {
		return nil
	}
	if image.SecurityVersion != m.security.Version || image.UIDMapStart != m.security.UIDMapStart ||
		image.GIDMapStart != m.security.GIDMapStart || image.IDMapSize != m.security.IDMapSize {
		return errors.New("镜像未使用当前 User Namespace 映射，请删除旧镜像后重新导入")
	}
	return nil
}

func (m *Manager) validateInstanceSecurity(instance *Instance) error {
	if m.security.Version == 0 {
		return nil
	}
	if instance.SecurityVersion != m.security.Version || instance.UIDMapStart != m.security.UIDMapStart ||
		instance.GIDMapStart != m.security.GIDMapStart || instance.IDMapSize != m.security.IDMapSize {
		return errors.New("实例未使用当前 User Namespace 映射，请删除旧实例后重新创建")
	}
	return nil
}

func (m *Manager) writeJSON(path string, value interface{}) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
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
