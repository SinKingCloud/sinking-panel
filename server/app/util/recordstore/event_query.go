package recordstore

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type eventFileSnapshot struct {
	day  string
	size int64
}

type eventCursor struct {
	Day    string `json:"day"`
	Offset int64  `json:"offset"`
	Query  string `json:"query"`
}

// Query 按文件偏移续读，游标绑定时间与过滤条件；相同毫秒的事件也不会漏读。
// 查询取各日文件大小快照，扫描时不持有写锁；分页期间新增的更早日事件应重新查询。
func (s *Events) Query(query EventQuery) (EventPage, error) {
	return s.QueryContext(context.Background(), query)
}

func (s *Events) QueryContext(ctx context.Context, query EventQuery) (EventPage, error) {
	page := EventPage{Items: make([]Event, 0)}
	if err := ctx.Err(); err != nil {
		return page, err
	}
	if query.Limit == 0 {
		query.Limit = 100
	}
	if query.Limit < 1 || query.Limit > 1000 {
		return page, errors.New("事件分页条数必须为 1 至 1000")
	}
	filters, err := eventNormalizeFilters(query.Filters)
	if err != nil {
		return page, err
	}
	snapshots, start, err := s.snapshots(query.Start, query.End)
	if err != nil {
		return page, err
	}
	fingerprint := eventFingerprint(query.Start, query.End, filters)
	cursor, err := eventDecodeCursor(query.Cursor, fingerprint, snapshots)
	if err != nil {
		return page, err
	}
	pageBytes := 0
	for _, snapshot := range snapshots {
		if cursor.Day != "" && snapshot.day < cursor.Day {
			continue
		}
		offset := int64(0)
		if snapshot.day == cursor.Day {
			offset = cursor.Offset
		}
		err = s.scanDay(ctx, snapshot, offset, func(event Event, position int64, size int) (bool, error) {
			if event.Time < start || event.Time > query.End || !eventMatches(event.Data, filters) {
				return false, nil
			}
			// 单条上限之外再限制整页大小，大字段日志不会撑满查询内存。
			if len(page.Items) == query.Limit || (len(page.Items) > 0 && pageBytes+size > 8<<20) {
				page.NextCursor = eventEncodeCursor(eventCursor{Day: snapshot.day, Offset: position, Query: fingerprint})
				return true, nil
			}
			page.Items = append(page.Items, event)
			pageBytes += size
			return false, nil
		})
		if err != nil {
			return page, err
		}
		if page.NextCursor != "" {
			break
		}
	}
	return page, nil
}

func (s *Events) snapshots(start, end int64) ([]eventFileSnapshot, int64, error) {
	if start <= 0 || end < start || end-start > 365*eventDayMillis || end > time.Now().Add(time.Minute).UnixMilli() {
		return nil, start, errors.New("事件查询时间无效，跨度不能超过 365 天")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return nil, start, errors.New("事件存储已关闭")
	}
	cutoff := time.Now().UnixMilli() - int64(s.config.RetentionDays)*eventDayMillis
	start = max(start, cutoff)
	if start > end {
		return nil, start, nil
	}
	first := time.UnixMilli(start).UTC().Format("2006-01-02")
	last := time.UnixMilli(end).UTC().Format("2006-01-02")
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, start, err
	}
	files := make([]eventFileSnapshot, 0)
	for _, entry := range entries {
		if !eventIsDayFile(entry.Name()) {
			continue
		}
		day := strings.TrimSuffix(entry.Name(), eventSuffix)
		if day < first || day > last {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, start, err
		}
		if !info.Mode().IsRegular() {
			return nil, start, fmt.Errorf("事件文件不是普通文件: %s", entry.Name())
		}
		files = append(files, eventFileSnapshot{day: day, size: info.Size()})
	}
	return files, start, nil
}

