package datastore

import (
	"bytes"
	"context"
	"crypto/sha256"
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
	"strings"
	"sync"
	"time"
)

const (
	timelineMagic       = "SPTLN001"
	timelineSharedMagic = "SPTLN002"
	timelineHeaderSize  = 32
	timelineIndexSize   = 32
	timelineDayMillis   = int64(24 * time.Hour / time.Millisecond)
	timelineMaxTime     = int64(253402300799999) // 9999-12-31，日文件名固定四位年份。
)

// TimelineOptions 使用单流 Interval 或多流 Streams，二者不能同时指定。
type TimelineOptions struct {
	Interval      time.Duration
	MaxValueBytes int // 默认 64 KiB，上限 16 MiB。
	Streams       []TimelineStreamOptions
	ResetInvalid  bool // 初始化时重建格式不兼容或文件头损坏的日文件；默认保留并报错。
}

// TimelineStreamOptions 为同一日文件中的独立时间流定义名称、间隔和数据格式摘要。
// 每个间隔必须整除一天且每天最多 1440 槽；Schema 由调用方定义，修改后拒绝读取旧格式。
type TimelineStreamOptions struct {
	Name          string
	Interval      time.Duration
	MaxValueBytes int
	Schema        uint32
}

// TimelineEntry 保存调用方编码后的任意字节，Time 为 Unix 毫秒。
type TimelineEntry struct {
	Time int64
	Data []byte
}

// Timeline 按天保存变长记录，通过文件内时间索引读取；不创建数据库或后台任务。
// 同一目录只允许一个根实例写入；各 Stream 共享锁，Scan 回调不可重入同一根实例的任何流。
// Stream 视图不拥有文件生命周期，只有打开时返回的根实例负责 Close。
type Timeline struct {
	*timelineFile
	interval int64
	slots    int
	offset   int
	maximum  int
	schema   uint32
	owner    bool
	streams  map[string]*Timeline
}

type timelineFile struct {
	mu        sync.RWMutex
	dir       string
	magic     string
	signature uint64
	total     int
	closed    bool
}

type timelineIndex struct {
	generation uint64
	offset     int64
	length     uint32
	checksum   uint32
	copy       int
}

func OpenTimeline(dir string, options TimelineOptions) (*Timeline, error) {
	if dir == "" || len(options.Streams) > 16 || len(options.Streams) > 0 && (options.Interval != 0 || options.MaxValueBytes != 0) {
		return nil, errors.New("记录目录或流配置无效，最多 16 个流且不能同时指定单流选项")
	}
	definitions := slices.Clone(options.Streams)
	shared := len(definitions) > 0
	if !shared {
		definitions = []TimelineStreamOptions{{Interval: options.Interval, MaxValueBytes: options.MaxValueBytes}}
	} else {
		slices.SortFunc(definitions, func(a, b TimelineStreamOptions) int { return strings.Compare(a.Name, b.Name) })
	}
	file := &timelineFile{dir: dir, magic: timelineMagic}
	root := &Timeline{timelineFile: file, owner: true}
	if shared {
		file.magic = timelineSharedMagic
		root.streams = make(map[string]*Timeline, len(definitions))
	}
	for index := range definitions {
		definition := &definitions[index]
		interval := definition.Interval.Milliseconds()
		if interval <= 0 || definition.Interval%time.Millisecond != 0 ||
			timelineDayMillis%interval != 0 || timelineDayMillis/interval > 1440 {
			return nil, errors.New("记录时间间隔必须整除一天且每天最多 1440 槽")
		}
		if definition.MaxValueBytes == 0 {
			definition.MaxValueBytes = 64 << 10
		}
		if definition.MaxValueBytes < 1 || definition.MaxValueBytes > 16<<20 {
			return nil, errors.New("时间记录大小限制必须为 1 字节至 16 MiB")
		}
		view := &Timeline{timelineFile: file, interval: interval, slots: int(timelineDayMillis / interval),
			offset: file.total, maximum: definition.MaxValueBytes, schema: definition.Schema}
		file.total += view.slots
		if shared {
			if definition.Name == "" || len(definition.Name) > 128 || strings.TrimSpace(definition.Name) != definition.Name || root.streams[definition.Name] != nil {
				return nil, errors.New("时间记录流名称必须非空、唯一且不能带首尾空格")
			}
			root.streams[definition.Name] = view
		} else {
			root.interval, root.slots, root.maximum = interval, view.slots, view.maximum
			file.signature = uint64(interval)
		}
	}
	if shared {
		encoded, _ := json.Marshal(definitions)
		digest := sha256.Sum256(encoded)
		file.signature = binary.LittleEndian.Uint64(digest[:8])
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	if shared || options.ResetInvalid {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			day, valid := timelineDay(entry.Name())
			if !valid {
				continue
			}
			opened, _, err := root.open(day, false, options.ResetInvalid)
			if err != nil {
				return nil, fmt.Errorf("初始化时间记录 %s 失败: %w", entry.Name(), err)
			}
			if err := opened.Close(); err != nil {
				return nil, err
			}
		}
	}
	return root, nil
}

