package server

import (
	"server/app/model"
	"server/app/util/str"
	"time"
)

// UpdateByIds 通过ID列表更新
func (r *Repository) UpdateByIds(ids []int64, data *UpdateServer) error {
	if len(ids) == 0 || data == nil {
		return nil
	}
	updates := make(map[string]interface{})
	if data.Ip != nil {
		updates["ip"] = data.Ip
	}
	if data.Port != nil {
		updates["port"] = data.Port
	}
	if data.User != nil {
		updates["user"] = data.User
	}
	if data.AuthType != nil {
		updates["auth_type"] = data.AuthType
	}
	if data.Password != nil {
		updates["password"] = data.Password
	}
	if data.Name != nil {
		updates["name"] = data.Name
	}
	if len(updates) == 0 {
		return nil
	}
	updates["update_time"] = str.DateTime(time.Now())
	return r.Database.Db.Model(&model.Server{}).Where("id IN ?", ids).Updates(updates).Error
}
