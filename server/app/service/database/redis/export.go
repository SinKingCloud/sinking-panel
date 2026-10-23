package redis

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	client "github.com/redis/go-redis/v9"
)

// Export 按键写出标准 RESP 命令流，可用 redis-cli --pipe 恢复；单键内存有界。
// PTTL 保留导出时的剩余毫秒，恢复后重新计时；不依赖 Redis 5 才支持的 ABSTTL。
func (m *Manager) Export(ctx context.Context, path, match string, key *string, overwrite bool, progress ...func(float64, string)) (result map[string]interface{}, err error) {
	result = map[string]interface{}{"path": path, "size": int64(0), "total": int64(0), "success": int64(0), "unit": "键"}
	defer func() { err = m.cleanError(err) }()
	lastProgress, lastUpdate := -1.0, time.Time{}
	report := func(percent float64, message string, values ...interface{}) {
		if len(progress) == 0 {
			return
		}
		percent = min(99, max(lastProgress, percent))
		now := time.Now()
		if lastProgress >= 0 && percent-lastProgress < 1 && now.Sub(lastUpdate) < 250*time.Millisecond && (percent < 99 || lastProgress == 99) {
			return
		}
		lastProgress, lastUpdate = percent, now
		if len(values) > 0 {
			message = fmt.Sprintf(message, values...)
		}
		for _, update := range progress {
			if update != nil {
				update(percent, message)
			}
		}
	}
	if !filepath.IsAbs(path) || filepath.Base(path) == "." {
		return result, errors.New("请填写服务器上的完整导出文件路径")
	}
	if err = ctx.Err(); err != nil {
		return result, err
	}
	report(0, "正在统计导出数据")
	if info, e := os.Lstat(path); e == nil {
		if !overwrite {
			return result, errors.New("同名文件已存在，请修改文件名或选择覆盖")
		}
		if !info.Mode().IsRegular() {
			return result, errors.New("不能覆盖目录或特殊文件")
		}
	} else if !errors.Is(e, os.ErrNotExist) {
		return result, e
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".database-export-*")
	if err != nil {
		return result, err
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	defer file.Close()
	writer := bufio.NewWriterSize(file, 64<<10)
	var total int64

	if len(match) > 65536 || (key != nil && len(*key) > 65536) {
		return result, errors.New("导出键名或筛选条件过长")
	}
	if match == "" {
		match = "*"
	}
	// 分批统计匹配键，不保留全库键名；统计与导出之间的增删以实际处理结果为准。
	var expected, processed int64
	if key != nil {
		expected = 1
	} else if len(progress) > 0 {
		var cursor uint64
		for {
			if err = ctx.Err(); err != nil {
				return result, err
			}
			keys, next, e := m.scanKeys(ctx, cursor, match, 100)
			if e != nil {
				return result, e
			}
			expected += int64(len(keys))
			report(0, "正在统计导出键，已统计 %d 个", expected)
			if next == 0 {
				break
			}
			cursor = next
		}
	}
	report(5, "共 %d 个键，开始导出", expected)
	reportKey := func() {
		processed++
		if expected > 0 {
			report(min(95, 5+90*float64(processed)/float64(expected)), "已处理 %d / %d 个键，导出 %d 个", processed, expected, total)
		}
	}
	writeCommand := func(arguments ...string) error {
		if _, e := fmt.Fprintf(writer, "*%d\r\n", len(arguments)); e != nil {
			return e
		}
		for _, value := range arguments {
			if _, e := fmt.Fprintf(writer, "$%d\r\n", len(value)); e != nil {
				return e
			}
			if _, e := writer.WriteString(value); e != nil {
				return e
			}
			if _, e := writer.WriteString("\r\n"); e != nil {
				return e
			}
		}
		return nil
	}
	if err = writeCommand("SELECT", strconv.Itoa(m.database)); err != nil {
		return result, err
	}
	var nonce [16]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return result, err
	}
	restorePrefix := "__panel_restore_" + hex.EncodeToString(nonce[:]) + "_"
	// 一次只读脚本取得 DUMP/PTTL，避免并发改键导致内容和有效期错配。
	// 给键名和 RESP 头留出空间，生成的命令也满足导入的 64 MB 单条上限。
	script := `local value = redis.call('DUMP', KEYS[1])
if not value then return nil end
if #value > 66060288 then return redis.error_reply('single key exceeds 63 MB export limit') end
local ttl = redis.call('PTTL', KEYS[1])
if ttl == -2 or ttl == 0 then return nil end
return {value, ttl}`
	writeKey := func(name string) error {
		if len(name) > 65536 {
			return errors.New("导出键名超过 64 KB")
		}
		if e := ctx.Err(); e != nil {
			return e
		}
		value, e := m.client.Eval(ctx, script, []string{name}).Result()
		if e == client.Nil || (e == nil && value == nil) {
			return nil
		}
		if e != nil {
			return e
		}
		parts, ok := value.([]interface{})
		if !ok || len(parts) != 2 {
			return errors.New("Redis DUMP/PTTL 响应不正确")
		}
		dump, ok := parts[0].(string)
		ttl, ttlOK := parts[1].(int64)
		if !ok || !ttlOK {
			return errors.New("Redis DUMP/PTTL 响应不正确")
		}
		if ttl < 0 {
			ttl = 0
		}
		// 先校验并恢复到唯一临时键，再替换原键；损坏或不兼容的 DUMP 不会先删原数据。
		// 标准事务命令兼容 Redis 2.6，SCAN 重复键也不需要全库去重表。
		temporary := restorePrefix + strconv.FormatInt(total, 10)
		if e = writeCommand("MULTI"); e != nil {
			return e
		}
		if e = writeCommand("RESTORE", temporary, strconv.FormatInt(ttl, 10), dump); e != nil {
			return e
		}
		if e = writeCommand("RENAME", temporary, name); e != nil {
			return e
		}
		if e = writeCommand("EXEC"); e != nil {
			return e
		}
		total++
		result["total"], result["success"] = total, total
		return nil
	}
	if key != nil {
		if err = writeKey(*key); err != nil {
			return result, err
		}
		reportKey()
	} else {
		var cursor uint64
		for {
			keys, next, e := m.scanKeys(ctx, cursor, match, 100)
			if e != nil {
				return result, e
			}
			for _, name := range keys {
				if e = writeKey(name); e != nil {
					return result, e
				}
				reportKey()
			}
			if next == 0 {
				break
			}
			cursor = next
		}
	}

	report(95, "正在写入导出文件")
	if err = ctx.Err(); err != nil {
		return result, err
	}
	if err = writer.Flush(); err != nil {
		return result, err
	}
	report(96, "正在保存导出文件")
	if err = file.Sync(); err != nil {
		return result, err
	}
	info, err := file.Stat()
	if err != nil {
		return result, err
	}
	if err = file.Close(); err != nil {
		return result, err
	}
	report(98, "正在完成导出文件")
	if err = ctx.Err(); err != nil {
		return result, err
	}
	if overwrite {
		err = os.Rename(temporary, path)
	} else if err = os.Link(temporary, path); err != nil && !errors.Is(err, os.ErrExist) {
		// 同目录硬链接原子发布；不支持硬链接时排他创建，仍不会竞争覆盖。
		err = func() error {
			input, e := os.Open(temporary)
			if e != nil {
				return e
			}
			defer input.Close()
			output, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if e != nil {
				return e
			}
			complete := false
			defer func() {
				_ = output.Close()
				if !complete {
					_ = os.Remove(path)
				}
			}()
			buffer := make([]byte, 64<<10)
			for {
				if e = ctx.Err(); e != nil {
					return e
				}
				read, readErr := input.Read(buffer)
				if read > 0 {
					written, writeErr := output.Write(buffer[:read])
					if writeErr != nil {
						return writeErr
					}
					if written != read {
						return io.ErrShortWrite
					}
				}
				if readErr == io.EOF {
					break
				}
				if readErr != nil {
					return readErr
				}
			}
			if e = output.Sync(); e != nil {
				return e
			}
			if e = output.Close(); e != nil {
				return e
			}
			if e = ctx.Err(); e != nil {
				return e
			}
			complete = true
			return nil
		}()
	}
	if err != nil {
		return result, err
	}
	result["size"], result["total"], result["success"] = info.Size(), total, total
	report(99, "导出文件已保存")
	return result, nil
}
