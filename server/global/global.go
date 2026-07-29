package global

import (
	"log"

	"server/app/util/cache"
	"server/app/util/database"

	"github.com/spf13/viper"
)

// Application 应用全局依赖
type Application struct {
	Database *database.Database //数据库
	Config   *viper.Viper       //配置文件
	Cache    cache.Interface    //缓存系统
	Log      *log.Logger        //日志系统
}
