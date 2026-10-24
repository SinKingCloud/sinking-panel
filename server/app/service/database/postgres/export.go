package postgres

import (
	"bufio"
	"bytes"
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
)

// Export 输出 plain SQL：结构、原生 COPY 数据、序列、视图、程序、索引和触发器。
// 空 table 导出整个数据库，schema 可缩小范围；只有完整写入后才发布文件。
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
	if table != "" && schema == "" {
		schema = "public"
	}
	if old, e := os.Lstat(path); e == nil {
		if !old.Mode().IsRegular() {
			return nil, errors.New("导出路径必须是普通文件")
		}
	} else if !errors.Is(e, os.ErrNotExist) {
		return nil, e
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".postgres-export-*.sql")
	if err != nil {
		return nil, err
	}
	temporary := file.Name()
	defer func() { _ = file.Close(); _ = os.Remove(temporary) }()
	writer := bufio.NewWriterSize(file, 64<<10)
	var count, statements int64
	write := func(text string) error {
		if e := ctx.Err(); e != nil {
			return e
		}
		if _, e := writer.WriteString(strings.TrimSuffix(strings.TrimSpace(text), ";") + ";\n\n"); e != nil {
			return e
		}
		statements++
		return nil
	}
	literal := func(text string) string {
		return "E'" + strings.ReplaceAll(strings.ReplaceAll(text, "\\", "\\\\"), "'", "''") + "'"
	}
	conn, err := m.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Raw(func(interface{}) error { return driver.ErrBadConn }); _ = conn.Close() }()
	if _, err = conn.ExecContext(ctx, "SET client_encoding='UTF8'"); err != nil {
		return nil, err
	}
	if _, err = conn.ExecContext(ctx, "BEGIN ISOLATION LEVEL REPEATABLE READ READ ONLY"); err != nil {
		return nil, err
	}
	records := func(query string, args ...interface{}) ([]map[string]interface{}, error) {
		rows, e := conn.QueryContext(ctx, query, args...)
		if e != nil {
			return nil, e
		}
		defer rows.Close()
		names, e := rows.Columns()
		if e != nil {
			return nil, e
		}
		result := []map[string]interface{}{}
		values, pointers := make([]interface{}, len(names)), make([]interface{}, len(names))
		for i := range values {
			pointers[i] = &values[i]
		}
		for rows.Next() {
			if len(result) >= 10000 {
				return nil, errors.New("对象超过 10000，请按模式或数据表分别导出")
			}
			if e = rows.Scan(pointers...); e != nil {
				return nil, e
			}
			record := map[string]interface{}{}
			for i, value := range values {
				if b, ok := value.([]byte); ok {
					value = string(b)
				}
				record[names[i]] = value
			}
			result = append(result, record)
		}
		return result, rows.Err()
	}
	version, err := m.serverVersion(ctx)
	if err != nil {
		return nil, err
	}
	selected := "n.nspname NOT IN ('information_schema','pg_catalog') AND substr(n.nspname, 1, 3) <> 'pg_'"
	args := []interface{}{}
	if schema != "" {
		if schema == "information_schema" || strings.HasPrefix(schema, "pg_") {
			return nil, errors.New("请选择业务模式")
		}
		selected += " AND n.nspname=$1"
		args = append(args, schema)
	}
	tableWhere := selected
	if table != "" {
		tableWhere += " AND c.relname=$2"
		args = append(args, table)
	}
	objects, err := records("SELECT c.oid::bigint AS oid, n.nspname AS schema, c.relname AS name, c.relkind AS kind FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE c.relkind IN ('r','p','v','m','f') AND "+tableWhere+" ORDER BY n.nspname,c.relname", args...)
	if err != nil {
		return nil, err
	}
	if table != "" && len(objects) == 0 {
		return nil, errors.New("表或视图不存在")
	}
	byId := map[int64]map[string]interface{}{}
	for _, item := range objects {
		byId[item["oid"].(int64)] = item
		if item["kind"] == "f" {
			return nil, errors.New("外部表需要对应 FDW 与远端数据源，不能作为普通 SQL 备份导出")
		}
	}
	dependencyArgs := args
	if table != "" {
		dependencyArgs = args[:1]
	}
	// 不发布缺少安全策略、类型或独立数据的“成功备份”；这些对象交给 pg_dump。
	rejectUnsupported := func(query, reason string, args ...interface{}) error {
		items, e := records(query, args...)
		if e != nil {
			return e
		}
		if len(items) > 0 {
			return fmt.Errorf("%s [%s]，请使用 pg_dump 导出", reason, fmt.Sprint(items[0]["name"]))
		}
		return nil
	}
	if version >= 90500 {
		// off 不会绕过 RLS；权限不足时会报错，防止只导出被允许看到的部分行。
		if _, err = conn.ExecContext(ctx, "SET LOCAL row_security=off"); err != nil {
			return nil, err
		}
		if err = rejectUnsupported("SELECT n.nspname||'.'||c.relname AS name FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE c.relkind IN ('r','p') AND "+tableWhere+" AND (c.relrowsecurity OR c.relforcerowsecurity OR EXISTS(SELECT 1 FROM pg_policy p WHERE p.polrelid=c.oid)) ORDER BY n.nspname,c.relname LIMIT 1", "暂不支持包含行级安全或策略的 SQL 备份", args...); err != nil {
			return nil, err
		}
	}
	if err = rejectUnsupported("SELECT n.nspname||'.'||c.relname AS name FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE c.relkind IN ('v','m') AND "+tableWhere+" AND array_length(c.reloptions,1)>0 ORDER BY n.nspname,c.relname LIMIT 1", "暂不支持带安全或存储选项的视图备份", args...); err != nil {
		return nil, err
	}
	if err = rejectUnsupported("SELECT n.nspname||'.'||c.relname||'.'||r.rulename AS name FROM pg_rewrite r JOIN pg_class c ON c.oid=r.ev_class JOIN pg_namespace n ON n.oid=c.relnamespace WHERE r.rulename<>'_RETURN' AND "+tableWhere+" ORDER BY n.nspname,c.relname,r.rulename LIMIT 1", "暂不支持自定义规则的 SQL 备份", args...); err != nil {
		return nil, err
	}
	unmanagedType := ""
	if version >= 90100 {
		unmanagedType = " AND NOT EXISTS(SELECT 1 FROM pg_depend d WHERE d.deptype='e' AND ((d.classid='pg_type'::regclass AND d.objid=t.oid) OR (d.classid='pg_class'::regclass AND d.objid=c.oid)))"
	}
	if version >= 140000 {
		// 扩展范围类型的 multirange 是自动创建的，不当作缺少定义的独立类型。
		unmanagedType += " AND NOT EXISTS(SELECT 1 FROM pg_range r JOIN pg_depend d ON d.classid='pg_type'::regclass AND d.objid=r.rngtypid AND d.deptype='e' WHERE r.rngmultitypid=t.oid)"
	}
	if err = rejectUnsupported("SELECT n.nspname||'.'||t.typname AS name FROM pg_type t JOIN pg_namespace n ON n.oid=t.typnamespace LEFT JOIN pg_class c ON c.oid=t.typrelid WHERE "+selected+" AND t.typtype NOT IN ('e','d') AND (t.typtype<>'c' OR c.relkind='c') AND NOT (t.typelem<>0 AND t.typinput='pg_catalog.array_in'::regproc)"+unmanagedType+" ORDER BY n.nspname,t.typname LIMIT 1", "暂不支持独立复合、范围或自定义基类型的 SQL 备份", dependencyArgs...); err != nil {
		return nil, err
	}
	if version >= 90100 {
		if err = rejectUnsupported("SELECT n.nspname||'.'||c.relname AS name FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE "+tableWhere+" AND EXISTS(SELECT 1 FROM pg_depend d WHERE d.classid='pg_class'::regclass AND d.objid=c.oid AND d.deptype='e') ORDER BY n.nspname,c.relname LIMIT 1", "暂不支持扩展受管关系对象的 SQL 备份", args...); err != nil {
			return nil, err
		}
	}
	if table == "" {
		catalog := "pg_largeobject"
		if version >= 90000 {
			catalog = "pg_largeobject_metadata"
		}
		if err = rejectUnsupported("SELECT 'large object' AS name FROM "+catalog+" LIMIT 1", "暂不支持包含大对象的 SQL 备份"); err != nil {
			return nil, err
		}
	}
	// 继承、分区与视图按真实依赖排序，不依赖表名的字母顺序。
	dependencies, err := records("SELECT i.inhrelid::bigint AS child,i.inhparent::bigint AS parent FROM pg_inherits i JOIN pg_class c ON c.oid=i.inhrelid JOIN pg_namespace n ON n.oid=c.relnamespace WHERE "+selected+" UNION SELECT r.ev_class::bigint AS child,d.refobjid::bigint AS parent FROM pg_rewrite r JOIN pg_depend d ON d.classid='pg_rewrite'::regclass AND d.objid=r.oid AND d.refclassid='pg_class'::regclass JOIN pg_class c ON c.oid=r.ev_class JOIN pg_namespace n ON n.oid=c.relnamespace WHERE r.ev_class<>d.refobjid AND "+selected, dependencyArgs...)
	if err != nil {
		return nil, err
	}
	depends := map[int64][]int64{}
	for _, item := range dependencies {
		child, parent := item["child"].(int64), item["parent"].(int64)
		if byId[child] != nil && byId[parent] != nil {
			depends[child] = append(depends[child], parent)
		}
	}
	ranks := map[int64]int{}
	visiting := map[int64]bool{}
	var rank func(int64) (int, error)
	rank = func(id int64) (int, error) {
		if visiting[id] {
			return 0, errors.New("对象存在循环依赖，无法生成可靠的 SQL 恢复顺序")
		}
		if n, ok := ranks[id]; ok {
			return n, nil
		}
		visiting[id] = true
		n := 0
		for _, parent := range depends[id] {
			r, e := rank(parent)
			if e != nil {
				return 0, e
			}
			if r+1 > n {
				n = r + 1
			}
		}
		visiting[id] = false
		ranks[id] = n
		return n, nil
	}
	for id := range byId {
		if _, err = rank(id); err != nil {
			return nil, err
		}
	}
	sort.SliceStable(objects, func(i, j int) bool { return ranks[objects[i]["oid"].(int64)] < ranks[objects[j]["oid"].(int64)] })
	rowCounts := map[int64]int64{}
	totalWeight := float64(0)
	if len(progress) > 0 && progress[0] != nil {
		// COUNT 与 COPY 使用同一快照；空表也分配一份工作量，完成后推进进度。
		for index, item := range objects {
			if item["kind"] == "r" {
				_, target, e := m.target(fmt.Sprint(item["schema"]), fmt.Sprint(item["name"]))
				if e != nil {
					return nil, e
				}
				var rows int64
				if e = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM ONLY "+target).Scan(&rows); e != nil {
					return nil, e
				}
				rowCounts[item["oid"].(int64)] = rows
				totalWeight += float64(max(rows, 1))
			}
			report(1+9*float64(index+1)/float64(len(objects)), "统计导出数据")
		}
	}
	report(10, "写入数据库结构")
	if _, err = writer.WriteString("-- Postgres plain SQL backup\n-- 在已连接的数据库恢复，不创建或切换数据库。\n\n"); err != nil {
		return nil, err
	}
	if err = write("SET client_encoding='UTF8'"); err != nil {
		return nil, err
	}
	if err = write("SET standard_conforming_strings=on"); err != nil {
		return nil, err
	}
	if err = write("SET check_function_bodies=false"); err != nil {
		return nil, err
	}
	if _, err = conn.ExecContext(ctx, "SET LOCAL DateStyle='ISO'; SET LOCAL TIME ZONE 'UTC'"); err != nil {
		return nil, err
	}
	if err = write("SET DateStyle='ISO'"); err != nil {
		return nil, err
	}
	if version >= 80400 {
		if _, err = conn.ExecContext(ctx, "SET LOCAL IntervalStyle='postgres'"); err != nil {
			return nil, err
		}
		if err = write("SET IntervalStyle='postgres'"); err != nil {
			return nil, err
		}
	}
	// escape 格式的 bytea 也能在尚未支持十六进制输入的旧版本恢复。
	if version >= 90000 {
		if _, err = conn.ExecContext(ctx, "SET LOCAL bytea_output='escape'"); err != nil {
			return nil, err
		}
	}
	// 显式限定 search_path，避免默认表达式恢复时误引用其他模式的同名对象。
	if _, err = conn.ExecContext(ctx, "SET LOCAL search_path=pg_catalog"); err != nil {
		return nil, err
	}
	if err = write("SET search_path=pg_catalog"); err != nil {
		return nil, err
	}
	schemas, err := records("SELECT n.nspname AS name FROM pg_namespace n WHERE "+selected+" ORDER BY n.nspname", dependencyArgs...)
	if err != nil {
		return nil, err
	}
	for _, item := range schemas {
		quoted, e := m.quote(fmt.Sprint(item["name"]))
		if e != nil {
			return nil, e
		}
		if version >= 90300 {
			err = write("CREATE SCHEMA IF NOT EXISTS " + quoted)
		} else if quoted != `"public"` {
			err = write("CREATE SCHEMA " + quoted)
		}
		if err != nil {
			return nil, err
		}
	}
	if table == "" && version >= 90100 {
		extensions, e := records("SELECT e.extname AS name,n.nspname AS schema FROM pg_extension e JOIN pg_namespace n ON n.oid=e.extnamespace WHERE e.extname<>'plpgsql' AND "+selected+" ORDER BY e.extname", dependencyArgs...)
		if e != nil {
			return nil, e
		}
		for _, item := range extensions {
			name, _ := m.quote(fmt.Sprint(item["name"]))
			ns, _ := m.quote(fmt.Sprint(item["schema"]))
			if err = write("CREATE EXTENSION IF NOT EXISTS " + name + " WITH SCHEMA " + ns); err != nil {
				return nil, err
			}
		}
	}
	// ENUM 与 DOMAIN 必须在引用这些类型的数据表之前创建。
	typeExclusion, routineExclusion := "", ""
	if version >= 90100 {
		typeExclusion = " AND NOT EXISTS(SELECT 1 FROM pg_depend d WHERE d.classid='pg_type'::regclass AND d.objid=t.oid AND d.deptype='e')"
		routineExclusion = " AND NOT EXISTS(SELECT 1 FROM pg_depend d WHERE d.classid='pg_proc'::regclass AND d.objid=p.oid AND d.deptype='e')"
	}
	types, e := records("SELECT t.oid::bigint AS oid,n.nspname AS schema,t.typname AS name,t.typtype AS kind,format_type(t.typbasetype,t.typtypmod) AS base,t.typnotnull AS not_null,t.typdefault AS default_value FROM pg_type t JOIN pg_namespace n ON n.oid=t.typnamespace WHERE t.typtype IN ('e','d') AND "+selected+typeExclusion+" ORDER BY CASE WHEN t.typtype='e' THEN 0 ELSE 1 END,n.nspname,t.typname", dependencyArgs...)
	if e != nil {
		return nil, e
	}
	for _, item := range types {
		ns, _ := m.quote(fmt.Sprint(item["schema"]))
		name, _ := m.quote(fmt.Sprint(item["name"]))
		target := ns + "." + name
		if item["kind"] == "e" {
			order := "oid"
			if version >= 90100 {
				order = "enumsortorder"
			}
			labels, e := records("SELECT enumlabel AS label FROM pg_enum WHERE enumtypid=$1::oid ORDER BY "+order, item["oid"])
			if e != nil {
				return nil, e
			}
			values := []string{}
			for _, label := range labels {
				values = append(values, literal(fmt.Sprint(label["label"])))
			}
			err = write("CREATE TYPE " + target + " AS ENUM (" + strings.Join(values, ",") + ")")
		} else {
			definition := "CREATE DOMAIN " + target + " AS " + fmt.Sprint(item["base"])
			if item["default_value"] != nil {
				definition += " DEFAULT " + fmt.Sprint(item["default_value"])
			}
			if m.truth(item["not_null"]) {
				definition += " NOT NULL"
			}
			constraints, e := records("SELECT pg_get_constraintdef(oid,true) AS definition FROM pg_constraint WHERE contypid=$1::oid ORDER BY conname", item["oid"])
			if e != nil {
				return nil, e
			}
			for _, constraint := range constraints {
				definition += " " + fmt.Sprint(constraint["definition"])
			}
			err = write(definition)
		}
		if err != nil {
			return nil, err
		}
	}
	// serial 与独立序列预先创建；identity 的序列由 CREATE TABLE 创建。
	sequenceWhere := ""
	if table != "" {
		oid := strconv.FormatInt(objects[0]["oid"].(int64), 10)
		sequenceWhere = " AND (EXISTS(SELECT 1 FROM pg_depend d WHERE d.classid='pg_class'::regclass AND d.objid=c.oid AND d.refobjid=" + oid + " AND d.deptype IN ('a','i')) OR EXISTS(SELECT 1 FROM pg_depend d JOIN pg_attrdef a ON a.oid=d.objid WHERE d.classid='pg_attrdef'::regclass AND d.refclassid='pg_class'::regclass AND d.refobjid=c.oid AND a.adrelid=" + oid + "))"
	}
	identityExclusion := ""
	if version >= 100000 {
		identityExclusion = " AND NOT EXISTS (SELECT 1 FROM pg_depend d WHERE d.classid='pg_class'::regclass AND d.objid=c.oid AND d.deptype='i')"
	}
	sequences, e := records("SELECT c.oid::bigint AS oid,n.nspname AS schema,c.relname AS name FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE c.relkind='S' AND "+selected+sequenceWhere+identityExclusion+" ORDER BY n.nspname,c.relname", dependencyArgs...)
	if e != nil {
		return nil, e
	}
	for _, item := range sequences {
		ns, _ := m.quote(fmt.Sprint(item["schema"]))
		name, _ := m.quote(fmt.Sprint(item["name"]))
		target := ns + "." + name
		query := "SELECT start_value AS start,increment_by AS increment,min_value AS min,max_value AS max,cache_value AS cache,is_cycled AS cycle FROM " + target
		if version >= 100000 {
			query = "SELECT seqstart AS start,seqincrement AS increment,seqmin AS min,seqmax AS max,seqcache AS cache,seqcycle AS cycle FROM pg_sequence WHERE seqrelid=" + strconv.FormatInt(item["oid"].(int64), 10)
		}
		options, e := records(query)
		if e != nil {
			return nil, e
		}
		if len(options) != 1 {
			return nil, errors.New("序列参数不存在")
		}
		option := options[0]
		cycle := "NO CYCLE"
		if m.truth(option["cycle"]) {
			cycle = "CYCLE"
		}
		if err = write("CREATE SEQUENCE " + target + " INCREMENT BY " + fmt.Sprint(option["increment"]) + " MINVALUE " + fmt.Sprint(option["min"]) + " MAXVALUE " + fmt.Sprint(option["max"]) + " START WITH " + fmt.Sprint(option["start"]) + " CACHE " + fmt.Sprint(option["cache"]) + " " + cycle); err != nil {
			return nil, err
		}
	}
	routineWhere := "NOT p.proisagg"
	if version >= 110000 {
		routineWhere = "p.prokind IN ('f','p')"
	}
	routines, e := records("SELECT pg_get_functiondef(p.oid) AS definition,EXISTS(SELECT 1 FROM pg_type rt JOIN pg_class rc ON rc.oid=rt.typrelid WHERE rc.relkind IN ('r','p') AND (rt.oid=p.prorettype OR rt.oid=ANY(COALESCE(p.proallargtypes,p.proargtypes::oid[])))) AS table_dependency FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace WHERE "+routineWhere+" AND "+selected+routineExclusion+" ORDER BY n.nspname,p.proname,p.oid", dependencyArgs...)
	if e != nil {
		return nil, e
	}
	for _, routine := range routines {
		if m.truth(routine["table_dependency"]) {
			continue
		}
		if err = write(fmt.Sprint(routine["definition"])); err != nil {
			return nil, err
		}
	}
	for index, item := range objects {
		kind := fmt.Sprint(item["kind"])
		if kind == "v" || kind == "m" {
			continue
		}
		definition, e := m.DDL(ctx, fmt.Sprint(item["schema"]), fmt.Sprint(item["name"]), conn)
		if e != nil {
			return nil, e
		}
		if err = write(definition); err != nil {
			return nil, err
		}
		report(10+10*float64(index+1)/float64(len(objects)), "写入数据表结构")
	}
	// 返回表行类型或以行类型为参数的程序必须等表存在后创建。
	for _, routine := range routines {
		if m.truth(routine["table_dependency"]) {
			if err = write(fmt.Sprint(routine["definition"])); err != nil {
				return nil, err
			}
		}
	}
	// 引用程序的默认值后置；COPY 显式写入每个普通字段，不会调用这些默认值。
	for _, item := range objects {
		if item["kind"] == "v" || item["kind"] == "m" {
			continue
		}
		_, target, _ := m.target(fmt.Sprint(item["schema"]), fmt.Sprint(item["name"]))
		generated := ""
		if version >= 120000 {
			generated = " AND a.attgenerated=''"
		}
		defaults, e := records("SELECT quote_ident(a.attname) AS name,pg_get_expr(d.adbin,d.adrelid) AS definition FROM pg_attribute a JOIN pg_attrdef d ON d.adrelid=a.attrelid AND d.adnum=a.attnum WHERE a.attrelid=$1::oid"+generated+" AND EXISTS(SELECT 1 FROM pg_depend fd WHERE fd.classid='pg_attrdef'::regclass AND fd.objid=d.oid AND fd.refclassid='pg_proc'::regclass) ORDER BY a.attnum", item["oid"])
		if e != nil {
			return nil, e
		}
		for _, column := range defaults {
			if err = write("ALTER TABLE " + target + " ALTER COLUMN " + fmt.Sprint(column["name"]) + " SET DEFAULT " + fmt.Sprint(column["definition"])); err != nil {
				return nil, err
			}
		}
	}
	// COPY 使用服务端类型的输出/输入函数，保留 bytea、数组、大数与时间精度。
	completedWeight := float64(0)
	report(20, "导出数据")
	for _, item := range objects {
		kind := fmt.Sprint(item["kind"])
		if kind == "v" || kind == "m" || kind == "p" {
			continue
		}
		_, target, e := m.target(fmt.Sprint(item["schema"]), fmt.Sprint(item["name"]))
		if e != nil {
			return nil, e
		}
		generated := ""
		if version >= 120000 {
			generated = " AND a.attgenerated=''"
		}
		columns, e := records("SELECT quote_ident(a.attname) AS name FROM pg_attribute a WHERE a.attrelid=$1::oid AND a.attnum>0 AND NOT a.attisdropped"+generated+" ORDER BY a.attnum", item["oid"])
		if e != nil {
			return nil, e
		}
		names := []string{}
		for _, column := range columns {
			names = append(names, fmt.Sprint(column["name"]))
		}
		if len(names) == 0 {
			return nil, errors.New("表没有可导出的普通字段")
		}
		if _, err = writer.WriteString("COPY " + target + " (" + strings.Join(names, ",") + ") FROM stdin;\n"); err != nil {
			return nil, err
		}
		statements++
		var copied int64
		var writtenRows int64
		rowBytes := 0
		output := copyOutput(func(data []byte) (int, error) {
			if e := ctx.Err(); e != nil {
				return 0, e
			}
			for offset := 0; offset < len(data); {
				newline := bytes.IndexByte(data[offset:], '\n')
				end := len(data)
				if newline >= 0 {
					end = offset + newline + 1
				}
				rowBytes += end - offset
				if rowBytes > 64<<20 {
					return 0, errors.New("COPY 单行超过 64 MB，请使用数据库原生备份工具处理超大字段")
				}
				if newline >= 0 {
					rowBytes = 0
				}
				offset = end
			}
			n, e := writer.Write(data)
			if len(progress) > 0 && progress[0] != nil {
				writtenRows += int64(bytes.Count(data[:n], []byte{'\n'}))
				weight := min(float64(writtenRows), float64(rowCounts[item["oid"].(int64)]))
				report(20+65*(completedWeight+weight)/max(totalWeight, 1), "导出数据 "+target)
			}
			return n, e
		})
		err = conn.Raw(func(raw interface{}) error {
			native, ok := raw.(*stdlib.Conn)
			if !ok {
				return errors.New("Postgres 连接不支持原生 COPY")
			}
			tag, e := native.Conn().PgConn().CopyTo(ctx, output, "COPY (SELECT "+strings.Join(names, ",")+" FROM ONLY "+target+") TO STDOUT")
			copied = tag.RowsAffected()
			return e
		})
		if err != nil {
			return nil, err
		}
		count += copied
		if _, err = writer.WriteString("\\.\n\n"); err != nil {
			return nil, err
		}
		completedWeight += float64(max(rowCounts[item["oid"].(int64)], 1))
		report(20+65*completedWeight/max(totalWeight, 1), "导出数据 "+target)
	}
	// 数据加载完毕后再恢复外键和触发器，防止循环引用及重复业务副作用。
	report(85, "写入索引与关联对象")
	for index, item := range objects {
		kind := fmt.Sprint(item["kind"])
		if kind == "v" || kind == "m" {
			definition, e := m.DDL(ctx, fmt.Sprint(item["schema"]), fmt.Sprint(item["name"]), conn)
			if e != nil {
				return nil, e
			}
			if kind == "m" {
				definition = strings.TrimSuffix(strings.TrimSpace(definition), ";") + " WITH NO DATA"
			}
			if err = write(definition); err != nil {
				return nil, err
			}
		}
		_, target, _ := m.target(fmt.Sprint(item["schema"]), fmt.Sprint(item["name"]))
		if kind != "v" {
			indexes, e := records("SELECT pg_get_indexdef(x.indexrelid) AS definition FROM pg_index x WHERE x.indrelid=$1::oid AND NOT EXISTS(SELECT 1 FROM pg_constraint p WHERE p.conindid=x.indexrelid AND p.contype IN ('p','u','x')) ORDER BY x.indexrelid", item["oid"])
			if e != nil {
				return nil, e
			}
			for _, index := range indexes {
				if err = write(fmt.Sprint(index["definition"])); err != nil {
					return nil, err
				}
			}
			constraints, e := records("SELECT quote_ident(conname) AS name,pg_get_constraintdef(oid,true) AS definition FROM pg_constraint WHERE conrelid=$1::oid AND contype IN ('f','c') ORDER BY conname", item["oid"])
			if e != nil {
				return nil, e
			}
			for _, constraint := range constraints {
				if err = write("ALTER TABLE " + target + " ADD CONSTRAINT " + fmt.Sprint(constraint["name"]) + " " + fmt.Sprint(constraint["definition"])); err != nil {
					return nil, err
				}
			}
		}
		visible, definition := "NOT tgisinternal", "pg_get_triggerdef(oid,true)"
		if version < 90000 {
			visible, definition = "tgconstraint=0", "pg_get_triggerdef(oid)"
		}
		triggers, e := records("SELECT "+definition+" AS definition,tgenabled AS enabled,quote_ident(tgname) AS name FROM pg_trigger WHERE tgrelid=$1::oid AND "+visible+" ORDER BY tgname", item["oid"])
		if e != nil {
			return nil, e
		}
		for _, trigger := range triggers {
			if err = write(fmt.Sprint(trigger["definition"])); err != nil {
				return nil, err
			}
			mode := ""
			switch trigger["enabled"] {
			case "D":
				mode = "DISABLE"
			case "R":
				mode = "ENABLE REPLICA"
			case "A":
				mode = "ENABLE ALWAYS"
			}
			if mode != "" {
				if err = write("ALTER TABLE " + target + " " + mode + " TRIGGER " + fmt.Sprint(trigger["name"])); err != nil {
					return nil, err
				}
			}
		}
		report(85+9*float64(index+1)/float64(len(objects)), "写入索引与关联对象")
	}
	// 同步 serial/identity 的当前位置；不采用 MAX(id)，保留空表与已耗用序列。
	allSequences, e := records("SELECT c.oid::bigint AS oid,n.nspname AS schema,c.relname AS name FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE c.relkind='S' AND "+selected+sequenceWhere+" ORDER BY n.nspname,c.relname", dependencyArgs...)
	if e != nil {
		return nil, e
	}
	for index, item := range allSequences {
		ns, _ := m.quote(fmt.Sprint(item["schema"]))
		name, _ := m.quote(fmt.Sprint(item["name"]))
		target := ns + "." + name
		owners, e := records("SELECT quote_ident(n.nspname)||'.'||quote_ident(c.relname)||'.'||quote_ident(a.attname) AS owner,d.deptype AS kind FROM pg_depend d JOIN pg_class c ON c.oid=d.refobjid JOIN pg_namespace n ON n.oid=c.relnamespace JOIN pg_attribute a ON a.attrelid=c.oid AND a.attnum=d.refobjsubid WHERE d.classid='pg_class'::regclass AND d.objid=$1::oid AND d.refclassid='pg_class'::regclass AND d.deptype IN ('a','i')", item["oid"])
		if e != nil {
			return nil, e
		}
		if len(owners) > 0 && owners[0]["kind"] == "a" {
			if err = write("ALTER SEQUENCE " + target + " OWNED BY " + fmt.Sprint(owners[0]["owner"])); err != nil {
				return nil, err
			}
		}
		if version >= 100000 {
			options, e := records("SELECT seqincrement AS increment,seqmin AS min,seqmax AS max,seqstart AS start,seqcache AS cache,seqcycle AS cycle,format_type(seqtypid,NULL) AS type FROM pg_sequence WHERE seqrelid=$1::oid", item["oid"])
			if e != nil {
				return nil, e
			}
			if len(options) != 1 {
				return nil, errors.New("序列参数不存在")
			}
			option := options[0]
			cycle := "NO CYCLE"
			if m.truth(option["cycle"]) {
				cycle = "CYCLE"
			}
			if err = write("ALTER SEQUENCE " + target + " AS " + fmt.Sprint(option["type"]) + " INCREMENT BY " + fmt.Sprint(option["increment"]) + " MINVALUE " + fmt.Sprint(option["min"]) + " MAXVALUE " + fmt.Sprint(option["max"]) + " START WITH " + fmt.Sprint(option["start"]) + " CACHE " + fmt.Sprint(option["cache"]) + " " + cycle); err != nil {
				return nil, err
			}
		}
		var last int64
		var called bool
		if err = conn.QueryRowContext(ctx, "SELECT last_value,is_called FROM "+target).Scan(&last, &called); err != nil {
			return nil, err
		}
		if err = write("SELECT pg_catalog.setval(" + literal(target) + ", " + strconv.FormatInt(last, 10) + ", " + strconv.FormatBool(called) + ")"); err != nil {
			return nil, err
		}
		report(94+2*float64(index+1)/float64(len(allSequences)), "写入序列状态")
	}
	for _, item := range objects {
		if item["kind"] == "m" {
			_, target, _ := m.target(fmt.Sprint(item["schema"]), fmt.Sprint(item["name"]))
			if err = write("REFRESH MATERIALIZED VIEW " + target); err != nil {
				return nil, err
			}
		}
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
	stat, e := file.Stat()
	if e != nil {
		return nil, e
	}
	if err = file.Close(); err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
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
