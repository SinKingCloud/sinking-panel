package cache

import (
	goCache "github.com/patrickmn/go-cache"
	"sync"
	"time"
)

func NewCache(defaultExpiration, cleanupInterval time.Duration) *Cache {
	c := &Cache{
		memory: goCache.New(
			defaultExpiration,
			cleanupInterval,
		),
		locks:    make(map[string]*lockInfo),
		mu:       sync.Mutex{},
		stopChan: make(chan struct{}),
	}
	go c.cleanupRoutine(cleanupInterval)
	return c
}

type lockInfo struct {
	mutex      sync.Mutex
	expiration time.Time
}

type Cache struct {
	memory   *goCache.Cache
	locks    map[string]*lockInfo
	mu       sync.Mutex
	stopChan chan struct{}
}

func (c *Cache) Get(key string) interface{} {
	obj, found := c.memory.Get(key)
	if found {
		return obj
	}
	return nil
}

func (c *Cache) Set(key string, value interface{}) {
	c.memory.SetDefault(key, value)
}

func (c *Cache) SetWithExpire(key string, value interface{}, expiration time.Duration) {
	c.memory.Set(key, value, expiration)
}

func (c *Cache) Delete(key string) {
	c.memory.Delete(key)
}

func (c *Cache) Remember(key string, fn func() interface{}, expiration time.Duration) interface{} {
	if value, exists := c.memory.Get(key); exists {
		return value
	}
	value := fn()
	c.SetWithExpire(key, value, expiration)
	return value
}

func (c *Cache) Lock(key string, expiration time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	info, exists := c.locks[key]
	if exists && info.expiration.After(now) {
		return false
	}
	if !exists {
		info = &lockInfo{}
		c.locks[key] = info
	}

	info.expiration = now.Add(expiration)
	locked := info.mutex.TryLock()
	if !locked {
		delete(c.locks, key)
		return false
	}
	return true
}

func (c *Cache) UnLock(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	info, exists := c.locks[key]
	if !exists {
		return
	}
	info.mutex.Unlock()
	if info.expiration.Before(time.Now()) {
		delete(c.locks, key)
	}
}

func (c *Cache) cleanupRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.CleanExpiredLock()
		case <-c.stopChan:
			return
		}
	}
}

func (c *Cache) CleanExpiredLock() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for key, info := range c.locks {
		if info.expiration.Before(now) {
			info.mutex.Unlock()
			delete(c.locks, key)
		}
	}
}

// Close 停止清理goroutine并释放资源
func (c *Cache) Close() {
	close(c.stopChan)

	// 清理所有锁
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, info := range c.locks {
		info.mutex.Unlock()
		delete(c.locks, key)
	}
}
