package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	client "github.com/redis/go-redis/v9"
)

// Keys 分页列举键。
func (m *Manager) Keys(ctx context.Context, match, cursorText string, pageSize int) (result interface{}, err error) {
	if pageSize < 0 {
		return nil, errors.New("分页容量不合法")
	}
	if len(match) > 65536 || len(cursorText) > 4<<20 {
		return nil, errors.New("操作参数超过大小限制")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	size := pageSize
	if size <= 0 {
		size = 100
	}
	if size > 500 {
		size = 500
	}
	var state struct {
		Cursor uint64   `json:"cursor"`
		Keys   []string `json:"keys,omitempty"`
	}
	if cursorText != "" && cursorText != "0" {
		if value, parseErr := strconv.ParseUint(cursorText, 10, 64); parseErr == nil {
			state.Cursor = value
		} else {
			data, decodeErr := base64.RawURLEncoding.DecodeString(cursorText)
			if decodeErr != nil || len(data) > 4<<20 || json.Unmarshal(data, &state) != nil {
				return nil, errors.New("Redis 分页游标不合法")
			}
		}
	}
	keys := state.Keys
	if len(keys) == 0 {
		keys, state.Cursor, err = m.scanKeys(ctx, state.Cursor, match, int64(size))
		if err != nil {
			return nil, err
		}
	}
	unsupported := 0
	visible := keys[:0]
	for _, key := range keys {
		if utf8.ValidString(key) {
			visible = append(visible, key)
		} else {
			unsupported++
		}
	}
	keys = visible
	state.Keys = nil
	if len(keys) > size {
		state.Keys = keys[size:]
		keys = keys[:size]
	}
	rows := make([]map[string]interface{}, 0, len(keys))
	pipe := m.client.Pipeline()
	types := make([]*client.StatusCmd, len(keys))
	ttls := make([]*client.DurationCmd, len(keys))
	for i, key := range keys {
		types[i] = pipe.Type(ctx, key)
		ttls[i] = pipe.TTL(ctx, key)
	}
	if _, err = pipe.Exec(ctx); err != nil && err != client.Nil {
		return nil, err
	}
	for i, key := range keys {
		if types[i].Val() != "none" {
			ttl := int64(ttls[i].Val() / time.Second)
			if ttls[i].Val() < 0 {
				ttl = int64(ttls[i].Val())
			}
			rows = append(rows, map[string]interface{}{
				"key":  key,
				"type": types[i].Val(),
				"ttl":  ttl,
			})
		}
	}
	cursor := "0"
	if state.Cursor != 0 || len(state.Keys) > 0 {
		encoded, encodeErr := json.Marshal(state)
		if encodeErr != nil || len(encoded) > 3<<20 {
			return nil, errors.New("分页快照超过大小限制，请缩小键名筛选范围")
		}
		cursor = base64.RawURLEncoding.EncodeToString(encoded)
	}
	return map[string]interface{}{
		"list":                    rows,
		"cursor":                  cursor,
		"has_more":                cursor != "0",
		"unsupported_binary_keys": unsupported,
	}, nil
}

// Value 读取键值、类型和有效期。
func (m *Manager) Value(ctx context.Context, key, match, cursorText string, start, pageSize int) (result interface{}, err error) {
	if start < 0 || pageSize < 0 {
		return nil, errors.New("分页和起始位置不合法")
	}
	if len(key) > 65536 || len(match) > 65536 || len(cursorText) > 4<<20 {
		return nil, errors.New("操作参数超过大小限制")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	size := pageSize
	if size <= 0 {
		size = 100
	}
	if size > 500 {
		size = 500
	}
	kind, e := m.client.Type(ctx, key).Result()
	if e != nil {
		return nil, e
	}
	if kind == "none" {
		return nil, errors.New("键不存在或已过期")
	}
	ttl, e := m.client.TTL(ctx, key).Result()
	if e != nil {
		return nil, e
	}
	seconds := int64(ttl / time.Second)
	if ttl < 0 {
		seconds = int64(ttl)
	}
	value := map[string]interface{}{
		"key":       key,
		"type":      kind,
		"ttl":       seconds,
		"cursor":    "0",
		"has_more":  false,
		"truncated": false,
	}
	cursor := uint64(0)
	if cursorText != "" {
		cursor, e = strconv.ParseUint(cursorText, 10, 64)
		if e != nil && kind != "stream" {
			return nil, errors.New("Redis 分页游标不合法")
		}
	}
	const maxValue = 8 << 20
	switch kind {
	case "string":
		length, e := m.client.StrLen(ctx, key).Result()
		if e != nil {
			return nil, e
		}
		text, e := m.client.GetRange(ctx, key, 0, maxValue-1).Result()
		if e != nil {
			return nil, e
		}
		value["size"], value["value"], value["truncated"] = length, text, length > maxValue
		if !utf8.ValidString(text) {
			value["value"] = base64.StdEncoding.EncodeToString([]byte(text))
			value["encoding"] = "base64"
		}
	case "hash":
		length, e := m.client.HLen(ctx, key).Result()
		if e != nil {
			return nil, e
		}
		items, next, e := m.scanValue(ctx, "hash", key, cursor, match, int64(size))
		if e != nil {
			return nil, e
		}
		data := map[string]string{}
		bytes := 0
		for i := 0; i+1 < len(items); i += 2 {
			bytes += len(items[i]) + len(items[i+1])
			if bytes > maxValue {
				value["truncated"] = true
				break
			}
			data[items[i]] = items[i+1]
		}
		value["size"], value["value"], value["cursor"], value["has_more"] = length, data, strconv.FormatUint(next, 10), next != 0
	case "list", "set":
		var items []string
		var length int64
		var next uint64
		if kind == "list" {
			length, e = m.client.LLen(ctx, key).Result()
			if e == nil {
				items, e = m.client.LRange(ctx, key, int64(start), int64(start+size-1)).Result()
			}
			value["has_more"] = int64(start+len(items)) < length
		} else {
			length, e = m.client.SCard(ctx, key).Result()
			if e == nil {
				items, next, e = m.scanValue(ctx, "set", key, cursor, match, int64(size))
			}
			value["cursor"], value["has_more"] = strconv.FormatUint(next, 10), next != 0
		}
		if e != nil {
			return nil, e
		}
		bytes := 0
		for i, text := range items {
			bytes += len(text)
			if bytes > maxValue {
				items = items[:i]
				value["truncated"] = true
				break
			}
		}
		if items == nil {
			items = []string{}
		}
		value["size"], value["value"] = length, items
	case "zset":
		length, e := m.client.ZCard(ctx, key).Result()
		if e != nil {
			return nil, e
		}
		items, e := m.client.ZRangeWithScores(ctx, key, int64(start), int64(start+size-1)).Result()
		if e != nil {
			return nil, e
		}
		data := make([]map[string]interface{}, 0, len(items))
		bytes := 0
		for _, item := range items {
			member := fmt.Sprint(item.Member)
			bytes += len(member)
			if bytes > maxValue {
				value["truncated"] = true
				break
			}
			data = append(data, map[string]interface{}{
				"member": member,
				"score":  item.Score,
			})
		}
		value["size"], value["value"], value["has_more"] = length, data, int64(start+len(items)) < length
	case "stream":
		length, e := m.client.XLen(ctx, key).Result()
		if e != nil {
			return nil, e
		}
		start := "-"
		if cursorText != "" && cursorText != "0" {
			start = cursorText
		}
		items, e := m.client.XRangeN(ctx, key, start, "+", int64(size+2)).Result()
		if e != nil {
			return nil, e
		}
		if start != "-" && len(items) > 0 && items[0].ID == start {
			items = items[1:]
		}
		more := len(items) > size
		if more {
			items = items[:size]
		}
		data := make([]map[string]interface{}, 0, len(items))
		bytes := 0
		for _, item := range items {
			encoded, _ := json.Marshal(item.Values)
			bytes += len(encoded)
			if bytes > maxValue {
				value["truncated"] = true
				break
			}
			data = append(data, map[string]interface{}{
				"id":     item.ID,
				"values": item.Values,
			})
		}
		next := "0"
		if len(data) > 0 {
			next = data[len(data)-1]["id"].(string)
		}
		value["size"], value["value"], value["cursor"], value["has_more"] = length, data, next, more
	default:
		return nil, fmt.Errorf("暂不支持可视化读取 %s 类型，请使用命令操作", kind)
	}
	// 非 UTF-8 内容只读展示，避免 JSON 替换原始字节后被整体覆盖。
	binaryValue := false
	var display func(interface{}) interface{}
	display = func(item interface{}) interface{} {
		switch data := item.(type) {
		case string:
			if !utf8.ValidString(data) {
				binaryValue = true
				return "base64:" + base64.StdEncoding.EncodeToString([]byte(data))
			}
			return data
		case []string:
			for i := range data {
				data[i] = display(data[i]).(string)
			}
			return data
		case map[string]string:
			out := make(map[string]string, len(data))
			for key, entry := range data {
				out[display(key).(string)] = display(entry).(string)
			}
			return out
		case map[string]interface{}:
			out := make(map[string]interface{}, len(data))
			for key, entry := range data {
				out[display(key).(string)] = display(entry)
			}
			return out
		case []map[string]interface{}:
			for i := range data {
				data[i] = display(data[i]).(map[string]interface{})
			}
			return data
		default:
			return item
		}
	}
	value["value"] = display(value["value"])
	if binaryValue {
		value["encoding"] = "base64"
	}
	return value, nil
}

// Save 原子保存键；ttl 为 nil 保留原有效期，-1 永不过期，0 立即过期（单位秒）。
func (m *Manager) Save(ctx context.Context, key, kind, value string, ttl *int64) (result interface{}, err error) {
	if len(key) > 65536 {
		return nil, errors.New("键名超过大小限制")
	}
	if ttl != nil && (*ttl < -1 || *ttl > math.MaxInt64/int64(time.Second)) {
		return nil, errors.New("有效期不合法")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	if len(value) > 8<<20 {
		return nil, errors.New("单次写入内容不能超过 8 MB")
	}
	var payload interface{}
	switch kind {
	case "string":
		payload = value
	case "hash":
		var values map[string]string
		if json.Unmarshal([]byte(value), &values) != nil || len(values) == 0 {
			return nil, errors.New("Hash 内容必须是非空的字符串对象")
		}
		payload = values
	case "list", "set":
		var values []string
		if json.Unmarshal([]byte(value), &values) != nil || len(values) == 0 {
			return nil, errors.New("内容必须是非空字符串数组")
		}
		args := make([]interface{}, len(values))
		for i, v := range values {
			args[i] = v
		}
		payload = args
	case "zset":
		var values []struct {
			Member string  `json:"member"`
			Score  float64 `json:"score"`
		}
		if json.Unmarshal([]byte(value), &values) != nil || len(values) == 0 {
			return nil, errors.New("有序集合必须是包含 member、score 的非空数组")
		}
		args := make([]client.Z, len(values))
		for i, v := range values {
			if math.IsNaN(v.Score) || math.IsInf(v.Score, 0) {
				return nil, errors.New("有序集合分值不合法")
			}
			args[i] = client.Z{
				Member: v.Member,
				Score:  v.Score,
			}
		}
		payload = args
	default:
		return nil, errors.New("此类型不能整体保存，请使用命令操作")
	}
	var nonce [16]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return nil, err
	}
	temporary := "__panel_edit_" + hex.EncodeToString(nonce[:])
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = m.client.Del(cleanup, temporary).Err()
	}()
	pipe := m.client.TxPipeline()
	switch kind {
	case "string":
		pipe.Set(ctx, temporary, payload, 0)
	case "hash":
		pipe.HMSet(ctx, temporary, payload)
	case "list":
		pipe.RPush(ctx, temporary, payload.([]interface{})...)
	case "set":
		pipe.SAdd(ctx, temporary, payload.([]interface{})...)
	case "zset":
		pipe.ZAdd(ctx, temporary, payload.([]client.Z)...)
	}
	// 创建临时键和失效时间在同一事务，面板异常退出也不会留下永久临时数据。
	pipe.PExpire(ctx, temporary, 5*time.Minute)
	if _, err = pipe.Exec(ctx); err != nil {
		return nil, err
	}

	// 在替换的同一原子操作中读取旧 TTL，失败时不删除原键。
	ttlMillis := int64(-2)
	if ttl != nil {
		ttlMillis = *ttl
		if ttlMillis >= 0 {
			ttlMillis *= 1000
		}
	}
	err = m.client.Eval(ctx, `local ttl=tonumber(ARGV[1]); if ttl==-2 then ttl=redis.call('PTTL',KEYS[2]); end; redis.call('RENAME',KEYS[1],KEYS[2]); if ttl>=0 then redis.call('PEXPIRE',KEYS[2],ttl); else redis.call('PERSIST',KEYS[2]); end; return 1`, []string{temporary, key}, ttlMillis).Err()
	return map[string]interface{}{"affected": 1}, err
}

// Delete 删除指定数据。
func (m *Manager) Delete(ctx context.Context, keys []string) (result interface{}, err error) {
	if len(keys) == 0 || len(keys) > 1000 {
		return nil, errors.New("请选择 1 至 1000 个键")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	bytes := 0
	for _, key := range keys {
		bytes += len(key)
	}
	if bytes > 1<<20 {
		return nil, errors.New("键名列表超过大小限制")
	}
	affected, e := m.client.Del(ctx, keys...).Result()
	return map[string]interface{}{"affected": affected}, e
}

// Rename 重命名键，禁止覆盖已存在的目标。
func (m *Manager) Rename(ctx context.Context, key, newKey string) (result interface{}, err error) {
	if len(key) > 65536 || len(newKey) > 65536 {
		return nil, errors.New("键名超过大小限制")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	if newKey == key {
		return map[string]interface{}{"affected": 0}, nil
	}
	ok, e := m.client.RenameNX(ctx, key, newKey).Result()
	if e != nil {
		return nil, e
	}
	if !ok {
		return nil, errors.New("目标键已存在")
	}
	return map[string]interface{}{"affected": 1}, nil
}

// Expire 修改键的有效期（秒）；-1 永不过期，0 立即过期。
func (m *Manager) Expire(ctx context.Context, key string, ttl int64) (result interface{}, err error) {
	if len(key) > 65536 {
		return nil, errors.New("键名超过大小限制")
	}
	if ttl < -1 || ttl > math.MaxInt64/int64(time.Second) {
		return nil, errors.New("有效期不合法")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	var ok bool
	var e error
	if ttl == -1 {
		ok, e = m.client.Persist(ctx, key).Result()
	} else {
		ok, e = m.client.Expire(ctx, key, time.Duration(ttl)*time.Second).Result()
	}
	return map[string]interface{}{"changed": ok}, e
}

// Command 执行有界命令；arguments 第一项为命令名，其余项按原样传递。
func (m *Manager) Command(ctx context.Context, arguments []string) (result interface{}, err error) {
	if len(arguments) == 0 || len(arguments) > 10000 {
		return nil, errors.New("请提供 1 至 10000 个命令参数")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	args := make([]interface{}, len(arguments))
	bytes := 0
	for i, value := range arguments {
		bytes += len(value)
		args[i] = value
	}
	if bytes > 1<<20 {
		return nil, errors.New("命令参数超过大小限制")
	}
	command := strings.ToUpper(arguments[0])
	switch command {
	case "MONITOR", "SUBSCRIBE", "PSUBSCRIBE", "SSUBSCRIBE", "SYNC", "PSYNC", "BLPOP", "BRPOP", "BRPOPLPUSH", "BLMOVE", "BLMPOP", "BZPOPMIN", "BZPOPMAX", "BZMPOP", "SHUTDOWN", "DEBUG", "MULTI", "EXEC", "DISCARD", "WATCH", "UNWATCH", "AUTH", "HELLO", "SELECT", "QUIT", "RESET", "EVAL", "EVALSHA", "EVAL_RO", "EVALSHA_RO", "FCALL", "FCALL_RO":
		return nil, errors.New("此命令需要独立会话或可能阻塞服务器，请使用终端执行")
	}
	if command == "KEYS" {
		return nil, errors.New("请使用 SCAN 浏览键，避免阻塞数据库")
	}
	if command == "XREAD" || command == "XREADGROUP" {
		for _, arg := range arguments[1:] {
			if strings.EqualFold(arg, "BLOCK") {
				return nil, errors.New("命令模式不支持 BLOCK 选项")
			}
		}
	}
	if command == "CLIENT" && len(arguments) > 1 {
		switch strings.ToUpper(arguments[1]) {
		case "PAUSE", "REPLY", "TRACKING":
			return nil, errors.New("此命令会改变连接状态，请使用终端执行")
		}
	}
	value, e := m.client.Do(ctx, args...).Result()
	if e == client.Nil {
		e = nil
		value = nil
	}
	if e != nil {
		return nil, e
	}
	// Redis 整数回复可超过 JavaScript 安全整数范围，命令结果使用十进制文本。
	var preserveNumber func(interface{}) interface{}
	preserveNumber = func(item interface{}) interface{} {
		switch typed := item.(type) {
		case int64:
			return strconv.FormatInt(typed, 10)
		case []interface{}:
			for i := range typed {
				typed[i] = preserveNumber(typed[i])
			}
			return typed
		case map[interface{}]interface{}:
			out := make(map[string]interface{}, len(typed))
			for key, entry := range typed {
				out[fmt.Sprint(key)] = preserveNumber(entry)
			}
			return out
		default:
			return item
		}
	}
	value = preserveNumber(value)
	encoded, e := json.Marshal(value)
	if e != nil {
		return nil, e
	}
	if len(encoded) > 8<<20 {
		return map[string]interface{}{
			"result":    string(encoded[:8<<20]),
			"truncated": true,
		}, nil
	}
	return map[string]interface{}{
		"result":    value,
		"truncated": false,
	}, nil
}

// Info 读取服务器运行信息。
func (m *Manager) Info(ctx context.Context) (result interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	defer func() { err = m.cleanError(err) }()
	info, e := m.client.Info(ctx).Result()
	return map[string]interface{}{
		"database": m.database,
		"info":     info,
	}, e
}
