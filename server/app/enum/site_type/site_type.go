package site_type

import (
	"server/app/enum/type_module"
	"server/app/service"
)

// Map 网站分类数据
func Map() map[int64]string {
	data, err := service.Type.GetEnum(type_module.Site)
	if err != nil {
		return make(map[int64]string)
	}
	return data
}
