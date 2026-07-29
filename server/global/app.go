package global

import (
	"log"

	"server/app/util/cache"
	"server/app/util/database"

	"github.com/spf13/viper"
)

var App = &Application{}

// SetDataBase 设置数据库实例，支持链式调用
func (a *Application) SetDataBase(d *database.Database) *Application {
	a.Database = d
	return a
}

// SetConfig 设置配置文件实例，支持链式调用
func (a *Application) SetConfig(c *viper.Viper) *Application {
	a.Config = c
	return a
}

// SetCache 设置缓存系统实例，支持链式调用
func (a *Application) SetCache(c cache.Interface) *Application {
	a.Cache = c
	return a
}

// SetLog 设置日志实例，支持链式调用
func (a *Application) SetLog(l *log.Logger) *Application {
	a.Log = l
	return a
}
