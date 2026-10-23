package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"server/app/util/page"
	"strings"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

// Manager 管理单次请求使用的 Sqlite 文件连接。
type Manager struct {
	db *sql.DB
}

// NewManager 只打开已有文件，不因连接测试或数据浏览创建新的数据库。
func NewManager(ctx context.Context, config, databaseName string) (*Manager, error) {
	if len(config) > 1<<20 {
		return nil, errors.New("连接配置超过 1 MB")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var values map[string]json.RawMessage
	if err := json.Unmarshal([]byte(config), &values); err != nil || values == nil {
		return nil, errors.New("连接配置必须是 JSON 对象")
	}
	value := func(name string) string {
		var text string
		if json.Unmarshal(values[name], &text) == nil {
			return text
		}
		return string(values[name])
	}
	if databaseName == "" {
		databaseName = value("database")
	}
	if databaseName != "" && databaseName != "main" {
		return nil, errors.New("Sqlite 仅支持已配置文件的 main 数据库")
	}
	path := value("path")
	if path == "" {
		return nil, errors.New("请填写 Sqlite 文件路径")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.New("Sqlite 文件路径不正确")
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return nil, errors.New("Sqlite 数据库文件不存在或不是普通文件")
	}
	u := &url.URL{Scheme: "file", Path: filepath.ToSlash(path)}
	q := u.Query()
	q.Set("mode", "rw")
	if value("read_only") == "true" {
		q.Set("mode", "ro")
	}
	q.Add("_pragma", "busy_timeout(5000)")
	q.Add("_pragma", "foreign_keys(1)")
	u.RawQuery = q.Encode()
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Minute)
	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Manager{db: db}, nil
}

// Close 释放本次请求的文件连接。
func (m *Manager) Close() error { return m.db.Close() }

// Test 读取实际使用的 Sqlite 引擎版本。
func (m *Manager) Test(ctx context.Context) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var version string
	if err := m.db.QueryRowContext(ctx, "SELECT sqlite_version()").Scan(&version); err != nil {
		return nil, err
	}
	return map[string]interface{}{"version": version, "database": "main"}, nil
}

// Databases 按名称搜索并分页返回当前账户可访问的数据库。
func (m *Manager) Databases(ctx context.Context, keyword string, pageNumber, pageSize int) (_ *page.Result[string], err error) {
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

	if err = ctx.Err(); err != nil {
		return nil, err
	}
	result := []string{}
	if strings.Contains("main", strings.ToLower(strings.TrimSpace(keyword))) {
		result = append(result, "main")
	}
	return page.New(&page.Query{Page: 1, PageSize: pageSize}, int64(len(result)), result), nil
}

func (m *Manager) quote(name string) (string, error) {
	if name == "" || strings.ContainsRune(name, 0) || len(name) > 1024 {
		return "", errors.New("表名或字段名不合法")
	}
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`, nil
}

func (m *Manager) target(schemaName, tableName string) (string, error) {
	if schemaName != "" && schemaName != "main" {
		return "", errors.New("Sqlite 仅支持 main 数据库")
	}
	table, err := m.quote(tableName)
	if err != nil {
		return "", err
	}
	return `"main".` + table, nil
}
