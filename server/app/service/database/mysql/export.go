package mysql

import (
	"bufio"
	"context"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Export 输出可直接恢复的 SQL，包含表结构、完整数据、视图、存储程序和触发器。
// 空 table 导出整个数据库；文件成功后才发布到目标路径，不留下半份备份。
func (m *Manager) Export(ctx context.Context, path, schema, table string, progress ...func(float64, string)) (out map[string]interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	lastPercent := float64(-1)
	lastProgress := time.Time{}
	report := func(percent float64, message string) {
		if len(progress) == 0 || progress[0] == nil {
			return
		}
		percent = max(lastPercent, min(99, max(0, percent)))
		if int(percent) == int(lastPercent) && time.Since(lastProgress) < 250*time.Millisecond {
			return
		}
		lastPercent, lastProgress = percent, time.Now()
		progress[0](percent, message)
	}
	report(0, "读取数据库结构")
	if !filepath.IsAbs(path) {
		return nil, errors.New("请选择服务器上的 SQL 保存路径")
	}
	path = filepath.Clean(path)
	if schema != "" && schema != m.database {
		return nil, errors.New("只能导出当前登记的数据库")
	}
	if old, e := os.Lstat(path); e == nil {
		if !old.Mode().IsRegular() {
			return nil, errors.New("导出路径必须是普通文件")
		}
	} else if !errors.Is(e, os.ErrNotExist) {
		return nil, e
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".mysql-export-*.sql")
	if err != nil {
		return nil, err
	}
	temporary := file.Name()
	defer func() { _ = file.Close(); _ = os.Remove(temporary) }()
	writer := bufio.NewWriterSize(file, 64<<10)
	var count, statements int64
	write := func(sqlText string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := writer.WriteString(strings.TrimSuffix(strings.TrimSpace(sqlText), ";") + ";\n\n"); err != nil {
			return err
		}
		statements++
		return nil
	}
	objects, err := m.Tables(ctx, schema)
	if err != nil {
		return nil, err
	}
	tables := objects.([]map[string]interface{})
	if table != "" {
		filtered := tables[:0]
		for _, item := range tables {
			if item["name"] == table {
				filtered = append(filtered, item)
			}
		}
		tables = filtered
		if len(tables) == 0 {
			return nil, errors.New("表或视图不存在")
		}
	}
	conn, err := m.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Raw(func(interface{}) error { return driver.ErrBadConn }); _ = conn.Close() }()
	if _, err = conn.ExecContext(ctx, "SET TRANSACTION ISOLATION LEVEL REPEATABLE READ"); err != nil {
		return nil, err
	}
	if _, err = conn.ExecContext(ctx, "START TRANSACTION WITH CONSISTENT SNAPSHOT"); err != nil {
		return nil, err
	}
	rowCounts := map[string]int64{}
	totalWeight := float64(0)
	if len(progress) > 0 && progress[0] != nil {
		// 同一快照统计真实行数；每张表额外计一份结构工作量，空表也能推进。
		for index, item := range tables {
			if item["type"] != "VIEW" {
				name := fmt.Sprint(item["name"])
				_, target, e := m.target(schema, name)
				if e != nil {
					return nil, e
				}
				var rows int64
				if e = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+target).Scan(&rows); e != nil {
					return nil, e
				}
				rowCounts[name] = rows
				totalWeight += float64(rows) + 1
			}
			report(1+9*float64(index+1)/float64(len(tables)), "统计导出数据")
		}
	}
	report(10, "写入数据表与数据")
	var charset, collation, mode string
	if err = conn.QueryRowContext(ctx, "SELECT @@character_set_connection, @@collation_connection, @@SESSION.sql_mode").Scan(&charset, &collation, &mode); err != nil {
		return nil, err
	}
	if _, err = writer.WriteString("-- MySQL SQL backup\n-- 表结构与数据来自当前数据库；导入时执行其中的 SQL。\n\n"); err != nil {
		return nil, err
	}
	// HEX 字面量不受 SQL_MODE、反斜杠或二进制数据影响。
	if err = write("SET @SINKING_OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS"); err != nil {
		return nil, err
	}
	if err = write("SET FOREIGN_KEY_CHECKS=0"); err != nil {
		return nil, err
	}
	if err = write("SET @SINKING_OLD_SQL_MODE=@@SQL_MODE"); err != nil {
		return nil, err
	}
	literal := func(text string) string {
		return "'" + strings.ReplaceAll(strings.ReplaceAll(text, "\\", "\\\\"), "'", "''") + "'"
	}
	if mode != "" {
		mode += ","
	}
	mode += "NO_AUTO_VALUE_ON_ZERO"
	if err = write("SET SQL_MODE=X'" + hex.EncodeToString([]byte(mode)) + "'"); err != nil {
		return nil, err
	}
	if err = write("SET NAMES " + literal(charset) + " COLLATE " + literal(collation)); err != nil {
		return nil, err
	}
	if _, err = conn.ExecContext(ctx, "SET time_zone='+00:00'"); err != nil {
		return nil, err
	}
	if err = write("SET @SINKING_OLD_TIME_ZONE=@@TIME_ZONE"); err != nil {
		return nil, err
	}
	if err = write("SET TIME_ZONE='+00:00'"); err != nil {
		return nil, err
	}
	numberLiteral := regexp.MustCompile(`^[+-]?(?:\d+(?:\.\d*)?|\.\d+)(?:[eE][+-]?\d+)?$`)
	encodeValue := func(data []byte, kind string) (string, error) {
		if len(data) > 32<<20 {
			return "", errors.New("单个字段超过 32 MB，请使用数据库原生备份工具处理超大字段")
		}
		kind = strings.ToUpper(strings.TrimPrefix(kind, "UNSIGNED "))
		switch kind {
		case "DECIMAL", "NUMERIC", "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "BIGINT", "YEAR", "FLOAT", "DOUBLE":
			if !numberLiteral.Match(data) {
				return "", errors.New("数值字段返回不合法的数据")
			}
			return string(data), nil
		case "BIT":
			if len(data) > 8 {
				return "", errors.New("BIT 字段超过 64 位")
			}
			var number uint64
			for _, part := range data {
				number = number<<8 | uint64(part)
			}
			return strconv.FormatUint(number, 10), nil
		}
		value := "X'" + hex.EncodeToString(data) + "'"
		if strings.Contains(kind, "BLOB") || strings.Contains(kind, "BINARY") || kind == "GEOMETRY" || kind == "VECTOR" {
			return value, nil
		}
		// 文本字节先按连接字符集解释，再由服务端转换到目标字段字符集。
		return "CONVERT(" + value + " USING " + charset + ")", nil
	}
	completedWeight := float64(0)
	for _, item := range tables {
		name := fmt.Sprint(item["name"])
		if item["type"] == "VIEW" {
			continue
		}
		_, target, _ := m.target(schema, name)
		definition, definitionErr := m.DDL(ctx, schema, name)
		if definitionErr != nil {
			return nil, definitionErr
		}
		if err = write("DROP TABLE IF EXISTS " + target); err != nil {
			return nil, err
		}
		if err = write(definition); err != nil {
			return nil, err
		}
		report(10+75*(completedWeight+1)/max(totalWeight, 1), "写入数据表结构 "+name)
		fields, fieldErr := m.Columns(ctx, schema, name)
		if fieldErr != nil {
			return nil, fieldErr
		}
		names := make([]string, 0)
		for _, column := range fields.([]map[string]interface{}) {
			if m.truth(column["generated"]) {
				continue
			}
			quoted, e := m.quote(fmt.Sprint(column["name"]))
			if e != nil {
				return nil, e
			}
			names = append(names, quoted)
		}
		if len(names) == 0 {
			return nil, fmt.Errorf("表 %s 没有可导出的普通字段", name)
		}
		rows, queryErr := conn.QueryContext(ctx, "SELECT "+strings.Join(names, ",")+" FROM "+target)
		if queryErr != nil {
			return nil, queryErr
		}
		types, typeErr := rows.ColumnTypes()
		if typeErr != nil {
			_ = rows.Close()
			return nil, typeErr
		}
		values, pointers := make([]interface{}, len(names)), make([]interface{}, len(names))
		for i := range values {
			pointers[i] = &values[i]
		}
		prefix := "INSERT INTO " + target + " (" + strings.Join(names, ",") + ") VALUES ("
		writtenRows := int64(0)
		for rows.Next() {
			if err = ctx.Err(); err != nil {
				_ = rows.Close()
				return nil, err
			}
			if err = rows.Scan(pointers...); err != nil {
				_ = rows.Close()
				return nil, err
			}
			var data strings.Builder
			data.WriteString(prefix)
			for i, value := range values {
				if i > 0 {
					data.WriteByte(',')
				}
				switch v := value.(type) {
				case nil:
					data.WriteString("NULL")
				case int64:
					data.WriteString(strconv.FormatInt(v, 10))
				case uint64:
					data.WriteString(strconv.FormatUint(v, 10))
				case float64:
					if math.IsNaN(v) || math.IsInf(v, 0) {
						_ = rows.Close()
						return nil, fmt.Errorf("表 %s 存在不能写入 SQL 的浮点数", name)
					}
					data.WriteString(strconv.FormatFloat(v, 'g', -1, 64))
				case bool:
					if v {
						data.WriteByte('1')
					} else {
						data.WriteByte('0')
					}
				case time.Time:
					data.WriteString("'")
					if strings.EqualFold(types[i].DatabaseTypeName(), "DATE") {
						data.WriteString(v.Format("2006-01-02"))
					} else {
						data.WriteString(v.Format("2006-01-02 15:04:05.999999"))
					}
					data.WriteString("'")
				case []byte:
					encoded, e := encodeValue(v, types[i].DatabaseTypeName())
					if e != nil {
						_ = rows.Close()
						return nil, e
					}
					data.WriteString(encoded)
				case string:
					encoded, e := encodeValue([]byte(v), types[i].DatabaseTypeName())
					if e != nil {
						_ = rows.Close()
						return nil, e
					}
					data.WriteString(encoded)
				default:
					_ = rows.Close()
					return nil, fmt.Errorf("表 %s 返回不支持的字段类型 %T", name, value)
				}
				if data.Len() > 64<<20 {
					_ = rows.Close()
					return nil, fmt.Errorf("表 %s 的单行 SQL 超过 64 MB", name)
				}
			}
			data.WriteByte(')')
			if err = write(data.String()); err != nil {
				_ = rows.Close()
				return nil, err
			}
			count++
			writtenRows++
			report(10+75*(completedWeight+1+min(float64(writtenRows), float64(rowCounts[name])))/max(totalWeight, 1), "导出数据 "+name)
		}
		err = rows.Err()
		_ = rows.Close()
		if err != nil {
			return nil, err
		}
		completedWeight += float64(rowCounts[name]) + 1
		report(10+75*completedWeight/max(totalWeight, 1), "导出数据 "+name)
	}
	// 建立视图的临时结构，使视图依赖另一个视图时不会因名称排序而失败。
	report(85, "写入视图结构")
	for _, item := range tables {
		if item["type"] != "VIEW" {
			continue
		}
		name := fmt.Sprint(item["name"])
		_, target, _ := m.target(schema, name)
		fields, e := m.Columns(ctx, schema, name)
		if e != nil {
			return nil, e
		}
		columns := []string{}
		for _, column := range fields.([]map[string]interface{}) {
			quoted, e := m.quote(fmt.Sprint(column["name"]))
			if e != nil {
				return nil, e
			}
			columns = append(columns, quoted+" "+fmt.Sprint(column["type"]))
		}
		if err = write("DROP VIEW IF EXISTS " + target); err != nil {
			return nil, err
		}
		if err = write("CREATE TABLE " + target + " (" + strings.Join(columns, ",") + ")"); err != nil {
			return nil, err
		}
	}
	// 原生备份以 binary 结果字符集读取 SHOW CREATE，保留创建字符集的原始字节。
	// 这个专用连接在请求结束时丢弃，不把修改后的会话状态交回连接池。
	if _, err = conn.ExecContext(ctx, "SET character_set_results=binary"); err != nil {
		return nil, err
	}
	databaseCollation := ""
	delimiterWrite := func(kind, name string) error {
		objectSchema, target, targetErr := m.target(schema, name)
		if targetErr != nil {
			return targetErr
		}
		definition := ""
		settings := make(map[string]string)
		statement := "SHOW CREATE " + kind + " " + target
		arguments := []interface{}{}
		if kind == "TRIGGER" {
			version, maria, versionErr := m.serverVersion(ctx)
			if versionErr != nil {
				return versionErr
			}
			// SHOW CREATE TRIGGER 与创建字符集元数据在 MySQL 5.1.21 一起加入。
			if !maria && version < 50121 {
				var definitionErr error
				definition, definitionErr = m.ObjectDefinition(ctx, schema, "trigger", name)
				if definitionErr != nil {
					return definitionErr
				}
				statement = "SELECT SQL_MODE FROM information_schema.TRIGGERS WHERE TRIGGER_SCHEMA = ? AND TRIGGER_NAME = ?"
				arguments = []interface{}{objectSchema, name}
			}
		}
		rows, queryErr := conn.QueryContext(ctx, statement, arguments...)
		if queryErr != nil {
			return queryErr
		}
		columns, columnErr := rows.Columns()
		if columnErr != nil {
			_ = rows.Close()
			return columnErr
		}
		if !rows.Next() {
			rowErr := rows.Err()
			_ = rows.Close()
			if rowErr != nil {
				return rowErr
			}
			return errors.New("对象不存在或无权查看")
		}
		values, pointers := make([]interface{}, len(columns)), make([]interface{}, len(columns))
		for index := range values {
			pointers[index] = &values[index]
		}
		if scanErr := rows.Scan(pointers...); scanErr != nil {
			_ = rows.Close()
			return scanErr
		}
		for index, column := range columns {
			var value string
			switch text := values[index].(type) {
			case string:
				value = text
			case []byte:
				value = string(text)
			default:
				continue
			}
			label := strings.ToLower(column)
			if strings.HasPrefix(label, "create ") || label == "sql original statement" {
				definition = value
			} else {
				settings[label] = value
			}
		}
		if closeErr := rows.Close(); closeErr != nil {
			return closeErr
		}
		if definition == "" {
			return errors.New("当前账户没有查看对象完整定义的权限")
		}
		if len(definition) > 1<<20 {
			return errors.New("对象定义超过 1 MB")
		}
		objectMode, ok := settings["sql_mode"]
		if !ok {
			return errors.New("当前服务器未返回对象创建时的 SQL 模式")
		}
		objectCollation := settings["database collation"]
		if objectCollation != "" && databaseCollation == "" {
			if err := conn.QueryRowContext(ctx, "SELECT DEFAULT_COLLATION_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?", objectSchema).Scan(&databaseCollation); err != nil {
				return err
			}
		}
		if err := write("SET SQL_MODE=X'" + hex.EncodeToString([]byte(objectMode)) + "'"); err != nil {
			return err
		}
		// 旧版本的 SHOW CREATE 没有这些字段，按实际返回的列保存。
		for _, variable := range []string{"character_set_client", "collation_connection"} {
			if value := settings[variable]; value != "" {
				if err := write("SET " + variable + "=X'" + hex.EncodeToString([]byte(value)) + "'"); err != nil {
					return err
				}
			}
		}
		databaseTarget, quoteErr := m.quote(objectSchema)
		if quoteErr != nil {
			return quoteErr
		}
		if objectCollation != "" && objectCollation != databaseCollation {
			if err := write("ALTER DATABASE " + databaseTarget + " COLLATE " + literal(objectCollation)); err != nil {
				return err
			}
		}
		delimiter := "$$"
		for strings.Contains(definition, delimiter) {
			delimiter += "$"
			if len(delimiter) > 64 {
				return errors.New("存储程序中无法找到可用的分隔符")
			}
		}
		if _, err := writer.WriteString("DELIMITER " + delimiter + "\n" + strings.TrimSuffix(strings.TrimSpace(definition), ";") + delimiter + "\nDELIMITER ;\n\n"); err != nil {
			return err
		}
		statements++
		if err := write("SET SQL_MODE=X'" + hex.EncodeToString([]byte(mode)) + "'"); err != nil {
			return err
		}
		if err := write("SET NAMES " + literal(charset) + " COLLATE " + literal(collation)); err != nil {
			return err
		}
		if objectCollation != "" && objectCollation != databaseCollation {
			if err := write("ALTER DATABASE " + databaseTarget + " COLLATE " + literal(databaseCollation)); err != nil {
				return err
			}
		}
		return nil
	}
	if table == "" {
		routines, e := m.Objects(ctx, schema, "routine", "", 0, 0)
		if e != nil {
			return nil, e
		}
		for index, routine := range routines.([]map[string]interface{}) {
			name := fmt.Sprint(routine["name"])
			_, target, _ := m.target(schema, name)
			kind := fmt.Sprint(routine["type"])
			if err = write("DROP " + kind + " IF EXISTS " + target); err != nil {
				return nil, err
			}
			if err = delimiterWrite(kind, name); err != nil {
				return nil, err
			}
			report(87+3*float64(index+1)/float64(len(routines.([]map[string]interface{}))), "写入存储程序")
		}
	}
	report(90, "写入视图定义")
	for index, item := range tables {
		if item["type"] != "VIEW" {
			continue
		}
		name := fmt.Sprint(item["name"])
		_, target, _ := m.target(schema, name)
		definition, e := m.ObjectDefinition(ctx, schema, "view", name)
		if e != nil {
			return nil, e
		}
		if err = write("DROP TABLE IF EXISTS " + target); err != nil {
			return nil, err
		}
		if err = write(definition); err != nil {
			return nil, err
		}
		report(90+3*float64(index+1)/float64(len(tables)), "写入视图定义")
	}
	triggers, e := m.Objects(ctx, schema, "trigger", "", 0, 0)
	if e != nil {
		return nil, e
	}
	report(93, "写入触发器")
	for index, trigger := range triggers.([]map[string]interface{}) {
		if table != "" && trigger["table"] != table {
			continue
		}
		if err = delimiterWrite("TRIGGER", fmt.Sprint(trigger["name"])); err != nil {
			return nil, err
		}
		report(93+3*float64(index+1)/float64(len(triggers.([]map[string]interface{}))), "写入触发器")
	}
	if err = write("SET TIME_ZONE=@SINKING_OLD_TIME_ZONE"); err != nil {
		return nil, err
	}
	if err = write("SET SQL_MODE=@SINKING_OLD_SQL_MODE"); err != nil {
		return nil, err
	}
	if err = write("SET FOREIGN_KEY_CHECKS=@SINKING_OLD_FOREIGN_KEY_CHECKS"); err != nil {
		return nil, err
	}
	if _, err = conn.ExecContext(ctx, "COMMIT"); err != nil {
		return nil, err
	}
	report(97, "保存 SQL 文件")
	if err = writer.Flush(); err != nil {
		return nil, err
	}
	if err = file.Sync(); err != nil {
		return nil, err
	}
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if err = file.Close(); err != nil {
		return nil, err
	}
	report(98, "发布 SQL 文件")
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	err = os.Rename(temporary, path)
	if err != nil {
		return nil, err
	}
	report(99, "SQL 文件已导出")
	return map[string]interface{}{"path": path, "size": stat.Size(), "total": count, "success": count, "statements": statements, "unit": "记录"}, nil
}
