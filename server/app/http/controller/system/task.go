package system

import (
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/context"
	"strconv"
	"strings"
)

// Task 获取任务列表、详情、日志，取消或删除任务。
func Task(c *context.Context) {
	var form struct {
		ID       string `json:"id" default:"" validate:"omitempty" label:"任务ID"`
		Status   string `json:"status" default:"" validate:"omitempty" label:"任务状态"`
		Action   string `json:"action" default:"" validate:"omitempty,oneof=list info log delete cancel" label:"操作类型"`
		After    int64  `json:"after" default:"0" validate:"numeric,min=0" label:"增量游标"`
		Before   int64  `json:"before" default:"0" validate:"numeric,min=0" label:"历史游标"`
		PageSize int    `json:"page_size" default:"300" validate:"numeric,min=1,max=10000" label:"读取行数"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	action := strings.ToLower(strings.TrimSpace(form.Action))
	switch action {
	case "cancel":
		if form.ID == "" {
			c.Error("任务ID不能为空")
			return
		}
		if !service.System.TaskCancel(form.ID) {
			c.Error("取消任务失败，任务可能不存在或已结束")
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventUpdate, "取消系统任务", "取消系统任务["+form.ID+"]")
		c.Success("任务已取消")
	case "delete":
		if form.ID == "" {
			c.Error("任务ID不能为空")
			return
		}
		if !service.System.TaskDelete(form.ID) {
			c.Error("删除任务失败，任务可能不存在")
			return
		}
		service.Log.Create(c.GetRequestIp(), log_type.EventDelete, "删除系统任务", "删除系统任务["+form.ID+"]")
		c.Success("任务已删除")
	case "log":
		if form.ID == "" || service.System.GetTask(form.ID) == nil {
			c.Error("任务不存在")
			return
		}
		c.SuccessWithData("获取成功", service.System.TaskLog(form.ID, form.After, form.Before, form.PageSize))
	case "info":
		if form.ID == "" {
			c.Error("任务ID不能为空")
			return
		}
		task := service.System.GetTask(form.ID)
		if task == nil {
			c.Error("任务不存在")
			return
		}
		c.SuccessWithData("获取成功", task)
	case "", "list":
		if action == "" && form.ID != "" {
			task := service.System.GetTask(form.ID)
			if task == nil {
				c.Error("任务不存在")
				return
			}
			c.SuccessWithData("获取成功", task)
			return
		}
		tasks := service.System.TaskList()
		if form.Status != "" {
			filteredTasks := make([]interface{}, 0)
			for _, task := range tasks {
				if strconv.Itoa(task.Status) == form.Status {
					filteredTasks = append(filteredTasks, task)
				}
			}
			c.SuccessWithData("获取成功", filteredTasks)
			return
		}
		result := make([]interface{}, len(tasks))
		for i, task := range tasks {
			result[i] = task
		}
		c.SuccessWithData("获取成功", result)
	default:
		c.Error("任务操作不支持")
	}
}
