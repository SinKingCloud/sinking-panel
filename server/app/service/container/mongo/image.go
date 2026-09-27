package mongo

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
	image, err := m.container.Image(id)
	if err != nil {
		return nil, err
	}
	if image.OS != "linux" || image.Architecture != runtime.GOARCH {
		return nil, fmt.Errorf("镜像平台 %s/%s 与当前平台 linux/%s 不匹配", image.OS, image.Architecture, runtime.GOARCH)
	}
	command := "mongod"
	if len(image.Command) != 1 || path.Base(image.Command[0]) != command || len(image.Entrypoint) != 1 || (path.Base(image.Entrypoint[0]) != "docker-entrypoint.sh" && path.Base(image.Entrypoint[0]) != "entrypoint.sh") {
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
	versionKey := "MONGO_VERSION"
	result.Version = variables[versionKey]
	majorText, _, _ := strings.Cut(result.Version, ".")
	major, err := strconv.Atoi(majorText)
	if err != nil || major <= 0 {
		return nil, errors.New("镜像未提供可识别的数据库版本")
	}

	if major < 2 || major == 2 && !strings.HasPrefix(result.Version, "2.6.") {
		return nil, errors.New("请使用提供 YAML 配置和角色权限管理的官方 MongoDB 镜像（2.6 起）")
	}
	result.DataPath, result.Config = "/data/db", m.defaultConfig(result.Version)
	result.Command = []string{"sh", "/etc/sinking-database/entrypoint.sh", image.Entrypoint[0], image.Command[0], "--config", result.ConfigPath, "--dbpath", result.DataPath}

	return result, nil
}
