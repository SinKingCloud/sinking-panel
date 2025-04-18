package system

import (
	"server/app/service"
	"server/app/util/server"
	"strconv"
)

// Task 获取任务列表或任务详情，支持取消任务
func Task(c *server.Context) {
	type Form struct {
		ID     string `json:"id" default:"" validate:"omitempty" label:"任务ID"`
		Status string `json:"status" default:"" validate:"omitempty" label:"任务状态"`
		Action string `json:"action" default:"" validate:"omitempty" label:"操作类型"`
	}
	form := &Form{}
	if ok, msg := c.ValidatorAll(form); !ok {
		c.Error(msg)
		return
	}
	// 如果提供了ID和取消操作，则取消任务
	if form.ID != "" && form.Action == "cancel" {
		success := service.Task.Cancel(form.ID)
		if !success {
			c.Error("取消任务失败，任务可能不存在或已结束")
			return
		}
		c.Success("任务已取消")
		return
	}
	// 如果提供了ID，则获取任务详情
	if form.ID != "" {
		task := service.Task.Get(form.ID)
		if task == nil {
			c.Error("任务不存在")
			return
		}
		c.SuccessWithData("获取成功", task)
		return
	}
	// 否则获取任务列表
	tasks := service.Task.List()
	// 如果指定了状态，则过滤
	if form.Status != "" {
		filteredTasks := make([]interface{}, 0)
		for _, task := range tasks {
			if strconv.Itoa(int(task.Status)) == form.Status {
				filteredTasks = append(filteredTasks, task)
			}
		}
		c.SuccessWithData("获取成功", filteredTasks)
		return
	}
	// 转换为接口切片
	result := make([]interface{}, len(tasks))
	for i, task := range tasks {
		result[i] = task
	}
	c.SuccessWithData("获取成功", result)
}
