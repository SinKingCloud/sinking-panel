package system

import "context"

// Task 任务信息
type Task struct {
	ID         string      `json:"id"`          // 任务ID
	Name       string      `json:"name"`        // 任务名称
	Status     int         `json:"status"`      // 任务状态
	Progress   float64     `json:"progress"`    // 进度（0-100）
	Message    string      `json:"message"`     // 消息
	Data       interface{} `json:"data"`        // 数据
	StartTime  int64       `json:"start_time"`  // 开始时间
	EndTime    int64       `json:"end_time"`    // 结束时间
	CreateTime int64       `json:"create_time"` // 创建时间
	UpdateTime int64       `json:"update_time"` // 更新时间
	cancel     context.CancelFunc
}

type job struct {
	id   string
	ctx  context.Context
	data interface{}
	run  func(context.Context, interface{}, func(int, float64, string))
}
