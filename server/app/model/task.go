package model

import (
	"gorm.io/gorm"
	"server/app/util/str"
	"time"
)

// Task 计划任务表
type Task struct {
	Id         int64        `gorm:"column:id;PRIMARY_KEY" json:"id"`       // 任务ID
	EntryID    int          `gorm:"column:entry_id" json:"entry_id"`       // 调度器任务ID
	TypeId     int64        `gorm:"column:type_id" json:"type_id"`         // 任务类型ID
	ExecType   int          `gorm:"column:exec_type" json:"exec_type"`     // 执行类型
	Name       string       `gorm:"column:name" json:"name"`               // 任务名称
	Spec       string       `gorm:"column:spec" json:"spec"`               // 调度表达式
	Script     string       `gorm:"column:script" json:"script"`           // 执行脚本
	Status     int          `gorm:"column:status" json:"status"`           // 任务状态
	RunTime    str.DateTime `gorm:"column:run_time" json:"run_time"`       // 最近运行时间
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"` // 创建时间
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"` // 更新时间
}

// TableName 获取表名
func (*Task) TableName() string {
	return "cloud_tasks"
}

// BeforeCreate 创建前
func (t *Task) BeforeCreate(_ *gorm.DB) error {
	now := str.DateTime(time.Now())
	t.CreateTime = now
	t.UpdateTime = now
	return nil
}

// BeforeUpdate 更新前
func (t *Task) BeforeUpdate(_ *gorm.DB) error {
	t.UpdateTime = str.DateTime(time.Now())
	return nil
}
