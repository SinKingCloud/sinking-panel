package sqlserver

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

	"github.com/golang-sql/sqlexp"
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

// statementArgs 校验 SQL 并转换绑定参数。
func (m *Manager) statementArgs(query, argsJSON string) ([]interface{}, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("请输入 SQL 内容")
	}
	if len(query) > 1<<20 {
		return nil, errors.New("SQL 内容超过 1 MB")
	}
	args := make([]interface{}, 0)
	if strings.TrimSpace(argsJSON) != "" {
		if err := m.decode(argsJSON, &args); err != nil {
			return nil, err
		}
	}
	if len(args) > 1000 {
		return nil, errors.New("绑定参数不能超过 1000 个")
	}
	for index, value := range args {
		converted, err := m.argument(value)
		if err != nil {
			return nil, err
		}
		args[index] = converted
	}

	return args, nil
}

// Query 执行带结果集的 SQL，保留原始参数精度。
func (m *Manager) Query(ctx context.Context, query, argsJSON string, rowLimit int) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	if len(query)+len(argsJSON) > 8<<20 {
		return nil, errors.New("操作参数超过 8 MB")
	}
	if rowLimit < 0 {
		return nil, errors.New("查询行数不能小于 0")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	args, err := m.statementArgs(query, argsJSON)
	if err != nil {
		return nil, err
	}
	started := time.Now()

	limit := rowLimit
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}
	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result, err := m.read(rows, limit)
	if err != nil {
		return nil, err
	}
	result.ElapsedMs = time.Since(started).Milliseconds()
	return result, nil
}

