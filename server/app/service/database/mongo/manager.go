package mongo

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"net"
	"os"
	"regexp"
	"server/app/util/page"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	client "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Manager 管理指定数据库的连接，可独立完成文档、索引和账户操作。
type Manager struct {
	client   *client.Client
	database *client.Database
	secrets  []string
}

// NewManager 使用结构化参数配置 MongoDB 主机、认证库和 TLS。
func NewManager(ctx context.Context, config, databaseName string) (*Manager, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var cfg Config
	config = strings.TrimSpace(config)
	decoder := json.NewDecoder(strings.NewReader(config))
	decoder.DisallowUnknownFields()
	if len(config) == 0 || config[0] != '{' || !json.Valid([]byte(config)) || decoder.Decode(&cfg) != nil {
		return nil, errors.New("MongoDB 连接配置必须是 JSON 对象")
	}
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}
	if cfg.Port == 0 {
		cfg.Port = 27017
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, errors.New("MongoDB 端口不合法")
	}
	opts := options.Client().SetHosts([]string{net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))}).SetDirect(cfg.DirectConnection)
	if cfg.User != "" {
		if cfg.AuthSource == "" {
			cfg.AuthSource = "admin"
		}
		opts.SetAuth(options.Credential{
			Username:   cfg.User,
			Password:   cfg.Password,
			AuthSource: cfg.AuthSource,
		})
	}
	if cfg.ReplicaSet != "" {
		opts.SetReplicaSet(cfg.ReplicaSet)
	}
	if databaseName == "" {
		databaseName = cfg.Database
	}
	if databaseName == "" {
		databaseName = "admin"
	}
	mode := strings.ToLower(cfg.TLS)
	if mode != "" && mode != "false" && mode != "disable" {
		opts.TLSConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: mode == "skip-verify",
		}
		if cfg.SSLCA != "" {
			pem, err := os.ReadFile(cfg.SSLCA)
			if err != nil {
				return nil, errors.New("无法读取 MongoDB TLS CA 文件")
			}
			opts.TLSConfig.RootCAs = x509.NewCertPool()
			if !opts.TLSConfig.RootCAs.AppendCertsFromPEM(pem) {
				return nil, errors.New("MongoDB TLS CA 文件格式不正确")
			}
		}
		if cfg.SSLCert != "" {
			certificate, err := tls.LoadX509KeyPair(cfg.SSLCert, cfg.SSLKey)
			if err != nil {
				return nil, errors.New("无法读取 MongoDB TLS 客户端证书")
			}
			opts.TLSConfig.Certificates = []tls.Certificate{certificate}
		}
	}
	opts.SetConnectTimeout(5 * time.Second).SetServerSelectionTimeout(5 * time.Second).SetSocketTimeout(30 * time.Second).SetMaxPoolSize(2).SetMinPoolSize(0).SetRetryReads(false).SetRetryWrites(false)
	if err := opts.Validate(); err != nil {
		return nil, errors.New("MongoDB 连接配置不合法")
	}
	c, err := client.Connect(ctx, opts)
	if err != nil {
		return nil, errors.New("创建 MongoDB 连接失败")
	}
	m := &Manager{
		client:   c,
		database: c.Database(databaseName),
		secrets:  []string{cfg.Password},
	}
	if err = c.Ping(ctx, readpref.PrimaryPreferred()); err != nil {
		_ = m.CloseContext(ctx)
		return nil, m.cleanError(err)
	}
	return m, nil
}

// Close 释放连接和驱动监控协程。
func (m *Manager) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.CloseContext(ctx)
}

// CloseContext 让连接测试的会话清理也使用原请求的剩余时间。
func (m *Manager) CloseContext(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// Test 测试登录和读取服务器版本；无 buildInfo 权限仍可确认连接成功。
func (m *Manager) Test(ctx context.Context) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := m.client.Ping(ctx, readpref.PrimaryPreferred()); err != nil {
		return nil, m.cleanError(err)
	}
	result := map[string]interface{}{
		"connected": true,
		"database":  m.database.Name(),
	}
	var info bson.M
	if err := m.database.RunCommand(ctx, bson.D{{
		Key:   "buildInfo",
		Value: 1,
	}}).Decode(&info); err == nil {
		result["version"] = info["version"]
	} else if client.IsTimeout(err) {
		return nil, m.cleanError(err)
	}
	if err := ctx.Err(); err != nil {
		return nil, m.cleanError(err)
	}
	return result, nil
}

// Databases 按名称搜索并分页返回当前账户可访问的数据库，默认隐藏系统库。
func (m *Manager) Databases(ctx context.Context, keyword string, pageNumber, pageSize int, showSystem bool, orderByType string) (_ *page.Result[string], err error) {
	defer func() { err = m.cleanError(err) }()
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

	// listDatabases 不提供游标，先由服务器过滤系统库及关键词，再对名称分页。
	filter := bson.D{}
	nameFilter := bson.D{}
	if !showSystem {
		nameFilter = append(nameFilter, bson.E{
			Key:   "$nin",
			Value: bson.A{"admin", "config", "local"},
		})
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		nameFilter = append(nameFilter, bson.E{
			Key:   "$regex",
			Value: regexp.QuoteMeta(keyword),
		}, bson.E{
			Key:   "$options",
			Value: "i",
		})
	}
	if len(nameFilter) > 0 {
		filter = append(filter, bson.E{
			Key:   "name",
			Value: nameFilter,
		})
	}
	names, err := m.client.ListDatabaseNames(ctx, filter, options.ListDatabases().SetNameOnly(true))
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	if orderByType == "desc" {
		slices.Reverse(names)
	}
	total := int64(len(names))

	if total == 0 {
		pageNumber = 1
	} else if lastPage := (total-1)/int64(pageSize) + 1; int64(pageNumber) > lastPage {
		pageNumber = int(lastPage)
	}

	start := (pageNumber - 1) * pageSize
	end := min(start+pageSize, len(names))
	result := append([]string{}, names[start:end]...)
	return page.New(&page.Query{
		Page:     pageNumber,
		PageSize: pageSize,
	}, total, result), nil
}

func (m *Manager) cleanError(err error) error {
	if err == nil {
		return nil
	}
	text := err.Error()
	for _, secret := range m.secrets {
		if secret != "" {
			text = strings.ReplaceAll(text, secret, "***")
		}
	}
	return errors.New(text)
}
