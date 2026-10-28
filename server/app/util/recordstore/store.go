package recordstore

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"

	bolt "go.etcd.io/bbolt"
)

const storeHeaderSize = 18

var (
	storeRootBucket   = []byte("recordstore.collections.v1")
	storeValuesBucket = []byte("records")
	storeTimeBucket   = []byte("time")
	storeExpiryBucket = []byte("expires")
)

// Store 是独立于业务数据结构的持久化存储。同一路径共享一个实例，退出时 Close。
// 采用 bbolt 默认同步事务；不同集合共享文件，但主键、时间和过期索引相互隔离。
type Store struct {
	db            *bolt.DB
	maxValueBytes int
}

type Collection struct {
	store *Store
	name  string
}

func Open(dir string, options ...Options) (*Store, error) {
	if len(options) > 1 {
		return nil, errors.New("记录存储只接受一组选项")
	}
	maximum := 16 << 20
	if len(options) == 1 && options[0].MaxValueBytes != 0 {
		maximum = options[0].MaxValueBytes
	}
	if maximum < 1 || int64(maximum) > bolt.MaxValueSize-storeHeaderSize {
		return nil, errors.New("单条记录大小限制无效")
	}
	if dir == "" {
		return nil, errors.New("记录存储目录不能为空")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "records.db")
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return nil, errors.New("记录存储文件不是普通文件")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, fmt.Errorf("打开记录存储失败: %w", err)
	}
	if err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(storeRootBucket)
		return err
	}); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db, maxValueBytes: maximum}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// Collection 打开或创建命名集合；集合名称不会作为文件系统路径使用。
func (s *Store) Collection(name string) (*Collection, error) {
	if name == "" || len(name) > 128 || !utf8.ValidString(name) {
		return nil, errors.New("集合名称须为 1 至 128 字节的有效文本")
	}
	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.Bucket(storeRootBucket).CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return err
		}
		for _, name := range [][]byte{storeValuesBucket, storeTimeBucket, storeExpiryBucket} {
			if _, err = bucket.CreateBucketIfNotExists(name); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &Collection{store: s, name: name}, nil
}

func (c *Collection) buckets(tx *bolt.Tx) (*bolt.Bucket, *bolt.Bucket, *bolt.Bucket, error) {
	root := tx.Bucket(storeRootBucket)
	if root == nil {
		return nil, nil, nil, errors.New("记录存储结构不存在")
	}
	bucket := root.Bucket([]byte(c.name))
	if bucket == nil {
		return nil, nil, nil, errors.New("记录集合不存在")
	}
	records, order, expires := bucket.Bucket(storeValuesBucket), bucket.Bucket(storeTimeBucket), bucket.Bucket(storeExpiryBucket)
	if records == nil || order == nil || expires == nil {
		return nil, nil, nil, errors.New("记录集合索引不完整")
	}
	return records, order, expires, nil
}

func (c *Collection) Put(key string, value interface{}, options ...WriteOptions) error {
	if err := storeValidateKey(key); err != nil {
		return err
	}
	option, err := storeWriteOption(options)
	if err != nil {
		return err
	}
	_, err = c.PutMany([]Entry{{Key: key, Value: value, Options: option}})
	return err
}

// PutBytes 原样保存调用方编码的数据，空字节与不存在的记录有明确区别。
func (c *Collection) PutBytes(key string, value []byte, options ...WriteOptions) error {
	if err := storeValidateKey(key); err != nil {
		return err
	}
	option, err := storeWriteOption(options)
	if err != nil {
		return err
	}
	record, err := c.prepareRecord(key, value, EncodingBytes, option)
	if err != nil {
		return err
	}
	// 独立副本避免底层事务保留调用方切片。
	record.Data = append([]byte(nil), value...)
	return c.store.db.Update(func(tx *bolt.Tx) error {
		storeSetRecordTime(&record, option, time.Now().UnixMilli())
		return c.putRecord(tx, record)
	})
}

func (c *Collection) Append(value interface{}, options ...WriteOptions) (string, error) {
	option, err := storeWriteOption(options)
	if err != nil {
		return "", err
	}
	keys, err := c.PutMany([]Entry{{Value: value, Options: option}})
	if err != nil {
		return "", err
	}
	return keys[0], nil
}

