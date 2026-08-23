package bootstrap

import "server/global"

func Load() {
	LoadConf()
	LoadLog()
	LoadCache()
	LoadDatabase()
}

// Close 按依赖顺序释放全局资源。
func Close() error {
	if global.App.Cache != nil {
		global.App.Cache.Close()
		global.App.Cache = nil
	}
	var err error
	if global.App.Database != nil {
		err = global.App.Database.Close()
		global.App.Database = nil
	}
	return err
}
