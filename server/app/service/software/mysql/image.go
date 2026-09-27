package mysql

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
	command := "mysqld"
	if len(image.Command) != 1 || path.Base(image.Command[0]) != command || len(image.Entrypoint) != 1 || path.Base(image.Entrypoint[0]) != "docker-entrypoint.sh" && path.Base(image.Entrypoint[0]) != "entrypoint.sh" {
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
	versionKey := "MYSQL_VERSION"
	result.Version = variables[versionKey]
	majorText, _, _ := strings.Cut(result.Version, ".")
	major, err := strconv.Atoi(majorText)
	if err != nil || major <= 0 {
		return nil, errors.New("镜像未提供可识别的数据库版本")
	}

	result.DataPath, result.Config = "/var/lib/mysql", m.defaultConfig()
	// 旧入口不认识 _FILE；从只读文件加载初始化密码后仍交回原入口初始化。
	result.Command = []string{"sh", "-c", `MYSQL_ROOT_PASSWORD="$(cat /run/sinking-database/bootstrap-password)"; export MYSQL_ROOT_PASSWORD; unset MYSQL_ROOT_PASSWORD_FILE; exec "$@"`, "mysql-entrypoint", image.Entrypoint[0], image.Command[0], "--defaults-file=" + result.ConfigPath}

	return result, nil
}
