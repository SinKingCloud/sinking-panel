package elasticsearch

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"
)

// mappingDefinition 区分旧版类型包装和字段映射本身，包括只配置元信息的映射。
func mappingDefinition(value map[string]interface{}) bool {
	for _, key := range []string{"properties", "dynamic", "_source", "_meta", "_routing", "_all", "dynamic_templates", "date_detection", "numeric_detection", "dynamic_date_formats"} {
		if _, exists := value[key]; exists {
			return true
		}
	}
	return false
}

// Mapping 读取实际索引映射，保留旧版文档类型层级。
func (m *Manager) Mapping(ctx context.Context) (map[string]interface{}, error) {
	path, err := m.indexPath(false)
	if err != nil {
		return nil, err
	}
	data, _, err := m.doRequest(ctx, http.MethodGet, path+"/_mapping", nil)
	if err != nil {
		return nil, err
	}
	result := object(object(object(data)[m.index])["mappings"])
	if result == nil && m.major == 0 {
		result = object(object(data)[m.index])
	}
	if result == nil && m.major == 1 && len(object(data)) == 0 {
		result = map[string]interface{}{}
	}
	if result == nil {
		return nil, errors.New("Elasticsearch 响应缺少索引映射")
	}
	return result, nil
}

// UpdateMapping 添加映射字段，字段类型等不可修改项由服务器按版本校验。
func (m *Manager) UpdateMapping(ctx context.Context, mapping string) (map[string]interface{}, error) {
	path, err := m.indexPath(true)
	if err != nil {
		return nil, err
	}
	body, err := decodeObject(mapping, "字段映射", false)
	if err != nil {
		return nil, err
	}
	if m.major < 7 {
		typeName := ""
		if len(body) == 1 && !mappingDefinition(body) {
			for name, value := range body {
				if object(value) != nil {
					typeName, body = name, object(value)
				}
			}
		}
		typeName, err = m.documentType(ctx, typeName)
		if err != nil {
			return nil, err
		}
		if m.major == 0 {
			path += "/" + url.PathEscape(typeName) + "/_mapping"
		} else {
			path += "/_mapping/" + url.PathEscape(typeName)
		}
		if m.major < 2 {
			body = map[string]interface{}{typeName: body}
		}
	} else {
		path += "/_mapping"
	}
	data, _, err := m.doRequest(ctx, http.MethodPut, path, body)
	return object(data), err
}

// Settings 返回索引设置；不读取集群认证与密钥配置。
func (m *Manager) Settings(ctx context.Context) (map[string]interface{}, error) {
	path, err := m.indexPath(false)
	if err != nil {
		return nil, err
	}
	data, _, err := m.doRequest(ctx, http.MethodGet, path+"/_settings", nil)
	if err != nil {
		return nil, err
	}
	result := object(object(object(data)[m.index])["settings"])
	if result == nil {
		return nil, errors.New("Elasticsearch 响应缺少索引设置")
	}
	return result, nil
}

// UpdateSettings 修改当前普通索引的动态设置，不自动关闭索引或变更集群设置。
func (m *Manager) UpdateSettings(ctx context.Context, settings string) (map[string]interface{}, error) {
	path, err := m.indexPath(true)
	if err != nil {
		return nil, err
	}
	body, err := decodeObject(settings, "索引设置", false)
	if err != nil {
		return nil, err
	}
	data, _, err := m.doRequest(ctx, http.MethodPut, path+"/_settings", body)
	return object(data), err
}

// Status 并行读取集群健康和资源统计；旧版或无统计权限时仍返回健康信息。
func (m *Manager) Status(ctx context.Context) (map[string]interface{}, error) {
	requestCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var stats interface{}
	var statsErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		stats, _, statsErr = m.doRequest(requestCtx, http.MethodGet, "/_cluster/stats?timeout=10s", nil)
		if statsErr == nil && object(stats) == nil {
			statsErr = errors.New("Elasticsearch 响应缺少集群统计")
		}
	}()
	data, _, err := m.doRequest(requestCtx, http.MethodGet, "/_cluster/health", nil)
	if err != nil {
		cancel()
		<-done
		return nil, err
	}
	<-done
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result := map[string]interface{}{"health": data, "version": m.root}
	if statsErr != nil {
		result["stats_error"] = statsErr.Error()
	} else {
		result["stats"] = stats
	}
	return result, nil
}
