package secret_name

import "server/app/service"

// Map 密钥 ID 与名称映射，不包含密钥正文。
func Map() map[int64]string {
	if service.Secret == nil {
		return map[int64]string{}
	}
	data, err := service.Secret.GetIdNameMap(nil)
	if err != nil {
		return map[int64]string{}
	}
	return data
}
