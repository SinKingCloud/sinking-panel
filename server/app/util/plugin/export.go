package plugin

import (
	"fmt"
	"path"
	"reflect"
	"runtime/debug"
	webContext "server/app/util/context"
	"strings"
	"unicode"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// Exports 从应用对象自动发现公开类型和方法，业务调用直接使用原生实例。
// sdk 包提供应用类型的便捷名称，第三方类型保留原导入路径；包级函数可用 Options.Symbols 扩展。
func Exports(app interface{}) (interp.Exports, error) {
	symbols := interp.Exports{"sdk/sdk": {
		"Plugin":  reflect.ValueOf((*Plugin)(nil)),
		"Context": reflect.ValueOf((*webContext.Context)(nil)),
	}}
	root := reflect.TypeOf(app)
	if root == nil {
		root = reflect.TypeOf(Plugin{})
	}
	for root.Kind() == reflect.Pointer {
		root = root.Elem()
	}
	module := root.PkgPath()
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dependency := range append([]*debug.Module{&info.Main}, info.Deps...) {
			if dependency.Path != "" && (module == dependency.Path || strings.HasPrefix(root.PkgPath(), dependency.Path+"/")) {
				module = dependency.Path
				break
			}
		}
	}
	names := map[reflect.Type]string{
		reflect.TypeOf(Plugin{}):             "Plugin",
		reflect.TypeOf(webContext.Context{}): "Context",
	}
	title := func(value string) string {
		parts := strings.FieldsFunc(value, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
		for i, part := range parts {
			runes := []rune(part)
			runes[0] = unicode.ToUpper(runes[0])
			parts[i] = string(runes)
		}
		return strings.Join(parts, "")
	}
	var nameOf func(reflect.Type) string
	nameOf = func(t reflect.Type) string {
		if name := names[t]; name != "" {
			return name
		}
		if t.Kind() == reflect.Pointer {
			return nameOf(t.Elem())
		}
		name, pkg := t.Name(), t.PkgPath()
		if pkg == "" || pkg == root.PkgPath() {
			return title(name)
		}
		if strings.Contains(name, "[") {
			// 已实例化的泛型分页按元素类型命名，解释器无需处理泛型语法。
			if t.Kind() == reflect.Struct {
				if field, ok := t.FieldByName("List"); ok && field.Type.Kind() == reflect.Slice {
					return nameOf(field.Type.Elem()) + "Page"
				}
			}
		}
		prefix := title(path.Base(pkg))
		if prefix == "Types" {
			prefix = "Type"
		}
		switch {
		case path.Base(pkg) == "model":
			name = "Model" + name
		case path.Base(path.Dir(pkg)) == "repository":
			if !strings.HasPrefix(name, "Select") && !strings.HasPrefix(name, "Update") {
				name += "ListItem"
			}
		case path.Base(path.Dir(pkg)) == "util":
			name = "Util" + prefix + name
		case path.Base(path.Dir(pkg)) == "service" && name == prefix:
			name += "Record"
		case !strings.HasPrefix(name, prefix):
			name = prefix + name
		}
		return title(name)
	}
	seen := make(map[reflect.Type]bool)
	var visit func(reflect.Type) error
	visit = func(t reflect.Type) error {
		if seen[t] {
			return nil
		}
		seen[t] = true
		if t.Name() != "" && t.PkgPath() != "" {
			pkg := t.PkgPath()
			packageName, _, ok := strings.Cut(t.String(), ".")
			if !ok {
				packageName = path.Base(pkg)
			}
			native := pkg + "/" + packageName
			if symbols[native] == nil {
				symbols[native] = make(map[string]reflect.Value)
			}
			symbols[native][t.Name()] = reflect.Zero(reflect.PointerTo(t))
			if pkg == module || strings.HasPrefix(pkg, module+"/") {
				name := nameOf(t)
				if value, exists := symbols["sdk/sdk"][name]; exists && value.Type() != reflect.PointerTo(t) {
					return fmt.Errorf("插件导出名称 %s 冲突：%s、%s", name, value.Type(), t)
				}
				names[t] = name
				symbols["sdk/sdk"][name] = reflect.Zero(reflect.PointerTo(t))
			}
			// 标准库已有完整导出；替换被注入的类型即可，避免继续展开运行时内部类型。
			if _, standard := stdlib.Symbols[native]; standard {
				return nil
			}
		}
		types := []reflect.Type{}
		switch t.Kind() {
		case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Chan:
			types = append(types, t.Elem())
		case reflect.Map:
			types = append(types, t.Key(), t.Elem())
		case reflect.Func:
			for i := 0; i < t.NumIn(); i++ {
				types = append(types, t.In(i))
			}
			for i := 0; i < t.NumOut(); i++ {
				types = append(types, t.Out(i))
			}
		case reflect.Struct:
			for i := 0; i < t.NumField(); i++ {
				if field := t.Field(i); field.IsExported() {
					types = append(types, field.Type)
				}
			}
		}
		if t.Kind() != reflect.Pointer {
			for i := 0; i < t.NumMethod(); i++ {
				types = append(types, t.Method(i).Type)
			}
			if t.Kind() != reflect.Interface {
				pointer := reflect.PointerTo(t)
				for i := 0; i < pointer.NumMethod(); i++ {
					types = append(types, pointer.Method(i).Type)
				}
			}
		}
		for _, child := range types {
			if err := visit(child); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(reflect.TypeOf(Plugin{})); err != nil {
		return nil, err
	}
	if app != nil {
		return symbols, visit(reflect.TypeOf(app))
	}
	return symbols, nil
}
