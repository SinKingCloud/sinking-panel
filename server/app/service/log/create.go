package log

import (
	"fmt"
	"server/app/enum/log_type"
	"server/app/model"
	ip2 "server/app/util/ip"
	"server/app/util/str"
	"strings"
)

// Create 插入数据
func (s *service) Create(ip string, types int, title string, content string) {
	if _, ok := log_type.Map()[types]; !ok {
		return
	}
	location := "未知"
	ipInfo, err := ip2.Query(ip)
	if err == nil && ipInfo != nil {
		location = strings.ReplaceAll(fmt.Sprintf("%s %s %s %s", ipInfo.Country, ipInfo.Province, ipInfo.City, ipInfo.Isp), "  ", "")
	}
	_ = s.repositoryLog.Create(&model.Log{
		Id:       str.GetSnowWorkIns().GetId(),
		Type:     types,
		Ip:       ip,
		Location: location,
		Title:    title,
		Content:  content,
	})
}
