package database

import (
	"context"
	"errors"
	"fmt"
	"server/app/enum/database_kind"
	"server/app/enum/system_task_status"
	clientElasticsearch "server/app/service/database/elasticsearch"
	"server/app/util/str"
)

// RefreshElasticsearchFields 将历史文档回填交给系统任务，关闭管理窗口不会中断执行。
func (s *service) RefreshElasticsearchFields(id int64, databaseName string) (string, error) {
	if err := clientElasticsearch.ValidateIndexName(databaseName, true); err != nil {
		return "", err
	}
	if s.systemService == nil {
		return "", errors.New("后台任务服务不可用")
	}
	data, err := s.FindConnection(id, database_kind.Elasticsearch)
	if err != nil {
		return "", err
	}
	taskID := str.GetSnowWorkIns().GetUuid()
	params := map[string]interface{}{"id": id, "database_name": databaseName, "action": "refresh_fields"}
	s.systemService.TaskCreate(taskID, "Elasticsearch 回填文档: "+data.Name+"/"+databaseName, params, func(ctx context.Context, _ interface{}, update func(int, float64, string)) {
		percent := float64(0)
		update(system_task_status.Running, 0, "正在连接 Elasticsearch")
		var result map[string]interface{}
		manager, err := s.Elasticsearch(ctx, id, databaseName)
		if err == nil {
			defer manager.Close()
			result, err = manager.RefreshFields(ctx, func(current float64, message string) {
				percent = max(percent, min(99, current))
				update(system_task_status.Running, percent, message)
			})
		}
		status, message := system_task_status.Completed, fmt.Sprintf("回填完成，已处理 %v 条文档", result["updated"])
		if ctx.Err() != nil || errors.Is(err, context.Canceled) {
			status, message = system_task_status.Canceled, "任务已取消"
		} else if err != nil {
			status, message = system_task_status.Failed, "回填失败: "+err.Error()
		} else {
			percent = 100
		}
		s.systemService.TaskUpdate(taskID, status, percent, message, result)
	})
	if s.systemService.GetTask(taskID) == nil {
		return "", errors.New("创建后台任务失败，面板正在停止")
	}
	return taskID, nil
}
