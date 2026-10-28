package recordstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

func seriesFields(fields []Field) ([]Field, error) {
	if len(fields) == 0 || len(fields) > 128 {
		return nil, errors.New("数值历史字段数量必须为 1 至 128 个")
	}
	result := slices.Clone(fields)
	names := make(map[string]bool, len(fields))
	for index := range result {
		field := &result[index]
		if field.Name == "" || len(field.Name) > 128 || strings.TrimSpace(field.Name) != field.Name || names[field.Name] {
			return nil, errors.New("数值历史字段名称必须非空、唯一且不能带首尾空格")
		}
		names[field.Name] = true
		if field.Aggregate == "" {
			field.Aggregate = Mean
		}
		switch field.Aggregate {
		case Mean, Max, Min, Sum, Last:
		default:
			return nil, fmt.Errorf("不支持字段 %s 的聚合方式 %s", field.Name, field.Aggregate)
		}
	}
	return result, nil
}

// validateFiles 校验日文件内保存的字段摘要，不读取外部字段定义。
func (s *Series) validateFiles() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".series" {
			continue
		}
		if _, valid := seriesDay(entry.Name()); !valid {
			return fmt.Errorf("数值历史日文件名称无效: %s", entry.Name())
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("数值历史日文件不是普通文件: %s", entry.Name())
		}
		file, err := os.Open(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			return err
		}
		err = s.checkHeader(file)
		closeErr := file.Close()
		if err != nil {
			return fmt.Errorf("验证数值历史文件 %s 失败: %w", entry.Name(), err)
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func seriesDay(name string) (time.Time, bool) {
	extension := filepath.Ext(name)
	if extension != ".series" {
		return time.Time{}, false
	}
	day, err := time.Parse("2006-01-02", strings.TrimSuffix(name, extension))
	return day, err == nil && day.Format("2006-01-02")+extension == name
}

func seriesSchemaHash(fields []Field) uint32 {
	data, _ := json.Marshal(fields)
	return crc32.ChecksumIEEE(data)
}
