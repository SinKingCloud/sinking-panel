package request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

// Validate 校验请求配置并返回标准 JSON，单独填写 URL 时使用 GET。
func (m *Manager) Validate(content string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", errors.New("任务内容不能为空")
	}
	data := Config{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		data = Config{
			Method: http.MethodGet,
			Url:    strings.TrimSpace(content),
		}
	}
	data.Method = strings.ToUpper(strings.TrimSpace(data.Method))
	if data.Method == "" {
		data.Method = http.MethodGet
	}
	switch data.Method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions:
	default:
		return "", errors.New("请求方式不合法")
	}
	data.Url = strings.TrimSpace(data.Url)
	request, err := http.NewRequest(data.Method, data.Url, nil)
	if err != nil || request.URL.Host == "" || (!strings.EqualFold(request.URL.Scheme, "http") && !strings.EqualFold(request.URL.Scheme, "https")) {
		return "", errors.New("请求地址必须是有效的HTTP或HTTPS地址")
	}
	if len(data.Headers) > 100 {
		return "", errors.New("请求头数量不能超过100个")
	}
	validHeaderName := func(name string) bool {
		for i := 0; i < len(name); i++ {
			char := name[i]
			if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || strings.ContainsRune("!#$%&'*+-.^_`|~", rune(char)) {
				continue
			}
			return false
		}
		return name != ""
	}
	validHeaderValue := func(value string) bool {
		for i := 0; i < len(value); i++ {
			char := value[i]
			if char != '\t' && (char < 0x20 || char == 0x7f) {
				return false
			}
		}
		return true
	}
	headers := make(map[string]string, len(data.Headers))
	headerNames := make(map[string]struct{}, len(data.Headers))
	for name, value := range data.Headers {
		name = strings.TrimSpace(name)
		if !validHeaderName(name) {
			return "", errors.New("请求头名称不合法")
		}
		lowerName := strings.ToLower(name)
		if _, ok := headerNames[lowerName]; ok {
			return "", errors.New("请求头名称不能重复")
		}
		if lowerName == "content-length" || lowerName == "transfer-encoding" || lowerName == "trailer" {
			return "", errors.New("不支持自定义传输层请求头")
		}
		if !validHeaderValue(value) {
			return "", errors.New("请求头值不合法")
		}
		if lowerName == "host" && strings.TrimSpace(value) == "" {
			return "", errors.New("Host请求头不能为空")
		}
		headerNames[lowerName] = struct{}{}
		name = http.CanonicalHeaderKey(name)
		headers[name] = value
		if lowerName == "host" {
			request.Host = value
		} else {
			request.Header.Set(name, value)
		}
	}
	if err = request.Write(io.Discard); err != nil {
		return "", errors.New("请求配置不合法")
	}
	data.Headers = headers
	value, err := json.Marshal(data)
	if err != nil {
		return "", errors.New("请求配置格式化失败")
	}
	return string(value), nil
}
