package postgres

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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/stdlib"
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

// queryArguments 验证 SQL 内容并保留绑定参数的数值精度。
func (m *Manager) queryArguments(query, argsJSON string) ([]interface{}, error) {
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

// Query 执行返回结果集的 SQL，限制行数与响应大小。
func (m *Manager) Query(ctx context.Context, query, argsJSON string, rowLimit int) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	if len(query)+len(argsJSON) > 8<<20 {
		return nil, errors.New("操作参数超过 8 MB")
	}
	if rowLimit < 0 {
		return nil, errors.New("查询行数不能小于 0")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}

	args, err := m.queryArguments(query, argsJSON)
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
func (m *Manager) Execute(ctx context.Context, query, argsJSON string, limit int) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	if len(query)+len(argsJSON) > 8<<20 {
		return nil, errors.New("操作参数超过 8 MB")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err = ctx.Err(); err != nil {
		return nil, err
	}

	args, err := m.queryArguments(query, argsJSON)
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
	conn, err := m.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

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
	err = conn.Raw(func(connection interface{}) error {
		native := connection.(*stdlib.Conn).Conn()
		readResult := func(rows pgx.Rows) error {
			data := &Result{
				Columns: []string{},
				Rows:    [][]interface{}{},
			}
			if len(rows.FieldDescriptions()) > 0 {
				var readErr error
				data, readErr = m.read(rows, limit)
				if readErr != nil {
					rows.Close()
					return readErr
				}
			}
			rows.Close()
			if readErr := rows.Err(); readErr != nil {
				return readErr
			}
			data.Affected = rows.CommandTag().RowsAffected()
			result.Affected += data.Affected
			collect(data)
			return nil
		}

		if len(args) > 0 {
			rows, queryErr := native.Query(ctx, query, args...)
			if queryErr != nil {
				return queryErr
			}
			return readResult(rows)
		}
		// 原生批次一次提交所有语句，逐个消费结果并检查错误，保留原有多语句能力。
		results := native.PgConn().Exec(ctx, query)
		var readErr error
		for results.NextResult() {
			readErr = readResult(pgx.RowsFromResultReader(native.TypeMap(), results.ResultReader()))
			if readErr != nil {
				break
			}
		}
		closeErr := results.Close()
		if readErr != nil {
			return readErr
		}
		return closeErr
	})
	if err != nil {
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
func (m *Manager) read(rows interface {
	Next() bool
	Scan(...interface{}) error
	Err() error
}, limit int) (*Result, error) {
	var columns, types []string
	var err error
	scan := rows.Scan
	switch data := rows.(type) {
	case *sql.Rows:
		columns, err = data.Columns()
		if err != nil {
			return nil, err
		}
		columnTypes, typeErr := data.ColumnTypes()
		if typeErr != nil {
			return nil, typeErr
		}
		for _, column := range columnTypes {
			types = append(types, column.DatabaseTypeName())
		}
	case pgx.Rows:
		typeMap := pgtype.NewMap()
		if data.Conn() != nil {
			typeMap = data.Conn().TypeMap()
		}
		fields := data.FieldDescriptions()
		for _, field := range fields {
			columns = append(columns, field.Name)
			name := ""
			if columnType, ok := typeMap.TypeForOID(field.DataTypeOID); ok {
				name = columnType.Name
			}
			types = append(types, name)
		}
		// 与 database/sql 使用同一套驱动值转换，保留大数、日期和二进制的既有响应格式。
		scan = func(destinations ...interface{}) error {
			for index, raw := range data.RawValues() {
				var value interface{}
				if raw != nil {
					field := fields[index]
					if columnType, ok := typeMap.TypeForOID(field.DataTypeOID); ok {
						var decodeErr error
						value, decodeErr = columnType.Codec.DecodeDatabaseSQLValue(typeMap, field.DataTypeOID, field.Format, raw)
						if decodeErr != nil {
							return decodeErr
						}
					} else {
						value = string(raw)
					}
				}
				*destinations[index].(*interface{}) = value
			}
			return nil
		}
	default:
		return nil, errors.New("查询结果类型不支持")
	}
	result := &Result{
		Columns: columns,
		Rows:    make([][]interface{}, 0),
	}
	binaryColumns := make(map[string]bool)
	// 列类型和扫描地址按结果集复用，避免每一行重复分配、解析。
	binaryTypes := make([]bool, len(types))
	for index, columnType := range types {
		typeName := strings.ToUpper(columnType)
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
		if err = scan(destinations...); err != nil {
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
func (m *Manager) Rows(ctx context.Context, schemaName, tableName, filtersJSON, filterLogic, orderColumn, orderDirection string, pageNumber, pageSize int) (out interface{}, err error) {
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
	if err = m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table+whereSQL, args...).Scan(&total); err != nil {
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
	result.Total = total
	result.Page = page
	result.PageSize = size
	result.PrimaryKeys = primary
	result.KeyColumns = keyColumns
	result.ElapsedMs = time.Since(started).Milliseconds()
	return result, nil
}

// Insert 插入一行数据，字段名与值分别校验和绑定。
func (m *Manager) Insert(ctx context.Context, schemaName, tableName, valuesJSON string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	if len(valuesJSON) > 8<<20 {
		return nil, errors.New("操作参数超过 8 MB")
	}
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
		_, kind, err := m.maintenanceTable(ctx, schemaName, tableName)
		if err != nil {
			return nil, err
		}
		if kind != "r" && kind != "p" {
			return nil, errors.New("默认记录只能添加到数据表")
		}
	}
	columns := make([]string, 0, len(names))
	positions := make([]string, 0, len(names))
	for index, name := range names {
		column, _ := m.quote(name)
		columns = append(columns, column)
		positions = append(positions, m.placeholder(index+1))
	}
	query := "INSERT INTO " + table + " (" + strings.Join(columns, ",") + ") VALUES (" + strings.Join(positions, ",") + ")"
	if len(names) == 0 {
		query = "INSERT INTO " + table + " DEFAULT VALUES"
	}
	result, err := m.executeRow(ctx, query, args)
	if err != nil {
		return nil, err
	}
	result.ElapsedMs = time.Since(started).Milliseconds()
	return result, nil
}

// Update 按完整主键或唯一键更新一行数据，影响行数异常时回滚。
func (m *Manager) Update(ctx context.Context, schemaName, tableName, valuesJSON, keysJSON string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	if len(valuesJSON)+len(keysJSON) > 8<<20 {
		return nil, errors.New("操作参数超过 8 MB")
	}
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
	for index, name := range names {
		column, _ := m.quote(name)
		assignments = append(assignments, column+" = "+m.placeholder(index+1))
	}
	for index, key := range primary {
		column, _ := m.quote(key)
		args = append(args, keys[index])
		where = append(where, column+" = "+m.placeholder(len(args)))
	}
	query := "UPDATE " + table + " SET " + strings.Join(assignments, ",") + " WHERE " + strings.Join(where, " AND ")
	result, err := m.executeRow(ctx, query, args)
	if err != nil {
		return nil, err
	}
	result.ElapsedMs = time.Since(started).Milliseconds()
	return result, nil
}

// Delete 按完整主键或唯一键删除一行数据，无法唯一定位时拒绝操作。
func (m *Manager) Delete(ctx context.Context, schemaName, tableName, keysJSON string) (out interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	if len(keysJSON) > 8<<20 {
		return nil, errors.New("操作参数超过 8 MB")
	}
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
	for index, key := range primary {
		column, _ := m.quote(key)
		where = append(where, column+" = "+m.placeholder(index+1))
	}
	result, err := m.executeRow(ctx, "DELETE FROM "+table+" WHERE "+strings.Join(where, " AND "), args)
	if err != nil {
		return nil, err
	}
	result.ElapsedMs = time.Since(started).Milliseconds()
	return result, nil
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
	keyCount := "x.indnatts"
	if version >= 110000 {
		keyCount = "x.indnkeyatts"
	}
	// INCLUDE 不参与唯一性；部分、表达式、无效及延迟唯一索引不能用作行定位。
	query := "SELECT i.relname AS index_name, a.attname AS column_name FROM pg_index x JOIN pg_class t ON t.oid = x.indrelid JOIN pg_namespace n ON n.oid = t.relnamespace JOIN pg_class i ON i.oid = x.indexrelid CROSS JOIN generate_series(0, 31) AS k(position) JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = x.indkey[k.position] WHERE n.nspname = $1 AND t.relname = $2 AND t.relkind IN ('r','p') AND x.indisunique AND x.indisvalid AND x.indisready AND x.indimmediate AND x.indpred IS NULL AND x.indexprs IS NULL AND " + keyCount + " <= 32 AND k.position < " + keyCount + " AND NOT a.attisdropped ORDER BY i.relname, k.position"
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

// editColumns 读取行编辑所需的字段属性与唯一定位字段。
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
		name := fmt.Sprint(column["name"])
		known[name] = column
	}
	keys, err := m.keyColumns(ctx, schemaName, tableName, columns)
	return known, keys, err
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
		return nil, errors.New("当前表没有可用的主键或非空唯一键，请使用 SQL 编辑数据")
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
