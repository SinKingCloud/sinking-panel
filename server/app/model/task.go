package model

import (
	"gorm.io/gorm"
	"server/app/util/str"
	"time"
)

// Task 计划任务表
type Task struct {
	Id         int          `gorm:"column:id;PRIMARY_KEY" json:"id"`
	EntryID    int          `gorm:"column:entry_id" json:"entry_id"`
	Type       int          `gorm:"column:type" json:"type"`
	Name       string       `gorm:"column:name" json:"name"`
	Spec       string       `gorm:"column:spec" json:"spec"`
	Script     string       `gorm:"column:script" json:"script"`
	Status     int          `gorm:"column:status" json:"status"`
	RunTime    str.DateTime `gorm:"column:run_time" json:"run_time"`
	CreateTime str.DateTime `gorm:"column:create_time" json:"create_time"`
	UpdateTime str.DateTime `gorm:"column:update_time" json:"update_time"`
}

// TableName 获取表名
func (*Task) TableName() string {
	return "cloud_tasks"
}

// BeforeCreate 创建前
func (t *Task) BeforeCreate(_ *gorm.DB) error {
	t.CreateTime = str.DateTime(time.Now())
	return nil
}

// BeforeUpdate 更新前
func (t *Task) BeforeUpdate(_ *gorm.DB) error {
	t.UpdateTime = str.DateTime(time.Now())
	return nil
}
