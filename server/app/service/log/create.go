package log

import (
	"server/app/enum/log_type"
	"server/app/model"
	"server/app/util/str"
)

// Create 插入数据
func (s *service) Create(ip string, types int, title string, content string) {
	if _, ok := log_type.Map()[types]; !ok {
		return
	}
	id := str.GetSnowWorkIns().GetId()
	go func(id int64, types int, ip string, title string, content string) {
		_ = s.repositoryLog.Create(&model.Log{
			Id:      id,
			Type:    types,
			Ip:      ip,
			Title:   title,
			Content: content,
		})
	}(id, types, ip, title, content)
}
