package traffic

import (
	"bufio"
	"context"
	"errors"
	"hash/maphash"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"golang.org/x/time/rate"
)

const (
	clientShardCount  = 64       // 单站点客户端并发计数分片数
	responseChunkSize = 32 << 10 // 响应限速写入块大小
)

var (
	states  = &stateRegistry{items: make(map[string]*state)}
	buffers = sync.Pool{New: func() interface{} {
		value := make([]byte, responseChunkSize)
		return &value
	}}
)

// Handler 限制站点并发请求和单个请求的响应速度。
type Handler struct {
	Scope               string `json:"scope"`
	MaxConnections      int64  `json:"max_connections,omitempty"`
	MaxConnectionsPerIP int64  `json:"max_connections_per_ip,omitempty"`
	RatePerRequest      int64  `json:"rate_per_request,omitempty"`

	lifecycle sync.RWMutex
	state     *state
	release   *sync.Once
	closed    bool
}

type stateRegistry struct {
	sync.Mutex
	items map[string]*state
}

type state struct {
	scope  string
	seed   maphash.Seed
	active atomic.Int64
	refs   int
	shards [clientShardCount]clientShard
}

type clientShard struct {
	sync.Mutex
	clients map[string]int64
}

type limitedResponseWriter struct {
	*caddyhttp.ResponseWriterWrapper
	requestContext context.Context
	limiter        *rate.Limiter
	burst          int
}

type writeOnly struct {
	io.Writer
}

func init() {
	caddy.RegisterModule(new(Handler))
}

// CaddyModule 返回流量处理器模块信息。
func (*Handler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.traffic_limit",
		New: func() caddy.Module { return new(Handler) },
	}
}

// Provision 取得站点共享状态，HTTP、HTTPS 和热加载配置共用同一份计数。
func (h *Handler) Provision(caddy.Context) error {
	h.Scope = strings.TrimSpace(h.Scope)
	if err := h.Validate(); err != nil {
		return err
	}
	h.release = new(sync.Once)
	h.state = states.acquire(h.Scope)
	return nil
}

// Validate 校验流量限制配置。
func (h *Handler) Validate() error {
	if h.Scope == "" || strings.ContainsAny(h.Scope, "\x00\r\n") {
		return errors.New("流量限制作用域无效")
	}
	if h.MaxConnections < 0 || h.MaxConnectionsPerIP < 0 || h.RatePerRequest < 0 {
		return errors.New("流量限制参数不能小于 0")
	}
	if h.MaxConnections == 0 && h.MaxConnectionsPerIP == 0 && h.RatePerRequest == 0 {
		return errors.New("至少需要配置一项流量限制")
	}
	return nil
}

// Cleanup 释放当前配置对共享状态的引用，在途请求仍可正常归还计数。
func (h *Handler) Cleanup() error {
	if h.release == nil {
		return nil
	}
	h.lifecycle.Lock()
	defer h.lifecycle.Unlock()
	h.closed = true
	h.release.Do(func() { states.release(h.state) })
	return nil
}

// ServeHTTP 执行并发检查，并按配置限制实际响应体传输速度。
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	h.lifecycle.RLock()
	if h.closed || h.state == nil {
		h.lifecycle.RUnlock()
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Set("Retry-After", "1")
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return nil
	}
	current := h.state
	client := ""
	if h.MaxConnectionsPerIP > 0 {
		client = h.clientIP(r)
	}
	concurrencyEnabled := h.MaxConnections > 0 || h.MaxConnectionsPerIP > 0
	status := 0
	if concurrencyEnabled {
		status = current.acquire(client, h.MaxConnections, h.MaxConnectionsPerIP)
	}
	h.lifecycle.RUnlock()
	if status != 0 {
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Set("Retry-After", "1")
		http.Error(w, http.StatusText(status), status)
		return nil
	}
	if concurrencyEnabled {
		defer states.releaseRequest(current, client, h.MaxConnectionsPerIP > 0)
	}

	if h.RatePerRequest > 0 {
		burst := int(min(h.RatePerRequest/10, responseChunkSize))
		if burst < 1 {
			burst = 1
		}
		w = &limitedResponseWriter{
			ResponseWriterWrapper: &caddyhttp.ResponseWriterWrapper{ResponseWriter: w},
			requestContext:        r.Context(),
			limiter:               rate.NewLimiter(rate.Limit(h.RatePerRequest), burst),
			burst:                 burst,
		}
	}
	return next.ServeHTTP(w, r)
}

