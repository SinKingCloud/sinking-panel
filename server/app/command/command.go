package command

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"

	"server/app/util/daemon"
	"server/app/util/platform"
)

const usage = `使用方法:
  server [command] [options]

可用命令:
  start   启动服务
  stop    停止服务
  restart 重启服务
  run     直接运行(非守护进程模式)
  install 设置登录信息、安装系统服务并启动
  uninstall 卸载软件，询问是否保留容器、镜像及数据
  user    修改登录账号
  pwd     修改登录密码
  plugin  插件列表、创建、安装和卸载（plugin --help 查看用法）

启动参数（适用于 run、start、restart、install）:
  --mode dev|release  运行模式，默认 release
  --host HOST      监听地址，默认 0.0.0.0
  --port PORT      监听端口，默认 5678
  -h, --help       显示帮助

停止/重启参数（适用于 stop、restart）:
  --containers    同时停止或重启容器，默认保留容器运行

配置文件可选: data/server/application.yml，不自动创建。
优先级: 命令行参数 > 配置文件 > 默认值。`

// Server 统一管理面板和插件命令，零值可直接调用 Execute。
type Server struct {
	daemon *daemon.Daemon
}

// Execute 是命令的统一入口。
// init 必须在创建守护进程之前处理，避免 libcontainer 子进程误进入面板启动流程。
func (s *Server) Execute(args []string) error {
	if s.executeInit(args) {
		return nil
	}
	if len(args) > 0 && args[0] == "plugin" {
		return s.executePlugin(args[1:])
	}
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
	mode, host, port := "release", "0.0.0.0", 5678
	if startup {
		options.StringVar(&mode, "mode", mode, "运行模式")
		options.StringVar(&host, "host", host, "监听地址")
		options.IntVar(&port, "port", port, "监听端口")
	}
	var containers bool
	if command == "stop" || command == "restart" {
		options.BoolVar(&containers, "containers", false, "同时停止或重启容器")
	}
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
	if startup {
		if mode != "dev" && mode != "release" {
			return errors.New("mode 只支持 dev 或 release")
		}
		host = strings.TrimSpace(host)
		if host == "" {
			return errors.New("host 不能为空")
		}
		if port < 1 || port > 65535 {
			return errors.New("port 必须在 1 到 65535 之间")
		}
	}
	// 前台运行不要求提权，系统服务相关命令仍检查权限。
	if command != "run" {
		if err := platform.CheckPrivilege(); err != nil {
			return err
		}
	}
	if err := s.initDaemon(); err != nil {
		return err
	}
	if startup {
		if err := s.configureStartup(command, options); err != nil {
			return err
		}
	}
	switch command {
	case "run":
		log.Println("以前台模式运行服务...")
		return s.daemon.Run()
	case "install":
		return s.install()
	case "uninstall":
		return s.uninstall()
	case "user", "pwd":
		return s.updateAccount(command)
	case "start":
		return s.start()
	case "stop":
		return s.stop(containers)
	case "restart":
		return s.restart(containers)
	}
	return nil
}
