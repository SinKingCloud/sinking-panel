package bootstrap

import (
	"github.com/spf13/viper"
	"server/app/constant"
	"server/app/util/file"
	"server/global"
)

// LoadConf 加载本地配置
func LoadConf() {
	if global.App.Config != nil {
		return
	}
	path := constant.ConfPath
	fileName := constant.ConfFile
	disk := file.NewDisk(path)
	_ = disk.AutoCreate(fileName)
	config := viper.New()
	config.AutomaticEnv() //读取环境变量
	config.AddConfigPath(path)
	config.SetConfigName(fileName)
	config.SetConfigType("yaml")
	if err := config.ReadInConfig(); err != nil {
		panic(err)
	}
	global.App.SetConfig(config)
}
