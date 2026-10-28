package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
	webContext "server/app/util/context"
	"strings"
	"time"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

// Load 解释并缓存插件，重复加载不会再次执行 New；不满足运行条件时不执行代码。
func (m *Manager) Load(directory string) error {
	i, release, err := m.acquire(context.Background(), directory)
	if err != nil {
		return err
	}
	defer release()
	return m.load(m.ctx, directory, i)
}

func (m *Manager) load(ctx context.Context, directory string, instance *instance) (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = fmt.Errorf("加载插件代码失败: %v", value)
		}
	}()
	if instance.interpreter != nil {
		return m.checkSupport(instance.manifest)
	}
	info, err := m.info(directory)
	if err != nil {
		return err
	}
	if !info.Available {
		return errors.New(info.UnavailableReason)
	}
	root, err := os.OpenRoot(m.root)
	if err != nil {
		return err
	}
	defer root.Close()
	code, err := root.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer func() {
		if instance.root != code {
			_ = code.Close()
		}
	}()
	files, err := fs.ReadDir(code.FS(), "server")
	if err != nil {
		return fmt.Errorf("读取插件代码失败: %w", err)
	}
	packageName := ""
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") || strings.HasSuffix(file.Name(), "_test.go") {
			continue
		}
		path := filepath.Join("server", file.Name())
		source, err := code.ReadFile(path)
		if err != nil {
			return err
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, source, parser.PackageClauseOnly)
		if err != nil {
			return err
		}
		if packageName != "" && packageName != parsed.Name.Name {
			return errors.New("插件入口目录只能包含一个 Go 包")
		}
		packageName = parsed.Name.Name
	}
	if packageName == "" || packageName == "main" {
		return errors.New("插件必须提供带 New 入口的 Go 包")
	}
	var source fs.FS = code.FS()
	entryPath := "./server"
	if data, readErr := code.ReadFile("server/go.mod"); readErr == nil {
		definition, parseErr := modfile.Parse("server/go.mod", data, nil)
		if parseErr != nil {
			return fmt.Errorf("插件 go.mod 无效: %w", parseErr)
		}
		if definition.Module == nil || module.CheckImportPath(definition.Module.Mod.Path) != nil {
			return errors.New("插件 go.mod 缺少有效的 module 路径")
		}
		entryPath = definition.Module.Mod.Path
		source = moduleFS{
			FS:     source,
			module: entryPath,
		}
	} else if !errors.Is(readErr, fs.ErrNotExist) {
		return fmt.Errorf("读取插件 go.mod 失败: %w", readErr)
	}
	i := interp.New(interp.Options{
		GoPath:               ".",
		SourcecodeFilesystem: source,
		Unrestricted:         true,
		Stdout:               io.Discard,
		Stderr:               io.Discard,
	})
	if err = i.Use(stdlib.Symbols); err != nil {
		return err
	}
	// Yaegi 的标准库表止于 Go 1.22；补充目录约束文件 API，供插件安全管理持久文件。
	if err = i.Use(interp.Exports{"os/os": {
		"OpenRoot": reflect.ValueOf(os.OpenRoot),
		"Root":     reflect.ValueOf((*os.Root)(nil)),
	}}); err != nil {
		return err
	}
	var app interface{}
	if m.options.App != nil {
		app = m.options.App(info.Manifest.Id)
		if app == nil {
			return errors.New("插件应用对象不能为空")
		}
	}
	exports, err := Exports(app)
	if err != nil {
		return err
	}
	if err = i.Use(exports); err != nil {
		return err
	}
	if err = i.Use(m.options.Symbols); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if _, err = i.EvalPathWithContext(ctx, entryPath); err != nil {
		if strings.Contains(err.Error(), "unable to find source related") || strings.Contains(err.Error(), "not in GOPATH") {
			return fmt.Errorf("插件缺少依赖源码，请在 server 目录执行 go mod vendor 后重新打包: %w", err)
		}
		return fmt.Errorf("解释插件代码失败: %w", err)
	}
	if _, err = i.EvalWithContext(ctx, fmt.Sprintf("import %s %q", packageName, entryPath)); err != nil {
		return err
	}
	entry, err := i.EvalWithContext(ctx, packageName+".New")
	if err != nil {
		return fmt.Errorf("插件缺少 New 入口: %w", err)
	}
	var arguments []reflect.Value
	if m.options.App != nil {
		arguments = append(arguments, reflect.ValueOf(app))
	}
	if entry.Kind() != reflect.Func || entry.Type().IsVariadic() || entry.Type().NumIn() != len(arguments) || entry.Type().NumOut() != 1 || entry.Type().Out(0) != reflect.TypeOf((*Plugin)(nil)) {
		return errors.New("插件 New 入口必须接收宿主对象并返回 *sdk.Plugin；未配置宿主时不接收参数")
	}
	if len(arguments) > 0 && !arguments[0].Type().AssignableTo(entry.Type().In(0)) {
		return fmt.Errorf("插件 New 入口参数为 %s，宿主提供的类型为 %s", entry.Type().In(0), arguments[0].Type())
	}
	created := entry.Call(arguments)[0].Interface().(*Plugin)
	if created == nil {
		return errors.New("插件 New 入口返回了空对象")
	}
	instance.manifest, instance.plugin, instance.interpreter, instance.root = info.Manifest, created, i, code
	return nil
}

