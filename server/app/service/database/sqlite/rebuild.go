package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// rebuildTable 按 SQLite 的安全迁移顺序重建普通表，只修改指定字段或外键。
// 原字段片段、约束、索引和触发器直接复用；复制与校验失败时整个事务回滚。
func (m *Manager) rebuildTable(ctx context.Context, schemaName, tableName string, column *ColumnDefinition, oldName string, key *ForeignKeyDefinition, removeKey bool) (_ *Result, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	table, err := m.maintenanceTable(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	connection, err := m.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	var foreignKeys, legacy int
	if err = connection.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		return nil, err
	}
	if err = connection.QueryRowContext(ctx, "PRAGMA legacy_alter_table").Scan(&legacy); err != nil {
		return nil, err
	}
	// PRAGMA foreign_keys 在事务内不生效，独占连接并在事务结束后恢复。
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, first := connection.ExecContext(cleanupCtx, "PRAGMA legacy_alter_table = "+strconv.Itoa(legacy))
		_, second := connection.ExecContext(cleanupCtx, "PRAGMA foreign_keys = "+strconv.Itoa(foreignKeys))
		if first != nil || second != nil {
			_ = connection.Raw(func(interface{}) error { return driver.ErrBadConn })
			if err == nil {
				err = errors.New("迁移已完成，但连接设置恢复失败，该连接已丢弃，请重新读取结构")
			}
		}
	}()
	if _, err = connection.ExecContext(ctx, "PRAGMA foreign_keys = OFF"); err != nil {
		return nil, err
	}
	// 临时表改回原名时保留外部视图、触发器和外键中的原名称。
	if _, err = connection.ExecContext(ctx, "PRAGMA legacy_alter_table = ON"); err != nil {
		return nil, err
	}
	tx, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if column != nil && column.Name != oldName {
		old, err := m.quote(oldName)
		if err != nil {
			return nil, err
		}
		name, err := m.quote(column.Name)
		if err != nil {
			return nil, err
		}
		// 字段改名先交给原生实现，同一事务更新所有引用，再读取新定义。
		if _, err = tx.ExecContext(ctx, "ALTER TABLE "+table+" RENAME COLUMN "+old+" TO "+name); err != nil {
			return nil, err
		}
		oldName = column.Name
	}
	var definition string
	if err = tx.QueryRowContext(ctx, "SELECT name,sql FROM main.sqlite_master WHERE type = 'table' AND name = ? COLLATE NOCASE", tableName).Scan(&tableName, &definition); err != nil {
		return nil, err
	}
	// 只在本方法中保留 token 的原文位置，避免重写不相关的 SQL 片段。
	type token struct {
		text   string
		start  int
		end    int
		quoted bool
	}
	number := regexp.MustCompile(`(?i)^(?:0x[0-9a-f](?:_?[0-9a-f])*|(?:[0-9](?:_?[0-9])*(?:\.(?:[0-9](?:_?[0-9])*)?)?|\.[0-9](?:_?[0-9])*)(?:e[+-]?[0-9](?:_?[0-9])*)?)`)
	tokenize := func(text string) ([]token, error) {
		if len(text) > 1<<20 || strings.ContainsRune(text, 0) {
			return nil, errors.New("数据表定义不合法或超过 1 MB")
		}
		result := make([]token, 0)
		for index := 0; index < len(text); {
			c := text[index]
			if unicode.IsSpace(rune(c)) {
				index++
				continue
			}
			if strings.HasPrefix(text[index:], "--") {
				if end := strings.IndexByte(text[index:], '\n'); end >= 0 {
					index += end + 1
				} else {
					break
				}
				continue
			}
			if strings.HasPrefix(text[index:], "/*") {
				end := strings.Index(text[index+2:], "*/")
				if end < 0 {
					return nil, errors.New("数据表定义存在未闭合注释")
				}
				index += end + 4
				continue
			}
			start := index
			if c == '\'' || c == '"' || c == '`' || c == '[' {
				closing := c
				if c == '[' {
					closing = ']'
				}
				index++
				var value strings.Builder
				closed := false
				for index < len(text) {
					c := text[index]
					index++
					if c == closing {
						if index < len(text) && text[index] == closing {
							value.WriteByte(c)
							index++
							continue
						}
						closed = true
						break
					}
					value.WriteByte(c)
				}
				if !closed {
					return nil, errors.New("数据表定义存在未闭合引号")
				}
				result = append(result, token{
					text:   value.String(),
					start:  start,
					end:    index,
					quoted: true,
				})
				continue
			}
			if value := number.FindString(text[index:]); value != "" {
				index += len(value)
			} else if strings.ContainsRune("(),.;+-=*/,", rune(c)) {
				index++
			} else {
				for index < len(text) && !unicode.IsSpace(rune(text[index])) && !strings.ContainsRune("(),.;+-=*/'\"`[]", rune(text[index])) {
					index++
				}
			}
			result = append(result, token{
				text:  strings.ToUpper(text[start:index]),
				start: start,
				end:   index,
			})
		}
		return result, nil
	}
	tokens, err := tokenize(definition)
	if err != nil {
		return nil, err
	}
	if len(tokens) < 5 || tokens[0].text != "CREATE" || tokens[1].text != "TABLE" {
		return nil, errors.New("仅支持普通 SQLite 数据表，虚拟表请使用专用 SQL 修改")
	}
	opening, closing, depth := -1, -1, 0
	clauses := make([]string, 0)
	last := 0
	for _, token := range tokens {
		if token.quoted {
			continue
		}
		if token.text == "(" {
			depth++
			if opening < 0 {
				opening, last = token.start, token.end
			}
		} else if token.text == ")" {
			depth--
			if depth == 0 && opening >= 0 {
				clauses = append(clauses, definition[last:token.start])
				closing = token.start
				break
			}
		} else if token.text == "," && depth == 1 {
			clauses = append(clauses, definition[last:token.start])
			last = token.end
		}
	}
	if opening < 0 || closing < 0 || len(clauses) == 0 {
		return nil, errors.New("无法可靠解析数据表字段定义")
	}
	withoutRowid := false
	suffix, err := tokenize(definition[closing+1:])
	if err != nil {
		return nil, err
	}
	for index, token := range suffix {
		if !token.quoted && token.text == "WITHOUT" && index+1 < len(suffix) && suffix[index+1].text == "ROWID" {
			withoutRowid = true
		}
	}
	// xinfo 包含生成字段，复制时不对其赋值，让原表达式重新计算。
	quotedTable, _ := m.quote(tableName)
	rows, err := tx.QueryContext(ctx, "PRAGMA main.table_xinfo("+quotedTable+")")
	if err != nil {
		return nil, err
	}
	fields := make([]string, 0)
	names, types, defaults, notNull, primary := make(map[string]bool), make(map[string]string), make(map[string]sql.NullString), make(map[string]bool), make(map[string]int)
	primaryCount := 0
	for rows.Next() {
		var cid, nullable, pk, hidden int
		var name, typeName string
		var value sql.NullString
		if err = rows.Scan(&cid, &name, &typeName, &nullable, &value, &pk, &hidden); err != nil {
			rows.Close()
			return nil, err
		}
		lower := strings.ToLower(name)
		names[lower], types[lower], defaults[lower], notNull[lower], primary[lower] = true, typeName, value, nullable != 0, pk
		if pk > 0 {
			primaryCount++
		}
		if hidden == 0 {
			quoted, _ := m.quote(name)
			fields = append(fields, quoted)
		} else if hidden == 1 {
			rows.Close()
			return nil, errors.New("虚拟表不支持通用结构迁移")
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, errors.New("没有可安全复制的字段")
	}
	// INTEGER PRIMARY KEY 的 rowid 别名没有主键自动索引，其机制不能隐式变更。
	rows, err = tx.QueryContext(ctx, "PRAGMA main.index_list("+quotedTable+")")
	if err != nil {
		return nil, err
	}
	primaryIndex := false
	for rows.Next() {
		var seq, unique, partial int
		var name, origin string
		if err = rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			rows.Close()
			return nil, err
		}
		primaryIndex = primaryIndex || origin == "pk"
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	rowidAlias := ""
	if !withoutRowid && !primaryIndex && primaryCount == 1 {
		for name, pk := range primary {
			if pk > 0 && strings.EqualFold(types[name], "INTEGER") {
				rowidAlias = name
			}
		}
	}
	if !withoutRowid && rowidAlias == "" {
		alias := ""
		for _, candidate := range []string{"rowid", "_rowid_", "oid"} {
			if !names[candidate] {
				alias = candidate
				break
			}
		}
		if alias == "" {
			return nil, errors.New("rowid 的三个别名均被字段占用，无法可靠保留内部行号")
		}
		quoted, _ := m.quote(alias)
		fields = append([]string{quoted}, fields...)
	}
	type part struct {
		clause int
		start  int
		end    int
		name   string
	}
	foreignParts := make([]part, 0)
	changedColumn := false
	for clauseIndex, original := range clauses {
		items, err := tokenize(original)
		if err != nil || len(items) == 0 {
			return nil, errors.New("无法可靠解析字段或约束定义")
		}
		isTableConstraint := !items[0].quoted && (items[0].text == "CONSTRAINT" || items[0].text == "PRIMARY" || items[0].text == "UNIQUE" || items[0].text == "CHECK" || items[0].text == "FOREIGN")
		constraintStart := 1
		if !isTableConstraint {
			level := 0
			constraintStart = len(items)
			for index := 1; index < len(items); index++ {
				item := items[index]
				if item.quoted {
					continue
				}
				if item.text == "(" {
					level++
				} else if item.text == ")" {
					level--
				} else if level == 0 {
					switch item.text {
					case "CONSTRAINT", "PRIMARY", "NOT", "NULL", "UNIQUE", "CHECK", "DEFAULT", "COLLATE", "REFERENCES", "GENERATED", "AS":
						constraintStart = index
					}
					if constraintStart != len(items) {
						break
					}
				}
			}
		}
		// 外键序号是 SQLite 从定义末端开始分配的目录 id；同时支持实际命名约束。
		level := 0
		for index, item := range items {
			if item.quoted {
				continue
			}
			if item.text == "(" {
				level++
			} else if item.text == ")" {
				level--
			} else if level == 0 && item.text == "REFERENCES" {
				foreign := part{
					clause: clauseIndex,
					start:  0,
					end:    len(original),
				}
				if isTableConstraint {
					if items[0].text == "CONSTRAINT" && len(items) > 1 {
						foreign.name = items[1].text
					}
				} else {
					foreign.start = item.start
					if index >= 2 && items[index-2].text == "CONSTRAINT" && !items[index-2].quoted {
						foreign.start, foreign.name = items[index-2].start, items[index-1].text
					}
					end := index + 1
					if end >= len(items) {
						return nil, errors.New("外键引用表定义不完整")
					}
					end++
					if end < len(items) && items[end].text == "." {
						return nil, errors.New("Sqlite 外键不能跨数据库引用")
					}
					if end < len(items) && items[end].text == "(" && !items[end].quoted {
						depth := 1
						end++
						for end < len(items) && depth > 0 {
							if !items[end].quoted && items[end].text == "(" {
								depth++
							}
							if !items[end].quoted && items[end].text == ")" {
								depth--
							}
							end++
						}
					}
					for end < len(items) {
						next := items[end]
						if next.quoted {
							break
						}
						if next.text == "ON" && end+2 < len(items) {
							end += 3
							if items[end-1].text == "SET" || items[end-1].text == "NO" {
								end++
							}
						} else if next.text == "MATCH" && end+1 < len(items) {
							end += 2
						} else if next.text == "DEFERRABLE" {
							end++
						} else if next.text == "NOT" && end+1 < len(items) && items[end+1].text == "DEFERRABLE" {
							end += 2
						} else if next.text == "INITIALLY" && end+1 < len(items) {
							end += 2
						} else {
							break
						}
					}
					if end == 0 || end > len(items) {
						return nil, errors.New("无法可靠解析外键动作")
					}
					foreign.end = items[end-1].end
				}
				foreignParts = append(foreignParts, foreign)
			}
		}
		if column == nil || isTableConstraint || !strings.EqualFold(items[0].text, oldName) {
			continue
		}
		changedColumn = true
		lower := strings.ToLower(oldName)
		if rowidAlias == lower && !strings.EqualFold(strings.TrimSpace(column.Type), "INTEGER") {
			return nil, errors.New("不能隐式改变 INTEGER 主键的 rowid 机制，请使用显式迁移")
		}
		if withoutRowid && primary[lower] > 0 && column.Nullable {
			return nil, errors.New("WITHOUT ROWID 表的主键字段必须为非空")
		}
		candidate := *column
		candidate.Name = "__column"
		validated, err := m.columnDefinition(candidate)
		if err != nil {
			return nil, err
		}
		type replacement struct {
			start int
			end   int
			text  string
		}
		edits := make([]replacement, 0)
		if !strings.EqualFold(strings.TrimSpace(types[lower]), strings.TrimSpace(column.Type)) {
			end := len(original)
			if constraintStart < len(items) {
				end = items[constraintStart].start
			}
			edits = append(edits, replacement{
				start: items[0].end,
				end:   end,
				text:  " " + strings.TrimSpace(column.Type) + " ",
			})
		}
		currentDefault := defaults[lower]
		defaultSame := !currentDefault.Valid && column.Default == nil
		if currentDefault.Valid && column.Default != nil {
			value := *column.Default
			if !column.DefaultExpression {
				value = "'" + strings.ReplaceAll(value, "'", "''") + "'"
			}
			defaultSame = strings.TrimSpace(value) == strings.TrimSpace(currentDefault.String)
		}
		currentNullable := !notNull[lower]
		if column.AutoIncrement {
			currentNullable = false
		}
		nullableSame := column.Nullable == currentNullable
		level = 0
		for index := constraintStart; index < len(items); index++ {
			item := items[index]
			if item.quoted {
				continue
			}
			if item.text == "(" {
				level++
				continue
			}
			if item.text == ")" {
				level--
				continue
			}
			if level != 0 {
				continue
			}
			if item.text == "CONSTRAINT" || item.text == "COLLATE" {
				// 约束名和排序规则名可以与关键字同名，不把它们当作属性。
				index++
				continue
			}
			if item.text == "REFERENCES" {
				// ON DELETE SET NULL 属于外键动作，不能当作字段可空属性删除。
				for _, foreign := range foreignParts {
					if foreign.clause == clauseIndex && foreign.start <= item.start && foreign.end >= item.end {
						for index+1 < len(items) && items[index+1].end <= foreign.end {
							index++
						}
						break
					}
				}
				continue
			}
			start := item.start
			if index >= 2 && items[index-2].text == "CONSTRAINT" && !items[index-2].quoted {
				start = items[index-2].start
			}
			if item.text == "DEFAULT" {
				end := index + 1
				if end >= len(items) {
					return nil, errors.New("无法可靠解析字段默认值")
				}
				if !items[end].quoted && items[end].text == "(" {
					depth := 1
					end++
					for end < len(items) && depth > 0 {
						if !items[end].quoted && items[end].text == "(" {
							depth++
						}
						if !items[end].quoted && items[end].text == ")" {
							depth--
						}
						end++
					}
					if depth != 0 {
						return nil, errors.New("无法可靠解析字段默认值")
					}
				} else {
					if !items[end].quoted && (items[end].text == "+" || items[end].text == "-") {
						end++
					}
					if end >= len(items) {
						return nil, errors.New("无法可靠解析字段默认值")
					}
					if !items[end].quoted && items[end].text == "X" && end+1 < len(items) && items[end+1].quoted {
						end++
					}
					end++
				}
				if !defaultSame {
					edits = append(edits, replacement{
						start: start,
						end:   items[end-1].end,
					})
				}
				index = end - 1
			} else if !nullableSame && ((item.text == "NOT" && index+1 < len(items) && items[index+1].text == "NULL") || item.text == "NULL") {
				end := index + 1
				if item.text == "NOT" {
					end++
				}
				if end+2 < len(items) && items[end].text == "ON" && items[end+1].text == "CONFLICT" {
					end += 3
				}
				edits = append(edits, replacement{
					start: start,
					end:   items[end-1].end,
				})
				index = end - 1
			}
		}
		sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
		for _, edit := range edits {
			original = original[:edit.start] + edit.text + original[edit.end:]
		}
		if !nullableSame && !column.Nullable {
			original += "\n NOT NULL"
		}
		if !defaultSame && column.Default != nil {
			position := strings.Index(validated, " DEFAULT ")
			if position < 0 {
				return nil, errors.New("生成字段或自增字段不允许设置默认值")
			}
			original += "\n" + validated[position:]
		}
		clauses[clauseIndex] = original
	}
	if column != nil && !changedColumn {
		return nil, errors.New("字段不存在或无法可靠定位字段定义")
	}
	if column == nil {
		if key != nil {
			for _, field := range key.Columns {
				if !names[strings.ToLower(field)] {
					return nil, errors.New("外键字段不存在")
				}
			}
		}
		position := -1
		foreignOptions := ""
		if oldName != "" {
			for index, part := range foreignParts {
				if part.name != "" && strings.EqualFold(part.name, oldName) {
					if position >= 0 {
						return nil, errors.New("外键名称重复，请使用目录中的唯一标识修改")
					}
					position = index
				}
			}
			if position < 0 {
				if separator := strings.LastIndexByte(oldName, '.'); separator >= 0 && strings.EqualFold(oldName[:separator], tableName) {
					id, parseErr := strconv.Atoi(oldName[separator+1:])
					if parseErr == nil && id >= 0 && id < len(foreignParts) {
						position = len(foreignParts) - id - 1
					}
				}
			}
			if position < 0 {
				return nil, errors.New("当前表中不存在该外键")
			}
			part := foreignParts[position]
			original := clauses[part.clause]
			if key != nil && !removeKey {
				// 编辑界面仅修改引用和动作，保留原外键的 MATCH 与延迟检查属性。
				items, tokenErr := tokenize(original[part.start:part.end])
				if tokenErr != nil {
					return nil, tokenErr
				}
				level, reference := 0, false
				for index := 0; index < len(items); index++ {
					item := items[index]
					if item.quoted {
						continue
					}
					if item.text == "(" {
						level++
					} else if item.text == ")" {
						level--
					} else if level == 0 && item.text == "REFERENCES" {
						reference = true
						index++ // 跳过被引用表名。
					} else if level == 0 && reference {
						end := index
						if item.text == "MATCH" && index+1 < len(items) {
							end++
						} else if item.text == "DEFERRABLE" {
							if index > 0 && items[index-1].text == "NOT" && !items[index-1].quoted {
								item.start = items[index-1].start
							}
							if end+2 < len(items) && items[end+1].text == "INITIALLY" {
								end += 2
							}
						} else {
							continue
						}
						foreignOptions += " " + original[part.start+item.start:part.start+items[end].end]
						index = end
					}
				}
			}
			clauses[part.clause] = original[:part.start] + original[part.end:]
		}
		if key != nil && !removeKey {
			for index, part := range foreignParts {
				if index != position && key.Name != "" && strings.EqualFold(part.name, key.Name) {
					return nil, errors.New("外键名称已存在")
				}
			}
			newDefinition, err := m.foreignKeyDefinition(*key)
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, "\n  "+newDefinition+foreignOptions+"\n")
		}
	}
	retained := clauses[:0]
	for _, clause := range clauses {
		if strings.TrimSpace(clause) != "" {
			retained = append(retained, clause)
		}
	}
	clauses = retained
	// 原始索引/触发器在原生字段改名之后读取，保证引用使用新字段名。
	rows, err = tx.QueryContext(ctx, "SELECT sql FROM main.sqlite_master WHERE tbl_name = ? COLLATE NOCASE AND type IN ('index','trigger') AND sql IS NOT NULL ORDER BY CASE type WHEN 'index' THEN 0 ELSE 1 END,name", tableName)
	if err != nil {
		return nil, err
	}
	dependencies := make([]string, 0)
	for rows.Next() {
		var statement string
		if err = rows.Scan(&statement); err != nil {
			rows.Close()
			return nil, err
		}
		dependencies = append(dependencies, statement)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	var sequence sql.NullInt64
	var hasSequence int
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM main.sqlite_master WHERE type = 'table' AND name = 'sqlite_sequence'").Scan(&hasSequence); err != nil {
		return nil, err
	}
	if hasSequence != 0 {
		err = tx.QueryRowContext(ctx, "SELECT seq FROM main.sqlite_sequence WHERE name = ?", tableName).Scan(&sequence)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	temporaryName := "__sinking_rebuild_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	temporary, _ := m.quote(temporaryName)
	newDefinition := "CREATE TABLE \"main\"." + temporary + " (" + strings.Join(clauses, ",") + definition[closing:]
	if _, err = tx.ExecContext(ctx, newDefinition); err != nil {
		return nil, err
	}
	copyFields := strings.Join(fields, ", ")
	if _, err = tx.ExecContext(ctx, "INSERT INTO \"main\"."+temporary+" ("+copyFields+") SELECT "+copyFields+" FROM "+table); err != nil {
		return nil, fmt.Errorf("已有数据不符合新结构，修改已回滚：%w", err)
	}
	if _, err = tx.ExecContext(ctx, "DROP TABLE "+table); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, "ALTER TABLE \"main\"."+temporary+" RENAME TO "+quotedTable); err != nil {
		return nil, err
	}
	for _, statement := range dependencies {
		if _, err = tx.ExecContext(ctx, statement); err != nil {
			return nil, err
		}
	}
	if sequence.Valid {
		result, err := tx.ExecContext(ctx, "UPDATE main.sqlite_sequence SET seq = MAX(seq, ?) WHERE name = ?", sequence.Int64, tableName)
		if err != nil {
			return nil, err
		}
		changed, err := result.RowsAffected()
		if err != nil {
			return nil, err
		}
		if changed == 0 {
			if _, err = tx.ExecContext(ctx, "INSERT INTO main.sqlite_sequence(name,seq) VALUES (?,?)", tableName, sequence.Int64); err != nil {
				return nil, err
			}
		}
	}
	violations, err := tx.QueryContext(ctx, "PRAGMA main.foreign_key_check")
	if err != nil {
		return nil, err
	}
	if violations.Next() {
		var child, parent string
		var rowid sql.NullInt64
		var id int
		if err = violations.Scan(&child, &rowid, &parent, &id); err != nil {
			violations.Close()
			return nil, err
		}
		violations.Close()
		return nil, fmt.Errorf("外键校验失败，修改已回滚：%s 引用 %s", child, parent)
	}
	err = violations.Err()
	violations.Close()
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("结构修改提交结果未确认，请重新读取结构：%w", err)
	}
	return &Result{
		Columns:   []string{},
		Rows:      [][]interface{}{},
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}
