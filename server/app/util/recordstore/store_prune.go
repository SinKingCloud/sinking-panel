package recordstore

import (
	"context"

	bolt "go.etcd.io/bbolt"
)

// DeleteBefore 按业务时间索引分批删除严格早于 before 的记录；不扫描保留范围。
// 每批最多 256 条并单独提交，适合 service 动态缩短保留期。
func (c *Collection) DeleteBefore(ctx context.Context, before int64) (int, error) {
	total := 0
	for before > 0 {
		if err := ctx.Err(); err != nil {
			return total, err
		}
		count := 0
		err := c.store.db.Update(func(tx *bolt.Tx) error {
			_, order, _, err := c.buckets(tx)
			if err != nil {
				return err
			}
			keys := make([]string, 0, 256)
			cursor := order.Cursor()
			for key, _ := cursor.First(); key != nil && len(keys) < 256; key, _ = cursor.Next() {
				if err := ctx.Err(); err != nil {
					return err
				}
				timestamp, err := storeQueryIndexTime(key)
				if err != nil {
					return err
				}
				if timestamp >= before {
					break
				}
				keys = append(keys, string(key[8:]))
			}
			for _, key := range keys {
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
			break
		}
	}
	return total, nil
}
