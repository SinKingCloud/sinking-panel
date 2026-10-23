package sqlserver

import (
	"bufio"
	"bytes"
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Import 流式执行 SSMS/sqlcmd SQL，GO 是客户端批次分隔符而非 SQL 语句。
func (m *Manager) Import(ctx context.Context, path string, progress ...func(float64, string)) (map[string]interface{}, error) {
	if !filepath.IsAbs(path) {
		return nil, errors.New("请选择服务器上的 SQL 文件")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("导入路径不是普通文件")
	}
	connection, err := m.db.Conn(ctx)
	if err != nil {
		return nil, m.safeError(err)
	}
	defer connection.Close()
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = connection.ExecContext(cleanup, "IF @@TRANCOUNT > 0 ROLLBACK TRANSACTION")
		_ = connection.Raw(func(interface{}) error { return driver.ErrBadConn })
	}()
	reader := bufio.NewReaderSize(file, 64<<10)
	sourceReader := reader
	if bom, _ := reader.Peek(2); bytes.Equal(bom, []byte{0xff, 0xfe}) || bytes.Equal(bom, []byte{0xfe, 0xff}) {
		reader = bufio.NewReaderSize(transform.NewReader(reader, unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM).NewDecoder()), 64<<10)
	} else if bom, _ := reader.Peek(3); bytes.Equal(bom, []byte{0xef, 0xbb, 0xbf}) {
		_, _ = reader.Discard(3)
	}
	boundary := regexp.MustCompile(`(?i)^\s*GO(?:\s+([0-9]+))?\s*(?:(?:--.*)|(?:/\*.*\*/))?\s*$`)
	result := map[string]interface{}{"path": path, "size": info.Size(), "total": int64(0), "success": int64(0), "failed": 0, "unit": "批次"}
	var batch bytes.Buffer
	var quote byte
	depth := 0
	line, startLine := int64(0), int64(1)
	lastUpdate, lastPercent := time.Time{}, float64(-1)
	report := func() {
		if len(progress) == 0 || progress[0] == nil || ctx.Err() != nil {
			return
		}
		position, err := file.Seek(0, io.SeekCurrent)
		if err != nil {
			return
		}
		position -= int64(sourceReader.Buffered())
		percent := float64(0)
		if info.Size() > 0 {
			percent = float64(position) / float64(info.Size()) * 99
		}
		percent = max(lastPercent, min(99, max(0, percent)))
		now := time.Now()
		if now.Sub(lastUpdate) < 250*time.Millisecond && percent-lastPercent < 1 {
			return
		}
		lastUpdate, lastPercent = now, percent
		progress[0](percent, fmt.Sprintf("正在读取 SQL；已执行 %v 批次", result["success"]))
	}
	execute := func(repeat int) error {
		text := strings.TrimSpace(batch.String())
		if text == "" {
			batch.Reset()
			return nil
		}
		for i := 0; i < repeat; i++ {
			result["total"] = result["total"].(int64) + 1
			if _, err := connection.ExecContext(ctx, text); err != nil {
				return fmt.Errorf("第 %d 行批次执行失败：%w", startLine, m.safeError(err))
			}
			result["success"] = result["success"].(int64) + 1
			report()
		}
		batch.Reset()
		startLine = line + 1
		return nil
	}
	for {
		if err = ctx.Err(); err != nil {
			break
		}
		var current bytes.Buffer
		for {
			part, readErr := reader.ReadSlice('\n')
			current.Write(part)
			report()
			if current.Len()+batch.Len() > 32<<20 {
				err = errors.New("单个 SQL 批次超过 32 MB，请使用 GO 拆分脚本")
				break
			}
			if readErr == bufio.ErrBufferFull {
				continue
			}
			err = readErr
			break
		}
		if err != nil && !errors.Is(err, io.EOF) {
			break
		}
		if current.Len() == 0 && errors.Is(err, io.EOF) {
			err = nil
			break
		}
		line++
		text := current.String()
		if !utf8.ValidString(text) || strings.ContainsRune(text, 0) {
			err = errors.New("SQL 文件编码不正确，请使用 UTF-8 或带 BOM 的 UTF-16")
			break
		}
		if quote == 0 && depth == 0 {
			if match := boundary.FindStringSubmatch(text); match != nil {
				repeat := 1
				if match[1] != "" {
					repeat, err = strconv.Atoi(match[1])
					if err != nil || repeat < 1 || repeat > 10000 {
						err = fmt.Errorf("第 %d 行 GO 次数必须在 1 到 10000 之间", line)
						break
					}
				}
				if err = execute(repeat); err != nil {
					break
				}
				continue
			}
			if strings.HasPrefix(strings.TrimSpace(text), ":") || strings.HasPrefix(strings.TrimSpace(text), "!!") {
				err = fmt.Errorf("第 %d 行为 sqlcmd 环境命令，请先展开为 SQL 脚本", line)
				break
			}
		}
		batch.WriteString(text)
		for i := 0; i < len(text); i++ {
			c := text[i]
			next := byte(0)
			if i+1 < len(text) {
				next = text[i+1]
			}
			if depth > 0 {
				if c == '/' && next == '*' {
					depth++
					i++
				} else if c == '*' && next == '/' {
					depth--
					i++
				}
				continue
			}
			if quote != 0 {
				if c == quote {
					if next == quote {
						i++
					} else {
						quote = 0
					}
				}
				continue
			}
			if c == '-' && next == '-' {
				break
			}
			if c == '/' && next == '*' {
				depth++
				i++
				continue
			}
			if c == '\'' || c == '"' || c == '[' {
				quote = c
				if c == '[' {
					quote = ']'
				}
			}
		}
		if errors.Is(err, io.EOF) {
			err = nil
			break
		}
	}
	if err == nil && (quote != 0 || depth != 0) {
		err = fmt.Errorf("第 %d 行批次包含未闭合的引号或注释", startLine)
	}
	if err == nil {
		err = execute(1)
	}
	if err != nil {
		result["failed"] = 1
		result["error"] = m.safeError(err).Error()
		return result, m.safeError(err)
	}
	report()
	return result, nil
}
