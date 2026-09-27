package bootstrap

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"server/app/constant"
	"server/app/util/container"
	"server/app/util/platform"
	"server/global"
)

// LoadContainer 在服务启动时加载容器管理器，接管运行实例并启动已停止的容器。
func LoadContainer() error {
	if global.App.Container != nil {
		return nil
	}
	// 普通用户仍可运行面板，只有 Linux 容器运行时需要 root。
	if platform.IsLinux() && os.Geteuid() != 0 {
		log.Println("当前用户非 root，跳过轻量容器初始化")
		return nil
	}
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取面板程序路径失败: %w", err)
	}
	root, err := filepath.EvalSymlinks(filepath.Dir(executable))
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
		return fmt.Errorf("初始化轻量容器失败: %w", err)
	}
	global.App.SetContainer(manager)
	for _, instance := range manager.List() {
		if instance.Status != container.StatusStopped && instance.Status != container.StatusFailed {
			continue
		}
		if _, err := manager.Start(instance.ID); err != nil {
			// 自动启动可能已先完成；已运行的容器直接接管，不重复启动。
			if info, infoErr := manager.Info(instance.ID); infoErr == nil && info.Status == container.StatusRunning {
				continue
			}
			log.Printf("启动容器 %s 失败: %v", instance.ID, err)
		}
	}
	return nil
}
