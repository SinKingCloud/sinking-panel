package system

import (
	"errors"
	"maps"
	"math"
	"server/app/util/recordstore"
)

const monitorHistoryMaxPoints = 1440

// fillMonitorHistory 从首条真实采样开始补齐时间桶，避免将采样前的空白显示为零值。
// 不改变真实聚合和历史文件；中间及末尾的空桶仍通过 samples=0 标记。
func fillMonitorHistory(result recordstore.RecordResult) (recordstore.RecordResult, error) {
	if result.Start <= 0 || result.End <= result.Start || result.Interval <= 0 || result.Interval > math.MaxInt64/1000 {
		return result, errors.New("监控历史时间桶无效")
	}
	step := result.Interval * 1000
	first, last := result.Start/step, result.End/step
	count := last - first + 1
	if count < 1 || count > monitorHistoryMaxPoints+1 {
		return result, errors.New("监控历史时间桶数量超出限制")
	}

	if len(result.Points) == 0 {
		return result, nil
	}
	// Query 已按时间排序并聚合，顺序归位即可，无需再建立桶索引。
	previous := int64(-1)
	for index, point := range result.Points {
		timestamp, ok := point["time"].(int64)
		if !ok || timestamp < result.Start || timestamp > result.End {
			return result, errors.New("监控历史采样时间无效")
		}
		bucket := timestamp / step
		if bucket <= previous {
			return result, errors.New("监控历史时间桶重复或顺序无效")
		}
		if index == 0 {
			first = bucket
		}
		previous = bucket
	}

	result.Start = result.Points[0]["time"].(int64)
	points := make([]map[string]interface{}, 0, int(last-first+1))
	index := 0
	for bucket := first; bucket <= last; bucket++ {
		timestamp := max(result.Start, bucket*step)
		if index < len(result.Points) && result.Points[index]["time"].(int64)/step == bucket {
			row := maps.Clone(result.Points[index])
			row["time"] = timestamp
			points = append(points, row)
			index++
			continue
		}
		row := make(map[string]interface{}, len(systemMonitorFields)+2)
		row["time"], row["samples"] = timestamp, uint32(0)
		for _, field := range systemMonitorFields {
			row[field.Name] = float64(0)
		}
		points = append(points, row)
	}
	result.Points = points
	return result, nil
}
