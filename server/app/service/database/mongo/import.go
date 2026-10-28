package mongo

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	client "go.mongodb.org/mongo-driver/v2/mongo"
)

// Import 逐文档读取 mongodump 的单集合 BSON 文件，不加载整个备份或自动重试写入。
func (m *Manager) Import(ctx context.Context, path, collectionName string, progress ...func(float64, string)) (result map[string]interface{}, err error) {
	result = map[string]interface{}{
		"path":    path,
		"size":    int64(0),
		"total":   int64(0),
		"success": int64(0),
		"failed":  int64(0),
		"unit":    "文档",
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
		return result, errors.New("请选择服务器上的 BSON 文件")
	}
	if collectionName == "" {
		collectionName = strings.TrimSuffix(filepath.Base(path), ".bson")
	}
	if strings.HasPrefix(collectionName, "system.") {
		return result, errors.New("不能向 MongoDB 系统集合导入数据")
	}
	collection, err := m.collection(collectionName)
	if err != nil {
		return result, err
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
	var header [4]byte
	var document bson.Raw
	var total, success, offset int64
	for {
		if err = ctx.Err(); err != nil {
			return result, err
		}
		_, err = io.ReadFull(reader, header[:])
		if err == io.EOF {
			report(99, "已导入 %d 个文档", success)
			return result, nil
		}
		if err != nil {
			return result, fmt.Errorf("第 %d 个 BSON 文档长度不完整: %w", total+1, err)
		}
		length := int64(binary.LittleEndian.Uint32(header[:]))
		// BSON 长度包含自身四字节与终止符；MongoDB 的单文档上限为 16 MB。
		if length < 5 || length > 16<<20 || length > info.Size()-offset {
			return result, fmt.Errorf("第 %d 个 BSON 文档长度不合法或超过 16 MB", total+1)
		}
		if int64(cap(document)) < length {
			document = make(bson.Raw, length)
		} else {
			document = document[:length]
		}
		copy(document, header[:])
		if _, err = io.ReadFull(reader, document[4:]); err != nil {
			return result, fmt.Errorf("第 %d 个 BSON 文档不完整: %w", total+1, err)
		}
		offset += length
		if err = document.Validate(); err != nil {
			return result, fmt.Errorf("第 %d 个 BSON 文档格式不合法: %w", total+1, err)
		}
		if err = ctx.Err(); err != nil {
			return result, err
		}
		total++
		result["total"] = total
		if _, err = collection.InsertOne(ctx, document); err != nil {
			result["failed"] = int64(1)
			var write client.WriteException
			if client.IsNetworkError(err) || client.IsTimeout(err) || (errors.As(err, &write) && write.WriteConcernError != nil) {
				result["unknown"] = true
				return result, fmt.Errorf("第 %d 个文档写入结果未确认，已确认成功 %d 个，请核对数据后再操作: %w", total, success, err)
			}
			return result, fmt.Errorf("导入第 %d 个文档失败，已成功 %d 个: %w", total, success, err)
		}
		success++
		result["success"] = success
		if len(progress) > 0 && info.Size() > 0 {
			report(99*float64(offset)/float64(info.Size()), "已导入 %d 个文档", success)
		}
	}
}
