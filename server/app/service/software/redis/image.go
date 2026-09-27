package redis

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/moby/sys/user"
)

// Image 查询已导入的镜像并生成 Redis 安装预览，不导入镜像或启动容器。
// 这里只处理 Redis 版本、配置及非 root 运行策略，启动元数据由 Manager 提供。
func (m *Manager) Image(id string) (*ImageInfo, error) {
	if err := m.available(); err != nil {
		return nil, err
	}
	info, err := m.container.Image(id)
	if err != nil {
		return nil, err
	}
	image := info.Metadata
	if image.OS != "linux" || image.Architecture != runtime.GOARCH {
		return nil, fmt.Errorf("镜像平台 %s/%s 与当前运行平台 linux/%s 不匹配", image.OS, image.Architecture, runtime.GOARCH)
	}
	result := &ImageInfo{
		ID: image.ID, Name: image.Name, Tags: image.Tags,
		OS: image.OS, Architecture: image.Architecture,
		User: image.User, Entrypoint: image.Entrypoint, WorkingDir: image.WorkingDir,
		ConfigPath: "/usr/local/etc/sinking-redis/redis.conf",
	}
	searchPath := "/usr/local/bin:/usr/bin:/bin"
	for _, value := range image.Env {
		key, value, _ := strings.Cut(value, "=")
		switch key {
		case "REDIS_VERSION":
			result.Version = value
		case "PATH":
			searchPath = value
		}
	}
	var major, minor int
	if _, err = fmt.Sscanf(result.Version, "%d.%d", &major, &minor); err != nil || major <= 0 || minor < 0 {
		return nil, errors.New("无法识别 Redis 版本，请使用包含 REDIS_VERSION 的官方镜像")
	}
	// 官方镜像的 CMD 为 redis-server，可附带镜像内的默认配置文件。
	// 不猜测 shell 包装或第三方初始化脚本的参数，避免启动时丢失必要逻辑。
	if len(image.Command) == 0 || len(image.Command) > 2 || path.Base(image.Command[0]) != "redis-server" {
		return nil, errors.New("无法识别 Redis 启动命令，请使用官方 Redis 镜像")
	}
	if result.WorkingDir == "" {
		result.WorkingDir = "/data"
	}
	if !path.IsAbs(result.WorkingDir) || path.Clean(result.WorkingDir) == "/" {
		return nil, errors.New("Redis 镜像需要独立的绝对工作目录")
	}
	result.WorkingDir = path.Clean(result.WorkingDir)
	root, err := os.OpenRoot(image.Rootfs)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	// 根据镜像 PATH 验证所需命令，保留原 ENTRYPOINT，由 Manager 统一组合执行。
	commands := []string{image.Command[0], "redis-cli"}
	if len(image.Entrypoint) > 0 {
		commands = append(commands, image.Entrypoint[0])
	}
	for _, command := range commands {
		candidates := []string{command}
		if !strings.Contains(command, "/") {
			candidates = nil
			for _, directory := range strings.Split(searchPath, ":") {
				if path.IsAbs(directory) {
					candidates = append(candidates, path.Join(directory, command))
				}
			}
		}
		found := false
		for _, candidate := range candidates {
			if !path.IsAbs(candidate) {
				candidate = path.Join(result.WorkingDir, candidate)
			}
			info, statErr := root.Stat(strings.TrimPrefix(candidate, "/"))
			if statErr == nil && info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0 {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("Redis 镜像缺少可执行文件 %s", command)
		}
	}
	// 官方入口脚本通常以 root 启动再降权；面板直接使用镜像中的 redis 用户。
	if name, _, _ := strings.Cut(result.User, ":"); name == "" || name == "root" || name == "0" {
		result.User = "redis"
	}
	passwd, err := root.Open("etc/passwd")
	if err != nil {
		return nil, fmt.Errorf("读取镜像运行用户失败: %w", err)
	}
	defer passwd.Close()
	group, err := root.Open("etc/group")
	if err != nil {
		return nil, fmt.Errorf("读取镜像运行用户组失败: %w", err)
	}
	defer group.Close()
	identity, err := user.GetExecUser(result.User, nil, passwd, group)
	if err != nil || identity.Uid <= 0 {
		return nil, errors.New("Redis 镜像缺少有效的非 root 运行用户")
	}
	result.DataPath = result.WorkingDir
	result.Config = m.defaultConfig(major, minor)
	// 某些官方版本带有模块配置。保留镜像专属指令，面板管理项使用安全默认值。
	if len(image.Command) == 2 {
		configPath := image.Command[1]
		if strings.HasPrefix(configPath, "-") {
			return nil, errors.New("Redis 镜像启动参数不是配置文件路径")
		}
		if !path.IsAbs(configPath) {
			configPath = path.Join(result.WorkingDir, configPath)
		}
		file, err := root.Open(strings.TrimPrefix(path.Clean(configPath), "/"))
		if err != nil {
			return nil, fmt.Errorf("读取镜像默认配置失败: %w", err)
		}
		info, err := file.Stat()
		if err != nil || !info.Mode().IsRegular() || info.Size() > redisConfigMaxSize {
			_ = file.Close()
			return nil, errors.New("镜像默认配置必须是不超过 256 KiB 的普通文件")
		}
		content, err := io.ReadAll(io.LimitReader(file, redisConfigMaxSize+1))
		_ = file.Close()
		if err != nil || len(content) > redisConfigMaxSize {
			return nil, errors.New("读取镜像默认配置失败或内容过大")
		}
		managed := make(map[string]bool)
		for line := range strings.SplitSeq(result.Config, "\n") {
			args, _ := m.configArguments(line)
			if len(args) > 0 {
				managed[args[0]] = true
			}
		}
		var preserved strings.Builder
		for line := range strings.SplitSeq(string(content), "\n") {
			args, err := m.configArguments(line)
			if err != nil {
				return nil, fmt.Errorf("镜像默认配置无效: %w", err)
			}
			if len(args) == 0 || managed[strings.ToLower(args[0])] {
				continue
			}
			preserved.WriteString(line + "\n")
		}
		result.Config = preserved.String() + result.Config
	}
	config, err := m.patchConfig(result.Config, map[string][]string{"dir": {result.DataPath}})
	if err != nil {
		return nil, err
	}
	result.Config = config.content
	result.Command = []string{image.Command[0], result.ConfigPath}
	return result, nil
}
