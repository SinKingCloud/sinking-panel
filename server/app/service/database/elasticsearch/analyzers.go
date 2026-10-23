package elasticsearch

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Analyzers 返回当前索引可选的自定义、插件及内置分析器。
func (m *Manager) Analyzers(ctx context.Context) ([]map[string]interface{}, error) {
	settings, err := m.Settings(ctx)
	if err != nil {
		return nil, err
	}
	names := make(map[string]bool)
	for name := range object(object(indexSetting(settings, "analysis"))["analyzer"]) {
		names[name] = true
	}
	// 旧版本和 flat_settings 响应使用平铺键名。
	if index := object(settings["index"]); index != nil {
		settings = index
	}
	for key := range settings {
		key = strings.TrimPrefix(key, "index.")
		if name, ok := strings.CutPrefix(key, "analysis.analyzer."); ok {
			name, _, _ = strings.Cut(name, ".")
			if name != "" {
				names[name] = true
			}
		}
	}
	custom := make([]string, 0, len(names))
	for name := range names {
		custom = append(custom, name)
	}
	sort.Strings(custom)
	options := make([][2]string, 0, len(custom))
	for _, name := range custom {
		options = append(options, [2]string{name, name})
	}
	groups := make([]map[string]interface{}, 0, 3)
	seen := make(map[string]bool)
	addGroup := func(label string, options [][2]string) {
		items := make([]map[string]string, 0, len(options))
		for _, option := range options {
			if !seen[option[0]] {
				seen[option[0]] = true
				items = append(items, map[string]string{"value": option[0], "label": option[1]})
			}
		}
		if len(items) > 0 {
			groups = append(groups, map[string]interface{}{"label": label, "options": items})
		}
	}
	addGroup("索引自定义", options)

	// 节点信息属于集群权限；无权限时仍返回索引自定义和内置分析器。
	pluginCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	data, _, pluginErr := m.doRequest(pluginCtx, http.MethodGet, "/_nodes/plugins?timeout=5s", nil)
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	nodes := object(object(data)["nodes"])
	installed := func(name string) bool {
		if pluginErr != nil || len(nodes) == 0 || number(object(object(data)["_nodes"])["failed"]) > 0 {
			return false
		}
		for _, node := range nodes {
			plugins, _ := object(node)["plugins"].([]interface{})
			found := false
			for _, plugin := range plugins {
				if object(plugin)["name"] == name {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}
	options = nil
	if installed("analysis-ik") {
		options = append(options, [2]string{"ik_smart", "ik_smart（中文智能分词）"}, [2]string{"ik_max_word", "ik_max_word（中文细粒度分词）"})
	}
	if installed("analysis-pinyin") {
		options = append(options, [2]string{"pinyin", "pinyin（拼音）"})
	}
	addGroup("插件分析器", options)
	options = [][2]string{
		{"standard", "standard（标准）"},
		{"simple", "simple（简单）"},
		{"whitespace", "whitespace（空白分词）"},
		{"stop", "stop（停用词）"},
		{"keyword", "keyword（完整内容）"},
		{"pattern", "pattern（正则分词）"},
		{"cjk", "cjk（中日韩）"},
		{"english", "english（英语）"},
	}
	if m.major >= 5 {
		options = append(options, [2]string{"fingerprint", "fingerprint（指纹）"})
	}
	addGroup("内置分析器", options)
	return groups, nil
}