// Call 使用面板上下文调用业务接口，参数绑定和响应由插件自行处理。
func (m *Manager) Call(ctx context.Context, directory, method string, c *webContext.Context) (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = fmt.Errorf("插件执行失败: %v", value)
		}
	}()
	if c == nil || c.Context == nil || c.Request == nil {
		return errors.New("插件请求上下文不能为空")
	}
	i, release, err := m.acquire(ctx, directory)
	if err != nil {
		return err
	}
	defer release()
	ctx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(m.ctx, cancel)
	defer func() { stop(); cancel() }()
	if err = m.load(ctx, directory, i); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	action := i.plugin.Actions[method]
	if action == nil {
		return fmt.Errorf("插件不支持 %s 操作", method)
	}
	request := c.Request
	current := request.WithContext(ctx)
	c.Request = current
	defer func() {
		request.MultipartForm = current.MultipartForm
		c.Request = request
	}()
	action(c)
	return nil
}

// Lifecycle 调用宿主指定的生命周期；卸载可加载未运行的插件，损坏的代码不阻塞移除。
func (m *Manager) Lifecycle(ctx context.Context, directory, event, params string) (result interface{}, err error) {
	defer func() {
		if value := recover(); value != nil {
			result, err = nil, fmt.Errorf("插件执行失败: %v", value)
		}
	}()
	if params == "" {
		params = "{}"
	}
	var object map[string]json.RawMessage
	if json.Unmarshal([]byte(params), &object) != nil || object == nil {
		return nil, errors.New("插件参数必须是 JSON 对象")
	}
	i, release, err := m.acquire(ctx, directory)
	if err != nil {
		return nil, err
	}
	defer release()
	ctx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(m.ctx, cancel)
	defer func() { stop(); cancel() }()
	if event == "stop" {
		return i.stop(ctx, params)
	}
	if event == "uninstall" {
		if i.plugin == nil {
			if err = m.load(ctx, directory, i); err != nil {
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}
				log.Printf("插件 %s 无法执行数据清理，继续移除代码: %v", directory, err)
				return nil, nil
			}
		}
	} else if err = m.load(ctx, directory, i); err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	var action func(context.Context, string) (interface{}, error)
	switch event {
	case "start":
		return i.start(ctx, m.ctx, params)
	case "install":
		action = i.plugin.Install
	case "uninstall":
		action = i.plugin.Uninstall
	case "upgrade":
		action = i.plugin.Upgrade
	default:
		return nil, fmt.Errorf("插件生命周期 %s 不支持", event)
	}
	if action == nil {
		return nil, nil
	}
	return action(ctx, params)
}

// start 的运行上下文独立于启动请求；请求结束后后台任务仍可继续运行。
func (i *instance) start(ctx, parent context.Context, params string) (result interface{}, err error) {
	if i.runCancel != nil {
		if i.runContext.Err() != nil {
			return nil, errors.New("插件上次停止尚未完成，请先重试停止")
		}
		return nil, nil
	}
	running, cancel := context.WithCancel(parent)
	i.runCancel = cancel
	i.runContext = running
	canceled := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		cancel()
		close(canceled)
	})
	defer func() {
		if !stop() {
			<-canceled
		}
		if value := recover(); value != nil {
			err = fmt.Errorf("启动插件失败: %v", value)
		}
		if err == nil {
			err = errors.Join(ctx.Err(), running.Err())
		}
		if err != nil {
			cleanup, release := context.WithTimeout(context.Background(), 30*time.Second)
			defer release()
			_, stopErr := i.stop(cleanup, "{}")
			err = errors.Join(err, stopErr)
		}
	}()
	if i.plugin.Start != nil {
		return i.plugin.Start(running, params)
	}
	return nil, nil
}

// stop 先通知后台任务退出，再执行清理；重复停止不会重复调用钩子。
func (i *instance) stop(ctx context.Context, params string) (result interface{}, err error) {
	if i.runCancel == nil {
		return nil, nil
	}
	i.runCancel()
	defer func() {
		if value := recover(); value != nil {
			err = fmt.Errorf("停止插件失败: %v", value)
		}
		if err == nil {
			i.runCancel, i.runContext = nil, nil
		}
	}()
	if i.plugin.Stop != nil {
		return i.plugin.Stop(ctx, params)
	}
	return nil, nil
}
