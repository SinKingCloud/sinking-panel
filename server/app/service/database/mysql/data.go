package mysql

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// decode 保留 JSON 数字的原始精度，数组/对象只在调用方预期的地方接受。
func (m *Manager) decode(value string, result interface{}) error {
	if len(value) > 8<<20 {
		return errors.New("操作参数超过 8 MB")
	}
	decoder := json.NewDecoder(strings.NewReader(value))
	decoder.UseNumber()
	if err := decoder.Decode(result); err != nil {
		return errors.New("操作参数不是有效 JSON")
	}
	var extra interface{}
	if decoder.Decode(&extra) != io.EOF {
		return errors.New("JSON 参数包含多余内容")
	}
	return nil
}

func (m *Manager) argument(value interface{}) (interface{}, error) {
	switch value := value.(type) {
	case json.Number:
		if n, err := value.Int64(); err == nil {
			return n, nil
		}
		return value.String(), nil
	case nil, string, bool, float64, int64, []byte:
		return value, nil
	case map[string]interface{}:
		if len(value) == 2 && value["encoding"] == "base64" {
			if data, ok := value["data"].(string); ok {
				bytes, err := base64.StdEncoding.DecodeString(data)
				if err != nil {
					return nil, errors.New("二进制数据的 Base64 格式不正确")
				}
				return bytes, nil
			}
		}
	}
	return nil, errors.New("字段值必须是字符串、数字、布尔值、null 或 Base64 二进制对象")
}

// Execute 执行一次 SQL，根据数据库返回的结果集自动识别查询，并保留影响行数。
func (m *Manager) Execute(ctx context.Context, query string, limit int) (out interface{}, err error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("请输入 SQL 内容")
	}
	if len(query) > 1<<20 {
		return nil, errors.New("SQL 内容超过 1 MB")
	}
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}

	started := time.Now()
	if limit < 0 {
		return nil, errors.New("返回行数不能小于 0")
	}
	if limit == 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}
	version, maria, err := m.serverVersion(ctx)
	if err != nil {
		return nil, err
	}
	// 同一连接执行一次用户 SQL，再读取该连接的执行状态，避免判断失败后重复执行。
	conn, err := m.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	// MySQL 8.0.32 起允许配置默认计划格式；仅设置当前请求连接，显式 FORMAT 和 ANALYZE 仍由数据库处理。
	if !maria && version >= 80032 {
		if _, err = conn.ExecContext(ctx, "SET SESSION explain_format = 'TRADITIONAL'"); err != nil {
			return nil, err
		}
	}
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result, err := m.read(rows, limit)
	if err != nil {
		return nil, err
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	if len(result.Columns) == 0 {
		result.Columns = []string{}
		if err = conn.QueryRowContext(ctx, "SELECT ROW_COUNT()").Scan(&result.Affected); err != nil {
			return nil, err
		}
	}
	result.ElapsedMs = time.Since(started).Milliseconds()
	return result, nil
}

