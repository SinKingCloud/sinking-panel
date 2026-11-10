package request

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Run 执行请求并写入响应日志，单次最多等待 30 秒、记录 64 KiB 响应体。
func (m *Manager) Run(ctx context.Context, content string, writeLog func(string)) error {
	data := Config{}
	if json.Unmarshal([]byte(content), &data) != nil || data.Url == "" {
		return errors.New("请求配置不合法")
	}
	startedAt := time.Now()
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("重定向次数过多")
			}
			return nil
		},
	}
	request, err := http.NewRequestWithContext(ctx, data.Method, data.Url, strings.NewReader(data.Body))
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	for name, value := range data.Headers {
		if strings.EqualFold(name, "Host") {
			request.Host = value
		} else {
			request.Header.Set(name, value)
		}
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 64*1024+1))
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	truncated := len(responseBody) > 64*1024
	if truncated {
		responseBody = responseBody[:64*1024]
	}
	result := "请求完成"
	if response.StatusCode >= http.StatusBadRequest {
		result = "请求异常"
	}
	log := fmt.Sprintf("%s: %s %s, 耗时: %s", result, data.Method, response.Status, time.Since(startedAt).Round(time.Millisecond))
	if len(responseBody) > 0 {
		log += "\n返回值:\n" + strings.ToValidUTF8(string(responseBody), "�")
	}
	if truncated {
		log += "\n[返回值超过64 KiB，已截断]"
	}
	if writeLog != nil {
		writeLog(log)
	}
	return nil
}
