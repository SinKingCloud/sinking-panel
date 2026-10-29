package system

import "time"

const maxTaskWorkers = 5 // 系统任务并发 worker 数量

const (
	monitorHistoryDefaultPoints = 480
	monitorHistoryMinPoints     = 120
	monitorHistoryMaxPoints     = 720
)

const (
	processMinuteMillis = int64(time.Minute / time.Millisecond)
	processHourMillis   = int64(time.Hour / time.Millisecond)
	processDayMillis    = 24 * processHourMillis
)
