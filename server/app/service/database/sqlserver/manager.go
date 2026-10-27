package sqlserver

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"server/app/util/page"
	"strconv"
	"strings"
	"sync"
	"time"

	mssql "github.com/microsoft/go-mssqldb"
)

// Manager 提供 SqlServer 连接、数据、账户和维护操作，连接池固定使用指定数据库。
type Manager struct {
	db        *sql.DB
	connector *connectionConnector
	database  string
	secrets   []string
	version   int // SQL Server 主版本，连接时读取一次。
}

// NewManager 打开已有数据库，调用方负责关闭；连接与执行均受请求 context 控制。
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
		connection.Port = 1433
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
	u := &url.URL{
		Scheme: "sqlserver",
		Host:   net.JoinHostPort(connection.Host, strconv.Itoa(connection.Port)),
		User:   url.UserPassword(connection.User, connection.Password),
	}
	q := u.Query()
	if databaseName != "" {
		q.Set("database", databaseName)
	}
	// 驱动的 connection timeout 同时限制业务读写；握手总时限由连接器单独控制。
	q.Set("connection timeout", "10")
	q.Set("dial timeout", "5")
	encrypt := connection.Encrypt
	if encrypt == "" {
		encrypt = "true"
	}
	q.Set("encrypt", encrypt)
	if connection.TrustServerCertificate != "" {
		q.Set("TrustServerCertificate", connection.TrustServerCertificate)
	}
	u.RawQuery = q.Encode()
	connector, err := mssql.NewConnector(u.String())
	if err != nil {
		return nil, m.safeError(err)
	}
	m.connector = &connectionConnector{Connector: connector}
	m.db = sql.OpenDB(m.connector)
	stop := m.connector.cancelConnections(ctx)
	defer stop()
	m.db.SetMaxOpenConns(2)
	m.db.SetMaxIdleConns(1)
	m.db.SetConnMaxLifetime(time.Minute)
	if err = m.db.PingContext(ctx); err != nil {
		_ = m.db.Close()
		return nil, m.safeError(err)
	}
	var version string
	if err = m.db.QueryRowContext(ctx, "SELECT CAST(SERVERPROPERTY('ProductVersion') AS VARCHAR(100))").Scan(&version); err != nil {
		_ = m.db.Close()
		return nil, m.safeError(err)
	}
	m.version, err = strconv.Atoi(strings.Split(version, ".")[0])
	if err != nil {
		_ = m.db.Close()
		return nil, errors.New("无法识别 SqlServer 版本")
	}
	return m, nil
}

func (dial connectionDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return dial(ctx, network, address)
}

// 驱动取消查询后仍等待服务端确认；连接测试到期直接关闭 socket。
func (c *connectionConnector) cancelConnections(ctx context.Context) func() bool {
	return context.AfterFunc(ctx, func() {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		for _, connection := range c.connections {
			_ = connection.Close()
		}
	})
}

// Connect 限制拨号、TLS、登录及故障转移的总耗时，连接成功后解除时限。
// go-mssqldb 只用传入 context 控制拨号，握手期间需主动关闭已建立的 socket。
func (c *connectionConnector) Connect(ctx context.Context) (driver.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	connector := *c.Connector
	var mutex sync.Mutex
	var connections []net.Conn
	stop := context.AfterFunc(ctx, func() {
		mutex.Lock()
		defer mutex.Unlock()
		for _, connection := range connections {
			_ = connection.Close()
		}
	})
	defer stop()
	connector.Dialer = connectionDialer(func(dialCtx context.Context, network, address string) (net.Conn, error) {
		connection, err := (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext(dialCtx, network, address)
		if err != nil {
			return nil, err
		}
		mutex.Lock()
		defer mutex.Unlock()
		if err = ctx.Err(); err != nil {
			_ = connection.Close()
			return nil, err
		}
		connections = append(connections, connection)
		c.mutex.Lock()
		c.connections = append(c.connections, connection)
		c.mutex.Unlock()
		return connection, nil
	})
	connection, err := connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	stop()
	if err = ctx.Err(); err != nil {
		_ = connection.Close()
		return nil, err
	}
	return connection, nil
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

// Test 返回服务端实际版本；不暴露认证信息。
func (m *Manager) Test(ctx context.Context) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if m.connector != nil {
		stop := m.connector.cancelConnections(ctx)
		defer stop()
	}
	query := "SELECT CAST(SERVERPROPERTY('ProductVersion') AS VARCHAR(100))"
	var version string
	if err := m.db.QueryRowContext(ctx, query).Scan(&version); err != nil {
		return nil, m.safeError(err)
	}
	return map[string]interface{}{
		"version":  version,
		"database": m.database,
	}, nil
}

// Databases 按名称搜索并分页返回当前账户可访问的数据库，默认隐藏系统库。
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

	source := " FROM sys.databases WHERE state = 0 AND HAS_DBACCESS(name) = 1"
	if !showSystem {
		source += " AND database_id > 4"
	}
	args := make([]interface{}, 0, 3)
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		source += " AND CHARINDEX(LOWER(@p1), LOWER(name)) > 0"
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

	// ROW_NUMBER 自 SQL Server 2005 可用，不要求 2012 才支持的 OFFSET/FETCH。
	query := "SELECT name FROM (SELECT name, ROW_NUMBER() OVER (ORDER BY name " + strings.ToUpper(orderByType) + ", database_id) AS row_number" + source + ") AS ordered WHERE row_number > " + m.placeholder(len(args)+1) + " AND row_number <= " + m.placeholder(len(args)+2) + " ORDER BY row_number"
	offset := int64(pageNumber-1) * int64(pageSize)
	args = append(args, offset, offset+int64(pageSize))
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
	return page.New(&page.Query{
		Page:     pageNumber,
		PageSize: pageSize,
	}, total, result), nil
}

func (m *Manager) quote(name string) (string, error) {
	if name == "" || strings.ContainsRune(name, 0) || len(name) > 1024 {
		return "", errors.New("表名或字段名不合法")
	}
	return "[" + strings.ReplaceAll(name, "]", "]]") + "]", nil
}

func (m *Manager) placeholder(position int) string {
	return "@p" + strconv.Itoa(position)
}

func (m *Manager) target(schemaName, tableName string) (string, string, error) {
	schema := schemaName
	if schema == "" {
		schema = "dbo"
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

// serverVersion 兼容直接构建 Manager 的调用方，正常连接不重复查询版本。
func (m *Manager) serverVersion(ctx context.Context) (int, error) {
	if m.version != 0 {
		return m.version, nil
	}
	var version string
	if err := m.db.QueryRowContext(ctx, "SELECT CAST(SERVERPROPERTY('ProductVersion') AS VARCHAR(100))").Scan(&version); err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.Split(version, ".")[0])
}
