package recordstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	bolt "go.etcd.io/bbolt"
)

const (
	storeQueryBatchSize = 128
	storeQueryByteLimit = 16 << 20
)

type storeQueryCursor struct {
	Version  int    `json:"version"`
	Query    string `json:"query"`
	Position []byte `json:"position"`
}

type storeQueryCandidate struct {
	position []byte
	record   *Record
}

// Query 通过业务时间索引读取，每个短事务最多检查 128 个候选并复制 16 MiB 数据。
// Match 在事务外执行；分页游标记录完整时间与 Key，删除游标对应记录不影响续读。
// 每页还受 16 MiB 数据量限制；单条超过此大小时独占一页，仍可通过游标继续。
// 字节边界上的下一条记录可能再次调用 Match，筛选函数不应依赖调用次数。
func (c *Collection) Query(ctx context.Context, query Query) (Page, error) {
	page := Page{Items: make([]Record, 0)}
	if err := ctx.Err(); err != nil {
		return page, err
	}
	if query.Start < 0 || query.End < 0 || (query.Start > 0 && query.End > 0 && query.End < query.Start) {
		return page, errors.New("记录查询时间范围无效")
	}
	if query.Limit == 0 {
		query.Limit = 100
	}
	if query.Limit < 1 || query.Limit > 1000 {
		return page, errors.New("记录查询条数必须为 1 至 1000")
	}
	fingerprint := storeQueryFingerprint(c.name, query)
	after, err := storeDecodeQueryCursor(query.Cursor, fingerprint, query)
	if err != nil {
		return page, err
	}
	pageBytes := 0
	for {
		batch, more, err := c.readQueryBatch(ctx, query, after)
		if err != nil {
			return page, err
		}
		for index, candidate := range batch {
			if err := ctx.Err(); err != nil {
				return page, err
			}
			record := candidate.record
			if record == nil || storeQueryExpired(*record, time.Now().UnixMilli()) {
				after = candidate.position
				continue
			}
			matches := true
			if query.Match != nil {
				matches, err = query.Match(*record)
				if err != nil {
					return page, err
				}
			}
			if err := ctx.Err(); err != nil {
				return page, err
			}
			if !matches || storeQueryExpired(*record, time.Now().UnixMilli()) {
				after = candidate.position
				continue
			}
			// 下一条匹配记录留到下一页；游标不能越过尚未返回的记录。
			if len(page.Items) > 0 && len(record.Data) > storeQueryByteLimit-pageBytes {
				return storeFinishQueryPage(page, after, fingerprint, true), nil
			}
			after = candidate.position
			page.Items = append(page.Items, *record)
			pageBytes += len(record.Data)
			if len(page.Items) >= query.Limit || pageBytes >= storeQueryByteLimit {
				return storeFinishQueryPage(page, after, fingerprint, more || index+1 < len(batch)), nil
			}
		}
		if !more {
			return storeFinishQueryPage(page, after, fingerprint, false), nil
		}
		// 即使当前批次全部被 Prefix、TTL 或 Match 排除，也继续沿索引前进。
		// readQueryBatch 有后续候选时必定返回至少一个已检查的位置。
		if len(batch) == 0 {
			return page, errors.New("记录查询游标未能前进")
		}
	}
}

// Scan 按页流式遍历，visit 与 Match 都在读事务外执行，可安全调用同一存储写入方法。
// 并发写入时每个批次有自己的快照；调用方应避免回调不断追加查询范围内的新记录。
func (c *Collection) Scan(ctx context.Context, query Query, visit func(Record) error) error {
	if visit == nil {
		return errors.New("记录遍历回调不能为空")
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		page, err := c.Query(ctx, query)
		if err != nil {
			return err
		}
		for _, record := range page.Items {
			if err := ctx.Err(); err != nil {
				return err
			}
			if storeQueryExpired(record, time.Now().UnixMilli()) {
				continue
			}
			if err := visit(record); err != nil {
				return err
			}
			if err := ctx.Err(); err != nil {
				return err
			}
		}
		if page.NextCursor == "" {
			return nil
		}
		query.Cursor = page.NextCursor
	}
}

