package redis

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	client "github.com/redis/go-redis/v9"
)

// Import 流式执行 Redis RESP/AOF 命令，兼容 redis-cli --pipe 使用的二进制命令格式。
func (m *Manager) Import(ctx context.Context, path string, progress ...func(float64, string)) (result map[string]interface{}, err error) {
	result = map[string]interface{}{
		"path":    path,
		"size":    int64(0),
		"total":   int64(0),
		"success": int64(0),
		"failed":  int64(0),
		"unit":    "命令",
	}
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
	if !filepath.IsAbs(path) {
		return result, errors.New("请选择服务器上的 RESP/AOF 文件")
	}
	file, err := os.Open(path)
	if err != nil {
		return result, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return result, err
	}
	if !info.Mode().IsRegular() {
		return result, errors.New("导入路径必须是普通文件")
	}
	result["size"] = info.Size()
	report(0, "正在读取导入文件")
	reader := bufio.NewReaderSize(file, 64<<10)
	if prefix, _ := reader.Peek(5); string(prefix) == "REDIS" {
		return result, errors.New("此文件包含 RDB 数据，请使用纯 RESP/AOF 命令文件")
	}
	var offset int64
	readLength := func(prefix byte) (int64, error) {
		line, e := reader.ReadSlice('\n')
		offset += int64(len(line))
		if e != nil {
			return 0, e
		}
		if len(line) < 4 || len(line) > 32 || line[0] != prefix || line[len(line)-2] != '\r' {
			return 0, errors.New("RESP 长度标记不合法")
		}
		return strconv.ParseInt(string(line[1:len(line)-2]), 10, 64)
	}
	// 独立连接保留 AOF 内 SELECT/MULTI 状态，不把状态带回 Manager 的其他操作。
	settings := *m.client.Options()
	connection := client.NewClient(&settings)
	defer connection.Close()
	commandConnection := connection.Conn()
	defer commandConnection.Close()
	var total, success, queued, transactionBytes int64
	reportRead := func() {
		if len(progress) > 0 && info.Size() > 0 {
			report(99*float64(offset)/float64(info.Size()), "已读取 %d 条命令，已确认成功 %d 条", total, success)
		}
	}
	transaction := false
	for {
		if err = ctx.Err(); err != nil {
			return result, err
		}
		prefix, e := reader.Peek(1)
		if e == io.EOF {
			if transaction {
				return result, errors.New("AOF 事务缺少 EXEC，未提交的命令未执行")
			}
			report(99, "已导入 %d 条命令", success)
			return result, nil
		}
		if e != nil {
			return result, e
		}
		// Redis 7 的可选 AOF 时间戳注释不属于命令。
		if prefix[0] == '#' {
			line, e := reader.ReadSlice('\n')
			offset += int64(len(line))
			if e != nil || len(line) > 128 || !strings.HasPrefix(string(line), "#TS:") {
				return result, errors.New("AOF 时间戳注释不合法")
			}
			reportRead()
			continue
		}
		count, e := readLength('*')
		if e != nil || count < 1 || count > 65536 {
			return result, fmt.Errorf("第 %d 条 RESP 命令参数数量不合法", total+1)
		}
		arguments := make([]interface{}, int(count))
		bytes := count * 16
		for index := range arguments {
			if err = ctx.Err(); err != nil {
				return result, err
			}
			length, e := readLength('$')
			// 仅限制单条命令；整个文件可以任意大小，不把所有命令积攒在内存。
			if e != nil || length < 0 || length > (64<<20)-bytes || length > info.Size() || (index == 0 && length > 256) {
				return result, fmt.Errorf("第 %d 条 RESP 命令长度不合法或超过 64 MB", total+1)
			}
			bytes += length
			value := make([]byte, int(length)+2)
			read, e := io.ReadFull(reader, value)
			offset += int64(read)
			if e != nil || value[length] != '\r' || value[length+1] != '\n' {
				return result, fmt.Errorf("第 %d 条 RESP 命令内容不完整", total+1)
			}
			if index == 0 {
				arguments[index] = string(value[:length])
			} else {
				arguments[index] = value[:length]
			}
		}
		name := strings.ToUpper(arguments[0].(string))
		if transaction && name != "EXEC" && name != "DISCARD" && (queued >= 65536 || bytes > (64<<20)-transactionBytes) {
			return result, errors.New("AOF 单个事务超过 64 MB 或 65536 条命令，未提交的命令未执行")
		}
		command := client.NewCmd(ctx, arguments...)
		total++
		result["total"] = total
		if e = commandConnection.Process(ctx, command); e != nil && e != client.Nil {
			result["failed"] = int64(1)
			var response client.Error
			if !errors.As(e, &response) {
				result["unknown"] = true
				return result, fmt.Errorf("第 %d 条命令结果未确认，已确认成功 %d 条，请核对数据后再操作: %w", total, success, e)
			}
			return result, fmt.Errorf("第 %d 条命令失败，已成功 %d 条: %w", total, success, e)
		}
		if e == client.Nil && name == "EXEC" && transaction {
			result["skipped"] = queued
			return result, errors.New("AOF 事务已中止，未提交的命令未执行")
		}
		if reply, ok := command.Val().(string); transaction && ok && reply == "QUEUED" {
			queued++
			transactionBytes += bytes
			reportRead()
			continue
		}
		if name == "MULTI" {
			transaction = true
		} else if name == "DISCARD" {
			result["skipped"] = queued
			queued, transactionBytes, transaction = 0, 0, false
		} else if name == "EXEC" && transaction {
			values, ok := command.Val().([]interface{})
			if !ok || int64(len(values)) != queued {
				result["unknown"] = true
				return result, errors.New("AOF 事务执行结果不完整，请核对目标数据")
			}
			var failed int64
			for _, value := range values {
				if _, ok := value.(error); ok {
					failed++
				} else {
					success++
				}
			}
			result["success"] = success
			if failed > 0 {
				result["failed"] = failed
				return result, fmt.Errorf("AOF 事务中 %d 条命令失败，已成功 %d 条", failed, success)
			}
			queued, transactionBytes, transaction = 0, 0, false
		}
		success++
		result["success"] = success
		reportRead()
	}
}
