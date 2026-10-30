package recordstore

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// RecordResult 将时间序列转换为可直接返回接口的字段记录。
type RecordResult struct {
	Points   []map[string]interface{} `json:"points"`
	Interval int64                    `json:"interval"`
	Start    int64                    `json:"start"`
	End      int64                    `json:"end"`
}

type queuedRecord struct {
	point      Point
	generation uint64
}

// Recorder 为数值序列提供异步写入、即时启停和自动清理；不负责业务采集。
type Recorder struct {
	series      *Series
	fields      []Field
	queue       chan queuedRecord
	done        chan struct{}
	inputMu     sync.Mutex
	writeMu     sync.Mutex
	healthMu    sync.Mutex
	enabled     atomic.Bool
	closed      atomic.Bool
	generation  atomic.Uint64
	dropped     atomic.Uint64
	sampleError error
	writeError  error
	closeError  error
}

// OpenRecorder 打开一个独立目录的记录器，调用方应在退出时 Close。
func OpenRecorder(dir string, fields []Field, options ...SeriesOptions) (*Recorder, error) {
	for _, field := range fields {
		if field.Name == "time" || field.Name == "samples" {
			return nil, fmt.Errorf("记录字段 %s 是保留名称", field.Name)
		}
	}
	series, err := OpenSeries(dir, fields, options...)
	if err != nil {
		return nil, err
	}
	settings, err := series.Settings()
	if err != nil {
		return nil, err
	}
	r := &Recorder{series: series, fields: append([]Field(nil), fields...), queue: make(chan queuedRecord, 120), done: make(chan struct{})}
	r.enabled.Store(settings.Enabled)
	go r.run()
	return r, nil
}

// Enabled 是否正在接受新记录。
func (r *Recorder) Enabled() bool { return r.enabled.Load() && !r.closed.Load() }

// Record 复制数值后加入有界队列，不在采集线程中执行文件 I/O。
// 每次提交完整的字段集合；停用时忽略记录，队列满时返回错误并计入健康状态。
func (r *Recorder) Record(timestamp int64, values map[string]float64) error {
	r.inputMu.Lock()
	defer r.inputMu.Unlock()
	if r.closed.Load() {
		return errors.New("历史记录器已关闭")
	}
	if !r.enabled.Load() {
		return nil
	}
	if timestamp <= 0 || timestamp > time.Now().Add(time.Minute).UnixMilli() {
		err := errors.New("历史记录时间无效")
		r.Report(err)
		return err
	}
	if len(values) != len(r.fields) {
		err := errors.New("记录字段与已定义的时间序列不一致")
		r.Report(err)
		return err
	}
	point := Point{Time: timestamp, Samples: 1, Values: make([]float64, len(r.fields))}
	for index, field := range r.fields {
		value, ok := values[field.Name]
		if !ok {
			err := fmt.Errorf("记录缺少字段 %s", field.Name)
			r.Report(err)
			return err
		}
		if math.IsNaN(value) || math.IsInf(value, 0) {
			err := fmt.Errorf("记录字段 %s 必须是有限数值", field.Name)
			r.Report(err)
			return err
		}
		point.Values[index] = value
	}
	select {
	case r.queue <- queuedRecord{point: point, generation: r.generation.Load()}:
		r.Report(nil)
		return nil
	default:
		r.dropped.Add(1)
		return errors.New("历史记录队列已满，本次采样已跳过")
	}
}

// Report 将采集错误暴露到设置接口，相同故障只输出一次日志。
func (r *Recorder) Report(err error) {
	r.setError(&r.sampleError, err)
}

func (r *Recorder) setError(target *error, err error) {
	r.healthMu.Lock()
	defer r.healthMu.Unlock()
	*target = err
}

func (r *Recorder) run() {
	defer close(r.done)
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	flushTicker := time.NewTicker(time.Minute)
	defer flushTicker.Stop()
	for {
		select {
		case record, ok := <-r.queue:
			r.writeMu.Lock()
			if !ok {
				r.closeError = r.series.Flush()
				r.setError(&r.writeError, r.closeError)
				r.writeMu.Unlock()
				return
			}
			if record.generation == r.generation.Load() && r.enabled.Load() {
				r.setError(&r.writeError, r.series.Add(record.point))
			}
			r.writeMu.Unlock()
		case now := <-ticker.C:
			// 清理故障由 Series.Settings 保留，正常写入不会抹掉该警告。
			_ = r.series.Cleanup(now)
		case <-flushTicker.C:
			// 稀疏业务即使没有下一条采样，也应定时保存内存中的记录。
			r.writeMu.Lock()
			r.setError(&r.writeError, r.series.Flush())
			r.writeMu.Unlock()
		}
	}
}

func (r *Recorder) Settings() (SeriesSettings, error) {
	settings, err := r.series.Settings()
	if err != nil {
		return settings, err
	}
	r.healthMu.Lock()
	for _, failure := range []error{r.sampleError, r.writeError} {
		if failure != nil {
			if settings.LastError != "" {
				settings.LastError += "；"
			}
			settings.LastError += failure.Error()
		}
	}
	r.healthMu.Unlock()
	if dropped := r.dropped.Load(); dropped > 0 {
		if settings.LastError != "" {
			settings.LastError += "；"
		}
		settings.LastError += fmt.Sprintf("本次运行累计跳过 %d 次队列溢出采样", dropped)
	}
	return settings, nil
}

// UpdateSettings 保存成功后立即切换代次，旧队列不会进入新的采集周期。
func (r *Recorder) UpdateSettings(enabled bool, days int) error {
	r.writeMu.Lock()
	defer r.writeMu.Unlock()
	if r.closed.Load() {
		return errors.New("历史记录器已关闭")
	}
	if err := r.series.UpdateSettings(enabled, days); err != nil {
		return err
	}
	r.inputMu.Lock()
	if r.enabled.Load() != enabled {
		r.generation.Add(1)
	}
	r.enabled.Store(enabled)
	r.inputMu.Unlock()
	return nil
}

func (r *Recorder) Query(start, end int64, options ...SeriesQueryOptions) (RecordResult, error) {
	result, err := r.series.Query(start, end, options...)
	rows := RecordResult{Points: make([]map[string]interface{}, 0, len(result.Points)), Start: result.Start, End: result.End, Interval: result.Interval}
	if err != nil {
		return rows, err
	}
	for _, point := range result.Points {
		row := make(map[string]interface{}, len(r.fields)+2)
		row["time"], row["samples"] = point.Time, point.Samples
		for index, field := range r.fields {
			row[field.Name] = point.Values[index]
		}
		rows.Points = append(rows.Points, row)
	}
	return rows, nil
}

// Close 停止接收记录，排空队列并保存不足一分钟的数据；可重复调用。
func (r *Recorder) Close() error {
	r.inputMu.Lock()
	if !r.closed.Swap(true) {
		close(r.queue)
	}
	r.inputMu.Unlock()
	<-r.done
	return r.closeError
}
