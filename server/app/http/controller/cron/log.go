package cron

import (
	"server/app/service"
	"server/app/util/page"
	"server/app/util/server"
)

func Log(c *server.Context) {
	pageInfo := page.ValidatePageDefault(c)
	type Form struct {
		Id int `json:"id" default:"" validate:"numeric,min=1" label:"记录ID"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	data, err := service.Cron.FindById(form.Id)
	if err != nil || data == nil {
		c.Error("获取失败")
	} else {
		c.SuccessWithData("获取成功", service.Cron.ReadLog(data.Id, pageInfo.Page, pageInfo.PageSize))
	}
}
