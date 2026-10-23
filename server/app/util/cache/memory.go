package cache

import (
	"sync"
	"time"

	goCache "github.com/patrickmn/go-cache"
)

// NewMem 实例化内存缓存
func NewMem(defaultExpiration, cleanupInterval time.Duration) *Cache {
	c := &Cache{
		memory:   goCache.New(defaultExpiration, 0),
		locks:    make(map[string]time.Time),
		stopChan: make(chan struct{}),
	}
	if cleanupInterval > 0 {
		c.waitGroup.Add(1)
		go c.cleanupRoutine(cleanupInterval)
	}
	return c
}

type Cache struct {
	memory    *goCache.Cache
	locks     map[string]time.Time
	mu        sync.Mutex
	stopChan  chan struct{}
	closeOnce sync.Once
	waitGroup sync.WaitGroup
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
	if c.locks[key].After(now) {
		return false
	}
	c.locks[key] = now.Add(expiration)
	return true
}

func (c *Cache) UnLock(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.locks, key)
}

// IsLock 检查指定键的锁是否处于上锁状态
func (c *Cache) IsLock(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.locks[key].After(time.Now())
}

func (c *Cache) cleanupRoutine(interval time.Duration) {
	defer c.waitGroup.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.memory.DeleteExpired()
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
	for key, expiration := range c.locks {
		if !expiration.After(now) {
			delete(c.locks, key)
		}
	}
}

// Close 停止清理goroutine并释放资源
func (c *Cache) Close() {
	c.closeOnce.Do(func() { close(c.stopChan) })
	c.waitGroup.Wait()
	c.mu.Lock()
	defer c.mu.Unlock()
	clear(c.locks)
}
