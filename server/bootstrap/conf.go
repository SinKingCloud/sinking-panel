package bootstrap

import (
	"github.com/spf13/viper"
	"server/app/util"
)

// LoadConf 加载本地配置
func LoadConf() {
	config := viper.New()
	config.AutomaticEnv() //读取环境变量
	config.AddConfigPath("./config")
	config.SetConfigName("application.yml")
	config.SetConfigType("yaml")
	config.WatchConfig()
	if err := config.ReadInConfig(); err != nil {
		panic(err)
		return
	}
	//赋值到util
	util.Conf = config
}