func (h *Handler) clientIP(r *http.Request) string {
	client, _ := caddyhttp.GetVar(r.Context(), caddyhttp.ClientIPVarKey).(string)
	client = strings.TrimSpace(client)
	if client == "" {
		client = strings.TrimSpace(r.RemoteAddr)
		if host, _, err := net.SplitHostPort(client); err == nil {
			client = host
		}
	}
	if parsed := net.ParseIP(strings.Trim(client, "[]")); parsed != nil {
		return parsed.String()
	}
	if client == "" {
		return "unknown"
	}
	return client
}

func (r *stateRegistry) acquire(scope string) *state {
	r.Lock()
	defer r.Unlock()
	current := r.items[scope]
	if current == nil {
		current = &state{scope: scope, seed: maphash.MakeSeed()}
		r.items[scope] = current
	}
	current.refs++
	return current
}

func (r *stateRegistry) release(current *state) {
	if current == nil {
		return
	}
	r.Lock()
	defer r.Unlock()
	if r.items[current.scope] != current || current.refs == 0 {
		return
	}
	current.refs--
	if current.refs == 0 && current.active.Load() == 0 {
		delete(r.items, current.scope)
	}
}

func (r *stateRegistry) releaseRequest(current *state, client string, trackedClient bool) {
	current.release(client, trackedClient)
	if current.active.Load() != 0 {
		return
	}
	r.Lock()
	defer r.Unlock()
	if r.items[current.scope] == current && current.refs == 0 && current.active.Load() == 0 {
		delete(r.items, current.scope)
	}
}

func (s *state) acquire(client string, maxConnections, maxConnectionsPerIP int64) int {
	for {
		active := s.active.Load()
		if maxConnections > 0 && active >= maxConnections {
			return http.StatusServiceUnavailable
		}
		if s.active.CompareAndSwap(active, active+1) {
			break
		}
	}
	if maxConnectionsPerIP == 0 {
		return 0
	}
	shard := &s.shards[maphash.String(s.seed, client)%clientShardCount]
	shard.Lock()
	if shard.clients == nil {
		shard.clients = make(map[string]int64)
	}
	if maxConnectionsPerIP > 0 && shard.clients[client] >= maxConnectionsPerIP {
		shard.Unlock()
		s.active.Add(-1)
		return http.StatusTooManyRequests
	}
	shard.clients[client]++
	shard.Unlock()
	return 0
}

func (s *state) release(client string, trackedClient bool) {
	if trackedClient {
		shard := &s.shards[maphash.String(s.seed, client)%clientShardCount]
		shard.Lock()
		if shard.clients[client] <= 1 {
			delete(shard.clients, client)
		} else {
			shard.clients[client]--
		}
		shard.Unlock()
	}
	s.active.Add(-1)
}

func (w *limitedResponseWriter) Write(value []byte) (int, error) {
	written := 0
	for len(value) > 0 {
		size := min(len(value), w.burst)
		if err := w.limiter.WaitN(w.requestContext, size); err != nil {
			return written, err
		}
		count, err := w.ResponseWriter.Write(value[:size])
		written += count
		value = value[count:]
		if err != nil {
			return written, err
		}
		if count == 0 {
			return written, io.ErrShortWrite
		}
	}
	return written, nil
}

func (w *limitedResponseWriter) ReadFrom(reader io.Reader) (int64, error) {
	buffer := buffers.Get().(*[]byte)
	defer buffers.Put(buffer)
	return io.CopyBuffer(writeOnly{Writer: w}, reader, *buffer)
}

// Flush 保留流式响应能力，写出的响应体仍先经过限速。
func (w *limitedResponseWriter) Flush() {
	_ = w.FlushError()
}

// FlushError 将底层流式刷新错误返回给 ResponseController。
func (w *limitedResponseWriter) FlushError() error {
	return http.NewResponseController(w.ResponseWriterWrapper).Flush()
}

// Hijack 保留 WebSocket 等协议升级能力，升级后的连接不再经过 HTTP 响应限速。
func (w *limitedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(w.ResponseWriterWrapper).Hijack()
}
