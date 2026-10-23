package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"server/app/util/page"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// Manager 提供 Postgres 连接、数据、账户和维护操作。
type Manager struct {
	db       *sql.DB
	database string
	config   string // 复用当前认证配置访问同一实例中的其他数据库。
	secrets  []string
	version  int // 连接时读取版本，结构查询只使用该版本存在的系统目录字段。
}

// NewManager 打开 Postgres 连接，调用方使用 Close 释放连接池。
func NewManager(ctx context.Context, config, databaseName string) (*Manager, error) {
	if len(config) > 1<<20 {
		return nil, errors.New("连接配置超过 1 MB")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var connection Config
	config = strings.TrimSpace(config)
	decoder := json.NewDecoder(strings.NewReader(config))
	decoder.DisallowUnknownFields()
	if len(config) == 0 || config[0] != '{' || !json.Valid([]byte(config)) || decoder.Decode(&connection) != nil {
		return nil, errors.New("连接配置必须是 JSON 对象")
	}
	if connection.Host == "" {
		connection.Host = "127.0.0.1"
	}
	if connection.Port == 0 {
		connection.Port = 5432
	}
	if connection.Port < 1 || connection.Port > 65535 {
		return nil, errors.New("连接端口不合法")
	}
	if databaseName == "" {
		databaseName = connection.Database
	}
	if databaseName == "" {
		databaseName = "postgres"
	}
	m := &Manager{config: config, secrets: []string{connection.Password}}
	u := &url.URL{Scheme: "postgres", Host: net.JoinHostPort(connection.Host, strconv.Itoa(connection.Port)), Path: "/" + databaseName, User: url.UserPassword(connection.User, connection.Password)}
	q := u.Query()
	mode := connection.SSLMode
	if mode == "" {
		mode = "prefer"
	}
	q.Set("sslmode", mode)
	for key, value := range map[string]string{"sslrootcert": connection.SSLRootCert, "sslcert": connection.SSLCert, "sslkey": connection.SSLKey} {
		if value != "" {
			q.Set(key, value)
		}
	}
	u.RawQuery = q.Encode()
	cfg, err := pgx.ParseConfig(u.String())
	if err != nil {
		return nil, errors.New("Postgres 连接配置不合法")
	}
	m.secrets = append(m.secrets, cfg.Password)
	cfg.ConnectTimeout = 5 * time.Second
	// Go 字符串和客户端 SQL 均为 UTF-8，避免非 UTF-8 数据库的文字与备份损坏。
	cfg.RuntimeParams["client_encoding"] = "UTF8"
	// 根据服务端字段类型绑定高精度数值，不依赖缓存的命名预编译语句。
	cfg.DefaultQueryExecMode = pgx.QueryExecModeDescribeExec
	m.database = cfg.Database
	m.db = stdlib.OpenDB(*cfg)
	m.db.SetMaxOpenConns(2)
	m.db.SetMaxIdleConns(1)
	m.db.SetConnMaxLifetime(time.Minute)
	if err = m.db.PingContext(ctx); err != nil {
		_ = m.db.Close()
		return nil, m.safeError(err)
	}
	if err = m.db.QueryRowContext(ctx, "SHOW server_version_num").Scan(&m.version); err != nil {
		_ = m.db.Close()
		return nil, m.safeError(err)
	}
	return m, nil
}

// Close 释放连接池。
func (m *Manager) Close() error { return m.db.Close() }

func (m *Manager) safeError(err error) error {
	if err == nil {
		return nil
	}
	message := err.Error()
	for _, secret := range m.secrets {
		if secret != "" {
			message = strings.ReplaceAll(message, secret, "***")
			message = strings.ReplaceAll(message, url.QueryEscape(secret), "***")
			message = strings.ReplaceAll(message, url.PathEscape(secret), "***")
		}
	}
	return errors.New(message)
}

// Test 返回 Postgres 服务端实际版本，不暴露认证信息。
func (m *Manager) Test(ctx context.Context) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var version string
	if err := m.db.QueryRowContext(ctx, "SELECT version()").Scan(&version); err != nil {
		return nil, m.safeError(err)
	}
	return map[string]interface{}{"version": version, "database": m.database}, nil
}

