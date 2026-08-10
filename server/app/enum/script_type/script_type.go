package script_type

import (
	"server/app/enum/type_module"
	"server/app/service"
)

// Map 脚本类型数据
func Map() map[int64]string {
	data, err := service.Type.GetEnum(type_module.Script)
	if err != nil {
		return map[int64]string{}
	}
	return data
}