// read 限制响应大小；大整数和高精度小数按字符串传递，二进制显式标记编码。
func (m *Manager) read(rows *sql.Rows, limit int) (*Result, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	result := &Result{
		Columns: columns,
		Rows:    make([][]interface{}, 0),
	}
	binaryColumns := make(map[string]bool)
	binaryTypes := make([]bool, len(types))
	for index, columnType := range types {
		typeName := strings.ToUpper(columnType.DatabaseTypeName())
		binaryTypes[index] = strings.Contains(typeName, "BLOB") || strings.Contains(typeName, "BINARY")
	}
	// 扫描地址复用；每行的数据单独保留，避免宽表分页时重复分配地址切片。
	destinations := make([]interface{}, len(columns))
	header, err := json.Marshal(columns)
	if err != nil {
		return nil, err
	}
	// 同时计入列名、二进制列名、主键名称和响应字段，避免宽表突破整体响应上限。
	used := len(header)*3 + 1024
	if used > 8<<20 {
		return nil, errors.New("查询列信息超过 8 MB")
	}
	for rows.Next() {
		if len(result.Rows) >= limit {
			result.HasMore = true
			break
		}
		values := make([]interface{}, len(columns))
		for index := range values {
			destinations[index] = &values[index]
		}
		if err = rows.Scan(destinations...); err != nil {
			return nil, err
		}
		truncated := make([][2]int, 0)
		for index, value := range values {
			switch value := value.(type) {
			case int64:
				values[index] = strconv.FormatInt(value, 10)
			case uint64:
				values[index] = strconv.FormatUint(value, 10)
			case float64:
				if math.IsInf(value, 0) || math.IsNaN(value) {
					values[index] = strconv.FormatFloat(value, 'g', -1, 64)
				}
			case time.Time:
				values[index] = value.Format(time.RFC3339Nano)
			case []byte:
				binary := binaryTypes[index] || !utf8.Valid(value)
				if len(value) > 64<<10 {
					value = value[:64<<10]
					truncated = append(truncated, [2]int{len(result.Rows), index})
				}
				if binary {
					values[index] = map[string]string{
						"encoding": "base64",
						"data":     base64.StdEncoding.EncodeToString(value),
					}
					binaryColumns[columns[index]] = true
				} else {
					for !utf8.Valid(value) && len(value) > 0 {
						value = value[:len(value)-1]
					}
					values[index] = string(value)
				}
			case string:
				if len(value) > 64<<10 {
					value = value[:64<<10]
					for !utf8.ValidString(value) && len(value) > 0 {
						value = value[:len(value)-1]
					}
					values[index] = value
					truncated = append(truncated, [2]int{len(result.Rows), index})
				}
			}
		}
		encoded, err := json.Marshal(values)
		if err != nil {
			return nil, err
		}
		rowSize := len(encoded) + 1 + len(truncated)*32
		if used+rowSize > 8<<20 {
			result.HasMore = true
			result.Truncated = true
			break
		}
		used += rowSize
		result.Rows = append(result.Rows, values)
		if len(truncated) > 0 {
			result.Truncated = true
			result.TruncatedCells = append(result.TruncatedCells, truncated...)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	for _, column := range columns {
		if binaryColumns[column] {
			result.BinaryColumns = append(result.BinaryColumns, column)
		}
	}
	return result, nil
}

// Rows 分页浏览表数据，使用完整主键保持排序稳定。
func (m *Manager) Rows(ctx context.Context, schemaName, tableName, filtersJSON, filterLogic, orderColumn, direction string, page, pageSize int) (out interface{}, err error) {
	if len(filtersJSON) > 8<<20 {
		return nil, errors.New("操作参数超过 8 MB")
	}
	if page < 0 || page > 1000000 || pageSize < 0 {
		return nil, errors.New("分页参数不合法")
	}
	if filterLogic == "" {
		filterLogic = "and"
	}
	if filterLogic != "and" && filterLogic != "or" {
		return nil, errors.New("筛选方式仅支持 and 或 or")
	}
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}

	started := time.Now()
	_, table, err := m.target(schemaName, tableName)
	if err != nil {
		return nil, err
	}
	data, err := m.Columns(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	columns := data.([]map[string]interface{})
	if len(columns) == 0 {
		return nil, errors.New("表不存在或当前账户无权访问")
	}
	known := make(map[string]bool, len(columns))
	for _, column := range columns {
		known[fmt.Sprint(column["name"])] = true
	}
	primary, keyColumns, err := m.rowKeyColumns(ctx, schemaName, tableName, columns)
	if err != nil {
		return nil, err
	}
	size := pageSize
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 1000 {
		size = 1000
	}
	args := make([]interface{}, 0)
	where := make([]string, 0)
	if filtersJSON != "" {
		var filters []struct {
			Column   string      `json:"column"`
			Operator string      `json:"operator"`
			Value    interface{} `json:"value"`
		}
		if err = m.decode(filtersJSON, &filters); err != nil {
			return nil, err
		}
		if len(filters) > 64 {
			return nil, errors.New("筛选条件不能超过 64 个")
		}
		for _, filter := range filters {
			if !known[filter.Column] {
				return nil, errors.New("筛选字段不存在")
			}
			column, _ := m.quote(filter.Column)
			switch filter.Operator {
			case "null":
				where = append(where, column+" IS NULL")
				continue
			case "not_null":
				where = append(where, column+" IS NOT NULL")
				continue
			case "in":
				list, ok := filter.Value.([]interface{})
				if !ok || len(list) == 0 || len(list) > 1000 {
					return nil, errors.New("IN 筛选需要包含 1 至 1000 项的数组")
				}
				positions := make([]string, 0, len(list))
				for _, value := range list {
					value, err = m.argument(value)
					if err != nil {
						return nil, err
					}
					args = append(args, value)
					positions = append(positions, "?")
				}
				where = append(where, column+" IN ("+strings.Join(positions, ",")+")")
				continue
			}
			op, ok := map[string]string{
				"eq":   "=",
				"ne":   "<>",
				"gt":   ">",
				"gte":  ">=",
				"lt":   "<",
				"lte":  "<=",
				"like": "LIKE",
			}[filter.Operator]
			if !ok {
				return nil, errors.New("不支持的筛选操作符")
			}
			if filter.Value == nil && (filter.Operator == "eq" || filter.Operator == "ne") {
				suffix := " IS NULL"
				if filter.Operator == "ne" {
					suffix = " IS NOT NULL"
				}
				where = append(where, column+suffix)
				continue
			}
			value, err := m.argument(filter.Value)
			if err != nil {
				return nil, err
			}
			args = append(args, value)
			where = append(where, column+" "+op+" ?")
		}
	}
	if len(args) > 2000 {
		return nil, errors.New("筛选绑定参数不能超过 2000 个")
	}
	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " "+strings.ToUpper(filterLogic)+" ")
	}
	order := strings.ToUpper(direction)
	if order == "" {
		order = "ASC"
	}
	if order != "ASC" && order != "DESC" {
		return nil, errors.New("排序方向仅支持 asc 或 desc")
	}
	orderBy := make([]string, 0)
	if orderColumn != "" {
		if !known[orderColumn] {
			return nil, errors.New("排序字段不存在")
		}
		quoted, _ := m.quote(orderColumn)
		orderBy = append(orderBy, quoted+" "+order)
	}
	for _, key := range keyColumns {
		if key != orderColumn {
			quoted, _ := m.quote(key)
			orderBy = append(orderBy, quoted+" ASC")
		}
	}
	orderSQL := ""
	if len(orderBy) > 0 {
		orderSQL = " ORDER BY " + strings.Join(orderBy, ",")
	}
	// 与数据查询共用筛选条件，先完成统计，避免同时占用两个结果游标。
	var total int64
	if err = m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table+whereSQL, args...).Scan(&total); err != nil {
		return nil, err
	}
	if total == 0 {
		page = 1
	} else if lastPage := (total-1)/int64(size) + 1; int64(page) > lastPage {
		// 删除末页最后一条数据后，自动回到仍有数据的最后一页。
		page = int(lastPage)
	}
	query := "SELECT * FROM " + table + whereSQL + orderSQL
	offset := (page - 1) * size
	query += " LIMIT " + strconv.Itoa(size+1) + " OFFSET " + strconv.Itoa(offset)
	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result, err := m.read(rows, size)
	if err != nil {
		return nil, err
	}
	result.Page = page
	result.PageSize = size
	result.Total = &total
	result.PrimaryKeys = primary
	result.KeyColumns = keyColumns
	result.ElapsedMs = time.Since(started).Milliseconds()
	return result, nil
}

