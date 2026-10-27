package plugin

import (
	"fmt"
	"go/format"
	"go/version"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/traefik/yaegi/stdlib"
)

// WriteSDK 从管理器注入的应用类型生成独立开发声明，不读取宿主源码或执行 Go 命令。
// directory 必须尚不存在；生成失败会清理本次创建的目录。
func (m *Manager) WriteSDK(directory string) error {
	var app interface{}
	if m.options.App != nil {
		app = m.options.App("")
	}
	symbols, err := Exports(app)
	if err != nil {
		return err
	}
	declared := make(map[reflect.Type]string)
	types := make(map[string]reflect.Type)
	for name, value := range symbols["sdk/sdk"] {
		if value.Kind() != reflect.Pointer || !value.IsNil() {
			continue
		}
		t := value.Type().Elem()
		types[name] = t
		if previous := declared[t]; previous == "" || name < previous {
			declared[t] = name
		}
	}
	names := make([]string, 0, len(types))
	for name := range types {
		names = append(names, name)
	}
	sort.Strings(names)
	publicImports := make(map[string]bool)
	generate := func(methods bool) ([]byte, error) {
		imports := make(map[string]string)
		aliases := make(map[string]bool)
		for name := range types {
			aliases[name] = true
		}
		importName := func(pkg, name string) string {
			if alias := imports[pkg]; alias != "" {
				return alias
			}
			if name == "" {
				name = path.Base(pkg)
			}
			base := name
			for suffix := 2; aliases[name]; suffix++ {
				name = base + strconv.Itoa(suffix)
			}
			imports[pkg], aliases[name], publicImports[pkg] = name, true, true
			return name
		}
		var reference func(reflect.Type) string
		var signature func(reflect.Type, int) string
		var underlying func(reflect.Type) string
		signature = func(t reflect.Type, skip int) string {
			var inputs, outputs []string
			for i := skip; i < t.NumIn(); i++ {
				parameter := reference(t.In(i))
				if t.IsVariadic() && i == t.NumIn()-1 {
					parameter = "..." + reference(t.In(i).Elem())
				}
				inputs = append(inputs, "arg"+strconv.Itoa(i-skip)+" "+parameter)
			}
			for i := 0; i < t.NumOut(); i++ {
				outputs = append(outputs, reference(t.Out(i)))
			}
			result := "(" + strings.Join(inputs, ", ") + ")"
			if len(outputs) == 1 {
				result += " " + outputs[0]
			} else if len(outputs) > 1 {
				result += " (" + strings.Join(outputs, ", ") + ")"
			}
			return result
		}
		reference = func(t reflect.Type) string {
			if name := declared[t]; name != "" {
				return name
			}
			if t.Name() != "" {
				if t.PkgPath() == "" {
					return t.Name()
				}
				packageName, _, _ := strings.Cut(t.String(), ".")
				return importName(t.PkgPath(), packageName) + "." + t.Name()
			}
			return underlying(t)
		}
		underlying = func(t reflect.Type) string {
			switch t.Kind() {
			case reflect.Pointer:
				return "*" + reference(t.Elem())
			case reflect.Slice:
				return "[]" + reference(t.Elem())
			case reflect.Array:
				return "[" + strconv.Itoa(t.Len()) + "]" + reference(t.Elem())
			case reflect.Map:
				return "map[" + reference(t.Key()) + "]" + reference(t.Elem())
			case reflect.Chan:
				prefix := "chan "
				if t.ChanDir() == reflect.RecvDir {
					prefix = "<-chan "
				} else if t.ChanDir() == reflect.SendDir {
					prefix = "chan<- "
				}
				return prefix + reference(t.Elem())
			case reflect.Func:
				return "func" + signature(t, 0)
			case reflect.Interface:
				if t.NumMethod() == 0 {
					return "interface{}"
				}
				var body strings.Builder
				body.WriteString("interface {\n")
				for i := 0; i < t.NumMethod(); i++ {
					method := t.Method(i)
					if method.PkgPath == "" {
						body.WriteString(method.Name + signature(method.Type, 0) + "\n")
					}
				}
				body.WriteString("}")
				return body.String()
			case reflect.Struct:
				if t.PkgPath() != "time" && t.ConvertibleTo(reflect.TypeOf(time.Time{})) {
					return importName("time", "time") + ".Time"
				}
				var body strings.Builder
				body.WriteString("struct {\n")
				for _, field := range reflect.VisibleFields(t) {
					if !field.IsExported() {
						continue
					}
					// 改名的嵌入类型保留原字段名，并补齐可见的提升字段供返回值读取。
					if len(field.Index) > 1 {
						root := t.Field(field.Index[0])
						embedded := root.Type
						if embedded.Kind() == reflect.Pointer {
							embedded = embedded.Elem()
						}
						if declared[embedded] == root.Name || declared[embedded] == "" {
							continue
						}
						body.WriteString("// 提升字段，仅用于读取原生返回值。\n")
					}
					fieldType := reference(field.Type)
					embedded := field.Type
					if embedded.Kind() == reflect.Pointer {
						embedded = embedded.Elem()
					}
					if field.Anonymous && (declared[embedded] == field.Name || declared[embedded] == "") {
						body.WriteString(fieldType)
					} else {
						body.WriteString(field.Name + " " + fieldType)
					}
					if field.Tag != "" {
						if strings.ContainsRune(string(field.Tag), '`') {
							body.WriteString(" " + strconv.Quote(string(field.Tag)))
						} else {
							body.WriteString(" `" + string(field.Tag) + "`")
						}
					}
					body.WriteString("\n")
				}
				body.WriteString("}")
				return body.String()
			case reflect.UnsafePointer:
				return importName("unsafe", "unsafe") + ".Pointer"
			default:
				return t.Kind().String()
			}
		}
		var body strings.Builder
		for _, name := range names {
			t := types[name]
			if declared[t] != name {
				if !methods {
					body.WriteString("type " + name + " = " + declared[t] + "\n\n")
				}
				continue
			}
			if !methods {
				body.WriteString("// " + name + " 是面板公开 API 的开发声明。\n")
				body.WriteString("type " + name + " " + underlying(t) + "\n\n")
				continue
			}
			if t.Kind() == reflect.Interface {
				continue
			}
			pointer := reflect.PointerTo(t)
			for i := 0; i < pointer.NumMethod(); i++ {
				method := pointer.Method(i)
				receiver := "*" + name
				if valueMethod, ok := t.MethodByName(method.Name); ok {
					method, receiver = valueMethod, name
				}
				body.WriteString("func (" + receiver + ") " + method.Name + signature(method.Type, 1) + " { panic(nativeOnly) }\n\n")
			}
		}
		var code strings.Builder
		code.WriteString("// Code generated by server plugin init; DO NOT EDIT.\n\npackage sdk\n\n")
		if len(imports) > 0 {
			paths := make([]string, 0, len(imports))
			for pkg := range imports {
				paths = append(paths, pkg)
			}
			sort.Strings(paths)
			code.WriteString("import (\n")
			for _, pkg := range paths {
				code.WriteString(imports[pkg] + " " + strconv.Quote(pkg) + "\n")
			}
			code.WriteString(")\n\n")
		}
		code.WriteString(body.String())
		return format.Source([]byte(code.String()))
	}
	entities, err := generate(false)
	if err != nil {
		return fmt.Errorf("生成 SDK 类型声明失败: %w", err)
	}
	methods, err := generate(true)
	if err != nil {
		return fmt.Errorf("生成 SDK 方法声明失败: %w", err)
	}
	build, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Errorf("当前程序缺少 Go 依赖信息，无法生成独立 SDK")
	}
	dependencies := make(map[string]*debug.Module)
	for pkg := range publicImports {
		standard := pkg == "unsafe"
		for native := range stdlib.Symbols {
			if strings.HasPrefix(native, pkg+"/") && !strings.Contains(strings.TrimPrefix(native, pkg+"/"), "/") {
				standard = true
				break
			}
		}
		if standard {
			continue
		}
		var dependency *debug.Module
		for _, candidate := range build.Deps {
			if (pkg == candidate.Path || strings.HasPrefix(pkg, candidate.Path+"/")) && (dependency == nil || len(candidate.Path) > len(dependency.Path)) {
				dependency = candidate
			}
		}
		if dependency == nil || dependency.Version == "" || dependency.Version == "(devel)" {
			return fmt.Errorf("公开 SDK 类型仍引用未导出的宿主包或缺少版本的依赖: %s", pkg)
		}
		if dependency.Replace != nil && dependency.Replace.Version == "" {
			return fmt.Errorf("SDK 依赖 %s 使用本地替换，无法生成公开模块", dependency.Path)
		}
		dependencies[dependency.Path] = dependency
	}
	goVersion := strings.TrimPrefix(version.Lang(runtime.Version()), "go")
	if goVersion == "" {
		return fmt.Errorf("无法识别当前 Go 版本: %s", runtime.Version())
	}
	var module strings.Builder
	module.WriteString("module sdk\n\ngo " + goVersion + ".0\n")
	dependencyPaths := make([]string, 0, len(dependencies))
	for pkg := range dependencies {
		dependencyPaths = append(dependencyPaths, pkg)
	}
	sort.Strings(dependencyPaths)
	if len(dependencyPaths) > 0 {
		module.WriteString("\nrequire (\n")
		for _, pkg := range dependencyPaths {
			module.WriteString("\t" + pkg + " " + dependencies[pkg].Version + "\n")
		}
		module.WriteString(")\n")
		for _, pkg := range dependencyPaths {
			if replacement := dependencies[pkg].Replace; replacement != nil {
				module.WriteString("\nreplace " + pkg + " => " + replacement.Path + " " + replacement.Version + "\n")
			}
		}
	}
	files := map[string][]byte{
		"go.mod":      []byte(module.String()),
		"entities.go": entities,
		"methods.go":  methods,
		"sdk.go":      []byte("// Package sdk 提供插件开发补全与类型检查；运行时由面板注入真实类型和服务。\npackage sdk\n\nconst nativeOnly = \"软件 SDK 只能在面板插件运行时调用\"\n"),
	}
	if err := os.Mkdir(directory, 0755); err != nil {
		return fmt.Errorf("创建 SDK 目录失败: %w", err)
	}
	complete := false
	defer func() {
		if !complete {
			_ = os.RemoveAll(directory)
		}
	}()
	for _, name := range []string{"go.mod", "entities.go", "methods.go", "sdk.go"} {
		file, err := os.OpenFile(filepath.Join(directory, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return err
		}
		_, writeErr := file.Write(files[name])
		closeErr := file.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	complete = true
	return nil
}
