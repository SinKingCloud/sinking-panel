package task

import "server/app/util/str"

// SelectTask 任务查询条件
type SelectTask struct {
	Type            string
	Name            string
	Status          string
	RunTimeStart    string
	RunTimeEnd      string
	CreateTimeStart string
	CreateTimeEnd   string
	UpdateTimeStart string
	UpdateTimeEnd   string
}

// UpdateTask 任务更新
type UpdateTask struct {
	Type   interface{}
	Name   interface{}
	Spec   interface{}
	Script interface{}
	Status interface{}
}

// Task 计划任务表
type Task struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`
	Type       int          `gorm:"column:type" json:"type"`
	Name       string       `gorm:"column:name" json:"name"`
	Spec       string       `gorm:"column:spec" json:"spec"`
	Status     string       `gorm:"column:status" json:"status"`
	RunTime    str.DateTime `gorm:"column:run_time" json:"run_time"`
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"`
}
