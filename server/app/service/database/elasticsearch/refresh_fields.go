package elasticsearch

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// RefreshFields 使用现有文档按当前映射重新索引，使新增的多字段生效。
func (m *Manager) RefreshFields(ctx context.Context, progress func(float64, string)) (map[string]interface{}, error) {
	if m.major < 5 {
		return nil, errors.New("后台更新文档需要 Elasticsearch 5.0 或更高版本")
	}
	path, err := m.indexPath(true)
	if err != nil {
		return nil, err
	}
	query := url.Values{
		"wait_for_completion": {"false"},
		"conflicts":           {"proceed"},
		"refresh":             {"true"},
	}
	// 6.5 起才支持默认管道和 _none；旧版会把 _none 当作不存在的管道执行。
	if m.atLeast(6, 5) {
		query.Set("pipeline", "_none")
	}
	data, _, err := m.doRequest(ctx, http.MethodPost, path+"/_update_by_query?"+query.Encode(), map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	taskID, _ := object(data)["task"].(string)
	if taskID == "" {
		return nil, errors.New("Elasticsearch 未返回历史文档回填任务ID")
	}
	taskPath := "/_tasks/" + url.PathEscape(taskID)
	finished := false
	// 仅清理本次任务；未完成时尽力取消，已完成时删除持久结果，均不影响原操作结果。
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if !finished {
			_, _, _ = m.doRequest(cleanup, http.MethodPost, taskPath+"/_cancel", nil)
			return
		}
		documentType := "_doc"
		if m.major < 7 {
			documentType = "task"
		}
		_, _, _ = m.doRequest(cleanup, http.MethodDelete, "/.tasks/"+documentType+"/"+url.PathEscape(taskID), nil)
	}()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		data, _, err = m.doRequest(ctx, http.MethodGet, taskPath, nil)
		if err != nil {
			return nil, err
		}
		result := object(data)
		completed, valid := result["completed"].(bool)
		if !valid {
			return nil, errors.New("Elasticsearch 回填任务状态不完整")
		}
		if completed {
			finished = true
			if failure := result["error"]; failure != nil {
				reason, _ := object(failure)["reason"].(string)
				if reason == "" {
					reason = fmt.Sprint(failure)
				}
				return result, m.cleanError(errors.New(reason))
			}
			response := object(result["response"])
			if response == nil || response["total"] == nil || response["updated"] == nil {
				response = object(result["results"])
			}
			if response == nil || response["total"] == nil || response["updated"] == nil {
				return nil, errors.New("Elasticsearch 回填任务缺少执行结果")
			}
			if failures, _ := response["failures"].([]interface{}); len(failures) > 0 {
				cause := object(object(failures[0])["cause"])
				reason, _ := cause["reason"].(string)
				if reason == "" {
					reason = fmt.Sprint(failures[0])
				}
				return response, m.cleanError(fmt.Errorf("部分文档回填失败：%s", reason))
			}
			if response["timed_out"] == true {
				return response, errors.New("历史文档回填超时，部分文档可能尚未处理")
			}
			if conflicts := number(response["version_conflicts"]); conflicts > 0 {
				return response, fmt.Errorf("已有 %d 条文档完成回填，%d 条文档发生版本冲突，请重新回填", number(response["updated"]), conflicts)
			}
			if progress != nil {
				progress(100, fmt.Sprintf("已回填 %d 条文档", number(response["updated"])))
			}
			return response, nil
		}
		status := object(object(result["task"])["status"])
		if progress != nil && status != nil {
			updated, total := number(status["updated"]), number(status["total"])
			done := updated + number(status["deleted"]) + number(status["noops"]) + number(status["version_conflicts"])
			percent := float64(0)
			if total > 0 {
				percent = min(99, float64(done)*100/float64(total))
			}
			progress(percent, fmt.Sprintf("已处理 %d / %d 条文档", done, total))
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
