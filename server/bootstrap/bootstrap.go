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

// Close 释放已加载的全局资源，默认保留容器运行；传入 true 时一并停止。
func Close(stopContainers ...bool) error {
	var err error
	if global.App.Container != nil {
		if len(stopContainers) > 0 && stopContainers[0] {
			err = global.App.Container.StopAll()
		} else {
			err = global.App.Container.Close()
		}
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