// Execute 执行 SQL；根据数据库返回的结果自动识别查询，并保留影响行数。
func (m *Manager) Execute(ctx context.Context, query, argsJSON string, limit int) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	if len(query)+len(argsJSON) > 8<<20 {
		return nil, errors.New("操作参数超过 8 MB")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	args, err := m.statementArgs(query, argsJSON)
	if err != nil {
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
	// 使用驱动的原生结果消息，兼容 SELECT、OUTPUT 和带结果集的存储过程。
	messages := &sqlexp.ReturnMessage{}
	rows, err := m.db.QueryContext(ctx, query, append(args, messages)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resultSets := make([]*Result, 0)
	resultBytes, resultRows := 0, 0
	resultsTruncated := false
	collect := func(data *Result) {
		encoded, _ := json.Marshal(data)
		if len(resultSets) >= 20 || resultBytes+len(encoded) > 8<<20 || resultRows+len(data.Rows) > 1000 {
			resultsTruncated = true
			return
		}
		resultSets = append(resultSets, data)
		resultBytes += len(encoded)
		resultRows += len(data.Rows)
	}
	result := &Result{
		Columns: []string{},
		Rows:    [][]interface{}{},
	}
	for active := true; active; {
		switch message := messages.Message(ctx).(type) {
		case sqlexp.MsgNext:
			data, readErr := m.read(rows, limit)
			if readErr != nil {
				return nil, readErr
			}
			collect(data)
			for rows.Next() {
			}
			if err = rows.Err(); err != nil {
				return nil, err
			}

		case sqlexp.MsgRowsAffected:
			result.Affected += message.Count
		case sqlexp.MsgNextResultSet:
			active = rows.NextResultSet()
		case sqlexp.MsgError:
			return nil, message.Error
		}
		if err = ctx.Err(); err != nil {
			return nil, err
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	affected := result.Affected
	if len(resultSets) > 0 {
		copyResult := *resultSets[0]
		result = &copyResult
	}
	result.Affected = affected
	result.Truncated = result.Truncated || resultsTruncated
	if len(resultSets) > 1 {
		result.Results = resultSets
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
	// 列类型和扫描地址按结果集复用，避免每一行重复分配、解析。
	binaryTypes := make([]bool, len(types))
	for index, columnType := range types {
		typeName := strings.ToUpper(columnType.DatabaseTypeName())
		binaryTypes[index] = strings.Contains(typeName, "BLOB") || strings.Contains(typeName, "BINARY") || typeName == "BYTEA" || typeName == "IMAGE"
	}
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

// Rows 按筛选与排序条件分页读取数据。
func (m *Manager) Rows(ctx context.Context, schemaName, tableName, filtersJSON, filterLogic, orderColumn, orderDirection string, pageNumber, pageSize int) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	join := " AND "
	switch strings.ToLower(strings.TrimSpace(filterLogic)) {
	case "", "and":
	case "or":
		join = " OR "
	default:
		return nil, errors.New("筛选条件关系不合法")
	}
	if len(filtersJSON) > 8<<20 {
		return nil, errors.New("操作参数超过 8 MB")
	}
	if pageNumber < 0 || pageNumber > 1000000 || pageSize < 0 {
		return nil, errors.New("分页参数不合法")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
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
	primary := make([]string, 0)
	for _, column := range columns {
		name := fmt.Sprint(column["name"])
		known[name] = true
		if m.truth(column["primary_key"]) {
			primary = append(primary, name)
		}
	}
	keyColumns, err := m.keyColumns(ctx, schemaName, tableName, columns)
	if err != nil {
		return nil, err
	}
	page, size := pageNumber, pageSize
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 1000 {
		size = 1000
	}
	if page > 1000000 {
		return nil, errors.New("页码过大，请使用筛选条件定位数据")
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
					positions = append(positions, m.placeholder(len(args)))
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
			where = append(where, column+" "+op+" "+m.placeholder(len(args)))
		}
	}
	if len(args) > 2000 {
		return nil, errors.New("筛选绑定参数不能超过 2000 个")
	}
	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, join)
	}
	var total int64
	if err = m.db.QueryRowContext(ctx, "SELECT COUNT_BIG(*) FROM "+table+whereSQL, args...).Scan(&total); err != nil {
		return nil, err
	}
	if total > 0 && int64((page-1)*size) >= total {
		page = int((total-1)/int64(size)) + 1
	}
	order := strings.ToUpper(orderDirection)
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
	offset := (page - 1) * size
	if orderSQL == "" {
		orderSQL = " ORDER BY (SELECT 0)"
	}
	alias := "__panel_row_number"
	for known[alias] {
		alias += "_"
	}
	number, _ := m.quote(alias)
	projection := make([]string, 0, len(columns))
	for _, column := range columns {
		quoted, _ := m.quote(fmt.Sprint(column["name"]))
		projection = append(projection, quoted)
	}
	query := "SELECT " + strings.Join(projection, ",") + " FROM (SELECT " + strings.Join(projection, ",") + ", ROW_NUMBER() OVER (" + orderSQL + ") AS " + number + " FROM " + table + whereSQL + ") AS _panel_rows WHERE " + number + " > " + strconv.Itoa(offset) + " AND " + number + " <= " + strconv.Itoa(offset+size+1) + " ORDER BY " + number
	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result, err := m.read(rows, size)
	if err != nil {
		return nil, err
	}
	result.Total = total
	result.Page = page
	result.PageSize = size
	result.PrimaryKeys = primary
	result.KeyColumns = keyColumns
	result.ElapsedMs = time.Since(started).Milliseconds()
	return result, nil
}

// Insert 添加单行记录。
func (m *Manager) Insert(ctx context.Context, schemaName, tableName, valuesJSON string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	if len(valuesJSON) > 8<<20 {
		return nil, errors.New("操作参数超过 8 MB")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	table, columns, err := m.rowColumns(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}

	names, args, err := m.writeValues(valuesJSON, columns, true)
	if err != nil {
		return nil, err
	}

	if len(names) == 0 {
		if _, err = m.maintenanceTable(ctx, schemaName, tableName); err != nil {
			return nil, err
		}
	}

	quoted := make([]string, len(names))
	positions := make([]string, len(names))
	for i, name := range names {
		quoted[i], _ = m.quote(name)
		positions[i] = m.placeholder(i + 1)
	}
	if len(names) == 0 {
		return m.writeOne(ctx, "INSERT INTO "+table+" DEFAULT VALUES", args)
	}
	return m.writeOne(ctx, "INSERT INTO "+table+" ("+strings.Join(quoted, ",")+") VALUES ("+strings.Join(positions, ",")+")", args)
}

// Update 按完整主键或唯一键修改单行记录。
func (m *Manager) Update(ctx context.Context, schemaName, tableName, valuesJSON, keysJSON string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	if len(valuesJSON)+len(keysJSON) > 8<<20 {
		return nil, errors.New("操作参数超过 8 MB")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	table, columns, err := m.rowColumns(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}

	names, args, err := m.writeValues(valuesJSON, columns)
	if err != nil {
		return nil, err
	}

	primary, keys, err := m.keyValues(ctx, schemaName, tableName, keysJSON, columns)
	if err != nil {
		return nil, err
	}
	assignments := make([]string, len(names))
	for i, name := range names {
		quoted, _ := m.quote(name)
		assignments[i] = quoted + " = " + m.placeholder(i+1)
	}
	where := make([]string, len(primary))
	for i, name := range primary {
		quoted, _ := m.quote(name)
		where[i] = quoted + " = " + m.placeholder(len(args)+i+1)
	}
	args = append(args, keys...)
	return m.writeOne(ctx, "UPDATE "+table+" SET "+strings.Join(assignments, ",")+" WHERE "+strings.Join(where, " AND "), args)
}

// Delete 按完整主键或唯一键删除单行记录。
func (m *Manager) Delete(ctx context.Context, schemaName, tableName, keysJSON string) (_ interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	if len(keysJSON) > 8<<20 {
		return nil, errors.New("操作参数超过 8 MB")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	table, columns, err := m.rowColumns(ctx, schemaName, tableName)
	if err != nil {
		return nil, err
	}

	primary, args, err := m.keyValues(ctx, schemaName, tableName, keysJSON, columns)
	if err != nil {
		return nil, err
	}
	where := make([]string, len(primary))
	for i, name := range primary {
		quoted, _ := m.quote(name)
		where[i] = quoted + " = " + m.placeholder(i+1)
	}
	return m.writeOne(ctx, "DELETE FROM "+table+" WHERE "+strings.Join(where, " AND "), args)
}

// rowColumns 读取目标表及字段，行编辑只接受实际存在的表。
func (m *Manager) rowColumns(ctx context.Context, schemaName, tableName string) (string, []map[string]interface{}, error) {
	_, table, err := m.target(schemaName, tableName)
	if err != nil {
		return "", nil, err
	}
	data, err := m.Columns(ctx, schemaName, tableName)
	if err != nil {
		return "", nil, err
	}
	columns := data.([]map[string]interface{})
	if len(columns) == 0 {
		return "", nil, errors.New("表不存在或当前账户无权访问")
	}
	return table, columns, nil
}

// writeValues 校验写入字段并按固定字段顺序转换值。
func (m *Manager) writeValues(valuesJSON string, columns []map[string]interface{}, allowEmpty ...bool) ([]string, []interface{}, error) {
	values := make(map[string]interface{})
	if err := m.decode(valuesJSON, &values); err != nil {
		return nil, nil, err
	}
	if values == nil {
		return nil, nil, errors.New("字段值必须是 JSON 对象")
	}
	if len(values) == 0 && (len(allowEmpty) == 0 || !allowEmpty[0]) {
		return nil, nil, errors.New("请至少填写一个字段")
	}
	known := make(map[string]map[string]interface{}, len(columns))
	for _, column := range columns {
		known[fmt.Sprint(column["name"])] = column
	}
	names := make([]string, 0, len(values))
	for name := range values {
		column, ok := known[name]
		if !ok {
			return nil, nil, fmt.Errorf("字段 %s 不存在", name)
		}
		if m.truth(column["generated"]) || m.truth(column["identity_column"]) {
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

// keyColumns 优先使用主键，没有主键时使用有效、完整且非空的唯一索引。
func (m *Manager) keyColumns(ctx context.Context, schemaName, tableName string, columns []map[string]interface{}) ([]string, error) {
	keys := make([]string, 0)
	known := make(map[string]bool, len(columns))
	for _, column := range columns {
		name := fmt.Sprint(column["name"])
		known[name] = !m.truth(column["nullable"])
		if m.truth(column["primary_key"]) {
			keys = append(keys, name)
		}
	}
	if len(keys) > 0 {
		return keys, nil
	}
	schema, _, err := m.target(schemaName, tableName)
	if err != nil {
		return nil, err
	}
	version, err := m.serverVersion(ctx)
	if err != nil {
		return nil, err
	}
	filter := ""
	if version >= 10 {
		filter = " AND i.has_filter = 0"
	}
	// 仅接受真实数据表的有效唯一键，过滤索引和包含列不参与行定位。
	query := "SELECT i.name AS index_name, c.name AS column_name FROM sys.indexes i JOIN sys.objects o ON o.object_id = i.object_id JOIN sys.schemas s ON s.schema_id = o.schema_id JOIN sys.index_columns ic ON ic.object_id = i.object_id AND ic.index_id = i.index_id JOIN sys.columns c ON c.object_id = ic.object_id AND c.column_id = ic.column_id WHERE s.name = @p1 AND o.name = @p2 AND o.type = 'U' AND i.is_unique = 1 AND i.is_disabled = 0 AND i.is_hypothetical = 0 AND ic.is_included_column = 0 AND ic.key_ordinal > 0" + filter + " ORDER BY i.name, ic.key_ordinal"
	records, err := m.records(ctx, query, schema, tableName)
	if err != nil {
		return nil, err
	}
	index, valid := "", true
	for _, record := range records {
		name := fmt.Sprint(record["index_name"])
		if name != index {
			if len(keys) > 0 && valid {
				return keys, nil
			}
			index, keys, valid = name, keys[:0], true
		}
		column := fmt.Sprint(record["column_name"])
		valid = valid && known[column]
		keys = append(keys, column)
	}
	if !valid {
		keys = keys[:0]
	}
	return keys, nil
}

// keyValues 要求完整的主键或唯一键，避免修改或删除多行。
func (m *Manager) keyValues(ctx context.Context, schemaName, tableName, keysJSON string, columns []map[string]interface{}) ([]string, []interface{}, error) {
	primary, err := m.keyColumns(ctx, schemaName, tableName, columns)
	if err != nil {
		return nil, nil, err
	}
	if len(primary) == 0 {
		return nil, nil, errors.New("当前表没有可用的主键或非空唯一键，请使用 SQL 编辑数据")
	}
	keys := make(map[string]interface{})
	if err := m.decode(keysJSON, &keys); err != nil {
		return nil, nil, err
	}
	if len(keys) != len(primary) {
		return nil, nil, errors.New("修改或删除需要提供完整的行定位字段")
	}
	args := make([]interface{}, 0, len(primary))
	for _, key := range primary {
		value, ok := keys[key]
		if !ok || value == nil {
			return nil, nil, errors.New("行定位字段不能为空或遗漏")
		}
		value, err := m.argument(value)
		if err != nil {
			return nil, nil, err
		}
		args = append(args, value)
	}
	return primary, args, nil
}

// writeOne 在事务内检查影响行数，不符合单行修改预期时回滚。
func (m *Manager) writeOne(ctx context.Context, query string, args []interface{}) (*Result, error) {
	started := time.Now()
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
		Columns:   []string{},
		Rows:      [][]interface{}{},
		Affected:  affected,
		ElapsedMs: time.Since(started).Milliseconds(),
	}, nil
}
