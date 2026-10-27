package mysql

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"os"
	"regexp"
	"server/app/util/page"
	"strconv"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var serverVersionPattern = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)

// Manager 管理一次请求使用的 MySQL 连接，不在连接池间切换数据库。
type Manager struct {
	db       *sql.DB
	database string
	secrets  []string
	version  string // 服务器原始版本，避免每个字段重复读取。
}

// NewManager 打开已有 MySQL 数据库，调用方负责关闭；连接与执行均受请求 context 控制。
func NewManager(ctx context.Context, config string, databaseName string) (*Manager, error) {
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
		connection.Port = 3306
	}
	if connection.Port < 1 || connection.Port > 65535 {
		return nil, errors.New("连接端口不合法")
	}
	if databaseName == "" {
		databaseName = connection.Database
	}
	m := &Manager{
		database: databaseName,
		secrets:  []string{connection.Password},
	}
	// 多字符集仅由内部固定参数配置，依次协商 utf8mb4、utf8。
	cfg, err := mysqlDriver.ParseDSN("/?charset=utf8mb4,utf8")
	if err != nil {
		return nil, errors.New("MySQL 字符集配置不合法")
	}
	cfg.User, cfg.Passwd = connection.User, connection.Password
	cfg.Net, cfg.Addr = "tcp", net.JoinHostPort(connection.Host, strconv.Itoa(connection.Port))
	if connection.Socket != "" {
		cfg.Net, cfg.Addr = "unix", connection.Socket
	}
	cfg.DBName = databaseName
	// 执行由 context 控制，避免大型 SQL 恢复被固定读写超时提前中断。
	cfg.Timeout, cfg.ReadTimeout, cfg.WriteTimeout = 5*time.Second, 0, 0
	cfg.ClientFoundRows, cfg.MultiStatements, cfg.AllowAllFiles = true, false, false
	if mode := strings.ToLower(connection.TLS); mode != "" && mode != "false" && mode != "disable" {
		cfg.TLS = &tls.Config{
			ServerName: connection.Host,
			MinVersion: tls.VersionTLS12,
		}
		// 与 MySQL 客户端的 preferred 一致；指定证书后必须使用 TLS 并验证服务器。
		cfg.AllowFallbackToPlaintext = mode == "preferred" && connection.SSLCA == "" && connection.SSLCert == "" && connection.SSLKey == ""
		cfg.TLS.InsecureSkipVerify = mode == "skip-verify" || mode == "require" || cfg.AllowFallbackToPlaintext
		if path := connection.SSLCA; path != "" {
			pem, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil, errors.New("无法读取 TLS CA 文件")
			}
			cfg.TLS.RootCAs = x509.NewCertPool()
			if !cfg.TLS.RootCAs.AppendCertsFromPEM(pem) {
				return nil, errors.New("TLS CA 文件格式不正确")
			}
		}
		if connection.SSLCert != "" || connection.SSLKey != "" {
			if connection.SSLCert == "" || connection.SSLKey == "" {
				return nil, errors.New("客户端证书与私钥必须同时填写")
			}
			certificate, loadErr := tls.LoadX509KeyPair(connection.SSLCert, connection.SSLKey)
			if loadErr != nil {
				return nil, errors.New("无法读取 TLS 客户端证书")
			}
			cfg.TLS.Certificates = []tls.Certificate{certificate}
		}
	}
	open := func(connection *mysqlDriver.Config) (*sql.DB, error) {
		connector, connectorErr := mysqlDriver.NewConnector(connection)
		if connectorErr != nil {
			return nil, errors.New("MySQL 连接配置不合法")
		}
		db := sql.OpenDB(connector)
		db.SetMaxOpenConns(2)
		db.SetMaxIdleConns(1)
		db.SetConnMaxLifetime(time.Minute)
		if pingErr := db.PingContext(ctx); pingErr != nil {
			_ = db.Close()
			return nil, pingErr
		}
		return db, nil
	}
	m.db, err = open(cfg)
	var serverError *mysqlDriver.MySQLError
	// MySQL 5.0/5.1 在握手时尚不认识 utf8mb4。仅收到字符集错误才降级，
	// 不重试认证、网络或 TLS 错误。
	if errors.As(err, &serverError) && (serverError.Number == 1115 || serverError.Number == 1273) {
		legacy := cfg.Clone()
		if applyErr := legacy.Apply(mysqlDriver.Charset("utf8", "utf8_general_ci")); applyErr != nil {
			return nil, m.safeError(applyErr)
		}
		m.db, err = open(legacy)
	}
	if err != nil {
		return nil, m.safeError(err)
	}
	if m.database == "" {
		var name sql.NullString
		if err = m.db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&name); err != nil {
			_ = m.Close()
			return nil, m.safeError(err)
		}
		m.database = name.String
	}
	if err = m.db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&m.version); err != nil {
		_ = m.Close()
		return nil, m.safeError(err)
	}
	return m, nil
}

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

