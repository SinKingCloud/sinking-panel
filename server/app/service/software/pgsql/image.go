package pgsql

import (
	"errors"
	"fmt"
	"path"
	"runtime"
	"strconv"
	"strings"
)

// Image 读取已导入的官方镜像元数据，保留其 ENTRYPOINT 初始化流程。
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
		return nil, fmt.Errorf("镜像平台 %s/%s 与当前平台 linux/%s 不匹配", image.OS, image.Architecture, runtime.GOARCH)
	}
	if len(image.Command) != 1 || path.Base(image.Command[0]) != "postgres" || len(image.Entrypoint) != 1 {
		return nil, errors.New("镜像启动信息不匹配，请使用对应数据库的官方镜像")
	}
	if entrypoint := path.Base(image.Entrypoint[0]); entrypoint != "docker-entrypoint.sh" && entrypoint != "entrypoint.sh" {
		return nil, errors.New("镜像启动信息不匹配，请使用对应数据库的官方镜像")
	}
	if image.User != "" && image.User != "0" && image.User != "root" && image.User != "0:0" {
		return nil, errors.New("请使用保留官方入口初始化用户的数据库镜像")
	}
	result := &ImageInfo{ID: image.ID, Name: image.Name, Tags: image.Tags, OS: image.OS, Architecture: image.Architecture, Entrypoint: image.Entrypoint, ConfigPath: "/etc/sinking-database/server.conf"}
	variables := make(map[string]string)
	for _, env := range image.Env {
		key, value, _ := strings.Cut(env, "=")
		variables[key] = value
	}
	result.Version = variables["PG_VERSION"]
	if result.Version == "" {
		result.Version = variables["PG_MAJOR"]
	}
	version := m.versionNumber(result.Version)
	if version == 0 {
		return nil, errors.New("镜像未提供可识别的数据库版本")
	}
	result.DataPath = "/var/lib/postgresql/data"
	if version >= 180000 {
		result.DataPath = "/var/lib/postgresql"
	}
	pgdata := path.Clean(variables["PGDATA"])
	if pgdata != result.DataPath && !strings.HasPrefix(pgdata, result.DataPath+"/") {
		return nil, errors.New("镜像 PGDATA 不在官方持久化目录内")
	}
	result.Config, result.Command = m.defaultConfig(version), []string{image.Command[0], "-c", "config_file=" + result.ConfigPath}

	return result, nil
}

// versionNumber 保留 9.x 的次版本分界，10 起按主版本判断功能；兼容镜像中的包版本后缀。
func (m *Manager) versionNumber(version string) int {
	parts := strings.FieldsFunc(version, func(r rune) bool { return r < '0' || r > '9' })
	if len(parts) == 0 {
		return 0
	}
	major, _ := strconv.Atoi(parts[0])
	if major <= 0 {
		return 0
	}
	if major < 10 {
		if len(parts) < 2 {
			return 0
		}
		minor, _ := strconv.Atoi(parts[1])
		return major*10000 + minor*100
	}
	return major * 10000
}