// Databases 按名称搜索并分页返回当前账户可访问的数据库，默认隐藏模板库。
func (m *Manager) Databases(ctx context.Context, keyword string, pageNumber, pageSize int, showSystem bool, orderByType string) (_ *page.Result[string], err error) {
	defer func() { err = m.safeError(err) }()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if pageNumber == 0 {
		pageNumber = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}
	if pageNumber < 1 || pageNumber > 1000000 || pageSize < 1 || pageSize > 1000 {
		return nil, errors.New("页码应为 1 至 1000000，每页数量应为 1 至 1000")
	}
	if len(keyword) > 4096 {
		return nil, errors.New("数据库名称关键词过长")
	}
	if orderByType == "" {
		orderByType = "asc"
	}
	if orderByType != "asc" && orderByType != "desc" {
		return nil, errors.New("排序类型不合法")
	}

	source := " FROM pg_database WHERE datallowconn AND has_database_privilege(datname, 'CONNECT')"
	if !showSystem {
		source += " AND NOT datistemplate"
	}
	args := make([]interface{}, 0, 3)
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		source += " AND POSITION(LOWER($1) IN LOWER(datname)) > 0"
		args = append(args, keyword)
	}
	var total int64
	if err = m.db.QueryRowContext(ctx, "SELECT COUNT(*)"+source, args...).Scan(&total); err != nil {
		return nil, err
	}

	if total == 0 {
		pageNumber = 1
	} else if lastPage := (total-1)/int64(pageSize) + 1; int64(pageNumber) > lastPage {
		pageNumber = int(lastPage)
	}

	query := "SELECT datname" + source + " ORDER BY datname " + strings.ToUpper(orderByType) + ", oid LIMIT " + m.placeholder(len(args)+1) + " OFFSET " + m.placeholder(len(args)+2)
	args = append(args, pageSize, int64(pageNumber-1)*int64(pageSize))
	rows, err := m.db.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]string, 0)
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, err
		}
		result = append(result, name)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return page.New(&page.Query{Page: pageNumber, PageSize: pageSize}, total, result), nil
}

func (m *Manager) quote(name string) (string, error) {
	if name == "" || strings.ContainsRune(name, 0) || len(name) > 1024 {
		return "", errors.New("表名或字段名不合法")
	}
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`, nil
}

func (m *Manager) placeholder(position int) string { return "$" + strconv.Itoa(position) }

func (m *Manager) target(schemaName, tableName string) (string, string, error) {
	schema := schemaName
	if schema == "" {
		schema = "public"
	}
	table, err := m.quote(tableName)
	if err != nil {
		return "", "", err
	}
	name, err := m.quote(schema)
	if err != nil {
		return "", "", err
	}
	return schema, name + "." + table, nil
}

// serverVersion 兼容直接构建 Manager 的调用方；正常连接只在初始化时读取一次。
func (m *Manager) serverVersion(ctx context.Context) (int, error) {
	if m.version != 0 {
		return m.version, nil
	}
	var version int
	err := m.db.QueryRowContext(ctx, "SHOW server_version_num").Scan(&version)
	return version, err
}

// routineCatalog PostgreSQL 11 用 prokind 替代 proisagg；不依赖 JSON 系统目录转换。
func (m *Manager) routineCatalog(ctx context.Context) (string, string, error) {
	version, err := m.serverVersion(ctx)
	if err != nil {
		return "", "", err
	}
	if version >= 110000 {
		return "CASE WHEN p.prokind = 'p' THEN 'PROCEDURE' ELSE 'FUNCTION' END", "p.prokind IN ('f','p')", nil
	}
	return "'FUNCTION'", "NOT p.proisagg", nil
}
