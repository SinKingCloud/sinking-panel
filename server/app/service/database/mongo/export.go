package mongo

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Export 将一个集合逐文档写入标准 BSON 文件，可交给 mongorestore 恢复。
// 此文件仅包含文档，集合选项和索引由独立的结构管理功能维护。
func (m *Manager) Export(ctx context.Context, path, collectionName string, progress ...func(float64, string)) (result map[string]interface{}, err error) {
	result = map[string]interface{}{"path": path, "size": int64(0), "total": int64(0), "success": int64(0), "unit": "文档"}
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

	collection, err := m.collection(collectionName)
	if err != nil {
		return result, err
	}
	var expected int64
	if len(progress) > 0 {
		expected, err = collection.CountDocuments(ctx, bson.D{})
		if err != nil {
			return result, err
		}
	}
	report(5, "共 %d 个文档，开始导出", expected)
	// 每批一条，避免大型 BSON 文档和游标批次同时占用大量内存。
	cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetBatchSize(1))
	if err != nil {
		return result, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		if err = ctx.Err(); err != nil {
			return result, err
		}
		if len(cursor.Current) > 16<<20 {
			return result, errors.New("BSON 文档超过 16 MB")
		}
		if err = cursor.Current.Validate(); err != nil {
			return result, err
		}
		if _, err = writer.Write(cursor.Current); err != nil {
			return result, err
		}
		total++
		result["total"], result["success"] = total, total
		if expected > 0 {
			report(min(95, 5+90*float64(total)/float64(expected)), "已导出 %d / %d 个文档", total, expected)
		}
	}
	if err = cursor.Err(); err != nil {
		return result, err
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
	if err = os.Rename(temporary, path); err != nil {
		return result, err
	}
	result["size"], result["total"], result["success"] = info.Size(), total, total
	report(99, "导出文件已保存")
	return result, nil
}
