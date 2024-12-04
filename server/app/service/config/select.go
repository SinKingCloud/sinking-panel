package config

import (
	"server/app/model"
	"server/app/util"
)

// SelectByGroup 查看配置数据
func (*Service) SelectByGroup(group string) map[string]string {
	var configs []*model.Config
	temp := make(map[string]string)
	query := util.Database.Db.Model(&model.Config{})
	if group != "" {
		query.Where("`key` like ?", group+"%")
	}
	query.Find(&configs)
	for _, v := range configs {
		temp[v.Key] = v.Value
	}
	return temp
}
