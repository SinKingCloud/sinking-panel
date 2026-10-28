package recordstore

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Query 只读取命中日期和分钟的记录，边读边聚合，内存占用不随保留天数增长。
func (s *Series) Query(start, end int64, options ...SeriesQueryOptions) (SeriesResult, error) {
	result := SeriesResult{Points: make([]Point, 0), Start: start, End: end}
	now := time.Now().UnixMilli()
	if start <= 0 || end <= start || end-start > 365*seriesDayMillis || end > now+seriesMinuteMillis {
		return result, errors.New("数值历史查询时间无效，查询跨度不能超过 365 天")
	}
	// 按完整请求跨度确定精度，不能在保留期截断后重新缩小间隔。
	step, err := seriesQueryStep(end-start, options)
	if err != nil {
		return result, err
	}
	result.Interval = step / 1000
	s.mu.RLock()
	defer s.mu.RUnlock()
	cutoff := now - int64(s.config.RetentionDays)*seriesDayMillis
	if start < cutoff {
		start = cutoff
	}
	if end < start {
		return result, nil
	}
	// 分钟记录以窗口起点标记，边界处不返回区间外的数据。
	first := (start + seriesMinuteMillis - 1) / seriesMinuteMillis * seriesMinuteMillis
	last := end / seriesMinuteMillis * seriesMinuteMillis
	if first > last {
		return result, nil
	}
	// 复用单日缓冲区，全年查询也不按日重复分配内存。
	buffer := make([]byte, min(int64(1440), (last-first)/seriesMinuteMillis+1)*int64(s.recordSize))
	appendPoint := func(sample Point) error {
		bucket := sample.Time / step * step
		// 使用实际首个样本的时间，保持返回点处于请求区间内。
		if len(result.Points) > 0 && result.Points[len(result.Points)-1].Time/step*step == bucket {
			return s.merge(&result.Points[len(result.Points)-1], sample)
		} else {
			point := Point{}
			if err := s.merge(&point, sample); err != nil {
				return err
			}
			result.Points = append(result.Points, point)
		}
		return nil
	}
	values := make([]float64, len(s.fields))
	for day := first / seriesDayMillis * seriesDayMillis; day <= last; day += seriesDayMillis {
		file, err := os.Open(s.path(day))
		if errors.Is(err, os.ErrNotExist) {
			if s.pending.Samples > 0 && s.pending.Time/seriesDayMillis*seriesDayMillis == day && s.pending.Time >= first && s.pending.Time <= last {
				if err := appendPoint(s.pending); err != nil {
					return result, err
				}
			}
			continue
		}
		if err != nil {
			return result, err
		}
		err = func() error {
			defer file.Close()
			if err := s.checkHeader(file); err != nil {
				return err
			}
			begin, finish := day, day+seriesDayMillis-seriesMinuteMillis
			if begin < first {
				begin = first
			}
			if finish > last {
				finish = last
			}
			// 一次最多读取一个日文件，避免长区间产生大量小随机读取。
			data := buffer[:(finish-begin)/seriesMinuteMillis*int64(s.recordSize)+int64(s.recordSize)]
			n, err := file.ReadAt(data, s.slotOffset(begin))
			if err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			for offset := 0; offset < len(data); offset += s.recordSize {
				expected := begin + int64(offset/s.recordSize)*seriesMinuteMillis
				if s.pending.Samples > 0 && s.pending.Time == expected {
					if err := appendPoint(s.pending); err != nil {
						return err
					}
					continue
				}
				if offset+s.recordSize > n {
					continue
				}
				sample, ok := s.decode(data[offset:offset+s.recordSize], values)
				if !ok || sample.Time != expected {
					continue
				}
				if err := appendPoint(sample); err != nil {
					return err
				}
			}
			return nil
		}()
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

// seriesQueryStep 显式点数预算使用常见间隔，避免长时间范围产生大量前端绘图点。
// 默认调用保持原有整分钟精度；最大查询跨度为 365 天，最小预算 1 也有对应档位。
func seriesQueryStep(span int64, options []SeriesQueryOptions) (int64, error) {
	if len(options) > 1 {
		return 0, errors.New("数值历史查询只接受一组选项")
	}
	maximum := 0
	if len(options) == 1 {
		maximum = options[0].MaxPoints
	}
	if maximum < 0 || maximum > seriesMaxPoints {
		return 0, fmt.Errorf("数值历史查询点数必须为 1 至 %d，零值使用默认精度", seriesMaxPoints)
	}
	if maximum == 0 {
		step := (span + seriesDefaultPoints*seriesMinuteMillis - 1) / (seriesDefaultPoints * seriesMinuteMillis) * seriesMinuteMillis
		return max(seriesMinuteMillis, step), nil
	}
	minimum := (span + int64(maximum) - 1) / int64(maximum)
	for _, minutes := range [...]int64{
		1, 2, 5, 10, 15, 30, 60, 120, 180, 360, 720, 1440,
		2 * 1440, 7 * 1440, 14 * 1440, 30 * 1440, 60 * 1440,
		90 * 1440, 180 * 1440, 365 * 1440,
	} {
		step := minutes * seriesMinuteMillis
		if step >= minimum {
			return step, nil
		}
	}
	return 0, errors.New("数值历史查询跨度超出时间间隔支持范围")
}

// Cleanup 删除完整过期日文件；跨日的边界记录在 Query 中按时间过滤。
func (s *Series) Cleanup(now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cleanupStatus(now)
}

func (s *Series) cleanupStatus(now time.Time) error {
	err := s.cleanup(now)
	if err != nil {
		s.lastError = fmt.Sprintf("清理过期数值历史数据失败: %v", err)
	} else {
		s.lastError = ""
	}
	return err
}

func (s *Series) cleanup(now time.Time) error {
	cutoff := now.UnixMilli() - int64(s.config.RetentionDays)*seriesDayMillis
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		day, valid := seriesDay(entry.Name())
		if entry.IsDir() || !valid {
			continue
		}
		if day.UnixMilli()+seriesDayMillis > cutoff {
			continue
		}
		if err = os.Remove(filepath.Join(s.dir, entry.Name())); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if s.pending.Samples > 0 && s.pending.Time < cutoff {
		s.pending = Point{}
		s.dirty = false
	}
	return nil
}
