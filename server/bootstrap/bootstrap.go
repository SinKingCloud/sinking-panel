package bootstrap

import (
	"errors"
	"server/global"
)

func Load() {
	LoadConf()
	LoadLog()
	LoadCache()
	LoadDatabase()
}

// Close 按依赖顺序释放全局资源。
func Close() error {
	var err error
	if global.App.Container != nil {
		err = global.App.Container.StopAll()
		global.App.Container = nil
	}
	if global.App.Cache != nil {
		global.App.Cache.Close()
		global.App.Cache = nil
	}
	if global.App.Database != nil {
		err = errors.Join(err, global.App.Database.Close())
		global.App.Database = nil
	}
	return err
}
