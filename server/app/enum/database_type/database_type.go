package database_type

import (
	"server/app/enum/type_module"
	"server/app/service"
)

// Map 数据库分类数据
func Map() map[int64]string {
	data, err := service.Type.GetEnum(type_module.Database)
	if err != nil {
		return make(map[int64]string)
	}
	return data
}
