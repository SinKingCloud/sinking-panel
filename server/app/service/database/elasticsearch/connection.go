package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const maxPayload = 8 << 20

// ValidateIndexName 只允许一个具体索引；写操作拒绝隐藏及系统索引。
// 此方法不访问连接，Manager 零值也可用于提交任务前的参数校验。
func (m *Manager) ValidateIndexName(name string, writable bool) error {
	if name == "" || len(name) > 255 || !utf8.ValidString(name) || strings.ToLower(name) != name || name == "." || name == ".." ||
		strings.ContainsAny(name, "\\/*?\"<>| ,#:%") || strings.ContainsAny(name[:1], "_+-") || strings.IndexFunc(name, func(r rune) bool { return unicode.IsControl(r) || unicode.IsSpace(r) }) >= 0 {
		return errors.New("索引名应为小写的具体名称，不超过 255 字节，不能包含通配符、路径或空白")
	}
	if writable && strings.HasPrefix(name, ".") {
		return errors.New("不能修改或删除 Elasticsearch 系统索引")
	}
	return nil
}

func (m *Manager) indexPath(writable bool) (string, error) {
	if err := m.ValidateIndexName(m.index, writable); err != nil {
		return "", err
	}
	return "/" + url.PathEscape(m.index), nil
}

func (m *Manager) decodeObject(text, label string, optional bool) (map[string]interface{}, error) {
	if strings.TrimSpace(text) == "" && optional {
		return map[string]interface{}{}, nil
	}
	if len(text) > maxPayload || !json.Valid([]byte(text)) {
		return nil, errors.New(label + "必须是 JSON 对象且不超过 8 MB")
	}
	var value map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()
	if decoder.Decode(&value) != nil || value == nil {
		return nil, errors.New(label + "必须是 JSON 对象")
	}
	return value, nil
}

func (m *Manager) number(value interface{}) int64 {
	switch value := value.(type) {
	case json.Number:
		result, _ := value.Int64()
		return result
	case string:
		result, _ := strconv.ParseInt(value, 10, 64)
		return result
	case float64:
		return int64(value)
	case int:
		return int64(value)
	case int64:
		return value
	}
	return 0
}

func (m *Manager) object(value interface{}) map[string]interface{} {
	result, _ := value.(map[string]interface{})
	return result
}

// doRequest 统一限时、限制响应体并脱敏错误，写请求不自动重试或跳转。
func (m *Manager) doRequest(ctx context.Context, method, relative string, body interface{}) (interface{}, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var content []byte
	var err error
	if body != nil {
		content, err = json.Marshal(body)
		if err != nil || len(content) > maxPayload {
			return nil, 0, errors.New("Elasticsearch 请求体不合法或超过 8 MB")
		}
	}
	address := strings.TrimRight(m.base.String(), "/") + relative
	request, err := http.NewRequestWithContext(ctx, method, address, bytes.NewReader(content))
	if err != nil {
		return nil, 0, m.cleanError(err)
	}
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if m.apiKey != "" {
		request.Header.Set("Authorization", "ApiKey "+m.apiKey)
	} else if m.user != "" || m.password != "" {
		request.SetBasicAuth(m.user, m.password)
	}
	response, err := m.client.Do(request)
	if err != nil {
		return nil, 0, m.cleanError(err)
	}
	defer response.Body.Close()
	content, err = io.ReadAll(io.LimitReader(response.Body, maxPayload+1))
	if err != nil {
		return nil, response.StatusCode, m.cleanError(err)
	}
	if len(content) > maxPayload {
		return nil, response.StatusCode, errors.New("Elasticsearch 响应超过 8 MB，请缩小查询或字段范围")
	}
	var result interface{}
	if len(content) != 0 {
		decoder := json.NewDecoder(bytes.NewReader(content))
		decoder.UseNumber()
		if decoder.Decode(&result) != nil {
			result = string(content)
		}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		reason := http.StatusText(response.StatusCode)
		if failure := m.object(m.object(result)["error"]); failure != nil {
			if text, ok := failure["reason"].(string); ok {
				reason = text
			}
			if cause, ok := failure["root_cause"].([]interface{}); ok && len(cause) > 0 {
				if text, ok := m.object(cause[0])["reason"].(string); ok {
					reason = text
				}
			}
		} else if text, ok := m.object(result)["error"].(string); ok {
			reason = text
		} else if text, ok := result.(string); ok && text != "" {
			reason = text
		}
		if len(reason) > 4096 {
			reason = reason[:4096]
		}
		return result, response.StatusCode, m.cleanError(fmt.Errorf("Elasticsearch HTTP %d：%s", response.StatusCode, reason))
	}
	return result, response.StatusCode, nil
}
