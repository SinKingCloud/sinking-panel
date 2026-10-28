package recordstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	eventMaxRecord = 1 << 20
	eventDayMillis = int64(24 * time.Hour / time.Millisecond)
	eventSuffix    = ".events.jsonl"
)

// Events 是按日追加的事件存储。每个目录由一个 Events 实例独占写入。
// Append 将一条完整 JSON 写入操作系统缓存；调用方按需定时 Sync，Close 也会同步。
// 不包含后台协程；与系统/网站的采集、排队、调度解耦。
type Events struct {
	mu        sync.RWMutex
	dir       string
	config    eventConfig
	file      *os.File
	day       string
	offset    int64
	closed    bool
	lastError string
}

type eventConfig struct {
	Enabled       bool `json:"enabled"`
	RetentionDays int  `json:"retention_days"`
}

func OpenEvents(dir string) (*Events, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	s := &Events{dir: dir, config: eventConfig{Enabled: true, RetentionDays: 30}}
	data, err := os.ReadFile(filepath.Join(dir, "events.config.json"))
	if err == nil {
		if err = json.Unmarshal(data, &s.config); err != nil {
			return nil, fmt.Errorf("读取事件配置失败: %w", err)
		}
		if err = validateDays(s.config.RetentionDays); err != nil {
			return nil, err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if err = writeJSON(s.dir, "events.config.json", s.config); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	_ = s.Cleanup(time.Now())
	return s, nil
}

func (s *Events) Settings() (EventSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	settings := EventSettings{Enabled: s.config.Enabled, RetentionDays: s.config.RetentionDays, LastError: s.lastError}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return settings, err
	}
	for _, entry := range entries {
		if entry.IsDir() || (!eventIsDayFile(entry.Name()) && entry.Name() != "events.config.json") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return settings, err
		}
		settings.StorageBytes += info.Size()
	}
	return settings, nil
}

func (s *Events) UpdateSettings(enabled bool, days int) error {
	if err := validateDays(days); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("事件存储已关闭")
	}
	if s.file != nil {
		if err := s.file.Sync(); err != nil {
			return err
		}
	}
	next := eventConfig{Enabled: enabled, RetentionDays: days}
	if err := writeJSON(s.dir, "events.config.json", next); err != nil {
		return err
	}
	s.config = next
	// 设置已经持久化；清理失败作为健康状态返回，避免“失败但已生效”。
	_ = s.cleanup(time.Now())
	return nil
}

// Append 支持保留范围内的乱序事件，按事件时间选择 UTC 日文件。
func (s *Events) Append(timestamp int64, fields map[string]interface{}) (Event, error) {
	data, err := eventNormalizeData(fields)
	if err != nil {
		return Event{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return Event{}, errors.New("事件存储已关闭")
	}
	if !s.config.Enabled {
		return Event{}, ErrEventsDisabled
	}
	now := time.Now().UnixMilli()
	if timestamp <= 0 || timestamp > now+60000 || timestamp < now-int64(s.config.RetentionDays)*eventDayMillis {
		return Event{}, errors.New("事件时间超出可记录范围")
	}
	day := time.UnixMilli(timestamp).UTC().Format("2006-01-02")
	if s.file == nil || s.day != day {
		if err := s.openDay(day); err != nil {
			return Event{}, err
		}
	}
	event := Event{ID: day + ":" + strconv.FormatInt(s.offset, 10), Time: timestamp, Data: data}
	encoded, err := json.Marshal(event)
	if err != nil {
		return Event{}, err
	}
	if len(encoded)+1 > eventMaxRecord {
		return Event{}, errors.New("单条事件不能超过 1 MiB")
	}
	encoded = append(encoded, '\n')
	n, err := s.file.Write(encoded)
	if err == nil && n != len(encoded) {
		err = io.ErrShortWrite
	}
	if err != nil {
		// 失败写入不能成为下一条记录的前缀，保持完整行边界。
		truncateErr := s.file.Truncate(s.offset)
		_, seekErr := s.file.Seek(s.offset, io.SeekStart)
		return Event{}, errors.Join(err, truncateErr, seekErr)
	}
	s.offset += int64(n)
	return event, nil
}

func (s *Events) openDay(day string) error {
	if err := s.closeFile(); err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(s.dir, day+eventSuffix), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	offset, err := eventRecoverTail(file)
	if err != nil {
		_ = file.Close()
		return err
	}
	if _, err = file.Seek(offset, io.SeekStart); err != nil {
		_ = file.Close()
		return err
	}
	s.file, s.day, s.offset = file, day, offset
	return nil
}

// eventRecoverTail 只删除崩溃留下的不完整尾行，不隐藏完整行的格式错误。
func eventRecoverTail(file *os.File) (int64, error) {
	info, err := file.Stat()
	if err != nil {
		return 0, err
	}
	size := info.Size()
	if size == 0 {
		return 0, nil
	}
	buffer := make([]byte, min(size, int64(eventMaxRecord)))
	if _, err = file.ReadAt(buffer, size-int64(len(buffer))); err != nil {
		return 0, err
	}
	if buffer[len(buffer)-1] == '\n' {
		return size, nil
	}
	last := bytes.LastIndexByte(buffer, '\n')
	if last < 0 && size > int64(len(buffer)) {
		return 0, errors.New("事件文件尾部损坏超过单条记录上限")
	}
	offset := size - int64(len(buffer)) + int64(last+1)
	if err = file.Truncate(offset); err != nil {
		return 0, err
	}
	return offset, file.Sync()
}

func (s *Events) Sync() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("事件存储已关闭")
	}
	if s.file == nil {
		return nil
	}
	return s.file.Sync()
}

