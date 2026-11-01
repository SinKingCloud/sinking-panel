package oss

import (
	"server/app/enum/log_type"
	repositoryOss "server/app/repository/oss"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// List 按名称和关联密钥分页查询 OSS 存储配置。
func List(c *context.Context) {
	query := c.ValidatePage("id", "desc", "id", "id,name,create_time,update_time")
	var form struct {
		Keyword  string `json:"keyword" default:"" validate:"omitempty,max=100" label:"关键词"`
		SecretId string `json:"secret_id" default:"" validate:"omitempty,numeric" label:"密钥ID"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	where := &repositoryOss.SelectOss{}
	if form.Keyword != "" {
		where.Keyword = &form.Keyword
	}
	if form.SecretId != "" {
		secretId, err := strconv.ParseInt(form.SecretId, 10, 64)
		if err != nil || secretId <= 0 {
			c.Error("密钥参数错误")
			return
		}
		where.SecretId = &secretId
	}
	data, err := service.Oss.Select(where, query)
	if err != nil {
		c.Error("获取失败")
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看 OSS 列表", "查看 OSS 列表数据")
		c.SuccessWithData("获取成功", data)
	}
}
