package recordstore

import "errors"

// Event 保存一条带时间的结构化事件；Data 支持 JSON 对象、数组与标量。
// ID 由 UTC 日期和文件偏移组成，适合访问日志、审计事件等追加记录。
type Event struct {
	ID   string                 `json:"id"`
	Time int64                  `json:"time"`
	Data map[string]interface{} `json:"data"`
}

type EventQuery struct {
	Start   int64                  `json:"start"`
	End     int64                  `json:"end"`
	Limit   int                    `json:"limit"`
	Cursor  string                 `json:"cursor"`
	Filters map[string]interface{} `json:"filters"`
}

// EventPage 按 UTC 日期、日文件追加顺序分页，不按事件时间重新排序。
// 每页除 Limit 外另受 8 MiB 序列化大小限制；使用 NextCursor 持续读取。
type EventPage struct {
	Items      []Event `json:"items"`
	NextCursor string  `json:"next_cursor,omitempty"`
}

type StatsQuery struct {
	Start     int64                  `json:"start"`
	End       int64                  `json:"end"`
	Filters   map[string]interface{} `json:"filters"`
	GroupBy   []string               `json:"group_by"`
	Sum       []string               `json:"sum"`
	MaxGroups int                    `json:"max_groups"`
}

// NumberStats 仅统计存在且为数值的字段；Count 可以小于事件数。
type NumberStats struct {
	Count int64   `json:"count"`
	Sum   float64 `json:"sum"`
	Mean  float64 `json:"mean"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
}

type EventGroup struct {
	Values map[string]interface{} `json:"values"`
	Count  int64                  `json:"count"`
	Fields map[string]NumberStats `json:"fields"`
}

// EventStats 超过分组上限时 Truncated 为 true，总计仍覆盖所有匹配事件。
type EventStats struct {
	Count     int64                  `json:"count"`
	Fields    map[string]NumberStats `json:"fields"`
	Groups    []EventGroup           `json:"groups"`
	Truncated bool                   `json:"truncated"`
}

type EventSettings struct {
	Enabled       bool   `json:"enabled"`
	RetentionDays int    `json:"retention_days"`
	StorageBytes  int64  `json:"storage_bytes"`
	LastError     string `json:"last_error,omitempty"`
}

var ErrEventsDisabled = errors.New("事件记录已关闭")
