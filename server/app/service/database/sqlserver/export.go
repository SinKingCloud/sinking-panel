package sqlserver

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Export 生成 SSMS/sqlcmd 可执行的 SQL；数据逐行写入，外键与触发器在数据后恢复。
func (m *Manager) Export(ctx context.Context, path, schemaName, tableName string, progress ...func(float64, string)) (result map[string]interface{}, err error) {
	result = map[string]interface{}{"path": path, "size": int64(0), "total": int64(0), "success": int64(0), "failed": 0, "unit": "记录"}
	defer func() { err = m.safeError(err) }()
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
	if err = ctx.Err(); err != nil {
		return result, err
	}
	report(0, "正在读取数据库结构")
	if !filepath.IsAbs(path) {
		return result, errors.New("请选择服务器上的 SQL 保存路径")
	}
	if info, e := os.Lstat(path); e == nil {
		if !info.Mode().IsRegular() {
			return result, errors.New("不能覆盖目录或特殊文件")
		}
	} else if !os.IsNotExist(e) {
		return result, e
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".sqlserver-export-*")
	if err != nil {
		return result, err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	writer := bufio.NewWriterSize(file, 64<<10)
	batchSize := -1
	write := func(text string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if batchSize >= 0 {
			batchSize += len(text)
			if batchSize > 32<<20 {
				return errors.New("单条数据生成的 SQL 超过 32 MB，请使用 SQL Server 原生工具备份")
			}
		}
		_, err := writer.WriteString(text)
		return err
	}
	quoteValue := func(value string) string {
		// SQL 脚本不写入 NUL 字节，保留字符串内的原始零字符。
		return "N'" + strings.ReplaceAll(strings.ReplaceAll(value, "'", "''"), "\x00", "'+NCHAR(0)+N'") + "'"
	}
	// 与数据表列表保持一致，避免导出 tempdb 时读取其他会话的临时对象。
	objects, err := m.records(ctx, "SELECT o.object_id AS id, s.name AS [schema], o.name, o.type, o.parent_object_id AS parent, OBJECT_DEFINITION(o.object_id) AS definition, sm.uses_ansi_nulls AS ansi_nulls, sm.uses_quoted_identifier AS quoted_identifier, tr.is_disabled, OBJECT_SCHEMA_NAME(o.parent_object_id) AS parent_schema, OBJECT_NAME(o.parent_object_id) AS parent_name FROM sys.objects o JOIN sys.schemas s ON s.schema_id=o.schema_id LEFT JOIN sys.sql_modules sm ON sm.object_id=o.object_id LEFT JOIN sys.triggers tr ON tr.object_id=o.object_id WHERE o.is_ms_shipped=0 AND (o.name NOT LIKE '#%' OR o.name LIKE '##%') AND o.type IN ('U','V','P','PC','FN','IF','TF','FS','FT','TR','TA') AND (@p1='' OR s.name=@p1) AND (@p2='' OR o.name=@p2 OR o.parent_object_id=OBJECT_ID(QUOTENAME(s.name)+'.'+QUOTENAME(@p2))) ORDER BY s.name,o.name", schemaName, tableName)
	if err != nil {
		return result, err
	}
	if tableName != "" && len(objects) == 0 {
		return result, errors.New("数据表或视图不存在")
	}
	if err = write("-- SQL Server database SQL backup\nSET ANSI_NULLS ON;\nSET QUOTED_IDENTIFIER ON;\nGO\n"); err != nil {
		return result, err
	}
	schemas := map[string]bool{}
	ensureSchema := func(name string) error {
		if schemas[name] || name == "dbo" {
			return nil
		}
		schemas[name] = true
		quoted, e := m.quote(name)
		if e != nil {
			return e
		}
		statement := "CREATE SCHEMA " + quoted
		return write("IF SCHEMA_ID(N'" + strings.ReplaceAll(name, "'", "''") + "') IS NULL EXEC(N'" + strings.ReplaceAll(statement, "'", "''") + "');\nGO\n")
	}
	for _, object := range objects {
		if err = ensureSchema(fmt.Sprint(object["schema"])); err != nil {
			return result, err
		}
	}
	// 用户别名类型与序列必须先于引用它们的字段和默认表达式。
	tableTypes := ""
	if m.version >= 10 {
		var count int
		if err = m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys.table_types WHERE (@p1='' OR SCHEMA_NAME(schema_id)=@p1)", schemaName).Scan(&count); err != nil {
			return result, err
		}
		if count > 0 {
			return result, errors.New("数据库包含用户表类型，请使用 SQL Server 原生工具生成完整 SQL 脚本")
		}
		tableTypes = " AND t.is_table_type=0"
	}
	types, e := m.records(ctx, "SELECT SCHEMA_NAME(t.schema_id) AS [schema],t.name,b.name AS base,t.max_length,t.precision,t.scale,t.is_nullable,t.is_assembly_type FROM sys.types t LEFT JOIN sys.types b ON b.user_type_id=t.system_type_id AND b.user_type_id=b.system_type_id WHERE t.is_user_defined=1"+tableTypes+" AND (@p1='' OR SCHEMA_NAME(t.schema_id)=@p1) ORDER BY t.schema_id,t.name", schemaName)
	if e != nil {
		return result, e
	}
	for _, kind := range types {
		if err = ensureSchema(fmt.Sprint(kind["schema"])); err != nil {
			return result, err
		}
		if m.truth(kind["is_assembly_type"]) {
			return result, errors.New("数据库包含 CLR 类型，请使用 SQL Server 原生备份")
		}
		schema, _ := m.quote(fmt.Sprint(kind["schema"]))
		name, _ := m.quote(fmt.Sprint(kind["name"]))
		base := fmt.Sprint(kind["base"])
		length, _ := strconv.Atoi(fmt.Sprint(kind["max_length"]))
		if base == "nvarchar" || base == "nchar" {
			if length > 0 {
				length /= 2
			}
		}
		switch base {
		case "varchar", "char", "varbinary", "binary", "nvarchar", "nchar":
			size := strconv.Itoa(length)
			if length < 0 {
				size = "MAX"
			}
			base += "(" + size + ")"
		case "decimal", "numeric":
			base += "(" + fmt.Sprint(kind["precision"]) + "," + fmt.Sprint(kind["scale"]) + ")"
		case "time", "datetime2", "datetimeoffset":
			base += "(" + fmt.Sprint(kind["scale"]) + ")"
		}
		nullable := " NOT NULL"
		if m.truth(kind["is_nullable"]) {
			nullable = " NULL"
		}
		if err = write("CREATE TYPE " + schema + "." + name + " FROM " + base + nullable + ";\nGO\n"); err != nil {
			return result, err
		}
	}
	if m.version >= 11 {
		used := "1 AS used"
		if m.version >= 14 {
			used = "CASE WHEN last_used_value IS NULL THEN 0 ELSE 1 END AS used"
		}
		sequences, e := m.records(ctx, "SELECT SCHEMA_NAME(schema_id) AS [schema],name,TYPE_NAME(user_type_id) AS type,precision,CONVERT(VARCHAR(100),current_value) AS current_value,CONVERT(VARCHAR(100),start_value) AS start_value,CONVERT(VARCHAR(100),increment) AS increment,CONVERT(VARCHAR(100),minimum_value) AS minimum_value,CONVERT(VARCHAR(100),maximum_value) AS maximum_value,is_cycling,is_exhausted,"+used+" FROM sys.sequences WHERE (@p1='' OR SCHEMA_NAME(schema_id)=@p1) ORDER BY schema_id,name", schemaName)
		if e != nil {
			return result, e
		}
		for _, sequence := range sequences {
			if err = ensureSchema(fmt.Sprint(sequence["schema"])); err != nil {
				return result, err
			}
			schema, _ := m.quote(fmt.Sprint(sequence["schema"]))
			name, _ := m.quote(fmt.Sprint(sequence["name"]))
			cycle := "NO CYCLE"
			if m.truth(sequence["is_cycling"]) {
				cycle = "CYCLE"
			}
			current, ok := new(big.Int).SetString(fmt.Sprint(sequence["current_value"]), 10)
			if !ok {
				return result, errors.New("无法读取序列当前值")
			}
			increment, ok := new(big.Int).SetString(fmt.Sprint(sequence["increment"]), 10)
			if !ok {
				return result, errors.New("无法读取序列步长")
			}
			if (m.truth(sequence["used"]) || fmt.Sprint(sequence["start_value"]) != fmt.Sprint(sequence["current_value"])) && !m.truth(sequence["is_exhausted"]) {
				current.Add(current, increment)
			}
			if m.truth(sequence["is_cycling"]) {
				minimum, a := new(big.Int).SetString(fmt.Sprint(sequence["minimum_value"]), 10)
				maximum, b := new(big.Int).SetString(fmt.Sprint(sequence["maximum_value"]), 10)
				if !a || !b {
					return result, errors.New("无法读取序列范围")
				}
				if current.Cmp(maximum) > 0 {
					current = minimum
				}
				if current.Cmp(minimum) < 0 {
					current = maximum
				}
			}
			kind := fmt.Sprint(sequence["type"])
			if kind == "decimal" || kind == "numeric" {
				kind += "(" + fmt.Sprint(sequence["precision"]) + ",0)"
			}
			if err = write("CREATE SEQUENCE " + schema + "." + name + " AS " + kind + " START WITH " + current.String() + " INCREMENT BY " + fmt.Sprint(sequence["increment"]) + " MINVALUE " + fmt.Sprint(sequence["minimum_value"]) + " MAXVALUE " + fmt.Sprint(sequence["maximum_value"]) + " " + cycle + ";\nGO\n"); err != nil {
				return result, err
			}
			if m.truth(sequence["is_exhausted"]) {
				if err = write("SELECT NEXT VALUE FOR " + schema + "." + name + ";\nGO\n"); err != nil {
					return result, err
				}
			}
		}
	}
	dependencyQuery := "SELECT referencing_id, referenced_id FROM sys.sql_expression_dependencies WHERE referenced_id IS NOT NULL"
	if m.version < 10 {
		dependencyQuery = "SELECT object_id AS referencing_id, referenced_major_id AS referenced_id FROM sys.sql_dependencies WHERE referenced_major_id > 0"
	}
	dependencies, e := m.records(ctx, dependencyQuery)
	if e != nil {
		return result, e
	}
	report(2, "正在导出结构对象")
	pending := map[string]map[string]interface{}{}
	var tables, triggers int
	for _, object := range objects {
		kind := strings.TrimSpace(fmt.Sprint(object["type"]))
		if kind != "TR" && kind != "TA" {
			pending[fmt.Sprint(object["id"])] = object
		} else {
			triggers++
		}
		if kind == "U" {
			tables++
		}
	}
	objectCount, created := len(pending), 0
	for len(pending) > 0 {
		advanced := false
		for _, object := range objects {
			id := fmt.Sprint(object["id"])
			if pending[id] == nil {
				continue
			}
			ready := true
			for _, dependency := range dependencies {
				if fmt.Sprint(dependency["referencing_id"]) == id && fmt.Sprint(dependency["referenced_id"]) != id && pending[fmt.Sprint(dependency["referenced_id"])] != nil {
					ready = false
					break
				}
			}
			if !ready {
				continue
			}
			var statement string
			if strings.TrimSpace(fmt.Sprint(object["type"])) == "U" {
				statement, err = m.DDL(ctx, fmt.Sprint(object["schema"]), fmt.Sprint(object["name"]), false)
			} else {
				if object["definition"] == nil {
					return result, errors.New("对象定义不可读，请检查权限或加密状态")
				}
				statement = fmt.Sprint(object["definition"])
				ansi, quoted := "OFF", "OFF"
				if m.truth(object["ansi_nulls"]) {
					ansi = "ON"
				}
				if m.truth(object["quoted_identifier"]) {
					quoted = "ON"
				}
				if err = write("SET ANSI_NULLS " + ansi + ";\nSET QUOTED_IDENTIFIER " + quoted + ";\nGO\n"); err != nil {
					return result, err
				}
			}
			if err != nil {
				return result, err
			}
			if len(statement) >= 32<<20 {
				return result, errors.New("单个对象的 SQL 超过 32 MB，请使用 SQL Server 原生工具备份")
			}
			if err = write(statement + "\nGO\n"); err != nil {
				return result, err
			}
			if strings.TrimSpace(fmt.Sprint(object["type"])) == "U" {
				// 历史数据可能违反未信任的 CHECK，装载后再还原原来的状态。
				_, target, e := m.target(fmt.Sprint(object["schema"]), fmt.Sprint(object["name"]))
				if e != nil {
					return result, e
				}
				if err = write("ALTER TABLE " + target + " NOCHECK CONSTRAINT ALL;\nGO\n"); err != nil {
					return result, err
				}
			}
			delete(pending, id)
			advanced = true
			created++
			report(2+8*float64(created)/float64(max(1, objectCount)), "已导出 %d / %d 个结构对象", created, objectCount)
		}
		if !advanced {
			return result, errors.New("对象存在循环依赖，请拆分对象后导出")
		}
	}
	tx, err := m.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	var expected, count int64
	if len(progress) > 0 {
		report(10, "正在统计导出数据")
		counted := 0
		for _, object := range objects {
			if strings.TrimSpace(fmt.Sprint(object["type"])) != "U" {
				continue
			}
			_, target, e := m.target(fmt.Sprint(object["schema"]), fmt.Sprint(object["name"]))
			if e != nil {
				return result, e
			}
			var rows int64
			if err = tx.QueryRowContext(ctx, "SELECT COUNT_BIG(*) FROM "+target).Scan(&rows); err != nil {
				return result, err
			}
			expected += rows
			counted++
			report(10+5*float64(counted)/float64(max(1, tables)), "已统计 %d / %d 张表，共 %d 条记录", counted, tables, expected)
		}
	}
	report(15, "开始导出 %d 条记录", expected)
	for _, object := range objects {
		if strings.TrimSpace(fmt.Sprint(object["type"])) != "U" {
			continue
		}
		schema, name := fmt.Sprint(object["schema"]), fmt.Sprint(object["name"])
		_, target, e := m.target(schema, name)
		if e != nil {
			return result, e
		}
		metadata, e := m.Columns(ctx, schema, name)
		if e != nil {
			return result, e
		}
		fields, binary, numeric, kinds := make([]string, 0), make([]bool, 0), make([]bool, 0), make([]string, 0)
		identity := false
		for _, column := range metadata.([]map[string]interface{}) {
			if m.truth(column["generated"]) && !m.truth(column["auto_increment"]) {
				continue
			}
			field, e := m.quote(fmt.Sprint(column["name"]))
			if e != nil {
				return result, e
			}
			fields = append(fields, field)
			kind := strings.ToLower(fmt.Sprint(column["base_type"]))
			kinds = append(kinds, kind)
			binary = append(binary, kind == "binary" || kind == "varbinary" || kind == "image")
			numeric = append(numeric, kind == "decimal" || kind == "numeric" || kind == "money" || kind == "smallmoney")
			identity = identity || m.truth(column["auto_increment"])
		}
		if len(fields) == 0 {
			return result, errors.New("数据表没有可恢复的字段")
		}
		if identity {
			if err = write("SET IDENTITY_INSERT " + target + " ON;\nGO\n"); err != nil {
				return result, err
			}
		}
		data, e := tx.QueryContext(ctx, "SELECT "+strings.Join(fields, ",")+" FROM "+target)
		if e != nil {
			return result, e
		}
		values, destinations := make([]interface{}, len(fields)), make([]interface{}, len(fields))
		for i := range values {
			destinations[i] = &values[i]
		}
		var tableRows int64
		for data.Next() {
			if err = data.Scan(destinations...); err != nil {
				break
			}
			batchSize = 0
			if err = write("INSERT INTO " + target + " (" + strings.Join(fields, ",") + ") VALUES ("); err != nil {
				break
			}
			for i, value := range values {
				if i > 0 {
					if err = write(","); err != nil {
						break
					}
				}
				var literal string
				switch value := value.(type) {
				case nil:
					literal = "NULL"
				case int64:
					literal = strconv.FormatInt(value, 10)
				case float64:
					literal = strconv.FormatFloat(value, 'g', -1, 64)
				case bool:
					literal = "0"
					if value {
						literal = "1"
					}
				case time.Time:
					format := "2006-01-02T15:04:05.9999999"
					switch kinds[i] {
					case "date":
						format = "2006-01-02"
					case "time":
						format = "15:04:05.9999999"
					case "datetimeoffset":
						format = "2006-01-02T15:04:05.9999999-07:00"
					case "smalldatetime":
						format = "2006-01-02T15:04:05"
					}
					literal = "N'" + value.Format(format) + "'"
				case []byte:
					if len(value) > 32<<20 || (binary[i] && len(value) > (32<<20)/2) {
						err = errors.New("单条数据生成的 SQL 超过 32 MB，请使用 SQL Server 原生工具备份")
						break
					}
					if binary[i] {
						literal = "0x" + hex.EncodeToString(value)
					} else if numeric[i] {
						literal = string(value)
					} else {
						literal = quoteValue(string(value))
					}
				case string:
					if len(value) > 32<<20 {
						err = errors.New("单条数据生成的 SQL 超过 32 MB，请使用 SQL Server 原生工具备份")
						break
					}
					if numeric[i] {
						literal = value
					} else {
						literal = quoteValue(value)
					}
				default:
					err = fmt.Errorf("不支持导出字段类型 %T", value)
				}
				if err != nil {
					break
				}
				if err = write(literal); err != nil {
					break
				}
			}
			if err != nil {
				break
			}
			if err = write(");\nGO\n"); err != nil {
				break
			}
			batchSize = -1
			count++
			tableRows++
			result["total"], result["success"] = count, count
			if expected > 0 {
				report(min(90, 15+75*float64(count)/float64(expected)), "已导出 %d / %d 条记录", count, expected)
			}
		}
		if err == nil {
			err = data.Err()
		}
		_ = data.Close()
		if err != nil {
			return result, err
		}
		if identity {
			if err = write("SET IDENTITY_INSERT " + target + " OFF;\nGO\n"); err != nil {
				return result, err
			}
			var last sql.NullString
			var increment string
			if err = tx.QueryRowContext(ctx, "SELECT CONVERT(VARCHAR(100), last_value), CONVERT(VARCHAR(100), increment_value) FROM sys.identity_columns WHERE object_id=OBJECT_ID(@p1)", target).Scan(&last, &increment); err != nil {
				return result, err
			}
			if last.Valid {
				if tableRows == 0 {
					// 新建空表 RESEED 后首次插入不会加步长，补回 DELETE 清空前的下一值。
					current, a := new(big.Int).SetString(last.String, 10)
					step, b := new(big.Int).SetString(increment, 10)
					if !a || !b {
						return result, errors.New("无法读取自增字段当前值或步长")
					}
					last.String = current.Add(current, step).String()
					if err = write("-- Empty identity table: next value is last_value + increment; verify manually if reseeded before its first INSERT or after TRUNCATE.\n"); err != nil {
						return result, err
					}
				}
				if err = write("DBCC CHECKIDENT (N'" + strings.ReplaceAll(target, "'", "''") + "', RESEED, " + last.String + ");\nGO\n"); err != nil {
					return result, err
				}
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return result, err
	}
	report(90, "正在还原约束状态")
	checked := 0
	for _, object := range objects {
		if strings.TrimSpace(fmt.Sprint(object["type"])) != "U" {
			continue
		}
		_, target, e := m.target(fmt.Sprint(object["schema"]), fmt.Sprint(object["name"]))
		if e != nil {
			return result, e
		}
		checks, e := m.records(ctx, "SELECT name, is_disabled, is_not_trusted FROM sys.check_constraints WHERE parent_object_id = OBJECT_ID(@p1) ORDER BY name", target)
		if e != nil {
			return result, e
		}
		for _, check := range checks {
			name, e := m.quote(fmt.Sprint(check["name"]))
			if e != nil {
				return result, e
			}
			state := "WITH CHECK CHECK"
			if m.truth(check["is_disabled"]) {
				state = "NOCHECK"
			} else if m.truth(check["is_not_trusted"]) {
				state = "CHECK"
			}
			if err = write("ALTER TABLE " + target + " " + state + " CONSTRAINT " + name + ";\nGO\n"); err != nil {
				return result, err
			}
		}
		checked++
		report(90+2*float64(checked)/float64(max(1, tables)), "已导出 %d / %d 张表的约束状态", checked, tables)
	}
	keys, e := m.Objects(ctx, schemaName, "foreign_key")
	if e != nil {
		return result, e
	}
	foreignKeys := keys.([]map[string]interface{})
	for index, key := range foreignKeys {
		if tableName != "" && !strings.EqualFold(fmt.Sprint(key["table"]), tableName) {
			continue
		}
		statement, e := m.ObjectDefinition(ctx, fmt.Sprint(key["schema"]), "foreign_key", fmt.Sprint(key["name"]))
		if e != nil {
			return result, e
		}
		if err = write(statement + "\nGO\n"); err != nil {
			return result, err
		}
		report(92+3*float64(index+1)/float64(max(1, len(foreignKeys))), "已导出 %d / %d 个外键", index+1, len(foreignKeys))
	}
	report(95, "正在导出触发器")
	triggerCount := 0
	for _, object := range objects {
		if kind := strings.TrimSpace(fmt.Sprint(object["type"])); kind == "TR" || kind == "TA" {
			if object["definition"] == nil {
				return result, errors.New("触发器定义不可读")
			}
			if len(fmt.Sprint(object["definition"])) >= 32<<20 {
				return result, errors.New("单个对象的 SQL 超过 32 MB，请使用 SQL Server 原生工具备份")
			}
			ansi, quoted := "OFF", "OFF"
			if m.truth(object["ansi_nulls"]) {
				ansi = "ON"
			}
			if m.truth(object["quoted_identifier"]) {
				quoted = "ON"
			}
			if err = write("SET ANSI_NULLS " + ansi + ";\nSET QUOTED_IDENTIFIER " + quoted + ";\nGO\n"); err != nil {
				return result, err
			}
			if err = write(fmt.Sprint(object["definition"]) + "\nGO\n"); err != nil {
				return result, err
			}
			if m.truth(object["is_disabled"]) {
				_, trigger, e := m.target(fmt.Sprint(object["schema"]), fmt.Sprint(object["name"]))
				if e != nil {
					return result, e
				}
				_, target, e := m.target(fmt.Sprint(object["parent_schema"]), fmt.Sprint(object["parent_name"]))
				if e != nil {
					return result, e
				}
				if err = write("DISABLE TRIGGER " + trigger + " ON " + target + ";\nGO\n"); err != nil {
					return result, err
				}
			}
			triggerCount++
			report(95+2*float64(triggerCount)/float64(max(1, triggers)), "已导出 %d / %d 个触发器", triggerCount, triggers)
		}
	}
	report(97, "正在写入导出文件")
	if err = ctx.Err(); err != nil {
		return result, err
	}
	if err = writer.Flush(); err != nil {
		return result, err
	}
	report(98, "正在保存导出文件")
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
	if err = ctx.Err(); err != nil {
		return result, err
	}
	err = os.Rename(file.Name(), path)
	if err != nil {
		return result, err
	}
	result["size"] = info.Size()
	report(99, "导出文件已保存")
	return result, nil
}
