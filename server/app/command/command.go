package command

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"server/app"
	"server/app/enum/log_type"
	"server/app/service"
	"server/app/util/daemon"
	"server/bootstrap"
	"server/global"
)

const (
	pidFileName = "server.pid"
	logFileName = "server.log"
	serviceName = "sinking-panel"
	usage       = `使用方法:
  server [command] [options]

可用命令:
  start   启动服务
  stop    停止服务
  restart 重启服务
  run     直接运行(非守护进程模式)
  install 设置登录信息并安装系统自启动
  uninstall 卸载软件，保留网站数据并询问是否保留容器
  user    修改登录账号
  pwd     修改登录密码

启动参数（适用于 run、start、restart、install）:
  --mode dev|release  运行模式，默认 release
  --host HOST      监听地址，默认 0.0.0.0
  --port PORT      监听端口，默认 5678
  -h, --help       显示帮助

配置文件可选: data/server/application.yml，不自动创建。
优先级: 命令行参数 > 配置文件 > 默认值。`
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
	server.daemon = d.SetAutoStartOptions(daemon.AutoStartOptions{Name: serviceName})
	return server, nil
}

// Execute 执行服务命令。
func (s *Server) Execute(args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		log.Println(usage)
		return nil
	}
	command := args[0]
	startup := false
	switch command {
	case "run", "start", "restart", "install":
		startup = true
	case "stop", "uninstall", "user", "pwd":
	default:
		return fmt.Errorf("未知命令: %s\n%s", command, usage)
	}
	options := flag.NewFlagSet(command, flag.ContinueOnError)
	options.SetOutput(io.Discard)
	mode := options.String("mode", "release", "运行模式")
	host := options.String("host", "0.0.0.0", "监听地址")
	port := options.Int("port", 5678, "监听端口")
	if err := options.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			log.Println(usage)
			return nil
		}
		return fmt.Errorf("命令参数错误: %w", err)
	}
	if options.NArg() > 0 {
		return fmt.Errorf("无法识别的命令参数: %s", strings.Join(options.Args(), " "))
	}
	if !startup && options.NFlag() > 0 {
		return fmt.Errorf("%s 命令不接受启动参数", command)
	}
	if startup {
		if *mode != "dev" && *mode != "release" {
			return errors.New("mode 只支持 dev 或 release")
		}
		*host = strings.TrimSpace(*host)
		if *host == "" {
			return errors.New("host 不能为空")
		}
		if *port < 1 || *port > 65535 {
			return errors.New("port 必须在 1 到 65535 之间")
		}
		bootstrap.LoadConf()
		arguments := make([]string, 0, options.NFlag())
		options.Visit(func(option *flag.Flag) {
			global.App.Config.Set("server."+option.Name, option.Value.String())
			arguments = append(arguments, "--"+option.Name+"="+option.Value.String())
		})
		s.daemon.SetChildArgs(append([]string{os.Args[0], "start"}, arguments...)...)
		s.daemon.SetAutoStartOptions(daemon.AutoStartOptions{
			Name:      serviceName,
			Arguments: append([]string{"run"}, arguments...),
		})
	}
	switch command {
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
	case "start":
		return s.start()
	case "stop":
		return s.stop()
	case "restart":
		return s.restart()
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
