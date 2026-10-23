package mysql

import (
	"bufio"
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Import 在同一会话逐条恢复 SQL，支持 mysqldump 的 DELIMITER、版本注释及二进制字面量。
// SQL 自身控制事务；已执行后失败会返回 error 字段与处理位置，避免重复导入。
func (m *Manager) Import(ctx context.Context, path string, progress ...func(float64, string)) (out map[string]interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	if !filepath.IsAbs(path) {
		return nil, errors.New("请选择服务器上的 SQL 文件")
	}
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !stat.Mode().IsRegular() {
		return nil, errors.New("SQL 路径必须是普通文件")
	}
	conn, err := m.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	// 脚本可能修改 SQL_MODE、数据库或事务，结束后不将会话放回池中。
	defer func() { _ = conn.Raw(func(interface{}) error { return driver.ErrBadConn }); _ = conn.Close() }()
	var mode string
	if err = conn.QueryRowContext(ctx, "SELECT @@SESSION.sql_mode").Scan(&mode); err != nil {
		return nil, err
	}
	noBackslash := strings.Contains(strings.ToUpper(mode), "NO_BACKSLASH_ESCAPES")
	out = map[string]interface{}{"path": filepath.Clean(path), "size": stat.Size(), "total": int64(0), "success": int64(0), "affected": int64(0), "unit": "语句"}
	var total, success, affected int64
	line, startLine := 1, 1
	defer func() {
		out["total"], out["success"], out["affected"] = total, success, affected
		if err != nil {
			out["error"], out["failed"], out["line"] = m.safeError(err).Error(), 1, startLine
		}
	}()
	reader := bufio.NewReaderSize(file, 64<<10)
	lastProgress := time.Time{}
	lastPercent := float64(0)
	// 只统计已消耗的字节，不把 bufio 预读到内存的数据算作已处理。
	report := func(force bool) {
		if len(progress) == 0 || progress[0] == nil || (!force && time.Since(lastProgress) < 250*time.Millisecond) {
			return
		}
		offset, e := file.Seek(0, io.SeekCurrent)
		if e != nil {
			return
		}
		percent := float64(0)
		if stat.Size() > 0 {
			percent = min(99, float64(offset-int64(reader.Buffered()))/float64(stat.Size())*99)
		}
		lastPercent = max(lastPercent, percent)
		lastProgress = time.Now()
		progress[0](lastPercent, "正在导入 SQL")
	}
	report(true)
	if bom, _ := reader.Peek(3); string(bom) == "\xef\xbb\xbf" {
		_, _ = reader.Discard(3)
	}
	var statement strings.Builder
	delimiter := ";"
	quote := byte(0)
	lineComment, blockComment, hasSQL, beginning := false, false, false, true
	execute := func() error {
		if !hasSQL {
			statement.Reset()
			return nil
		}
		sqlText := strings.TrimSpace(statement.String())
		total++
		result, executeErr := conn.ExecContext(ctx, sqlText)
		if executeErr != nil {
			return fmt.Errorf("第 %d 行，第 %d 条 SQL 执行失败：%w", startLine, total, executeErr)
		}
		success++
		report(false)
		if n, countErr := result.RowsAffected(); countErr == nil && n > 0 {
			affected += n
		}
		if strings.Contains(strings.ToUpper(sqlText), "SQL_MODE") {
			if executeErr = conn.QueryRowContext(ctx, "SELECT @@SESSION.sql_mode").Scan(&mode); executeErr != nil {
				return executeErr
			}
			noBackslash = strings.Contains(strings.ToUpper(mode), "NO_BACKSLASH_ESCAPES")
		}
		statement.Reset()
		hasSQL = false
		return nil
	}
	for {
		if statement.Len() > 64<<20 {
			return out, fmt.Errorf("第 %d 行的 SQL 超过 64 MB，请减小备份中的单条 INSERT", startLine)
		}
		if reader.Buffered() == 0 || statement.Len()&65535 == 0 {
			report(false)
			if err = ctx.Err(); err != nil {
				return out, err
			}
		}
		ch, readErr := reader.ReadByte()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return out, readErr
		}
		if beginning && quote == 0 && !lineComment && !blockComment && !hasSQL && (ch == 'D' || ch == 'd') {
			next, _ := reader.Peek(9)
			if len(next) >= 9 && strings.EqualFold(string(next[:8]), "ELIMITER") && (next[8] == ' ' || next[8] == '\t') {
				_, _ = reader.Discard(8)
				var command strings.Builder
				for {
					part, e := reader.ReadByte()
					if e != nil && !errors.Is(e, io.EOF) {
						return out, e
					}
					if e != nil || part == '\n' {
						if part == '\n' && e == nil {
							line++
						}
						break
					}
					if command.Len() > 128 {
						return out, fmt.Errorf("第 %d 行的 DELIMITER 不合法", line)
					}
					command.WriteByte(part)
				}
				delimiter = strings.TrimSpace(command.String())
				if delimiter == "" || len(delimiter) > 64 || strings.ContainsAny(delimiter, " \t\r\n\x00") {
					return out, fmt.Errorf("第 %d 行的 DELIMITER 不合法", line)
				}
				statement.Reset()
				beginning = true
				continue
			}
		}
		if quote == 0 && !lineComment && !blockComment && ch == delimiter[0] {
			next, _ := reader.Peek(len(delimiter) - 1)
			if len(next) == len(delimiter)-1 && string(next) == delimiter[1:] {
				_, _ = reader.Discard(len(delimiter) - 1)
				if err = execute(); err != nil {
					return out, err
				}
				beginning = false
				continue
			}
		}
		statement.WriteByte(ch)
		if ch == '\n' {
			line++
			beginning = true
		} else if ch != ' ' && ch != '\t' && ch != '\r' {
			beginning = false
		}
		if lineComment {
			if ch == '\n' {
				lineComment = false
			}
			continue
		}
		if blockComment {
			if ch == '*' {
				next, _ := reader.Peek(1)
				if len(next) == 1 && next[0] == '/' {
					_, _ = reader.ReadByte()
					statement.WriteByte('/')
					blockComment = false
				}
			}
			continue
		}
		if quote != 0 {
			if ch == '\\' && quote != '`' && !noBackslash {
				escaped, e := reader.ReadByte()
				if e != nil {
					return out, fmt.Errorf("第 %d 行的字符串未结束", startLine)
				}
				statement.WriteByte(escaped)
				if escaped == '\n' {
					line++
					beginning = true
				}
				continue
			}
			if ch == quote {
				next, _ := reader.Peek(1)
				if len(next) == 1 && next[0] == quote {
					_, _ = reader.ReadByte()
					statement.WriteByte(quote)
				} else {
					quote = 0
				}
			}
			continue
		}
		next, _ := reader.Peek(2)
		if ch == '#' {
			lineComment = true
			continue
		}
		if ch == '-' && len(next) > 0 && next[0] == '-' && (len(next) == 1 || next[1] <= ' ') {
			_, _ = reader.ReadByte()
			statement.WriteByte('-')
			lineComment = true
			continue
		}
		if ch == '/' && len(next) > 0 && next[0] == '*' {
			_, _ = reader.ReadByte()
			statement.WriteByte('*')
			blockComment = true
			if len(next) > 1 && (next[1] == '!' || next[1] == 'M') {
				if !hasSQL {
					startLine = line
				}
				hasSQL = true
			}
			continue
		}
		if ch > ' ' {
			if !hasSQL {
				startLine = line
			}
			hasSQL = true
		}
		if ch == '\'' || ch == '"' || ch == '`' {
			quote = ch
		}
	}
	if quote != 0 || blockComment {
		return out, fmt.Errorf("第 %d 行的 SQL 字符串或注释未结束", startLine)
	}
	if err = execute(); err != nil {
		return out, err
	}
	current, statErr := file.Stat()
	if statErr != nil {
		return out, statErr
	}
	if current.Size() != stat.Size() || !current.ModTime().Equal(stat.ModTime()) {
		return out, errors.New("导入期间 SQL 文件发生变化，请核对已处理的数据")
	}
	report(true)
	if err = ctx.Err(); err != nil {
		return out, err
	}
	return out, nil
}
