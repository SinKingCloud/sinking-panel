package site_name

import "server/app/service"

// Map 可选默认站点名称映射。
func Map() map[int64]string {
	data, err := service.Site.GetIdNameMap(false)
	if err != nil {
		return map[int64]string{}
	}
	return data
}
