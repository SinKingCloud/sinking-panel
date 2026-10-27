package sqlite

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
	"strings"
	"time"

	"modernc.org/libc"
	sqlite3 "modernc.org/sqlite/lib"
)

// Import 逐条执行标准 .dump SQL；使用 SQLite 自身判断语句边界，保留完整触发器。
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
		return nil, err
	}
	defer connection.Close()
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = connection.ExecContext(cleanup, "ROLLBACK")
		// 脚本的 PRAGMA、临时对象和事务只属于本次专用连接。
		_ = connection.Raw(func(interface{}) error { return driver.ErrBadConn })
	}()
	tls := libc.NewTLS()
	defer tls.Close()
	complete := func(statement string) (bool, error) {
		text, err := libc.CString(statement)
		if err != nil {
			return false, err
		}
		defer libc.Xfree(tls, text)
		return sqlite3.Xsqlite3_complete(tls, text) != 0, nil
	}
	result := map[string]interface{}{
		"path":    path,
		"size":    info.Size(),
		"total":   int64(0),
		"success": int64(0),
		"failed":  0,
		"unit":    "语句",
	}
	reader := bufio.NewReaderSize(file, 64<<10)
	if bom, _ := reader.Peek(3); bytes.Equal(bom, []byte{0xef, 0xbb, 0xbf}) {
		_, _ = reader.Discard(3)
	}
	var statement bytes.Buffer
	var quote byte
	line, startLine := int64(1), int64(1)
	lineComment, blockComment := false, false
	hasSQL := false
	lastUpdate, lastPercent := time.Time{}, float64(-1)
	report := func(force bool) {
		if len(progress) == 0 || progress[0] == nil || ctx.Err() != nil {
			return
		}
		// 先节流再读取文件位置，避免每条短 SQL 都发起 Seek 系统调用。
		if !force && time.Since(lastUpdate) < 250*time.Millisecond {
			return
		}
		position, err := file.Seek(0, io.SeekCurrent)
		if err != nil {
			return
		}
		position -= int64(reader.Buffered())
		percent := float64(0)
		if info.Size() > 0 {
			percent = float64(position) / float64(info.Size()) * 99
		}
		percent = max(lastPercent, min(99, max(0, percent)))
		now := time.Now()
		lastUpdate, lastPercent = now, percent
		progress[0](percent, fmt.Sprintf("正在读取 SQL；已执行 %v 语句", result["success"]))
	}
	report(true)
	execute := func() error {
		text := strings.TrimSpace(statement.String())
		statement.Reset()
		if text == "" || !hasSQL {
			return nil
		}
		hasSQL = false
		result["total"] = result["total"].(int64) + 1
		if _, err := connection.ExecContext(ctx, text); err != nil {
			return fmt.Errorf("第 %d 行 SQL 执行失败：%w", startLine, err)
		}
		result["success"] = result["success"].(int64) + 1
		report(false)
		startLine = line
		return nil
	}
	for {
		if err = ctx.Err(); err != nil {
			break
		}
		if reader.Buffered() == 0 || statement.Len()&65535 == 0 {
			report(false)
		}
		var c byte
		c, err = reader.ReadByte()
		if errors.Is(err, io.EOF) {
			err = nil
			break
		}
		if err != nil {
			break
		}
		if c == 0 {
			err = errors.New("SQL 文件包含零字节，请使用 UTF-8 的 SQL 脚本")
			break
		}
		statement.WriteByte(c)
		if statement.Len() > 32<<20 {
			err = errors.New("单条 SQL 超过 32 MB，请关闭备份工具的扩展批量 INSERT")
			break
		}
		if c == '\n' {
			line++
		}
		next, _ := reader.Peek(1)
		peek := byte(0)
		if len(next) > 0 {
			peek = next[0]
		}
		if lineComment {
			if c == '\n' {
				lineComment = false
			}
			continue
		}
		if blockComment {
			if c == '*' && peek == '/' {
				_, _ = reader.ReadByte()
				statement.WriteByte('/')
				blockComment = false
			}
			continue
		}
		if quote != 0 {
			if c == quote {
				if peek == quote && quote != ']' {
					_, _ = reader.ReadByte()
					statement.WriteByte(peek)
				} else {
					quote = 0
				}
			}
			continue
		}
		if c == '-' && peek == '-' {
			_, _ = reader.ReadByte()
			statement.WriteByte('-')
			lineComment = true
			continue
		}
		if c == '/' && peek == '*' {
			_, _ = reader.ReadByte()
			statement.WriteByte('*')
			blockComment = true
			continue
		}
		if c == '\'' || c == '"' || c == '`' || c == '[' {
			hasSQL = true
			quote = c
			if c == '[' {
				quote = ']'
			}
			continue
		}
		if c != ';' && !strings.ContainsRune(" \t\r\n", rune(c)) {
			hasSQL = true
		}
		if c == ';' {
			var done bool
			done, err = complete(statement.String())
			if err != nil {
				break
			}
			if done {
				err = execute()
				if err != nil {
					break
				}
			}
		}
	}
	if err == nil && (hasSQL || quote != 0 || blockComment) {
		// 末尾允许注释或没有分号的普通语句，未闭合引号/触发器不得当作成功。
		var done bool
		done, err = complete(statement.String() + "\n;")
		if err == nil && (!done || quote != 0 || blockComment) {
			err = fmt.Errorf("第 %d 行 SQL 未结束", startLine)
		}
		if err == nil {
			err = execute()
		}
	}
	if err != nil {
		result["failed"] = 1
		result["error"] = err.Error()
		return result, err
	}
	report(true)
	return result, nil
}