// Stream 获取共享日文件内的命名流；返回的视图不单独持有文件句柄或后台任务。
func (s *Timeline) Stream(name string) (*Timeline, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return nil, os.ErrClosed
	}
	stream := s.streams[name]
	if stream == nil {
		return nil, fmt.Errorf("时间记录流不存在: %s", name)
	}
	return stream, nil
}

func (s *Timeline) Get(timestamp int64) ([]byte, error) {
	if s.interval == 0 || timestamp < 0 || timestamp > timelineMaxTime {
		return nil, errors.New("记录时间无效")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return nil, os.ErrClosed
	}
	day := timestamp / timelineDayMillis * timelineDayMillis
	file, size, err := s.open(day, false, false)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	slot := s.offset + int((timestamp-day)/s.interval)
	var pair [2 * timelineIndexSize]byte
	if _, err = file.ReadAt(pair[:], timelineHeaderSize+int64(slot)*int64(len(pair))); err != nil {
		return nil, err
	}
	index, err := s.index(pair[:], slot, size)
	if err != nil {
		return nil, err
	}
	return s.read(file, index)
}

// PutMany 同一时间槽以最后一项覆盖。先同步数据，再发布双份索引中的下一代。
// 每槽独立提交，不保证跨槽、跨日事务；失败后可重试，相同内容不会重复追加。
func (s *Timeline) PutMany(entries []TimelineEntry) error {
	if s.interval == 0 {
		return errors.New("请指定要写入的时间记录流")
	}
	groups := make(map[int64]map[int]TimelineEntry)
	for _, entry := range entries {
		if entry.Time < 0 || entry.Time > timelineMaxTime || len(entry.Data) > s.maximum {
			return errors.New("记录时间或大小无效")
		}
		day := entry.Time / timelineDayMillis * timelineDayMillis
		if groups[day] == nil {
			groups[day] = make(map[int]TimelineEntry)
		}
		groups[day][s.offset+int((entry.Time-day)/s.interval)] = entry
	}
	days := make([]int64, 0, len(groups))
	for day := range groups {
		days = append(days, day)
	}
	slices.Sort(days)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return os.ErrClosed
	}
	for _, day := range days {
		if err := s.writeDay(day, groups[day]); err != nil {
			return err
		}
	}
	return nil
}

