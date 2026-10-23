package postgres

import (
	"bufio"
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
)

// Import 逐条恢复 pg_dump plain SQL，支持 dollar quote、嵌套注释及 COPY FROM stdin。
// 使用一个连接保留 SET、临时对象及事务；SQL 失败后返回统计与行号，不继续执行。
func (m *Manager) Import(ctx context.Context, path string, progress ...func(float64, string)) (out map[string]interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	if !filepath.IsAbs(path) {
		return nil, errors.New("请选择服务器上的 SQL 文件")
	}
	path = filepath.Clean(path)
	file, err := os.Open(path)
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
	reader := bufio.NewReaderSize(file, 64<<10)
	lastProgress := time.Time{}
	lastPercent := float64(0)
	// 扣除 bufio 预读部分，COPY 与普通 SQL 共用真实文件读取进度。
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
	magic, _ := reader.Peek(5)
	if string(magic) == "PGDMP" {
		return nil, errors.New("此文件是 pg_dump archive，请选择 plain SQL 格式的备份")
	}
	conn, err := m.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Raw(func(interface{}) error { return driver.ErrBadConn }); _ = conn.Close() }()
	var standard string
	if err = conn.QueryRowContext(ctx, "SHOW standard_conforming_strings").Scan(&standard); err != nil {
		return nil, err
	}
	standardStrings := strings.EqualFold(standard, "on")
	var total, success, affected int64
	line, startLine := 1, 1
	out = map[string]interface{}{"path": path, "size": stat.Size(), "total": int64(0), "success": int64(0), "affected": int64(0), "unit": "语句"}
	defer func() {
		out["total"], out["success"], out["affected"] = total, success, affected
		if err != nil {
			out["error"], out["failed"], out["line"] = m.safeError(err).Error(), 1, startLine
		}
	}()
	var statement strings.Builder
	quote := byte(0)
	escapedString := false
	dollar := ""
	blockDepth := 0
	lineComment, hasSQL, beginning := false, false, true
	dollarTag := regexp.MustCompile(`^\$(?:[a-zA-Z_][a-zA-Z0-9_]*)?\$`)
	copyCommand := regexp.MustCompile(`(?is)^COPY\s+.+\s+FROM\s+stdin\b`)
	// 仅剥离开头的注释用于识别 COPY，SQL 原文仍完整交给服务端。
	leadingSQL := func(text string) string {
		for {
			text = strings.TrimSpace(text)
			if strings.HasPrefix(text, "--") {
				at := strings.IndexByte(text, '\n')
				if at < 0 {
					return ""
				}
				text = text[at+1:]
				continue
			}
			if strings.HasPrefix(text, "/*") {
				depth, pos := 1, 2
				for pos < len(text) && depth > 0 {
					if strings.HasPrefix(text[pos:], "/*") {
						depth++
						pos += 2
					} else if strings.HasPrefix(text[pos:], "*/") {
						depth--
						pos += 2
					} else {
						pos++
					}
				}
				text = text[pos:]
				continue
			}
			return text
		}
	}
	execute := func() error {
		if !hasSQL {
			statement.Reset()
			return nil
		}
		text := strings.TrimSpace(statement.String())
		total++
		if copyCommand.MatchString(leadingSQL(text)) {
			// COPY 数据从下一行开始；分号后的尾部只能包含空白。
			for {
				ch, e := reader.ReadByte()
				if e != nil {
					return fmt.Errorf("第 %d 行的 COPY 缺少数据段：%w", startLine, e)
				}
				if ch == '\n' {
					line++
					break
				}
				if ch != ' ' && ch != '\t' && ch != '\r' {
					return fmt.Errorf("第 %d 行的 COPY 数据应从下一行开始", line)
				}
			}
			copyLine, rowBytes := line, 0
			var pending []byte
			done := false
			input := copyInput(func(data []byte) (int, error) {
				if err := ctx.Err(); err != nil {
					return 0, err
				}
				if len(pending) > 0 {
					n := copy(data, pending)
					pending = pending[n:]
					return n, nil
				}
				if done {
					return 0, io.EOF
				}
				if rowBytes == 0 {
					prefix, _ := reader.Peek(4)
					ending := 0
					if len(prefix) >= 3 && string(prefix[:3]) == "\\.\n" {
						ending = 3
					} else if len(prefix) >= 4 && string(prefix[:4]) == "\\.\r\n" {
						ending = 4
					} else if len(prefix) == 2 && string(prefix) == "\\." {
						ending = 2
					}
					if ending > 0 {
						_, _ = reader.Discard(ending)
						done = true
						copyLine++
						return 0, io.EOF
					}
				}
				fragment, readErr := reader.ReadSlice('\n')
				report(false)
				rowBytes += len(fragment)
				if rowBytes > 64<<20 {
					return 0, fmt.Errorf("第 %d 行的 COPY 数据超过 64 MB", copyLine)
				}
				if readErr != nil && !errors.Is(readErr, bufio.ErrBufferFull) {
					return 0, fmt.Errorf("第 %d 行的 COPY 数据未找到结束标记 \\.：%w", copyLine, readErr)
				}
				if len(fragment) > 0 && fragment[len(fragment)-1] == '\n' {
					copyLine++
					rowBytes = 0
				}
				n := copy(data, fragment)
				pending = fragment[n:]
				return n, nil
			})
			var copied int64
			executeErr := conn.Raw(func(raw interface{}) error {
				native, ok := raw.(*stdlib.Conn)
				if !ok {
					return errors.New("Postgres 连接不支持原生 COPY")
				}
				tag, e := native.Conn().PgConn().CopyFrom(ctx, input, text)
				copied = tag.RowsAffected()
				return e
			})
			if executeErr != nil {
				return fmt.Errorf("第 %d 行，第 %d 条 COPY 执行失败：%w", startLine, total, executeErr)
			}
			line = copyLine
			beginning = true
			affected += copied
		} else {
			result, e := conn.ExecContext(ctx, text)
			if e != nil {
				return fmt.Errorf("第 %d 行，第 %d 条 SQL 执行失败：%w", startLine, total, e)
			}
			if n, e := result.RowsAffected(); e == nil && n > 0 {
				affected += n
			}
			if strings.Contains(strings.ToLower(text), "standard_conforming_strings") {
				if e = conn.QueryRowContext(ctx, "SHOW standard_conforming_strings").Scan(&standard); e != nil {
					return e
				}
				standardStrings = strings.EqualFold(standard, "on")
			}
		}
		success++
		report(false)
		statement.Reset()
		hasSQL = false
		return nil
	}
	for {
		if statement.Len() > 64<<20 {
			return out, fmt.Errorf("第 %d 行的 SQL 超过 64 MB，请减小单条 INSERT", startLine)
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
		if beginning && quote == 0 && dollar == "" && blockDepth == 0 && !lineComment && !hasSQL && ch == '\\' {
			var command strings.Builder
			for {
				part, e := reader.ReadByte()
				if e != nil && !errors.Is(e, io.EOF) {
					return out, e
				}
				if e != nil || part == '\n' {
					if e == nil {
						line++
					}
					break
				}
				if command.Len() > 4096 {
					return out, fmt.Errorf("第 %d 行的 psql 指令过长", line)
				}
				command.WriteByte(part)
			}
			values := strings.Fields(command.String())
			if len(values) == 0 {
				return out, fmt.Errorf("第 %d 行的 psql 指令为空", line)
			}
			switch values[0] {
			case "restrict", "unrestrict":
				if len(values) != 2 {
					return out, fmt.Errorf("第 %d 行的 psql 限制指令不合法", line)
				}
			case "connect", "c":
				if len(values) != 2 || strings.Trim(values[1], `"'`) != m.database {
					return out, fmt.Errorf("第 %d 行的备份要求切换数据库，请选择该数据库后导入", line)
				}
			default:
				return out, fmt.Errorf("第 %d 行包含不支持的 psql 指令 \\%s", line, values[0])
			}
			statement.Reset()
			beginning = true
			continue
		}
		if quote == 0 && dollar == "" && blockDepth == 0 && !lineComment && ch == ';' {
			if err = execute(); err != nil {
				return out, err
			}
			continue
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
		if blockDepth > 0 {
			next, _ := reader.Peek(1)
			if len(next) == 1 && ch == '/' && next[0] == '*' {
				_, _ = reader.ReadByte()
				statement.WriteByte('*')
				blockDepth++
			} else if len(next) == 1 && ch == '*' && next[0] == '/' {
				_, _ = reader.ReadByte()
				statement.WriteByte('/')
				blockDepth--
			}
			continue
		}
		if dollar != "" {
			if ch == dollar[0] {
				next, _ := reader.Peek(len(dollar) - 1)
				if len(next) == len(dollar)-1 && string(next) == dollar[1:] {
					_, _ = reader.Discard(len(dollar) - 1)
					statement.WriteString(dollar[1:])
					dollar = ""
				}
			}
			continue
		}
		if quote != 0 {
			if ch == '\\' && quote == '\'' && escapedString {
				next, e := reader.ReadByte()
				if e != nil {
					return out, fmt.Errorf("第 %d 行的字符串未结束", startLine)
				}
				statement.WriteByte(next)
				if next == '\n' {
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
		next, _ := reader.Peek(1)
		if ch == '-' && len(next) == 1 && next[0] == '-' {
			_, _ = reader.ReadByte()
			statement.WriteByte('-')
			lineComment = true
			continue
		}
		if ch == '/' && len(next) == 1 && next[0] == '*' {
			_, _ = reader.ReadByte()
			statement.WriteByte('*')
			blockDepth = 1
			continue
		}
		if ch > ' ' {
			if !hasSQL {
				startLine = line
			}
			hasSQL = true
		}
		if ch == '\'' || ch == '"' {
			quote = ch
			escapedString = !standardStrings
			if ch == '\'' {
				text := statement.String()
				if len(text) >= 2 && (text[len(text)-2] == 'E' || text[len(text)-2] == 'e') && (len(text) == 2 || !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_$", rune(text[len(text)-3]))) {
					escapedString = true
				}
			}
		}
		if ch == '$' {
			text := statement.String()
			if len(text) > 1 && strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_$", rune(text[len(text)-2])) {
				continue
			}
			prefix, _ := reader.Peek(64)
			tag := dollarTag.FindString("$" + string(prefix))
			if tag != "" {
				dollar = tag
				_, _ = reader.Discard(len(tag) - 1)
				statement.WriteString(tag[1:])
			}
		}
	}
	if quote != 0 || dollar != "" || blockDepth > 0 {
		return out, fmt.Errorf("第 %d 行的 SQL 字符串或注释未结束", startLine)
	}
	if err = execute(); err != nil {
		return out, err
	}
	current, e := file.Stat()
	if e != nil {
		return out, e
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
