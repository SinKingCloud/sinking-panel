package cache

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/darkweak/storages/core"
	"github.com/dustin/go-humanize"
)

// SimpleFS 保持 cache-handler 的 storages.cache.simplefs 配置协议。
// Provision 只登记共享存储，不创建目录、文件或后台协程。
type SimpleFS struct {
	core.Configuration

	logger  core.Logger
	store   *store
	uuid    string
	release *sync.Once
}

func init() {
	caddy.RegisterModule(SimpleFS{})
}

// CaddyModule 返回与上游 simplefs 完全相同的模块 ID。
func (SimpleFS) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "storages.cache.simplefs",
		New: func() caddy.Module { return &SimpleFS{release: new(sync.Once)} },
	}
}

// Provision 校验配置并取得共享存储租约，实际数据库仍保持关闭。
func (s *SimpleFS) Provision(ctx caddy.Context) error {
	if s.release == nil {
		s.release = new(sync.Once)
	}
	s.logger = ctx.Logger(s).Sugar()
	path := s.Provider.Path
	allowedRoot := ""
	size := "0"
	maxSize := int64(-1)
	if configuration, ok := s.Provider.Configuration.(map[string]interface{}); ok {
		if value, exists := configuration["allowed_root"]; exists && value != nil {
			allowedRoot = strings.TrimSpace(fmt.Sprint(value))
		}
		if value, exists := configuration["path"]; path == "" && exists && value != nil {
			path = strings.TrimSpace(fmt.Sprint(value))
		}
		if value, exists := configuration["size"]; exists && value != nil {
			size = fmt.Sprint(value)
		}
		if value, exists := configuration["directory_size"]; exists && value != nil {
			switch current := value.(type) {
			case float64:
				if current > math.MaxInt64 || current < math.MinInt64 {
					return fmt.Errorf("simplefs directory_size 超出 int64 范围")
				}
				maxSize = int64(current)
			case float32:
				maxSize = int64(current)
			case int:
				maxSize = int64(current)
			case int64:
				maxSize = current
			case json.Number:
				var err error
				maxSize, err = strconv.ParseInt(string(current), 10, 64)
				if err != nil {
					return fmt.Errorf("解析 simplefs directory_size 失败: %w", err)
				}
			case string:
				parsed, err := humanize.ParseBytes(strings.TrimSpace(current))
				if err != nil {
					return fmt.Errorf("解析 simplefs directory_size 失败: %w", err)
				}
				if parsed > uint64(^uint64(0)>>1) {
					return fmt.Errorf("simplefs directory_size 超出 int64 范围")
				}
				maxSize = int64(parsed)
			default:
				return fmt.Errorf("simplefs directory_size 类型无效: %T", value)
			}
		}
	}
	if path == "" {
		current, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("获取 simplefs 默认目录失败: %w", err)
		}
		path = current
	}
	absolute, err := Registry.canonical(path)
	if err != nil {
		return fmt.Errorf("解析 simplefs 路径失败: %w", err)
	}
	if allowedRoot != "" {
		absolute, err = Registry.treePath(allowedRoot, path)
		if err != nil {
			return fmt.Errorf("simplefs 路径超出受控缓存目录: %w", err)
		}
	}
	if maxSize == 0 {
		maxSize = -1
	}
	if maxSize < -1 {
		return fmt.Errorf("simplefs directory_size 不能为负数")
	}
	s.store = Registry.acquire(absolute, maxSize, s.Stale)
	s.uuid = path + "-" + size
	Registry.register(s)
	core.RegisterStorage(s)
	return nil
}

// Cleanup 释放当前 HTTP 配置持有的租约。
func (s *SimpleFS) Cleanup() error {
	var err error
	if s.release == nil {
		return nil
	}
	s.release.Do(func() {
		Registry.unregister(s)
		err = Registry.release(s.store)
	})
	return err
}

// Name 返回 cache-handler 识别的存储名称。
func (*SimpleFS) Name() string { return "SIMPLEFS" }

// Uuid 必须保持 cache-handler v0.16 的 path-size 拼接格式。
func (s *SimpleFS) Uuid() string { return s.uuid }

// Init 不打开数据库，避免 Validate 和热加载阶段产生资源。
func (*SimpleFS) Init() error { return nil }

// MapKeys 返回指定前缀下尚未过期的键值，并移除返回键的前缀。
func (s *SimpleFS) MapKeys(prefix string) map[string]string {
	if s.store == nil {
		return map[string]string{}
	}
	values, err := s.store.mapKeys(prefix)
	if err != nil && s.logger != nil {
		s.logger.Errorf("读取 simplefs 键失败: %v", err)
	}
	return values
}

// ListKeys 返回多级缓存映射中的真实缓存键。
func (s *SimpleFS) ListKeys() []string {
	if s.store == nil {
		return []string{}
	}
	keys, err := s.store.listKeys()
	if err != nil && s.logger != nil {
		s.logger.Errorf("列出 simplefs 键失败: %v", err)
	}
	return keys
}

// Get 返回尚未过期的值。
func (s *SimpleFS) Get(key string) []byte {
	if s.store == nil {
		return nil
	}
	value, err := s.store.get(key)
	if err != nil && s.logger != nil {
		s.logger.Errorf("读取 simplefs 数据失败: %v", err)
	}
	return value
}

// Set 写入普通键值，duration 小于等于零表示不过期。
func (s *SimpleFS) Set(key string, value []byte, duration time.Duration) error {
	if s.store == nil {
		return errStoreClosed
	}
	return s.store.set(key, value, duration)
}

// Delete 删除一个键。
func (s *SimpleFS) Delete(key string) {
	if s.store == nil {
		return
	}
	if err := s.store.delete(key); err != nil && s.logger != nil {
		s.logger.Errorf("删除 simplefs 数据失败: %v", err)
	}
}

// DeleteMany 按正则表达式批量删除键。
func (s *SimpleFS) DeleteMany(pattern string) {
	if s.store == nil {
		return
	}
	if err := s.store.deleteMany(pattern); err != nil && s.logger != nil {
		s.logger.Errorf("批量删除 simplefs 数据失败: %v", err)
	}
}

// Reset 清空当前持久化存储。
func (s *SimpleFS) Reset() error {
	if s.store == nil {
		return errStoreClosed
	}
	return s.store.reset()
}

// GetMultiLevel 按 Vary、ETag、fresh 和 stale 规则读取响应。
func (s *SimpleFS) GetMultiLevel(key string, request *http.Request, validator *core.Revalidator) (fresh *http.Response, stale *http.Response) {
	value := s.Get(core.MappingKeyPrefix + key)
	if len(value) == 0 || request == nil || validator == nil {
		return nil, nil
	}
	fresh, stale, err := core.MappingElection(s, value, request, validator, s.logger)
	if err != nil && s.logger != nil {
		s.logger.Errorf("解析 simplefs 多级缓存失败: %v", err)
	}
	return fresh, stale
}

// SetMultiLevel 原子写入压缩响应及其 Vary/ETag 映射。
func (s *SimpleFS) SetMultiLevel(baseKey, variedKey string, value []byte, variedHeaders http.Header, etag string, duration time.Duration, realKey string) error {
	if s.store == nil {
		return errStoreClosed
	}
	return s.store.setMultiLevel(baseKey, variedKey, value, variedHeaders, etag, duration, realKey, s.logger)
}