func (s *Timeline) writeDay(day int64, entries map[int]TimelineEntry) error {
	file, size, err := s.open(day, true, false)
	if err != nil {
		return err
	}
	defer file.Close()
	type update struct {
		offset int64
		index  [timelineIndexSize]byte
	}
	updates := make([]update, 0, len(entries))
	var pair [2 * timelineIndexSize]byte
	for slot, entry := range entries {
		position := int64(timelineHeaderSize + slot*len(pair))
		if _, err = file.ReadAt(pair[:], position); err != nil {
			return err
		}
		previous, err := s.index(pair[:], slot, size)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return err
		}
		if err == nil {
			data, err := s.read(file, previous)
			if err != nil {
				return err
			}
			if bytes.Equal(data, entry.Data) {
				continue
			}
		}
		if previous.generation == math.MaxUint64 || size > math.MaxInt64-int64(len(entry.Data)) {
			return errors.New("时间记录索引超出范围")
		}
		if _, err = file.WriteAt(entry.Data, size); err != nil {
			return err
		}
		copyIndex := 0
		if previous.generation > 0 {
			copyIndex = 1 - previous.copy
		}
		next := update{offset: position + int64(copyIndex*timelineIndexSize)}
		binary.LittleEndian.PutUint64(next.index[:8], previous.generation+1)
		binary.LittleEndian.PutUint64(next.index[8:16], uint64(size))
		binary.LittleEndian.PutUint32(next.index[16:20], uint32(len(entry.Data)))
		binary.LittleEndian.PutUint32(next.index[20:24], crc32.ChecksumIEEE(entry.Data))
		binary.LittleEndian.PutUint32(next.index[24:28], uint32(slot))
		binary.LittleEndian.PutUint32(next.index[28:], crc32.ChecksumIEEE(next.index[:28]))
		updates = append(updates, next)
		size += int64(len(entry.Data))
	}
	if len(updates) == 0 {
		return file.Sync()
	}
	if err = file.Sync(); err != nil {
		return err
	}
	for _, update := range updates {
		if _, err = file.WriteAt(update.index[:], update.offset); err != nil {
			return err
		}
	}
	return file.Sync()
}

// Scan 包含两端所在的时间槽，回调获得独立的数据副本。
func (s *Timeline) Scan(ctx context.Context, start, end int64, visit func(TimelineEntry) error) error {
	return s.scan(ctx, start, end, false, visit)
}

// ScanView 复用读取缓冲；Data 只在当前回调内有效，不可修改或保留，需保留时自行复制。
// 与 Scan 一样，回调不可重入同一根实例下的任何流。
func (s *Timeline) ScanView(ctx context.Context, start, end int64, visit func(TimelineEntry) error) error {
	return s.scan(ctx, start, end, true, visit)
}