// Insert 插入一行数据，字段名与值分别校验和绑定。
func (m *Manager) Insert(ctx context.Context, schemaName, tableName, valuesJSON string) (out interface{}, err error) {
	if len(valuesJSON) > 8<<20 {
		return nil, errors.New("操作参数超过 8 MB")
	}
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}

	started := time.Now()
	_, table, err := m.target(schemaName, tableName)
	if err != nil {
		return nil, err
	}
	known, _, err := m.editColumns(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	names, args, err := m.editValues(valuesJSON, known, true)
	if err != nil {
		return nil, err
	}

	if len(names) == 0 {
		_, kind, _, err := m.maintenanceTable(ctx, schemaName, tableName)
		if err != nil {
			return nil, err
		}
		if kind != "BASE TABLE" {
			return nil, errors.New("默认记录只能添加到数据表")
		}
	}
	columns := make([]string, 0, len(names))
	positions := make([]string, 0, len(names))
	for _, name := range names {
		column, _ := m.quote(name)
		columns = append(columns, column)
		positions = append(positions, "?")
	}
	query := "INSERT INTO " + table + " (" + strings.Join(columns, ",") + ") VALUES (" + strings.Join(positions, ",") + ")"
	result, err := m.executeRow(ctx, query, args)
	if err != nil {
		return nil, err
	}
	result.ElapsedMs = time.Since(started).Milliseconds()
	return result, nil
}