func (c *Collection) readQueryBatch(ctx context.Context, query Query, after []byte) ([]storeQueryCandidate, bool, error) {
	batch := make([]storeQueryCandidate, 0, storeQueryBatchSize)
	more := false
	err := c.store.db.View(func(tx *bolt.Tx) error {
		records, order, _, err := c.buckets(tx)
		if err != nil {
			return err
		}
		cursor := order.Cursor()
		key := storeSeekQueryCursor(cursor, query, after)
		copiedBytes := 0
		for key != nil {
			if err := ctx.Err(); err != nil {
				return err
			}
			timestamp, err := storeQueryIndexTime(key)
			if err != nil {
				return err
			}
			if query.Start > 0 && timestamp < query.Start || query.End > 0 && timestamp > query.End {
				// Seek 已定位到所需边界，之后越界意味着当前方向已没有匹配记录。
				break
			}
			if len(batch) == storeQueryBatchSize {
				more = true
				break
			}
			candidate := storeQueryCandidate{position: bytes.Clone(key)}
			recordKey := key[8:]
			if bytes.HasPrefix(recordKey, []byte(query.Prefix)) {
				encoded := records.Get(recordKey)
				if encoded != nil {
					record, err := decodeRecord(string(recordKey), encoded)
					if err != nil {
						return fmt.Errorf("解码记录失败: %w", err)
					}
					if record.Time != timestamp {
						return errors.New("记录时间索引与数据不一致")
					}
					if !storeQueryExpired(record, time.Now().UnixMilli()) {
						// 超大单条允许独占批次；其余批次限制总复制量，避免长期占用读事务。
						if copiedBytes > 0 && len(record.Data) > storeQueryByteLimit-copiedBytes {
							more = true
							break
						}
						record.Data = bytes.Clone(record.Data)
						copiedBytes += len(record.Data)
						candidate.record = &record
					}
				}
			}
			batch = append(batch, candidate)
			if query.Desc {
				key, _ = cursor.Prev()
			} else {
				key, _ = cursor.Next()
			}
			if copiedBytes >= storeQueryByteLimit {
				if key != nil {
					nextTime, err := storeQueryIndexTime(key)
					if err != nil {
						return err
					}
					more = (query.Start == 0 || nextTime >= query.Start) && (query.End == 0 || nextTime <= query.End)
				}
				break
			}
		}
		return nil
	})
	return batch, more, err
}

func storeSeekQueryCursor(cursor *bolt.Cursor, query Query, after []byte) []byte {
	var key []byte
	if len(after) > 0 {
		key, _ = cursor.Seek(after)
		if query.Desc {
			if key == nil {
				key, _ = cursor.Last()
			} else {
				// 精确命中与游标记录已删除时，都取第一个严格小于游标的键。
				key, _ = cursor.Prev()
			}
		} else if bytes.Equal(key, after) {
			key, _ = cursor.Next()
		}
		return key
	}
	if !query.Desc {
		key, _ = cursor.Seek(storeIndexKey(query.Start, ""))
		return key
	}
	if query.End == 0 || query.End == math.MaxInt64 {
		key, _ = cursor.Last()
		return key
	}
	// End 包含边界，Seek 到下一毫秒的起点后回退，覆盖同毫秒的全部 Key。
	key, _ = cursor.Seek(storeIndexKey(query.End+1, ""))
	if key == nil {
		key, _ = cursor.Last()
	} else {
		key, _ = cursor.Prev()
	}
	return key
}

func storeQueryIndexTime(key []byte) (int64, error) {
	if len(key) <= 8 {
		return 0, errors.New("记录时间索引格式无效")
	}
	value := binary.BigEndian.Uint64(key[:8])
	if value > math.MaxInt64 {
		return 0, errors.New("记录时间索引超出有效范围")
	}
	return int64(value), nil
}

func storeQueryExpired(record Record, now int64) bool {
	return record.ExpiresAt > 0 && record.ExpiresAt <= now
}

func storeFinishQueryPage(page Page, after []byte, fingerprint string, more bool) Page {
	// Match 可能耗时，返回前再次排除处理期间已到期的记录。
	now := time.Now().UnixMilli()
	valid := page.Items[:0]
	for _, record := range page.Items {
		if !storeQueryExpired(record, now) {
			valid = append(valid, record)
		}
	}
	clear(page.Items[len(valid):])
	page.Items = valid
	if more && len(after) > 0 {
		data, _ := json.Marshal(storeQueryCursor{Version: 1, Query: fingerprint, Position: after})
		page.NextCursor = base64.RawURLEncoding.EncodeToString(data)
	}
	return page
}

func storeQueryFingerprint(collection string, query Query) string {
	encoded, _ := json.Marshal(struct {
		Collection string
		Start      int64
		End        int64
		Prefix     string
		Desc       bool
	}{collection, query.Start, query.End, query.Prefix, query.Desc})
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}

func storeDecodeQueryCursor(encoded, fingerprint string, query Query) ([]byte, error) {
	if encoded == "" {
		return nil, nil
	}
	if len(encoded) > 128<<10 {
		return nil, errors.New("记录查询游标无效")
	}
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.New("记录查询游标无效")
	}
	var cursor storeQueryCursor
	if json.Unmarshal(data, &cursor) != nil || cursor.Version != 1 || cursor.Query != fingerprint || len(cursor.Position) > bolt.MaxKeySize {
		return nil, errors.New("记录查询游标无效或查询条件已改变")
	}
	timestamp, err := storeQueryIndexTime(cursor.Position)
	if err != nil || query.Start > 0 && timestamp < query.Start || query.End > 0 && timestamp > query.End {
		return nil, errors.New("记录查询游标超出时间范围")
	}
	return cursor.Position, nil
}