func (s *Timeline) scan(ctx context.Context, start, end int64, borrowed bool, visit func(TimelineEntry) error) error {
	if s.interval == 0 || start < 0 || end < start || end > timelineMaxTime || visit == nil {
		return errors.New("时间记录查询范围或回调无效")
	}
	start, end = start/s.interval*s.interval, end/s.interval*s.interval
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return os.ErrClosed
	}
	buffer := make([]byte, 2*timelineIndexSize*min(int64(s.slots), (end-start)/s.interval+1))
	var payloadBuffer []byte
	for day := start / timelineDayMillis * timelineDayMillis; day <= end; day += timelineDayMillis {
		if err := ctx.Err(); err != nil {
			return err
		}
		file, size, err := s.open(day, false, false)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		err = func() error {
			defer file.Close()
			first, last := int((max(start, day)-day)/s.interval), int((min(end, day+timelineDayMillis-1)-day)/s.interval)
			data := buffer[:(last-first+1)*2*timelineIndexSize]
			if _, err := file.ReadAt(data, int64(timelineHeaderSize+(s.offset+first)*2*timelineIndexSize)); err != nil {
				return err
			}
			var payloadStart int64 = -1
			for slot := first; slot <= last; slot++ {
				if err := ctx.Err(); err != nil {
					return err
				}
				offset := (slot - first) * 2 * timelineIndexSize
				index, err := s.index(data[offset:offset+2*timelineIndexSize], s.offset+slot, size)
				if errors.Is(err, ErrNotFound) {
					continue
				}
				if err != nil {
					return err
				}
				var payload []byte
				if !borrowed && index.length > 64<<10 {
					// 独立副本模式的大记录直接读入结果，避免额外保留一份大缓冲。
					payload, err = s.read(file, index)
					if err != nil {
						return err
					}
				} else {
					// 小记录有界预读，大记录也复用同一个缓冲，容量不超过 max(64KiB, MaxValueBytes)。
					if payloadStart < 0 || index.offset < payloadStart || index.offset+int64(index.length) > payloadStart+int64(len(payloadBuffer)) {
						length := int(min(max(int64(64<<10), int64(index.length)), size-index.offset))
						if cap(payloadBuffer) < length {
							payloadBuffer = make([]byte, length)
						} else {
							payloadBuffer = payloadBuffer[:length]
						}
						if _, err := file.ReadAt(payloadBuffer, index.offset); err != nil {
							return err
						}
						payloadStart = index.offset
					}
					begin, end := index.offset-payloadStart, index.offset-payloadStart+int64(index.length)
					payload = payloadBuffer[begin:end:end]
					if crc32.ChecksumIEEE(payload) != index.checksum {
						return errors.New("时间记录内容校验失败")
					}
					if !borrowed {
						payload = bytes.Clone(payload)
					}
				}
				if err = visit(TimelineEntry{Time: day + int64(slot)*s.interval, Data: payload}); err != nil {
					return err
				}
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return ctx.Err()
}

func (s *Timeline) open(day int64, create, resetInvalid bool) (*os.File, int64, error) {
	path := filepath.Join(s.dir, time.UnixMilli(day).UTC().Format("2006-01-02")+".data")
	initialize := func(replace bool) error {
		file, err := os.CreateTemp(s.dir, ".timeline-*.tmp")
		if err != nil {
			return err
		}
		defer os.Remove(file.Name())
		var header [timelineHeaderSize]byte
		copy(header[:8], s.magic)
		binary.LittleEndian.PutUint64(header[8:16], s.signature)
		binary.LittleEndian.PutUint64(header[16:24], uint64(day))
		binary.LittleEndian.PutUint32(header[24:28], uint32(s.total))
		binary.LittleEndian.PutUint32(header[28:], crc32.ChecksumIEEE(header[:28]))
		if _, err = file.WriteAt(header[:], 0); err == nil {
			err = file.Truncate(int64(timelineHeaderSize + s.total*2*timelineIndexSize))
		}
		if err == nil {
			err = file.Sync()
		}
		closeErr := file.Close()
		if err != nil || closeErr != nil {
			return errors.Join(err, closeErr)
		}
		if replace {
			// 新文件完整落盘后才替换旧文件，创建失败不会先丢弃原文件。
			return os.Rename(file.Name(), path)
		}
		// 同目录硬链接原子发布完整文件且不覆盖已有目标，关闭句柄后也兼容 Windows。
		return os.Link(file.Name(), path)
	}
	info, err := os.Lstat(path)
	if create && errors.Is(err, os.ErrNotExist) {
		err = initialize(false)
		if err != nil {
			return nil, 0, err
		}
		info, err = os.Lstat(path)
	}
	if err != nil {
		return nil, 0, err
	}
	if !info.Mode().IsRegular() {
		return nil, 0, errors.New("时间记录日文件不是普通文件")
	}
	flags := os.O_RDONLY
	if create {
		flags = os.O_RDWR
	}
	file, err := os.OpenFile(path, flags, 0600)
	if err != nil {
		return nil, 0, err
	}
	stat, err := file.Stat()
	if err == nil && (!stat.Mode().IsRegular() || (info != nil && !os.SameFile(info, stat))) {
		err = errors.New("时间记录日文件已替换或不是普通文件")
	}
	if err == nil {
		current, checkErr := os.Lstat(path)
		if checkErr != nil {
			err = checkErr
		} else if !current.Mode().IsRegular() || !os.SameFile(current, stat) {
			err = errors.New("时间记录日文件已替换或不是普通文件")
		}
	}
	if err != nil {
		_ = file.Close()
		return nil, 0, err
	}
	var header [timelineHeaderSize]byte
	_, err = file.ReadAt(header[:], 0)
	invalid := errors.Is(err, io.EOF) || err == nil && (string(header[:8]) != s.magic || binary.LittleEndian.Uint64(header[8:16]) != s.signature ||
		binary.LittleEndian.Uint64(header[16:24]) != uint64(day) || binary.LittleEndian.Uint32(header[24:28]) != uint32(s.total) ||
		binary.LittleEndian.Uint32(header[28:]) != crc32.ChecksumIEEE(header[:28]) || stat.Size() < int64(timelineHeaderSize+s.total*2*timelineIndexSize))
	if invalid {
		err = errors.New("时间记录文件格式、流配置或数据结构不兼容")
	}
	if err != nil {
		closeErr := file.Close()
		if !resetInvalid || !invalid || closeErr != nil {
			return nil, 0, errors.Join(err, closeErr)
		}
		if err = initialize(true); err != nil {
			return nil, 0, fmt.Errorf("重建时间记录文件失败: %w", err)
		}
		return s.open(day, create, false)
	}
	return file, stat.Size(), nil
}

func (s *Timeline) index(pair []byte, slot int, size int64) (timelineIndex, error) {
	var result timelineIndex
	var empty [timelineIndexSize]byte
	corrupt := false
	for copyIndex := 0; copyIndex < 2; copyIndex++ {
		data := pair[copyIndex*timelineIndexSize : (copyIndex+1)*timelineIndexSize]
		if bytes.Equal(data, empty[:]) {
			continue
		}
		generation, offset := binary.LittleEndian.Uint64(data[:8]), binary.LittleEndian.Uint64(data[8:16])
		length := binary.LittleEndian.Uint32(data[16:20])
		if generation == 0 || offset > uint64(size) || offset < uint64(timelineHeaderSize+s.total*2*timelineIndexSize) ||
			uint64(length) > uint64(size)-offset || uint64(length) > uint64(s.maximum) ||
			binary.LittleEndian.Uint32(data[24:28]) != uint32(slot) || binary.LittleEndian.Uint32(data[28:]) != crc32.ChecksumIEEE(data[:28]) {
			corrupt = true
			continue
		}
		if generation > result.generation {
			result = timelineIndex{generation: generation, offset: int64(offset), length: length, checksum: binary.LittleEndian.Uint32(data[20:24]), copy: copyIndex}
		}
	}
	if result.generation == 0 {
		if corrupt {
			return result, fmt.Errorf("时间记录槽 %d 的索引损坏", slot)
		}
		return result, ErrNotFound
	}
	return result, nil
}

func (s *Timeline) read(file *os.File, index timelineIndex) ([]byte, error) {
	data := make([]byte, int(index.length))
	if _, err := file.ReadAt(data, index.offset); err != nil {
		return nil, err
	}
	if crc32.ChecksumIEEE(data) != index.checksum {
		return nil, errors.New("时间记录内容校验失败")
	}
	return data, nil
}

// Cleanup 删除完全早于 cutoff 的日文件，边界日交由调用方按时间过滤。
func (s *Timeline) Cleanup(cutoff int64) error {
	if cutoff < 0 || cutoff > timelineMaxTime {
		return errors.New("时间记录清理时间无效")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return os.ErrClosed
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		day, ok := timelineDay(entry.Name())
		if !ok || day+timelineDayMillis > cutoff {
			continue
		}
		if !entry.Type().IsRegular() {
			return errors.New("时间记录日文件不是普通文件")
		}
		if err := os.Remove(filepath.Join(s.dir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func (s *Timeline) Size() (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return 0, os.ErrClosed
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return 0, err
	}
	var size int64
	for _, entry := range entries {
		if _, ok := timelineDay(entry.Name()); !ok {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return size, err
		}
		if !info.Mode().IsRegular() {
			return size, errors.New("时间记录日文件不是普通文件")
		}
		size += info.Size()
	}
	return size, nil
}

func (s *Timeline) Close() error {
	if !s.owner {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return nil
}

func timelineDay(name string) (int64, bool) {
	if !strings.HasSuffix(name, ".data") {
		return 0, false
	}
	day, err := time.Parse("2006-01-02", strings.TrimSuffix(name, ".data"))
	return day.UnixMilli(), err == nil && day.UnixMilli() >= 0 && day.Format("2006-01-02")+".data" == name
}
