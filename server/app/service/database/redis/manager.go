package redis

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"net"
	"os"
	"server/app/util/page"
	"slices"
	"strconv"
	"strings"
	"time"

	client "github.com/redis/go-redis/v9"
)

// Manager 使用独立连接选择逻辑库，避免不同请求相互影响 SELECT 状态。
type Manager struct {
	client   *client.Client
	database int
	secrets  []string
}

// NewManager 使用结构化参数配置 Redis 连接。
func NewManager(ctx context.Context, config, databaseName string) (*Manager, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var cfg Config
	config = strings.TrimSpace(config)
	decoder := json.NewDecoder(strings.NewReader(config))
	decoder.DisallowUnknownFields()
	if len(config) == 0 || config[0] != '{' || !json.Valid([]byte(config)) || decoder.Decode(&cfg) != nil {
		return nil, errors.New("Redis 连接配置必须是 JSON 对象")
	}
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}
	if cfg.Port == 0 {
		cfg.Port = 6379
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, errors.New("Redis 端口不合法")
	}
	opts := &client.Options{
		Addr:     net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Username: cfg.User,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	// 单参数 AUTH 在 Redis 6+ 同样代表 default 用户，也兼容没有 ACL 的旧版本。
	if opts.Username == "default" {
		opts.Username = ""
	}
	if cfg.Socket != "" {
		opts.Network, opts.Addr = "unix", cfg.Socket
	}
	if databaseName != "" {
		var err error
		opts.DB, err = strconv.Atoi(databaseName)
		if err != nil {
			return nil, errors.New("Redis 数据库名称必须是编号")
		}
	}
	if opts.DB < 0 {
		return nil, errors.New("Redis 数据库编号不能小于 0")
	}
	mode := strings.ToLower(cfg.TLS)
	if mode != "" && mode != "false" && mode != "disable" {
		opts.TLSConfig = &tls.Config{
			ServerName:         cfg.Host,
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: mode == "skip-verify",
		}
		if cfg.SSLCA != "" {
			pem, readErr := os.ReadFile(cfg.SSLCA)
			if readErr != nil {
				return nil, errors.New("无法读取 Redis TLS CA 文件")
			}
			opts.TLSConfig.RootCAs = x509.NewCertPool()
			if !opts.TLSConfig.RootCAs.AppendCertsFromPEM(pem) {
				return nil, errors.New("Redis TLS CA 文件格式不正确")
			}
		}
		if cfg.SSLCert != "" {
			certificate, loadErr := tls.LoadX509KeyPair(cfg.SSLCert, cfg.SSLKey)
			if loadErr != nil {
				return nil, errors.New("无法读取 Redis TLS 客户端证书")
			}
			opts.TLSConfig.Certificates = []tls.Certificate{certificate}
		}
	}
	// RESP2 与禁用客户端身份指令兼容旧版本；禁用重试，避免写操作重复执行。
	opts.Protocol, opts.DisableIdentity = 2, true
	opts.DialTimeout, opts.ReadTimeout, opts.WriteTimeout = 5*time.Second, 30*time.Second, 30*time.Second
	// 默认 TLS 拨号忽略 context；显式传递以遵守测试连接的剩余预算。
	opts.Dialer = func(ctx context.Context, network, address string) (net.Conn, error) {
		dialer := &net.Dialer{Timeout: 5 * time.Second}
		if opts.TLSConfig != nil {
			return (&tls.Dialer{
				NetDialer: dialer,
				Config:    opts.TLSConfig,
			}).DialContext(ctx, network, address)
		}
		return dialer.DialContext(ctx, network, address)
	}
	opts.DialerRetries = 1
	opts.ContextTimeoutEnabled, opts.MaxRetries = true, -1
	opts.PoolSize, opts.MinIdleConns, opts.MaxActiveConns = 1, 0, 1
	m := &Manager{
		client:   client.NewClient(opts),
		database: opts.DB,
		secrets:  []string{cfg.Password},
	}
	if err := m.client.Ping(ctx).Err(); err != nil {
		_ = m.Close()
		return nil, m.cleanError(err)
	}
	return m, nil
}

// Close 释放本次请求的连接。
func (m *Manager) Close() error { return m.client.Close() }

// Test 验证连接，INFO 权限不足不影响 PING 成功。
func (m *Manager) Test(ctx context.Context) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := m.client.Ping(ctx).Err(); err != nil {
		return nil, m.cleanError(err)
	}
	result := map[string]interface{}{
		"database":  m.database,
		"connected": true,
	}
	if info, err := m.info(ctx, "server"); err == nil {
		for _, line := range strings.Split(info, "\n") {
			if strings.HasPrefix(line, "redis_version:") {
				result["version"] = strings.TrimSpace(strings.TrimPrefix(line, "redis_version:"))
			}
		}
	} else {
		var networkError net.Error
		if errors.As(err, &networkError) && networkError.Timeout() {
			return nil, m.cleanError(err)
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, m.cleanError(err)
	}
	return result, nil
}

// Databases 分页查询逻辑库；受限账号仅返回可见库和当前库。
func (m *Manager) Databases(ctx context.Context, keyword string, pageNumber, pageSize int) (*page.Result[string], error) {
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
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	count := 0
	var numbers []int
	if config, err := m.client.ConfigGet(ctx, "databases").Result(); err == nil {
		if value, parseErr := strconv.Atoi(config["databases"]); parseErr == nil && value > 0 && value <= 65536 {
			count = value
		}
	}
	if count == 0 {
		info, err := m.info(ctx, "keyspace")
		if err != nil {
			if !strings.Contains(strings.ToUpper(err.Error()), "NOPERM") {
				return nil, m.cleanError(err)
			}
			// 初始化已通过 PING；无法读取实例配置时只显示已连接的逻辑库。
			numbers = []int{m.database}
		} else {
			seen := map[int]bool{
				0:          true,
				m.database: true,
			}
			for _, line := range strings.Split(info, "\n") {
				if key, _, ok := strings.Cut(line, ":"); ok && strings.HasPrefix(key, "db") {
					if number, parseErr := strconv.Atoi(strings.TrimPrefix(key, "db")); parseErr == nil && number >= 0 && number <= 65535 {
						seen[number] = true
					}
				}
			}
			numbers = make([]int, 0, len(seen))
			for number := range seen {
				numbers = append(numbers, number)
			}
			slices.Sort(numbers)
		}
		count = len(numbers)
	}
	// 配置可读时按编号生成当前页，不构造全部逻辑库列表。
	numberAt := func(index int) int {
		if numbers != nil {
			return numbers[index]
		}
		return index
	}
	total := int64(count)
	if keyword != "" {
		total = 0
		for i := 0; i < count; i++ {
			if strings.Contains("db "+strconv.Itoa(numberAt(i)), keyword) {
				total++
			}
		}
	}
	if total == 0 {
		pageNumber = 1
	} else if lastPage := (total-1)/int64(pageSize) + 1; int64(pageNumber) > lastPage {
		pageNumber = int(lastPage)
	}
	offset := (pageNumber - 1) * pageSize
	result := make([]string, 0, pageSize)
	matched := 0
	start := 0
	if keyword == "" {
		start, matched = offset, offset
	}
	for i := start; i < count && len(result) < pageSize; i++ {
		name := strconv.Itoa(numberAt(i))
		if keyword != "" && !strings.Contains("db "+name, keyword) {
			continue
		}
		if matched >= offset {
			result = append(result, name)
		}
		matched++
	}
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
