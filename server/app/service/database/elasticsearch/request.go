package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"
)

// Request 执行显式 REST 请求；写请求只作用于当前普通索引，不开放跨索引批量操作。
func (m *Manager) Request(ctx context.Context, method, path, body string) (map[string]interface{}, error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method != http.MethodGet && method != http.MethodHead && method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete {
		return nil, errors.New("REST 方法只支持 GET、HEAD、POST、PUT 和 DELETE")
	}
	if path == "" {
		path = "/"
	}
	if len(path) > 8192 || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.IndexFunc(path, unicode.IsControl) >= 0 || strings.ContainsAny(path, "\\#") {
		return nil, errors.New("REST 路径应为以 / 开头的相对地址，不超过 8192 字节")
	}
	parsed, err := url.Parse(path)
	if err != nil || parsed.Host != "" || parsed.Scheme != "" || parsed.Fragment != "" || parsed.Opaque != "" {
		return nil, errors.New("REST 路径不合法")
	}
	segments := strings.Split(strings.Trim(parsed.EscapedPath(), "/"), "/")
	decoded := make([]string, len(segments))
	for i, segment := range segments {
		decoded[i], err = url.PathUnescape(segment)
		if err != nil || decoded[i] == "." || decoded[i] == ".." || strings.ContainsAny(decoded[i], "/\\%\x00\r\n") {
			return nil, errors.New("REST 路径不能包含目录穿越或重复编码")
		}
	}
	if method != http.MethodGet && method != http.MethodHead {
		if err := ValidateIndexName(m.index, true); err != nil {
			return nil, err
		}
		if len(decoded) == 0 || decoded[0] != m.index {
			return nil, errors.New("写入 REST 请求必须明确限定当前索引")
		}
		if method == http.MethodDelete {
			valid := len(decoded) == 1 || len(decoded) == 3 && (decoded[1] == "_doc" || m.major < 7 && !strings.HasPrefix(decoded[1], "_"))
			if !valid {
				return nil, errors.New("DELETE 仅允许当前具体索引或具体文档，不能批量或通配删除")
			}
			if len(decoded) == 3 {
				if err := validateDocumentID(decoded[2]); err != nil {
					return nil, err
				}
				if m.major < 7 && legacyDeleteEndpoint(decoded[2]) {
					return nil, errors.New("旧版 DELETE 文档路径不能使用保留端点名称，避免删除类型或按查询删除")
				}
			}
		} else if len(decoded) > 1 {
			allowed := map[string]bool{
				"_search":     true,
				"_count":      true,
				"_analyze":    true,
				"_mapping":    true,
				"_settings":   true,
				"_doc":        true,
				"_create":     true,
				"_update":     true,
				"_refresh":    true,
				"_flush":      true,
				"_forcemerge": true,
				"_open":       true,
				"_close":      true,
				"_validate":   true,
				"_explain":    true,
			}
			if !allowed[decoded[1]] && !(m.major < 7 && !strings.HasPrefix(decoded[1], "_") && len(decoded) <= 3) {
				return nil, errors.New("REST 写入路径不支持此操作；批量、跨索引和按查询删除均未开放")
			}
			for _, segment := range decoded {
				if strings.ContainsAny(segment, "*?,") {
					return nil, errors.New("REST 写入路径不能包含通配符或多个目标")
				}
			}
		}
	}
	var content interface{}
	if strings.TrimSpace(body) != "" {
		content, err = decodeObject(body, "REST 请求体", false)
		if err != nil {
			return nil, err
		}
	}
	data, status, err := m.doRequest(ctx, method, parsed.RequestURI(), content)
	if err != nil {
		return nil, err
	}
	// 文档编辑使用原始 JSON 和读取时版本，避免浏览器丢失大整数精度或覆盖并发修改。
	if value := object(data); method == http.MethodGet && value != nil && value["_source"] != nil && value["_id"] != nil {
		source, err := json.MarshalIndent(value["_source"], "", "  ")
		if err != nil {
			return nil, err
		}
		value["_source_json"] = string(source)
		for _, key := range []string{"_seq_no", "_primary_term", "_version"} {
			if value[key] != nil {
				value[key] = strconv.FormatInt(number(value[key]), 10)
			}
		}
	}
	return map[string]interface{}{
		"status": status,
		"data":   data,
	}, nil
}
