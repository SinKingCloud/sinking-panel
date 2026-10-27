package software

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"server/public"
)

// Init 创建可直接编辑和安装的插件工程，SDK 从当前二进制公开的 API 生成。
func (s *service) Init(directory, id, name string) (err error) {
	if s.plugins == nil {
		return errors.New("插件管理器未初始化")
	}
	if !regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`).MatchString(id) {
		return errors.New("插件标识须以小写字母开头，只能包含小写字母、数字、下划线和短横线，最长64位")
	}
	if name = strings.TrimSpace(name); name == "" {
		name = id
	}
	if len(name) > 100 {
		return errors.New("插件名称不能超过100字节")
	}
	directory, err = filepath.Abs(directory)
	if err != nil {
		return err
	}
	if _, err = os.Lstat(directory); err == nil {
		return errors.New("目标目录已存在，请选择一个新目录")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(directory), 0755); err != nil {
		return err
	}
	if err = os.Mkdir(directory, 0755); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(directory)
		}
	}()
	if err = s.plugins.WriteSDK(filepath.Join(directory, "sdk")); err != nil {
		return fmt.Errorf("生成插件 SDK 失败: %w", err)
	}
	sdkModule, err := os.ReadFile(filepath.Join(directory, "sdk", "go.mod"))
	if err != nil {
		return err
	}
	goVersion := "1.23.0"
	for line := range strings.SplitSeq(string(sdkModule), "\n") {
		if version, ok := strings.CutPrefix(line, "go "); ok {
			goVersion = strings.TrimSpace(version)
			break
		}
	}
	data := map[string]string{
		"ID":        id,
		"Name":      name,
		"Directory": directory,
		"GoVersion": goVersion,
	}
	return fs.WalkDir(public.Plugin, "plugin", func(source string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		content, err := public.Plugin.ReadFile(source)
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(source, "plugin/")
		if strings.HasSuffix(name, ".tmpl") {
			// JSX 的对象属性使用 {{ }}，模板采用独立分隔符，避免解析前端代码。
			file, err := template.New(name).Delims("[[", "]]").Option("missingkey=error").Funcs(template.FuncMap{
				"quote": strconv.Quote,
				"json":  func(value string) string { encoded, _ := json.Marshal(value); return string(encoded) },
			}).Parse(string(content))
			if err != nil {
				return err
			}
			var output bytes.Buffer
			if err = file.Execute(&output, data); err != nil {
				return err
			}
			content = output.Bytes()
			name = strings.TrimSuffix(name, ".tmpl")
		}
		if strings.HasSuffix(name, ".go") {
			content, err = format.Source(content)
			if err != nil {
				return err
			}
		}
		path := filepath.Join(directory, filepath.FromSlash(name))
		if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		return os.WriteFile(path, content, 0644)
	})
}
