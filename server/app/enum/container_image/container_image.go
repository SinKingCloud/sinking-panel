package container_image

import "server/app/service"

// Map 已导入镜像的 ID 与名称、标签映射。
func Map() map[string]string {
	if service.Container == nil {
		return map[string]string{}
	}
	data, err := service.Container.GetImageNameMap()
	if err != nil {
		return map[string]string{}
	}
	return data
}
