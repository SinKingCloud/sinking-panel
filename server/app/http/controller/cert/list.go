package cert

import (
	"server/app/enum/log_type"
	repositoryCert "server/app/repository/cert"
	"server/app/service"
	"server/app/util/context"
)

// List 获取证书列表，列表不返回证书和私钥正文。
func List(c *context.Context) {
	query := c.ValidatePage("id", "desc", "id", "id,name,type,start_time,expire_time,create_time,update_time")
	var form struct {
		Keyword         string `json:"keyword" default:"" validate:"omitempty,max=100" label:"关键词"`
		Name            string `json:"name" default:"" validate:"omitempty,max=100" label:"证书名称"`
		Type            string `json:"type" default:"" validate:"omitempty,numeric" label:"证书类型"`
		ExpireTimeStart string `json:"expire_time_start" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"到期起始时间"`
		ExpireTimeEnd   string `json:"expire_time_end" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"到期结束时间"`
		CreateTimeStart string `json:"create_time_start" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"创建起始时间"`
		CreateTimeEnd   string `json:"create_time_end" default:"" validate:"omitempty,datetime=2006-01-02 15:04:05" label:"创建结束时间"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	where := &repositoryCert.SelectCert{}
	if form.Keyword != "" {
		where.Keyword = form.Keyword
	}
	if form.Name != "" {
		where.Name = form.Name
	}
	if form.Type != "" {
		where.Type = form.Type
	}
	if form.ExpireTimeStart != "" {
		where.ExpireTimeStart = form.ExpireTimeStart
	}
	if form.ExpireTimeEnd != "" {
		where.ExpireTimeEnd = form.ExpireTimeEnd
	}
	if form.CreateTimeStart != "" {
		where.CreateTimeStart = form.CreateTimeStart
	}
	if form.CreateTimeEnd != "" {
		where.CreateTimeEnd = form.CreateTimeEnd
	}
	data, err := service.Site.SelectCert(where, query)
	if err != nil {
		c.Error("获取失败")
	} else {
		service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看证书列表", "查看网站证书列表数据")
		c.SuccessWithData("获取成功", data)
	}
}