// Test 返回 MySQL 实际版本，不暴露认证信息。
func (m *Manager) Test(ctx context.Context) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var version string
	if err := m.db.QueryRowContext(ctx, "SELECT version()").Scan(&version); err != nil {
		return nil, m.safeError(err)
	}
	return map[string]interface{}{
		"version":  version,
		"database": m.database,
	}, nil
}

// Databases 按名称分页查询当前账户有权访问的数据库，默认隐藏系统库。
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
	// SCHEMATA 与 SHOW DATABASES 一样按当前连接的权限限制可见范围。
	source := " FROM information_schema.SCHEMATA WHERE 1 = 1"
	if !showSystem {
		source += " AND SCHEMA_NAME NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys')"
	}
	args := make([]interface{}, 0, 3)
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		source += " AND LOCATE(LOWER(?), LOWER(SCHEMA_NAME)) > 0"
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
	args = append(args, pageSize, int64(pageNumber-1)*int64(pageSize))
	rows, err := m.db.QueryContext(ctx, "SELECT SCHEMA_NAME"+source+" ORDER BY BINARY SCHEMA_NAME "+strings.ToUpper(orderByType)+" LIMIT ? OFFSET ?", args...)
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
	return page.New(&page.Query{
		Page:     pageNumber,
		PageSize: pageSize,
	}, total, result), nil
}

// quote 转义 MySQL 标识符，点号作为名称本身处理。
func (m *Manager) quote(name string) (string, error) {
	if name == "" || strings.ContainsRune(name, 0) || len(name) > 1024 {
		return "", errors.New("表名或字段名不合法")
	}
	return "`" + strings.ReplaceAll(name, "`", "``") + "`", nil
}

// target 限定到当前数据库，避免表名被解释为跨库引用。
func (m *Manager) target(schemaName, tableName string) (string, string, error) {
	schema := schemaName
	if schema == "" {
		schema = m.database
	}
	if schema != m.database {
		return "", "", errors.New("只能操作当前数据库的表")
	}
	table, err := m.quote(tableName)
	if err != nil {
		return "", "", err
	}
	if schema != "" {
		name, err := m.quote(schema)
		if err != nil {
			return "", "", err
		}
		table = name + "." + table
	}
	return schema, table, nil
}

// serverVersion 同时识别 MySQL 和带 5.5.5 兼容前缀的 MariaDB 版本。
func (m *Manager) serverVersion(ctx context.Context) (int, bool, error) {
	version := m.version
	if version == "" {
		if err := m.db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version); err != nil {
			return 0, false, err
		}
	}
	maria := strings.Contains(strings.ToLower(version), "mariadb")
	matches := serverVersionPattern.FindAllStringSubmatch(version, -1)
	if len(matches) == 0 {
		return 0, maria, errors.New("无法识别数据库服务器版本")
	}
	match := matches[0]
	if maria && len(matches) > 1 && match[0] == "5.5.5" {
		match = matches[1]
	}
	major, _ := strconv.Atoi(match[1])
	minor, _ := strconv.Atoi(match[2])
	patch, _ := strconv.Atoi(match[3])
	return major*10000 + minor*100 + patch, maria, nil
}
