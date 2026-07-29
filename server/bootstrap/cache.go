package bootstrap

import (
	"server/app/util/cache"
	"server/global"
	"time"
)

// LoadCache 初始化缓存
func LoadCache() {
	global.App.SetCache(cache.NewMem(3600*time.Second, 60*time.Second))
}
