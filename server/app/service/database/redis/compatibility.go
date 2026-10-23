package redis

import (
	"context"
	"errors"
	"strings"
)

// scanKeys 使用 SCAN；Redis 2.6 没有该命令时，仅允许对小数据库生成有界快照。
// 不把权限错误当作版本差异，也不在大数据库上回退到阻塞的 KEYS。
func (m *Manager) scanKeys(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	keys, next, err := m.client.Scan(ctx, cursor, match, count).Result()
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "unknown command") {
		return keys, next, err
	}
	if cursor != 0 {
		return nil, 0, errors.New("此 Redis 版本不支持 SCAN 游标，请从首页重新加载")
	}
	if match == "" {
		match = "*"
	}
	keys, err = m.client.Eval(ctx, `if redis.call('DBSIZE') > 10000 then return redis.error_reply('Redis 2.6 key browsing is limited to 10000 keys; upgrade for SCAN') end
local keys=redis.call('KEYS',ARGV[1]); local bytes=0
for _,key in ipairs(keys) do bytes=bytes+#key; if bytes>4194304 then return redis.error_reply('key names exceed snapshot limit') end end
return keys`, nil, match).StringSlice()
	return keys, 0, err
}

// scanValue 为 Redis 2.6 的小 Hash/Set 提供分页，返回值格式与 HSCAN/SSCAN 一致。
func (m *Manager) scanValue(ctx context.Context, kind, key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	var values []string
	var next uint64
	var err error
	if kind == "hash" {
		values, next, err = m.client.HScan(ctx, key, cursor, match, count).Result()
	} else {
		values, next, err = m.client.SScan(ctx, key, cursor, match, count).Result()
	}
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "unknown command") {
		return values, next, err
	}
	if match != "" && match != "*" {
		return nil, 0, errors.New("此 Redis 版本不支持 Hash/Set 的模式筛选，请清空筛选条件")
	}
	value, err := m.client.Eval(ctx, `local kind=ARGV[1]; local size=redis.call(kind=='hash' and 'HLEN' or 'SCARD',KEYS[1])
if size>10000 then return redis.error_reply('Redis 2.6 value browsing is limited to 10000 entries; upgrade for SCAN') end
local names=redis.call(kind=='hash' and 'HKEYS' or 'SMEMBERS',KEYS[1]); table.sort(names)
local start=tonumber(ARGV[2]); local finish=math.min(start+tonumber(ARGV[3]),#names); local output={}; local bytes=0
for index=start+1,finish do
 local name=names[index]; local item=kind=='hash' and redis.call('HGET',KEYS[1],name) or nil
 bytes=bytes+#name+(item and #item or 0); if bytes>8388608 then return redis.error_reply('value page exceeds 8 MB') end
 table.insert(output,name); if kind=='hash' then table.insert(output,item or '') end
end
return {finish<#names and finish or 0,output}`, []string{key}, kind, cursor, count).Result()
	if err != nil {
		return nil, 0, err
	}
	parts, ok := value.([]interface{})
	if !ok || len(parts) != 2 {
		return nil, 0, errors.New("Redis 旧版本分页响应不正确")
	}
	position, ok := parts[0].(int64)
	items, itemsOK := parts[1].([]interface{})
	if !ok || !itemsOK || position < 0 {
		return nil, 0, errors.New("Redis 旧版本分页响应不正确")
	}
	values = make([]string, len(items))
	for index, item := range items {
		text, ok := item.(string)
		if !ok {
			return nil, 0, errors.New("Redis 旧版本分页响应不正确")
		}
		values[index] = text
	}
	return values, uint64(position), nil
}
