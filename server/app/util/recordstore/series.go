package recordstore

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

const (
	seriesMinuteMillis  = int64(time.Minute / time.Millisecond)
	seriesDayMillis     = int64(24 * time.Hour / time.Millisecond)
	seriesHeaderSize    = 16
	seriesDefaultPoints = 1200
	seriesMaxPoints     = 1440
	seriesMagic         = "SPHIS001"
)

// Series 是并发安全的分钟数值历史，不启动后台协程。同一目录只应由一个实例写入。
type Series struct {
	mu                sync.RWMutex
	dir               string
	fields            []Field
	recordSize        int
	schemaHash        uint32
	config            SeriesConfiguration
	saveConfiguration func(SeriesConfiguration) error
	pending           Point
	dirty             bool
	lastError         string
}

// OpenSeries 为不同消费者打开独立目录，并验证已持久化的字段定义。
func OpenSeries(dir string, fields []Field, options ...SeriesOptions) (*Series, error) {
	fields, err := seriesFields(fields)
	if err != nil {
		return nil, err
	}
	if len(options) > 1 {
		return nil, errors.New("数值历史只接受一组选项")
	}
	var option SeriesOptions
	if len(options) == 1 {
		option = options[0]
	}
	if (option.Configuration == nil) != (option.SaveConfiguration == nil) {
		return nil, errors.New("外部数值历史配置与保存方法必须同时提供")
	}
	if option.Configuration != nil {
		if err = validateDays(option.Configuration.RetentionDays); err != nil {
			return nil, err
		}
	}
	if err = os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("创建数值历史目录失败: %w", err)
	}
	s := &Series{dir: dir, fields: fields, recordSize: 16 + 8*len(fields), schemaHash: seriesSchemaHash(fields),
		config: SeriesConfiguration{Enabled: true, RetentionDays: 30}}
	if err = s.validateFiles(); err != nil {
		return nil, err
	}
	if option.Configuration != nil {
		s.config = *option.Configuration
		s.saveConfiguration = option.SaveConfiguration
	} else {
		s.saveConfiguration = func(config SeriesConfiguration) error {
			return writeJSON(s.dir, "config.json", config)
		}
		data, err := os.ReadFile(filepath.Join(dir, "config.json"))
		if err == nil {
			if err = json.Unmarshal(data, &s.config); err != nil {
				return nil, fmt.Errorf("读取数值历史配置失败: %w", err)
			}
			if err = validateDays(s.config.RetentionDays); err != nil {
				return nil, err
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		} else if err = s.saveConfiguration(s.config); err != nil {
			return nil, err
		}
	}
	// 过期文件删除失败只展示警告，不阻止已配置的写入。
	_ = s.Cleanup(time.Now())
	return s, nil
}

func (s *Series) Settings() (SeriesSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := SeriesSettings{Enabled: s.config.Enabled, RetentionDays: s.config.RetentionDays, Interval: 60, LastError: s.lastError}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return result, err
	}
	for _, entry := range entries {
		_, valid := seriesDay(entry.Name())
		if entry.IsDir() || (!valid && entry.Name() != "config.json") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return result, err
		}
		result.StorageBytes += info.Size()
	}
	return result, nil
}

func (s *Series) UpdateSettings(enabled bool, days int) error {
	if err := validateDays(days); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.flush(); err != nil {
		return err
	}
	next := SeriesConfiguration{Enabled: enabled, RetentionDays: days}
	if err := s.saveConfiguration(next); err != nil {
		return err
	}
	s.config = next
	s.pending = Point{}
	// 已提交的配置保持成功语义，清理错误独立展示并在下次成功时清除。
	_ = s.cleanupStatus(time.Now())
	return nil
}

// Add 合并实际采样，不填补空白分钟。调用方应按时间顺序传入，Last 保留最后传入值。
func (s *Series) Add(point Point) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.config.Enabled {
		return nil
	}
	if point.Time <= 0 || point.Time > time.Now().Add(time.Minute).UnixMilli() {
		return errors.New("数值历史采样时间无效")
	}
	if point.Samples == 0 {
		point.Samples = 1
	}
	if !s.validPoint(point) {
		return errors.New("数值历史字段数量或数值无效")
	}
	point.Time = point.Time / seriesMinuteMillis * seriesMinuteMillis
	if point.Time < time.Now().UnixMilli()-int64(s.config.RetentionDays)*seriesDayMillis ||
		(s.pending.Samples > 0 && point.Time < s.pending.Time) {
		return nil
	}
	if s.pending.Time != point.Time {
		if err := s.flush(); err != nil {
			return err
		}
		previous, err := s.readSlot(point.Time)
		if err != nil {
			return err
		}
		s.pending = previous
	}
	if err := s.merge(&s.pending, point); err != nil {
		return err
	}
	s.dirty = true
	return nil
}

func (s *Series) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flush()
}

