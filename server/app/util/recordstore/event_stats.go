package recordstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
)

// Stats 流式计算匹配事件的数量、数值统计与分组；不会将整个日文件加载到内存。
func (s *Events) Stats(query StatsQuery) (EventStats, error) {
	return s.StatsContext(context.Background(), query)
}

func (s *Events) StatsContext(ctx context.Context, query StatsQuery) (EventStats, error) {
	result := EventStats{Fields: make(map[string]NumberStats), Groups: make([]EventGroup, 0)}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if query.MaxGroups == 0 {
		query.MaxGroups = 1000
	}
	if query.MaxGroups < 1 || query.MaxGroups > 1000 || len(query.GroupBy) > 4 || len(query.Sum) > 16 {
		return result, errors.New("最多按 4 个字段分组、统计 16 个数值字段，分组上限为 1 至 1000")
	}
	for _, names := range [][]string{query.GroupBy, query.Sum} {
		seen := make(map[string]bool)
		for _, name := range names {
			if name == "" || len(name) > 128 || seen[name] {
				return result, errors.New("统计字段不能重复，名称须为 1 至 128 字节")
			}
			seen[name] = true
		}
	}
	filters, err := eventNormalizeFilters(query.Filters)
	if err != nil {
		return result, err
	}
	snapshots, start, err := s.snapshots(query.Start, query.End)
	if err != nil {
		return result, err
	}
	groups := make(map[string]int)
	for _, snapshot := range snapshots {
		err = s.scanDay(ctx, snapshot, 0, func(event Event, _ int64, _ int) (bool, error) {
			if event.Time < start || event.Time > query.End || !eventMatches(event.Data, filters) {
				return false, nil
			}
			result.Count++
			if err := eventAddStats(result.Fields, event.Data, query.Sum); err != nil {
				return false, err
			}
			if len(query.GroupBy) == 0 {
				return false, nil
			}
			key, values, ok := eventGroupKey(event.Data, query.GroupBy)
			if !ok {
				result.Truncated = true
				return false, nil
			}
			index, exists := groups[key]
			if !exists {
				if len(result.Groups) == query.MaxGroups {
					result.Truncated = true
					return false, nil
				}
				index = len(result.Groups)
				groups[key] = index
				result.Groups = append(result.Groups, EventGroup{Values: values, Fields: make(map[string]NumberStats)})
			}
			group := &result.Groups[index]
			group.Count++
			return false, eventAddStats(group.Fields, event.Data, query.Sum)
		})
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func eventAddStats(target map[string]NumberStats, values map[string]interface{}, names []string) error {
	for _, name := range names {
		stats := target[name]
		number, ok := values[name].(json.Number)
		if !ok {
			target[name] = stats
			continue
		}
		value, err := number.Float64()
		if err != nil || math.IsInf(value, 0) || math.IsNaN(value) {
			return fmt.Errorf("字段 %s 的数值无效", name)
		}
		sum := stats.Sum + value
		if math.IsInf(sum, 0) || math.IsNaN(sum) {
			return fmt.Errorf("字段 %s 的统计总和超出浮点数范围", name)
		}
		if stats.Count == 0 {
			stats.Min, stats.Max = value, value
		} else {
			stats.Min, stats.Max = math.Min(stats.Min, value), math.Max(stats.Max, value)
		}
		stats.Count++
		stats.Sum = sum
		stats.Mean = sum / float64(stats.Count)
		target[name] = stats
	}
	return nil
}

// eventGroupKey 除了分组数量，还约束键长度，避免超长路径占用无界内存。
func eventGroupKey(fields map[string]interface{}, names []string) (string, map[string]interface{}, bool) {
	values := make(map[string]interface{}, len(names))
	keys := make(map[string]interface{}, len(names))
	for _, name := range names {
		value, exists := fields[name]
		if !exists {
			continue
		}
		if text, ok := value.(string); ok && len(text) > 4096 {
			return "", nil, false
		}
		values[name] = value
		key := struct {
			Type  string      `json:"type"`
			Value interface{} `json:"value"`
		}{Value: value}
		switch number := value.(type) {
		case json.Number:
			rational, ok := new(big.Rat).SetString(string(number))
			if !ok {
				return "", nil, false
			}
			key.Type, key.Value = "number", rational.RatString()
		case string:
			key.Type = "string"
		case bool:
			key.Type = "bool"
		case nil:
			key.Type = "null"
		default:
			return "", nil, false
		}
		keys[name] = key
	}
	encoded, err := json.Marshal(keys)
	if err != nil || len(encoded) > 4096 {
		return "", nil, false
	}
	return string(encoded), values, true
}
