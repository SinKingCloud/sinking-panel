package system

import "time"

const (
	monitorHistoryDefaultPoints = 480
	monitorHistoryMinPoints     = 120
	monitorHistoryMaxPoints     = 720
)

const (
	processSampleInterval = 5 * time.Second
	processViewerTimeout  = 15 * time.Second
	processMinuteMillis   = int64(time.Minute / time.Millisecond)
	processHourMillis     = int64(time.Hour / time.Millisecond)
	processDayMillis      = 24 * processHourMillis
)
