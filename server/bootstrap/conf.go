package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"server/app/constant"
	"server/global"

	"github.com/spf13/viper"
)

// LoadConf 加载可选本地配置，缺失时使用默认值，不创建目录或文件。
func LoadConf() {
	if global.App.Config != nil {
		return
	}
	config := viper.New()
	config.SetDefault(constant.ServerMode, "release")
	config.SetDefault(constant.ServerHost, "0.0.0.0")
	config.SetDefault(constant.ServerPort, 5678)
	config.SetConfigFile(filepath.Join(constant.ConfPath, constant.ConfFile))
	config.SetConfigType("yaml")
	if err := config.ReadInConfig(); err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(fmt.Errorf("读取配置文件失败: %w", err))
	}
	global.App.SetConfig(config)
}
