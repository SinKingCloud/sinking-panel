package elasticsearch

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Manager 每次操作使用独立 HTTP 客户端，不依赖跨大版本的 Elasticsearch SDK。
type Manager struct {
	client    *http.Client
	transport *http.Transport
	base      *url.URL
	index     string
	user      string
	password  string
	apiKey    string
	secrets   []string
	root      map[string]interface{}
	version   string
	major     int
	minor     int
}

// NewManager 使用主机和端口验证连接并读取版本，支持 Basic、API Key 和 TLS。
func NewManager(ctx context.Context, config, databaseName string) (*Manager, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var cfg Config
	config = strings.TrimSpace(config)
	decoder := json.NewDecoder(strings.NewReader(config))
	decoder.DisallowUnknownFields()
	if len(config) == 0 || len(config) > 1<<20 || config[0] != '{' || !json.Valid([]byte(config)) || decoder.Decode(&cfg) != nil {
		return nil, errors.New("Elasticsearch 连接配置必须是 JSON 对象")
	}
	if databaseName != "" {
		if err := ValidateIndexName(databaseName, false); err != nil {
			return nil, err
		}
	}
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}
	if cfg.Port == 0 {
		cfg.Port = 9200
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, errors.New("Elasticsearch 端口不合法")
	}
	mode := strings.ToLower(cfg.TLS)
	if mode != "" && mode != "false" && mode != "true" && mode != "skip-verify" {
		return nil, errors.New("Elasticsearch TLS 选项不合法")
	}
	scheme := "http"
	if mode == "true" || mode == "skip-verify" {
		scheme = "https"
	}
	base := &url.URL{Scheme: scheme, Host: net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))}
	if strings.ContainsAny(cfg.APIKey, "\r\n") {
		return nil, errors.New("Elasticsearch API Key 格式不正确")
	}
	secure := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: base.Hostname(), InsecureSkipVerify: mode == "skip-verify"}
	if scheme == "https" && cfg.SSLCA != "" {
		pem, readErr := os.ReadFile(cfg.SSLCA)
		if readErr != nil {
			return nil, errors.New("无法读取 Elasticsearch CA 证书")
		}
		secure.RootCAs, _ = x509.SystemCertPool()
		if secure.RootCAs == nil {
			secure.RootCAs = x509.NewCertPool()
		}
		if !secure.RootCAs.AppendCertsFromPEM(pem) {
			return nil, errors.New("Elasticsearch CA 证书格式不正确")
		}
	}
	if scheme == "https" && (cfg.SSLCert != "" || cfg.SSLKey != "") {
		if cfg.SSLCert == "" || cfg.SSLKey == "" {
			return nil, errors.New("客户端证书与私钥必须同时填写")
		}
		certificate, loadErr := tls.LoadX509KeyPair(cfg.SSLCert, cfg.SSLKey)
		if loadErr != nil {
			return nil, errors.New("无法读取 Elasticsearch 客户端证书或私钥")
		}
		secure.Certificates = []tls.Certificate{certificate}
	}
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment, DialContext: (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
		TLSClientConfig: secure, TLSHandshakeTimeout: 5 * time.Second, ResponseHeaderTimeout: 15 * time.Second,
		MaxIdleConns: 2, MaxIdleConnsPerHost: 2, IdleConnTimeout: 30 * time.Second}
	m := &Manager{transport: transport, base: base, index: databaseName, user: cfg.User, password: cfg.Password, apiKey: cfg.APIKey,
		secrets: []string{cfg.Password, cfg.APIKey}}
	m.client = &http.Client{Transport: transport, Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	data, _, err := m.doRequest(ctx, http.MethodGet, "/", nil)
	if err != nil {
		_ = m.Close()
		return nil, err
	}
	m.root, _ = data.(map[string]interface{})
	version, _ := m.root["version"].(map[string]interface{})
	m.version, _ = version["number"].(string)
	parts := strings.Split(m.version, ".")
	var majorErr, minorErr error
	if len(parts) >= 2 {
		m.major, majorErr = strconv.Atoi(parts[0])
		m.minor, minorErr = strconv.Atoi(parts[1])
	}
	if len(parts) < 2 || majorErr != nil || minorErr != nil || m.major < 0 || m.minor < 0 {
		_ = m.Close()
		return nil, errors.New("无法识别 Elasticsearch 服务器版本")
	}
	return m, nil
}

// Close 释放本次操作创建的 HTTP 连接。
func (m *Manager) Close() error { m.transport.CloseIdleConnections(); return nil }

// Test 返回连接状态和服务器版本，认证信息不会出现在结果中。
func (m *Manager) Test(ctx context.Context) (map[string]interface{}, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return map[string]interface{}{"connected": true, "version": m.version, "database": m.index, "cluster_name": m.root["cluster_name"]}, nil
}

// Version 通过 REST 功能适配有类型和无类型文档版本。
func (m *Manager) Version(ctx context.Context) (map[string]interface{}, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return map[string]interface{}{"name": "Elasticsearch", "version": m.version, "features": map[string]bool{
		"documents": true, "mapping": true, "settings": true, "rest": true, "accounts": false, "typed_documents": m.major < 7,
		"typed_mapping": m.major < 7, "text_keyword_fields": m.major >= 5,
		"mapping_backfill": m.major >= 5,
		"deep_pagination":  true, "pit_pagination": m.atLeast(7, 12),
		"multi_fields": m.major >= 1, "search_analyzer_update": m.major >= 2,
		"max_result_window": m.atLeast(2, 1), "api_key": m.atLeast(6, 7),
	}}, nil
}

// atLeast 比较功能引入的主次版本，不把 0.90 当作无法识别的版本。
func (m *Manager) atLeast(major, minor int) bool {
	return m.major > major || m.major == major && m.minor >= minor
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