func (s *Events) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.closeFile()
}

func (s *Events) closeFile() error {
	if s.file == nil {
		return nil
	}
	err := errors.Join(s.file.Sync(), s.file.Close())
	s.file, s.day, s.offset = nil, "", 0
	return err
}

func (s *Events) Cleanup(now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("事件存储已关闭")
	}
	return s.cleanup(now)
}

func (s *Events) cleanup(now time.Time) (result error) {
	defer func() {
		s.lastError = ""
		if result != nil {
			s.lastError = result.Error()
		}
	}()
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}
	cutoff := now.UnixMilli() - int64(s.config.RetentionDays)*eventDayMillis
	for _, entry := range entries {
		if entry.IsDir() || !eventIsDayFile(entry.Name()) {
			continue
		}
		day := strings.TrimSuffix(entry.Name(), eventSuffix)
		date, _ := time.Parse("2006-01-02", day)
		if date.UnixMilli()+eventDayMillis > cutoff {
			continue
		}
		if s.day == day {
			if err := s.closeFile(); err != nil {
				result = errors.Join(result, err)
				continue
			}
		}
		if err := os.Remove(filepath.Join(s.dir, entry.Name())); err != nil && !errors.Is(err, os.ErrNotExist) {
			result = errors.Join(result, err)
		}
	}
	return result
}

func eventIsDayFile(name string) bool {
	if len(name) != 10+len(eventSuffix) || !strings.HasSuffix(name, eventSuffix) {
		return false
	}
	day := strings.TrimSuffix(name, eventSuffix)
	date, err := time.Parse("2006-01-02", day)
	return err == nil && date.Format("2006-01-02") == day
}

func eventNormalizeFilters(fields map[string]interface{}) (map[string]interface{}, error) {
	if len(fields) > 128 {
		return nil, errors.New("事件字段不能超过 128 个")
	}
	for name, value := range fields {
		if name == "" || len(name) > 128 {
			return nil, errors.New("事件字段名称须为 1 至 128 字节")
		}
		switch value.(type) {
		case nil, string, bool, json.Number, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		default:
			return nil, fmt.Errorf("事件字段 %s 只支持字符串、布尔、数值或 nil", name)
		}
	}
	data, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	if len(data) > eventMaxRecord {
		return nil, errors.New("单条事件不能超过 1 MiB")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	normalized := make(map[string]interface{})
	if err = decoder.Decode(&normalized); err != nil {
		return nil, err
	}
	if normalized == nil {
		normalized = make(map[string]interface{})
	}
	for _, value := range normalized {
		if number, ok := value.(json.Number); ok {
			if err := eventValidateNumber(number); err != nil {
				return nil, err
			}
		}
	}
	return normalized, nil
}

// eventNormalizeData 深拷贝任意 JSON 字段，嵌套结构可记录 headers、标签等上下文。
func eventNormalizeData(fields map[string]interface{}) (map[string]interface{}, error) {
	if len(fields) > 128 {
		return nil, errors.New("事件顶层字段不能超过 128 个")
	}
	for name := range fields {
		if name == "" || len(name) > 128 {
			return nil, errors.New("事件字段名称须为 1 至 128 字节")
		}
	}
	data, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	if len(data) > eventMaxRecord {
		return nil, errors.New("单条事件不能超过 1 MiB")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	normalized := make(map[string]interface{})
	if err = decoder.Decode(&normalized); err != nil {
		return nil, err
	}
	if normalized == nil {
		normalized = make(map[string]interface{})
	}
	nodes := 0
	if err = eventValidateJSON(normalized, 0, &nodes); err != nil {
		return nil, err
	}
	return normalized, nil
}

func eventValidateJSON(value interface{}, depth int, nodes *int) error {
	*nodes = *nodes + 1
	if depth > 32 || *nodes > 16384 {
		return errors.New("事件 JSON 最多嵌套 32 层、包含 16384 个节点")
	}
	switch item := value.(type) {
	case map[string]interface{}:
		for _, child := range item {
			if err := eventValidateJSON(child, depth+1, nodes); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, child := range item {
			if err := eventValidateJSON(child, depth+1, nodes); err != nil {
				return err
			}
		}
	case json.Number:
		return eventValidateNumber(item)
	}
	return nil
}

// eventValidateNumber 同时约束指数，避免精确过滤时展开恶意超大有理数。
func eventValidateNumber(value json.Number) error {
	number, err := value.Float64()
	if err != nil || len(value) > 64 || math.IsNaN(number) || math.IsInf(number, 0) {
		return errors.New("事件数值必须是有限实数")
	}
	if index := strings.LastIndexAny(string(value), "eE"); index >= 0 {
		exponent, err := strconv.ParseInt(string(value[index+1:]), 10, 32)
		if err != nil || exponent < -1024 || exponent > 1024 {
			return errors.New("事件数值指数超出支持范围")
		}
	}
	return nil
}
