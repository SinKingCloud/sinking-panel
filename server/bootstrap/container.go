package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"server/app/constant"
	"server/app/util/container"
	"server/global"
)

// LoadContainer 仅在运行服务时加载，避免管理命令接管运行中的容器。
func LoadContainer() {
	if global.App.Container != nil {
		return
	}
	executable, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("获取面板程序路径失败: %w", err))
	}
	if err := loadContainer(filepath.Dir(executable)); err != nil {
		panic(fmt.Errorf("初始化轻量容器失败: %w", err))
	}
}

func loadContainer(installationRoot string) error {
	if global.App.Container != nil {
		return nil
	}
	root, err := filepath.EvalSymlinks(installationRoot)
	if err != nil {
		return fmt.Errorf("解析安装目录失败: %w", err)
	}
	directory := filepath.Join(root, constant.ContainerPath)
	// 与卸载一致，安装目录内的容器路径不能通过符号链接重定向。
	for current := directory; current != root; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("读取容器目录失败: %w", err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("容器路径必须是普通目录，不能通过符号链接指向其他目录: %s", current)
		}
	}
	manager, err := container.NewManager(directory)
	if err != nil {
		return err
	}
	global.App.SetContainer(manager)
	return nil
}
