package command

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"server/app"
	"server/app/constant"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/container"
	"server/app/util/daemon"
	"server/bootstrap"
	"server/global"
)

const (
	logFileName = "server.log"
	serviceName = "sinking-panel"
)

// initDaemon 仅在执行面板命令时初始化守护进程，插件命令独立处理。
func (s *Server) initDaemon() error {
	if s.daemon != nil {
		return nil
	}
	d, err := daemon.NewDaemon(constant.DaemonPath, logFileName, s.run)
	if err != nil {
		return fmt.Errorf("创建守护进程管理器失败: %w", err)
	}
	s.daemon = d.SetAutoStartOptions(daemon.AutoStartOptions{
		Name:               serviceName,
		KeepChildProcesses: true,
	})
	global.App.Daemon = s.daemon
	return nil
}

func (s *Server) configureStartup(command string, options *flag.FlagSet) error {
	bootstrap.LoadConf()
	arguments := make([]string, 0, options.NFlag())
	options.Visit(func(option *flag.Flag) {
		// 容器选项只控制本次退出，不写入配置或下次启动参数。
		if option.Name == "containers" {
			return
		}
		global.App.Config.Set("server."+option.Name, option.Value.String())
		arguments = append(arguments, "--"+option.Name+"="+option.Value.String())
	})
	s.daemon.SetChildArgs(append([]string{os.Args[0], "start"}, arguments...)...)
	workingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("读取程序工作目录失败: %w", err)
	}
	s.daemon.SetAutoStartOptions(daemon.AutoStartOptions{
		Name:               serviceName,
		WorkingDirectory:   workingDirectory,
		Arguments:          append([]string{"run"}, arguments...),
		KeepChildProcesses: true,
	})
	// 没有传入新配置时沿用安装服务时保存的端口等参数。
	if len(arguments) > 0 && (command == "start" || command == "restart") && !s.daemon.IsChildProcess() {
		if err := s.daemon.SyncAutoStart(); err != nil {
			return fmt.Errorf("同步系统启动配置失败: %w", err)
		}
	}
	return nil
}

func (s *Server) run(stop <-chan struct{}) {
	defer func() {
		if err := bootstrap.Close(); err != nil {
			log.Printf("释放程序资源失败: %v", err)
		}
	}()
	bootstrap.Load()
	if err := bootstrap.LoadContainer(); err != nil {
		panic(fmt.Errorf("初始化轻量容器失败: %w", err))
	}
	service.Init()
	shutdown := service.Boot()
	defer shutdown()
	if closeCommand, err := s.listenPluginCommand(service.Software); err != nil {
		log.Printf("启动插件命令服务失败: %v", err)
	} else {
		defer closeCommand()
	}
	app.Run(stop)
}

func (s *Server) install() error {
	if err := s.requireInteractive("install"); err != nil {
		return err
	}
	account, err := s.readAccount("请输入登录账号: ")
	if err != nil {
		return err
	}
	password, err := s.readNewPassword("请输入登录密码: ", "请再次输入登录密码: ", true)
	if err != nil {
		return err
	}
	defer clear(password)

	bootstrap.Load()
	defer bootstrap.Close()
	service.Init()
	if err = service.Auth.UpdateAccount(account, string(password)); err != nil {
		return fmt.Errorf("设置登录账号密码失败: %w", err)
	}
	service.Log.Create("127.0.0.1", log_type.EventUpdate, "初始化登录信息", "通过安装命令设置登录账号密码")
	log.Println("登录账号密码设置成功")

	if err = s.daemon.InstallAutoStart(); err != nil {
		return fmt.Errorf("登录账号密码已设置，但安装系统自启动失败: %w", err)
	}
	log.Println("系统自启动安装成功")
	// 启动前释放安装命令持有的数据库和缓存。
	if err := bootstrap.Close(); err != nil {
		return err
	}
	return s.start()
}

func (s *Server) start() error {
	log.Println("正在启动服务...")
	if err := s.daemon.Start(); err != nil {
		return fmt.Errorf("启动失败: %w", err)
	}
	if !s.daemon.IsChildProcess() {
		log.Println("服务已启动")
	}
	return nil
}

func (s *Server) stop(containers bool) error {
	log.Println("正在停止服务...")
	if err := s.daemon.Stop(); err != nil {
		return fmt.Errorf("停止失败: %w", err)
	}
	if containers {
		// 旧面板退出后统一接管，面板此前已停止也走相同流程。
		executable, err := os.Executable()
		if err != nil {
			return err
		}
		root, err := filepath.EvalSymlinks(filepath.Dir(executable))
		if err != nil {
			return err
		}
		directory := filepath.Join(root, constant.ContainerPath)
		resolved, err := filepath.EvalSymlinks(directory)
		if !errors.Is(err, os.ErrNotExist) {
			if err != nil {
				return fmt.Errorf("读取容器目录失败: %w", err)
			}
			if resolved != directory {
				return errors.New("容器目录不能通过符号链接指向其他目录")
			}
			// 清理时仅加载已有状态，不能把已停止的容器重新启动。
			manager, err := container.NewManager(directory, container.ManagerOptions{DisableAutoRestart: true})
			if err != nil {
				return err
			}
			if err := manager.StopAll(); err != nil {
				return fmt.Errorf("停止容器失败: %w", err)
			}
		}
	}
	log.Println("服务已停止")
	return nil
}

func (s *Server) restart(containers bool) error {
	if err := s.stop(containers); err != nil {
		return err
	}
	return s.start()
}
