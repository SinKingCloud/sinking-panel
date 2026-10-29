package system

import (
	"fmt"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Monitor 查询监控历史、读取或修改历史采集设置。
func Monitor(c *context.Context) {
	var form struct {
		Action        string `json:"action" default:"history" validate:"required,oneof=settings set history" label:"操作类型"`
		Enabled       *bool  `json:"enabled" label:"开启监控"`
		RetentionDays int    `json:"retention_days" validate:"omitempty,min=1,max=365" label:"保留天数"`
		Start         int64  `json:"start" validate:"omitempty,min=1" label:"开始时间"`
		End           int64  `json:"end" validate:"omitempty,min=1" label:"结束时间"`
		MaxPoints     *int   `json:"max_points" validate:"omitempty,min=120,max=720" label:"查询点数"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	message := "获取监控设置成功"
	switch form.Action {
	case "history":
		// 通用绑定会忽略文本转整数失败；显式无效值不能误用默认查询精度。
		for _, values := range []map[string]string{c.AllQuery(), c.AllForm()} {
			if raw, exists := values["max_points"]; exists {
				if _, err := strconv.Atoi(raw); err != nil {
					c.Error("查询点数必须为整数")
					return
				}
			}
		}
		var maxPoints []int
		if form.MaxPoints != nil {
			maxPoints = []int{*form.MaxPoints}
		}
		data, err := service.System.GetMonitorHistory(form.Start, form.End, maxPoints...)
		if err != nil {
			c.Error(err.Error())
		} else {
			c.SuccessWithData("获取监控历史成功", data)
		}
		return
	case "set":
		if form.Enabled == nil || form.RetentionDays < 1 {
			c.Error("请指定是否开启监控及保留天数")
			return
		}
		if err := service.System.SetMonitorSettings(*form.Enabled, form.RetentionDays); err != nil {
			c.Error(err.Error())
			return
		}
		state := "关闭"
		if *form.Enabled {
			state = "开启"
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "修改系统监控", fmt.Sprintf("%s历史监控，保留 %d 天", state, form.RetentionDays))
		message = "保存监控设置成功"
	case "settings":
		// 与保存设置共用下方的配置读取及返回逻辑。
	default:
		c.Error("不支持该操作")
		return
	}
	data, err := service.System.GetMonitorSettings()
	if err != nil {
		c.Error(err.Error())
	} else {
		c.SuccessWithData(message, data)
	}
}
