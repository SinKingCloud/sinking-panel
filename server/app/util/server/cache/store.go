package cache

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/darkweak/storages/core"
	"github.com/pierrec/lz4/v4"
	bolt "go.etcd.io/bbolt"
)

const databaseName = "cache.db"

var (
	valuesBucket = []byte("values")
	orderBucket  = []byte("order")
	metaBucket   = []byte("meta")
	totalKey     = []byte("logical_size")
	sequenceKey  = []byte("sequence")
)

type backend struct {
	path string
	refs int

	mu      sync.RWMutex
	db      *bolt.DB
	retired bool
}

func (b *backend) view(run func(*bolt.Tx) error) error {
	for {
		b.mu.RLock()
		if b.retired {
			b.mu.RUnlock()
			return errStoreClosed
		}
		if b.db != nil {
			err := b.db.View(run)
			b.mu.RUnlock()
			return err
		}
		b.mu.RUnlock()
		if _, err := os.Stat(filepath.Join(b.path, databaseName)); errors.Is(err, os.ErrNotExist) {
			return nil
		} else if err != nil {
			return fmt.Errorf("读取 simplefs 数据库状态失败: %w", err)
		}
		if err := b.open(); err != nil {
			return err
		}
	}
}

func (b *backend) update(run func(*bolt.Tx) error) error {
	for {
		b.mu.RLock()
		if b.retired {
			b.mu.RUnlock()
			return errStoreClosed
		}
		if b.db != nil {
			err := b.db.Update(run)
			b.mu.RUnlock()
			return err
		}
		b.mu.RUnlock()
		if err := b.open(); err != nil {
			return err
		}
	}
}

func (b *backend) open() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.retired {
		return errStoreClosed
	}
	if b.db != nil {
		return nil
	}
	if err := os.MkdirAll(b.path, 0o700); err != nil {
		return fmt.Errorf("创建 simplefs 目录失败: %w", err)
	}
	if err := os.Chmod(b.path, 0o700); err != nil {
		return fmt.Errorf("设置 simplefs 目录权限失败: %w", err)
	}
	databasePath := filepath.Join(b.path, databaseName)
	if info, err := os.Lstat(databasePath); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("simplefs 数据库不能是符号链接")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("读取 simplefs 数据库状态失败: %w", err)
	}
	db, err := bolt.Open(databasePath, 0o600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return fmt.Errorf("打开 simplefs 数据库失败: %w", err)
	}
	if err = os.Chmod(databasePath, 0o600); err != nil {
		_ = db.Close()
		return fmt.Errorf("设置 simplefs 数据库权限失败: %w", err)
	}
	err = db.Update(func(transaction *bolt.Tx) error {
		values, createErr := transaction.CreateBucketIfNotExists(valuesBucket)
		if createErr != nil {
			return createErr
		}
		if deleteErr := transaction.DeleteBucket(orderBucket); deleteErr != nil && !errors.Is(deleteErr, bolt.ErrBucketNotFound) {
			return deleteErr
		}
		order, createErr := transaction.CreateBucket(orderBucket)
		if createErr != nil {
			return createErr
		}
		meta, createErr := transaction.CreateBucketIfNotExists(metaBucket)
		if createErr != nil {
			return createErr
		}
		var total uint64
		var maximumSequence uint64
		cursor := values.Cursor()
		for key, encoded := cursor.First(); key != nil; key, encoded = cursor.Next() {
			expiresAt, sequence, value, decodeErr := Registry.decodeRecord(encoded)
			if decodeErr != nil || expiresAt > 0 && expiresAt <= time.Now().UnixNano() {
				if sequence > 0 {
					_ = order.Delete(Registry.uint64Bytes(sequence))
				}
				if deleteErr := cursor.Delete(); deleteErr != nil {
					return deleteErr
				}
				continue
			}
			total += uint64(len(value))
			if sequence > maximumSequence {
				maximumSequence = sequence
			}
			if sequence > 0 {
				if putErr := order.Put(Registry.uint64Bytes(sequence), append([]byte(nil), key...)); putErr != nil {
					return putErr
				}
			}
		}
		if err := meta.Put(totalKey, Registry.uint64Bytes(total)); err != nil {
			return err
		}
		return meta.Put(sequenceKey, Registry.uint64Bytes(maximumSequence))
	})
	if err != nil {
		_ = db.Close()
		return fmt.Errorf("初始化 simplefs 数据库失败: %w", err)
	}
	b.db = db
	return nil
}

