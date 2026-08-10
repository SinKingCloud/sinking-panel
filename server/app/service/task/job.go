package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"server/app/constant"
	"server/app/enum/task_exec_type"
	"server/app/model"
	"server/app/util/cmd"
	"strings"
	"time"
)

type job struct {
	*model.Task
	service *service
}

func newJob(task *model.Task, service *service) *job {
	if task == nil || service == nil || task.Spec == "" {
		return nil
	}
	content, err := service.checkContent(task.Script, task.ExecType)
	if err != nil {
		return nil
	}
	task.Script = content
	return &job{Task: task, service: service}
}

func (j *job) Run() {
	defer func() {
		_ = j.service.updateRuntimeById(j.Id, time.Now())
	}()
	switch j.ExecType {
	case task_exec_type.Script:
		c := cmd.NewScriptExec(constant.TempPath, 43200, func(s string) {
			_ = j.service.WriteLog(j.Id, s)
		})
		_, _, _ = c.Execute(j.Script)
	case task_exec_type.Request:
		data := Request{}
		if json.Unmarshal([]byte(j.Script), &data) != nil || data.Url == "" {
			_ = j.service.WriteLog(j.Id, "请求配置不合法")
			return
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
		request, err := http.NewRequest(data.Method, data.Url, strings.NewReader(data.Body))
		if err != nil {
			_ = j.service.WriteLog(j.Id, "请求失败: "+err.Error())
			return
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
			_ = j.service.WriteLog(j.Id, "请求失败: "+err.Error())
			return
		}
		defer response.Body.Close()
		responseBody, err := io.ReadAll(io.LimitReader(response.Body, 64*1024+1))
		if err != nil {
			_ = j.service.WriteLog(j.Id, "读取响应失败: "+err.Error())
			return
		}
		truncated := len(responseBody) > 64*1024
		if truncated {
			responseBody = responseBody[:64*1024]
		}
		result := "请求完成"
		if response.StatusCode >= http.StatusBadRequest {
			result = "请求异常"
		}
		content := fmt.Sprintf("%s: %s %s, 耗时: %s", result, data.Method, response.Status, time.Since(startedAt).Round(time.Millisecond))
		if len(responseBody) > 0 {
			content += "\n返回值:\n" + strings.ToValidUTF8(string(responseBody), "�")
		}
		if truncated {
			content += "\n[返回值超过64 KiB，已截断]"
		}
		_ = j.service.WriteLog(j.Id, content)
	}
}
