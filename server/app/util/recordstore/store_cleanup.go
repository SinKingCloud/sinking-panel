package recordstore

import (
	"context"
	"encoding/binary"
	"errors"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Cleanup 按过期索引分批回收数据，由 service 调度；读取时无需等待清理即可屏蔽过期记录。
func (s *Store) Cleanup(ctx context.Context) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	var names []string
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(storeRootBucket).ForEach(func(key, value []byte) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			if value == nil {
				names = append(names, string(key))
			}
			return nil
		})
	})
	if err != nil {
		return 0, err
	}
	total := 0
	for _, name := range names {
		collection := &Collection{store: s, name: name}
		count, err := collection.Cleanup(ctx)
		total += count
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

func (c *Collection) Cleanup(ctx context.Context) (int, error) {
	now := time.Now().UnixMilli()
	total := 0
	for {
		if err := ctx.Err(); err != nil {
			return total, err
		}
		count := 0
		err := c.store.db.Update(func(tx *bolt.Tx) error {
			records, _, expires, err := c.buckets(tx)
			if err != nil {
				return err
			}
			keys := make([]string, 0, 256)
			cursor := expires.Cursor()
			for key, _ := cursor.First(); key != nil && len(keys) < 256; key, _ = cursor.Next() {
				if err := ctx.Err(); err != nil {
					return err
				}
				if len(key) <= 8 {
					return errors.New("记录过期索引无效")
				}
				if binary.BigEndian.Uint64(key[:8]) > uint64(now) {
					break
				}
				record, err := decodeRecord(string(key[8:]), records.Get(key[8:]))
				if err != nil {
					return err
				}
				if record.ExpiresAt != int64(binary.BigEndian.Uint64(key[:8])) {
					return errors.New("记录过期索引与元数据不匹配")
				}
				keys = append(keys, record.Key)
			}
			// 遍历结束再删除，避免修改当前游标导致漏清理相邻记录。
			for _, key := range keys {
				if err := ctx.Err(); err != nil {
					return err
				}
				if err := c.deleteRecord(tx, key); err != nil {
					return err
				}
				count++
			}
			return nil
		})
		if err != nil {
			return total, err
		}
		total += count
		if count < 256 {
			return total, nil
		}
	}
}
