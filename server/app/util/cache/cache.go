package cache

type Cache interface {
	Get(key string) (string, bool)
	SetWithExpire(key string, value string, ttl int64) bool
	SetExpire(key string, ttl int64) bool
	Set(key string, value string) bool
	Remember(key string, fun func() string, ttl int64) string
	Delete(key string) bool
	Lock(key string, ttl int64) bool
	IsLock(key string) bool
	UnLock(key string) bool
	Inc(key string) int
	Dec(key string) int
}