func (s *Series) flush() error {
	if !s.dirty || s.pending.Samples == 0 {
		return nil
	}
	file, err := os.OpenFile(s.path(s.pending.Time), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.Size() == 0 {
		var header [seriesHeaderSize]byte
		copy(header[:], seriesMagic)
		binary.LittleEndian.PutUint32(header[8:12], uint32(s.recordSize))
		binary.LittleEndian.PutUint32(header[12:16], s.schemaHash)
		if _, err = file.WriteAt(header[:], 0); err != nil {
			return err
		}
	} else if err = s.checkHeader(file); err != nil {
		return err
	}
	data := s.encode(s.pending)
	if _, err = file.WriteAt(data, s.slotOffset(s.pending.Time)); err != nil {
		return err
	}
	if err = file.Sync(); err != nil {
		return err
	}
	s.dirty = false
	return nil
}

func (s *Series) path(timestamp int64) string {
	return filepath.Join(s.dir, time.UnixMilli(timestamp).UTC().Format("2006-01-02")+".series")
}

func (s *Series) slotOffset(timestamp int64) int64 {
	return seriesHeaderSize + timestamp%seriesDayMillis/seriesMinuteMillis*int64(s.recordSize)
}

func (s *Series) checkHeader(file *os.File) error {
	var header [seriesHeaderSize]byte
	if _, err := file.ReadAt(header[:], 0); err != nil {
		return fmt.Errorf("数值历史文件头损坏: %w", err)
	}
	if string(header[:8]) != seriesMagic || binary.LittleEndian.Uint32(header[12:16]) != s.schemaHash ||
		binary.LittleEndian.Uint32(header[8:12]) != uint32(s.recordSize) {
		return errors.New("数值历史文件版本或字段定义不兼容")
	}
	return nil
}

func (s *Series) readSlot(timestamp int64) (Point, error) {
	file, err := os.Open(s.path(timestamp))
	if errors.Is(err, os.ErrNotExist) {
		return Point{}, nil
	}
	if err != nil {
		return Point{}, err
	}
	defer file.Close()
	if err = s.checkHeader(file); err != nil {
		return Point{}, err
	}
	data := make([]byte, s.recordSize)
	if _, err = file.ReadAt(data, s.slotOffset(timestamp)); errors.Is(err, io.EOF) {
		return Point{}, nil
	} else if err != nil {
		return Point{}, err
	}
	point, ok := s.decode(data, make([]float64, len(s.fields)))
	if !ok || point.Time != timestamp {
		return Point{}, nil
	}
	return point, nil
}

func (s *Series) validPoint(point Point) bool {
	if len(point.Values) != len(s.fields) || point.Samples == 0 {
		return false
	}
	for _, value := range point.Values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}

func (s *Series) encode(point Point) []byte {
	data := make([]byte, s.recordSize)
	binary.LittleEndian.PutUint64(data[:8], uint64(point.Time))
	binary.LittleEndian.PutUint32(data[8:12], point.Samples)
	for index, value := range point.Values {
		binary.LittleEndian.PutUint64(data[12+index*8:], math.Float64bits(value))
	}
	binary.LittleEndian.PutUint32(data[s.recordSize-4:], crc32.ChecksumIEEE(data[:s.recordSize-4]))
	return data
}

// decode 复用调用方提供的数组，查询时只为最终返回的时间桶复制 Values。
func (s *Series) decode(data []byte, values []float64) (Point, bool) {
	if binary.LittleEndian.Uint32(data[s.recordSize-4:]) != crc32.ChecksumIEEE(data[:s.recordSize-4]) {
		return Point{}, false
	}
	point := Point{Time: int64(binary.LittleEndian.Uint64(data[:8])), Samples: binary.LittleEndian.Uint32(data[8:12]), Values: values}
	for index := range values {
		values[index] = math.Float64frombits(binary.LittleEndian.Uint64(data[12+index*8:]))
	}
	return point, point.Time > 0 && s.validPoint(point)
}

func (s *Series) merge(target *Point, point Point) error {
	if target.Samples == 0 {
		*target = Point{Time: point.Time, Samples: point.Samples, Values: slices.Clone(point.Values)}
		return nil
	}
	count := uint64(target.Samples) + uint64(point.Samples)
	if count > math.MaxUint32 {
		return errors.New("数值历史聚合样本数超出范围")
	}
	weight := float64(point.Samples) / float64(count)
	// 先验证整个聚合结果，避免溢出时留下部分已修改的字段。
	for index, field := range s.fields {
		value := seriesAggregate(field.Aggregate, target.Values[index], point.Values[index], weight)
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return errors.New("数值历史聚合结果超出范围")
		}
	}
	for index, field := range s.fields {
		target.Values[index] = seriesAggregate(field.Aggregate, target.Values[index], point.Values[index], weight)
	}
	target.Samples = uint32(count)
	return nil
}

func seriesAggregate(method Aggregate, left, right, weight float64) float64 {
	switch method {
	case Max:
		return math.Max(left, right)
	case Min:
		return math.Min(left, right)
	case Sum:
		return left + right
	case Last:
		return right
	default:
		return left*(1-weight) + right*weight
	}
}
