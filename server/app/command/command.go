package command

import (
	"fmt"
	"log"
	"os"

	"server/app"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/daemon"
	"server/bootstrap"
)

const (
	pidFileName = "server.pid"
	logFileName = "server.log"
	serviceName = "sinking-panel"
	usage       = `使用方法:
  server [command]

可用命令:
  start   启动服务
  stop    停止服务
  restart 重启服务
  run     直接运行(非守护进程模式)
  install 设置登录信息并安装系统自启动
  uninstall 卸载软件，保留网站数据并询问是否保留容器
  user    修改登录账号
  pwd     修改登录密码`
)

// Server 管理服务运行命令。
type Server struct {
	daemon *daemon.Daemon
}

// Run 是服务命令的统一入口。
// init 必须在创建守护进程之前处理，避免 libcontainer 子进程误进入面板启动流程。
func Run(args []string) error {
	if ExecuteInit(args) {
		return nil
	}
	server, err := NewServer()
	if err != nil {
		return err
	}
	return server.Execute(args)
}

// NewServer 创建服务命令。
func NewServer() (*Server, error) {
	server := &Server{}
	d, err := daemon.NewDaemon(pidFileName, logFileName, server.run)
	if err != nil {
		return nil, fmt.Errorf("创建守护进程管理器失败: %w", err)
	}
	server.daemon = d.SetChildArgs(os.Args[0], "start").SetAutoStartOptions(daemon.AutoStartOptions{
		Name:      serviceName,
		Arguments: []string{"run"},
	})
	return server, nil
}

// Execute 执行服务命令。
func (s *Server) Execute(args []string) error {
	if len(args) == 0 {
		log.Println(usage)
		return nil
	}
	switch args[0] {
	case "run":
		log.Println("以前台模式运行服务...")
		return s.daemon.Run()
	case "install":
		return s.install()
	case "uninstall":
		return s.uninstall()
	case "user":
		return s.user()
	case "pwd":
		return s.pwd()
	case "start", "stop", "restart":
		switch args[0] {
		case "start":
			return s.start()
		case "stop":
			return s.stop()
		default:
			return s.restart()
		}
	default:
		return fmt.Errorf("未知命令: %s\n%s", args[0], usage)
	}
}

func (s *Server) run(stop <-chan struct{}) {
	defer func() {
		if err := bootstrap.Close(); err != nil {
			log.Printf("释放程序资源失败: %v", err)
		}
	}()
	bootstrap.Load()
	bootstrap.LoadContainer()
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
	password, err := s.readNewPassword("请输入登录密码: ", "请再次输入登录密码: ")
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
	return nil
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

func (s *Server) stop() error {
	log.Println("正在停止服务...")
	if err := s.daemon.Stop(); err != nil {
		return fmt.Errorf("停止失败: %w", err)
	}
	log.Println("服务已停止")
	return nil
}

func (s *Server) restart() error {
	log.Println("正在重启服务...")
	if err := s.daemon.Reload(); err != nil {
		return fmt.Errorf("重启失败: %w", err)
	}
	log.Println("服务已重启")
	return nil
}
