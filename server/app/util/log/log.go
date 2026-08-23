package log

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	readChunkSize = 64 << 10  // 单次文件读取缓冲大小
	maxReadSize   = 8 << 20   // 单次请求最多读取的日志字节数
	maxLineSize   = 256 << 10 // 单行最多保留的日志字节数
	longLineText  = "[日志行过长，已省略]"
)

// Read 按字节游标读取日志。首次读取最新内容，after 读取新增内容，before 读取更早内容。
// 路径安全和并发锁由调用方负责；文件不存在时返回空结果且不会创建文件。
func Read(path string, after int64, before int64, pageSize int) map[string]interface{} {
	if pageSize < 1 {
		pageSize = 300
	}
	if pageSize > 10000 {
		pageSize = 10000
	}
	fileName := ""
	if path != "" {
		fileName = filepath.Base(path)
	}
	result := map[string]interface{}{
		"file_name":    fileName,
		"lines":        []string{},
		"cursor":       after,
		"start_cursor": after,
		"has_previous": false,
		"end":          true,
	}
	file, err := os.Open(path)
	if err != nil {
		return result
	}
	defer func() { _ = file.Close() }()

	stat, err := file.Stat()
	if err != nil {
		return result
	}
	if stat.Size() == 0 {
		result["cursor"] = int64(0)
		result["start_cursor"] = int64(0)
		return result
	}
	result["end"] = false

	readBefore := func(end int64) ([]string, []int64) {
		if end < 0 {
			end = 0
		}
		if end > stat.Size() {
			end = stat.Size()
		}
		position := end
		lineCount := 0
		chunks := make([][]byte, 0, 2)
		totalSize := 0
		for position > 0 && lineCount <= pageSize && totalSize < maxReadSize {
			readSize := int64(readChunkSize)
			if position < readSize {
				readSize = position
			}
			if remaining := int64(maxReadSize - totalSize); readSize > remaining {
				readSize = remaining
			}
			position -= readSize
			chunk := make([]byte, int(readSize))
			readCount, readErr := file.ReadAt(chunk, position)
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				return []string{}, []int64{}
			}
			if readCount == 0 {
				break
			}
			chunk = chunk[:readCount]
			chunks = append(chunks, chunk)
			totalSize += len(chunk)
			lineCount += bytes.Count(chunk, []byte{'\n'})
		}
		data := make([]byte, 0, totalSize)
		for index := len(chunks) - 1; index >= 0; index-- {
			data = append(data, chunks[index]...)
		}
		lines := make([]string, 0, pageSize)
		starts := make([]int64, 0, pageSize)
		appendLine := func(start, lineEnd int, complete bool) {
			line := data[start:lineEnd]
			if !complete || len(line) > maxLineSize {
				lines = append(lines, longLineText)
			} else {
				content := strings.TrimSuffix(strings.ToValidUTF8(string(line), "�"), "\r")
				lines = append(lines, content)
			}
			starts = append(starts, position+int64(start))
		}
		start := 0
		leftComplete := position == 0
		for start < len(data) {
			relativeEnd := bytes.IndexByte(data[start:], '\n')
			if relativeEnd < 0 {
				appendLine(start, len(data), leftComplete && end == stat.Size())
				break
			}
			lineEnd := start + relativeEnd
			if lineEnd > start || leftComplete {
				appendLine(start, lineEnd, leftComplete)
			}
			start = lineEnd + 1
			leftComplete = true
		}
		if len(lines) > pageSize {
			first := len(lines) - pageSize
			lines = lines[first:]
			starts = starts[first:]
		}
		return lines, starts
	}

	readAfter := func(start int64) ([]string, int64) {
		if start < 0 {
			start = 0
		}
		if start >= stat.Size() {
			return []string{}, start
		}
		lines := make([]string, 0, pageSize)
		next := start
		position := start
		totalSize := 0
		line := make([]byte, 0, min(maxLineSize, readChunkSize))
		lineTooLong := false
		if start > 0 {
			var previous [1]byte
			if count, _ := file.ReadAt(previous[:], start-1); count == 1 {
				totalSize++
				lineTooLong = previous[0] != '\n'
			}
		}
		buffer := make([]byte, readChunkSize)
		for position < stat.Size() && len(lines) < pageSize && totalSize < maxReadSize {
			readSize := int64(len(buffer))
			if remaining := stat.Size() - position; readSize > remaining {
				readSize = remaining
			}
			if remaining := int64(maxReadSize - totalSize); readSize > remaining {
				readSize = remaining
			}
			readCount, readErr := file.ReadAt(buffer[:readSize], position)
			if readCount == 0 {
				break
			}
			for index, value := range buffer[:readCount] {
				if value == '\n' {
					if lineTooLong {
						lines = append(lines, longLineText)
					} else {
						content := strings.TrimSuffix(strings.ToValidUTF8(string(line), "�"), "\r")
						lines = append(lines, content)
					}
					next = position + int64(index) + 1
					line = line[:0]
					lineTooLong = false
					if len(lines) >= pageSize {
						return lines, next
					}
					continue
				}
				if !lineTooLong {
					if len(line) < maxLineSize {
						line = append(line, value)
					} else {
						line = line[:0]
						lineTooLong = true
					}
				}
			}
			position += int64(readCount)
			totalSize += readCount
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				break
			}
		}
		if lineTooLong {
			lines = append(lines, longLineText)
			next = position
		}
		return lines, next
	}

	if after > 0 {
		if after > stat.Size() {
			lines, starts := readBefore(stat.Size())
			result["lines"] = lines
			result["cursor"] = stat.Size()
			result["end"] = true
			if len(starts) > 0 {
				result["start_cursor"] = starts[0]
				result["has_previous"] = starts[0] > 0
			}
			return result
		}
		lines, next := readAfter(after)
		result["lines"] = lines
		result["cursor"] = next
		result["start_cursor"] = after
		result["has_previous"] = after > 0
		result["end"] = next >= stat.Size()
		return result
	}

	end := stat.Size()
	if before > 0 {
		end = before
	}
	lines, starts := readBefore(end)
	result["lines"] = lines
	result["cursor"] = stat.Size()
	if len(starts) > 0 {
		result["start_cursor"] = starts[0]
		result["has_previous"] = starts[0] > 0
		result["end"] = starts[0] <= 0
	} else {
		result["start_cursor"] = int64(0)
		result["end"] = true
	}
	return result
}

