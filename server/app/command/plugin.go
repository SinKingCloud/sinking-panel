package command

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"server/app/constant"
	"server/app/enum/software_status"
	repositorySoftware "server/app/repository/software"
	"server/app/service"
	serviceSoftware "server/app/service/software"
	"server/app/service/software/sdk"
	"server/app/util/container"
	"server/app/util/database"
	"server/app/util/ipc"
	"server/app/util/page"
	"server/app/util/platform"
	"server/bootstrap"
	"server/global"
)

const pluginUsage = `插件命令:
  server plugin list
  server plugin init ID [DIR] [--name 名称]
  server plugin install ZIP或目录
  server plugin uninstall 标识或ID

init 默认创建以 ID 命名的新目录，包含插件示例和当前版本 SDK。
init 不需要启动面板或管理员权限，也不读取面板数据库。
其余命令从面板工作目录执行；安装目录会自动打包。
面板运行时通过本机通信安装、卸载，停止时直接调用软件服务；list 直接读取列表。
uninstall 同时删除插件数据。`

// executePlugin 在守护进程、权限检查和数据库初始化之前处理插件命令。
func (s *Server) executePlugin(args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		fmt.Println(pluginUsage)
		return nil
	}
	action := args[0]
	if action != "list" && action != "init" && action != "install" && action != "uninstall" {
		return fmt.Errorf("未知插件命令: %s\n%s", action, pluginUsage)
	}
	options := flag.NewFlagSet("plugin "+action, flag.ContinueOnError)
	options.SetOutput(io.Discard)
	var name string
	if action == "init" {
		options.StringVar(&name, "name", "", "插件名称")
	}
	// flag 默认只解析位置参数前的选项，这里允许两种顺序，保持命令使用自然。
	var flags, positional []string
	for i := 1; i < len(args); i++ {
		argument := args[i]
		if argument == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if strings.HasPrefix(argument, "-") {
			flags = append(flags, argument)
			if (argument == "--name" || argument == "-name") && action == "init" && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
		} else {
			positional = append(positional, argument)
		}
	}
	if err := options.Parse(flags); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Println(pluginUsage)
			return nil
		}
		return fmt.Errorf("插件命令参数错误: %w", err)
	}
	minimum, maximum := 1, 1
	switch action {
	case "list":
		minimum, maximum = 0, 0
	case "init":
		maximum = 2
	}
	if options.NArg() != 0 || len(positional) < minimum || len(positional) > maximum {
		return errors.New(pluginUsage)
	}
	if action == "init" {
		id, directory := positional[0], positional[0]
		if len(positional) == 2 {
			directory = positional[1]
		}
		if name == "" {
			name = id
		}
		software := serviceSoftware.NewService(nil, nil, sdk.NewManager(constant.SoftwarePath, nil))
		defer software.Close()
		if err := software.Init(directory, id, name); err != nil {
			return err
		}
		absolute, err := filepath.Abs(directory)
		if err != nil {
			return err
		}
		fmt.Printf("插件已创建: %s\n", absolute)
		return nil
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	var target string
	if len(positional) > 0 {
		target = positional[0]
	}
	if action == "install" {
		var err error
		target, err = filepath.Abs(target)
		if err != nil {
			return err
		}
	}
	if action != "list" {
		handled, message, err := s.sendPluginCommand(ctx, action, target)
		if err != nil {
			return err
		}
		if handled {
			fmt.Println(message)
			return nil
		}
		// 面板与离线插件命令共用单例端点，覆盖初始化和退出清理期间。
		guard, err := ipc.Listen(constant.DaemonPath)
		if err != nil {
			return fmt.Errorf("无法离线操作插件，面板或其他插件命令可能正在运行: %w", err)
		}
		defer guard.Close()
	}

	path, err := filepath.Abs(filepath.Join(constant.DBPath, constant.DBFile))
	if err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return errors.New("面板数据库不存在或不可读取，请在面板工作目录先初始化面板")
	}
	// 只打开现有数据库，不同步表结构，不触及插件自行创建的数据表。
	mode := "rw"
	if action == "list" {
		mode = "ro"
	}
	// Windows 盘符需要位于 URI 路径内，不能被识别为 file:// 的主机名。
	dsn := filepath.ToSlash(path)
	if !strings.HasPrefix(dsn, "/") {
		dsn = "/" + dsn
	}
	db := database.NewSqlite((&url.URL{
		Scheme:   "file",
		Path:     dsn,
		RawQuery: "mode=" + mode,
	}).String())
	if db.DbError != nil {
		return fmt.Errorf("读取面板数据库失败: %w", db.DbError)
	}
	global.App.Database = db
	defer bootstrap.Close()
	if action == "list" {
		software := serviceSoftware.NewService(repositorySoftware.NewRepository(db), nil, nil)
		writer := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		defer writer.Flush()
		statuses := software_status.Map()
		fmt.Fprintln(writer, "ID\t标识\t名称\t版本\t状态")
		for number := 1; ; number++ {
			data, err := software.Select(nil, &page.Query{
				Page:         number,
				PageSize:     100,
				OrderByField: "id",
				OrderByType:  "asc",
			})
			if err != nil {
				return err
			}
			for _, record := range data.List {
				if record == nil {
					continue
				}
				status := statuses[record.Status]
				if status == "" {
					status = strconv.Itoa(record.Status)
				}
				fmt.Fprintf(writer, "%d\t%s\t%s\t%s\t%s\n", record.Id, record.Code, record.Name, record.Version, status)
			}
			if len(data.List) < 100 || (data.Total != nil && int64(number*100) >= *data.Total) {
				break
			}
		}
		return nil
	}

	// 只有离线安装、卸载需要加载可写业务服务；list 保持只读。
	bootstrap.LoadConf()
	bootstrap.LoadLog()
	bootstrap.LoadCache()
	if platform.CanRunContainer() {
		executable, err := os.Executable()
		if err != nil {
			return err
		}
		root, err := filepath.EvalSymlinks(filepath.Dir(executable))
		if err != nil {
			return err
		}
		manager, err := container.NewManager(filepath.Join(root, constant.ContainerPath), container.ManagerOptions{DisableAutoRestart: true})
		if err != nil {
			return fmt.Errorf("初始化插件容器服务失败: %w", err)
		}
		global.App.Container = manager
	}
	service.Init()
	service.System.Start(false)
	defer func() {
		service.Software.Close()
		service.System.Close()
		service.Task.Close()
	}()
	message, err := s.runPluginCommand(service.Software, action, target)
	if err != nil {
		return err
	}
	fmt.Println(message)
	return nil
}
