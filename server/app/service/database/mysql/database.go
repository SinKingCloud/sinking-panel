package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

// Charsets 读取远端实际支持的字符集和排序规则，供创建、修改数据库共用。
func (m *Manager) Charsets(ctx context.Context) (_ *CharsetOptions, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	data := &CharsetOptions{Charsets: []Charset{}}
	if err = m.db.QueryRowContext(ctx, "SELECT @@character_set_server, @@collation_server").Scan(&data.Charset, &data.Collation); err != nil {
		return nil, err
	}
	// 新版 MariaDB 的通用 UCA 排序规则需要完整名称，MySQL 和旧版保留原字段。
	collationField := "c.COLLATION_NAME"
	if strings.Contains(strings.ToLower(m.version), "mariadb") {
		collationField = "c.FULL_COLLATION_NAME"
	}
	query := "SELECT s.CHARACTER_SET_NAME, s.DEFAULT_COLLATE_NAME, %s FROM information_schema.CHARACTER_SETS s LEFT JOIN information_schema.COLLATION_CHARACTER_SET_APPLICABILITY c ON c.CHARACTER_SET_NAME = s.CHARACTER_SET_NAME ORDER BY s.CHARACTER_SET_NAME, %s"
	rows, err := m.db.QueryContext(ctx, fmt.Sprintf(query, collationField, collationField))
	var serverError *mysqlDriver.MySQLError
	if collationField != "c.COLLATION_NAME" && errors.As(err, &serverError) && serverError.Number == 1054 {
		rows, err = m.db.QueryContext(ctx, fmt.Sprintf(query, "c.COLLATION_NAME", "c.COLLATION_NAME"))
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name, defaultCollation string
		var collation sql.NullString
		if err = rows.Scan(&name, &defaultCollation, &collation); err != nil {
			return nil, err
		}
		if len(data.Charsets) == 0 || data.Charsets[len(data.Charsets)-1].Name != name {
			data.Charsets = append(data.Charsets, Charset{Name: name, DefaultCollation: defaultCollation, Collations: []string{}})
		}
		if collation.Valid && collation.String != "" {
			charset := &data.Charsets[len(data.Charsets)-1]
			charset.Collations = append(charset.Collations, collation.String)
		}
	}
	return data, rows.Err()
}

// DatabaseInfo 读取真实数据库的字符集、排序规则和表空间信息。
func (m *Manager) DatabaseInfo(ctx context.Context, name string) (_ map[string]interface{}, err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if _, err = m.quote(name); err != nil || strings.TrimSpace(name) == "" || utf8.RuneCountInString(name) > 64 || strings.ContainsAny(name, "/\\.") || strings.HasSuffix(name, " ") {
		return nil, errors.New("数据库名称不合法，最多允许 64 个字符")
	}
	var databaseName, charset, collation string
	err = m.db.QueryRowContext(ctx, "SELECT SCHEMA_NAME, DEFAULT_CHARACTER_SET_NAME, DEFAULT_COLLATION_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?", name).Scan(&databaseName, &charset, &collation)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("数据库不存在或当前账户无权访问")
	}
	if err != nil {
		return nil, err
	}
	var count, size int64
	if err = m.db.QueryRowContext(ctx, "SELECT COUNT(*), COALESCE(SUM(DATA_LENGTH + INDEX_LENGTH), 0) FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'", databaseName).Scan(&count, &size); err != nil {
		return nil, err
	}
	return map[string]interface{}{"name": databaseName, "charset": charset, "collation": collation, "table_count": count, "size": size}, nil
}

// CreateDatabase 创建真实数据库；字符集和排序规则使用服务器支持的名称。
func (m *Manager) CreateDatabase(ctx context.Context, name, charset, collation string) (err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	name = strings.TrimSpace(name)
	databaseName, err := m.quote(name)
	if err != nil || utf8.RuneCountInString(name) > 64 || strings.ContainsAny(name, "/\\.") || strings.HasSuffix(name, " ") {
		return errors.New("数据库名称不合法，最多允许 64 个字符")
	}
	switch strings.ToLower(name) {
	case "mysql", "information_schema", "performance_schema", "sys":
		return errors.New("不能创建系统数据库")
	}
	statement := "CREATE DATABASE " + databaseName
	if charset != "" || collation != "" {
		options, optionErr := m.Charsets(ctx)
		if optionErr != nil {
			return optionErr
		}
		var supportedCharset, supportedCollation, collationCharset string
		for _, option := range options.Charsets {
			if strings.EqualFold(option.Name, charset) {
				supportedCharset = option.Name
			}
			for _, name := range option.Collations {
				if strings.EqualFold(name, collation) {
					supportedCollation, collationCharset = name, option.Name
					break
				}
			}
		}
		if charset != "" && supportedCharset == "" {
			return errors.New("当前服务器不支持此字符集")
		}
		if collation != "" && supportedCollation == "" {
			return errors.New("当前服务器不支持此排序规则")
		}
		if supportedCharset != "" && supportedCollation != "" && !strings.EqualFold(supportedCharset, collationCharset) {
			return errors.New("字符集与排序规则不匹配")
		}
		if supportedCharset != "" {
			quoted, quoteErr := m.quote(supportedCharset)
			if quoteErr != nil {
				return quoteErr
			}
			statement += " CHARACTER SET " + quoted
		}
		if supportedCollation != "" {
			quoted, quoteErr := m.quote(supportedCollation)
			if quoteErr != nil {
				return quoteErr
			}
			statement += " COLLATE " + quoted
		}
	}
	_, err = m.db.ExecContext(ctx, statement)
	return err
}

