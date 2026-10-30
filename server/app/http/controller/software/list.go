package software

import (
	"server/app/enum/log_type"
	repositorySoftware "server/app/repository/software"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"strings"
)

// List 分页查询已安装的软件。
func List(c *context.Context) {
	query := c.ValidatePage("id", "desc", "id", "id,name,code,version,status,state,create_time,update_time")
	var form struct {
		Keyword string `json:"keyword" default:"" validate:"omitempty,max=100" label:"关键词"`
		Name    string `json:"name" default:"" validate:"omitempty,max=100" label:"软件名称"`
		Code    string `json:"code" default:"" validate:"omitempty,max=100" label:"软件标识"`
		State   string `json:"state" default:"" validate:"omitempty,numeric" label:"安装状态"`
		Status  string `json:"status" default:"" validate:"omitempty,numeric" label:"启用状态"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	where := &repositorySoftware.SelectSoftware{}
	if keyword := strings.TrimSpace(form.Keyword); keyword != "" {
		where.Keyword = &keyword
	}
	if form.Name != "" {
		where.Name = &form.Name
	}
	if form.Code != "" {
		where.Code = &form.Code
	}
	if form.Status != "" {
		status, err := strconv.Atoi(form.Status)
		if err != nil {
			c.Error("启用状态参数错误")
			return
		}
		where.Status = &status
	}
	if form.State != "" {
		state, err := strconv.Atoi(form.State)
		if err != nil {
			c.Error("安装状态参数错误")
			return
		}
		where.State = &state
	}
	data, err := service.Software.Select(where, query)
	if err != nil {
		c.Error(err.Error())
		return
	}
	service.Log.Create(c.GetRequestIp(), log_type.EventShow, "查看软件列表", "查看软件管理列表")
	c.SuccessWithData("获取成功", data)
}
