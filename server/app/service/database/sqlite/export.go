package sqlite

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Export 输出 sqlite3 .read 可恢复的 SQL，逐行写入并包含表、索引、视图、触发器和自增序列。
func (m *Manager) Export(ctx context.Context, path, schemaName, tableName string, progress ...func(float64, string)) (result map[string]interface{}, err error) {
	var count int64
	defer func() {
		if err != nil {
			result = map[string]interface{}{
				"path":    path,
				"success": count,
				"total":   count,
				"failed":  1,
				"unit":    "记录",
				"error":   err.Error(),
			}
		}
	}()
	if !filepath.IsAbs(path) || (schemaName != "" && schemaName != "main") {
		return nil, errors.New("请选择服务器 SQL 保存路径，Sqlite 仅支持 main 数据库")
	}
	var databaseFile string
	if err := m.db.QueryRowContext(ctx, "SELECT file FROM pragma_database_list WHERE name='main'").Scan(&databaseFile); err != nil {
		return nil, err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	targetPath := filepath.Join(parent, filepath.Base(path))
	databaseFile, err = filepath.EvalSymlinks(databaseFile)
	if err != nil {
		return nil, err
	}
	// 导出的 SQL 不能替换正在连接的数据库及其事务文件。
	for _, source := range []string{databaseFile, databaseFile + "-wal", databaseFile + "-shm", databaseFile + "-journal"} {
		if targetPath == source {
			return nil, errors.New("SQL 保存路径不能是当前数据库或其事务文件")
		}
		if target, err := os.Stat(path); err == nil {
			if current, err := os.Stat(source); err == nil && os.SameFile(target, current) {
				return nil, errors.New("SQL 保存路径不能是当前数据库或其事务文件")
			}
		}
	}
	if info, err := os.Lstat(path); err == nil {
		if !info.Mode().IsRegular() {
			return nil, errors.New("不能覆盖目录或特殊文件")
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".sqlite-export-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	writer := bufio.NewWriterSize(file, 64<<10)
	tx, err := m.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, "SELECT type, name, tbl_name, sql FROM main.sqlite_master WHERE sql IS NOT NULL AND name NOT GLOB 'sqlite_*' AND (? = '' OR name = ? OR tbl_name = ?) ORDER BY CASE type WHEN 'table' THEN 0 WHEN 'view' THEN 1 WHEN 'index' THEN 2 ELSE 3 END, name", tableName, tableName, tableName)
	if err != nil {
		return nil, err
	}
	objects := make([]map[string]string, 0)
	for rows.Next() {
		var kind, name, table, statement string
		if err = rows.Scan(&kind, &name, &table, &statement); err != nil {
			break
		}
		if len(objects) >= 10000 {
			err = errors.New("数据库对象超过 10000 个，请按表导出")
			break
		}
		if strings.HasPrefix(strings.ToUpper(statement), "CREATE VIRTUAL TABLE") {
			err = errors.New("虚拟表依赖扩展模块，请使用 SQLite 原生备份工具")
			break
		}
		if len(statement) >= 32<<20 {
			err = errors.New("单个对象的 SQL 超过 32 MB，请使用 SQLite 原生备份工具")
			break
		}
		objects = append(objects, map[string]string{
			"type": kind,
			"name": name,
			"sql":  statement,
		})
	}
	if err == nil {
		err = rows.Err()
	}
	_ = rows.Close()
	if err != nil {
		return nil, err
	}
	if tableName != "" && len(objects) == 0 {
		return nil, errors.New("数据表或视图不存在")
	}
	statementSize := -1
	progressEnabled := len(progress) > 0 && progress[0] != nil
	lastUpdate, lastPercent := time.Time{}, float64(-1)
	report := func(percent float64, message string) {
		if !progressEnabled || ctx.Err() != nil {
			return
		}
		percent = max(lastPercent, min(99, max(0, percent)))
		now := time.Now()
		if now.Sub(lastUpdate) < 250*time.Millisecond && percent-lastPercent < 1 {
			return
		}
		lastUpdate, lastPercent = now, percent
		progress[0](percent, message)
	}
	report(1, "正在读取数据库结构")
	totalWork, doneWork := int64(len(objects)), int64(0)
	if progressEnabled {
		for _, object := range objects {
			if object["type"] != "table" {
				continue
			}
			quoted, err := m.quote(object["name"])
			if err != nil {
				return nil, err
			}
			var rows int64
			if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+quoted).Scan(&rows); err != nil {
				return nil, err
			}
			totalWork += rows
			report(2, "正在统计记录: "+object["name"])
		}
	}
	advance := func(message string) {
		if !progressEnabled {
			return
		}
		doneWork++
		if totalWork > 0 {
			report(5+90*float64(doneWork)/float64(totalWork), message)
		}
	}
	write := func(text string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if statementSize >= 0 {
			statementSize += len(text)
			if statementSize > 32<<20 {
				return errors.New("单条数据生成的 SQL 超过 32 MB，请使用 SQLite 原生备份工具")
			}
		}
		_, err := writer.WriteString(text)
		return err
	}
	if err = write("PRAGMA foreign_keys=OFF;\nBEGIN TRANSACTION;\n"); err != nil {
		return nil, err
	}
	exportRows := func(name string) error {
		message := ""
		if progressEnabled {
			message = "正在导出: " + name
		}
		quoted, err := m.quote(name)
		if err != nil {
			return err
		}
		// table_xinfo 的 hidden=2/3 为生成字段，恢复时由数据库重新计算。
		columns, err := tx.QueryContext(ctx, "PRAGMA main.table_xinfo("+quoted+")")
		if err != nil {
			return err
		}
		fields := make([]string, 0)
		for columns.Next() {
			var cid, notNull, primary, hidden int
			var field, kind string
			var defaultValue interface{}
			if err = columns.Scan(&cid, &field, &kind, &notNull, &defaultValue, &primary, &hidden); err != nil {
				break
			}
			if hidden == 0 {
				field, err = m.quote(field)
				if err != nil {
					break
				}
				fields = append(fields, field)
			}
		}
		if err == nil {
			err = columns.Err()
		}
		_ = columns.Close()
		if err != nil {
			return err
		}
		if len(fields) == 0 {
			return errors.New("数据表没有可导出的字段")
		}
		data, err := tx.QueryContext(ctx, "SELECT "+strings.Join(fields, ",")+" FROM "+quoted)
		if err != nil {
			return err
		}
		defer data.Close()
		values, destinations := make([]interface{}, len(fields)), make([]interface{}, len(fields))
		for i := range values {
			destinations[i] = &values[i]
		}
		for data.Next() {
			if err = data.Scan(destinations...); err != nil {
				return err
			}
			statementSize = 0
			if err = write("INSERT INTO " + quoted + " (" + strings.Join(fields, ",") + ") VALUES ("); err != nil {
				return err
			}
			for i, value := range values {
				if i > 0 {
					if err = write(","); err != nil {
						return err
					}
				}
				var text string
				switch value := value.(type) {
				case nil:
					text = "NULL"
				case int64:
					text = strconv.FormatInt(value, 10)
				case float64:
					if math.IsNaN(value) {
						text = "NULL"
					} else if math.IsInf(value, 1) {
						text = "9.0e999"
					} else if math.IsInf(value, -1) {
						text = "-9.0e999"
					} else {
						text = strconv.FormatFloat(value, 'g', -1, 64)
						if !strings.ContainsAny(text, ".eE") {
							text += ".0"
						}
					}
				case []byte:
					if len(value) > (32<<20)/2 {
						return errors.New("单条数据生成的 SQL 超过 32 MB，请使用 SQLite 原生备份工具")
					}
					text = "X'" + hex.EncodeToString(value) + "'"
				case string:
					if len(value) > 32<<20 {
						return errors.New("单条数据生成的 SQL 超过 32 MB，请使用 SQLite 原生备份工具")
					}
					if strings.ContainsRune(value, 0) {
						if len(value) > (32<<20)/2 {
							return errors.New("单条数据生成的 SQL 超过 32 MB，请使用 SQLite 原生备份工具")
						}
						text = "CAST(X'" + hex.EncodeToString([]byte(value)) + "' AS TEXT)"
					} else {
						text = "'" + strings.ReplaceAll(value, "'", "''") + "'"
					}
				case time.Time:
					text = "'" + value.Format("2006-01-02 15:04:05.999999999-07:00") + "'"
				default:
					return fmt.Errorf("不能导出字段类型 %T", value)
				}
				if err = write(text); err != nil {
					return err
				}
			}
			if err = write(");\n"); err != nil {
				return err
			}
			statementSize = -1
			count++
			advance(message)
		}
		return data.Err()
	}
	hasSequence := false
	for _, object := range objects {
		if object["type"] != "table" {
			continue
		}
		if strings.Contains(strings.ToUpper(object["sql"]), "AUTOINCREMENT") {
			tokens, e := m.schemaTokens(object["sql"])
			if e != nil {
				return nil, e
			}
			for _, token := range tokens {
				hasSequence = hasSequence || token == "AUTOINCREMENT"
			}
		}
		if err = write(strings.TrimRight(object["sql"], "; \r\n") + ";\n"); err != nil {
			return nil, err
		}
		if err = exportRows(object["name"]); err != nil {
			return nil, err
		}
		advance("正在导出: " + object["name"])
	}
	var sequence int
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM main.sqlite_master WHERE name = 'sqlite_sequence'").Scan(&sequence); err != nil {
		return nil, err
	}
	if sequence > 0 && hasSequence {
		if tableName == "" {
			err = write("DELETE FROM sqlite_sequence;\n")
		} else {
			err = write("DELETE FROM sqlite_sequence WHERE name='" + strings.ReplaceAll(tableName, "'", "''") + "';\n")
		}
		if err != nil {
			return nil, err
		}
		sequences, e := tx.QueryContext(ctx, "SELECT name, seq FROM sqlite_sequence WHERE (? = '' OR name = ?)", tableName, tableName)
		if e != nil {
			return nil, e
		}
		for sequences.Next() {
			var name string
			var seq int64
			if err = sequences.Scan(&name, &seq); err != nil {
				break
			}
			if err = write("INSERT INTO sqlite_sequence VALUES ('" + strings.ReplaceAll(name, "'", "''") + "'," + strconv.FormatInt(seq, 10) + ");\n"); err != nil {
				break
			}
		}
		if err == nil {
			err = sequences.Err()
		}
		_ = sequences.Close()
		if err != nil {
			return nil, err
		}
	}
	for _, object := range objects {
		if object["type"] != "table" {
			if err = write(strings.TrimRight(object["sql"], "; \r\n") + ";\n"); err != nil {
				return nil, err
			}
			advance("正在写入数据库对象")
		}
	}
	if err = write("COMMIT;\nPRAGMA foreign_keys=ON;\n"); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	report(96, "正在写入 SQL 文件")
	if err = writer.Flush(); err != nil {
		return nil, err
	}
	if err = file.Sync(); err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if err = file.Close(); err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	report(98, "正在保存 SQL 文件")
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	err = os.Rename(file.Name(), path)
	if err != nil {
		return nil, err
	}
	report(99, "SQL 文件已保存")
	return map[string]interface{}{
		"path":    path,
		"size":    info.Size(),
		"total":   count,
		"success": count,
		"failed":  0,
		"unit":    "记录",
	}, nil
}