// Update 按完整主键或非空唯一键更新一行数据，影响行数异常时回滚。
func (m *Manager) Update(ctx context.Context, schemaName, tableName, valuesJSON, keysJSON string) (out interface{}, err error) {
	if len(valuesJSON)+len(keysJSON) > 8<<20 {
		return nil, errors.New("操作参数超过 8 MB")
	}
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}

	started := time.Now()
	_, table, err := m.target(schemaName, tableName)
	if err != nil {
		return nil, err
	}
	known, primary, err := m.editColumns(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	names, args, err := m.editValues(valuesJSON, known)
	if err != nil {
		return nil, err
	}
	keys, err := m.editKeys(keysJSON, primary)
	if err != nil {
		return nil, err
	}
	assignments := make([]string, 0, len(names))
	where := make([]string, 0, len(primary))
	for _, name := range names {
		column, _ := m.quote(name)
		assignments = append(assignments, column+" = "+"?")
	}
	for index, key := range primary {
		column, _ := m.quote(key)
		args = append(args, keys[index])
		where = append(where, column+" = "+"?")
	}
	query := "UPDATE " + table + " SET " + strings.Join(assignments, ",") + " WHERE " + strings.Join(where, " AND ")
	result, err := m.executeRow(ctx, query, args)
	if err != nil {
		return nil, err
	}
	result.ElapsedMs = time.Since(started).Milliseconds()
	return result, nil
}

// Delete 按完整主键或非空唯一键删除一行数据，没有可靠定位条件时拒绝删除。
func (m *Manager) Delete(ctx context.Context, schemaName, tableName, keysJSON string) (out interface{}, err error) {
	if len(keysJSON) > 8<<20 {
		return nil, errors.New("操作参数超过 8 MB")
	}
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}

	started := time.Now()
	_, table, err := m.target(schemaName, tableName)
	if err != nil {
		return nil, err
	}
	_, primary, err := m.editColumns(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	args, err := m.editKeys(keysJSON, primary)
	if err != nil {
		return nil, err
	}
	where := make([]string, 0, len(primary))
	for _, key := range primary {
		column, _ := m.quote(key)
		where = append(where, column+" = "+"?")
	}
	result, err := m.executeRow(ctx, "DELETE FROM "+table+" WHERE "+strings.Join(where, " AND "), args)
	if err != nil {
		return nil, err
	}
	result.ElapsedMs = time.Since(started).Milliseconds()
	return result, nil
}

// editColumns 读取行编辑所需的字段属性与可靠定位字段。
func (m *Manager) editColumns(ctx context.Context, schemaName, tableName string) (map[string]map[string]interface{}, []string, error) {
	data, err := m.Columns(ctx, schemaName, tableName)
	if err != nil {
		return nil, nil, err
	}
	columns := data.([]map[string]interface{})
	if len(columns) == 0 {
		return nil, nil, errors.New("表不存在或当前账户无权访问")
	}
	known := make(map[string]map[string]interface{}, len(columns))
	for _, column := range columns {
		known[fmt.Sprint(column["name"])] = column
	}
	_, keyColumns, err := m.rowKeyColumns(ctx, schemaName, tableName, columns)
	return known, keyColumns, err
}