// Clear 截断日志；文件不存在时创建空文件。路径安全和并发锁由调用方负责。
func Clear(path string) error {
	if path == "" {
		return errors.New("日志路径不能为空")
	}
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("日志文件不能是符号链接")
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	if err = file.Chmod(0600); err != nil {
		_ = file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}

	// 滚动日志格式为 name-时间-原因.log，并可能追加 gzip 或 zstd 后缀。
	directory := filepath.Dir(path)
	base := filepath.Base(path)
	extension := filepath.Ext(base)
	prefix := strings.TrimSuffix(base, extension) + "-"
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	var clearErr error
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == base {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".gz") {
			name = strings.TrimSuffix(name, ".gz")
		} else if strings.HasSuffix(name, ".zst") {
			name = strings.TrimSuffix(name, ".zst")
		}
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, extension) {
			continue
		}
		value := strings.TrimSuffix(strings.TrimPrefix(name, prefix), extension)
		separator := strings.LastIndexByte(value, '-')
		if separator < 1 || separator == len(value)-1 {
			continue
		}
		if _, parseErr := time.Parse("2006-01-02T15-04-05.000", value[:separator]); parseErr != nil {
			continue
		}
		reason := value[separator+1:]
		if reason != "size" && reason != "time" {
			continue
		}
		if removeErr := os.Remove(filepath.Join(directory, entry.Name())); removeErr != nil && !os.IsNotExist(removeErr) {
			clearErr = errors.Join(clearErr, removeErr)
		}
	}
	return clearErr
}
