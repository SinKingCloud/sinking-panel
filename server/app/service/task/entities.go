package task

import "server/app/util/str"

// Info 任务详情
type Info struct {
	Id         int64        `json:"id"`
	Type       int          `json:"type"`
	Name       string       `json:"name"`
	Spec       string       `json:"spec"`
	Script     string       `json:"script"`
	Status     int          `json:"status"`
	RunTime    str.DateTime `json:"run_time"`
	CreateTime str.DateTime `json:"create_time"`
	UpdateTime str.DateTime `json:"update_time"`
}