// rowKeyColumns 优先使用真实主键；没有主键时，选择完整、非空的唯一索引。
// 查询和写入共用相同选择顺序，视图、表达式和前缀索引不能作为行定位条件。
func (m *Manager) rowKeyColumns(ctx context.Context, schemaName, tableName string, columns []map[string]interface{}) ([]string, []string, error) {
	schema, table, err := m.target(schemaName, tableName)
	if err != nil {
		return nil, nil, err
	}
	var kind string
	if err = m.db.QueryRowContext(ctx, "SELECT TABLE_TYPE FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?", schema, tableName).Scan(&kind); err != nil {
		return nil, nil, err
	}
	primary, keys := []string{}, []string{}
	if kind != "BASE TABLE" {
		return primary, keys, nil
	}
	records, err := m.records(ctx, "SHOW INDEX FROM "+table)
	if err != nil {
		return nil, nil, err
	}
	known := make(map[string]map[string]interface{}, len(columns))
	for _, column := range columns {
		known[fmt.Sprint(column["name"])] = column
	}
	candidates := make(map[string]map[int]string)
	invalid := make(map[string]bool)
	for _, record := range records {
		name := fmt.Sprint(record["key_name"])
		if candidates[name] == nil {
			candidates[name] = make(map[int]string)
		}
		field, ok := record["column_name"].(string)
		column := known[field]
		position, positionErr := strconv.Atoi(fmt.Sprint(record["seq_in_index"]))
		if m.truth(record["non_unique"]) || !ok || column == nil || m.truth(column["nullable"]) || record["sub_part"] != nil || positionErr != nil || position < 1 {
			invalid[name] = true
		}
		candidates[name][position] = field
	}
	names := make([]string, 0, len(candidates))
	for name := range candidates {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if invalid[name] {
			continue
		}
		fields := make([]string, 0, len(candidates[name]))
		for position := 1; position <= len(candidates[name]); position++ {
			field, ok := candidates[name][position]
			if !ok {
				fields = nil
				break
			}
			fields = append(fields, field)
		}
		if len(fields) == 0 {
			continue
		}
		if name == "PRIMARY" {
			return fields, fields, nil
		}
		if len(keys) == 0 {
			keys = fields
		}
	}
	return primary, keys, nil
}

func (m *Manager) editValues(input string, known map[string]map[string]interface{}, allowEmpty ...bool) ([]string, []interface{}, error) {
	values := make(map[string]interface{})
	if err := m.decode(input, &values); err != nil {
		return nil, nil, err
	}
	if values == nil {
		return nil, nil, errors.New("字段值必须是 JSON 对象")
	}
	if len(values) == 0 && (len(allowEmpty) == 0 || !allowEmpty[0]) {
		return nil, nil, errors.New("请至少填写一个字段")
	}
	names := make([]string, 0, len(values))
	for name := range values {
		column, ok := known[name]
		if !ok {
			return nil, nil, fmt.Errorf("字段 %s 不存在", name)
		}
		if m.truth(column["generated"]) {
			return nil, nil, fmt.Errorf("字段 %s 由数据库自动生成，不能直接写入", name)
		}
		names = append(names, name)
	}
	sort.Strings(names)
	args := make([]interface{}, 0, len(names))
	for _, name := range names {
		value, err := m.argument(values[name])
		if err != nil {
			return nil, nil, err
		}
		args = append(args, value)
	}
	return names, args, nil
}

func (m *Manager) editKeys(input string, primary []string) ([]interface{}, error) {
	if len(primary) == 0 {
		return nil, errors.New("当前表没有主键或非空唯一键，请使用 SQL 编辑数据")
	}
	keys := make(map[string]interface{})
	if err := m.decode(input, &keys); err != nil {
		return nil, err
	}
	if len(keys) != len(primary) {
		return nil, errors.New("修改或删除需要提供完整的行定位字段")
	}
	args := make([]interface{}, 0, len(primary))
	for _, name := range primary {
		value, ok := keys[name]
		if !ok || value == nil {
			return nil, errors.New("行定位字段不能为空或遗漏")
		}
		value, err := m.argument(value)
		if err != nil {
			return nil, err
		}
		args = append(args, value)
	}
	return args, nil
}

// executeRow 单行写操作共用事务与影响行数检查，不决定具体业务操作。
func (m *Manager) executeRow(ctx context.Context, query string, args []interface{}) (*Result, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected != 1 {
		return nil, fmt.Errorf("预期修改 1 行，实际影响 %d 行，已回滚；请刷新数据后重试", affected)
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &Result{
		Columns:  []string{},
		Rows:     [][]interface{}{},
		Affected: affected,
	}, nil
}