// PutMany 在一次事务中提交整个批次，任一写入失败则全部回滚，返回主键与输入一一对应。
func (c *Collection) PutMany(entries []Entry) ([]string, error) {
	prepared := make([]Record, 0, len(entries))
	for _, entry := range entries {
		if entry.Key != "" {
			if err := storeValidateKey(entry.Key); err != nil {
				return nil, err
			}
		}
		data, err := json.Marshal(entry.Value)
		if err != nil {
			return nil, fmt.Errorf("编码记录失败: %w", err)
		}
		record, err := c.prepareRecord(entry.Key, data, EncodingJSON, entry.Options)
		if err != nil {
			return nil, err
		}
		prepared = append(prepared, record)
	}
	keys := make([]string, len(entries))
	if len(entries) == 0 {
		return keys, nil
	}
	err := c.store.db.Update(func(tx *bolt.Tx) error {
		records, _, _, err := c.buckets(tx)
		if err != nil {
			return err
		}
		now := time.Now().UnixMilli()
		// 批次中的显式主键也须避让，不能被同批自动生成的主键覆盖。
		reserved := make(map[string]bool, len(entries))
		for _, record := range prepared {
			if record.Key != "" {
				reserved[record.Key] = true
			}
		}
		for index, record := range prepared {
			storeSetRecordTime(&record, entries[index].Options, now)
			if record.Key == "" {
				for {
					sequence, err := records.NextSequence()
					if err != nil {
						return err
					}
					if sequence == 0 {
						return errors.New("记录主键序列已耗尽")
					}
					record.Key = fmt.Sprintf("%016x", sequence)
					if records.Get([]byte(record.Key)) == nil && !reserved[record.Key] {
						break
					}
				}
			}
			if err := c.putRecord(tx, record); err != nil {
				return err
			}
			keys[index] = record.Key
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (c *Collection) prepareRecord(key string, data []byte, encoding string, options WriteOptions) (Record, error) {
	record := Record{Key: key, Time: options.Time, Data: data, Encoding: encoding}
	if len(data) > c.store.maxValueBytes {
		return Record{}, fmt.Errorf("单条记录超过 %d 字节", c.store.maxValueBytes)
	}
	if options.Time < 0 || options.TTL < 0 {
		return Record{}, errors.New("记录时间与存活时长不能为负数")
	}
	return record, nil
}

// 在取得写事务后计算时间，编码及排队等待不会消耗记录的 TTL。
func storeSetRecordTime(record *Record, options WriteOptions, now int64) {
	if record.Time == 0 {
		record.Time = now
	}
	if options.TTL > 0 {
		record.ExpiresAt = now + int64((options.TTL-1)/time.Millisecond+1)
	}
}

func (c *Collection) putRecord(tx *bolt.Tx, record Record) error {
	records, order, expires, err := c.buckets(tx)
	if err != nil {
		return err
	}
	if err = c.deleteRecord(tx, record.Key); err != nil {
		return err
	}
	encoded := make([]byte, storeHeaderSize+len(record.Data))
	encoded[0] = 1
	if record.Encoding == EncodingBytes {
		encoded[1] = 1
	}
	binary.BigEndian.PutUint64(encoded[2:10], uint64(record.Time))
	binary.BigEndian.PutUint64(encoded[10:18], uint64(record.ExpiresAt))
	copy(encoded[storeHeaderSize:], record.Data)
	if err = records.Put([]byte(record.Key), encoded); err != nil {
		return err
	}
	if err = order.Put(storeIndexKey(record.Time, record.Key), nil); err != nil {
		return err
	}
	if record.ExpiresAt > 0 {
		return expires.Put(storeIndexKey(record.ExpiresAt, record.Key), nil)
	}
	return nil
}

func (c *Collection) Get(key string) (Record, error) {
	if err := storeValidateKey(key); err != nil {
		return Record{}, err
	}
	var result Record
	err := c.store.db.View(func(tx *bolt.Tx) error {
		records, _, _, err := c.buckets(tx)
		if err != nil {
			return err
		}
		encoded := records.Get([]byte(key))
		if encoded == nil {
			return ErrNotFound
		}
		result, err = decodeRecord(key, encoded)
		if err != nil {
			return err
		}
		if result.ExpiresAt > 0 && result.ExpiresAt <= time.Now().UnixMilli() {
			return ErrNotFound
		}
		result.Data = append([]byte(nil), result.Data...)
		return nil
	})
	if err != nil {
		return Record{}, err
	}
	return result, nil
}

func (c *Collection) GetInto(key string, target interface{}) error {
	record, err := c.Get(key)
	if err != nil {
		return err
	}
	return record.Decode(target)
}

// Delete 幂等删除记录，并在同一事务内删除所有索引。
func (c *Collection) Delete(key string) error {
	if err := storeValidateKey(key); err != nil {
		return err
	}
	return c.store.db.Update(func(tx *bolt.Tx) error { return c.deleteRecord(tx, key) })
}

func (c *Collection) deleteRecord(tx *bolt.Tx, key string) error {
	records, order, expires, err := c.buckets(tx)
	if err != nil {
		return err
	}
	encoded := records.Get([]byte(key))
	if encoded == nil {
		return nil
	}
	record, err := decodeRecord(key, encoded)
	if err != nil {
		return err
	}
	if err = order.Delete(storeIndexKey(record.Time, key)); err != nil {
		return err
	}
	if record.ExpiresAt > 0 {
		if err = expires.Delete(storeIndexKey(record.ExpiresAt, key)); err != nil {
			return err
		}
	}
	return records.Delete([]byte(key))
}

func decodeRecord(key string, encoded []byte) (Record, error) {
	if len(encoded) < storeHeaderSize || encoded[0] != 1 || encoded[1] > 1 {
		return Record{}, errors.New("记录格式无效或版本不兼容")
	}
	record := Record{Key: key, Time: int64(binary.BigEndian.Uint64(encoded[2:10])),
		ExpiresAt: int64(binary.BigEndian.Uint64(encoded[10:18])), Encoding: EncodingJSON, Data: encoded[storeHeaderSize:]}
	if record.Time <= 0 || record.ExpiresAt < 0 {
		return Record{}, errors.New("记录时间元数据无效")
	}
	if encoded[1] == 1 {
		record.Encoding = EncodingBytes
	}
	return record, nil
}

func storeIndexKey(timestamp int64, key string) []byte {
	encoded := make([]byte, 8+len(key))
	binary.BigEndian.PutUint64(encoded[:8], uint64(timestamp))
	copy(encoded[8:], key)
	return encoded
}

func storeValidateKey(key string) error {
	if key == "" || len(key) > bolt.MaxKeySize-8 || !utf8.ValidString(key) {
		return errors.New("记录主键为空、过长或不是有效文本")
	}
	return nil
}

func storeWriteOption(options []WriteOptions) (WriteOptions, error) {
	if len(options) > 1 {
		return WriteOptions{}, errors.New("记录写入只接受一组选项")
	}
	if len(options) == 1 {
		return options[0], nil
	}
	return WriteOptions{}, nil
}