func (b *backend) close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.retired = true
	if b.db == nil {
		return nil
	}
	err := b.db.Close()
	b.db = nil
	return err
}

func (b *backend) reset() error {
	return b.update(func(transaction *bolt.Tx) error {
		for _, bucket := range [][]byte{valuesBucket, orderBucket, metaBucket} {
			if err := transaction.DeleteBucket(bucket); err != nil && !errors.Is(err, bolt.ErrBucketNotFound) {
				return err
			}
			if _, err := transaction.CreateBucket(bucket); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *store) mapKeys(prefix string) (map[string]string, error) {
	result := map[string]string{}
	now := time.Now().UnixNano()
	err := s.backend.view(func(transaction *bolt.Tx) error {
		cursor := transaction.Bucket(valuesBucket).Cursor()
		search := []byte(prefix)
		for key, encoded := cursor.Seek(search); key != nil && bytes.HasPrefix(key, search); key, encoded = cursor.Next() {
			expiresAt, _, value, err := Registry.decodeRecord(encoded)
			if err != nil || expiresAt > 0 && expiresAt <= now {
				continue
			}
			result[strings.TrimPrefix(string(key), prefix)] = string(value)
		}
		return nil
	})
	return result, err
}

func (s *store) listKeys() ([]string, error) {
	result := map[string]struct{}{}
	now := time.Now()
	err := s.backend.view(func(transaction *bolt.Tx) error {
		cursor := transaction.Bucket(valuesBucket).Cursor()
		for key, encoded := cursor.Seek([]byte(core.MappingKeyPrefix)); key != nil && bytes.HasPrefix(key, []byte(core.MappingKeyPrefix)); key, encoded = cursor.Next() {
			expiresAt, _, value, err := Registry.decodeRecord(encoded)
			if err != nil || expiresAt > 0 && expiresAt <= now.UnixNano() {
				continue
			}
			mapping, err := core.DecodeMapping(value)
			if err != nil {
				continue
			}
			for _, item := range mapping.GetMapping() {
				if item.GetStaleTime() != nil && item.GetStaleTime().AsTime().After(now) {
					result[item.GetRealKey()] = struct{}{}
				}
			}
		}
		return nil
	})
	keys := make([]string, 0, len(result))
	for key := range result {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys, err
}

func (s *store) get(key string) ([]byte, error) {
	var value []byte
	err := s.backend.view(func(transaction *bolt.Tx) error {
		encoded := transaction.Bucket(valuesBucket).Get([]byte(key))
		if encoded == nil {
			return nil
		}
		expiresAt, _, current, err := Registry.decodeRecord(encoded)
		if err != nil {
			return err
		}
		if expiresAt > 0 && expiresAt <= time.Now().UnixNano() {
			return nil
		}
		value = append([]byte(nil), current...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *store) set(key string, value []byte, duration time.Duration) error {
	return s.backend.update(func(transaction *bolt.Tx) error {
		expiresAt := int64(0)
		if duration > 0 {
			expiresAt = time.Now().Add(duration).UnixNano()
		}
		return s.setRecordsTransaction(transaction, map[string]pendingRecord{key: {value: value, expiresAt: expiresAt}})
	})
}

func (s *store) setMultiLevel(baseKey, variedKey string, value []byte, variedHeaders http.Header, etag string, duration time.Duration, realKey string, logger core.Logger) error {
	compressed := new(bytes.Buffer)
	writer := lz4.NewWriter(compressed)
	if _, err := writer.Write(value); err != nil {
		_ = writer.Close()
		return fmt.Errorf("压缩 simplefs 响应失败: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("完成 simplefs 响应压缩失败: %w", err)
	}
	now := time.Now()
	mappingKey := core.MappingKeyPrefix + baseKey
	return s.backend.update(func(transaction *bolt.Tx) error {
		var mappingValue []byte
		if encoded := transaction.Bucket(valuesBucket).Get([]byte(mappingKey)); encoded != nil {
			_, _, current, err := Registry.decodeRecord(encoded)
			if err != nil {
				return err
			}
			mappingValue = current
		}
		mappingValue, err := core.MappingUpdater(variedKey, mappingValue, logger, now, now.Add(duration), now.Add(duration+s.stale), variedHeaders, etag, realKey)
		if err != nil {
			return err
		}
		expiresAt := int64(0)
		if duration+s.stale > 0 {
			expiresAt = now.Add(duration + s.stale).UnixNano()
		}
		mappingExpiresAt := expiresAt
		if mapping, decodeErr := core.DecodeMapping(mappingValue); decodeErr == nil {
			for _, item := range mapping.GetMapping() {
				if item.GetStaleTime() != nil && item.GetStaleTime().AsTime().UnixNano() > mappingExpiresAt {
					mappingExpiresAt = item.GetStaleTime().AsTime().UnixNano()
				}
			}
		}
		return s.setRecordsTransaction(transaction, map[string]pendingRecord{
			variedKey:  {value: compressed.Bytes(), expiresAt: expiresAt},
			mappingKey: {value: mappingValue, expiresAt: mappingExpiresAt},
		})
	})
}

func (s *store) setRecordsTransaction(transaction *bolt.Tx, records map[string]pendingRecord) error {
	values := transaction.Bucket(valuesBucket)
	order := transaction.Bucket(orderBucket)
	meta := transaction.Bucket(metaBucket)
	total := Registry.readUint64(meta.Get(totalKey))
	sequence := Registry.readUint64(meta.Get(sequenceKey))
	now := time.Now().UnixNano()

	cursor := values.Cursor()
	for key, encoded := cursor.First(); key != nil; key, encoded = cursor.Next() {
		expiresAt, currentSequence, currentValue, err := Registry.decodeRecord(encoded)
		if err != nil || expiresAt > 0 && expiresAt <= now {
			if err == nil {
				total -= min(total, uint64(len(currentValue)))
			}
			if currentSequence > 0 {
				_ = order.Delete(Registry.uint64Bytes(currentSequence))
			}
			if err := cursor.Delete(); err != nil {
				return err
			}
		}
	}

	for key := range records {
		encoded := values.Get([]byte(key))
		if encoded == nil {
			continue
		}
		_, currentSequence, currentValue, err := Registry.decodeRecord(encoded)
		if err == nil {
			total -= min(total, uint64(len(currentValue)))
			_ = order.Delete(Registry.uint64Bytes(currentSequence))
		}
		if err := values.Delete([]byte(key)); err != nil {
			return err
		}
	}

	var required uint64
	for _, record := range records {
		required += uint64(len(record.value))
	}
	if s.maxSize > 0 && required > uint64(s.maxSize) {
		return fmt.Errorf("simplefs 写入数据 %d 字节超过目录上限 %d 字节", required, s.maxSize)
	}
	for s.maxSize > 0 && total+required > uint64(s.maxSize) {
		orderCursor := order.Cursor()
		var evicted bool
		for sequenceKey, key := orderCursor.First(); sequenceKey != nil; sequenceKey, key = orderCursor.Next() {
			if _, protected := records[string(key)]; protected {
				continue
			}
			removed, err := s.evictGroup(transaction, key)
			if err != nil {
				return err
			}
			total -= min(total, removed)
			evicted = true
			break
		}
		if !evicted {
			return fmt.Errorf("simplefs 无法回收足够空间")
		}
	}

	keys := make([]string, 0, len(records))
	for key := range records {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		sequence++
		record := records[key]
		if err := values.Put([]byte(key), Registry.encodeRecord(record.expiresAt, sequence, record.value)); err != nil {
			return err
		}
		if err := order.Put(Registry.uint64Bytes(sequence), []byte(key)); err != nil {
			return err
		}
		total += uint64(len(record.value))
	}
	if err := meta.Put(totalKey, Registry.uint64Bytes(total)); err != nil {
		return err
	}
	return meta.Put(sequenceKey, Registry.uint64Bytes(sequence))
}

func (s *store) evictGroup(transaction *bolt.Tx, key []byte) (uint64, error) {
	values := transaction.Bucket(valuesBucket)
	keys := map[string]struct{}{string(key): {}}
	if strings.HasPrefix(string(key), core.MappingKeyPrefix) {
		if encoded := values.Get(key); encoded != nil {
			_, _, value, err := Registry.decodeRecord(encoded)
			if err != nil {
				return 0, err
			}
			if mapping, err := core.DecodeMapping(value); err == nil {
				for variedKey := range mapping.GetMapping() {
					keys[variedKey] = struct{}{}
				}
			}
		}
	} else {
		cursor := values.Cursor()
		prefix := []byte(core.MappingKeyPrefix)
		for mappingKey, encoded := cursor.Seek(prefix); mappingKey != nil && bytes.HasPrefix(mappingKey, prefix); mappingKey, encoded = cursor.Next() {
			_, _, value, err := Registry.decodeRecord(encoded)
			if err != nil {
				continue
			}
			mapping, err := core.DecodeMapping(value)
			if err != nil {
				continue
			}
			if _, exists := mapping.GetMapping()[string(key)]; !exists {
				continue
			}
			keys[string(mappingKey)] = struct{}{}
			for variedKey := range mapping.GetMapping() {
				keys[variedKey] = struct{}{}
			}
		}
	}

	var removed uint64
	for currentKey := range keys {
		encoded := values.Get([]byte(currentKey))
		if encoded == nil {
			continue
		}
		_, sequence, value, err := Registry.decodeRecord(encoded)
		if err != nil {
			return 0, err
		}
		removed += uint64(len(value))
		if err = values.Delete([]byte(currentKey)); err != nil {
			return 0, err
		}
		if err = transaction.Bucket(orderBucket).Delete(Registry.uint64Bytes(sequence)); err != nil {
			return 0, err
		}
	}
	return removed, nil
}

func (s *store) delete(key string) error {
	return s.backend.update(func(transaction *bolt.Tx) error {
		return s.deleteRecord(transaction, []byte(key))
	})
}

func (s *store) deleteMany(pattern string) error {
	expression, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("解析 simplefs 删除正则失败: %w", err)
	}
	return s.backend.update(func(transaction *bolt.Tx) error {
		values := transaction.Bucket(valuesBucket)
		keys := make([][]byte, 0)
		cursor := values.Cursor()
		for key, _ := cursor.First(); key != nil; key, _ = cursor.Next() {
			if expression.Match(key) {
				keys = append(keys, append([]byte(nil), key...))
			}
		}
		for _, key := range keys {
			if err := s.deleteRecord(transaction, key); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *store) reset() error {
	return s.backend.reset()
}

type pendingRecord struct {
	value     []byte
	expiresAt int64
}

func (s *store) deleteRecord(transaction *bolt.Tx, key []byte) error {
	values := transaction.Bucket(valuesBucket)
	encoded := values.Get(key)
	if encoded == nil {
		return nil
	}
	_, sequence, value, err := Registry.decodeRecord(encoded)
	if err != nil {
		return err
	}
	meta := transaction.Bucket(metaBucket)
	total := Registry.readUint64(meta.Get(totalKey))
	total -= min(total, uint64(len(value)))
	if err = values.Delete(key); err != nil {
		return err
	}
	if err = transaction.Bucket(orderBucket).Delete(Registry.uint64Bytes(sequence)); err != nil {
		return err
	}
	return meta.Put(totalKey, Registry.uint64Bytes(total))
}

func (*StoreRegistry) encodeRecord(expiresAt int64, sequence uint64, value []byte) []byte {
	encoded := make([]byte, 16+len(value))
	binary.BigEndian.PutUint64(encoded[:8], uint64(expiresAt))
	binary.BigEndian.PutUint64(encoded[8:16], sequence)
	copy(encoded[16:], value)
	return encoded
}

func (*StoreRegistry) decodeRecord(encoded []byte) (int64, uint64, []byte, error) {
	if len(encoded) < 16 {
		return 0, 0, nil, errors.New("simplefs 数据记录损坏")
	}
	return int64(binary.BigEndian.Uint64(encoded[:8])), binary.BigEndian.Uint64(encoded[8:16]), encoded[16:], nil
}

func (*StoreRegistry) uint64Bytes(value uint64) []byte {
	encoded := make([]byte, 8)
	binary.BigEndian.PutUint64(encoded, value)
	return encoded
}

func (*StoreRegistry) readUint64(value []byte) uint64 {
	if len(value) != 8 {
		return 0
	}
	return binary.BigEndian.Uint64(value)
}
