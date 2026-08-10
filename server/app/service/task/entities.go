package task

import "server/app/util/str"

// Info 任务详情
type Info struct {
	Id         int64        `json:"id"`
	TypeId     int64        `json:"type_id"`
	ExecType   int          `json:"exec_type"`
	Name       string       `json:"name"`
	Spec       string       `json:"spec"`
	Script     interface{}  `json:"script"`
	Status     int          `json:"status"`
	RunTime    str.DateTime `json:"run_time"`
	CreateTime str.DateTime `json:"create_time"`
	UpdateTime str.DateTime `json:"update_time"`
}

// Request HTTP请求配置
type Request struct {
	Method  string            `json:"method"`
	Url     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}
