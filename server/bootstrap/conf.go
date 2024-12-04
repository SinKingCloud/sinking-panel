package bootstrap

import (
	"github.com/spf13/viper"
	"server/app/constant"
	"server/app/util"
	file2 "server/app/util/file"
	"strings"
)

// LoadConf 加载本地配置
func LoadConf() {
	path := constant.ConfPath
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	fileName := constant.ConfFile
	file := file2.New(path)
	_ = file.EnsureDirectory(path, 0777)
	if !file.Has(fileName) {
		_, _ = file.Write(fileName, "", 0)
	}
	config := viper.New()
	config.AutomaticEnv() //读取环境变量
	config.AddConfigPath(path)
	config.SetConfigName(fileName)
	config.SetConfigType("yaml")
	config.WatchConfig()
	if err := config.ReadInConfig(); err != nil {
		panic(err)
		return
	}
	//赋值到util
	util.Conf = config
}