// UpdateDatabase 修改真实数据库的默认字符集和排序规则，不移动或重命名数据库。
func (m *Manager) UpdateDatabase(ctx context.Context, name, charset, collation string) (err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	databaseName, err := m.quote(name)
	if err != nil || strings.TrimSpace(name) == "" || utf8.RuneCountInString(name) > 64 || strings.ContainsAny(name, "/\\.") || strings.HasSuffix(name, " ") {
		return errors.New("数据库名称不合法，最多允许 64 个字符")
	}
	switch strings.ToLower(name) {
	case "mysql", "information_schema", "performance_schema", "sys":
		return errors.New("不能修改系统数据库")
	}
	if charset == "" && collation == "" {
		_, err = m.DatabaseInfo(ctx, name)
		return err
	}
	statement := "ALTER DATABASE " + databaseName
	if charset != "" || collation != "" {
		options, optionErr := m.Charsets(ctx)
		if optionErr != nil {
			return optionErr
		}
		var supportedCharset, supportedCollation, collationCharset string
		for _, option := range options.Charsets {
			if strings.EqualFold(option.Name, charset) {
				supportedCharset = option.Name
			}
			for _, name := range option.Collations {
				if strings.EqualFold(name, collation) {
					supportedCollation, collationCharset = name, option.Name
					break
				}
			}
		}
		if charset != "" && supportedCharset == "" {
			return errors.New("当前服务器不支持此字符集")
		}
		if collation != "" && supportedCollation == "" {
			return errors.New("当前服务器不支持此排序规则")
		}
		if supportedCharset != "" && supportedCollation != "" && !strings.EqualFold(supportedCharset, collationCharset) {
			return errors.New("字符集与排序规则不匹配")
		}
		if supportedCharset != "" {
			quoted, quoteErr := m.quote(supportedCharset)
			if quoteErr != nil {
				return quoteErr
			}
			statement += " CHARACTER SET " + quoted
		}
		if supportedCollation != "" {
			quoted, quoteErr := m.quote(supportedCollation)
			if quoteErr != nil {
				return quoteErr
			}
			statement += " COLLATE " + quoted
		}
	}
	_, err = m.db.ExecContext(ctx, statement)
	return err
}

// DropDatabase 删除真实业务数据库；已不存在的数据库无需重复删除。
func (m *Manager) DropDatabase(ctx context.Context, name string) (err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	databaseName, err := m.dropDatabaseTarget(name)
	if err != nil {
		return err
	}
	_, err = m.db.ExecContext(ctx, "DROP DATABASE IF EXISTS "+databaseName)
	return err
}

// DropDatabases 在发送任何删除前检查整批目标，并报告失败前已完成的数量。
func (m *Manager) DropDatabases(ctx context.Context, names []string) (err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if len(names) == 0 {
		return errors.New("请选择数据库")
	}
	for _, name := range names {
		if _, err = m.dropDatabaseTarget(name); err != nil {
			return err
		}
	}
	for completed, name := range names {
		if err = m.DropDatabase(ctx, name); err != nil {
			return fmt.Errorf("已删除 %d 个数据库，删除 %s 失败：%w", completed, name, err)
		}
	}
	return nil
}

func (m *Manager) dropDatabaseTarget(name string) (string, error) {
	databaseName, err := m.quote(name)
	if err != nil || strings.TrimSpace(name) == "" || utf8.RuneCountInString(name) > 64 || strings.ContainsAny(name, "/\\.") || strings.HasSuffix(name, " ") {
		return "", errors.New("数据库名称不合法，最多允许 64 个字符")
	}
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "mysql", "information_schema", "performance_schema", "sys":
		return "", errors.New("不能删除系统数据库")
	}
	return databaseName, nil
}
