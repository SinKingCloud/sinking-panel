package secret

import (
	"server/app/enum/log_type"
	repositorySecret "server/app/repository/secret"
	"server/app/service"
	"server/app/util/context"
	"strings"
)

// List 获取密钥凭据列表，不返回密钥正文。
func List(c *context.Context) {
	query := c.ValidatePage("id", "desc", "id", "id,name,provider,create_time,update_time")
	var form struct {
		Keyword  string `json:"keyword" default:"" validate:"omitempty,max=100" label:"关键词"`
		Provider string `json:"provider" default:"" validate:"omitempty,numeric" label:"服务商"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	where := &repositorySecret.SelectSecret{}
	if form.Keyword != "" {
		where.Keyword = strings.TrimSpace(form.Keyword)
	}
	if form.Provider != "" {
		where.Provider = form.Provider
	}
	data, err := service.Secret.Select(where, query)
	if err != nil {
		c.Error("获取失败")
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看密钥列表", "查看密钥列表数据")
		c.SuccessWithData("获取成功", data)
	}
}