func (s *Events) scanDay(ctx context.Context, snapshot eventFileSnapshot, offset int64, visit func(Event, int64, int) (bool, error)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	file, err := os.Open(filepath.Join(s.dir, snapshot.day+eventSuffix))
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("事件文件已过期，请重新查询")
	}
	if err != nil {
		return err
	}
	defer file.Close()
	if offset < 0 || offset > snapshot.size {
		return errors.New("事件游标偏移无效")
	}
	if offset > 0 {
		var previous [1]byte
		if _, err := file.ReadAt(previous[:], offset-1); err != nil || previous[0] != '\n' {
			return errors.New("事件游标不在完整记录边界")
		}
	}
	reader := bufio.NewReaderSize(io.NewSectionReader(file, offset, snapshot.size-offset), 64<<10)
	for position := offset; position < snapshot.size; {
		if err := ctx.Err(); err != nil {
			return err
		}
		line, err := eventReadLine(reader)
		if errors.Is(err, io.EOF) {
			// 崩溃留下的不完整尾行在下次 Append 时修复，查询不显示半条事件。
			return nil
		}
		if err != nil {
			return err
		}
		var event Event
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.UseNumber()
		if err = decoder.Decode(&event); err != nil {
			return fmt.Errorf("事件记录损坏 (%s:%d): %w", snapshot.day, position, err)
		}
		var extra interface{}
		if decoder.Decode(&extra) != io.EOF {
			return fmt.Errorf("事件记录包含额外内容 (%s:%d)", snapshot.day, position)
		}
		expectedID := snapshot.day + ":" + strconv.FormatInt(position, 10)
		if event.ID != expectedID || event.Time <= 0 || time.UnixMilli(event.Time).UTC().Format("2006-01-02") != snapshot.day {
			return fmt.Errorf("事件记录索引无效 (%s:%d)", snapshot.day, position)
		}
		nodes := 0
		if err := eventValidateJSON(event.Data, 0, &nodes); err != nil {
			return fmt.Errorf("事件记录字段无效 (%s:%d): %w", snapshot.day, position, err)
		}
		stop, err := visit(event, position, len(line))
		if err != nil || stop {
			return err
		}
		position += int64(len(line))
	}
	return nil
}

func eventReadLine(reader *bufio.Reader) ([]byte, error) {
	var line []byte
	for {
		part, err := reader.ReadSlice('\n')
		if len(line)+len(part) > eventMaxRecord {
			return nil, errors.New("事件记录超过 1 MiB 或文件已损坏")
		}
		line = append(line, part...)
		if err == nil {
			return line, nil
		}
		if !errors.Is(err, bufio.ErrBufferFull) {
			return nil, err
		}
	}
}

func eventFingerprint(start, end int64, filters map[string]interface{}) string {
	encoded, _ := json.Marshal(struct {
		Start   int64
		End     int64
		Filters map[string]interface{}
	}{start, end, filters})
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}

func eventEncodeCursor(cursor eventCursor) string {
	data, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(data)
}

func eventDecodeCursor(encoded, fingerprint string, snapshots []eventFileSnapshot) (eventCursor, error) {
	var cursor eventCursor
	if encoded == "" {
		return cursor, nil
	}
	if len(encoded) > 512 {
		return cursor, errors.New("事件游标无效")
	}
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || json.Unmarshal(data, &cursor) != nil || cursor.Query != fingerprint ||
		!eventIsDayFile(cursor.Day+eventSuffix) || cursor.Offset < 0 {
		return cursor, errors.New("事件游标无效或查询条件已改变")
	}
	for _, snapshot := range snapshots {
		if snapshot.day == cursor.Day && cursor.Offset <= snapshot.size {
			return cursor, nil
		}
	}
	return cursor, errors.New("事件游标已过期或偏移超出文件范围")
}

// eventMatches 对已归一化的标量做精确相等匹配；缺失与显式 nil 不相等。
func eventMatches(fields, filters map[string]interface{}) bool {
	for name, wanted := range filters {
		actual, exists := fields[name]
		if !exists {
			return false
		}
		if number, ok := wanted.(json.Number); ok {
			other, ok := actual.(json.Number)
			if !ok {
				return false
			}
			left, leftOK := new(big.Rat).SetString(string(number))
			right, rightOK := new(big.Rat).SetString(string(other))
			if !leftOK || !rightOK || left.Cmp(right) != 0 {
				return false
			}
		} else if actual != wanted {
			return false
		}
	}
	return true
}
