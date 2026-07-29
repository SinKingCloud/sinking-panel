package system

import (
	"server/app/service"
	"server/app/util/context"
	"strconv"
)

// Task 获取任务列表 获取任务详情 取消任务
func Task(c *context.Context) {
	var form struct {
		ID     string `json:"id" default:"" validate:"omitempty" label:"任务ID"`
		Status string `json:"status" default:"" validate:"omitempty" label:"任务状态"`
		Action string `json:"action" default:"" validate:"omitempty" label:"操作类型"`
	}
	if ok, msg := c.ValidatorAll(&form); !ok {
		c.Error(msg)
		return
	}
	if form.ID != "" && form.Action == "cancel" {
		success := service.System.TaskCancel(form.ID)
		if !success {
			c.Error("取消任务失败，任务可能不存在或已结束")
			return
		}
		c.Success("任务已取消")
		return
	}
	if form.ID != "" {
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
			if strconv.Itoa(int(task.Status)) == form.Status {
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
}
